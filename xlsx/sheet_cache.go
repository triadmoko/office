package xlsx

import (
	"bytes"
	"fmt"

	"github.com/triadmoko/office/internal/sml"
)

const maxRandomSheetBytes = 100 << 20

func (s *Sheet) ensureRandomCells() error {
	if s == nil || s.wb == nil {
		return ErrMissingMainPart
	}
	if s.ws != nil {
		return ErrReadOnlySheet
	}
	s.randOnce.Do(s.loadRandomCells)
	return s.randErr
}

func (s *Sheet) loadRandomCells() {
	body, err := s.wb.pkg.ReadFile(s.part)
	if err != nil {
		s.randErr = fmt.Errorf("xlsx: read sheet: %w", err)
		return
	}
	if int64(len(body)) > maxRandomSheetBytes {
		s.randErr = ErrSheetTooLargeRandom
		return
	}
	if err := s.wb.ensureStyles(); err != nil {
		s.randErr = err
		return
	}
	if err := s.wb.ensureSharedStrings(); err != nil {
		s.randErr = err
		return
	}
	m := make(map[string]*Cell)
	err = sml.StreamWorksheetRows(bytes.NewReader(body), func(rd sml.RowData) error {
		for i := range rd.Cells {
			cd := &rd.Cells[i]
			can, err := canonicalCellRef(cd.Ref)
			if err != nil {
				return err
			}
			cell, err := newCellFromData(s.wb, s, cd)
			if err != nil {
				return err
			}
			m[can] = cell
		}
		return nil
	})
	if err != nil {
		s.randErr = err
		return
	}
	s.cells = m
}

func (s *Sheet) ensureLayout() error {
	if s == nil || s.wb == nil {
		return ErrMissingMainPart
	}
	if s.ws != nil {
		return ErrReadOnlySheet
	}
	s.layOnce.Do(s.loadLayout)
	return s.layErr
}

func (s *Sheet) loadLayout() {
	body, err := s.wb.pkg.ReadFile(s.part)
	if err != nil {
		s.layErr = fmt.Errorf("xlsx: read sheet layout: %w", err)
		return
	}
	lay, err := sml.ScanWorksheetLayout(bytes.NewReader(body))
	if err != nil {
		s.layErr = err
		return
	}
	s.layout = lay
}
