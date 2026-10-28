package book

import (
	"reflect"
	"testing"
)

func TestSaddleStitchEightPages(t *testing.T) {
	plan, err := NewPlan(8, 0, BlanksAtEnd)
	if err != nil {
		t.Fatal(err)
	}
	want := []Sheet{
		{Signature: 0, Nest: 0, NestCount: 2, Front: Side{8, 1}, Back: Side{2, 7}},
		{Signature: 0, Nest: 1, NestCount: 2, Front: Side{6, 3}, Back: Side{4, 5}},
	}
	if !reflect.DeepEqual(plan.Sheets, want) {
		t.Fatalf("sheets = %+v, want %+v", plan.Sheets, want)
	}
	if plan.BookPages != 8 || len(plan.BlankAt) != 0 || plan.Signatures != 1 {
		t.Fatalf("unexpected plan %+v", plan)
	}
}

func TestPaddingAtEnd(t *testing.T) {
	plan, _ := NewPlan(6, 0, BlanksAtEnd)
	if plan.BookPages != 8 || !reflect.DeepEqual(plan.BlankAt, []int{7, 8}) {
		t.Fatalf("plan = %+v", plan)
	}
	// Outer sheet front: page 8 (blank) | page 1.
	if plan.Sheets[0].Front != (Side{Blank, 1}) {
		t.Fatalf("outer front = %+v", plan.Sheets[0].Front)
	}
}

func TestPaddingBeforeBackCover(t *testing.T) {
	plan, _ := NewPlan(6, 0, BlanksBeforeBackCover)
	if !reflect.DeepEqual(plan.BlankAt, []int{6, 7}) {
		t.Fatalf("blankAt = %v", plan.BlankAt)
	}
	// Back cover (source page 6) must stay the last page of the book,
	// which sits on the outer sheet's front, left half.
	if plan.Sheets[0].Front != (Side{6, 1}) || plan.Sheets[0].Back != (Side{2, Blank}) {
		t.Fatalf("outer sheet = %+v", plan.Sheets[0])
	}
}

func TestSignatures(t *testing.T) {
	// 20 pages = 5 sheets; 2 sheets per signature -> 2 + 2 + 1.
	plan, _ := NewPlan(20, 2, BlanksAtEnd)
	if plan.Signatures != 3 || len(plan.Sheets) != 5 {
		t.Fatalf("plan = %+v", plan)
	}
	// Second signature covers pages 9..16.
	got := plan.Sheets[2]
	want := Sheet{Signature: 1, Nest: 0, NestCount: 2, Front: Side{16, 9}, Back: Side{10, 15}}
	if got != want {
		t.Fatalf("sheet 2 = %+v, want %+v", got, want)
	}
	last := plan.Sheets[4]
	if last.NestCount != 1 || last.Front != (Side{20, 17}) || last.Back != (Side{18, 19}) {
		t.Fatalf("last sheet = %+v", last)
	}
}

func TestEveryPageAppearsOnce(t *testing.T) {
	for pages := 1; pages <= 41; pages++ {
		for _, per := range []int{0, 1, 3, 4} {
			plan, err := NewPlan(pages, per, BlanksBeforeBackCover)
			if err != nil {
				t.Fatal(err)
			}
			seen := map[int]int{}
			for _, s := range plan.Sheets {
				for _, p := range []int{s.Front.Left, s.Front.Right, s.Back.Left, s.Back.Right} {
					seen[p]++
				}
			}
			for p := 1; p <= pages; p++ {
				if seen[p] != 1 {
					t.Fatalf("pages=%d per=%d: page %d placed %d times", pages, per, p, seen[p])
				}
			}
			if seen[Blank] != plan.BookPages-pages {
				t.Fatalf("pages=%d per=%d: %d blanks, want %d", pages, per, seen[Blank], plan.BookPages-pages)
			}
		}
	}
}

func TestNoPages(t *testing.T) {
	if _, err := NewPlan(0, 0, BlanksAtEnd); err == nil {
		t.Fatal("expected error")
	}
}
