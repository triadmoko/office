package wml

import (
	"strings"
	"testing"
)

func TestParseDocumentWithPPrAndNumPr(t *testing.T) {
	const doc = `<?xml version="1.0"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
		`<w:body><w:p><w:pPr><w:pStyle w:val="ListParagraph"/>` +
		`<w:numPr><w:ilvl w:val="0"></w:ilvl><w:numId w:val="1"></w:numId></w:numPr>` +
		`</w:pPr><w:r><w:t>a</w:t></w:r></w:p></w:body></w:document>`
	d, err := ParseDocument(strings.NewReader(doc))
	if err != nil {
		t.Fatal(err)
	}
	ps := d.DirectParagraphs()
	if len(ps) != 1 || ps[0].PPr.Numbering == nil || ps[0].PPr.Numbering.NumID != 1 {
		t.Fatalf("numPr: %+v", ps[0].PPr.Numbering)
	}
}
