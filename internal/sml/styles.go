package sml

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
)

// StylesTable holds minimal style data for cell value interpretation (dates, etc.).
type StylesTable struct {
	NumFmtIDs    []uint32 // index = cell style s (cellXfs index) -> numFmtId
	CustomNumFmt map[uint32]string
}

type stylesRoot struct {
	NumFmts *struct {
		NumFmt []struct {
			NumFmtID   string `xml:"numFmtId,attr"`
			FormatCode string `xml:"formatCode,attr"`
		} `xml:"numFmt"`
	} `xml:"numFmts"`
	CellXfs *struct {
		Xf []struct {
			NumFmtID string `xml:"numFmtId,attr"`
		} `xml:"xf"`
	} `xml:"cellXfs"`
}

// ParseStylesTable parses xl/styles.xml for cellXfs -> numFmtId mapping.
func ParseStylesTable(r io.Reader) (*StylesTable, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	var root stylesRoot
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}
	if root.CellXfs == nil {
		return &StylesTable{NumFmtIDs: nil}, nil
	}
	out := make([]uint32, 0, len(root.CellXfs.Xf))
	for _, xf := range root.CellXfs.Xf {
		id, _ := strconv.ParseUint(strings.TrimSpace(xf.NumFmtID), 10, 32)
		out = append(out, uint32(id))
	}
	custom := map[uint32]string{}
	if root.NumFmts != nil {
		for _, nf := range root.NumFmts.NumFmt {
			id, _ := strconv.ParseUint(strings.TrimSpace(nf.NumFmtID), 10, 32)
			custom[uint32(id)] = nf.FormatCode
		}
	}
	return &StylesTable{NumFmtIDs: out, CustomNumFmt: custom}, nil
}

// IsDateStyle reports whether cell style index s refers to a date-like number format.
func (st *StylesTable) IsDateStyle(styleIdx int) bool {
	if st == nil || styleIdx < 0 || styleIdx >= len(st.NumFmtIDs) {
		return false
	}
	id := st.NumFmtIDs[styleIdx]
	if IsBuiltInDateNumFmt(id) {
		return true
	}
	if code, ok := st.CustomNumFmt[id]; ok {
		return IsDateLikeFormatString(code)
	}
	return false
}

// NumberFormatForStyle resolves the number format string for a cell style index (cell s attribute).
func (st *StylesTable) NumberFormatForStyle(styleIdx int) string {
	if st == nil || styleIdx < 0 || styleIdx >= len(st.NumFmtIDs) {
		return "General"
	}
	id := st.NumFmtIDs[styleIdx]
	if s, ok := NumberFormatBuiltIn(id); ok {
		return s
	}
	if c, ok := st.CustomNumFmt[id]; ok && c != "" {
		return c
	}
	return "General"
}

// IsBuiltInDateNumFmt returns true for Excel built-in format ids commonly used for dates/times.
func IsBuiltInDateNumFmt(numFmtID uint32) bool {
	switch numFmtID {
	case 14, 15, 16, 17, 18, 19, 20, 21, 22, 45, 46, 47:
		return true
	default:
		return false
	}
}

// IsDateLikeFormatString uses a coarse heuristic (y/m/d/h/s) for custom formats.
func IsDateLikeFormatString(code string) bool {
	c := strings.ToLower(code)
	return strings.ContainsAny(c, "y") && (strings.Contains(c, "d") || strings.Contains(c, "m")) ||
		strings.Contains(c, "h") && strings.Contains(c, "m") && strings.Contains(c, "s")
}
