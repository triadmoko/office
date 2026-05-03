package docx

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/triadmoko/office/internal/wml"
)

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
