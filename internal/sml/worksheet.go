package sml

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
)

// CellData is one parsed <c> element within a row.
type CellData struct {
	Ref         string
	T           string // s, str, inlineStr, b, e, n (implicit)
	S           int    // style index, -1 if absent
	Raw         string // <v> text
	Formula     string // <f> text
	InlinePlain string // joined <is> text runs
}

// RowData is one non-empty worksheet row (may have zero cells after filter — caller skips).
type RowData struct {
	Index int // 1-based from r attribute
	Cells []CellData
}

type rowDecode struct {
	R string    `xml:"r,attr"`
	C []cellDec `xml:"c"`
}

type cellDec struct {
	R string `xml:"r,attr"`
	T string `xml:"t,attr"`
	S string `xml:"s,attr"`
	F *struct {
		Inner string `xml:",chardata"`
	} `xml:"f"`
	V *struct {
		Inner string `xml:",chardata"`
	} `xml:"v"`
	IS *struct {
		T []struct {
			Space string `xml:"http://www.w3.org/XML/1998/namespace space,attr"`
			Text  string `xml:",chardata"`
		} `xml:"t"`
	} `xml:"is"`
}

// StreamWorksheetRows calls fn for each <row> that contains at least one <c>.
func StreamWorksheetRows(r io.Reader, fn func(RowData) error) error {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "row" {
			continue
		}
		var rd rowDecode
		if err := dec.DecodeElement(&rd, &se); err != nil {
			return err
		}
		rowIdx, _ := strconv.Atoi(strings.TrimSpace(rd.R))
		if rowIdx < 1 {
			rowIdx = 0
		}
		if len(rd.C) == 0 {
			continue
		}
		row := RowData{Index: rowIdx, Cells: make([]CellData, 0, len(rd.C))}
		for i := range rd.C {
			cd := cellDecToData(&rd.C[i], rowIdx, i+1)
			row.Cells = append(row.Cells, cd)
		}
		if err := fn(row); err != nil {
			return err
		}
	}
}

func cellDecToData(c *cellDec, rowIdx, colGuess int) CellData {
	ref := strings.TrimSpace(c.R)
	style := -1
	if s := strings.TrimSpace(c.S); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			style = v
		}
	}
	raw := ""
	if c.V != nil {
		raw = strings.TrimSpace(c.V.Inner)
	}
	formula := ""
	if c.F != nil {
		formula = strings.TrimSpace(c.F.Inner)
	}
	inline := ""
	if c.IS != nil {
		var b strings.Builder
		for _, t := range c.IS.T {
			txt := t.Text
			if t.Space != "preserve" {
				txt = strings.TrimSpace(txt)
			}
			b.WriteString(txt)
		}
		inline = b.String()
	}
	if ref == "" && rowIdx >= 1 {
		ref = IndexesToCellRef(colGuess, rowIdx)
	}
	return CellData{
		Ref:         ref,
		T:           strings.TrimSpace(c.T),
		S:           style,
		Raw:         raw,
		Formula:     formula,
		InlinePlain: inline,
	}
}
