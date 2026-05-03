package xlsx

import (
	"io"

	"github.com/triadmoko/office/internal/ooxml"
)

// Workbook is an opened .xlsx (SpreadsheetML) package.
type Workbook struct {
	pkg  *ooxml.Package
	main string // e.g. /xl/workbook.xml
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
	return &Workbook{pkg: pkg, main: main}, nil
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
