package docx

import (
	"bytes"
	"io"
	"strings"
	"sync"

	"github.com/triadmoko/office/internal/ooxml"
	"github.com/triadmoko/office/internal/wml"
)

const (
	relTypeStyles    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"
	relTypeNumbering = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering"
	relTypeFooter    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer"
	relTypeHeader    = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/header"
)

// Document is an opened or newly built .docx (WordprocessingML) package.
type Document struct {
	pkg  *ooxml.Package
	main string

	once      sync.Once
	wmlDoc    *wml.Document
	styles    *wml.Styles
	numbering *wml.Numbering
	loadErr   error

	partData map[string][]byte // snapshot of all parts when opened (OFFICE-109)
	origCT   *ooxml.ContentTypes

	fromNew bool

	// footerPageNumber: when true, Save (NewDocument only) emits /word/footer1.xml with PAGE field and footerReference in sectPr.
	footerPageNumber bool
	// footerPageTemplate: literal + placeholders [FooterPlaceholderPage] / [FooterPlaceholderNumPages]; kosong = default "Hal. {{PAGE}}".
	footerPageTemplate string

	// headerPageNumber: when true, Save (NewDocument only) emits /word/header1.xml with PAGE field and headerReference in sectPr.
	headerPageNumber bool
	// headerPageTemplate: same placeholders as footer; kosong = default "Hal. {{PAGE}}" (rata tengah).
	headerPageTemplate string

	// stripLayoutHints: when true, [Save] omits w:lastRenderedPageBreak from serialized runs.
	stripLayoutHints bool
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

// SetFooterPageNumber enables a default footer with a PAGE field (nomor halaman) on Save.
// MVP: only for documents from [NewDocument]; [Open] returns [ErrFooterPageNumberOpenDoc] if this is set when saving.
func (d *Document) SetFooterPageNumber(v bool) {
	if d == nil {
		return
	}
	d.footerPageNumber = v
}

// SetFooterPageNumberTemplate mengatur teks footer bila [SetFooterPageNumber](true): gabungan teks biasa
// dan placeholder (case-sensitive) [FooterPlaceholderPage] (halaman saat ini) serta [FooterPlaceholderNumPages] (jumlah halaman).
// Contoh: "No. {{PAGE}}", "Page {{PAGE}} of {{NUMPAGES}}", "Hal. {{PAGE}} / {{NUMPAGES}}".
// Jika string kosong (atau hanya spasi) saat Save dengan footer aktif, dipakai default "Hal. {{PAGE}}".
func (d *Document) SetFooterPageNumberTemplate(layout string) {
	if d == nil {
		return
	}
	d.footerPageTemplate = layout
}

// SetHeaderPageNumber enables a default header with PAGE (and optional NUMPAGES) fields on Save.
// MVP: only for documents from [NewDocument]; [Open] returns [ErrHeaderPageNumberOpenDoc] if set when saving.
func (d *Document) SetHeaderPageNumber(v bool) {
	if d == nil {
		return
	}
	d.headerPageNumber = v
}

// SetHeaderPageNumberTemplate sets header text when [SetHeaderPageNumber](true).
// Uses the same placeholders as the footer: [HeaderPlaceholderPage] (= [FooterPlaceholderPage]) and [HeaderPlaceholderNumPages].
// Default layout when empty: "Hal. {{PAGE}}" (paragraph alignment center).
func (d *Document) SetHeaderPageNumberTemplate(layout string) {
	if d == nil {
		return
	}
	d.headerPageTemplate = layout
}

// SetStripLayoutHints controls whether Word layout hints (w:lastRenderedPageBreak) are written on [Save].
// When true, those elements are omitted from runs; pagination instructions (w:br, w:sectPr, etc.) are unchanged.
func (d *Document) SetStripLayoutHints(v bool) {
	if d == nil {
		return
	}
	d.stripLayoutHints = v
}

// NewDocument returns an empty in-memory document ready for the builder API and Save.
func NewDocument() *Document {
	return &Document{
		main:      "/word/document.xml",
		wmlDoc:    wml.EmptyDocument(),
		styles:    wml.DefaultStyles(),
		numbering: &wml.Numbering{Abstract: make(map[int]*wml.AbstractNum), Nums: make(map[int]*wml.NumDef)},
		fromNew:   true,
	}
}

func (d *Document) loadFromPackage() {
	if d == nil || d.fromNew {
		return
	}
	if d.pkg == nil {
		d.loadErr = ErrMissingMainPart
		return
	}
	names, err := d.pkg.FileNames()
	if err != nil {
		d.loadErr = err
		return
	}
	d.partData = make(map[string][]byte)
	for _, n := range names {
		b, err := d.pkg.ReadFile(n)
		if err != nil {
			d.loadErr = err
			return
		}
		nb := make([]byte, len(b))
		copy(nb, b)
		d.partData[n] = nb
	}
	for k, v := range d.partData {
		if strings.HasSuffix(strings.ToLower(strings.TrimPrefix(k, "/")), "[content_types].xml") {
			d.origCT, _ = ooxml.ParseContentTypes(bytes.NewReader(v))
			break
		}
	}
	mainBody := d.partData[d.main]
	if len(mainBody) == 0 {
		d.loadErr = ErrMissingMainPart
		return
	}
	d.wmlDoc, err = wml.ParseDocument(bytes.NewReader(mainBody))
	if err != nil {
		d.loadErr = err
		return
	}
	if _, data, ok, err := lookupRelatedPart(d.pkg, d.main, relTypeStyles); err == nil && ok {
		st, err := wml.ParseStyles(bytes.NewReader(data))
		if err == nil {
			d.styles = st
		}
	}
	if d.styles == nil {
		d.styles = wml.DefaultStyles()
	}
	if _, data, ok, err := lookupRelatedPart(d.pkg, d.main, relTypeNumbering); err == nil && ok {
		num, err := wml.ParseNumbering(bytes.NewReader(data))
		if err == nil {
			d.numbering = num
		}
	}
}

func (d *Document) ensureLoaded() (*wml.Document, error) {
	if d == nil {
		return nil, ErrMissingMainPart
	}
	if d.fromNew {
		if d.wmlDoc == nil {
			d.wmlDoc = wml.EmptyDocument()
		}
		if d.styles == nil {
			d.styles = wml.DefaultStyles()
		}
		if d.numbering == nil {
			d.numbering = &wml.Numbering{Abstract: make(map[int]*wml.AbstractNum), Nums: make(map[int]*wml.NumDef)}
		}
		return d.wmlDoc, nil
	}
	d.once.Do(d.loadFromPackage)
	if d.loadErr != nil {
		return nil, d.loadErr
	}
	return d.wmlDoc, nil
}

// MainPart returns the package part path of the main document (e.g. /word/document.xml).
func (d *Document) MainPart() string {
	if d == nil {
		return ""
	}
	return d.main
}

// Package exposes the underlying OPC package for advanced use (nil for NewDocument until saved).
func (d *Document) Package() *ooxml.Package {
	if d == nil {
		return nil
	}
	return d.pkg
}

// PartBytes returns a snapshot of raw part bytes after Open (OFFICE-109); nil for NewDocument.
func (d *Document) PartBytes() map[string][]byte {
	if d == nil {
		return nil
	}
	_, _ = d.ensureLoaded()
	if d.partData == nil {
		return nil
	}
	out := make(map[string][]byte, len(d.partData))
	for k, v := range d.partData {
		cp := make([]byte, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
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
	return ooxml.ResolveTarget("/_rels/.rels", r.Target)
}
