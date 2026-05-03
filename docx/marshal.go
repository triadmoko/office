package docx

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/triadmoko/office/internal/wml"
)

const (
	nsW = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	nsR = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
)

func escapeCharData(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// MarshalDocumentXML serializes the main document body to WordprocessingML.
func MarshalDocumentXML(doc *wml.Document) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("docx: nil document model")
	}
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:document xmlns:w="` + nsW + `" xmlns:r="` + nsR + `">`)
	b.WriteString(`<w:body>`)
	for _, bl := range doc.Body.Blocks {
		if err := marshalBodyBlock(&b, bl); err != nil {
			return nil, err
		}
	}
	if len(doc.Body.SectPr) > 0 {
		b.Write(doc.Body.SectPr)
	} else {
		b.WriteString(`<w:sectPr><w:pgSz w:w="` + strconv.FormatInt(wml.PageA4W, 10) + `" w:h="` + strconv.FormatInt(wml.PageA4H, 10) + `"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440" w:header="708" w:footer="708" w:gutter="0"/></w:sectPr>`)
	}
	b.WriteString(`</w:body></w:document>`)
	return b.Bytes(), nil
}

func marshalBodyBlock(b *bytes.Buffer, bl wml.BodyBlock) error {
	switch {
	case bl.Para != nil:
		marshalParagraph(b, bl.Para)
	case bl.Table != nil:
		marshalTable(b, bl.Table)
	case len(bl.Unknown) > 0:
		b.Write(bl.Unknown)
	}
	return nil
}

func marshalParagraph(b *bytes.Buffer, p *wml.Paragraph) {
	if p == nil {
		return
	}
	b.WriteString(`<w:p>`)
	if len(p.PPr.RawPPr) > 0 {
		b.Write(p.PPr.RawPPr)
	} else {
		marshalPPrStructured(b, p.PPr)
	}
	for _, r := range p.Runs {
		marshalRun(b, r)
	}
	if len(p.Unknown) > 0 {
		b.Write(p.Unknown)
	}
	b.WriteString(`</w:p>`)
}

func marshalPPrStructured(b *bytes.Buffer, pr wml.ParagraphProps) {
	if pPrIsEmpty(pr) {
		return
	}
	b.WriteString(`<w:pPr>`)
	if pr.Alignment != wml.AlignLeft {
		b.WriteString(`<w:jc w:val="` + jcString(pr.Alignment) + `"/>`)
	}
	if pr.Indent.Left != 0 || pr.Indent.Right != 0 || pr.Indent.FirstLine != 0 || pr.Indent.Hanging != 0 {
		b.WriteString(`<w:ind`)
		writeI64Attr(b, "w:left", pr.Indent.Left)
		writeI64Attr(b, "w:right", pr.Indent.Right)
		writeI64Attr(b, "w:firstLine", pr.Indent.FirstLine)
		writeI64Attr(b, "w:hanging", pr.Indent.Hanging)
		b.WriteString(`/>`)
	}
	if pr.Spacing.Before != 0 || pr.Spacing.After != 0 || pr.Spacing.Line != 0 || pr.Spacing.LineRule != wml.LineRuleUnset {
		b.WriteString(`<w:spacing`)
		writeI64Attr(b, "w:before", pr.Spacing.Before)
		writeI64Attr(b, "w:after", pr.Spacing.After)
		writeI64Attr(b, "w:line", pr.Spacing.Line)
		if pr.Spacing.LineRule != wml.LineRuleUnset {
			b.WriteString(` w:lineRule="` + lineRuleString(pr.Spacing.LineRule) + `"`)
		}
		b.WriteString(`/>`)
	}
	if pr.StyleID != "" {
		b.WriteString(`<w:pStyle w:val="` + escapeAttr(pr.StyleID) + `"/>`)
	}
	if pr.Numbering != nil {
		// Use explicit close tags: some xml.Decoder paths reject sibling self-closing tags in numPr.
		b.WriteString(`<w:numPr><w:ilvl w:val="` + strconv.Itoa(pr.Numbering.Ilvl) + `"></w:ilvl><w:numId w:val="` + strconv.Itoa(pr.Numbering.NumID) + `"></w:numId></w:numPr>`)
	}
	if len(pr.SectPr) > 0 {
		b.Write(pr.SectPr)
	}
	b.WriteString(`</w:pPr>`)
}

