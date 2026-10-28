package book

// Paper is a sheet size the printer can feed, stored in portrait orientation.
type Paper struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size Size   `json:"size"`
}

// Papers lists the sizes offered in the UI. The Canon G6020 rear tray and
// cassette both take all of these.
var Papers = []Paper{
	{ID: "letter", Name: "US Letter (8.5 × 11 in)", Size: Size{Inch(8.5), Inch(11)}},
	{ID: "legal", Name: "US Legal (8.5 × 14 in)", Size: Size{Inch(8.5), Inch(14)}},
	{ID: "a4", Name: "A4 (210 × 297 mm)", Size: Size{MM(210), MM(297)}},
	{ID: "a5", Name: "A5 (148 × 210 mm)", Size: Size{MM(148), MM(210)}},
}

// PaperByID returns the paper with the given id, or false.
func PaperByID(id string) (Paper, bool) {
	for _, p := range Papers {
		if p.ID == id {
			return p, true
		}
	}
	return Paper{}, false
}
