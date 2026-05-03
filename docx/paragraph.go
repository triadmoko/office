package docx

import (
	"github.com/triadmoko/office/internal/wml"
)

// Paragraph is a document paragraph (w:p).
type Paragraph struct {
	x   *wml.Paragraph
	doc *Document
}

// Runs returns w:r children in order.
func (p *Paragraph) Runs() []*Run {
	if p == nil || p.x == nil {
		return nil
	}
	out := make([]*Run, len(p.x.Runs))
	for i, r := range p.x.Runs {
		out[i] = &Run{x: r, doc: p.doc}
	}
	return out
}

// Alignment returns w:pPr/w:jc (default Left).
func (p *Paragraph) Alignment() Alignment {
	if p == nil || p.x == nil {
		return AlignLeft
	}
	return fromWMLAlignment(p.x.PPr.Alignment)
}

// Indent returns w:ind-derived values (twentieths of a point).
func (p *Paragraph) Indent() Indent {
	if p == nil || p.x == nil {
		return Indent{}
	}
	return fromWMLIndent(p.x.PPr.Indent)
}

// Spacing returns w:spacing.
func (p *Paragraph) Spacing() Spacing {
	if p == nil || p.x == nil {
		return Spacing{}
	}
	return fromWMLSpacing(p.x.PPr.Spacing)
}

// StyleID returns w:pStyle/@w:val.
func (p *Paragraph) StyleID() string {
	if p == nil || p.x == nil {
		return ""
	}
	return p.x.PPr.StyleID
}

// NumberingRef returns w:numPr or nil.
func (p *Paragraph) NumberingRef() *NumPr {
	if p == nil || p.x == nil {
		return nil
	}
	return fromWMLNum(p.x.PPr.Numbering)
}

// SetIndent sets w:ind (values in twips / dxa: left, right, firstLine, hanging).
func (p *Paragraph) SetIndent(i Indent) {
	if p == nil || p.x == nil {
		return
	}
	p.x.PPr.Indent = toWMLIndent(i)
}

// SetAlignment sets w:jc (paragraph alignment).
func (p *Paragraph) SetAlignment(a Alignment) {
	if p == nil || p.x == nil {
		return
	}
	p.x.PPr.Alignment = toWMLAlignment(a)
}

// SetSpacing sets w:spacing (before/after/line in twips; line rule optional).
func (p *Paragraph) SetSpacing(s Spacing) {
	if p == nil || p.x == nil {
		return
	}
	p.x.PPr.Spacing = toWMLSpacing(s)
}

func (p *Paragraph) applyListRef(numID, ilvl int) {
	if p == nil || p.x == nil {
		return
	}
	p.x.PPr.Numbering = &wml.NumPr{NumID: numID, Ilvl: ilvl}
	p.x.PPr.StyleID = "ListParagraph"
}

// AppendRun adds a run with plain text to this paragraph.
func (p *Paragraph) AppendRun(text string) *Run {
	if p == nil || p.x == nil {
		return nil
	}
	r := &wml.Run{Parts: []wml.RunPart{{Text: text}}}
	r.RebuildText()
	p.x.Runs = append(p.x.Runs, r)
	return &Run{x: r, doc: p.doc}
}

// AppendPageBreak inserts w:br w:type="page" (mulai isi berikutnya di halaman baru).
func (p *Paragraph) AppendPageBreak() {
	if p == nil || p.x == nil {
		return
	}
	r := &wml.Run{Parts: []wml.RunPart{{PageBreak: true}}}
	r.RebuildText()
	p.x.Runs = append(p.x.Runs, r)
}

// SetSectionBreak writes w:pPr/w:sectPr: isi setelah paragraf ini memakai bagian baru dengan ukuran/orientasi cfg.
// Pastikan dokumen punya sectPr penutup di body: jika belum, diisi default A4 portrait.
func (p *Paragraph) SetSectionBreak(cfg SectionBreakConfig) {
	if p == nil || p.x == nil || p.doc == nil {
		return
	}
	m, _ := p.doc.ensureLoaded()
	if m == nil {
		return
	}
	p.x.PPr.SectPr = marshalSectPrBytes(sectionFromBreakConfig(cfg))
	if len(m.Body.SectPr) == 0 {
		m.Body.SectPr = marshalSectPrBytes(wml.ParseSectPr(nil))
	}
}
