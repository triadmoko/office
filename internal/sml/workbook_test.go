package sml

import (
	"strings"
	"testing"
)

func TestParseWorkbookSheetsAndNames(t *testing.T) {
	const xml = `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="A" sheetId="1" r:id="r1"/>
    <sheet name="B" sheetId="2" state="hidden" r:id="r2"/>
    <sheet name="C" sheetId="3" state="veryHidden" r:id="r3"/>
  </sheets>
  <definedNames>
    <definedName name="Rng">A!$A$1</definedName>
    <definedName name="Local" localSheetId="0">A!$B$1</definedName>
  </definedNames>
</workbook>`
	sheets, names, err := ParseWorkbook(strings.NewReader(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets) != 3 {
		t.Fatalf("sheets: %d", len(sheets))
	}
	if sheets[1].State != "hidden" || sheets[2].State != "veryHidden" {
		t.Fatalf("state: %#v %#v", sheets[1].State, sheets[2].State)
	}
	if len(names) != 2 || names[0].Name != "Rng" || names[0].Formula != "A!$A$1" {
		t.Fatalf("names0: %#v", names[0])
	}
	if names[1].LocalSheetID == nil || *names[1].LocalSheetID != 0 {
		t.Fatalf("local: %#v", names[1].LocalSheetID)
	}
}
