package docx

import (
	"bytes"
	"strings"

	"github.com/triadmoko/office/internal/ooxml"
)

// documentRelsPath returns /word/_rels/document.xml.rels for main /word/document.xml.
func documentRelsPath(main string) string {
	main = strings.TrimPrefix(main, "/")
	i := strings.LastIndex(main, "/")
	if i < 0 {
		return "/_rels/document.xml.rels"
	}
	dir := main[:i]
	file := main[i+1:]
	return "/" + dir + "/_rels/" + file + ".rels"
}

func lookupRelatedPart(pkg *ooxml.Package, mainPart, relType string) (path string, body []byte, ok bool, err error) {
	relsPath := documentRelsPath(mainPart)
	data, err := pkg.ReadFile(relsPath)
	if err != nil {
		return "", nil, false, nil
	}
	rels, err := ooxml.ParseRelationships(bytes.NewReader(data))
	if err != nil {
		return "", nil, false, err
	}
	r := rels.ByType(relType)
	if r == nil || r.Target == "" {
		return "", nil, false, nil
	}
	path, err = ooxml.ResolveTarget(relsPath, r.Target)
	if err != nil {
		return "", nil, false, err
	}
	body, err = pkg.ReadFile(path)
	if err != nil || body == nil {
		return path, nil, false, nil
	}
	return path, body, true, nil
}
