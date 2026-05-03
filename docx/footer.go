package docx

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/triadmoko/office/internal/ooxml"
	"github.com/triadmoko/office/internal/wml"
)

// Placeholder teks untuk [SetFooterPageNumberTemplate] / [SetHeaderPageNumberTemplate] (case-sensitive).
const (
	FooterPlaceholderPage     = "{{PAGE}}"
	FooterPlaceholderNumPages = "{{NUMPAGES}}"
)

// HeaderPlaceholderPage matches [FooterPlaceholderPage] (token yang sama untuk header).
const HeaderPlaceholderPage = FooterPlaceholderPage

// HeaderPlaceholderNumPages matches [FooterPlaceholderNumPages].
const HeaderPlaceholderNumPages = FooterPlaceholderNumPages

type footerSegKind int

const (
	footerSegText footerSegKind = iota
	footerSegPage
	footerSegNumPages
)

type footerSeg struct {
	kind footerSegKind
	text string
}

func parseFooterLayout(s string) []footerSeg {
	var out []footerSeg
	for len(s) > 0 {
		i := strings.Index(s, "{{")
		if i < 0 {
			if s != "" {
				out = append(out, footerSeg{kind: footerSegText, text: s})
			}
			break
		}
		if i > 0 {
			out = append(out, footerSeg{kind: footerSegText, text: s[:i]})
		}
		s = s[i:]
		if strings.HasPrefix(s, FooterPlaceholderPage) {
			out = append(out, footerSeg{kind: footerSegPage})
			s = s[len(FooterPlaceholderPage):]
			continue
		}
		if strings.HasPrefix(s, FooterPlaceholderNumPages) {
			out = append(out, footerSeg{kind: footerSegNumPages})
			s = s[len(FooterPlaceholderNumPages):]
			continue
		}
		out = append(out, footerSeg{kind: footerSegText, text: "{{"})
		s = s[2:]
	}
	return out
}

// newDocumentRels builds /word/document.xml.rels for a new package; returns relationship ids used in w:sectPr.
func newDocumentRels(withNumbering, withFooter, withHeader bool) (*ooxml.Relationships, string, string) {
	var footerRID, headerRID string
	rels := []ooxml.Relationship{
		{ID: "rId1", Type: relTypeStyles, Target: "styles.xml"},
	}
	next := 2
	alloc := func() string {
		s := "rId" + strconv.Itoa(next)
		next++
		return s
	}
	if withNumbering {
		rels = append(rels, ooxml.Relationship{ID: alloc(), Type: relTypeNumbering, Target: "numbering.xml"})
	}
	if withFooter {
		footerRID = alloc()
		rels = append(rels, ooxml.Relationship{ID: footerRID, Type: relTypeFooter, Target: "footer1.xml"})
	}
	if withHeader {
		headerRID = alloc()
		rels = append(rels, ooxml.Relationship{ID: headerRID, Type: relTypeHeader, Target: "header1.xml"})
	}
	return &ooxml.Relationships{Relationship: rels}, footerRID, headerRID
}

// marshalFooterPageXML builds footer1.xml: teks bebas + bidang PAGE / NUMPAGES sesuai layout.
// layout kosong (setelah trim) memakai default "Hal. {{PAGE}}".
func marshalFooterPageXML(layout string) []byte {
	if strings.TrimSpace(layout) == "" {
		layout = "Hal. " + FooterPlaceholderPage
	}
	segs := parseFooterLayout(layout)
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:ftr xmlns:w="` + nsW + `" xmlns:r="` + nsR + `">`)
	b.WriteString(`<w:p><w:pPr><w:jc w:val="right"/></w:pPr>`)
	writeFooterHeaderRuns(&b, segs)
	b.WriteString(`</w:p></w:ftr>`)
	return []byte(b.String())
}

// marshalHeaderPageXML builds header1.xml (default paragraph rata tengah).
func marshalHeaderPageXML(layout string) []byte {
	if strings.TrimSpace(layout) == "" {
		layout = "Hal. " + HeaderPlaceholderPage
	}
	segs := parseFooterLayout(layout)
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:hdr xmlns:w="` + nsW + `" xmlns:r="` + nsR + `">`)
	b.WriteString(`<w:p><w:pPr><w:jc w:val="center"/></w:pPr>`)
	writeFooterHeaderRuns(&b, segs)
	b.WriteString(`</w:p></w:hdr>`)
	return []byte(b.String())
}

