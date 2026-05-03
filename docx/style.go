package docx

import "github.com/triadmoko/office/internal/wml"

// Styles is the document style registry (styles.xml).
type Styles struct {
	x *wml.Styles
}

// Styles returns parsed styles, or nil if none could be loaded.
func (d *Document) Styles() *Styles {
	if d == nil {
		return nil
	}
	if _, err := d.ensureLoaded(); err != nil || d.styles == nil {
		return nil
	}
	return &Styles{x: d.styles}
}

// ByID returns a style by w:styleId.
func (s *Styles) ByID(id string) *Style {
	if s == nil || s.x == nil {
		return nil
	}
	stw := s.x.ByID[id]
	if stw == nil {
		return nil
	}
	return &Style{x: stw, reg: s.x}
}

// Style is one paragraph/character/table style.
type Style struct {
	x   *wml.Style
	reg *wml.Styles
}

// Type returns w:type (paragraph, character, table, numbering).
func (st *Style) Type() string {
	if st == nil || st.x == nil {
		return ""
	}
	return st.x.Type
}

// Name returns the display name.
func (st *Style) Name() string {
	if st == nil || st.x == nil {
		return ""
	}
	return st.x.Name
}

// BasedOn returns w:basedOn style id.
func (st *Style) BasedOn() string {
	if st == nil || st.x == nil {
		return ""
	}
	return st.x.BasedOn
}

// LinkedStyle returns w:link style id.
func (st *Style) LinkedStyle() string {
	if st == nil || st.x == nil {
		return ""
	}
	return st.x.LinkedStyle
}

// Resolved returns flattened formatting (OFFICE-104).
func (st *Style) Resolved() *ResolvedFormat {
	if st == nil || st.x == nil {
		return nil
	}
	r := st.x.Resolved(st.reg)
	if r == nil {
		return nil
	}
	return &ResolvedFormat{r: r}
}

// ResolvedFormat is merged rPr + pPr after style chain resolution.
type ResolvedFormat struct {
	r *wml.ResolvedFormat
}

// RunBold returns resolved bold.
func (rf *ResolvedFormat) RunBold() bool {
	if rf == nil || rf.r == nil {
		return false
	}
	return rf.r.RPr.Bold
}

// ParagraphAlignment returns resolved alignment.
func (rf *ResolvedFormat) ParagraphAlignment() Alignment {
	if rf == nil || rf.r == nil {
		return AlignLeft
	}
	return fromWMLAlignment(rf.r.PPr.Alignment)
}
