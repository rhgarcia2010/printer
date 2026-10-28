package book

// Placement says where one book page is drawn on a sheet side.
type Placement struct {
	Page  int     // 1-based source page, or Blank
	Slot  Rect    // the half-sheet; drawing is clipped to it
	Scale float64 // uniform scale applied to the page
	X, Y  float64 // where the page's lower-left corner lands
	Size  Size    // the page's placed (scaled) size
}

// PlaceSide lays out one side of a sheet. pageSizes[i] is the size of source
// page i+1; blanks get a Placement with only Slot set.
func PlaceSide(s Settings, sheet Sheet, side Side, pageSizes []Size) [2]Placement {
	size := s.SheetSize()
	fold := size.Width / 2
	left := Rect{X: 0, Y: 0, W: fold, H: size.Height}
	right := Rect{X: fold, Y: 0, W: fold, H: size.Height}

	creep := creepFor(s, sheet)
	return [2]Placement{
		place(s, side.Left, left, pageSizes, fold, true, creep),
		place(s, side.Right, right, pageSizes, fold, false, creep),
	}
}

func place(s Settings, page int, slot Rect, pageSizes []Size, fold float64, isLeft bool, creep float64) Placement {
	p := Placement{Page: page, Slot: slot}
	if page == Blank || page > len(pageSizes) {
		return p
	}
	src := pageSizes[page-1]

	p.Scale = 1
	if s.Scaling == ScaleFit {
		p.Scale = min(slot.W/src.Width, slot.H/src.Height)
	}
	p.Size = Size{Width: src.Width * p.Scale, Height: src.Height * p.Scale}

	// Pages sit against the fold so facing pages meet at the spine; any
	// leftover width ends up at the fore edge. Creep pushes them further
	// toward (and slightly past) the fold.
	if isLeft {
		p.X = fold - p.Size.Width + creep
	} else {
		p.X = fold - creep
	}
	p.Y = slot.Y + (slot.H-p.Size.Height)/2
	return p
}

// creepFor grows linearly from 0 on the outermost sheet to CreepMM on the
// innermost sheet of the signature.
func creepFor(s Settings, sheet Sheet) float64 {
	if s.CreepMM == 0 || sheet.NestCount < 2 {
		return 0
	}
	return MM(s.CreepMM) * float64(sheet.Nest) / float64(sheet.NestCount-1)
}
