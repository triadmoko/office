package docx

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/triadmoko/office/internal/wml"
)

// PageSize holds width and height in twips and orientation.
type PageSize struct {
	Width, Height int64
	Orient        Orientation
}

// Orientation is portrait or landscape.
type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

// Margins holds page margins in twips (twentieths of a point).
type Margins struct {
	Top, Bottom, Left, Right int64
	Header, Footer, Gutter   int64
}

// Columns describes text columns in a section.
type Columns struct {
	Num        int
	Sep        bool
	EqualWidth bool
}

// Section is a document section (merged view over w:sectPr).
type Section struct {
	doc *Document
	idx int
	sec wml.Section
}

// Sections returns parsed sections (body sectPr + paragraph section breaks).
func (d *Document) Sections() []*Section {
	if d == nil {
		return nil
	}
	m, err := d.ensureLoaded()
	if err != nil || m == nil {
		return nil
	}
	raw := wml.SectionsFromDocument(m)
	out := make([]*Section, len(raw))
	for i := range raw {
		out[i] = &Section{doc: d, idx: i, sec: raw[i]}
	}
	return out
}

// SectionAt returns the i-th section for editing (builder).
func (d *Document) SectionAt(i int) *Section {
	if d == nil || i < 0 {
		return nil
	}
	secs := d.Sections()
	if i >= len(secs) {
		return nil
	}
	return secs[i]
}

// PageSize returns section page dimensions.
func (s *Section) PageSize() PageSize {
	if s == nil {
		return PageSize{}
	}
	return PageSize{
		Width: s.sec.PageSize.Width, Height: s.sec.PageSize.Height,
		Orient: Orientation(s.sec.PageSize.Orient),
	}
}

// Margins returns section page margins.
func (s *Section) Margins() Margins {
	if s == nil {
		return Margins{}
	}
	m := s.sec.Margins
	return Margins{
		Top: m.Top, Bottom: m.Bottom, Left: m.Left, Right: m.Right,
		Header: m.Header, Footer: m.Footer, Gutter: m.Gutter,
	}
}

// Columns returns w:cols summary.
func (s *Section) Columns() Columns {
	if s == nil {
		return Columns{}
	}
	c := s.sec.Columns
	return Columns{Num: c.Num, Sep: c.Sep, EqualWidth: c.EqualWidth}
}

// SetPageSize applies a standard page size (twips from ECMA defaults).
func (s *Section) SetPageSize(kind PageSizeKind) {
	if s == nil || s.doc == nil {
		return
	}
	switch kind {
	case PageSizeLetter:
		s.sec.PageSize.Width, s.sec.PageSize.Height = wml.PageLetterW, wml.PageLetterH
	case PageSizeA4:
		fallthrough
	default:
		s.sec.PageSize.Width, s.sec.PageSize.Height = wml.PageA4W, wml.PageA4H
	}
	s.apply()
}

// SetOrientation sets landscape or portrait (swaps W/H when switching to landscape if needed).
func (s *Section) SetOrientation(o Orientation) {
	if s == nil || s.doc == nil {
		return
	}
	s.sec.PageSize.Orient = wml.Orientation(o)
	if o == Landscape && s.sec.PageSize.Width < s.sec.PageSize.Height {
		s.sec.PageSize.Width, s.sec.PageSize.Height = s.sec.PageSize.Height, s.sec.PageSize.Width
	}
	s.apply()
}

// SetMargins sets w:pgMar (twips).
func (s *Section) SetMargins(m Margins) {
	if s == nil || s.doc == nil {
		return
	}
	s.sec.Margins = wml.Margins{
		Top: m.Top, Bottom: m.Bottom, Left: m.Left, Right: m.Right,
		Header: m.Header, Footer: m.Footer, Gutter: m.Gutter,
	}
	s.apply()
}

func (s *Section) apply() {
	m, _ := s.doc.ensureLoaded()
	if m == nil {
		return
	}
	// Only the first section maps to body-level sectPr for MVP.
	if s.idx == 0 {
		m.Body.SectPr = marshalSectPrBytes(s.sec)
	}
}

func marshalSectPrBytes(sec wml.Section) []byte {
	var b bytes.Buffer
	b.WriteString(`<w:sectPr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	ori := "portrait"
	if sec.PageSize.Orient == wml.Landscape {
		ori = "landscape"
	}
	w, h := sec.PageSize.Width, sec.PageSize.Height
	if w == 0 {
		w = wml.PageA4W
	}
	if h == 0 {
		h = wml.PageA4H
	}
	b.WriteString(`<w:pgSz w:w="` + strconv.FormatInt(w, 10) + `" w:h="` + strconv.FormatInt(h, 10) + `" w:orient="` + ori + `"/>`)
	m := sec.Margins
	b.WriteString(fmt.Sprintf(`<w:pgMar w:top="%d" w:right="%d" w:bottom="%d" w:left="%d" w:header="%d" w:footer="%d" w:gutter="%d"/>`,
		m.Top, m.Right, m.Bottom, m.Left, m.Header, m.Footer, m.Gutter))
	c := sec.Columns
	if c.Num <= 0 {
		c.Num = 1
	}
	sep := "0"
	if c.Sep {
		sep = "1"
	}
	eq := "0"
	if c.EqualWidth {
		eq = "1"
	}
	b.WriteString(`<w:cols w:num="` + strconv.Itoa(c.Num) + `" w:sep="` + sep + `" w:equalWidth="` + eq + `"/>`)
	b.WriteString(`</w:sectPr>`)
	return b.Bytes()
}
