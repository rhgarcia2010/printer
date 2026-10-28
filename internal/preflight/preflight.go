// Package preflight checks a book and its print settings for problems that
// would only show up once paper comes out of the printer.
package preflight

import (
	"fmt"
	"image"
	"math"
	"sort"
	"strings"

	"printer/internal/book"
)

type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
	Info    Severity = "info"
)

// Issue is one finding, phrased for the person printing the book.
type Issue struct {
	Severity Severity `json:"severity"`
	Title    string   `json:"title"`
	Detail   string   `json:"detail"`
	Pages    []int    `json:"pages,omitempty"` // source pages involved
}

// Input is everything the checks look at.
type Input struct {
	Settings  book.Settings
	Plan      book.Plan
	PageSizes []book.Size
	// PageImages are grayscale renders of each source page (index = page-1)
	// at ImageDPI. Edge checks are skipped when nil.
	PageImages []*image.Gray
	ImageDPI   int
}

// Check runs every check. Issues come back most severe first.
func Check(in Input) []Issue {
	var issues []Issue
	issues = append(issues, checkPageCount(in)...)
	issues = append(issues, checkPageSizes(in)...)
	issues = append(issues, checkScaling(in)...)
	issues = append(issues, checkEdges(in)...)

	rank := map[Severity]int{Error: 0, Warning: 1, Info: 2}
	sort.SliceStable(issues, func(i, j int) bool {
		return rank[issues[i].Severity] < rank[issues[j].Severity]
	})
	return issues
}

func checkPageCount(in Input) []Issue {
	p := in.Plan
	sheets := len(p.Sheets)
	summary := fmt.Sprintf("%d pages on %d sheets of paper", p.BookPages, sheets)
	if p.Signatures > 1 {
		summary += fmt.Sprintf(" in %d signatures", p.Signatures)
	}
	if len(p.BlankAt) == 0 {
		return []Issue{{Severity: Info, Title: summary,
			Detail: fmt.Sprintf("%d pages is a multiple of 4, so no blank pages are needed.", p.SourcePages)}}
	}

	where := "at the end of the book"
	if in.Settings.Blanks == book.BlanksBeforeBackCover {
		where = "just before the back cover"
	}
	return []Issue{{
		Severity: Warning,
		Title:    fmt.Sprintf("%d blank page%s added", len(p.BlankAt), plural(len(p.BlankAt))),
		Detail: fmt.Sprintf("Folded sheets hold 4 pages each, and the document has %d. "+
			"Blank pages were added %s (book positions %s), giving %s. "+
			"Add or remove pages in Canva to avoid them.",
			p.SourcePages, where, pageList(p.BlankAt), summary),
	}}
}

func checkPageSizes(in Input) []Issue {
	if len(in.PageSizes) == 0 {
		return nil
	}
	common := mostCommonSize(in.PageSizes)
	var odd []int
	for i, s := range in.PageSizes {
		if math.Abs(s.Width-common.Width) > 1 || math.Abs(s.Height-common.Height) > 1 {
			odd = append(odd, i+1)
		}
	}
	if len(odd) == 0 {
		return nil
	}
	return []Issue{{
		Severity: Warning,
		Title:    "Pages have different sizes",
		Detail: fmt.Sprintf("Most pages are %s, but page%s %s %s not. They will be scaled differently.",
			sizeLabel(common), plural(len(odd)), pageList(odd), isAre(len(odd))),
		Pages: odd,
	}}
}

