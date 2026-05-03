// Package opcprops provides parsers and serializers for OPC document properties parts:
// docProps/core.xml ([CoreProperties]) and docProps/app.xml ([AppProperties]).
package opcprops

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"time"
)

// Version is the default AppVersion string reported in generated app.xml files.
const Version = "0.1"

// Application is the default application name reported in generated app.xml files.
const Application = "github.com/triadmoko/office"

// CoreProperties maps to the OPC Core Properties part (docProps/core.xml).
// Namespace prefixes used in the serialized XML: cp, dc, dcterms, xsi.
type CoreProperties struct {
	Title          string
	Subject        string
	Creator        string
	Keywords       string
	Description    string
	LastModifiedBy string
	Revision       string
	Category       string
	ContentStatus  string
	Language       string
	Version        string
	Created        time.Time
	Modified       time.Time
}

// AppProperties maps to the Office Extended Properties part (docProps/app.xml).
type AppProperties struct {
	Application       string // defaults to [Application] when empty
	AppVersion        string // defaults to [Version] when empty
	Company           string
	Manager           string
	DocSecurity       int
	ScaleCrop         bool
	LinksUpToDate     bool
	SharedDoc         bool
	HyperlinksChanged bool
}

// ParseCore decodes docProps/core.xml from r.
// Handles mixed namespaces: cp:, dc:, dcterms:.
func ParseCore(r io.Reader) (*CoreProperties, error) {
	dec := xml.NewDecoder(r)
	var c CoreProperties
	var current string
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			current = t.Name.Local
		case xml.EndElement:
			current = ""
		case xml.CharData:
			s := string(bytes.TrimSpace(t))
			if s == "" {
				continue
			}
			switch current {
			case "title":
				c.Title = s
			case "subject":
				c.Subject = s
			case "creator":
				c.Creator = s
			case "keywords":
				c.Keywords = s
			case "description":
				c.Description = s
			case "lastModifiedBy":
				c.LastModifiedBy = s
			case "revision":
				c.Revision = s
			case "category":
				c.Category = s
			case "contentStatus":
				c.ContentStatus = s
			case "language":
				c.Language = s
			case "version":
				c.Version = s
			case "created":
				c.Created, _ = time.Parse(time.RFC3339, s)
			case "modified":
				c.Modified, _ = time.Parse(time.RFC3339, s)
			}
		}
	}
	return &c, nil
}

// ParseApp decodes docProps/app.xml from r.
func ParseApp(r io.Reader) (*AppProperties, error) {
	type xmlApp struct {
		Application       string `xml:"Application"`
		AppVersion        string `xml:"AppVersion"`
		Company           string `xml:"Company"`
		Manager           string `xml:"Manager"`
		DocSecurity       int    `xml:"DocSecurity"`
		ScaleCrop         bool   `xml:"ScaleCrop"`
		LinksUpToDate     bool   `xml:"LinksUpToDate"`
		SharedDoc         bool   `xml:"SharedDoc"`
		HyperlinksChanged bool   `xml:"HyperlinksChanged"`
	}
	var x xmlApp
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&x); err != nil {
		return nil, err
	}
	return &AppProperties{
		Application:       x.Application,
		AppVersion:        x.AppVersion,
		Company:           x.Company,
		Manager:           x.Manager,
		DocSecurity:       x.DocSecurity,
		ScaleCrop:         x.ScaleCrop,
		LinksUpToDate:     x.LinksUpToDate,
		SharedDoc:         x.SharedDoc,
		HyperlinksChanged: x.HyperlinksChanged,
	}, nil
}

// WriteTo serializes c to core.xml XML and writes it to w.
func (c *CoreProperties) WriteTo(w io.Writer) (int64, error) {
	if c == nil {
		c = &CoreProperties{}
	}
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<cp:coreProperties`)
	buf.WriteString(` xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"`)
	buf.WriteString(` xmlns:dc="http://purl.org/dc/elements/1.1/"`)
	buf.WriteString(` xmlns:dcterms="http://purl.org/dc/terms/"`)
	buf.WriteString(` xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`)
	buf.WriteByte('>')

	writeElem(&buf, "dc:title", c.Title)
	writeElem(&buf, "dc:subject", c.Subject)
	writeElem(&buf, "dc:creator", c.Creator)
	writeElem(&buf, "cp:keywords", c.Keywords)
	writeElem(&buf, "dc:description", c.Description)
	writeElem(&buf, "cp:lastModifiedBy", c.LastModifiedBy)
	writeElem(&buf, "cp:revision", c.Revision)
	writeElem(&buf, "cp:category", c.Category)
	writeElem(&buf, "cp:contentStatus", c.ContentStatus)
	writeElem(&buf, "dc:language", c.Language)
	writeElem(&buf, "cp:version", c.Version)
	if !c.Created.IsZero() {
		fmt.Fprintf(&buf, `<dcterms:created xsi:type="dcterms:W3CDTF">%s</dcterms:created>`,
			c.Created.UTC().Format(time.RFC3339))
	}
	if !c.Modified.IsZero() {
		fmt.Fprintf(&buf, `<dcterms:modified xsi:type="dcterms:W3CDTF">%s</dcterms:modified>`,
			c.Modified.UTC().Format(time.RFC3339))
	}
	buf.WriteString(`</cp:coreProperties>`)

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// WriteTo serializes a to app.xml XML and writes it to w.
func (a *AppProperties) WriteTo(w io.Writer) (int64, error) {
	if a == nil {
		a = &AppProperties{}
	}
	app := a.Application
	if app == "" {
		app = Application
	}
	ver := a.AppVersion
	if ver == "" {
		ver = Version
	}

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"`)
	buf.WriteString(` xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"`)
	buf.WriteByte('>')

	writeElem(&buf, "Application", app)
	writeElem(&buf, "AppVersion", ver)
	writeElem(&buf, "Company", a.Company)
	writeElem(&buf, "Manager", a.Manager)
	if a.DocSecurity != 0 {
		fmt.Fprintf(&buf, "<DocSecurity>%d</DocSecurity>", a.DocSecurity)
	}
	writeBoolElem(&buf, "ScaleCrop", a.ScaleCrop)
	writeBoolElem(&buf, "LinksUpToDate", a.LinksUpToDate)
	writeBoolElem(&buf, "SharedDoc", a.SharedDoc)
	writeBoolElem(&buf, "HyperlinksChanged", a.HyperlinksChanged)

	buf.WriteString(`</Properties>`)

	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

func writeElem(buf *bytes.Buffer, tag, val string) {
	if val == "" {
		return
	}
	buf.WriteByte('<')
	buf.WriteString(tag)
	buf.WriteByte('>')
	xml.EscapeText(buf, []byte(val))
	buf.WriteString("</")
	buf.WriteString(tag)
	buf.WriteByte('>')
}

func writeBoolElem(buf *bytes.Buffer, tag string, val bool) {
	s := "false"
	if val {
		s = "true"
	}
	fmt.Fprintf(buf, "<%s>%s</%s>", tag, s, tag)
}
