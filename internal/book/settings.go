package book

import "fmt"

// Scaling decides how a book page is sized into its half of the sheet.
type Scaling string

const (
	// ScaleFit shrinks or enlarges the page to fill its half-sheet without
	// cropping.
	ScaleFit Scaling = "fit"
	// ScaleActual prints the page at its true size.
	ScaleActual Scaling = "actual"
)

// Duplex is how the second side of each sheet gets printed.
type Duplex string

const (
	DuplexAuto   Duplex = "auto"   // printer flips the sheet (short-edge binding)
	DuplexManual Duplex = "manual" // print fronts, reinsert the stack, print backs
)

// Settings is everything the user can tune for one print job.
type Settings struct {
	PaperID string `json:"paperId"`

	// SheetsPerSignature is 0 for a single saddle-stitched booklet, or the
	// number of sheets folded together per signature for sewn/glued books.
	SheetsPerSignature int            `json:"sheetsPerSignature"`
	Blanks             BlankPlacement `json:"blanks"`
	Scaling            Scaling        `json:"scaling"`

	// CreepMM shifts the pages of the innermost sheet toward the fold by this
	// much; outer sheets get proportionally less. Compensates for inner sheets
	// sticking out at the fore edge in thick booklets.
	CreepMM float64 `json:"creepMm"`

	// PrinterMarginMM is the band at the paper edges the printer cannot print.
	PrinterMarginMM float64 `json:"printerMarginMm"`
	// SafeMarginMM flags content closer than this to a page's edge.
	SafeMarginMM float64 `json:"safeMarginMm"`

	FoldMarks bool   `json:"foldMarks"`
	Duplex    Duplex `json:"duplex"`
}

// DefaultSettings matches the user's current setup: Legal paper folded in half,
// saddle stitched, on a Canon G6020.
func DefaultSettings() Settings {
	return Settings{
		PaperID:         "legal",
		Blanks:          BlanksBeforeBackCover,
		Scaling:         ScaleFit,
		PrinterMarginMM: 5,
		SafeMarginMM:    6,
		FoldMarks:       true,
		Duplex:          DuplexManual,
	}
}

// Validate reports settings that cannot produce a layout.
func (s Settings) Validate() error {
	if _, ok := PaperByID(s.PaperID); !ok {
		return fmt.Errorf("unknown paper size %q", s.PaperID)
	}
	if s.SheetsPerSignature < 0 {
		return fmt.Errorf("sheets per signature cannot be negative")
	}
	switch s.Blanks {
	case BlanksAtEnd, BlanksBeforeBackCover:
	default:
		return fmt.Errorf("unknown blank placement %q", s.Blanks)
	}
	switch s.Scaling {
	case ScaleFit, ScaleActual:
	default:
		return fmt.Errorf("unknown scaling %q", s.Scaling)
	}
	switch s.Duplex {
	case DuplexAuto, DuplexManual:
	default:
		return fmt.Errorf("unknown duplex mode %q", s.Duplex)
	}
	if s.CreepMM < 0 || s.PrinterMarginMM < 0 || s.SafeMarginMM < 0 {
		return fmt.Errorf("margins and creep cannot be negative")
	}
	return nil
}

// SheetSize is the paper in landscape, the orientation pages are laid out on.
func (s Settings) SheetSize() Size {
	p, _ := PaperByID(s.PaperID)
	return p.Size.Landscape()
}
