package book

import "fmt"

// Blank marks an empty slot in a plan (a padding page).
const Blank = 0

// Side is one face of a folded sheet: two book pages next to each other with
// the fold between them. Values are 1-based source page numbers, or Blank.
type Side struct {
	Left  int `json:"left"`
	Right int `json:"right"`
}

// Sheet is one physical piece of paper, printed on both sides and folded once.
type Sheet struct {
	Signature int  `json:"signature"` // 0-based signature this sheet belongs to
	Nest      int  `json:"nest"`      // 0 = outermost sheet of its signature
	NestCount int  `json:"nestCount"` // sheets in this signature
	Front     Side `json:"front"`
	Back      Side `json:"back"`
}

// Plan is the full assignment of book pages to sheets.
type Plan struct {
	SourcePages int     `json:"sourcePages"`
	BookPages   int     `json:"bookPages"` // after padding, always a multiple of 4
	BlankAt     []int   `json:"blankAt"`   // 1-based book positions that are blank padding
	Sheets      []Sheet `json:"sheets"`
	Signatures  int     `json:"signatures"`
}

// BlankPlacement decides where padding pages go.
type BlankPlacement string

const (
	// BlanksAtEnd appends blank pages after the last page.
	BlanksAtEnd BlankPlacement = "end"
	// BlanksBeforeBackCover keeps the last source page as the back cover and
	// inserts the blanks just before it.
	BlanksBeforeBackCover BlankPlacement = "before-back-cover"
)

// NewPlan distributes sourcePages over folded sheets.
//
// sheetsPerSignature is how many sheets are nested and folded together.
// 0 means a single signature holding every sheet (saddle stitch); any other
// value gives sewn/glued signatures, where the final signature may be smaller.
func NewPlan(sourcePages, sheetsPerSignature int, blanks BlankPlacement) (Plan, error) {
	if sourcePages < 1 {
		return Plan{}, fmt.Errorf("the document has no pages")
	}
	if sheetsPerSignature < 0 {
		return Plan{}, fmt.Errorf("sheets per signature must be positive")
	}

	order, blankAt := bookOrder(sourcePages, blanks)
	bookPages := len(order)
	totalSheets := bookPages / 4

	perSig := sheetsPerSignature
	if perSig == 0 || perSig > totalSheets {
		perSig = totalSheets
	}

	plan := Plan{SourcePages: sourcePages, BookPages: bookPages, BlankAt: blankAt}

	// Each signature is a small saddle-stitched booklet covering a
	// contiguous run of book pages.
	for start, sig := 0, 0; start < totalSheets; start, sig = start+perSig, sig+1 {
		nests := min(perSig, totalSheets-start)
		pages := order[start*4 : (start+nests)*4]
		n := len(pages)
		for i := range nests {
			// Positions are 0-based within this signature's pages.
			plan.Sheets = append(plan.Sheets, Sheet{
				Signature: sig,
				Nest:      i,
				NestCount: nests,
				Front:     Side{Left: pages[n-1-2*i], Right: pages[2*i]},
				Back:      Side{Left: pages[2*i+1], Right: pages[n-2-2*i]},
			})
		}
		plan.Signatures++
	}
	return plan, nil
}

// bookOrder returns the reading order of the finished book (source page
// numbers or Blank), padded to a multiple of 4, and the 1-based positions of
// the blanks.
func bookOrder(sourcePages int, blanks BlankPlacement) ([]int, []int) {
	padded := (sourcePages + 3) / 4 * 4
	missing := padded - sourcePages

	order := make([]int, 0, padded)
	for p := 1; p <= sourcePages; p++ {
		order = append(order, p)
	}

	insertAt := len(order) // append
	if blanks == BlanksBeforeBackCover && sourcePages > 1 {
		insertAt = len(order) - 1
	}

	var blankAt []int
	for i := range missing {
		blankAt = append(blankAt, insertAt+i+1)
	}
	tail := append([]int(nil), order[insertAt:]...)
	order = append(order[:insertAt], make([]int, missing)...)
	order = append(order, tail...)
	return order, blankAt
}
