package wml

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

// ParseSectPr parses a single w:sectPr fragment into a Section.
// Data may be a full <w:sectPr>...</w:sectPr> subtree (e.g. from captureSubtree) or inner elements only.
func ParseSectPr(data []byte) Section {
	var sec Section
	sec.Raw = append([]byte(nil), data...)
	if len(bytes.TrimSpace(data)) == 0 {
		sec.PageSize = PageSize{Width: PageA4W, Height: PageA4H, Orient: Portrait}
		return sec
	}
	d := xml.NewDecoder(bytes.NewReader(data))
	d.Strict = false
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if !isWML(se.Name.Space, se.Name.Local) {
			_ = skipSubtree(d, se)
			continue
		}
		if se.Name.Local == "sectPr" {
			_ = parseSectPrContents(d, &sec)
			break
		}
		parseSectPrChildElement(d, se, &sec)
	}
	if sec.PageSize.Width == 0 || sec.PageSize.Height == 0 {
		sec.PageSize = PageSize{Width: PageA4W, Height: PageA4H, Orient: sec.PageSize.Orient}
	}
	return sec
}

func parseSectPrContents(d *xml.Decoder, sec *Section) error {
	for {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if !isWML(t.Name.Space, t.Name.Local) {
				if err := skipSubtree(d, t); err != nil {
					return err
				}
				continue
			}
			parseSectPrChildElement(d, t, sec)
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "sectPr" {
				return nil
			}
		}
	}
}

func parseSectPrChildElement(d *xml.Decoder, se xml.StartElement, sec *Section) {
	switch se.Name.Local {
	case "type":
		sec.TypeVal = strings.TrimSpace(valAttr(se.Attr))
		_ = skipSubtree(d, se)
	case "pgSz":
		sec.PageSize.Width = twipAttr(attrLocal(se.Attr, "w"))
		sec.PageSize.Height = twipAttr(attrLocal(se.Attr, "h"))
		if strings.EqualFold(attrLocal(se.Attr, "orient"), "landscape") {
			sec.PageSize.Orient = Landscape
		}
		_ = skipSubtree(d, se)
	case "pgMar":
		sec.Margins.Top = twipAttr(attrLocal(se.Attr, "top"))
		sec.Margins.Bottom = twipAttr(attrLocal(se.Attr, "bottom"))
		sec.Margins.Left = twipAttr(attrLocal(se.Attr, "left"))
		sec.Margins.Right = twipAttr(attrLocal(se.Attr, "right"))
		sec.Margins.Header = twipAttr(attrLocal(se.Attr, "header"))
		sec.Margins.Footer = twipAttr(attrLocal(se.Attr, "footer"))
		sec.Margins.Gutter = twipAttr(attrLocal(se.Attr, "gutter"))
		_ = skipSubtree(d, se)
	case "cols":
		sec.Columns.Num = intAttr(attrLocal(se.Attr, "num"))
		if sec.Columns.Num == 0 {
			sec.Columns.Num = 1
		}
		sec.Columns.Sep = attrLocal(se.Attr, "sep") == "1" || strings.EqualFold(attrLocal(se.Attr, "sep"), "true")
		sec.Columns.EqualWidth = strings.EqualFold(attrLocal(se.Attr, "equalWidth"), "1") ||
			strings.EqualFold(attrLocal(se.Attr, "equalWidth"), "true")
		_ = skipSubtree(d, se)
	case "pgNumType":
		if v := strings.TrimSpace(attrLocal(se.Attr, "fmt")); v != "" {
			sec.PageNumFmt = v
		}
		if s := strings.TrimSpace(attrLocal(se.Attr, "start")); s != "" {
			sec.PageNumStart = intAttr(s)
			sec.PageNumStartSet = true
		}
		_ = skipSubtree(d, se)
	default:
		_ = skipSubtree(d, se)
	}
}

// SectionsFromDocument builds section list from body + paragraph breaks.
func SectionsFromDocument(doc *Document) []Section {
	if doc == nil {
		return nil
	}
	var out []Section
	// Section breaks from w:p/w:pPr/w:sectPr
	for _, p := range CollectParagraphsPreorder(&doc.Body) {
		if len(p.PPr.SectPr) > 0 {
			out = append(out, ParseSectPr(p.PPr.SectPr))
		}
	}
	if len(doc.Body.SectPr) > 0 {
		out = append(out, ParseSectPr(doc.Body.SectPr))
	} else if len(out) == 0 {
		out = append(out, ParseSectPr(nil))
	}
	return out
}

// CollectParagraphsPreorder returns every w:p in reading order (body blocks and table cells).
func CollectParagraphsPreorder(body *Body) []*Paragraph {
	if body == nil {
		return nil
	}
	var acc []*Paragraph
	var walk func(b *Body)
	walk = func(b *Body) {
		for _, bl := range b.Blocks {
			switch {
			case bl.Para != nil:
				acc = append(acc, bl.Para)
			case bl.Table != nil:
				for _, row := range bl.Table.Rows {
					if row == nil {
						continue
					}
					for _, cell := range row.Cells {
						if cell == nil {
							continue
						}
						for _, cb := range cell.Blocks {
							if cb.Para != nil {
								acc = append(acc, cb.Para)
							} else if cb.Table != nil {
								walk(&Body{Blocks: []BodyBlock{{Table: cb.Table}}})
							}
						}
					}
				}
			}
		}
	}
	walk(body)
	return acc
}
