package wml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseDocument parses word/document.xml from r.
func ParseDocument(r io.Reader) (*Document, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("wml: no document element")
			}
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if isWML(se.Name.Space, se.Name.Local) && se.Name.Local == "document" {
			return parseDocumentElement(dec)
		}
	}
}

func parseDocumentElement(dec *xml.Decoder) (*Document, error) {
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "body" {
				body, err := parseBody(dec)
				if err != nil {
					return nil, err
				}
				return &Document{Body: body}, nil
			}
			if err := skipSubtree(dec, t); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "document" {
				return &Document{Body: Body{}}, nil
			}
		}
	}
}

func parseBody(dec *xml.Decoder) (Body, error) {
	var b Body
	for {
		tok, err := dec.Token()
		if err != nil {
			return Body{}, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if !isWML(t.Name.Space, t.Name.Local) {
				raw, err := captureSubtree(dec, t)
				if err != nil {
					return Body{}, err
				}
				b.Blocks = append(b.Blocks, BodyBlock{Unknown: raw})
				continue
			}
			switch t.Name.Local {
			case "p":
				p, err := parseParagraph(dec, t)
				if err != nil {
					return Body{}, err
				}
				b.Blocks = append(b.Blocks, BodyBlock{Para: p})
			case "tbl":
				tbl, err := parseTable(dec, t)
				if err != nil {
					return Body{}, err
				}
				b.Blocks = append(b.Blocks, BodyBlock{Table: tbl})
			case "sectPr":
				sub, err := captureSubtree(dec, t)
				if err != nil {
					return Body{}, err
				}
				b.SectPr = sub
			default:
				raw, err := captureSubtree(dec, t)
				if err != nil {
					return Body{}, err
				}
				b.Blocks = append(b.Blocks, BodyBlock{Unknown: raw})
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "body" {
				return b, nil
			}
		}
	}
}

func parseParagraph(dec *xml.Decoder, start xml.StartElement) (*Paragraph, error) {
	p := &Paragraph{}
	var unk bytes.Buffer

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if !isWML(t.Name.Space, t.Name.Local) {
				raw, err := captureSubtree(dec, t)
				if err != nil {
					return nil, err
				}
				if _, err := unk.Write(raw); err != nil {
					return nil, err
				}
				continue
			}
			switch t.Name.Local {
			case "r":
				run, err := parseRun(dec, t)
				if err != nil {
					return nil, err
				}
				p.Runs = append(p.Runs, run)
			case "pPr":
				if err := parsePPr(dec, t, p); err != nil {
					return nil, err
				}
			default:
				raw, err := captureSubtree(dec, t)
				if err != nil {
					return nil, err
				}
				if _, err := unk.Write(raw); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "p" {
				p.Unknown = unk.Bytes()
				return p, nil
			}
		case xml.CharData:
			if len(bytes.TrimSpace([]byte(t))) > 0 {
				if _, err := unk.Write([]byte(t)); err != nil {
					return nil, err
				}
			}
		case xml.Comment:
			if _, err := unk.Write([]byte(fmt.Sprintf("<!--%s-->", string(t)))); err != nil {
				return nil, err
			}
		case xml.ProcInst:
			// ignore
		}
	}
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func parsePPr(dec *xml.Decoder, start xml.StartElement, p *Paragraph) error {
	// Walk w:pPr children without xml.Encoder round-tripping: EncodeToken breaks
	// namespace/prefix output for WordprocessingML on Go 1.23, corrupting RawPPr.
	_ = start
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if !isWML(t.Name.Space, t.Name.Local) {
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
				continue
			}
			switch t.Name.Local {
			case "jc":
				p.PPr.Alignment = parseJc(valAttr(t.Attr))
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "ind":
				parseInd(t, &p.PPr.Indent)
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "spacing":
				parseSpacing(t, &p.PPr.Spacing)
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "pStyle":
				p.PPr.StyleID = valAttr(t.Attr)
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "numPr":
				n, err := parseNumPrTokens(dec, t)
				if err != nil {
					return err
				}
				p.PPr.Numbering = n
			case "sectPr":
				sub, err := captureSubtree(dec, t)
				if err != nil {
					return err
				}
				p.PPr.SectPr = append([]byte(nil), sub...)
			default:
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "pPr" {
				return nil
			}
		case xml.CharData:
			// ignore whitespace in pPr
		}
	}
}

