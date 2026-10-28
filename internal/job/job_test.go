package job

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"printer/internal/book"
	"printer/internal/raster"
	"printer/internal/testpdf"
)

func TestAnalyzeEndToEnd(t *testing.T) {
	pages := make([]testpdf.Page, 10)
	for i := range pages {
		pages[i] = testpdf.Page{Width: book.Inch(7), Height: book.Inch(8.5)}
	}
	pages[0].EdgeMark = "right" // fore edge of the front cover
	path := filepath.Join(t.TempDir(), "book.pdf")
	if err := os.WriteFile(path, testpdf.Build(pages), 0o644); err != nil {
		t.Fatal(err)
	}

	r, err := raster.New()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	c := context.Background()
	doc, err := Load(c, r, path)
	if err != nil {
		t.Fatal(err)
	}
	res, err := doc.Analyze(c, r, book.DefaultSettings())
	if err != nil {
		t.Fatal(err)
	}

	// 10 pages -> 12 -> 3 sheets -> 6 sides.
	if len(res.Sides) != 6 {
		t.Fatalf("%d sides, want 6", len(res.Sides))
	}
	if s := res.Sides[0]; s.Left != "10" || s.Right != "1" || s.Back {
		t.Fatalf("first side = %+v", s)
	}
	if !strings.HasPrefix(res.Sides[0].Image, "data:image/jpeg;base64,") {
		t.Fatal("missing preview image")
	}

	var titles []string
	for _, is := range res.Issues {
		titles = append(titles, is.Title)
	}
	joined := strings.Join(titles, " | ")
	for _, want := range []string{"2 blank pages added", "Content will be cut off"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing issue %q in %s", want, joined)
		}
	}
}