func pPrIsEmpty(pr wml.ParagraphProps) bool {
	if pr.Numbering != nil {
		return false
	}
	if pr.Alignment != wml.AlignLeft || pr.StyleID != "" || len(pr.SectPr) > 0 {
		return false
	}
	if pr.Indent != (wml.Indent{}) || pr.Spacing != (wml.Spacing{}) {
		return false
	}
	return true
}

func writeI64Attr(b *bytes.Buffer, name string, v int64) {
	if v == 0 {
		return
	}
	b.WriteString(` ` + name + `="` + strconv.FormatInt(v, 10) + `"`)
}

func jcString(a wml.Alignment) string {
	switch a {
	case wml.AlignRight:
		return "right"
	case wml.AlignCenter:
		return "center"
	case wml.AlignJustify:
		return "both"
	case wml.AlignDistribute:
		return "distribute"
	case wml.AlignStart:
		return "start"
	case wml.AlignEnd:
		return "end"
	default:
		return "left"
	}
}

func lineRuleString(r wml.LineRule) string {
	switch r {
	case wml.LineRuleAuto:
		return "auto"
	case wml.LineRuleExact:
		return "exact"
	case wml.LineRuleAtLeast:
		return "atLeast"
	default:
		return "auto"
	}
}

func marshalRun(b *bytes.Buffer, r *wml.Run) {
	if r == nil {
		return
	}
	b.WriteString(`<w:r>`)
	if len(r.RPr.RawRPr) > 0 {
		b.Write(r.RPr.RawRPr)
	} else if !emptyRunProps(r.RPr) {
		b.WriteString(`<w:rPr>`)
		if r.RPr.Bold {
			b.WriteString(`<w:b/>`)
		}
		if r.RPr.Italic {
			b.WriteString(`<w:i/>`)
		}
		if r.RPr.Underline {
			b.WriteString(`<w:u w:val="single"/>`)
		}
		if r.RPr.Strike {
			b.WriteString(`<w:strike/>`)
		}
		if r.RPr.VertAlign == wml.VertAlignSuperscript {
			b.WriteString(`<w:vertAlign w:val="superscript"/>`)
		} else if r.RPr.VertAlign == wml.VertAlignSubscript {
			b.WriteString(`<w:vertAlign w:val="subscript"/>`)
		}
		if r.RPr.FontSizeHalf != 0 {
			b.WriteString(`<w:sz w:val="` + strconv.Itoa(r.RPr.FontSizeHalf) + `"/>`)
		}
		if r.RPr.Color != "" {
			b.WriteString(`<w:color w:val="` + escapeAttr(r.RPr.Color) + `"/>`)
		}
		if r.RPr.FontName != "" {
			b.WriteString(`<w:rFonts w:ascii="` + escapeAttr(r.RPr.FontName) + `" w:hAnsi="` + escapeAttr(r.RPr.FontName) + `"/>`)
		}
		b.WriteString(`</w:rPr>`)
	}
	for _, part := range r.Parts {
		switch {
		case part.Tab:
			b.WriteString(`<w:tab/>`)
		case part.Br:
			b.WriteString(`<w:br/>`)
		case part.SoftHyphen:
			b.WriteString(`<w:softHyphen/>`)
		case part.Text != "":
			b.WriteString(`<w:t xml:space="preserve">` + escapeCharData(part.Text) + `</w:t>`)
		case len(part.Unknown) > 0:
			b.Write(part.Unknown)
		}
	}
	b.WriteString(`</w:r>`)
}

func emptyRunProps(r wml.RunProps) bool {
	return !r.Bold && !r.Italic && !r.Underline && !r.Strike && r.VertAlign == wml.VertAlignBaseline &&
		r.FontSizeHalf == 0 && r.Color == "" && r.FontName == ""
}

