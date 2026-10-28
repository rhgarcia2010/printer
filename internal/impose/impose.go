// Package impose writes the print-ready PDF: book pages arranged onto
// two-up sheets according to a book.Plan.
package impose

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/filter"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"

	"printer/internal/book"
)

// Sides selects which sheet sides to write.
type Sides string

const (
	AllSides   Sides = "all"    // front, back, front, back… for automatic duplex
	FrontsOnly Sides = "fronts" // first pass of manual duplex
	BacksOnly  Sides = "backs"  // second pass of manual duplex
)

// Options tunes the output for how the paper gets fed back into the printer.
type Options struct {
	Sides        Sides
	ReverseOrder bool // write sheets last-to-first
	RotateBacks  bool // turn back sides 180°
}

// PageSizes reads the visible size of every page (crop box, rotation applied).
func PageSizes(c context.Context, src []byte) ([]book.Size, error) {
	ctx, err := api.ReadContext(c, bytes.NewReader(src), nil)
	if err != nil {
		return nil, fmt.Errorf("read PDF: %w", err)
	}
	if err := ctx.EnsurePageCount(); err != nil {
		return nil, err
	}
	sizes := make([]book.Size, ctx.PageCount)
	for i := range sizes {
		_, _, attrs, err := ctx.PageDict(c, i+1, false)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", i+1, err)
		}
		box := visibleBox(attrs)
		sizes[i] = book.Size{Width: box.Width(), Height: box.Height()}
	}
	return sizes, nil
}

// Write imposes the source PDF and writes the result to w.
func Write(c context.Context, src []byte, s book.Settings, plan book.Plan, opts Options, w io.Writer) error {
	ctx, err := api.ReadContext(c, bytes.NewReader(src), nil)
	if err != nil {
		return fmt.Errorf("read PDF: %w", err)
	}
	if err := ctx.EnsurePageCount(); err != nil {
		return err
	}
	if ctx.PageCount != plan.SourcePages {
		return fmt.Errorf("PDF has %d pages but the plan expects %d", ctx.PageCount, plan.SourcePages)
	}

	im := &imposer{c: c, ctx: ctx, settings: s, forms: map[int]formRef{}}
	if err := im.loadPages(); err != nil {
		return err
	}

	sheetSize := s.SheetSize()
	pagesDict := types.Dict{
		"Type":     types.Name("Pages"),
		"Count":    types.Integer(0),
		"MediaBox": types.RectForDim(sheetSize.Width, sheetSize.Height).Array(),
	}
	pagesRef, err := ctx.IndRefForNewObject(pagesDict)
	if err != nil {
		return err
	}

	sheets := plan.Sheets
	if opts.ReverseOrder {
		sheets = make([]book.Sheet, len(plan.Sheets))
		for i, sh := range plan.Sheets {
			sheets[len(sheets)-1-i] = sh
		}
	}
	pageCount := 0
	for _, sheet := range sheets {
		if opts.Sides != BacksOnly {
			if err := im.addSide(sheet, sheet.Front, false, pagesDict, pagesRef); err != nil {
				return err
			}
			pageCount++
		}
		if opts.Sides != FrontsOnly {
			if err := im.addSide(sheet, sheet.Back, opts.RotateBacks, pagesDict, pagesRef); err != nil {
				return err
			}
			pageCount++
		}
	}

	root, err := ctx.Catalog()
	if err != nil {
		return err
	}
	root.Update("Pages", *pagesRef)
	ctx.PageCount = pageCount

	return api.WriteContext(c, ctx, w)
}

type formRef struct {
	ref  types.IndirectRef
	name string
}

type imposer struct {
	c        context.Context
	ctx      *model.Context
	settings book.Settings
	sizes    []book.Size
	forms    map[int]formRef // source page -> form XObject
}

// loadPages turns every source page into a form XObject so it can be drawn
// anywhere on a sheet, and records its size.
func (im *imposer) loadPages() error {
	im.sizes = make([]book.Size, im.ctx.PageCount)
	for nr := 1; nr <= im.ctx.PageCount; nr++ {
		d, _, attrs, err := im.ctx.PageDict(im.c, nr, true)
		if err != nil {
			return fmt.Errorf("page %d: %w", nr, err)
		}
		box := visibleBox(attrs)
		im.sizes[nr-1] = book.Size{Width: box.Width(), Height: box.Height()}

		content, err := im.ctx.PageContent(d, nr)
		if errors.Is(err, model.ErrNoContent) {
			continue // an empty page; nothing to draw
		}
		if err != nil {
			return fmt.Errorf("page %d content: %w", nr, err)
		}
		ref, err := im.pageForm(attrs, content)
		if err != nil {
			return fmt.Errorf("page %d: %w", nr, err)
		}
		im.forms[nr] = formRef{ref: *ref, name: fmt.Sprintf("P%d", nr)}
	}
	return nil
}

