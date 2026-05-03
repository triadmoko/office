package wml

import (
	"bytes"
	"encoding/xml"
	"io"
)

// ParseStyles parses word/styles.xml.
func ParseStyles(r io.Reader) (*Styles, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	s := &Styles{ByID: make(map[string]*Style)}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if !isWML(se.Name.Space, se.Name.Local) {
			continue
		}
		switch se.Name.Local {
		case "docDefaults":
			if err := parseDocDefaults(dec, se, s); err != nil {
				return nil, err
			}
		case "style":
			id := attrLocal(se.Attr, "styleId")
			if id == "" {
				if err := skipSubtree(dec, se); err != nil {
					return nil, err
				}
				continue
			}
			raw, err := captureSubtree(dec, se)
			if err != nil {
				return nil, err
			}
			st := decodeStyleElement(raw, id, attrLocal(se.Attr, "type"))
			st.Raw = raw
			s.ByID[id] = st
		}
	}
	return s, nil
}

func parseDocDefaults(dec *xml.Decoder, start xml.StartElement, s *Styles) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "rPrDefault" {
				if err := parseRPrDefault(dec, t, s); err != nil {
					return err
				}
				continue
			}
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "pPrDefault" {
				if err := parsePPrDefault(dec, t, s); err != nil {
					return err
				}
				continue
			}
			if err := skipSubtree(dec, t); err != nil {
				return err
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "docDefaults" {
				return nil
			}
		}
	}
}

func parseRPrDefault(dec *xml.Decoder, start xml.StartElement, s *Styles) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "rPr" {
				run := &Run{}
				if err := parseRPr(dec, t, run); err != nil {
					return err
				}
				s.RDefaults = run.RPr
				continue
			}
			if err := skipSubtree(dec, t); err != nil {
				return err
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "rPrDefault" {
				return nil
			}
		}
	}
}

func parsePPrDefault(dec *xml.Decoder, start xml.StartElement, s *Styles) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "pPr" {
				p := &Paragraph{}
				if err := parsePPr(dec, t, p); err != nil {
					return err
				}
				s.DocDefaults = p.PPr
				continue
			}
			if err := skipSubtree(dec, t); err != nil {
				return err
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "pPrDefault" {
				return nil
			}
		}
	}
}

func decodeStyleElement(raw []byte, id, typ string) *Style {
	st := &Style{ID: id, Type: typ}
	d := xml.NewDecoder(bytes.NewReader(raw))
	d.Strict = false
	tok0, err := d.Token()
	if err != nil {
		return st
	}
	se0, ok := tok0.(xml.StartElement)
	if !ok || se0.Name.Local != "style" {
		return st
	}
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
			if err := skipSubtree(d, se); err != nil {
				break
			}
			continue
		}
		switch se.Name.Local {
		case "name":
			st.Name = valAttr(se.Attr)
			_ = skipSubtree(d, se)
		case "basedOn":
			st.BasedOn = valAttr(se.Attr)
			_ = skipSubtree(d, se)
		case "link":
			st.LinkedStyle = valAttr(se.Attr)
			_ = skipSubtree(d, se)
		case "rPr":
			run := &Run{}
			if err := parseRPr(d, se, run); err == nil {
				st.RPr = run.RPr
			}
		case "pPr":
			p := &Paragraph{}
			if err := parsePPr(d, se, p); err == nil {
				st.PPr = p.PPr
			}
		default:
			_ = skipSubtree(d, se)
		}
	}
	return st
}
