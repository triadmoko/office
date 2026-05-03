package pptx

import (
	"io"

	"github.com/triadmoko/office/internal/ooxml"
)

// Presentation is an opened .pptx (PresentationML) package.
type Presentation struct {
	pkg  *ooxml.Package
	main string // e.g. /ppt/presentation.xml
}

// Open opens a PPTX package for validation and future APIs.
func Open(ra io.ReaderAt, size int64) (*Presentation, error) {
	pkg, err := ooxml.Open(ra, size)
	if err != nil {
		return nil, err
	}
	main, err := resolveMainPresentation(pkg)
	if err != nil {
		return nil, err
	}
	if !pkg.HasPart(main) {
		return nil, ErrMissingMainPart
	}
	return &Presentation{pkg: pkg, main: main}, nil
}

func resolveMainPresentation(pkg *ooxml.Package) (string, error) {
	ct := pkg.ContentTypes()
	if ct != nil && ct.HasContentType(ooxml.CTPresentationMain) {
		pn := ct.PartNameForContentType(ooxml.CTPresentationMain)
		if pn != "" {
			return pn, nil
		}
	}
	rels, err := pkg.RootRelationships()
	if err != nil {
		return "", ErrNotPPTX
	}
	r := rels.ByType(ooxml.NSRelOfficeDocument)
	if r == nil || r.Target == "" {
		return "", ErrNotPPTX
	}
	return ooxml.ResolveTarget("/_rels/.rels", r.Target)
}

// MainPart returns the presentation part path (e.g. /ppt/presentation.xml).
func (p *Presentation) MainPart() string {
	if p == nil {
		return ""
	}
	return p.main
}

// Package exposes the underlying OPC package.
func (p *Presentation) Package() *ooxml.Package {
	if p == nil {
		return nil
	}
	return p.pkg
}

// Write is reserved for future PresentationML serialization.
func (p *Presentation) Write(_ io.Writer) error {
	return ErrNotImplemented
}