func parseNumPrTokens(dec *xml.Decoder, start xml.StartElement) (*NumPr, error) {
	n := &NumPr{}
	var seenID bool
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if isWML(t.Name.Space, t.Name.Local) {
				switch t.Name.Local {
				case "ilvl":
					n.Ilvl = intAttr(valAttr(t.Attr))
				case "numId":
					n.NumID = intAttr(valAttr(t.Attr))
					seenID = true
				}
			}
		case xml.EndElement:
			depth--
		}
	}
	if !seenID {
		return nil, nil
	}
	return n, nil
}

func intAttr(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

func parseJc(val string) Alignment {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "right":
		return AlignRight
	case "center":
		return AlignCenter
	case "both", "justify":
		return AlignJustify
	case "distribute":
		return AlignDistribute
	case "start":
		return AlignStart
	case "end":
		return AlignEnd
	default:
		return AlignLeft
	}
}

func parseInd(se xml.StartElement, ind *Indent) {
	for _, a := range se.Attr {
		if a.Name.Local == "left" {
			ind.Left = twipAttr(a.Value)
		}
		if a.Name.Local == "right" {
			ind.Right = twipAttr(a.Value)
		}
		if a.Name.Local == "firstLine" {
			ind.FirstLine = twipAttr(a.Value)
		}
		if a.Name.Local == "hanging" {
			ind.Hanging = twipAttr(a.Value)
		}
	}
}

func parseSpacing(se xml.StartElement, sp *Spacing) {
	for _, a := range se.Attr {
		switch a.Name.Local {
		case "before":
			sp.Before = twipAttr(a.Value)
		case "after":
			sp.After = twipAttr(a.Value)
		case "line":
			sp.Line = twipAttr(a.Value)
		case "lineRule":
			switch strings.ToLower(strings.TrimSpace(a.Value)) {
			case "auto":
				sp.LineRule = LineRuleAuto
			case "exact":
				sp.LineRule = LineRuleExact
			case "atLeast":
				sp.LineRule = LineRuleAtLeast
			}
		}
	}
}

func twipAttr(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func valAttr(attrs []xml.Attr) string {
	for _, a := range attrs {
		if a.Name.Local == "val" {
			return a.Value
		}
	}
	return ""
}

func parseRun(dec *xml.Decoder, start xml.StartElement) (*Run, error) {
	run := &Run{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if !isWML(t.Name.Space, t.Name.Local) {
				raw, err := captureSubtree(dec, t)
				if err != nil {
					return nil, err
				}
				run.Parts = append(run.Parts, RunPart{Unknown: raw})
				continue
			}
			switch t.Name.Local {
			case "rPr":
				if err := parseRPr(dec, t, run); err != nil {
					return nil, err
				}
			case "t":
				preserve := hasSpacePreserve(t.Attr)
				text, err := readTextUntilEnd(dec, "t")
				if err != nil {
					return nil, err
				}
				s := string(text)
				if !preserve {
					s = strings.TrimSpace(s) // Word often uses preserve incorrectly; spec says use attr
				}
				run.Parts = append(run.Parts, RunPart{Text: s})
			case "tab":
				run.Parts = append(run.Parts, RunPart{Tab: true})
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			case "br":
				run.Parts = append(run.Parts, RunPart{Br: true})
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			case "softHyphen":
				run.Parts = append(run.Parts, RunPart{SoftHyphen: true})
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			case "smartTag":
				if err := flattenSmartTag(dec, t, run); err != nil {
					return nil, err
				}
			default:
				raw, err := captureSubtree(dec, t)
				if err != nil {
					return nil, err
				}
				run.Parts = append(run.Parts, RunPart{Unknown: raw})
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "r" {
				run.RebuildText()
				var u bytes.Buffer
				for _, ch := range run.Parts {
					if len(ch.Unknown) > 0 {
						u.Write(ch.Unknown)
					}
				}
				run.Unknown = u.Bytes()
				return run, nil
			}
		case xml.CharData:
			if len(bytes.TrimSpace([]byte(t))) > 0 {
				run.Parts = append(run.Parts, RunPart{Unknown: append([]byte(nil), t...)})
			}
		}
	}
}

func flattenSmartTag(dec *xml.Decoder, start xml.StartElement, run *Run) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "r" {
				inner, err := parseRun(dec, t)
				if err != nil {
					return err
				}
				run.Parts = append(run.Parts, inner.Parts...)
			} else if err := skipSubtree(dec, t); err != nil {
				return err
			}
		case xml.EndElement:
			if t.Name.Local == start.Name.Local && t.Name.Space == start.Name.Space {
				return nil
			}
		}
	}
}

