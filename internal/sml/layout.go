package sml

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
)

// FreezePane describes a frozen split from sheetViews/pane.
type FreezePane struct {
	Row int // number of frozen rows (ySplit)
	Col int // number of frozen columns (xSplit)
}

// WorksheetLayout collects non-cell worksheet metadata.
type WorksheetLayout struct {
	MergedRefs   []string
	HiddenRows   []int
	HiddenCols   []int
	Freeze       *FreezePane
	RowHeights   map[int]float64 // 1-based row -> height in points
	ColumnWidths map[int]float64 // 1-based col -> width
}

// ScanWorksheetLayout scans worksheet XML for merges, hidden cols/rows, freeze, dimensions.
func ScanWorksheetLayout(r io.Reader) (*WorksheetLayout, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	out := &WorksheetLayout{
		RowHeights:   make(map[int]float64),
		ColumnWidths: make(map[int]float64),
	}
	hiddenRowSet := make(map[int]struct{})
	hiddenColSet := make(map[int]struct{})
	var inSheetData bool
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "sheetData":
				inSheetData = true
			case "mergeCell":
				for _, a := range t.Attr {
					if a.Name.Local == "ref" {
						out.MergedRefs = append(out.MergedRefs, strings.TrimSpace(a.Value))
					}
				}
			case "col":
				min, max := 1, 1
				hidden := false
				var width float64
				hasW := false
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "min":
						min, _ = strconv.Atoi(strings.TrimSpace(a.Value))
					case "max":
						max, _ = strconv.Atoi(strings.TrimSpace(a.Value))
					case "hidden":
						hidden = a.Value == "1" || strings.EqualFold(a.Value, "true")
					case "width":
						if w, err := strconv.ParseFloat(strings.TrimSpace(a.Value), 64); err == nil {
							width = w
							hasW = true
						}
					}
				}
				if hidden {
					for c := min; c <= max; c++ {
						hiddenColSet[c] = struct{}{}
					}
				}
				if hasW {
					for c := min; c <= max; c++ {
						out.ColumnWidths[c] = width
					}
				}
			case "pane":
				fp := parsePane(t)
				if fp != nil && (fp.Row > 0 || fp.Col > 0) {
					out.Freeze = fp
				}
			case "row":
				if !inSheetData {
					continue
				}
				rnum := 0
				ht := 0.0
				hasHt := false
				hidden := false
				for _, a := range t.Attr {
					switch a.Name.Local {
					case "r":
						rnum, _ = strconv.Atoi(strings.TrimSpace(a.Value))
					case "ht":
						if v, err := strconv.ParseFloat(strings.TrimSpace(a.Value), 64); err == nil {
							ht = v
							hasHt = true
						}
					case "hidden":
						if a.Value == "1" || strings.EqualFold(a.Value, "true") {
							hidden = true
						}
					}
				}
				if rnum >= 1 {
					if hasHt {
						out.RowHeights[rnum] = ht
					}
					if hidden {
						hiddenRowSet[rnum] = struct{}{}
					}
				}
			}
		case xml.EndElement:
			if t.Name.Local == "sheetData" {
				inSheetData = false
			}
		}
	}
	for r := range hiddenRowSet {
		out.HiddenRows = append(out.HiddenRows, r)
	}
	for c := range hiddenColSet {
		out.HiddenCols = append(out.HiddenCols, c)
	}
	sortInts(out.HiddenRows)
	sortInts(out.HiddenCols)
	return out, nil
}

func parsePane(se xml.StartElement) *FreezePane {
	var xSplit, ySplit float64
	for _, a := range se.Attr {
		switch a.Name.Local {
		case "xSplit":
			xSplit, _ = strconv.ParseFloat(strings.TrimSpace(a.Value), 64)
		case "ySplit":
			ySplit, _ = strconv.ParseFloat(strings.TrimSpace(a.Value), 64)
		}
	}
	if xSplit == 0 && ySplit == 0 {
		return nil
	}
	return &FreezePane{Row: int(ySplit), Col: int(xSplit)}
}

func sortInts(a []int) {
	// small slices — insertion sort
	for i := 1; i < len(a); i++ {
		j, v := i, a[i]
		for j > 0 && a[j-1] > v {
			a[j] = a[j-1]
			j--
		}
		a[j] = v
	}
}
