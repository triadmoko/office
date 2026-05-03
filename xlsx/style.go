package xlsx

// Style describes cell formatting resolved from styles.xml (read subset).
type Style struct {
	wb  *Workbook
	idx int // cellXfs index; -1 = none
}

// NumberFormat returns the number format string (built-in or custom).
func (s *Style) NumberFormat() string {
	if s == nil || s.wb == nil || s.idx < 0 {
		return "General"
	}
	return s.wb.stylesTable().NumberFormatForStyle(s.idx)
}

// Font returns font metadata when available (MVP: empty placeholder).
func (s *Style) Font() *FontFmt {
	return &FontFmt{}
}

// Fill returns fill metadata (MVP: empty placeholder).
func (s *Style) Fill() *FillFmt {
	return &FillFmt{}
}

// Border returns border metadata (MVP: empty placeholder).
func (s *Style) Border() *BorderFmt {
	return &BorderFmt{}
}

// Alignment returns alignment metadata (MVP: empty placeholder).
func (s *Style) Alignment() *AlignFmt {
	return &AlignFmt{}
}

// FontFmt is reserved for future style read support.
type FontFmt struct{}

// FillFmt is reserved for future style read support.
type FillFmt struct{}

// BorderFmt is reserved for future style read support.
type BorderFmt struct{}

// AlignFmt is reserved for future style read support.
type AlignFmt struct{}

// Style returns formatting for this cell when a style index is present.
func (c *Cell) Style() *Style {
	if c == nil || c.wb == nil || c.styleIdx < 0 {
		return &Style{wb: nil, idx: -1}
	}
	return &Style{wb: c.wb, idx: c.styleIdx}
}
