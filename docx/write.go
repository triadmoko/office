package docx

import "io"

// WriteMinimal writes a new .docx containing a single paragraph with plain text.
func WriteMinimal(w io.Writer, text string) error {
	d := NewDocument()
	d.Body().AppendParagraph().AppendRun(text)
	return d.Save(w)
}

// Write re-serializes the document via Save (full OPC package).
func (d *Document) Write(w io.Writer) error {
	return d.Save(w)
}
