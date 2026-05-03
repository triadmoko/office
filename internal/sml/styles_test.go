package sml

import (
	"strings"
	"testing"
)

func TestParseStylesTableCellXfs(t *testing.T) {
	const xml = `<?xml version="1.0"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
<cellXfs count="2">
<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>
<xf numFmtId="14" fontId="0" fillId="0" borderId="0" xfId="0"/>
</cellXfs>
</styleSheet>`
	st, err := ParseStylesTable(strings.NewReader(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(st.NumFmtIDs) != 2 || st.NumFmtIDs[1] != 14 {
		t.Fatalf("xfs: %#v", st.NumFmtIDs)
	}
	if !st.IsDateStyle(1) || st.IsDateStyle(0) {
		t.Fatalf("date style: %v %v", st.IsDateStyle(0), st.IsDateStyle(1))
	}
}
