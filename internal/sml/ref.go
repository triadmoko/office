package sml

import (
	"fmt"
	"strings"
	"unicode"
)

// CellRefToIndexes converts "A1", "$B$2", or "AA10" to 1-based column and row (Excel convention).
func CellRefToIndexes(ref string) (col, row int, err error) {
	ref = strings.TrimSpace(ref)
	ref = strings.ReplaceAll(ref, "$", "")
	if ref == "" {
		return 0, 0, fmt.Errorf("sml: empty cell reference")
	}
	i := 0
	for i < len(ref) && unicode.IsLetter(rune(ref[i])) {
		i++
	}
	if i == 0 || i >= len(ref) {
		return 0, 0, fmt.Errorf("sml: bad cell reference %q", ref)
	}
	colLetters := ref[:i]
	rowDigits := ref[i:]
	col = 0
	for _, ch := range colLetters {
		if ch < 'A' || ch > 'Z' {
			return 0, 0, fmt.Errorf("sml: bad column in %q", ref)
		}
		col = col*26 + int(ch-'A') + 1
	}
	var r int
	for _, ch := range rowDigits {
		if ch < '0' || ch > '9' {
			return 0, 0, fmt.Errorf("sml: bad row in %q", ref)
		}
		r = r*10 + int(ch-'0')
	}
	if r < 1 || col < 1 {
		return 0, 0, fmt.Errorf("sml: invalid indexes for %q", ref)
	}
	return col, r, nil
}

// IndexesToCellRef builds "A1" from 1-based column and row.
func IndexesToCellRef(col, row int) string {
	if col < 1 || row < 1 {
		return ""
	}
	var letters strings.Builder
	n := col
	for n > 0 {
		n--
		letters.WriteByte(byte('A' + n%26))
		n /= 26
	}
	s := []byte(letters.String())
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return string(s) + fmt.Sprintf("%d", row)
}
