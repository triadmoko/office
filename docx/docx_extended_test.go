package docx

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/triadmoko/office/internal/wml"
)

func TestSectionBreakTypeContinuousRoundTrip(t *testing.T) {
	d := NewDocument()
	p := d.Body().AppendParagraph()
	p.SetSectionBreak(SectionBreakConfig{
		PageKind: PageSizeA4,
		Orient:   Portrait,
		Break:    SectionBreakContinuous,
	})
	d.Body().AppendParagraph().AppendRun("after")
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if k := d2.SectionAt(0).BreakKind(); k != SectionBreakContinuous {
		t.Fatalf("break kind sec0: %v", k)
	}
}

func TestTwoSectionsDifferentOrientation(t *testing.T) {
	d := NewDocument()
	d.SectionAt(0).SetPageSize(PageSizeA4)
	d.SectionAt(0).SetOrientation(Portrait)
	p0 := d.Body().AppendParagraph()
	p0.AppendRun("bagian awal")
	p0.SetSectionBreak(SectionBreakConfig{PageKind: PageSizeA4, Orient: Landscape})
	d.Body().AppendParagraph().AppendRun("setelah break")
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	secs := d2.Sections()
	if len(secs) < 2 {
		t.Fatalf("want >=2 sections, got %d", len(secs))
	}
	var sawP, sawL bool
	for _, s := range secs {
		switch s.PageSize().Orient {
		case Portrait:
			sawP = true
		case Landscape:
			sawL = true
		}
	}
	if !sawP || !sawL {
		t.Fatalf("orientations: %+v", secs)
	}
}

func TestParagraphPageBreakRoundTrip(t *testing.T) {
	d := NewDocument()
	p := d.Body().AppendParagraph()
	p.AppendRun("satu")
	p.AppendPageBreak()
	p.AppendRun("dua")
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	ps := d2.Body().Paragraphs()
	var found bool
	for _, pp := range ps {
		rs := pp.Runs()
		if len(rs) >= 3 && rs[0].Text() == "satu" && rs[1].ContainsPageBreak() && rs[2].Text() == "dua" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("page break paragraph not found")
	}
}

func TestParagraphIndentSpacingRoundTrip(t *testing.T) {
	d := NewDocument()
	p := d.Body().AppendParagraph()
	p.SetIndent(Indent{Left: 500, FirstLine: 300})
	p.SetSpacing(Spacing{After: 120})
	p.SetAlignment(AlignCenter)
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
	var target *Paragraph
	for i := len(ps) - 1; i >= 0; i-- {
		if len(ps[i].Runs()) > 0 && ps[i].Runs()[0].Text() == "x" {
			target = ps[i]
			break
		}
	}
	if target == nil {
		t.Fatal("paragraph not found")
	}
	in := target.Indent()
	if in.Left != 500 || in.FirstLine != 300 {
		t.Fatalf("indent %+v", in)
	}
	if sp := target.Spacing(); sp.After != 120 {
		t.Fatalf("spacing %+v", sp)
	}
	if target.Alignment() != AlignCenter {
		t.Fatalf("alignment %v", target.Alignment())
	}
}

func TestRunHighlightRoundTrip(t *testing.T) {
	d := NewDocument()
	r := d.Body().AppendParagraph().AppendRun("sorot")
	r.SetHighlight("yellow")
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	rs := d2.Body().Paragraphs()[1].Runs()
	if len(rs) == 0 || rs[0].Highlight() != "yellow" {
		t.Fatalf("highlight readback: %q", rs[0].Highlight())
	}
}

func TestRunSubSuperscriptRoundTrip(t *testing.T) {
	d := NewDocument()
	p := d.Body().AppendParagraph()
	p.AppendRun("x")
	rs := p.AppendRun("2")
	rs.SetSubSuperscript(VertAlignSuperscript)
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	ps := d2.Body().Paragraphs()
	var last *Paragraph
	for i := len(ps) - 1; i >= 0; i-- {
		if len(ps[i].Runs()) > 0 {
			last = ps[i]
			break
		}
	}
	if last == nil {
		t.Fatal("no paragraph with runs")
	}
	rs2 := last.Runs()
	if len(rs2) < 2 || rs2[1].SubSuperscript() != VertAlignSuperscript {
		t.Fatalf("vertAlign readback: %#v", rs2)
	}
}

func TestRunEmphasisRoundTrip(t *testing.T) {
	d := NewDocument()
	r := d.Body().AppendParagraph().AppendRun("a")
	r.SetEmphasis("dot")
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	rs := d2.Body().Paragraphs()[1].Runs()
	if len(rs) == 0 || rs[0].Emphasis() != "dot" {
		t.Fatalf("emphasis readback: %q", rs[0].Emphasis())
	}
}