func marshalTable(b *bytes.Buffer, t *wml.Table) {
	if t == nil {
		return
	}
	b.WriteString(`<w:tbl>`)
	if len(t.Props.Raw) > 0 {
		b.Write(t.Props.Raw)
	} else if t.Props.Width.Value != 0 || t.Props.Width.Kind != wml.WidthAuto {
		b.WriteString(`<w:tblPr><w:tblW`)
		writeI64Attr(b, "w:w", t.Props.Width.Value)
		b.WriteString(` w:type="` + widthKind(t.Props.Width.Kind) + `"/></w:tblPr>`)
	}
	if len(t.Unknown) > 0 {
		b.Write(t.Unknown)
	}
	for _, row := range t.Rows {
		if row == nil {
			continue
		}
		b.WriteString(`<w:tr>`)
		for _, cell := range row.Cells {
			if cell == nil {
				continue
			}
			b.WriteString(`<w:tc>`)
			if cell.TcPr.GridSpan > 0 || cell.TcPr.VMerge != wml.VMergeNone || cell.TcPr.Width.Value != 0 ||
				cell.TcPr.Borders != nil || cell.TcPr.Shading != nil || len(cell.TcPr.RawTcPr) > 0 {
				if len(cell.TcPr.RawTcPr) > 0 {
					b.Write(cell.TcPr.RawTcPr)
				} else {
					marshalTcPr(b, cell.TcPr)
				}
			}
			for _, cb := range cell.Blocks {
				_ = marshalBodyBlock(b, cb)
			}
			b.WriteString(`</w:tc>`)
		}
		b.WriteString(`</w:tr>`)
	}
	b.WriteString(`</w:tbl>`)
}

func widthKind(k wml.WidthKind) string {
	switch k {
	case wml.WidthDxa:
		return "dxa"
	case wml.WidthPct:
		return "pct"
	default:
		return "auto"
	}
}

func marshalTcPr(b *bytes.Buffer, pr wml.TableCellProps) {
	b.WriteString(`<w:tcPr>`)
	if pr.GridSpan > 1 {
		b.WriteString(`<w:gridSpan w:val="` + strconv.Itoa(pr.GridSpan) + `"/>`)
	}
	if pr.VMerge != wml.VMergeNone {
		v := "restart"
		if pr.VMerge == wml.VMergeContinue {
			v = "continue"
		}
		b.WriteString(`<w:vMerge w:val="` + v + `"/>`)
	}
	if pr.Width.Value != 0 || pr.Width.Kind != wml.WidthAuto {
		b.WriteString(`<w:tcW`)
		writeI64Attr(b, "w:w", pr.Width.Value)
		b.WriteString(` w:type="` + widthKind(pr.Width.Kind) + `"/>`)
	}
	if pr.Borders != nil {
		b.WriteString(`<w:tcBorders>`)
		writeCellBorder(b, "w:top", pr.Borders.Top)
		writeCellBorder(b, "w:left", pr.Borders.Left)
		writeCellBorder(b, "w:bottom", pr.Borders.Bottom)
		writeCellBorder(b, "w:right", pr.Borders.Right)
		writeCellBorder(b, "w:insideH", pr.Borders.InsideH)
		writeCellBorder(b, "w:insideV", pr.Borders.InsideV)
		b.WriteString(`</w:tcBorders>`)
	}
	if pr.Shading != nil {
		b.WriteString(`<w:shd w:val="` + escapeAttr(pr.Shading.Val) + `" w:fill="` + escapeAttr(pr.Shading.Fill) + `" w:color="` + escapeAttr(pr.Shading.Color) + `"/>`)
	}
	b.WriteString(`</w:tcPr>`)
}

func writeCellBorder(b *bytes.Buffer, tag string, bd *wml.BorderDef) {
	if bd == nil {
		return
	}
	b.WriteString(`<` + tag + ` w:val="` + escapeAttr(bd.Val) + `" w:sz="` + strconv.Itoa(bd.Size) + `" w:space="` + strconv.Itoa(bd.Space) + `" w:color="` + escapeAttr(bd.Color) + `"/>`)
}
