package docx

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

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

// PageNumberFormat is w:pgNumType/@w:fmt (ST_NumberFormat). Empty ([PageNumberFormatDefault]) omits the attribute so Word uses its default (typically decimal).
// See ECMA-376 ST_NumberFormat; common values: [PageNumberFormatDecimal], [PageNumberFormatUpperRoman], [PageNumberFormatLowerRoman], letters, ordinal, cardinal.
type PageNumberFormat string

const (
	PageNumberFormatDefault     PageNumberFormat = ""
	PageNumberFormatDecimal     PageNumberFormat = "decimal"
	PageNumberFormatUpperRoman  PageNumberFormat = "upperRoman"
	PageNumberFormatLowerRoman  PageNumberFormat = "lowerRoman"
	PageNumberFormatUpperLetter PageNumberFormat = "upperLetter"
	PageNumberFormatLowerLetter PageNumberFormat = "lowerLetter"
	PageNumberFormatOrdinal     PageNumberFormat = "ordinal"
	PageNumberFormatCardinal    PageNumberFormat = "cardinal"
)

// SectionBreakKind selects w:type/@w:val (ST_SectionMark) inside w:sectPr.
type SectionBreakKind int

const (
	// SectionBreakUnset leaves w:type unset when editing an existing section with [Section.SetBreakKind];
	// for [Paragraph.SetSectionBreak], unset is treated as nextPage.
	SectionBreakUnset SectionBreakKind = iota
	SectionBreakNextPage
	SectionBreakContinuous
	SectionBreakNextColumn
	SectionBreakEvenPage
	SectionBreakOddPage
)

