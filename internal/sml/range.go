package sml

import (
	"fmt"
	"strings"
)

// ParseCellRange parses "A1:B10", "$A$1:$B$10" into inclusive min/max column and row (1-based).
func ParseCellRange(s string) (minCol, minRow, maxCol, maxRow int, err error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "$", "")
	i := strings.IndexByte(s, ':')
	if i < 0 {
		return 0, 0, 0, 0, fmt.Errorf("sml: not a range %q", s)
	}
	a, b := strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])
	c1, r1, err := CellRefToIndexes(a)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	c2, r2, err := CellRefToIndexes(b)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	minCol, maxCol = c1, c2
	minRow, maxRow = r1, r2
	if minCol > maxCol {
		minCol, maxCol = maxCol, minCol
	}
	if minRow > maxRow {
		minRow, maxRow = maxRow, minRow
	}
	return minCol, minRow, maxCol, maxRow, nil
}
