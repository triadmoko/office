package xlsx

import (
	"bytes"
	"strings"

	"github.com/triadmoko/office/internal/ooxml"
)

const (
	relTypeWorksheet     = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet"
	relTypeSharedStrings = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/sharedStrings"
)

// workbookRelsPath returns /xl/_rels/workbook.xml.rels for main /xl/workbook.xml.
func workbookRelsPath(main string) string {
	main = strings.TrimPrefix(main, "/")
	i := strings.LastIndex(main, "/")
	if i < 0 {
		return "/_rels/workbook.xml.rels"
	}
	dir := main[:i]
	file := main[i+1:]
	return "/" + dir + "/_rels/" + file + ".rels"
}

func resolveWorksheetParts(pkg *ooxml.Package, workbookMain string) (map[string]string, error) {
	relsPath := workbookRelsPath(workbookMain)
	data, err := pkg.ReadFile(relsPath)
	if err != nil {
		return nil, err
	}
	rels, err := ooxml.ParseRelationships(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for i := range rels.Relationship {
		r := &rels.Relationship[i]
		if r.Type != relTypeWorksheet || r.ID == "" || r.Target == "" {
			continue
		}
		path, err := ooxml.ResolveTarget(relsPath, r.Target)
		if err != nil {
			return nil, err
		}
		out[r.ID] = path
	}
	return out, nil
}

func resolveRelatedPartByType(pkg *ooxml.Package, workbookMain, relType string) (string, bool, error) {
	relsPath := workbookRelsPath(workbookMain)
	data, err := pkg.ReadFile(relsPath)
	if err != nil {
		return "", false, err
	}
	rels, err := ooxml.ParseRelationships(bytes.NewReader(data))
	if err != nil {
		return "", false, err
	}
	for i := range rels.Relationship {
		r := &rels.Relationship[i]
		if r.Type != relType || r.Target == "" {
			continue
		}
		path, err := ooxml.ResolveTarget(relsPath, r.Target)
		if err != nil {
			return "", false, err
		}
		return path, true, nil
	}
	return "", false, nil
}