// SectionBreakConfig describes w:sectPr stored on this paragraph (pemecah bagian: properti untuk isi setelah paragraf ini).
type SectionBreakConfig struct {
	PageKind PageSizeKind
	Orient   Orientation
	Margins  Margins
	// Break is w:type (nextPage, continuous, …). Zero defaults to nextPage for [Paragraph.SetSectionBreak].
	Break SectionBreakKind
	// PageNumberFormat is w:pgNumType/@w:fmt (romawi, huruf, …). [PageNumberFormatDefault] = omit.
	PageNumberFormat PageNumberFormat
	// PageNumberStart, if non-nil, sets w:pgNumType/@w:start (nomor halaman pertama section untuk bidang PAGE).
	PageNumberStart *int
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

// BreakKind returns w:type/@w:val as [SectionBreakKind], or [SectionBreakUnset] if absent.
func (s *Section) BreakKind() SectionBreakKind {
	if s == nil {
		return SectionBreakUnset
	}
	return sectionBreakKindFromWML(s.sec.TypeVal)
}

// PageNumberFormat returns w:pgNumType/@w:fmt, or [PageNumberFormatDefault] if unset.
func (s *Section) PageNumberFormat() PageNumberFormat {
	if s == nil {
		return PageNumberFormatDefault
	}
	return PageNumberFormat(strings.TrimSpace(s.sec.PageNumFmt))
}

// SetPageNumberFormat sets w:pgNumType/@w:fmt. Use [PageNumberFormatDefault] to clear format (still may emit pgNumType if start is set).
func (s *Section) SetPageNumberFormat(f PageNumberFormat) {
	if s == nil || s.doc == nil {
		return
	}
	s.sec.PageNumFmt = string(strings.TrimSpace(string(f)))
	s.apply()
}

// PageNumberStart returns (n, true) if w:pgNumType/@w:start is set.
func (s *Section) PageNumberStart() (n int, ok bool) {
	if s == nil || !s.sec.PageNumStartSet {
		return 0, false
	}
	return s.sec.PageNumStart, true
}

// SetPageNumberStart sets w:pgNumType/@w:start for this section. Pass nil to clear; otherwise first PAGE value uses *p (typically ≥ 1).
func (s *Section) SetPageNumberStart(p *int) {
	if s == nil || s.doc == nil {
		return
	}
	if p == nil {
		s.sec.PageNumStartSet = false
		s.sec.PageNumStart = 0
	} else {
		s.sec.PageNumStartSet = true
		s.sec.PageNumStart = *p
	}
	s.apply()
}

// SetBreakKind sets w:type for this section (empty [SectionBreakUnset] clears w:type on next marshal).
func (s *Section) SetBreakKind(k SectionBreakKind) {
	if s == nil || s.doc == nil {
		return
	}
	s.sec.TypeVal = sectionBreakKindToWML(k)
	s.apply()
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
	sinks := sectPrWriteTargets(m)
	if s.idx < 0 || s.idx >= len(sinks) {
		return
	}
	data := marshalSectPrBytes(s.sec)
	t := sinks[s.idx]
	if t.body {
		m.Body.SectPr = data
	} else {
		t.para.PPr.SectPr = data
	}
}

type sectSink struct {
	para *wml.Paragraph
	body bool
}

// sectPrWriteTargets mirrors [wml.SectionsFromDocument] storage order (paragraph sectPr in preorder, then body).
func sectPrWriteTargets(m *wml.Document) []sectSink {
	if m == nil {
		return nil
	}
	var sinks []sectSink
	for _, p := range wml.CollectParagraphsPreorder(&m.Body) {
		if len(p.PPr.SectPr) > 0 {
			sinks = append(sinks, sectSink{para: p})
		}
	}
	if len(m.Body.SectPr) > 0 {
		sinks = append(sinks, sectSink{body: true})
	} else if len(sinks) == 0 {
		sinks = append(sinks, sectSink{body: true})
	}
	return sinks
}

func sectionFromBreakConfig(c SectionBreakConfig) wml.Section {
	var sec wml.Section
	switch c.PageKind {
	case PageSizeLetter:
		sec.PageSize.Width, sec.PageSize.Height = wml.PageLetterW, wml.PageLetterH
	default:
		sec.PageSize.Width, sec.PageSize.Height = wml.PageA4W, wml.PageA4H
	}
	sec.PageSize.Orient = wml.Orientation(c.Orient)
	if c.Orient == Landscape && sec.PageSize.Width < sec.PageSize.Height {
		sec.PageSize.Width, sec.PageSize.Height = sec.PageSize.Height, sec.PageSize.Width
	}
	sec.Margins = wml.Margins{
		Top: c.Margins.Top, Bottom: c.Margins.Bottom, Left: c.Margins.Left, Right: c.Margins.Right,
		Header: c.Margins.Header, Footer: c.Margins.Footer, Gutter: c.Margins.Gutter,
	}
	sec.Columns = wml.Columns{Num: 1, Sep: false, EqualWidth: false}
	sec.TypeVal = sectionBreakKindToWML(c.Break)
	if sec.TypeVal == "" {
		sec.TypeVal = "nextPage"
	}
	sec.PageNumFmt = strings.TrimSpace(string(c.PageNumberFormat))
	if c.PageNumberStart != nil {
		sec.PageNumStartSet = true
		sec.PageNumStart = *c.PageNumberStart
	}
	return sec
}

func sectionBreakKindToWML(k SectionBreakKind) string {
	switch k {
	case SectionBreakNextPage:
		return "nextPage"
	case SectionBreakContinuous:
		return "continuous"
	case SectionBreakNextColumn:
		return "nextColumn"
	case SectionBreakEvenPage:
		return "evenPage"
	case SectionBreakOddPage:
		return "oddPage"
	case SectionBreakUnset:
		return ""
	default:
		return ""
	}
}

func sectionBreakKindFromWML(s string) SectionBreakKind {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "nextpage":
		return SectionBreakNextPage
	case "continuous":
		return SectionBreakContinuous
	case "nextcolumn":
		return SectionBreakNextColumn
	case "evenpage":
		return SectionBreakEvenPage
	case "oddpage":
		return SectionBreakOddPage
	default:
		return SectionBreakUnset
	}
}

func marshalSectPrBytes(sec wml.Section) []byte {
	var b bytes.Buffer
	b.WriteString(`<w:sectPr xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">`)
	if v := strings.TrimSpace(sec.TypeVal); v != "" {
		b.WriteString(`<w:type w:val="` + escapeAttr(v) + `"/>`)
	}
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
	if sec.PageNumStartSet || strings.TrimSpace(sec.PageNumFmt) != "" {
		b.WriteString(`<w:pgNumType`)
		if sec.PageNumStartSet {
			b.WriteString(` w:start="` + strconv.Itoa(sec.PageNumStart) + `"`)
		}
		if f := strings.TrimSpace(sec.PageNumFmt); f != "" {
			b.WriteString(` w:fmt="` + escapeAttr(f) + `"`)
		}
		b.WriteString(`/>`)
	}
	b.WriteString(`</w:sectPr>`)
	return b.Bytes()
}
