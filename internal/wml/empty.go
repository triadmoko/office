package wml

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
			"Heading1": {ID: "Heading1", Type: "paragraph", Name: "heading 1", BasedOn: "Normal"},
			"Heading2": {ID: "Heading2", Type: "paragraph", Name: "heading 2", BasedOn: "Normal"},
			"Heading3": {ID: "Heading3", Type: "paragraph", Name: "heading 3", BasedOn: "Normal"},
			"ListParagraph": {
				ID: "ListParagraph", Type: "paragraph", Name: "List Paragraph", BasedOn: "Normal",
			},
		},
	}
}
