package docx

import (
	"bytes"
	"fmt"

	"github.com/triadmoko/office/internal/wml"
)

// MarshalStylesXML emits a minimal valid styles.xml for the given registry.
func MarshalStylesXML(s *wml.Styles) ([]byte, error) {
	if s == nil {
		s = wml.DefaultStyles()
	}
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:styles xmlns:w="` + nsW + `">`)
	b.WriteString(`<w:docDefaults><w:rPrDefault><w:rPr><w:lang w:val="en-US" w:eastAsia="en-US" w:bidi="ar-SA"/></w:rPr></w:rPrDefault><w:pPrDefault/></w:docDefaults>`)
	b.WriteString(`<w:latentStyles w:defLockedState="0" w:defUIPriority="99" w:defSemiHidden="0" w:defUnhideWhenUsed="0" w:defQFormat="0" w:count="371"/>`)
	ids := []string{"Normal", "Heading1", "Heading2", "Heading3", "ListParagraph"}
	for _, id := range ids {
		st := s.ByID[id]
		if st == nil {
			continue
		}
		if err := marshalOneStyle(&b, st); err != nil {
			return nil, err
		}
	}
	for id, st := range s.ByID {
		skip := false
		for _, k := range ids {
			if id == k {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if err := marshalOneStyle(&b, st); err != nil {
			return nil, err
		}
	}
	b.WriteString(`</w:styles>`)
	return b.Bytes(), nil
}

func marshalOneStyle(b *bytes.Buffer, st *wml.Style) error {
	if st == nil || st.ID == "" {
		return nil
	}
	if len(st.Raw) > 0 {
		b.Write(st.Raw)
		return nil
	}
	typ := st.Type
	if typ == "" {
		typ = "paragraph"
	}
	name := st.Name
	if name == "" {
		name = st.ID
	}
	b.WriteString(`<w:style w:type="` + escapeAttr(typ) + `" w:styleId="` + escapeAttr(st.ID) + `">`)
	b.WriteString(`<w:name w:val="` + escapeAttr(name) + `"/>`)
	if st.BasedOn != "" {
		b.WriteString(`<w:basedOn w:val="` + escapeAttr(st.BasedOn) + `"/>`)
	}
	if st.LinkedStyle != "" {
		b.WriteString(`<w:link w:val="` + escapeAttr(st.LinkedStyle) + `"/>`)
	}
	if !emptyRunProps(st.RPr) {
		b.WriteString(`<w:rPr>`)
		if st.RPr.Bold {
			b.WriteString(`<w:b/>`)
		}
		if st.RPr.Italic {
			b.WriteString(`<w:i/>`)
		}
		if st.RPr.FontSizeHalf != 0 {
			b.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/>`, st.RPr.FontSizeHalf))
		}
		b.WriteString(`</w:rPr>`)
	}
	b.WriteString(`</w:style>`)
	return nil
}
