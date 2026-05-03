package xlsx

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/triadmoko/office/internal/ooxml"
)

// Save writes the workbook to w. New workbooks use the builder serializer; opened workbooks
// re-emit a snapshot of all package parts (round-trip).
func (w *Workbook) Save(out io.Writer) error {
	if w == nil {
		return ErrMissingMainPart
	}
	if w.fromNew {
		return w.saveNew(out)
	}
	return w.saveOpened(out)
}

// SaveFile writes the workbook to a file path.
func (w *Workbook) SaveFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := w.Save(f); err != nil {
		return err
	}
	return f.Close()
}

func (w *Workbook) saveOpened(out io.Writer) error {
	if len(w.partData) == 0 {
		return fmt.Errorf("xlsx: no part snapshot for Save (internal error)")
	}
	pw := ooxml.NewPackageWriter(out)
	paths := sortedPartKeys(w.partData)
	for _, path := range paths {
		if isContentTypesPart(path) {
			continue
		}
		data := w.partData[path]
		if isRelsPart(path) {
			rels, err := ooxml.ParseRelationships(bytes.NewReader(data))
			if err != nil {
				return fmt.Errorf("xlsx: parse %s: %w", path, err)
			}
			owner, err := ownerPartForRelsPart(path)
			if err != nil {
				return err
			}
			if err := pw.AddRelationships(owner, rels); err != nil {
				return fmt.Errorf("xlsx: relationships %s: %w", path, err)
			}
			continue
		}
		pn, err := ooxml.NormalizePartName(path)
		if err != nil {
			return err
		}
		ct := contentTypeForPartX(w.origCT, pn)
		if ct == "" {
			ct = guessContentTypeX(pn)
		}
		if err := pw.AddPartBytes(pn, ct, data); err != nil {
			return fmt.Errorf("xlsx: part %s: %w", pn, err)
		}
	}
	return pw.Close()
}

func sortedPartKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
		return "", fmt.Errorf("xlsx: invalid rels path %q", relsPath)
	}
	dir := rp[:i]
	file := rp[i+len(seg):]
	if !strings.HasSuffix(file, ".rels") {
		return "", fmt.Errorf("xlsx: invalid rels path %q", relsPath)
	}
	base := strings.TrimSuffix(file, ".rels")
	if dir == "" || dir == "/" {
		return ooxml.NormalizePartName("/" + base)
	}
	return ooxml.NormalizePartName(dir + "/" + base)
}

func contentTypeForPartX(ct *ooxml.ContentTypes, part string) string {
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

func guessContentTypeX(part string) string {
	p, _ := ooxml.NormalizePartName(part)
	switch p {
	case "/xl/workbook.xml":
		return ooxml.CTSpreadsheetMain
	case "/xl/sharedStrings.xml":
		return ctSharedStrings
	case "/xl/styles.xml":
		return ctStyles
	case "/docProps/core.xml":
		return ooxml.CTCoreProps
	case "/docProps/app.xml":
		return ooxml.CTAppProps
	default:
		if strings.Contains(p, "/xl/worksheets/") && strings.HasSuffix(p, ".xml") {
			return ctWorksheet
		}
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
	i := strings.LastIndex(part, ".")
	if i < 0 {
		return ""
	}
	return strings.ToLower(part[i+1:])
}