// RebuildText sets Text from Parts (tab, br, soft hyphen, text segments).
func (r *Run) RebuildText() {
	var b strings.Builder
	for _, ch := range r.Parts {
		switch {
		case ch.Tab:
			b.WriteByte('\t')
		case ch.Br:
			b.WriteByte('\n')
		case ch.SoftHyphen:
			b.WriteRune('\u00ad')
		case ch.Text != "":
			b.WriteString(ch.Text)
		}
	}
	r.Text = b.String()
}

func hasSpacePreserve(attrs []xml.Attr) bool {
	for _, a := range attrs {
		if isXMLSpace(a.Name) && a.Value == "preserve" {
			return true
		}
	}
	return false
}

func readTextUntilEnd(dec *xml.Decoder, local string) ([]byte, error) {
	var buf bytes.Buffer
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.CharData:
			buf.Write([]byte(t))
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == local {
				return buf.Bytes(), nil
			}
		}
	}
}

func parseRPr(dec *xml.Decoder, start xml.StartElement, run *Run) error {
	_ = start
	run.RPr.RawRPr = nil
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if !isWML(t.Name.Space, t.Name.Local) {
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
				continue
			}
			switch t.Name.Local {
			case "b":
				if !isOff(valAttr(t.Attr)) {
					run.RPr.Bold = true
				}
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "i":
				if !isOff(valAttr(t.Attr)) {
					run.RPr.Italic = true
				}
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "u":
				if valAttr(t.Attr) != "" && !strings.EqualFold(valAttr(t.Attr), "none") {
					run.RPr.Underline = true
				}
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "strike", "dstrike":
				if !isOff(valAttr(t.Attr)) {
					run.RPr.Strike = true
				}
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "sz", "szCs":
				if v := valAttr(t.Attr); v != "" {
					if n, err := strconv.Atoi(v); err == nil && n > 0 {
						if run.RPr.FontSizeHalf == 0 || t.Name.Local == "sz" {
							run.RPr.FontSizeHalf = n
						}
					}
				}
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "color":
				run.RPr.Color = strings.TrimSpace(valAttr(t.Attr))
				if run.RPr.Color == "auto" {
					run.RPr.Color = ""
				}
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "rFonts":
				if a := fontAttr(t.Attr, "ascii"); a != "" {
					run.RPr.FontName = a
				} else if h := fontAttr(t.Attr, "hAnsi"); h != "" {
					run.RPr.FontName = h
				}
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			case "vertAlign":
				switch strings.ToLower(valAttr(t.Attr)) {
				case "superscript":
					run.RPr.VertAlign = VertAlignSuperscript
				case "subscript":
					run.RPr.VertAlign = VertAlignSubscript
				default:
					run.RPr.VertAlign = VertAlignBaseline
				}
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			default:
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "rPr" {
				return nil
			}
		case xml.CharData:
			// ignore
		}
	}
}

