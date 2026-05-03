package sml

import "testing"

func TestParseCellRange(t *testing.T) {
	mc, mr, xc, xr, err := ParseCellRange("B2:$D$5")
	if err != nil {
		t.Fatal(err)
	}
	if mc != 2 || mr != 2 || xc != 4 || xr != 5 {
		t.Fatalf("%d,%d %d,%d", mc, mr, xc, xr)
	}
}
