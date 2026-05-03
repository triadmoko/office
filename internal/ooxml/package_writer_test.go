package ooxml

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
)

const testDocXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
	`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
	`<w:body><w:p><w:r><w:t>Hello</w:t></w:r></w:p></w:body></w:document>`


func buildTestPackage(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	pw := NewPackageWriter(&buf)

	rels := &Relationships{
		Relationship: []Relationship{{
			ID:     "rId1",
			Type:   NSRelOfficeDocument,
			Target: "word/document.xml",
		}},
	}
	if err := pw.AddRelationships("", rels); err != nil {
		t.Fatal(err)
	}
	if err := pw.AddPartBytes("/word/document.xml", CTWordDocumentMain, []byte(testDocXML)); err != nil {
		t.Fatal(err)
	}
	if err := pw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestPackageWriterRoundTrip(t *testing.T) {
	data := buildTestPackage(t)
	pkg, err := Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Open after PackageWriter.Close(): %v", err)
	}
	if !pkg.HasPart("/word/document.xml") {
		t.Error("expected /word/document.xml in opened package")
	}
	if ct := pkg.ContentTypes(); ct == nil {
		t.Error("ContentTypes should not be nil")
	} else if !ct.HasContentType(CTWordDocumentMain) {
		t.Error("missing CTWordDocumentMain in content types")
	}
}

func TestPackageWriterDeterministicHash(t *testing.T) {
	h1 := sha256.Sum256(buildTestPackage(t))
	h2 := sha256.Sum256(buildTestPackage(t))
	if h1 != h2 {
		t.Error("PackageWriter output is not deterministic across two identical builds")
	}
}

func TestPackageWriterContentTypesFirst(t *testing.T) {
	data := buildTestPackage(t)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(zr.File) == 0 {
		t.Fatal("ZIP has no entries")
	}
	if zr.File[0].Name != "[Content_Types].xml" {
		t.Errorf("first ZIP entry: got %q, want [Content_Types].xml", zr.File[0].Name)
	}
}

func TestPackageWriterDuplicatePart(t *testing.T) {
	var buf bytes.Buffer
	pw := NewPackageWriter(&buf)
	if err := pw.AddPartBytes("/word/document.xml", CTWordDocumentMain, []byte("body")); err != nil {
		t.Fatal(err)
	}
	err := pw.AddPartBytes("/word/document.xml", CTWordDocumentMain, []byte("body2"))
	if !errors.Is(err, ErrDuplicatePart) {
		t.Errorf("expected ErrDuplicatePart, got %v", err)
	}
}

func TestPackageWriterInvalidPartName(t *testing.T) {
	var buf bytes.Buffer
	pw := NewPackageWriter(&buf)
	err := pw.AddPartBytes("../etc/passwd", "text/plain", []byte("data"))
	if !errors.Is(err, ErrPathTraversal) {
		t.Errorf("expected ErrPathTraversal, got %v", err)
	}
}

func TestPackageWriterEpochModTime(t *testing.T) {
	data := buildTestPackage(t)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range zr.File {
		if !f.Modified.Equal(zipEpoch) {
			t.Errorf("entry %q has non-epoch ModTime: %v", f.Name, f.Modified)
		}
	}
}

func TestPackageWriterWithRels(t *testing.T) {
	data := buildTestPackage(t)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range zr.File {
		if f.Name == "_rels/.rels" {
			found = true
			break
		}
	}
	if !found {
		t.Error("_rels/.rels not found in ZIP entries")
	}
}

func TestPackageWriterMarshalContentTypes(t *testing.T) {
	ct := &ContentTypes{
		Default: []CTDefault{
			{Extension: "xml", ContentType: "application/xml"},
			{Extension: "rels", ContentType: CTRelsXML},
		},
		Override: []CTOverride{
			{PartName: "/word/document.xml", ContentType: CTWordDocumentMain},
		},
	}
	data, err := marshalContentTypes(ct)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, `<?xml`) {
		t.Error("missing XML declaration")
	}
	if !strings.Contains(s, `Extension="rels"`) {
		t.Error("missing rels Default entry")
	}
	if !strings.Contains(s, `PartName="/word/document.xml"`) {
		t.Error("missing Override PartName")
	}

	// Round-trip: parse the output and verify.
	ct2, err := ParseContentTypes(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseContentTypes after marshal: %v", err)
	}
	if !ct2.HasContentType(CTWordDocumentMain) {
		t.Error("round-trip: missing CTWordDocumentMain")
	}
}

func TestPackageWriterMarshalRelationships(t *testing.T) {
	r := &Relationships{
		Relationship: []Relationship{
			{ID: "rId1", Type: NSRelOfficeDocument, Target: "word/document.xml"},
		},
	}
	data, err := marshalRelationships(r)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, `<?xml`) {
		t.Error("missing XML declaration")
	}
	if !strings.Contains(s, `Id="rId1"`) {
		t.Error("missing relationship Id")
	}

	// Round-trip.
	r2, err := ParseRelationships(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ParseRelationships after marshal: %v", err)
	}
	if got := r2.ByType(NSRelOfficeDocument); got == nil {
		t.Error("round-trip: missing NSRelOfficeDocument relationship")
	}
}

func TestPackageWriterAddPartFromReader(t *testing.T) {
	var buf bytes.Buffer
	pw := NewPackageWriter(&buf)
	body := strings.NewReader("<root/>")
	if err := pw.AddPart("/test.xml", "application/xml", body); err != nil {
		t.Fatal(err)
	}
	if err := pw.Close(); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	pkg, err := Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if !pkg.HasPart("/test.xml") {
		t.Error("expected /test.xml")
	}
	b, err := pkg.ReadFile("/test.xml")
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "<root/>" {
		t.Errorf("body: got %q, want %q", string(b), "<root/>")
	}
}

func TestExtensionOf(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"/word/document.xml", "xml"},
		{"/images/photo.PNG", "png"},
		{"/no-extension", ""},
		{"/dir.name/file", ""},
		{"/a.b/c.d", "d"},
	}
	for _, tc := range cases {
		got := extensionOf(tc.in)
		if got != tc.want {
			t.Errorf("extensionOf(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPartToRelsPath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "/_rels/.rels"},
		{"/", "/_rels/.rels"},
		{"/word/document.xml", "/word/_rels/document.xml.rels"},
		{"/document.xml", "/_rels/document.xml.rels"},
	}
	for _, tc := range cases {
		got := partToRelsPath(tc.in)
		if got != tc.want {
			t.Errorf("partToRelsPath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPackageWriterClosedError(t *testing.T) {
	var buf bytes.Buffer
	pw := NewPackageWriter(&buf)
	pw.AddPartBytes("/a.xml", "application/xml", []byte("<a/>"))
	pw.Close()
	if err := pw.Close(); err == nil {
		t.Error("expected error on second Close()")
	}
}

func TestPackageWriterEmptyPartNameError(t *testing.T) {
	var buf bytes.Buffer
	pw := NewPackageWriter(&buf)
	err := pw.AddPartBytes("", "application/xml", []byte("<a/>"))
	if !errors.Is(err, ErrInvalidPartName) {
		t.Errorf("empty part name: expected ErrInvalidPartName, got %v", err)
	}
}

func TestXMLEscapeAttr(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "plain"},
		{"a & b", "a &amp; b"},
		{"<tag>", "&lt;tag&gt;"},
		{`say "hi"`, `say &quot;hi&quot;`},
	}
	for _, tc := range cases {
		got := xmlEscapeAttr(tc.in)
		if got != tc.want {
			t.Errorf("xmlEscapeAttr(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// Benchmark to ensure PackageWriter doesn't do unnecessary work.
func BenchmarkPackageWriter(b *testing.B) {
	body := []byte(testDocXML)
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		pw := NewPackageWriter(&buf)
		pw.AddPartBytes("/word/document.xml", CTWordDocumentMain, body)
		pw.Close()
		_ = buf.Bytes()
	}
}

// Verify io.ReadAll from a nil part body produces empty bytes (edge case).
func TestPackageWriterEmptyBody(t *testing.T) {
	var buf bytes.Buffer
	pw := NewPackageWriter(&buf)
	if err := pw.AddPartBytes("/empty.xml", "application/xml", nil); err != nil {
		t.Fatal(err)
	}
	if err := pw.Close(); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	pkg, err := Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	b, err := pkg.ReadFile("/empty.xml")
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 0 {
		t.Errorf("expected empty body, got %d bytes", len(b))
	}
}

func TestPackageWriterRelsReaderContent(t *testing.T) {
	data := buildTestPackage(t)
	pkg, err := Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	rels, err := pkg.RootRelationships()
	if err != nil {
		t.Fatal(err)
	}
	r := rels.ByType(NSRelOfficeDocument)
	if r == nil {
		t.Fatal("missing root relationship")
	}
	got, err := ResolveTarget("/_rels/.rels", r.Target)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/word/document.xml" {
		t.Errorf("resolved target: got %q, want /word/document.xml", got)
	}
}

