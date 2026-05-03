package docx

import (
	"bytes"
	"strings"
	"testing"
)

func TestAppendTOCFieldMarshal(t *testing.T) {
	d := NewDocument()
	d.Body().Paragraphs()[0].AppendTOCField(TOCFieldOptions{})
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	xmlb, err := d2.Package().ReadFile("/word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	xml := string(xmlb)
	if !strings.Contains(xml, "w:instrText") || !strings.Contains(xml, "TOC") {
		t.Fatalf("missing TOC instrText: snippet=%q", xml[:snippetLen(400, len(xml))])
	}
	if !strings.Contains(xml, `fldCharType="begin"`) || !strings.Contains(xml, `fldCharType="separate"`) || !strings.Contains(xml, `fldCharType="end"`) {
		t.Fatal("missing fldChar sequence")
	}
	// Instruction lives inside w:instrText; default uses \\h \\z but not \\u (\\u breaks style-only headings).
	if !strings.Contains(xml, `\h`) || !strings.Contains(xml, `\z`) || strings.Contains(xml, `\u`) {
		t.Fatalf("expected default \\h \\z without \\u in instruction: %q", xml)
	}
}

func TestAppendTOCFieldNoHyperlinkSwitch(t *testing.T) {
	d := NewDocument()
	d.Body().Paragraphs()[0].AppendTOCField(TOCFieldOptions{
		OutlineLevels:              "1-2",
		Hyperlinks:                 false,
		OmitPageNumbersInWebLayout: true,
		UseAppliedOutlineLevels:    true,
	})
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	xmlb, err := d2.Package().ReadFile("/word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	xml := string(xmlb)
	start := strings.Index(xml, "<w:instrText")
	if start < 0 {
		t.Fatal("no instrText")
	}
	end := strings.Index(xml[start:], "</w:instrText>")
	if end < 0 {
		t.Fatal("no end instrText")
	}
	instr := xml[start : start+end+len("</w:instrText>")]
	if strings.Contains(instr, `\h`) {
		t.Fatalf("did not want \\h in instrText: %q", instr)
	}
	if !strings.Contains(instr, `\o "1-2"`) {
		t.Fatalf("expected \\o 1-2 in instrText: %q", instr)
	}
}

func TestMarshalStylesHeadingOutlineLvl(t *testing.T) {
	d := NewDocument()
	var buf bytes.Buffer
	if err := d.Save(&buf); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	stxml, err := d2.Package().ReadFile("/word/styles.xml")
	if err != nil {
		t.Fatal(err)
	}
	s := string(stxml)
	if !strings.Contains(s, `w:styleId="Heading1"`) || strings.Count(s, "<w:outlineLvl") < 3 {
		t.Fatalf("styles.xml missing outline for Heading1–3: %q", s[:snippetLen(800, len(s))])
	}
}

// snippetLen returns min(limit, total) for safe string slicing.
func snippetLen(limit, total int) int {
	if total < limit {
		return total
	}
	return limit
}
