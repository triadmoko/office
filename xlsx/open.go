package xlsx

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/triadmoko/office/internal/ooxml"
	"github.com/triadmoko/office/internal/sml"
)

// Workbook is an opened .xlsx (SpreadsheetML) package.
type Workbook struct {
	pkg  *ooxml.Package
	main string // e.g. /xl/workbook.xml

	fromNew bool

	// In-memory workbook (from [NewWorkbook]).
	newSheets []*Sheet
	styleReg  *styleRegistry

	// Snapshot of all parts after [Open] (round-trip / preserve).
	partData map[string][]byte
	origCT   *ooxml.ContentTypes

	metaOnce sync.Once
	metaErr  error
	sheets   []*Sheet
	names    []*DefinedName

	sharedOnce    sync.Once
	sharedErr     error
	sharedStrings []string

	stylesOnce sync.Once
	stylesErr  error
	styles     *sml.StylesTable
}

// Open opens an XLSX package for validation and future APIs.
func Open(ra io.ReaderAt, size int64) (*Workbook, error) {
	pkg, err := ooxml.Open(ra, size)
	if err != nil {
		return nil, err
	}
	main, err := resolveMainWorkbook(pkg)
	if err != nil {
		return nil, err
	}
	if !pkg.HasPart(main) {
		return nil, ErrMissingMainPart
	}
	wb := &Workbook{pkg: pkg, main: main}
	if err := wb.snapshotParts(); err != nil {
		return nil, err
	}
	return wb, nil
}

func (w *Workbook) ensureMeta() error {
	if w == nil {
		return ErrMissingMainPart
	}
	if w.fromNew {
		return nil
	}
	w.metaOnce.Do(w.loadMeta)
	return w.metaErr
}

func (w *Workbook) snapshotParts() error {
	if w == nil || w.pkg == nil {
		return nil
	}
	names, err := w.pkg.FileNames()
	if err != nil {
		return fmt.Errorf("xlsx: list parts: %w", err)
	}
	w.partData = make(map[string][]byte)
	for _, n := range names {
		b, err := w.pkg.ReadFile(n)
		if err != nil {
			return fmt.Errorf("xlsx: read part %s: %w", n, err)
		}
		cp := make([]byte, len(b))
		copy(cp, b)
		w.partData[n] = cp
		if strings.HasSuffix(strings.ToLower(strings.TrimPrefix(n, "/")), "[content_types].xml") {
			w.origCT, _ = ooxml.ParseContentTypes(bytes.NewReader(cp))
		}
	}
	return nil
}

func (w *Workbook) loadMeta() {
	if w.fromNew {
		return
	}
	body, err := w.pkg.ReadFile(w.main)
	if err != nil {
		w.metaErr = fmt.Errorf("xlsx: read workbook: %w", err)
		return
	}
	sheetDefs, dnDefs, err := sml.ParseWorkbook(bytes.NewReader(body))
	if err != nil {
		w.metaErr = fmt.Errorf("xlsx: parse workbook.xml: %w", err)
		return
	}
	ridToPart, err := resolveWorksheetParts(w.pkg, w.main)
	if err != nil {
		w.metaErr = fmt.Errorf("xlsx: workbook relationships: %w", err)
		return
	}
	w.sheets = make([]*Sheet, 0, len(sheetDefs))
	for _, sd := range sheetDefs {
		part, ok := ridToPart[sd.RID]
		if !ok {
			w.metaErr = fmt.Errorf("xlsx: no worksheet relationship for r:id %q (sheet %q)", sd.RID, sd.Name)
			return
		}
		if !w.pkg.HasPart(part) {
			w.metaErr = fmt.Errorf("xlsx: worksheet part missing %q (sheet %q)", part, sd.Name)
			return
		}
		w.sheets = append(w.sheets, &Sheet{
			wb:      w,
			name:    sd.Name,
			sheetID: sd.SheetID,
			rid:     sd.RID,
			state:   parseSheetState(sd.State),
			part:    part,
		})
	}
	w.names = make([]*DefinedName, 0, len(dnDefs))
	for _, dn := range dnDefs {
		ls := dn.LocalSheetID
		w.names = append(w.names, &DefinedName{
			name:         dn.Name,
			formula:      dn.Formula,
			localSheetID: ls,
		})
	}
}

// Sheets returns worksheets in workbook order.
func (w *Workbook) Sheets() ([]*Sheet, error) {
	if w == nil {
		return nil, ErrMissingMainPart
	}
	if w.fromNew {
		out := make([]*Sheet, len(w.newSheets))
		copy(out, w.newSheets)
		return out, nil
	}
	if err := w.ensureMeta(); err != nil {
		return nil, err
	}
	out := make([]*Sheet, len(w.sheets))
	copy(out, w.sheets)
	return out, nil
}

// SheetByName returns the first sheet whose name matches case-insensitively, or nil if not found.
func (w *Workbook) SheetByName(name string) (*Sheet, error) {
	if w == nil {
		return nil, ErrMissingMainPart
	}
	if w.fromNew {
		for _, s := range w.newSheets {
			if s != nil && strings.EqualFold(s.name, name) {
				return s, nil
			}
		}
		return nil, nil
	}
	if err := w.ensureMeta(); err != nil {
		return nil, err
	}
	for _, s := range w.sheets {
		if s != nil && strings.EqualFold(s.name, name) {
			return s, nil
		}
	}
	return nil, nil
}

// DefinedNames returns defined names from workbook.xml (may be empty).
func (w *Workbook) DefinedNames() ([]*DefinedName, error) {
	if w == nil {
		return nil, ErrMissingMainPart
	}
	if w.fromNew {
		return nil, nil
	}
	if err := w.ensureMeta(); err != nil {
		return nil, err
	}
	out := make([]*DefinedName, len(w.names))
	copy(out, w.names)
	return out, nil
}

func resolveMainWorkbook(pkg *ooxml.Package) (string, error) {
	ct := pkg.ContentTypes()
	if ct != nil && ct.HasContentType(ooxml.CTSpreadsheetMain) {
		pn := ct.PartNameForContentType(ooxml.CTSpreadsheetMain)
		if pn != "" {
			return pn, nil
		}
	}
	rels, err := pkg.RootRelationships()
	if err != nil {
		return "", ErrNotXLSX
	}
	r := rels.ByType(ooxml.NSRelOfficeDocument)
	if r == nil || r.Target == "" {
		return "", ErrNotXLSX
	}
	return ooxml.ResolveTarget("/_rels/.rels", r.Target)
}

// MainPart returns the workbook part path (e.g. /xl/workbook.xml).
func (w *Workbook) MainPart() string {
	if w == nil {
		return ""
	}
	return w.main
}

// Package exposes the underlying OPC package.
func (w *Workbook) Package() *ooxml.Package {
	if w == nil {
		return nil
	}
	return w.pkg
}

// Write is reserved for future SpreadsheetML serialization.
func (w *Workbook) Write(_ io.Writer) error {
	return ErrNotImplemented
}
