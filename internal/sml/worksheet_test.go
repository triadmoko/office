package sml

import (
	"strings"
	"testing"
)

func TestStreamWorksheetRows(t *testing.T) {
	const xml = `<?xml version="1.0"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<sheetData>
<row r="2"><c r="A2" t="s"><v>0</v></c></row>
<row r="5"><c r="A5"><v>3.14</v></c><c r="B5" t="b"><v>1</v></c></row>
</sheetData>
</worksheet>`
	var got int
	err := StreamWorksheetRows(strings.NewReader(xml), func(rd RowData) error {
		got++
		if got == 1 && rd.Index != 2 {
			t.Fatalf("row1 idx %d", rd.Index)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if got != 2 {
		t.Fatalf("rows %d", got)
	}
}
