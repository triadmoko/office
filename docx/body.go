package docx

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/triadmoko/office/internal/ooxml"
)

// PlainText extracts concatenated text from w:t runs in the main document part.
func (d *Document) PlainText() (string, error) {
	if d == nil || d.pkg == nil {
		return "", ErrMissingMainPart
	}
	rc, err := d.pkg.OpenReader(d.main)
	if err != nil {
		return "", err
	}
	defer rc.Close()
	return plainTextFromWordML(rc)
}

func plainTextFromWordML(r io.Reader) (string, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	var b strings.Builder
	inT := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", ErrMalformedBody
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWT(t.Name) {
				inT = true
			}
		case xml.EndElement:
			if isWT(t.Name) {
				inT = false
			}
		case xml.CharData:
			if inT {
				b.Write([]byte(t))
			}
		}
	}
	return b.String(), nil
}

func isWT(n xml.Name) bool {
	if n.Local != "t" {
		return false
	}
	return n.Space == ooxml.NSWordprocessingMain || n.Space == ""
}