func checkScaling(in Input) []Issue {
	if len(in.PageSizes) == 0 {
		return nil
	}
	sheet := in.Settings.SheetSize()
	half := book.Size{Width: sheet.Width / 2, Height: sheet.Height}
	page := mostCommonSize(in.PageSizes)
	paper, _ := book.PaperByID(in.Settings.PaperID)

	ideal := fmt.Sprintf("For %s folded in half, design pages at %s.", shortPaperName(paper), sizeLabel(half))

	if in.Settings.Scaling == book.ScaleActual {
		if page.Width > half.Width+0.5 || page.Height > half.Height+0.5 {
			return []Issue{{
				Severity: Error,
				Title:    "Pages are larger than half a sheet",
				Detail: fmt.Sprintf("Pages are %s but half a sheet is %s, so printing at actual size cuts them off. "+
					"Switch scaling to \"Fit\" or pick larger paper. %s", sizeLabel(page), sizeLabel(half), ideal),
			}}
		}
		return nil
	}

	scale := min(half.Width/page.Width, half.Height/page.Height)
	placed := book.Size{Width: page.Width * scale, Height: page.Height * scale}
	spareW := book.ToMM(half.Width - placed.Width)
	spareH := book.ToMM(half.Height - placed.Height)

	var notes []string
	if math.Abs(scale-1) > 0.01 {
		notes = append(notes, fmt.Sprintf("Pages are %s and get scaled to %.0f%% to fit half a sheet.",
			sizeLabel(page), scale*100))
	}
	if spareW > 2 {
		notes = append(notes, fmt.Sprintf("That leaves %.0f mm of extra paper at each fore edge, which you can trim off.", spareW))
	}
	if spareH > 2 {
		notes = append(notes, fmt.Sprintf("That leaves %.0f mm of extra paper above and below each page.", spareH/2))
	}
	if len(notes) == 0 {
		return nil
	}
	notes = append(notes, ideal)
	return []Issue{{Severity: Info, Title: "Pages don't match the paper exactly", Detail: strings.Join(notes, " ")}}
}

// Edge identifies a side of a book page by its role in the bound book.
type Edge string

const (
	Top      Edge = "top"
	Bottom   Edge = "bottom"
	ForeEdge Edge = "outer"
	Spine    Edge = "spine"
)

// Full-bleed artwork covers most of the edge band; smaller hits are probably
// text or a detail that would be visibly cut.
const fullBleedCoverage = 0.6

// "Ink" is anything noticeably darker than white paper.
const inkThreshold = 245

type edgeHit struct {
	page     int
	edge     Edge
	coverage float64
}

func checkEdges(in Input) []Issue {
	if in.PageImages == nil {
		return nil
	}
	margin := book.MM(in.Settings.PrinterMarginMM)
	safe := book.MM(in.Settings.SafeMarginMM)
	sheet := in.Settings.SheetSize()

	var clipped, bleed, close []edgeHit
	for _, sh := range in.Plan.Sheets {
		for _, side := range []book.Side{sh.Front, sh.Back} {
			for i, p := range book.PlaceSide(in.Settings, sh, side, in.PageSizes) {
				if p.Page == book.Blank || p.Page > len(in.PageImages) {
					continue
				}
				img := in.PageImages[p.Page-1]
				onLeft := i == 0
				lostInk := map[Edge]bool{}
				for _, b := range lostBands(p, sheet, margin, onLeft) {
					cov := coverage(img, b, in.ImageDPI)
					if cov == 0 {
						continue
					}
					lostInk[b.edge] = true
					hit := edgeHit{p.Page, b.edge, cov}
					if cov >= fullBleedCoverage {
						bleed = append(bleed, hit)
					} else {
						clipped = append(clipped, hit)
					}
				}
				// Safe-zone check: printable but uncomfortably close to the
				// edge of the paper or the fold. Edges already reported as
				// cut off are not repeated.
				for _, b := range safeBands(p, sheet, margin, safe, onLeft) {
					if lostInk[b.edge] {
						continue
					}
					if cov := coverage(img, b, in.ImageDPI); cov > 0 && cov < fullBleedCoverage {
						close = append(close, edgeHit{p.Page, b.edge, cov})
					}
				}
			}
		}
	}

	var issues []Issue
	if len(clipped) > 0 {
		issues = append(issues, Issue{
			Severity: Warning,
			Title:    "Content will be cut off",
			Detail: fmt.Sprintf("The printer can't print within %.0f mm of the paper edge. Content falls in that strip on: %s.",
				in.Settings.PrinterMarginMM, describeHits(clipped)),
			Pages: hitPages(clipped),
		})
	}
	if len(bleed) > 0 {
		issues = append(issues, Issue{
			Severity: Info,
			Title:    "Full-bleed artwork will get a white border",
			Detail: fmt.Sprintf("Artwork runs to the edge on %s. The printer leaves about %.0f mm unprinted there, "+
				"so trim the printed book or keep important details away from the edges.",
				describeHits(bleed), in.Settings.PrinterMarginMM),
			Pages: hitPages(bleed),
		})
	}
	if len(close) > 0 {
		issues = append(issues, Issue{
			Severity: Warning,
			Title:    "Content is close to the edge",
			Detail: fmt.Sprintf("Content is within %.0f mm of the edge on %s. It will print, but may be uncomfortably "+
				"close to the fold or the trimmed edge.", in.Settings.SafeMarginMM, describeHits(close)),
			Pages: hitPages(close),
		})
	}
	return issues
}

