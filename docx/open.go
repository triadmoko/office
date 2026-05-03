package docx

import (
	"io"

	"github.com/triadmoko/office/internal/ooxml"
)

// Document is an opened .docx (WordprocessingML) package.
type Document struct {
	pkg  *ooxml.Package
	main string // e.g. /word/document.xml
}

// Open opens a DOCX from a ZIP-backed reader (e.g. *os.File or bytes.NewReader data).
func Open(ra io.ReaderAt, size int64) (*Document, error) {
	pkg, err := ooxml.Open(ra, size)
	if err != nil {
		return nil, err
	}
	main, err := resolveMainDocument(pkg)
	if err != nil {
		return nil, err
	}
	if !pkg.HasPart(main) {
		return nil, ErrMissingMainPart
	}
	return &Document{pkg: pkg, main: main}, nil
}

func resolveMainDocument(pkg *ooxml.Package) (string, error) {
	ct := pkg.ContentTypes()
	if ct != nil && ct.HasContentType(ooxml.CTWordDocumentMain) {
		pn := ct.PartNameForContentType(ooxml.CTWordDocumentMain)
		if pn != "" {
			return pn, nil
		}
	}
	rels, err := pkg.RootRelationships()
	if err != nil {
		return "", ErrNotDOCX
	}
	r := rels.ByType(ooxml.NSRelOfficeDocument)
	if r == nil || r.Target == "" {
		return "", ErrNotDOCX
	}
	return ooxml.ResolveTarget("/_rels/.rels", r.Target), nil
}

// MainPart returns the package part path of the main document (e.g. /word/document.xml).
func (d *Document) MainPart() string {
	if d == nil {
		return ""
	}
	return d.main
}

// Package exposes the underlying OPC package for advanced use.
func (d *Document) Package() *ooxml.Package {
	if d == nil {
		return nil
	}
	return d.pkg
}
