package xlsx

import (
	"archive/zip"
	"bytes"
	"testing"
	"time"
)

func TestRowsIteratorMixedTypes(t *testing.T) {
	data := buildRowsIteratorFixture(t)
	wb, err := Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	sheets, err := wb.Sheets()
	if err != nil {
		t.Fatal(err)
	}
	sh := sheets[0]
	rows, err := sh.Rows()
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		for _, c := range rows.Row().Cells() {
			_, err := c.Value()
			if err != nil && c.Type() != CellError {
				t.Fatalf("%s: %v", c.Address(), err)
			}
			n++
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if n != 6 {
		t.Fatalf("cells %d", n)
	}
}

func TestCellValueDateAndFormula(t *testing.T) {
	data := buildRowsIteratorFixture(t)
	wb, err := Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	sh, err := wb.Sheets()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := sh[0].Rows()
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("row1")
	}
	a1 := rows.Row().Cells()[0]
	if a1.Address() != "A1" || a1.Type() != CellString {
		t.Fatalf("A1: %s %v", a1.Address(), a1.Type())
	}
	v1, _ := a1.Value()
	if v1 != "hello" {
		t.Fatalf("A1 val %v", v1)
	}
	if !rows.Next() {
		t.Fatal("row2")
	}
	r2 := rows.Row().Cells()
	var a2, b2 *Cell
	for _, c := range r2 {
		switch c.Address() {
		case "A2":
			a2 = c
		case "B2":
			b2 = c
		}
	}
	if a2 == nil || a2.Type() != CellDate {
		t.Fatalf("A2 date: %#v", a2)
	}
	v2, err := a2.Value()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := v2.(time.Time); !ok {
		t.Fatalf("A2 type %T", v2)
	}
	if b2 == nil || b2.Formula() == "" {
		t.Fatal("B2 formula")
	}
	if b2.Type() != CellNumber {
		t.Fatalf("B2 type %v", b2.Type())
	}
}

func buildRowsIteratorFixture(t *testing.T) []byte {
	t.Helper()
	ws := `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<sheetData>
<row r="1">
<c r="A1" t="s"><v>0</v></c>
<c r="B1"><v>3.5</v></c>
<c r="C1" t="b"><v>1</v></c>
</row>
<row r="2">
<c r="A2" s="1"><v>1</v></c>
<c r="B2"><f>1+1</f><v>2</v></c>
<c r="C2" t="e"><v>#DIV/0!</v></c>
</row>
</sheetData>
</worksheet>`
	styles := `<?xml version="1.0" encoding="UTF-8"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<cellXfs count="2">
<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
<xf numFmtId="14" fontId="0" fillId="0" borderId="0" xfId="0"/>
</cellXfs>
</styleSheet>`
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	write := func(name, body string) {
		t.Helper()
		f, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	write("[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
</Types>`)
	write("_rels/.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`)
	write("xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`)
	write("xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheets><sheet name="Data" sheetId="1" r:id="rId1"/></sheets>
</workbook>`)
	write("xl/sharedStrings.xml", `<?xml version="1.0" encoding="UTF-8"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="1" uniqueCount="1">
<si><t>hello</t></si>
</sst>`)
	write("xl/styles.xml", styles)
	write("xl/worksheets/sheet1.xml", ws)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
