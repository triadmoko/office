package docx

import (
	"bytes"
	"strconv"

	"github.com/triadmoko/office/internal/wml"
)

// MarshalNumberingXML returns numbering.xml bytes or nil if no numbering is used.
func MarshalNumberingXML(n *wml.Numbering) []byte {
	if n == nil || len(n.Nums) == 0 {
		return nil
	}
	var b bytes.Buffer
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:numbering xmlns:w="` + nsW + `">`)
	for aid, ab := range n.Abstract {
		if ab == nil {
			continue
		}
		b.WriteString(`<w:abstractNum w:abstractNumId="` + strconv.Itoa(aid) + `">`)
		for _, lvl := range ab.Levels {
			if lvl == nil {
				continue
			}
			marshalLvl(&b, lvl)
		}
		b.WriteString(`</w:abstractNum>`)
	}
	for nid, def := range n.Nums {
		if def == nil {
			continue
		}
		b.WriteString(`<w:num w:numId="` + strconv.Itoa(nid) + `"><w:abstractNumId w:val="` + strconv.Itoa(def.AbstractID) + `"/></w:num>`)
	}
	b.WriteString(`</w:numbering>`)
	return b.Bytes()
}

func marshalLvl(b *bytes.Buffer, lvl *wml.NumLevel) {
	if lvl == nil {
		return
	}
	b.WriteString(`<w:lvl w:ilvl="` + strconv.Itoa(lvl.Ilvl) + `"><w:start w:val="` + strconv.Itoa(lvl.StartAt) + `"/><w:numFmt w:val="` + escapeAttr(lvl.Format) + `"/><w:lvlText w:val="` + escapeAttr(lvl.Text) + `"/><w:lvlJc w:val="left"/></w:lvl>`)
}
