package preflight

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"printer/internal/book"
)

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func isAre(n int) string {
	if n == 1 {
		return "is"
	}
	return "are"
}

// pageList formats sorted page numbers compactly: "1–3, 7, 9".
func pageList(pages []int) string {
	var parts []string
	for i := 0; i < len(pages); {
		j := i
		for j+1 < len(pages) && pages[j+1] == pages[j]+1 {
			j++
		}
		switch {
		case j == i:
			parts = append(parts, strconv.Itoa(pages[i]))
		case j == i+1:
			parts = append(parts, strconv.Itoa(pages[i]), strconv.Itoa(pages[j]))
		default:
			parts = append(parts, fmt.Sprintf("%d–%d", pages[i], pages[j]))
		}
		i = j + 1
	}
	return strings.Join(parts, ", ")
}

// sizeLabel shows a size in inches and millimetres: "7 × 8.5 in (178 × 216 mm)".
func sizeLabel(s book.Size) string {
	in := func(pt float64) string { return strconv.FormatFloat(round(book.ToInch(pt), 2), 'f', -1, 64) }
	return fmt.Sprintf("%s × %s in (%.0f × %.0f mm)", in(s.Width), in(s.Height), book.ToMM(s.Width), book.ToMM(s.Height))
}

func round(v float64, places int) float64 {
	f, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'f', places, 64), 64)
	return f
}

func shortPaperName(p book.Paper) string {
	name, _, _ := strings.Cut(p.Name, " (")
	return name
}

func mostCommonSize(sizes []book.Size) book.Size {
	type key struct{ w, h int }
	counts := map[key]int{}
	var best book.Size
	bestN := 0
	for _, s := range sizes {
		k := key{int(s.Width + 0.5), int(s.Height + 0.5)}
		counts[k]++
		if counts[k] > bestN {
			best, bestN = s, counts[k]
		}
	}
	return best
}

// describeHits lists affected pages and edges: "page 3 (outer edge, top),
// page 5 (spine)". Long lists are shortened.
func describeHits(hits []edgeHit) string {
	edges := map[int][]string{}
	for _, h := range hits {
		name := string(h.edge)
		if h.edge == ForeEdge {
			name = "outer edge"
		}
		if !contains(edges[h.page], name) {
			edges[h.page] = append(edges[h.page], name)
		}
	}
	pages := hitPages(hits)
	const maxListed = 8
	var parts []string
	for i, p := range pages {
		if i == maxListed {
			parts = append(parts, fmt.Sprintf("and %d more", len(pages)-maxListed))
			break
		}
		parts = append(parts, fmt.Sprintf("page %d (%s)", p, strings.Join(edges[p], ", ")))
	}
	return strings.Join(parts, ", ")
}

func hitPages(hits []edgeHit) []int {
	seen := map[int]bool{}
	var pages []int
	for _, h := range hits {
		if !seen[h.page] {
			seen[h.page] = true
			pages = append(pages, h.page)
		}
	}
	sort.Ints(pages)
	return pages
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
