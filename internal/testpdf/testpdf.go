// Package testpdf generates small PDFs for tests and manual trials.
package testpdf

import (
	"bytes"
	"fmt"
)

// Page describes one generated page.
type Page struct {
	Width, Height float64 // points
	// EdgeMark draws a solid square touching this edge: "left", "right",
	// "top", "bottom" or "" for none.
	EdgeMark string
}

// Build returns a PDF whose pages each show their page number in the middle.
func Build(pages []Page) []byte {
	var objs []string
	// 1: catalog, 2: pages, 3: font, then page + content pairs.
	kids := ""
	for i := range pages {
		kids += fmt.Sprintf("%d 0 R ", 4+2*i)
	}
	objs = append(objs,
		"<< /Type /Catalog /Pages 2 0 R >>",
		fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", kids, len(pages)),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	)
	for i, p := range pages {
		content := pageContent(i+1, p)
		objs = append(objs,
			fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Resources << /Font << /F1 3 0 R >> >> /Contents %d 0 R >>",
				p.Width, p.Height, 5+2*i),
			fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
		)
	}

	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objs))
	for i, o := range objs {
		offsets[i] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for _, off := range offsets {
		fmt.Fprintf(&b, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, xref)
	return b.Bytes()
}

func pageContent(nr int, p Page) string {
	var b bytes.Buffer
	// Light frame 20pt inside the edge, and the page number.
	fmt.Fprintf(&b, "0.8 G 1 w 20 20 %.2f %.2f re S\n", p.Width-40, p.Height-40)
	fmt.Fprintf(&b, "0 g BT /F1 72 Tf %.2f %.2f Td (%d) Tj ET\n", p.Width/2-20, p.Height/2-25, nr)

	const sq = 30.0
	switch p.EdgeMark {
	case "left":
		fmt.Fprintf(&b, "0 g 0 %.2f %.2f %.2f re f\n", p.Height/2, sq, sq)
	case "right":
		fmt.Fprintf(&b, "0 g %.2f %.2f %.2f %.2f re f\n", p.Width-sq, p.Height/2, sq, sq)
	case "top":
		fmt.Fprintf(&b, "0 g %.2f %.2f %.2f %.2f re f\n", p.Width/2, p.Height-sq, sq, sq)
	case "bottom":
		fmt.Fprintf(&b, "0 g %.2f 0 %.2f %.2f re f\n", p.Width/2, sq, sq)
	}
	return b.String()
}