func TestRunStrikeRoundTrip(t *testing.T) {
	d := NewDocument()
	r := d.Body().AppendParagraph().AppendRun("coret")
	r.SetStrike(true)
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	rs := d2.Body().Paragraphs()[1].Runs()
	if len(rs) == 0 || !rs[0].Strike() {
		t.Fatalf("strike readback: %+v", rs)
	}
}

func TestNewDocumentSaveAndReadParagraphs(t *testing.T) {
	d := NewDocument()
	p := d.Body().AppendParagraph()
	r := p.AppendRun("hello")
	if r.Text() != "hello" {
		t.Fatalf("run text: %q", r.Text())
	}
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	ps := d2.Body().Paragraphs()
	if len(ps) != 2 {
		t.Fatalf("paragraphs: got %d want 2 (empty + hello)", len(ps))
	}
	if got := ps[1].Runs()[0].Text(); got != "hello" {
		t.Fatalf("readback: %q", got)
	}
}

func TestListBuilder(t *testing.T) {
	d := NewDocument()
	list := d.Body().AppendList(ListBullet)
	list.AppendItem("a")
	list.AppendItem("b")
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if txt, err := d2.PlainText(); err != nil || txt != "ab" {
		t.Fatalf("plaintext: %q err %v", txt, err)
	}
}

func TestAppendTableDefaultWidthFullTextArea(t *testing.T) {
	d := NewDocument()
	tbl := d.Body().AppendTable(1, 1)
	if w := tbl.Width(); w.Kind != WidthPct || w.Value != 5000 {
		t.Fatalf("default table width: got %+v want {5000 WidthPct}", w)
	}
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if w := d2.Body().Tables()[0].Width(); w.Kind != WidthPct || w.Value != 5000 {
		t.Fatalf("round-trip default width: got %+v", w)
	}
}

func TestTableSetWidth(t *testing.T) {
	d := NewDocument()
	tbl := d.Body().AppendTable(1, 1)
	tbl.SetWidth(TableWidth{Value: 5000, Kind: WidthPct})
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	w := d2.Body().Tables()[0].Width()
	if w.Kind != WidthPct || w.Value != 5000 {
		t.Fatalf("width readback: got %+v", w)
	}
}

func TestTableGridColWidthsAndRowHeight(t *testing.T) {
	d := NewDocument()
	tbl := d.Body().AppendTable(1, 2)
	tbl.SetGridColWidths([]int64{3000, 5000})
	tbl.Rows()[0].SetHeight(720, RowHeightExact)
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	t2 := d2.Body().Tables()[0]
	gw := t2.GridColWidths()
	if len(gw) != 2 || gw[0] != 3000 || gw[1] != 5000 {
		t.Fatalf("grid readback: %v", gw)
	}
	h, rule := t2.Rows()[0].Height()
	if h != 720 || rule != RowHeightExact {
		t.Fatalf("row height readback: %d %v", h, rule)
	}
}

func TestTableSetBorder(t *testing.T) {
	d := NewDocument()
	tbl := d.Body().AppendTable(2, 2)
	tbl.SetBorder(BorderAll, BorderStyle{Color: "000000", Size: 4, Kind: BorderSingle})
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	c := d2.Body().Tables()[0].Rows()[0].Cells()[0]
	b := c.Borders()
	if b == nil || b.b.Top == nil {
		t.Fatal("expected top border")
	}
}

func TestSavePreservesPartCount(t *testing.T) {
	path := filepath.Join("testdata", "minimal.docx")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skip("testdata:", err)
	}
	d, err := Open(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	namesBefore, err := d.Package().FileNames()
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := d.Save(&out); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(out.Bytes()), int64(out.Len()))
	if err != nil {
		t.Fatal(err)
	}
	namesAfter, err := d2.Package().FileNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(namesAfter) < len(namesBefore) {
		t.Fatalf("lost parts: before %d after %d", len(namesBefore), len(namesAfter))
	}
}

func TestModifyFirstParagraphRoundTrip(t *testing.T) {
	d := NewDocument()
	d.Body().AppendParagraph().AppendRun("alpha")
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	ps := d2.Body().Paragraphs()
	if len(ps) < 2 {
		t.Fatal("expected paragraphs")
	}
	// Replace text in first non-empty body paragraph (skip leading empty).
	target := ps[len(ps)-1]
	rs := target.Runs()
	if len(rs) == 0 {
		t.Fatal("no runs")
	}
	rs[0].x.Parts = []wml.RunPart{{Text: "beta"}}
	rs[0].x.RebuildText()
	var out bytes.Buffer
	if err := d2.Save(&out); err != nil {
		t.Fatal(err)
	}
	d3, err := Open(bytes.NewReader(out.Bytes()), int64(out.Len()))
	if err != nil {
		t.Fatal(err)
	}
	txt, err := d3.PlainText()
	if err != nil {
		t.Fatal(err)
	}
	if txt != "beta" {
		t.Fatalf("got %q", txt)
	}
}
