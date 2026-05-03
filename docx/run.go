package docx

import "github.com/triadmoko/office/internal/wml"

// Run is a text run (w:r).
type Run struct {
	x   *wml.Run
	doc *Document
}

// Text returns logical text (w:t plus tab, line break, soft hyphen).
func (r *Run) Text() string {
	if r == nil || r.x == nil {
		return ""
	}
	return r.x.Text
}

// Bold reports w:b in w:rPr.
func (r *Run) Bold() bool {
	if r == nil || r.x == nil {
		return false
	}
	return r.x.RPr.Bold
}

// Italic reports w:i.
func (r *Run) Italic() bool {
	if r == nil || r.x == nil {
		return false
	}
	return r.x.RPr.Italic
}

// Underline reports w:u (non-none).
func (r *Run) Underline() bool {
	if r == nil || r.x == nil {
		return false
	}
	return r.x.RPr.Underline
}

// Strike reports w:strike.
func (r *Run) Strike() bool {
	if r == nil || r.x == nil {
		return false
	}
	return r.x.RPr.Strike
}

// SubSuperscript returns vertical alignment (baseline, superscript, subscript).
func (r *Run) SubSuperscript() VertAlign {
	if r == nil || r.x == nil {
		return VertAlignBaseline
	}
	return fromWMLVert(r.x.RPr.VertAlign)
}

// FontSize returns w:sz in half-points (0 if unset).
func (r *Run) FontSize() int {
	if r == nil || r.x == nil {
		return 0
	}
	return r.x.RPr.FontSizeHalf
}

// Color returns w:color as RRGGBB without # (empty if unset).
func (r *Run) Color() string {
	if r == nil || r.x == nil {
		return ""
	}
	return r.x.RPr.Color
}

// FontName returns w:rFonts ascii/hAnsi.
func (r *Run) FontName() string {
	if r == nil || r.x == nil {
		return ""
	}
	return r.x.RPr.FontName
}

// SetBold sets w:b on the run (builder).
func (r *Run) SetBold(v bool) {
	if r == nil || r.x == nil {
		return
	}
	r.x.RPr.Bold = v
}

// SetItalic sets w:i.
func (r *Run) SetItalic(v bool) {
	if r == nil || r.x == nil {
		return
	}
	r.x.RPr.Italic = v
}

// SetUnderline sets w:u.
func (r *Run) SetUnderline(v bool) {
	if r == nil || r.x == nil {
		return
	}
	r.x.RPr.Underline = v
}

// SetFont sets w:rFonts ascii and hAnsi.
func (r *Run) SetFont(name string) {
	if r == nil || r.x == nil {
		return
	}
	r.x.RPr.FontName = name
}

// SetSize sets w:sz in half-points.
func (r *Run) SetSize(halfPoints int) {
	if r == nil || r.x == nil {
		return
	}
	r.x.RPr.FontSizeHalf = halfPoints
}

// SetColor sets w:color (RRGGBB without #).
func (r *Run) SetColor(hex string) {
	if r == nil || r.x == nil {
		return
	}
	r.x.RPr.Color = hex
}
