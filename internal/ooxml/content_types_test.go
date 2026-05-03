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

func TestResolveTargetTraversal(t *testing.T) {
	cases := []struct {
		rels   string
		target string
		want   string
		errIs  error
	}{
		// TC1: traversal keluar dari root → ErrPathTraversal
		{"/_rels/.rels", "../../etc/passwd", "", ErrPathTraversal},
		// TC2: satu level naik dari depth-2 → legal
		{"/a/b/_rels/c.rels", "../x.xml", "/a/x.xml", nil},
		// TC3: dua level naik dari depth-3 → legal
		{"/a/b/c/_rels/d.rels", "../../x.xml", "/a/x.xml", nil},
		// TC4: leading "./" → diabaikan, resolve normal
		{"/_rels/.rels", "./word/document.xml", "/word/document.xml", nil},
		// TC5: absolute target → abaikan base dir
		{"/_rels/.rels", "/ppt/presentation.xml", "/ppt/presentation.xml", nil},
		// TC6: empty target → root
		{"/_rels/.rels", "", "/", nil},
		// TC7: ".." dari depth-1 melewati root → ErrPathTraversal
		{"/a/_rels/b.rels", "../../x.xml", "", ErrPathTraversal},
		// TC8: terlalu banyak ".." dari depth-1 → ErrPathTraversal
		{"/word/_rels/doc.rels", "../../../etc/x", "", ErrPathTraversal},
	}
	for _, tc := range cases {
		got, err := ResolveTarget(tc.rels, tc.target)
		if tc.errIs != nil {
			if !errors.Is(err, tc.errIs) {
				t.Errorf("ResolveTarget(%q, %q): want error %v, got %v", tc.rels, tc.target, tc.errIs, err)
			}
			continue
		}
		if err != nil {
			t.Errorf("ResolveTarget(%q, %q): unexpected error %v", tc.rels, tc.target, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ResolveTarget(%q, %q): got %q, want %q", tc.rels, tc.target, got, tc.want)
		}
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
