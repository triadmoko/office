package xlsx

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"

	"github.com/triadmoko/office/internal/sml"
)

// Rows iterates worksheet rows in document order (streaming).
type Rows struct {
	sheet *Sheet
	rc    io.ReadCloser
	dec   *xml.Decoder
	err   error
	cur   *Row
}

// Row is one non-empty row from the worksheet (at least one cell).
type Row struct {
	sheet *Sheet
	index int
	cells []*Cell
}

// Index returns the 1-based row index from the row r attribute.
func (r *Row) Index() int {
	if r == nil {
		return 0
	}
	return r.index
}

// Cells returns cells left-to-right in this row.
func (r *Row) Cells() []*Cell {
	if r == nil {
		return nil
	}
	return r.cells
}

// Rows starts a streaming row iterator for the sheet.
func (s *Sheet) Rows() (*Rows, error) {
	if s == nil || s.wb == nil {
		return nil, ErrMissingMainPart
	}
	if s.ws != nil {
		return nil, ErrReadOnlySheet
	}
	rc, err := s.wb.pkg.OpenReader(s.part)
	if err != nil {
		return nil, err
	}
	dec := xml.NewDecoder(rc)
	dec.Strict = false
	return &Rows{sheet: s, rc: rc, dec: dec}, nil
}

// Next advances to the next non-empty row. Returns false on EOF or error (see Err).
func (rs *Rows) Next() bool {
	if rs == nil || rs.dec == nil {
		return false
	}
	for {
		tok, err := rs.dec.Token()
		if err == io.EOF {
			rs.cur = nil
			return false
		}
		if err != nil {
			rs.err = err
			rs.cur = nil
			return false
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "row" {
			continue
		}
		var rd rowDecode
		if err := rs.dec.DecodeElement(&rd, &se); err != nil {
			rs.err = err
			rs.cur = nil
			return false
		}
		if len(rd.C) == 0 {
			continue
		}
		rowIdx, _ := strconv.Atoi(strings.TrimSpace(rd.R))
		if rowIdx < 1 {
			rowIdx = 0
		}
		cells := make([]*Cell, 0, len(rd.C))
		for i := range rd.C {
			cd := cellDecToCellData(&rd.C[i], rowIdx, i+1)
			cell, err := newCellFromData(rs.sheet.wb, rs.sheet, &cd)
			if err != nil {
				rs.err = err
				rs.cur = nil
				return false
			}
			cells = append(cells, cell)
		}
		if rowIdx < 1 && len(cells) > 0 && cells[0].Address() != "" {
			_, rowIdx, _ = sml.CellRefToIndexes(cells[0].Address())
		}
		rs.cur = &Row{sheet: rs.sheet, index: rowIdx, cells: cells}
		return true
	}
}

// Row returns the current row after a successful Next.
func (rs *Rows) Row() *Row {
	if rs == nil {
		return nil
	}
	return rs.cur
}

// Err returns a non-EOF error from Next or streaming.
func (rs *Rows) Err() error {
	if rs == nil {
		return nil
	}
	return rs.err
}

// Close releases the underlying worksheet reader.
func (rs *Rows) Close() error {
	if rs == nil || rs.rc == nil {
		return nil
	}
	err := rs.rc.Close()
	rs.rc = nil
	rs.dec = nil
	return err
}

// duplicate rowDecode from sml - we need same structs in xlsx OR export from sml

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

func cellDecToCellData(c *cellDec, rowIdx, colGuess int) sml.CellData {
	// duplicate logic from sml.cellDecToData - call sml by exporting or duplicate
	// Import cycle if sml imports xlsx - so duplicate minimal:
	ref := trim(c.R)
	style := -1
	if v := trim(c.S); v != "" {
		if n, err := parseInt(v); err == nil {
			style = n
		}
	}
	raw := ""
	if c.V != nil {
		raw = trim(c.V.Inner)
	}
	formula := ""
	if c.F != nil {
		formula = trim(c.F.Inner)
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
		ref = sml.IndexesToCellRef(colGuess, rowIdx)
	}
	return sml.CellData{
		Ref: ref, T: trim(c.T), S: style, Raw: raw, Formula: formula, InlinePlain: inline,
	}
}

func trim(s string) string { return strings.TrimSpace(s) }

func parseInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}
