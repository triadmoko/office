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

func TestParseTableTblGridAndTrHeight(t *testing.T) {
	const doc = `<?xml version="1.0"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
		`<w:body>` +
		`<w:tbl>` +
		`<w:tblGrid><w:gridCol w:w="1000"/><w:gridCol w:w="2000"/></w:tblGrid>` +
		`<w:tr><w:trPr><w:trHeight w:val="500" w:hRule="exact"/></w:trPr>` +
		`<w:tc><w:tcPr/><w:p/></w:tc><w:tc><w:p/></w:tc></w:tr>` +
		`</w:tbl>` +
		`</w:body></w:document>`
	d, err := ParseDocument(strings.NewReader(doc))
	if err != nil {
		t.Fatal(err)
	}
	tbl := d.Body.Blocks[0].Table
	if tbl == nil || len(tbl.Props.GridColWidths) != 2 {
		t.Fatalf("grid: %+v", tbl)
	}
	if tbl.Props.GridColWidths[0] != 1000 || tbl.Props.GridColWidths[1] != 2000 {
		t.Fatal(tbl.Props.GridColWidths)
	}
	if len(tbl.Rows) != 1 || tbl.Rows[0].HeightVal != 500 || tbl.Rows[0].HeightRule != TrHeightExact {
		t.Fatalf("row: %+v", tbl.Rows[0])
	}
}