func writeFooterHeaderRuns(b *strings.Builder, segs []footerSeg) {
	for _, seg := range segs {
		switch seg.kind {
		case footerSegText:
			if seg.text == "" {
				continue
			}
			b.WriteString(`<w:r><w:t xml:space="preserve">`)
			b.WriteString(escapeCharData(seg.text))
			b.WriteString(`</w:t></w:r>`)
		case footerSegPage:
			b.WriteString(complexFieldPAGE())
		case footerSegNumPages:
			b.WriteString(complexFieldNUMPAGES())
		}
	}
}

func complexFieldPAGE() string {
	return `<w:r><w:fldChar w:fldCharType="begin"/></w:r>` +
		`<w:r><w:instrText xml:space="preserve"> PAGE </w:instrText></w:r>` +
		`<w:r><w:fldChar w:fldCharType="separate"/></w:r>` +
		`<w:r><w:t>1</w:t></w:r>` +
		`<w:r><w:fldChar w:fldCharType="end"/></w:r>`
}

func complexFieldNUMPAGES() string {
	return `<w:r><w:fldChar w:fldCharType="begin"/></w:r>` +
		`<w:r><w:instrText xml:space="preserve"> NUMPAGES </w:instrText></w:r>` +
		`<w:r><w:fldChar w:fldCharType="separate"/></w:r>` +
		`<w:r><w:t>1</w:t></w:r>` +
		`<w:r><w:fldChar w:fldCharType="end"/></w:r>`
}

// injectSectPrHeaderFooterRefs inserts w:headerReference and w:footerReference before </w:sectPr> (header first).
func injectSectPrHeaderFooterRefs(sect []byte, headerRID, footerRID string) []byte {
	if headerRID == "" && footerRID == "" {
		return sect
	}
	needle := []byte("</w:sectPr>")
	idx := bytes.LastIndex(sect, needle)
	if idx < 0 {
		return sect
	}
	var ins strings.Builder
	if headerRID != "" {
		ins.WriteString(`<w:headerReference w:type="default" r:id="`)
		ins.WriteString(escapeAttr(headerRID))
		ins.WriteString(`"/>`)
	}
	if footerRID != "" {
		ins.WriteString(`<w:footerReference w:type="default" r:id="`)
		ins.WriteString(escapeAttr(footerRID))
		ins.WriteString(`"/>`)
	}
	out := make([]byte, 0, len(sect)+ins.Len())
	out = append(out, sect[:idx]...)
	out = append(out, ins.String()...)
	out = append(out, sect[idx:]...)
	return out
}

func defaultBodyClosingSectPr(headerRID, footerRID string) string {
	var b bytes.Buffer
	b.WriteString(`<w:sectPr>`)
	if headerRID != "" {
		b.WriteString(`<w:headerReference w:type="default" r:id="`)
		b.WriteString(escapeAttr(headerRID))
		b.WriteString(`"/>`)
	}
	if footerRID != "" {
		b.WriteString(`<w:footerReference w:type="default" r:id="`)
		b.WriteString(escapeAttr(footerRID))
		b.WriteString(`"/>`)
	}
	b.WriteString(`<w:pgSz w:w="`)
	b.WriteString(strconv.FormatInt(wml.PageA4W, 10))
	b.WriteString(`" w:h="`)
	b.WriteString(strconv.FormatInt(wml.PageA4H, 10))
	b.WriteString(`"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="708" w:footer="708" w:gutter="0"/></w:sectPr>`)
	return b.String()
}
