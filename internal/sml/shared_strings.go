package sml

import (
	"encoding/xml"
	"io"
	"strings"
)

// ParseSharedStrings streams sharedStrings.xml and appends one plain string per <si>
// (rich text runs are concatenated). It does not build a DOM of the whole file.
func ParseSharedStrings(r io.Reader) ([]string, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	var out []string
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "si" {
			continue
		}
		var si siElement
		if err := dec.DecodeElement(&si, &se); err != nil {
			return nil, err
		}
		out = append(out, si.plain())
	}
	return out, nil
}

type siElement struct {
	T []tElement `xml:"t"`
	R []rElement `xml:"r"`
}

type rElement struct {
	T []tElement `xml:"t"`
}

type tElement struct {
	Space string `xml:"http://www.w3.org/XML/1998/namespace space,attr"`
	Text  string `xml:",chardata"`
}

func (si *siElement) plain() string {
	var b strings.Builder
	for _, t := range si.T {
		b.WriteString(t.preserve())
	}
	for _, r := range si.R {
		for _, t := range r.T {
			b.WriteString(t.preserve())
		}
	}
	return b.String()
}

func (t *tElement) preserve() string {
	if t == nil {
		return ""
	}
	s := t.Text
	if t.Space == "preserve" {
		return s
	}
	return strings.TrimSpace(s)
}
