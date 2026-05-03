package sml

import "testing"

func TestCellRefToIndexes(t *testing.T) {
	tests := []struct {
		ref    string
		col, r int
	}{
		{"A1", 1, 1},
		{"$B$2", 2, 2},
		{"AA10", 27, 10},
		{"XFD1048576", 16384, 1048576},
	}
	for _, tc := range tests {
		c, rr, err := CellRefToIndexes(tc.ref)
		if err != nil {
			t.Fatalf("%s: %v", tc.ref, err)
		}
		if c != tc.col || rr != tc.r {
			t.Fatalf("%s: got %d,%d want %d,%d", tc.ref, c, rr, tc.col, tc.r)
		}
		if back := IndexesToCellRef(c, rr); back == "" {
			t.Fatalf("round trip empty %s", tc.ref)
		}
	}
	if IndexesToCellRef(1, 1) != "A1" || IndexesToCellRef(27, 10) != "AA10" {
		t.Fatalf("IndexesToCellRef: %q %q", IndexesToCellRef(1, 1), IndexesToCellRef(27, 10))
	}
}
