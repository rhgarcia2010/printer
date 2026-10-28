package impose

import (
	"bytes"
	"context"
	"image"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"

	"printer/internal/book"
	"printer/internal/raster"
	"printer/internal/testpdf"
)

func halfLetterPages(n int) []testpdf.Page {
	pages := make([]testpdf.Page, n)
	for i := range pages {
		pages[i] = testpdf.Page{Width: book.Inch(5.5), Height: book.Inch(8.5)}
	}
	return pages
}

func TestWriteProducesOnePagePerSheetSide(t *testing.T) {
	src := testpdf.Build(halfLetterPages(6))
	s := book.DefaultSettings()
	plan, _ := book.NewPlan(6, 0, s.Blanks)

	for _, tc := range []struct {
		sides Sides
		want  int
	}{{AllSides, 4}, {FrontsOnly, 2}, {BacksOnly, 2}} {
		var out bytes.Buffer
		if err := Write(context.Background(), src, s, plan, Options{Sides: tc.sides}, &out); err != nil {
			t.Fatal(err)
		}
		ctx, err := api.ReadContext(context.Background(), bytes.NewReader(out.Bytes()), nil)
		if err != nil {
			t.Fatalf("output is not a readable PDF: %v", err)
		}
		if err := ctx.EnsurePageCount(); err != nil {
			t.Fatal(err)
		}
		if err := api.ValidateContext(context.Background(), ctx); err != nil {
			t.Fatalf("output fails PDF validation: %v", err)
		}
		if ctx.PageCount != tc.want {
			t.Fatalf("%s: %d pages, want %d", tc.sides, ctx.PageCount, tc.want)
		}
	}
}

func TestPagesLandOnTheRightHalf(t *testing.T) {
	pages := halfLetterPages(8)
	pages[0].EdgeMark = "right" // page 1: fore edge of the front cover
	src := testpdf.Build(pages)
	s := book.DefaultSettings()
	s.FoldMarks = false
	plan, _ := book.NewPlan(8, 0, s.Blanks)

	var out bytes.Buffer
	if err := Write(context.Background(), src, s, plan, Options{Sides: AllSides}, &out); err != nil {
		t.Fatal(err)
	}

	r, err := raster.New()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	const dpi = 36
	err = r.Render(out.Bytes(), dpi, raster.Gray, func(page int, img image.Image) error {
		if page != 1 {
			return nil
		}
		// Sheet is Legal landscape (14 × 8.5 in). Page 1 sits right of the
		// fold (x = 7 in) and is 5.5 in wide, so its fore edge is at 12.5 in.
		// The square spans 4.25–4.67 in from the bottom, i.e. 3.83–4.25 in
		// from the top in image coordinates.
		px := func(in float64) int { return int(in * dpi) }
		y := px(4.0)
		if !dark(img, px(12.4), y) {
			t.Errorf("expected page 1 edge mark near x=12.4in")
		}
		if dark(img, px(1.6), y) {
			t.Errorf("edge mark showed up on the left half")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func dark(img image.Image, x, y int) bool {
	g := img.(*image.Gray)
	return g.GrayAt(x, y).Y < 100
}
