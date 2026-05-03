package ooxml

import (
	"errors"
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
	got, err := ResolveTarget("/_rels/.rels", "word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/word/document.xml" {
		t.Fatalf("root rels: got %q", got)
	}
	got, err = ResolveTarget("/word/_rels/document.xml.rels", "document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/word/document.xml" {
		t.Fatalf("part rels: got %q", got)
	}
}

func TestParseContentTypesInvalidOverridePartName(t *testing.T) {
	const raw = `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Override PartName="/word/../document.xml" ContentType="application/xml"/>
</Types>`
	_, err := ParseContentTypes(strings.NewReader(raw))
	if !errors.Is(err, ErrPathTraversal) {
		t.Fatalf("expected ErrPathTraversal, got %v", err)
	}
}
