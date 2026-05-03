package xlsx

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestOpenMinimalWorkbook(t *testing.T) {
	data := buildMinimalXLSX(t)
	wb, err := Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if wb.MainPart() != "/xl/workbook.xml" {
		t.Fatalf("main: %q", wb.MainPart())
	}
	if err := wb.Write(&bytes.Buffer{}); err != ErrNotImplemented {
		t.Fatalf("write: %v", err)
	}
	sheets, err := wb.Sheets()
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets) != 0 {
		t.Fatalf("minimal sheets: %d", len(sheets))
	}
}

func TestWorkbook201SheetsAndDefinedNames(t *testing.T) {
	data := buildWorkbook201Fixture(t)
	wb, err := Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	sheets, err := wb.Sheets()
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets) != 3 {
		t.Fatalf("sheets: %d", len(sheets))
	}
	if sheets[0].Name() != "Visible1" || sheets[0].State() != SheetVisible {
		t.Fatalf("sheet0: %q %v", sheets[0].Name(), sheets[0].State())
	}
	if sheets[1].Name() != "HiddenSheet" || sheets[1].State() != SheetHidden {
		t.Fatalf("sheet1: %q %v", sheets[1].Name(), sheets[1].State())
	}
	if sheets[2].State() != SheetVisible {
		t.Fatalf("sheet2 state: %v", sheets[2].State())
	}
	sh, err := wb.SheetByName("visible2")
	if err != nil || sh == nil || sh.Name() != "visible2" {
		t.Fatalf("SheetByName visible2: %v %#v", err, sh)
	}
	sh2, err := wb.SheetByName("VISIBLE2")
	if err != nil || sh2 == nil || sh2 != sh {
		t.Fatalf("case insensitive: %v", err)
	}
	names, err := wb.DefinedNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0].Name() != "MyRange" || names[0].Formula() != "Visible1!$A$1:$B$2" {
		t.Fatalf("defined: %#v", names)
	}
	s0, err := wb.SharedString(0)
	if err != nil {
		t.Fatal(err)
	}
	if s0 != "alpha" {
		t.Fatalf("SharedString(0): %q", s0)
	}
	s1, err := wb.SharedString(1)
	if err != nil {
		t.Fatal(err)
	}
	if s1 != "betacd" { // rich text join
		t.Fatalf("SharedString(1): %q", s1)
	}
	if _, err := wb.SharedString(99); err != ErrSharedStringOutOfRange {
		t.Fatalf("oob: %v", err)
	}
}

func buildMinimalXLSX(t *testing.T) []byte {
	t.Helper()
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
</Relationships>`)
	write("xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"/>`)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func buildWorkbook201Fixture(t *testing.T) []byte {
	t.Helper()
	ws := `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData/></worksheet>`
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
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
<Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet3.xml"/>
<Relationship Id="rId4" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings" Target="sharedStrings.xml"/>
</Relationships>`)
	write("xl/sharedStrings.xml", `<?xml version="1.0" encoding="UTF-8"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="2" uniqueCount="2">
<si><t>alpha</t></si>
<si><r><t>beta</t></r><r><t>cd</t></r></si>
</sst>`)
	write("xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
<sheets>
<sheet name="Visible1" sheetId="1" r:id="rId1"/>
<sheet name="HiddenSheet" sheetId="2" state="hidden" r:id="rId2"/>
<sheet name="visible2" sheetId="3" r:id="rId3"/>
</sheets>
<definedNames>
<definedName name="MyRange">Visible1!$A$1:$B$2</definedName>
</definedNames>
</workbook>`)
	write("xl/worksheets/sheet1.xml", ws)
	write("xl/worksheets/sheet2.xml", ws)
	write("xl/worksheets/sheet3.xml", ws)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