func fontAttr(attrs []xml.Attr, key string) string {
	for _, a := range attrs {
		if a.Name.Local == key && a.Value != "" {
			return a.Value
		}
	}
	return ""
}

func isOff(val string) bool {
	return strings.EqualFold(strings.TrimSpace(val), "0") || strings.EqualFold(strings.TrimSpace(val), "false")
}

func parseTable(dec *xml.Decoder, start xml.StartElement) (*Table, error) {
	tbl := &Table{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if !isWML(t.Name.Space, t.Name.Local) {
				raw, err := captureSubtree(dec, t)
				if err != nil {
					return nil, err
				}
				tbl.Unknown = append(tbl.Unknown, raw...)
				continue
			}
			switch t.Name.Local {
			case "tr":
				row, err := parseTableRow(dec, t)
				if err != nil {
					return nil, err
				}
				tbl.Rows = append(tbl.Rows, row)
			case "tblPr":
				if err := parseTblPr(dec, t, tbl); err != nil {
					return nil, err
				}
			case "tblGrid":
				sub, err := captureSubtree(dec, t)
				if err != nil {
					return nil, err
				}
				tbl.Unknown = append(tbl.Unknown, sub...)
			default:
				sub, err := captureSubtree(dec, t)
				if err != nil {
					return nil, err
				}
				tbl.Unknown = append(tbl.Unknown, sub...)
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "tbl" {
				return tbl, nil
			}
		}
	}
}

func parseTblPr(dec *xml.Decoder, start xml.StartElement, tbl *Table) error {
	_ = start
	tbl.Props.Raw = nil
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "tblW" {
				parseTcW(t, &tbl.Props.Width)
				if err := skipSubtree(dec, t); err != nil {
					return err
				}
				continue
			}
			if err := skipSubtree(dec, t); err != nil {
				return err
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "tblPr" {
				return nil
			}
		case xml.CharData:
			// ignore
		}
	}
}

func parseTableRow(dec *xml.Decoder, start xml.StartElement) (*TableRow, error) {
	row := &TableRow{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "tc" {
				cell, err := parseTableCell(dec, t)
				if err != nil {
					return nil, err
				}
				row.Cells = append(row.Cells, cell)
			} else if isWML(t.Name.Space, t.Name.Local) {
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			} else if err := skipSubtree(dec, t); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "tr" {
				return row, nil
			}
		}
	}
}

func parseTableCell(dec *xml.Decoder, start xml.StartElement) (*TableCell, error) {
	cell := &TableCell{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if !isWML(t.Name.Space, t.Name.Local) {
				raw, err := captureSubtree(dec, t)
				if err != nil {
					return nil, err
				}
				cell.Blocks = append(cell.Blocks, BodyBlock{Unknown: raw})
				continue
			}
			switch t.Name.Local {
			case "p":
				p, err := parseParagraph(dec, t)
				if err != nil {
					return nil, err
				}
				cell.Blocks = append(cell.Blocks, BodyBlock{Para: p})
			case "tbl":
				tbl, err := parseTable(dec, t)
				if err != nil {
					return nil, err
				}
				cell.Blocks = append(cell.Blocks, BodyBlock{Table: tbl})
			case "tcPr":
				if err := parseTcPr(dec, t, cell); err != nil {
					return nil, err
				}
			default:
				raw, err := captureSubtree(dec, t)
				if err != nil {
					return nil, err
				}
				cell.Blocks = append(cell.Blocks, BodyBlock{Unknown: raw})
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "tc" {
				return cell, nil
			}
		}
	}
}

