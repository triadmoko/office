package docx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/triadmoko/office/internal/ooxml"
)

// Save writes a complete .docx package to w (PackageWriter + merged parts when opened from disk).
func (d *Document) Save(w io.Writer) error {
	if d == nil {
		return ErrMissingMainPart
	}
	if _, err := d.ensureLoaded(); err != nil {
		return err
	}
	if d.footerPageNumber && !d.fromNew {
		return ErrFooterPageNumberOpenDoc
	}
	numXML := MarshalNumberingXML(d.numbering)
	withNum := numXML != nil
	footerRID := ""
	if d.footerPageNumber {
		footerRID = documentFooterRelID(withNum)
	}
	docXML, err := MarshalDocumentXML(d.wmlDoc, MarshalDocumentOpts{FooterRelationshipID: footerRID})
	if err != nil {
		return fmt.Errorf("docx save: marshal document: %w", err)
	}
	styXML, err := MarshalStylesXML(d.styles)
	if err != nil {
		return fmt.Errorf("docx save: marshal styles: %w", err)
	}

	pw := ooxml.NewPackageWriter(w)

	override := map[string][]byte{
		d.main:             docXML,
		"/word/styles.xml": styXML,
	}
	if numXML != nil {
		override["/word/numbering.xml"] = numXML
	}
	if d.footerPageNumber {
		override["/word/footer1.xml"] = marshalFooterPageXML(d.footerPageTemplate)
	}
	override["/docProps/core.xml"] = marshalCoreProps()
	override["/docProps/app.xml"] = marshalAppProps()

	if len(d.partData) > 0 {
		// Opened package: preserve unrelated parts; replace main, styles, numbering, props.
		paths := sortedKeys(d.partData)
		for _, path := range paths {
			data := d.partData[path]
			if isContentTypesPart(path) {
				continue
			}
			if isRelsPart(path) {
				rels, err := ooxml.ParseRelationships(bytes.NewReader(data))
				if err != nil {
					return fmt.Errorf("docx save: parse %s: %w", path, err)
				}
				owner, err := ownerPartForRelsPart(path)
				if err != nil {
					return err
				}
				if err := pw.AddRelationships(owner, rels); err != nil {
					return fmt.Errorf("docx save: relationships %s: %w", path, err)
				}
				continue
			}
			if _, ok := override[normalizePart(path)]; ok {
				continue
			}
			ct := contentTypeForPart(d.origCT, path)
			if ct == "" {
				ct = guessContentType(path)
			}
			if err := pw.AddPartBytes(path, ct, data); err != nil {
				return fmt.Errorf("docx save: part %s: %w", path, err)
			}
		}
	} else {
		// New in-memory document.
		if err := pw.AddRelationships("", newRootRels()); err != nil {
			return err
		}
		if err := pw.AddRelationships(d.main, newDocumentRels(numXML != nil, d.footerPageNumber)); err != nil {
			return err
		}
	}

	for _, path := range sortedKeys(override) {
		data := override[path]
		pn, err := ooxml.NormalizePartName(path)
		if err != nil {
			return err
		}
		ct := contentTypeForPart(d.origCT, pn)
		if ct == "" {
			ct = guessContentType(pn)
		}
		if err := pw.AddPartBytes(pn, ct, data); err != nil {
			return fmt.Errorf("docx save: add %s: %w", pn, err)
		}
	}

	return pw.Close()
}

func sortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// SaveFile writes the package to path (see Save).
func (d *Document) SaveFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := d.Save(f); err != nil {
		return err
	}
	return f.Close()
}

func normalizePart(p string) string {
	n, err := ooxml.NormalizePartName(p)
	if err != nil {
		return p
	}
	return n
}

func isContentTypesPart(p string) bool {
	return strings.HasSuffix(strings.ToLower(strings.TrimPrefix(p, "/")), "[content_types].xml")
}

func isRelsPart(p string) bool {
	return strings.Contains(p, "/_rels/") && strings.HasSuffix(p, ".rels")
}

func ownerPartForRelsPart(relsPath string) (string, error) {
	rp, err := ooxml.NormalizePartName(relsPath)
	if err != nil {
		return "", err
	}
	if rp == "/_rels/.rels" {
		return "", nil
	}
	const seg = "/_rels/"
	i := strings.LastIndex(rp, seg)
	if i < 0 {
		return "", fmt.Errorf("docx: invalid rels path %q", relsPath)
	}
	dir := rp[:i]
	file := rp[i+len(seg):]
	if !strings.HasSuffix(file, ".rels") {
		return "", fmt.Errorf("docx: invalid rels path %q", relsPath)
	}
	base := strings.TrimSuffix(file, ".rels")
	if dir == "" || dir == "/" {
		return ooxml.NormalizePartName("/" + base)
	}
	return ooxml.NormalizePartName(dir + "/" + base)
}

func contentTypeForPart(ct *ooxml.ContentTypes, part string) string {
	if ct == nil {
		return ""
	}
	pn, err := ooxml.NormalizePartName(part)
	if err != nil {
		return ""
	}
	for _, o := range ct.Override {
		op, err := ooxml.NormalizePartName(o.PartName)
		if err != nil {
			continue
		}
		if op == pn {
			return o.ContentType
		}
	}
	return ""
}

func guessContentType(part string) string {
	p, _ := ooxml.NormalizePartName(part)
	switch p {
	case "/word/document.xml":
		return ooxml.CTWordDocumentMain
	case "/word/styles.xml":
		return ooxml.CTWordStyles
	case "/word/numbering.xml":
		return ooxml.CTWordNumbering
	case "/word/settings.xml":
		return ooxml.CTWordSettings
	case "/word/fontTable.xml":
		return ooxml.CTWordFontTable
	case "/word/webSettings.xml":
		return ooxml.CTWordWebSettings
	case "/word/footer1.xml":
		return ooxml.CTWordFooter
	case "/docProps/core.xml":
		return ooxml.CTCoreProps
	case "/docProps/app.xml":
		return ooxml.CTAppProps
	default:
		if strings.HasSuffix(p, ".rels") {
			return ooxml.CTRelsXML
		}
		if ext := extensionOfPart(p); ext != "" {
			if ct, ok := ooxml.ExtensionToContentType[ext]; ok {
				return ct
			}
		}
	}
	return "application/xml"
}

func extensionOfPart(part string) string {
	part = strings.TrimPrefix(part, "/")
	i := strings.LastIndex(part, ".")
	if i < 0 {
		return ""
	}
	return strings.ToLower(part[i+1:])
}

func newRootRels() *ooxml.Relationships {
	return &ooxml.Relationships{
		Relationship: []ooxml.Relationship{
			{ID: "rId1", Type: ooxml.NSRelOfficeDocument, Target: "word/document.xml"},
			{ID: "rId2", Type: "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties", Target: "docProps/core.xml"},
			{ID: "rId3", Type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties", Target: "docProps/app.xml"},
		},
	}
}
