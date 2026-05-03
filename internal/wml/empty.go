package wml

func intPtr(n int) *int { p := n; return &p }

// EmptyDocument returns a minimal document with one empty paragraph (no runs).
func EmptyDocument() *Document {
	return &Document{
		Body: Body{
			Blocks: []BodyBlock{
				{Para: &Paragraph{Runs: nil}},
			},
		},
	}
}

// DefaultStyles returns minimal styles for a newly created document (OFFICE-107).
func DefaultStyles() *Styles {
	return &Styles{
		ByID: map[string]*Style{
			"Normal": {
				ID:   "Normal",
				Type: "paragraph",
				Name: "Normal",
			},
			"Heading1": {ID: "Heading1", Type: "paragraph", Name: "heading 1", BasedOn: "Normal", OutlineLevel: intPtr(0)},
			"Heading2": {ID: "Heading2", Type: "paragraph", Name: "heading 2", BasedOn: "Normal", OutlineLevel: intPtr(1)},
			"Heading3": {ID: "Heading3", Type: "paragraph", Name: "heading 3", BasedOn: "Normal", OutlineLevel: intPtr(2)},
			"ListParagraph": {
				ID: "ListParagraph", Type: "paragraph", Name: "List Paragraph", BasedOn: "Normal",
			},
		},
	}
}
