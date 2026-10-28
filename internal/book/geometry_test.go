package book

import (
	"math"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 0.01 }

func TestPlaceFitsHalfLegal(t *testing.T) {
	s := DefaultSettings() // Legal: landscape 14 × 8.5 in, halves 7 × 8.5 in
	pages := []Size{{Inch(5.5), Inch(8.5)}, {Inch(5.5), Inch(8.5)}}
	sheet := Sheet{NestCount: 1}
	got := PlaceSide(s, sheet, Side{Left: 2, Right: 1}, pages)

	l, r := got[0], got[1]
	if !near(l.Scale, 1) || !near(r.Scale, 1) {
		t.Fatalf("scale = %v, %v; want 1 (height already fits)", l.Scale, r.Scale)
	}
	// Left page ends at the fold, right page starts at it.
	if !near(l.X+l.Size.Width, Inch(7)) || !near(r.X, Inch(7)) {
		t.Fatalf("pages not flush with fold: left %+v right %+v", l, r)
	}
}

func TestPlaceScalesDownOversizedPage(t *testing.T) {
	s := DefaultSettings()
	pages := []Size{{Inch(8.5), Inch(11)}} // Letter page into a 7 × 8.5 half
	got := PlaceSide(s, Sheet{NestCount: 1}, Side{Left: Blank, Right: 1}, pages)
	if want := 8.5 / 11.0; !near(got[1].Scale, want) {
		t.Fatalf("scale = %v, want %v", got[1].Scale, want)
	}
	if got[0].Page != Blank || got[0].Scale != 0 {
		t.Fatalf("blank placement = %+v", got[0])
	}
}

func TestCreepGrowsInward(t *testing.T) {
	s := DefaultSettings()
	s.CreepMM = 1
	pages := []Size{{Inch(7), Inch(8.5)}, {Inch(7), Inch(8.5)}}
	outer := PlaceSide(s, Sheet{Nest: 0, NestCount: 3}, Side{2, 1}, pages)
	inner := PlaceSide(s, Sheet{Nest: 2, NestCount: 3}, Side{2, 1}, pages)
	if !near(outer[1].X, Inch(7)) {
		t.Fatalf("outer sheet should not creep: %v", outer[1].X)
	}
	if !near(inner[1].X, Inch(7)-MM(1)) || !near(inner[0].X, MM(1)) {
		t.Fatalf("inner sheet creep wrong: left %v right %v", inner[0].X, inner[1].X)
	}
}