func parseTcPr(dec *xml.Decoder, start xml.StartElement, cell *TableCell) error {
	_ = start
	cell.TcPr.RawTcPr = nil
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) {
				switch t.Name.Local {
				case "gridSpan":
					cell.TcPr.GridSpan = intAttr(valAttr(t.Attr))
					if err := skipSubtree(dec, t); err != nil {
						return err
					}
					continue
				case "vMerge":
					switch strings.ToLower(valAttr(t.Attr)) {
					case "restart":
						cell.TcPr.VMerge = VMergeRestart
					case "continue":
						cell.TcPr.VMerge = VMergeContinue
					default:
						if valAttr(t.Attr) == "" {
							cell.TcPr.VMerge = VMergeRestart
						}
					}
					if err := skipSubtree(dec, t); err != nil {
						return err
					}
					continue
				case "tcW":
					parseTcW(t, &cell.TcPr.Width)
					if err := skipSubtree(dec, t); err != nil {
						return err
					}
					continue
				case "tcBorders":
					b, err := parseTcBorders(dec, t)
					if err != nil {
						return err
					}
					cell.TcPr.Borders = b
					continue
				case "shd":
					cell.TcPr.Shading = parseShd(t)
					if err := skipSubtree(dec, t); err != nil {
						return err
					}
					continue
				default:
					if err := skipSubtree(dec, t); err != nil {
						return err
					}
					continue
				}
			}
			if err := skipSubtree(dec, t); err != nil {
				return err
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "tcPr" {
				return nil
			}
		case xml.CharData:
			// ignore
		}
	}
}

func parseTcW(se xml.StartElement, w *TableWidth) {
	w.Value = twipAttr(attrLocal(se.Attr, "w"))
	typeStr := attrLocal(se.Attr, "type")
	switch strings.ToLower(strings.TrimSpace(typeStr)) {
	case "dxa":
		w.Kind = WidthDxa
	case "pct":
		w.Kind = WidthPct
	case "auto", "":
		w.Kind = WidthAuto
	default:
		w.Kind = WidthAuto
	}
}

func attrLocal(attrs []xml.Attr, local string) string {
	for _, a := range attrs {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

func parseShd(se xml.StartElement) *Shading {
	sh := &Shading{}
	for _, a := range se.Attr {
		switch a.Name.Local {
		case "fill":
			sh.Fill = a.Value
		case "color":
			sh.Color = a.Value
		case "val":
			sh.Val = a.Value
		}
	}
	return sh
}

func parseTcBorders(dec *xml.Decoder, start xml.StartElement) (*TcBorders, error) {
	b := &TcBorders{}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if !isWML(t.Name.Space, t.Name.Local) {
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
				continue
			}
			switch t.Name.Local {
			case "top":
				b.Top = parseBorderDef(t)
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			case "left":
				b.Left = parseBorderDef(t)
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			case "bottom":
				b.Bottom = parseBorderDef(t)
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			case "right":
				b.Right = parseBorderDef(t)
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			case "insideH":
				b.InsideH = parseBorderDef(t)
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			case "insideV":
				b.InsideV = parseBorderDef(t)
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			default:
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "tcBorders" {
				return b, nil
			}
		}
	}
}

func parseBorderDef(se xml.StartElement) *BorderDef {
	b := &BorderDef{Val: valAttr(se.Attr)}
	for _, a := range se.Attr {
		switch a.Name.Local {
		case "color":
			b.Color = a.Value
		case "sz":
			b.Size = intAttr(a.Value)
		case "space":
			b.Space = intAttr(a.Value)
		}
	}
	return b
}

func captureSubtree(dec *xml.Decoder, start xml.StartElement) ([]byte, error) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	if err := enc.EncodeToken(start); err != nil {
		return nil, err
	}
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		if err := enc.EncodeToken(tok); err != nil {
			return nil, err
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	enc.Flush()
	return buf.Bytes(), nil
}

func skipSubtree(dec *xml.Decoder, start xml.StartElement) error {
	depth := 1
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
			if depth == 0 && t.Name == start.Name {
				return nil
			}
		}
	}
	return nil
}
