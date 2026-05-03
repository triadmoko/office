package xmlwriter

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
)

func newWriter() (*Writer, *bytes.Buffer) {
	var buf bytes.Buffer
	return New(&buf), &buf
}

func TestXMLDecl(t *testing.T) {
	w, buf := newWriter()
	if err := w.StartElement(xml.Name{Local: "root"}, nil); err != nil {
		t.Fatal(err)
	}
	w.EndElement()
	w.Close()
	got := buf.String()
	if !strings.HasPrefix(got, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`) {
		t.Errorf("missing XML decl: %q", got)
	}
}

func TestSelfClosing(t *testing.T) {
	w, buf := newWriter()
	w.StartElement(xml.Name{Local: "root"}, nil)
	w.StartElement(xml.Name{Local: "empty"}, nil)
	w.EndElement() // closes "empty" — should self-close
	w.EndElement() // closes "root"
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "<empty/>") {
		t.Errorf("expected self-closing <empty/>, got: %q", got)
	}
	if strings.Contains(got, "</empty>") {
		t.Errorf("unexpected closing tag </empty> in %q", got)
	}
}

func TestWithChildren(t *testing.T) {
	w, buf := newWriter()
	w.StartElement(xml.Name{Local: "p"}, nil)
	w.CharData("hello")
	w.EndElement()
	w.Close()
	got := buf.String()
	if !strings.Contains(got, "<p>hello</p>") {
		t.Errorf("expected <p>hello</p>, got: %q", got)
	}
}

func TestNamespacePrefix(t *testing.T) {
	w, buf := newWriter()
	const ns = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	w.DeclareNamespace("w", ns)
	w.StartElement(xml.Name{Space: ns, Local: "document"}, nil)
	w.StartElement(xml.Name{Space: ns, Local: "body"}, nil)
	w.EndElement()
	w.EndElement()
	w.Close()
	got := buf.String()
	if !strings.Contains(got, "<w:document") {
		t.Errorf("expected w:document prefix, got: %q", got)
	}
	if !strings.Contains(got, "<w:body/>") {
		t.Errorf("expected w:body self-closing, got: %q", got)
	}
	if !strings.Contains(got, "</w:document>") {
		t.Errorf("expected </w:document>, got: %q", got)
	}
}

func TestNamespacePrefixAttribute(t *testing.T) {
	w, buf := newWriter()
	const wns = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	const rns = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	w.DeclareNamespace("w", wns)
	w.DeclareNamespace("r", rns)
	attrs := []xml.Attr{
		{Name: xml.Name{Space: rns, Local: "id"}, Value: "rId1"},
	}
	w.StartElement(xml.Name{Space: wns, Local: "hyperlink"}, attrs)
	w.EndElement()
	w.Close()
	got := buf.String()
	if !strings.Contains(got, `r:id="rId1"`) {
		t.Errorf("expected r:id attribute, got: %q", got)
	}
}

func TestEscapeCharData(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain", "plain"},
		{"a & b", "a &amp; b"},
		{"<tag>", "&lt;tag&gt;"},
		{"a\x00b", "ab"}, // NUL stripped
	}
	for _, tc := range cases {
		got := escapeCharData(tc.in)
		if got != tc.want {
			t.Errorf("escapeCharData(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestEscapeAttr(t *testing.T) {
	cases := []struct{ in, want string }{
		{`say "hi"`, `say &quot;hi&quot;`},
		{"a & b", "a &amp; b"},
		{"<bad>", "&lt;bad&gt;"},
		{"a\x00b", "ab"},
	}
	for _, tc := range cases {
		got := escapeAttr(tc.in)
		if got != tc.want {
			t.Errorf("escapeAttr(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRoundTrip(t *testing.T) {
	const ns = "http://example.com/ns"
	w, buf := newWriter()
	w.DeclareNamespace("ex", ns)
	w.StartElement(xml.Name{Space: ns, Local: "root"}, []xml.Attr{
		{Name: xml.Name{Local: "id"}, Value: "42"},
	})
	w.CharData("content & <data>")
	w.StartElement(xml.Name{Space: ns, Local: "child"}, nil)
	w.EndElement()
	w.EndElement()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	// Parse with standard library and verify round-trip.
	dec := xml.NewDecoder(strings.NewReader(buf.String()))
	var elems []string
	var chardata []string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch v := tok.(type) {
		case xml.StartElement:
			elems = append(elems, v.Name.Local)
		case xml.CharData:
			s := strings.TrimSpace(string(v))
			if s != "" {
				chardata = append(chardata, s)
			}
		}
	}
	if len(elems) != 2 || elems[0] != "root" || elems[1] != "child" {
		t.Errorf("elements: %v", elems)
	}
	if len(chardata) != 1 || chardata[0] != "content & <data>" {
		t.Errorf("chardata: %v", chardata)
	}
}

func TestNestedElements(t *testing.T) {
	w, buf := newWriter()
	w.StartElement(xml.Name{Local: "a"}, nil)
	w.StartElement(xml.Name{Local: "b"}, nil)
	w.StartElement(xml.Name{Local: "c"}, nil)
	w.EndElement() // c
	w.EndElement() // b
	w.EndElement() // a
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	// c is self-closing (no children), b has child c so it gets > and </b>
	if !strings.Contains(got, "<c/>") {
		t.Errorf("expected <c/>, got: %q", got)
	}
	if !strings.Contains(got, "</b>") {
		t.Errorf("expected </b>, got: %q", got)
	}
	if !strings.Contains(got, "</a>") {
		t.Errorf("expected </a>, got: %q", got)
	}
}

func TestUnclosedError(t *testing.T) {
	w, _ := newWriter()
	w.StartElement(xml.Name{Local: "root"}, nil)
	// Intentionally do NOT call EndElement
	if err := w.Close(); err == nil {
		t.Error("expected error from Close() with unclosed element")
	}
}

func TestEndElementUnderflow(t *testing.T) {
	w, _ := newWriter()
	if err := w.EndElement(); err == nil {
		t.Error("expected error from EndElement() with empty stack")
	}
}
