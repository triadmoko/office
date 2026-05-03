package xlsx

import (
	"strings"
	"sync"

	"github.com/triadmoko/office/internal/sml"
)

// SheetState is the visibility of a sheet tab in Excel.
type SheetState int

const (
	// SheetVisible is the default tab visibility.
	SheetVisible SheetState = iota
	// SheetHidden hides the sheet (can be unhidden via UI).
	SheetHidden
	// SheetVeryHidden hides the sheet (VBA / strict hide).
	SheetVeryHidden
)

// Sheet is a worksheet in the workbook (read side).
type Sheet struct {
	wb      *Workbook
	name    string
	sheetID int
	rid     string
	state   SheetState
	part    string // OPC part path, e.g. /xl/worksheets/sheet1.xml

	randOnce sync.Once
	randErr  error
	cells    map[string]*Cell // canonical "A1" -> cell

	layOnce sync.Once
	layErr  error
	layout  *sml.WorksheetLayout

	ws *writeSheet // non-nil for sheets created via [Workbook.AddSheet] on a new workbook
}

// Name returns the sheet tab name.
func (s *Sheet) Name() string {
	if s == nil {
		return ""
	}
	return s.name
}

// SheetID returns the workbook sheetId attribute (1-based in typical files).
func (s *Sheet) SheetID() int {
	if s == nil {
		return 0
	}
	return s.sheetID
}

// RelationshipID returns the r:id for this sheet in workbook.xml.rels.
func (s *Sheet) RelationshipID() string {
	if s == nil {
		return ""
	}
	return s.rid
}

// State returns tab visibility.
func (s *Sheet) State() SheetState {
	if s == nil {
		return SheetVisible
	}
	return s.state
}

// Part returns the OPC part path of the worksheet XML.
func (s *Sheet) Part() string {
	if s == nil {
		return ""
	}
	return s.part
}

// Range is an inclusive rectangular cell region (1-based column/row indices).
type Range struct {
	MinCol, MinRow, MaxCol, MaxRow int
}

// Cell returns a cell by reference ("A1", "$B$2"). Missing cells return an empty cell.
func (s *Sheet) Cell(ref string) (*Cell, error) {
	can, err := canonicalCellRef(ref)
	if err != nil {
		return nil, err
	}
	if err := s.ensureRandomCells(); err != nil {
		return nil, err
	}
	if c, ok := s.cells[can]; ok {
		return c, nil
	}
	col, row, err := sml.CellRefToIndexes(can)
	if err != nil {
		return nil, err
	}
	return &Cell{wb: s.wb, sheet: s, addr: can, row: row, col: col, typ: CellEmpty}, nil
}

// CellAt returns a cell by 1-based Excel column and row indices.
func (s *Sheet) CellAt(row, col int) (*Cell, error) {
	if row < 1 || col < 1 {
		return nil, ErrInvalidCellRef
	}
	return s.Cell(sml.IndexesToCellRef(col, row))
}

// MergedRanges returns merged regions from the worksheet.
func (s *Sheet) MergedRanges() ([]Range, error) {
	if err := s.ensureLayout(); err != nil {
		return nil, err
	}
	var out []Range
	for _, ref := range s.layout.MergedRefs {
		mc, mr, xc, xr, err := sml.ParseCellRange(ref)
		if err != nil {
			continue
		}
		out = append(out, Range{MinCol: mc, MinRow: mr, MaxCol: xc, MaxRow: xr})
	}
	return out, nil
}

// HiddenRows returns 1-based row indices marked hidden.
func (s *Sheet) HiddenRows() ([]int, error) {
	if err := s.ensureLayout(); err != nil {
		return nil, err
	}
	return append([]int(nil), s.layout.HiddenRows...), nil
}

// HiddenCols returns 1-based column indices marked hidden.
func (s *Sheet) HiddenCols() ([]int, error) {
	if err := s.ensureLayout(); err != nil {
		return nil, err
	}
	return append([]int(nil), s.layout.HiddenCols...), nil
}

// FreezePane returns frozen split rows/columns or nil.
func (s *Sheet) FreezePane() *FreezePane {
	if s == nil {
		return nil
	}
	if err := s.ensureLayout(); err != nil || s.layout == nil || s.layout.Freeze == nil {
		return nil
	}
	return &FreezePane{Row: s.layout.Freeze.Row, Col: s.layout.Freeze.Col}
}

// RowHeight returns custom row height in points, or defaultHeight if unset.
func (s *Sheet) RowHeight(row int, defaultHeight float64) float64 {
	if s == nil || row < 1 {
		return defaultHeight
	}
	if err := s.ensureLayout(); err != nil || s.layout == nil {
		return defaultHeight
	}
	if h, ok := s.layout.RowHeights[row]; ok {
		return h
	}
	return defaultHeight
}

// ColumnWidth returns custom column width, or defaultWidth if unset.
func (s *Sheet) ColumnWidth(col int, defaultWidth float64) float64 {
	if s == nil || col < 1 {
		return defaultWidth
	}
	if err := s.ensureLayout(); err != nil || s.layout == nil {
		return defaultWidth
	}
	if w, ok := s.layout.ColumnWidths[col]; ok {
		return w
	}
	return defaultWidth
}

// FreezePane describes a frozen pane split (read).
type FreezePane struct {
	Row int
	Col int
}

func canonicalCellRef(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	ref = strings.ReplaceAll(ref, "$", "")
	if ref == "" {
		return "", ErrInvalidCellRef
	}
	col, row, err := sml.CellRefToIndexes(ref)
	if err != nil {
		return "", err
	}
	return sml.IndexesToCellRef(col, row), nil
}

func parseSheetState(raw string) SheetState {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "hidden":
		return SheetHidden
	case "veryhidden":
		return SheetVeryHidden
	default:
		return SheetVisible
	}
}
