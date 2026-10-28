// Package book holds the print-layout domain: paper sizes, settings, how
// pages are distributed over folded sheets (the binding plan) and where each
// page lands on a sheet (geometry). It has no PDF dependencies.
//
// All lengths are in PDF points (1/72 inch) unless a name says otherwise.
package book

const (
	PointsPerInch = 72.0
	PointsPerMM   = 72.0 / 25.4
)

func MM(v float64) float64      { return v * PointsPerMM }
func Inch(v float64) float64    { return v * PointsPerInch }
func ToMM(pt float64) float64   { return pt / PointsPerMM }
func ToInch(pt float64) float64 { return pt / PointsPerInch }

// Size is a width/height pair in points.
type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Landscape returns the size with the long side horizontal.
func (s Size) Landscape() Size {
	if s.Width >= s.Height {
		return s
	}
	return Size{Width: s.Height, Height: s.Width}
}

// Rect is an axis-aligned rectangle in PDF space (origin bottom-left).
type Rect struct {
	X, Y, W, H float64
}

func (r Rect) Right() float64 { return r.X + r.W }
func (r Rect) Top() float64   { return r.Y + r.H }
