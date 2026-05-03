package xlsx

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/triadmoko/office/internal/sml"
)

// CellType classifies cell content for read APIs.
type CellType int

const (
	CellEmpty CellType = iota
	CellNumber
	CellString
	CellBoolean
	CellDate
	CellFormula
	CellError
)

// Cell is one cell on a sheet (read iterator).
type Cell struct {
	wb    *Workbook
	sheet *Sheet

	addr     string
	row, col int
	raw      string
	formula  string
	t        string // SpreadsheetML t attribute ("" = number)
	styleIdx int
	typ      CellType
}

// Address returns the cell reference (e.g. "A1").
func (c *Cell) Address() string {
	if c == nil {
		return ""
	}
	return c.addr
}

// Row returns the 1-based row index.
func (c *Cell) Row() int {
	if c == nil {
		return 0
	}
	return c.row
}

// Col returns the 1-based column index.
func (c *Cell) Col() int {
	if c == nil {
		return 0
	}
	return c.col
}

// RawValue returns the text inside <v> if present.
func (c *Cell) RawValue() string {
	if c == nil {
		return ""
	}
	return c.raw
}

// Formula returns the formula text in <f>, or "".
func (c *Cell) Formula() string {
	if c == nil {
		return ""
	}
	return c.formula
}

// Type returns the classified cell type.
func (c *Cell) Type() CellType {
	if c == nil {
		return CellEmpty
	}
	return c.typ
}

// Value returns the cell value: string, float64, bool, time.Time for dates, or error for t="e".
func (c *Cell) Value() (any, error) {
	if c == nil {
		return nil, nil
	}
	switch c.typ {
	case CellEmpty:
		return nil, nil
	case CellString:
		return c.raw, nil
	case CellNumber:
		v, err := strconv.ParseFloat(strings.TrimSpace(c.raw), 64)
		if err != nil {
			return nil, err
		}
		return v, nil
	case CellBoolean:
		return strings.TrimSpace(c.raw) == "1", nil
	case CellDate:
		v, err := strconv.ParseFloat(strings.TrimSpace(c.raw), 64)
		if err != nil {
			return nil, err
		}
		return excelSerialToTime(v), nil
	case CellFormula:
		if c.raw != "" {
			return c.raw, nil
		}
		return nil, nil
	case CellError:
		return nil, errors.New(strings.TrimSpace(c.raw))
	default:
		return c.raw, nil
	}
}

func excelSerialToTime(serial float64) time.Time {
	whole, frac := math.Modf(serial)
	days := int(whole)
	base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	t := base.AddDate(0, 0, days)
	sec := frac * 86400.0
	return t.Add(time.Duration(sec * float64(time.Second)))
}

func cellTypeFromData(wb *Workbook, d *sml.CellData) CellType {
	if d.Formula != "" && d.Raw == "" && d.T != "s" && d.T != "str" && d.InlinePlain == "" && d.T != "inlineStr" && d.T != "b" && d.T != "e" {
		return CellFormula
	}
	switch d.T {
	case "s":
		return CellString
	case "inlineStr":
		return CellString
	case "b":
		return CellBoolean
	case "e":
		return CellError
	case "str":
		return CellString
	default:
		if strings.TrimSpace(d.Raw) == "" && d.InlinePlain == "" && d.Formula == "" {
			return CellEmpty
		}
		if d.T == "" || d.T == "n" {
			if wb != nil {
				st := wb.stylesTable()
				if d.S >= 0 && st.IsDateStyle(d.S) {
					return CellDate
				}
			}
			return CellNumber
		}
		return CellNumber
	}
}

func newCellFromData(wb *Workbook, sh *Sheet, d *sml.CellData) (*Cell, error) {
	col, row, err := sml.CellRefToIndexes(d.Ref)
	if err != nil {
		return nil, fmt.Errorf("xlsx: cell ref %q: %w", d.Ref, err)
	}
	addr := d.Ref
	raw := d.Raw
	if d.T == "inlineStr" {
		raw = d.InlinePlain
	}
	if d.T == "s" {
		idx, err := strconv.Atoi(strings.TrimSpace(d.Raw))
		if err != nil {
			return nil, fmt.Errorf("xlsx: shared string index %q: %w", d.Raw, err)
		}
		s, err := wb.SharedString(idx)
		if err != nil {
			return nil, err
		}
		raw = s
	}
	typ := cellTypeFromData(wb, d)
	if d.Formula != "" && typ != CellString && typ != CellBoolean && typ != CellError {
		if typ == CellNumber || typ == CellDate {
			// formula + cached value — treat as number/date not "formula-only"
		} else if typ == CellEmpty {
			typ = CellFormula
		}
	}
	return &Cell{
		wb:       wb,
		sheet:    sh,
		addr:     addr,
		row:      row,
		col:      col,
		raw:      raw,
		formula:  d.Formula,
		t:        d.T,
		styleIdx: d.S,
		typ:      typ,
	}, nil
}