// visibleBox is the crop box (or media box) with page rotation applied to
// its dimensions.
func visibleBox(attrs *model.InheritedPageAttrs) *types.Rectangle {
	box := attrs.MediaBox
	if attrs.CropBox != nil {
		box = attrs.CropBox
	}
	box = box.Clone()
	if attrs.Rotate%180 != 0 {
		w := box.Width()
		box.UR.X = box.LL.X + box.Height()
		box.UR.Y = box.LL.Y + w
	}
	return box
}

func (im *imposer) pageForm(attrs *model.InheritedPageAttrs, content []byte) (*types.IndirectRef, error) {
	resRef, err := im.ctx.IndRefForNewObject(attrs.Resources)
	if err != nil {
		return nil, err
	}
	box := visibleBox(attrs)
	if attrs.Rotate != 0 {
		content = append(model.ContentBytesForPageRotation(attrs.Rotate, box.Width(), box.Height()), content...)
	}
	sd := types.StreamDict{
		Dict: types.Dict{
			"Type":    types.Name("XObject"),
			"Subtype": types.Name("Form"),
			"BBox":    box.Array(),
			// Move the box's lower-left corner to the origin so placement
			// math only deals with width and height.
			"Matrix":    types.NewNumberArray(1, 0, 0, 1, -box.LL.X, -box.LL.Y),
			"Resources": *resRef,
		},
		Content:        content,
		FilterPipeline: []types.PDFFilter{{Name: filter.Flate}},
	}
	sd.InsertName("Filter", filter.Flate)
	if err := sd.Encode(); err != nil {
		return nil, err
	}
	return im.ctx.IndRefForNewObject(sd)
}

func (im *imposer) addSide(sheet book.Sheet, side book.Side, rotate bool, pagesDict types.Dict, pagesRef *types.IndirectRef) error {
	size := im.settings.SheetSize()
	var buf bytes.Buffer
	xobjects := types.Dict{}

	if rotate {
		fmt.Fprintf(&buf, "-1 0 0 -1 %.3f %.3f cm\n", size.Width, size.Height)
	}
	for _, p := range book.PlaceSide(im.settings, sheet, side, im.sizes) {
		form, ok := im.forms[p.Page]
		if !ok {
			continue // blank padding or empty source page
		}
		xobjects[form.name] = form.ref
		fmt.Fprintf(&buf, "q %.3f %.3f %.3f %.3f re W n %.5f 0 0 %.5f %.3f %.3f cm /%s Do Q\n",
			p.Slot.X, p.Slot.Y, p.Slot.W, p.Slot.H,
			p.Scale, p.Scale, p.X, p.Y, form.name)
	}
	if im.settings.FoldMarks {
		writeFoldMarks(&buf, size, book.MM(im.settings.PrinterMarginMM))
	}

	contents, err := im.ctx.NewStreamDictForBuf(buf.Bytes())
	if err != nil {
		return err
	}
	if err := contents.Encode(); err != nil {
		return err
	}
	contentsRef, err := im.ctx.IndRefForNewObject(*contents)
	if err != nil {
		return err
	}
	resRef, err := im.ctx.IndRefForNewObject(types.Dict{"XObject": xobjects})
	if err != nil {
		return err
	}
	pageRef, err := im.ctx.IndRefForNewObject(types.Dict{
		"Type":      types.Name("Page"),
		"Parent":    *pagesRef,
		"MediaBox":  types.RectForDim(size.Width, size.Height).Array(),
		"Resources": *resRef,
		"Contents":  *contentsRef,
	})
	if err != nil {
		return err
	}
	return model.AppendPageTree(pageRef, 1, pagesDict)
}

// writeFoldMarks draws short ticks on the fold line just inside the printable
// area at the top and bottom edges.
func writeFoldMarks(buf *bytes.Buffer, size book.Size, margin float64) {
	x := size.Width / 2
	tick := book.MM(4)
	fmt.Fprintf(buf, "q 0.6 G 0.4 w %.3f %.3f m %.3f %.3f l S %.3f %.3f m %.3f %.3f l S Q\n",
		x, margin, x, margin+tick,
		x, size.Height-margin, x, size.Height-margin-tick)
}
