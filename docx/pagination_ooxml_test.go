package docx

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/triadmoko/office/internal/wml"
)

func documentXMLFromSavedDoc(t *testing.T, d *Document) string {
	t.Helper()
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	rc, err := d2.Package().OpenReader("/word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestPPrPaginationAPIRoundTrip(t *testing.T) {
	d := NewDocument()
	p := d.Body().AppendParagraph()
	p.SetPageBreakBefore(true)
	p.SetKeepNext(true)
	p.SetKeepLines(true)
	on := true
	p.SetWidowControl(&on)
	p.AppendRun("x")

	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	ps := d2.Body().Paragraphs()
	var got *Paragraph
	for i := len(ps) - 1; i >= 0; i-- {
		if len(ps[i].Runs()) == 1 && ps[i].Runs()[0].Text() == "x" {
			got = ps[i]
			break
		}
	}
	if got == nil {
		t.Fatal("paragraph not found")
	}
	if !got.PageBreakBefore() || !got.KeepNext() || !got.KeepLines() {
		t.Fatalf("flags: pbb=%v kn=%v kl=%v", got.PageBreakBefore(), got.KeepNext(), got.KeepLines())
	}
	won, wok := got.WidowControl()
	if !wok || !won {
		t.Fatalf("widow: ok=%v on=%v", wok, won)
	}

	xml := documentXMLFromSavedDoc(t, d2)
	for _, sub := range []string{"<w:pageBreakBefore/>", "<w:keepNext/>", "<w:keepLines/>", "<w:widowControl/>"} {
		if !strings.Contains(xml, sub) {
			t.Fatalf("missing %q in %s", sub, xml)
		}
	}

	off := false
	got.SetWidowControl(&off)
	xml2 := documentXMLFromSavedDoc(t, d2)
	if !strings.Contains(xml2, `<w:widowControl w:val="0"/>`) {
		t.Fatalf("want widow off: %s", xml2)
	}
}

func TestColumnBreakRoundTrip(t *testing.T) {
	d := NewDocument()
	p := d.Body().AppendParagraph()
	p.AppendRun("a")
	p.AppendColumnBreak()
	p.AppendRun("b")
	xml := documentXMLFromSavedDoc(t, d)
	if !strings.Contains(xml, `w:type="column"`) {
		t.Fatalf("column br missing: %s", xml)
	}
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if txt, err := d2.PlainText(); err != nil || txt != "a\nb" {
		t.Fatalf("plaintext %q err %v", txt, err)
	}
}

func TestLastRenderedPageBreakMarshalStrip(t *testing.T) {
	const doc = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
		`<w:body><w:p><w:r><w:lastRenderedPageBreak/><w:t xml:space="preserve">z</w:t></w:r></w:p></w:body></w:document>`
	wd, err := wml.ParseDocument(strings.NewReader(doc))
	if err != nil {
		t.Fatal(err)
	}
	out, err := MarshalDocumentXML(wd, MarshalDocumentOpts{StripLayoutHints: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "lastRenderedPageBreak") {
		t.Fatalf("strip failed: %s", out)
	}
	out2, err := MarshalDocumentXML(wd, MarshalDocumentOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out2), "lastRenderedPageBreak") {
		t.Fatal("expected hint when not stripping")
	}
}

func TestDocumentStripLayoutHintsOnSave(t *testing.T) {
	d := NewDocument()
	d.SetStripLayoutHints(true)
	d.Body().AppendParagraph().AppendRun("only")
	s := documentXMLFromSavedDoc(t, d)
	if strings.Contains(s, "lastRenderedPageBreak") {
		t.Fatalf("unexpected hint: %s", s)
	}
}

func TestTblPrExtraRoundTripViaWML(t *testing.T) {
	const doc = `<?xml version="1.0"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
		`<w:body><w:tbl>` +
		`<w:tblPr><w:tblW w:w="5000" w:type="pct"/><w:tblLayout w:type="autofit"/></w:tblPr>` +
		`<w:tblGrid><w:gridCol w:w="1000"/></w:tblGrid>` +
		`<w:tr><w:tc><w:p/></w:tc></w:tr></w:tbl></w:body></w:document>`
	wd, err := wml.ParseDocument(strings.NewReader(doc))
	if err != nil {
		t.Fatal(err)
	}
	out, err := MarshalDocumentXML(wd, MarshalDocumentOpts{})
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "tblLayout") || !strings.Contains(s, "autofit") {
		t.Fatalf("tblPr extra lost: %s", s)
	}
}

func TestTableCantSplitAndHeaderRowRoundTrip(t *testing.T) {
	d := NewDocument()
	tbl := d.Body().AppendTable(1, 1)
	row := tbl.Rows()[0]
	row.SetCantSplit(true)
	row.SetRepeatAsHeaderRow(true)
	xml := documentXMLFromSavedDoc(t, d)
	if !strings.Contains(xml, "<w:cantSplit/>") || !strings.Contains(xml, "<w:tblHeader/>") {
		t.Fatalf("trPr flags: %s", xml)
	}
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	r0 := d2.Body().Tables()[0].Rows()[0]
	if !r0.CantSplit() || !r0.RepeatAsHeaderRow() {
		t.Fatalf("readback cant=%v hdr=%v", r0.CantSplit(), r0.RepeatAsHeaderRow())
	}
}

func TestPPrRawTailRoundTrip(t *testing.T) {
	const doc = `<?xml version="1.0"?>` +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
		`<w:body><w:p><w:pPr><w:suppressLineNumbers/><w:jc w:val="left"/></w:pPr><w:r><w:t>a</w:t></w:r></w:p></w:body></w:document>`
	wd, err := wml.ParseDocument(strings.NewReader(doc))
	if err != nil {
		t.Fatal(err)
	}
	ps := wd.DirectParagraphs()
	if len(ps) != 1 || !strings.Contains(string(ps[0].PPr.RawPPrTail), "suppressLineNumbers") {
		t.Fatalf("tail: %q", ps[0].PPr.RawPPrTail)
	}
	out, err := MarshalDocumentXML(wd, MarshalDocumentOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "suppressLineNumbers") {
		t.Fatalf("marshal lost tail: %s", out)
	}
}

func TestSectionPageNumberStartAndFormatRoundTrip(t *testing.T) {
	d := NewDocument()
	if sec := d.SectionAt(0); sec != nil {
		sec.SetPageNumberFormat(PageNumberFormatLowerRoman)
		start := 3
		sec.SetPageNumberStart(&start)
	}
	xml := documentXMLFromSavedDoc(t, d)
	if !strings.Contains(xml, `<w:pgNumType`) || !strings.Contains(xml, `w:start="3"`) || !strings.Contains(xml, `w:fmt="lowerRoman"`) {
		t.Fatalf("pgNumType missing: %s", xml)
	}
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	s0 := d2.SectionAt(0)
	if s0 == nil {
		t.Fatal("nil section")
	}
	if s0.PageNumberFormat() != PageNumberFormatLowerRoman {
		t.Fatalf("fmt got %q", s0.PageNumberFormat())
	}
	n, ok := s0.PageNumberStart()
	if !ok || n != 3 {
		t.Fatalf("start got %d ok=%v", n, ok)
	}
	s0.SetPageNumberFormat(PageNumberFormatDefault)
	s0.SetPageNumberStart(nil)
	xml2 := documentXMLFromSavedDoc(t, d2)
	if strings.Contains(xml2, `w:pgNumType`) {
		t.Fatalf("expected pgNumType cleared: %s", xml2)
	}
}

func TestSectionBreakConfigPageNumber(t *testing.T) {
	d := NewDocument()
	p := d.Body().AppendParagraph()
	start := 10
	p.SetSectionBreak(SectionBreakConfig{
		PageKind:         PageSizeA4,
		Orient:           Portrait,
		Break:            SectionBreakNextPage,
		PageNumberFormat: PageNumberFormatUpperRoman,
		PageNumberStart:  &start,
	})
	p.AppendRun("isi section baru")
	xml := documentXMLFromSavedDoc(t, d)
	if !strings.Contains(xml, `w:fmt="upperRoman"`) || !strings.Contains(xml, `w:start="10"`) {
		t.Fatalf("sect break pgNum missing: %s", xml)
	}
}
