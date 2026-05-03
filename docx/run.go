package docx

import "github.com/triadmoko/office/internal/wml"

// Run is a text run (w:r).
type Run struct {
	x   *wml.Run
	doc *Document
}

// Text returns logical text (w:t plus tab, line break, form feed U+000C for page break, soft hyphen).
func (r *Run) Text() string {
	if r == nil || r.x == nil {
		return ""
	}
	return r.x.Text
}

// ContainsPageBreak reports whether this run has w:br w:type="page".
func (r *Run) ContainsPageBreak() bool {
	if r == nil || r.x == nil {
		return false
	}
	for _, p := range r.x.Parts {
		if p.PageBreak {
			return true
		}
	}
	return false
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

// Emphasis returns w:em/@w:val (East Asian emphasis mark), empty if unset.
func (r *Run) Emphasis() string {
	if r == nil || r.x == nil {
		return ""
	}
	return r.x.RPr.Emphasis
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

// Highlight returns w:highlight/@w:val (ST_HighlightColor token, e.g. "yellow"), empty if unset.
func (r *Run) Highlight() string {
	if r == nil || r.x == nil {
		return ""
	}
	return r.x.RPr.Highlight
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

// SetStrike sets w:strike (single strikethrough). Marshal emits w:strike only (not w:dstrike).
func (r *Run) SetStrike(v bool) {
	if r == nil || r.x == nil {
		return
	}
	r.x.RPr.Strike = v
}

// SetEmphasis sets w:em (East Asian emphasis mark). Use OOXML ST_Em values, e.g. "dot", "comma", "circle", "underDot", or "" to clear.
func (r *Run) SetEmphasis(val string) {
	if r == nil || r.x == nil {
		return
	}
	r.x.RPr.Emphasis = val
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

// SetHighlight sets w:highlight (text marker / sorot). Use OOXML ST_HighlightColor tokens, e.g. "yellow", "green", "cyan", or "" to clear.
func (r *Run) SetHighlight(val string) {
	if r == nil || r.x == nil {
		return
	}
	r.x.RPr.Highlight = val
}

// SetSubSuperscript sets w:vertAlign: [VertAlignSuperscript], [VertAlignSubscript], or [VertAlignBaseline].
func (r *Run) SetSubSuperscript(v VertAlign) {
	if r == nil || r.x == nil {
		return
	}
	r.x.RPr.VertAlign = toWMLVert(v)
}
