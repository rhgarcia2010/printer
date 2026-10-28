// Package job ties the pieces together: load a book PDF, check it against
// print settings, and produce previews and print-ready output.
package job

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"printer/internal/book"
	"printer/internal/impose"
	"printer/internal/preflight"
	"printer/internal/raster"
)

// checkDPI is the resolution source pages are rendered at for edge checks;
// one pixel is about half a millimetre.
const checkDPI = 50

// previewDPI keeps a Legal sheet preview around 1100 px wide.
const previewDPI = 80

// Document is a loaded book PDF.
type Document struct {
	Path  string
	data  []byte
	sizes []book.Size
	gray  []*image.Gray
}

// Load reads the PDF and renders its pages for the edge checks.
func Load(c context.Context, r *raster.Renderer, path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sizes, err := impose.PageSizes(c, data)
	if err != nil {
		return nil, err
	}
	doc := &Document{Path: path, data: data, sizes: sizes}
	err = r.Render(data, checkDPI, raster.Gray, func(page int, img image.Image) error {
		doc.gray = append(doc.gray, cloneGray(img))
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(doc.gray) != len(sizes) {
		return nil, fmt.Errorf("rendered %d pages but the PDF has %d", len(doc.gray), len(sizes))
	}
	return doc, nil
}

func cloneGray(img image.Image) *image.Gray {
	out := image.NewGray(img.Bounds())
	draw.Draw(out, out.Bounds(), img, img.Bounds().Min, draw.Src)
	return out
}

// Info summarizes the document for the UI.
type Info struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	PageCount int       `json:"pageCount"`
	PageSize  book.Size `json:"pageSize"` // size of the first page
}

func (d *Document) Info() Info {
	return Info{
		Path:      d.Path,
		Name:      filepath.Base(d.Path),
		PageCount: len(d.sizes),
		PageSize:  d.sizes[0],
	}
}

// SidePreview is one printed side of one sheet.
type SidePreview struct {
	Sheet     int    `json:"sheet"` // 1-based
	Signature int    `json:"signature"`
	Back      bool   `json:"back"`
	Left      string `json:"left"`  // page label on the left half
	Right     string `json:"right"` // page label on the right half
	Image     string `json:"image"` // data: URL
}

// Result is what the UI shows for a document + settings combination.
type Result struct {
	Plan   book.Plan         `json:"plan"`
	Issues []preflight.Issue `json:"issues"`
	Sides  []SidePreview     `json:"sides"`
	Sheet  book.Size         `json:"sheet"`
}

// Analyze lays out the book, runs the checks and renders a preview of every
// sheet side.
func (d *Document) Analyze(c context.Context, r *raster.Renderer, s book.Settings) (*Result, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	plan, err := book.NewPlan(len(d.sizes), s.SheetsPerSignature, s.Blanks)
	if err != nil {
		return nil, err
	}
	res := &Result{
		Plan:  plan,
		Sheet: s.SheetSize(),
		Issues: preflight.Check(preflight.Input{
			Settings: s, Plan: plan, PageSizes: d.sizes, PageImages: d.gray, ImageDPI: checkDPI,
		}),
	}

	var pdf bytes.Buffer
	if err := impose.Write(c, d.data, s, plan, impose.Options{Sides: impose.AllSides}, &pdf); err != nil {
		return nil, err
	}
	err = r.Render(pdf.Bytes(), previewDPI, raster.Color, func(page int, img image.Image) error {
		sheet := plan.Sheets[(page-1)/2]
		back := page%2 == 0
		side := sheet.Front
		if back {
			side = sheet.Back
		}
		var jpg bytes.Buffer
		if err := jpeg.Encode(&jpg, img, &jpeg.Options{Quality: 85}); err != nil {
			return err
		}
		res.Sides = append(res.Sides, SidePreview{
			Sheet:     (page-1)/2 + 1,
			Signature: sheet.Signature + 1,
			Back:      back,
			Left:      pageLabel(side.Left),
			Right:     pageLabel(side.Right),
			Image:     "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(jpg.Bytes()),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func pageLabel(p int) string {
	if p == book.Blank {
		return "blank"
	}
	return strconv.Itoa(p)
}

// Write produces the print-ready PDF.
func (d *Document) Write(c context.Context, s book.Settings, opts impose.Options, w io.Writer) error {
	if err := s.Validate(); err != nil {
		return err
	}
	plan, err := book.NewPlan(len(d.sizes), s.SheetsPerSignature, s.Blanks)
	if err != nil {
		return err
	}
	return impose.Write(c, d.data, s, plan, opts, w)
}