// band is a strip along one page edge, in page points.
type band struct {
	edge     Edge
	physical side    // which side of the page image the strip is on
	from, to float64 // distance in from the edge where the strip starts and ends
}

type side int

const (
	sideTop side = iota
	sideBottom
	sideLeft
	sideRight
)

// physicalSide maps a book edge to the page image: on the left half of a
// sheet the fore edge is the page's left side, on the right half its right.
func physicalSide(e Edge, onLeft bool) side {
	switch e {
	case Top:
		return sideTop
	case Bottom:
		return sideBottom
	case ForeEdge:
		if onLeft {
			return sideLeft
		}
		return sideRight
	default: // Spine
		if onLeft {
			return sideRight
		}
		return sideLeft
	}
}

var allEdges = []Edge{Top, Bottom, ForeEdge, Spine}

// lostBands are the strips of a placed page that don't make it onto paper:
// they fall in the printer's unprintable margin or past the fold.
func lostBands(p book.Placement, sheet book.Size, margin float64, onLeft bool) []band {
	var out []band
	for _, e := range allEdges {
		if d := lostDepth(p, sheet, margin, onLeft, e); d > 0 {
			out = append(out, band{edge: e, physical: physicalSide(e, onLeft), to: d})
		}
	}
	return out
}

// safeBands are the printable strips closer than safe (in sheet points) to
// an edge.
func safeBands(p book.Placement, sheet book.Size, margin, safe float64, onLeft bool) []band {
	var out []band
	for _, e := range allEdges {
		from := lostDepth(p, sheet, margin, onLeft, e)
		to := safe / p.Scale
		if to > from {
			out = append(out, band{edge: e, physical: physicalSide(e, onLeft), from: from, to: to})
		}
	}
	return out
}

// lostDepth is how deep, in page points, the lost strip along edge e is.
func lostDepth(p book.Placement, sheet book.Size, margin float64, onLeft bool, e Edge) float64 {
	fold := sheet.Width / 2
	left, right := p.X, p.X+p.Size.Width
	var sheetDepth float64
	switch e {
	case Top:
		sheetDepth = p.Y + p.Size.Height - (sheet.Height - margin)
	case Bottom:
		sheetDepth = margin - p.Y
	case ForeEdge:
		if onLeft {
			sheetDepth = margin - left
		} else {
			sheetDepth = right - (sheet.Width - margin)
		}
	case Spine:
		// Anything pushed past the fold is clipped at the fold.
		if onLeft {
			sheetDepth = right - fold
		} else {
			sheetDepth = fold - left
		}
	}
	return max(0, sheetDepth/p.Scale)
}

// coverage is the fraction of inked pixels in the band.
func coverage(img *image.Gray, b band, dpi int) float64 {
	r := bandRect(img.Bounds(), b, dpi)
	if r.Empty() {
		return 0
	}
	inked := 0
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if img.GrayAt(x, y).Y < inkThreshold {
				inked++
			}
		}
	}
	return float64(inked) / float64(r.Dx()*r.Dy())
}

// bandRect converts a band to pixels. Image y grows downward.
func bandRect(bounds image.Rectangle, b band, dpi int) image.Rectangle {
	px := func(pt float64) int { return int(math.Round(pt / book.PointsPerInch * float64(dpi))) }
	from, to := px(b.from), max(px(b.to), px(b.from)+1)
	var r image.Rectangle
	switch b.physical {
	case sideTop:
		r = image.Rect(bounds.Min.X, bounds.Min.Y+from, bounds.Max.X, bounds.Min.Y+to)
	case sideBottom:
		r = image.Rect(bounds.Min.X, bounds.Max.Y-to, bounds.Max.X, bounds.Max.Y-from)
	case sideLeft:
		r = image.Rect(bounds.Min.X+from, bounds.Min.Y, bounds.Min.X+to, bounds.Max.Y)
	case sideRight:
		r = image.Rect(bounds.Max.X-to, bounds.Min.Y, bounds.Max.X-from, bounds.Max.Y)
	}
	return r.Intersect(bounds)
}
