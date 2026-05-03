package xlsx

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/triadmoko/office/internal/sml"
)

type writeCell struct {
	formula string
	val     any // string, float64, int, bool, time.Time
	styleID int // cellXfs index, -1 default
	hURL    string
	hDisp   string
}

type writeSheet struct {
	name   string
	id     int
	cells  map[string]*writeCell
	locked bool            // stream writer started
	stream *streamSheetBuf // non-nil after StreamWriter
}

func newWriteSheet(name string, id int) *writeSheet {
	return &writeSheet{name: name, id: id, cells: make(map[string]*writeCell)}
}

func (s *Sheet) mustWrite() (*writeSheet, error) {
	if s == nil || s.wb == nil || !s.wb.fromNew || s.ws == nil {
		return nil, ErrReadOnlySheet
	}
	if s.ws.locked {
		return nil, ErrSheetStreamLocked
	}
	return s.ws, nil
}

// SetCell sets a cell value. Optional last argument may be *StyleBuilder.
func (s *Sheet) SetCell(addr string, values ...any) error {
	ws, err := s.mustWrite()
	if err != nil {
		return err
	}
	can, err := canonicalCellRef(addr)
	if err != nil {
		return err
	}
	var styleID int = -1
	var payload []any
	for _, v := range values {
		if sb, ok := v.(*StyleBuilder); ok {
			sid, err := sb.register()
			if err != nil {
				return err
			}
			styleID = sid
			continue
		}
		payload = append(payload, v)
	}
	if len(payload) != 1 {
		return fmt.Errorf("xlsx: SetCell expects one value plus optional style, got %d values", len(payload))
	}
	ws.cells[can] = &writeCell{val: normalizeSetCellValue(payload[0]), styleID: styleID}
	return nil
}

func normalizeSetCellValue(v any) any {
	switch x := v.(type) {
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case uint:
		return float64(x)
	case uint64:
		return float64(x)
	case *Cell:
		if x == nil {
			return nil
		}
		vv, _ := x.Value()
		return vv
	default:
		return v
	}
}

// SetFormula sets a formula string (not evaluated).
func (s *Sheet) SetFormula(addr, formula string) error {
	ws, err := s.mustWrite()
	if err != nil {
		return err
	}
	can, err := canonicalCellRef(addr)
	if err != nil {
		return err
	}
	c := ws.cells[can]
	if c == nil {
		c = &writeCell{}
		ws.cells[can] = c
	}
	c.formula = formula
	return nil
}

// SetHyperlink sets display text and URL or internal location (e.g. "Sheet2!A1").
func (s *Sheet) SetHyperlink(addr, target, display string) error {
	ws, err := s.mustWrite()
	if err != nil {
		return err
	}
	can, err := canonicalCellRef(addr)
	if err != nil {
		return err
	}
	c := ws.cells[can]
	if c == nil {
		c = &writeCell{}
		ws.cells[can] = c
	}
	c.hURL = strings.TrimSpace(target)
	c.hDisp = strings.TrimSpace(display)
	if c.hDisp == "" {
		c.hDisp = c.hURL
	}
	return nil
}

func timeToExcelSerial(t time.Time) float64 {
	base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	return t.In(time.UTC).Sub(base).Seconds() / 86400.0
}

func floatStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func sortCellAddrs(addrs []string) {
	// sort by row then col
	type key struct{ r, c int }
	keys := make([]key, len(addrs))
	for i, a := range addrs {
		cc, rr, _ := sml.CellRefToIndexes(a)
		keys[i] = key{r: rr, c: cc}
	}
	for i := 1; i < len(addrs); i++ {
		for j := i; j > 0; j-- {
			pj, pk := keys[j-1], keys[j]
			if pj.r > pk.r || (pj.r == pk.r && pj.c > pk.c) {
				addrs[j-1], addrs[j] = addrs[j], addrs[j-1]
				keys[j-1], keys[j] = keys[j], keys[j-1]
			} else {
				break
			}
		}
	}
}
