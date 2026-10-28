package preflight

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"printer/internal/book"
)

const dpi = 36

// halfLegal pages fill exactly half a Legal sheet, so scale is 1.
var halfLegal = book.Size{Width: book.Inch(7), Height: book.Inch(8.5)}

func blankPage() *image.Gray {
	img := image.NewGray(image.Rect(0, 0, 7*dpi, int(8.5*dpi)))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	return img
}

// ink paints a rectangle given in inches from the page's top-left corner.
func ink(img *image.Gray, x, y, w, h float64) {
	r := image.Rect(int(x*dpi), int(y*dpi), int((x+w)*dpi), int((y+h)*dpi))
	for py := r.Min.Y; py < r.Max.Y; py++ {
		for px := r.Min.X; px < r.Max.X; px++ {
			img.SetGray(px, py, color.Gray{Y: 0})
		}
	}
}

func run(t *testing.T, s book.Settings, images []*image.Gray) []Issue {
	t.Helper()
	sizes := make([]book.Size, len(images))
	for i := range sizes {
		sizes[i] = halfLegal
	}
	plan, err := book.NewPlan(len(images), s.SheetsPerSignature, s.Blanks)
	if err != nil {
		t.Fatal(err)
	}
	return Check(Input{Settings: s, Plan: plan, PageSizes: sizes, PageImages: images, ImageDPI: dpi})
}

func pages(n int) []*image.Gray {
	out := make([]*image.Gray, n)
	for i := range out {
		out[i] = blankPage()
	}
	return out
}

func find(issues []Issue, title string) *Issue {
	for i := range issues {
		if issues[i].Title == title {
			return &issues[i]
		}
	}
	return nil
}

func TestCleanBookHasOnlySummary(t *testing.T) {
	issues := run(t, book.DefaultSettings(), pages(8))
	if len(issues) != 1 || issues[0].Severity != Info {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestBlankPagesReported(t *testing.T) {
	issues := run(t, book.DefaultSettings(), pages(10))
	got := find(issues, "2 blank pages added")
	if got == nil || !strings.Contains(got.Detail, "before the back cover") {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestTextAtForeEdgeIsCutOff(t *testing.T) {
	imgs := pages(8)
	// Page 1 is a right-hand page: its fore edge is the image's right side.
	ink(imgs[0], 6.95, 4, 0.05, 0.3)
	issues := run(t, book.DefaultSettings(), imgs)
	got := find(issues, "Content will be cut off")
	if got == nil || !strings.Contains(got.Detail, "page 1 (outer edge)") {
		t.Fatalf("issues = %+v", issues)
	}
	if find(issues, "Content is close to the edge") != nil {
		t.Fatalf("cut-off edge was also reported as close: %+v", issues)
	}
}

func TestFullBleedArtworkIsInfo(t *testing.T) {
	imgs := pages(8)
	ink(imgs[2], 0, 0, 7, 8.5)
	issues := run(t, book.DefaultSettings(), imgs)
	got := find(issues, "Full-bleed artwork will get a white border")
	if got == nil || got.Severity != Info || got.Pages[0] != 3 {
		t.Fatalf("issues = %+v", issues)
	}
	if find(issues, "Content will be cut off") != nil {
		t.Fatalf("full-bleed page reported as cut off: %+v", issues)
	}
}

func TestContentNearSpine(t *testing.T) {
	imgs := pages(8)
	// Page 2 is a left-hand page: its spine is the image's right side.
	// 4 mm from the fold: printable, but inside the 6 mm safe zone.
	ink(imgs[1], 7-book.ToInch(book.MM(4)), 4, 0.05, 0.3)
	issues := run(t, book.DefaultSettings(), imgs)
	got := find(issues, "Content is close to the edge")
	if got == nil || !strings.Contains(got.Detail, "page 2 (spine)") {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestActualSizeTooLarge(t *testing.T) {
	s := book.DefaultSettings()
	s.Scaling = book.ScaleActual
	s.PaperID = "letter" // half Letter is 5.5 in wide, pages are 7 in
	issues := run(t, s, pages(4))
	if issues[0].Severity != Error || issues[0].Title != "Pages are larger than half a sheet" {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestPageList(t *testing.T) {
	if got := pageList([]int{1, 2, 3, 5, 7, 8}); got != "1–3, 5, 7, 8" {
		t.Fatalf("pageList = %q", got)
	}
}
