package ooxml

import (
	"strings"
	"testing"
)

func TestParseContentTypes(t *testing.T) {
	const raw = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`
	ct, err := ParseContentTypes(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if !ct.HasContentType(CTWordDocumentMain) {
		t.Fatal("expected word main content type")
	}
	if got := ct.PartNameForContentType(CTWordDocumentMain); got != "/word/document.xml" {
		t.Fatalf("part name: got %q", got)
	}
}

func TestResolveTarget(t *testing.T) {
	if got := ResolveTarget("/_rels/.rels", "word/document.xml"); got != "/word/document.xml" {
		t.Fatalf("root rels: got %q", got)
	}
	if got := ResolveTarget("/word/_rels/document.xml.rels", "document.xml"); got != "/word/document.xml" {
		t.Fatalf("part rels: got %q", got)
	}
}
