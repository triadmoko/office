package docx

import (
	"archive/zip"
	"fmt"
	"io"
	"strings"
)

// WriteMinimal writes a new .docx containing a single paragraph with plain text.
func WriteMinimal(w io.Writer, text string) error {
	zw := zip.NewWriter(w)

	ct := `[Content_Types].xml`
	if err := writeZipFile(zw, ct, contentTypesXML()); err != nil {
		return err
	}
	rels := `_rels/.rels`
	if err := writeZipFile(zw, rels, rootRelsXML()); err != nil {
		return err
	}
	docPath := `word/document.xml`
	if err := writeZipFile(zw, docPath, documentXML(text)); err != nil {
		return err
	}
	return zw.Close()
}

func writeZipFile(zw *zip.Writer, name, body string) error {
	f, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, strings.NewReader(body))
	return err
}

func contentTypesXML() string {
	return xmlHeader + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
		`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
		`<Default Extension="xml" ContentType="application/xml"/>` +
		`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>` +
		`</Types>`
}

func rootRelsXML() string {
	return xmlHeader + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>` +
		`</Relationships>`
}

func documentXML(text string) string {
	escaped := escapeXML(text)
	return xmlHeader + `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` +
		`<w:body><w:p><w:r><w:t xml:space="preserve">` + escaped + `</w:t></w:r></w:p></w:body></w:document>`
}

const xmlHeader = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// Write re-serializes the document. Currently only round-trips PlainText via WriteMinimal.
func (d *Document) Write(w io.Writer) error {
	if d == nil {
		return ErrMissingMainPart
	}
	txt, err := d.PlainText()
	if err != nil {
		return fmt.Errorf("docx write: %w", err)
	}
	return WriteMinimal(w, txt)
}
