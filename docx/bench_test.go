package docx

import (
	"bytes"
	"testing"

	"github.com/triadmoko/office/internal/wml"
)

func BenchmarkParse100Paragraphs(b *testing.B) {
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for range 100 {
		buf.WriteString(`<w:p><w:r><w:t>hello</w:t></w:r></w:p>`)
	}
	buf.WriteString(`</w:body></w:document>`)
	data := buf.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = wml.ParseDocument(bytes.NewReader(data))
	}
}

func BenchmarkSave1000Paragraphs(b *testing.B) {
	d := NewDocument()
	body := d.Body()
	for range 1000 {
		body.AppendParagraph().AppendRun("x")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var w bytes.Buffer
		_ = d.Save(&w)
	}
}
