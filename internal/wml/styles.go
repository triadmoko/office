package wml

// Styles is a parsed word/styles.xml registry.
type Styles struct {
	DocDefaults ParagraphProps
	RDefaults   RunProps
	ByID        map[string]*Style
}

// Style is one w:style definition.
type Style struct {
	ID          string
	Type        string // paragraph, character, table, numbering
	Name        string
	BasedOn     string
	LinkedStyle string
	RPr         RunProps
	PPr         ParagraphProps
	Raw         []byte // full w:style element for round-trip
}

// ResolvedFormat is flattened character + paragraph formatting (OFFICE-104).
type ResolvedFormat struct {
	RPr RunProps
	PPr ParagraphProps
}

// Resolved merges this style's chain (BasedOn) with document defaults.
func (s *Style) Resolved(doc *Styles) *ResolvedFormat {
	if s == nil {
		return nil
	}
	var rf ResolvedFormat
	if doc != nil {
		rf.RPr = doc.RDefaults
		rf.PPr = doc.DocDefaults
	}
	seen := make(map[string]bool)
	cur := s
	for cur != nil {
		if cur.ID != "" && seen[cur.ID] {
			break
		}
		if cur.ID != "" {
			seen[cur.ID] = true
		}
		rf.RPr = mergeRunProps(cur.RPr, rf.RPr)
		rf.PPr = mergeParaProps(cur.PPr, rf.PPr)
		if cur.BasedOn == "" || doc == nil {
			break
		}
		cur = doc.ByID[cur.BasedOn]
	}
	return &rf
}

func mergeRunProps(overlay, base RunProps) RunProps {
	out := base
	if overlay.Bold {
		out.Bold = true
	}
	if overlay.Italic {
		out.Italic = true
	}
	if overlay.Underline {
		out.Underline = true
	}
	if overlay.Strike {
		out.Strike = true
	}
	if overlay.VertAlign != VertAlignBaseline {
		out.VertAlign = overlay.VertAlign
	}
	if overlay.FontSizeHalf != 0 {
		out.FontSizeHalf = overlay.FontSizeHalf
	}
	if overlay.Color != "" {
		out.Color = overlay.Color
	}
	if overlay.FontName != "" {
		out.FontName = overlay.FontName
	}
	return out
}

func mergeParaProps(overlay, base ParagraphProps) ParagraphProps {
	out := base
	if overlay.Alignment != AlignLeft || base.Alignment != AlignLeft {
		if overlay.Alignment != AlignLeft {
			out.Alignment = overlay.Alignment
		}
	}
	if overlay.StyleID != "" {
		out.StyleID = overlay.StyleID
	}
	if overlay.Numbering != nil {
		out.Numbering = overlay.Numbering
	}
	if overlay.Indent.Left != 0 || overlay.Indent.Right != 0 || overlay.Indent.FirstLine != 0 || overlay.Indent.Hanging != 0 {
		out.Indent = overlay.Indent
	}
	if overlay.Spacing.Before != 0 || overlay.Spacing.After != 0 || overlay.Spacing.Line != 0 || overlay.Spacing.LineRule != LineRuleUnset {
		out.Spacing = overlay.Spacing
	}
	return out
}
