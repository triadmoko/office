package ooxml

import (
	"encoding/xml"
	"io"
)

// ContentTypes reflects [Content_Types].xml (package root).
type ContentTypes struct {
	XMLName  xml.Name         `xml:"http://schemas.openxmlformats.org/package/2006/content-types Types"`
	Default  []CTDefault      `xml:"Default"`
	Override []CTOverride     `xml:"Override"`
}

// CTDefault is a Default entry (extension to content type mapping).
type CTDefault struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// CTOverride is an Override entry (part path to content type).
type CTOverride struct {
	PartName    string `xml:"PartName,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// ParseContentTypes reads and decodes [Content_Types].xml from r.
func ParseContentTypes(r io.Reader) (*ContentTypes, error) {
	var ct ContentTypes
	dec := xml.NewDecoder(r)
	dec.Strict = false
	if err := dec.Decode(&ct); err != nil {
		return nil, err
	}
	for i := range ct.Override {
		norm, err := NormalizePartName(ct.Override[i].PartName)
		if err != nil {
			return nil, err
		}
		ct.Override[i].PartName = norm
	}
	return &ct, nil
}

// HasContentType returns true if any Override uses the given content type.
func (c *ContentTypes) HasContentType(contentType string) bool {
	if c == nil {
		return false
	}
	for _, o := range c.Override {
		if o.ContentType == contentType {
			return true
		}
	}
	return false
}

// PartNameForContentType returns the first Override PartName for contentType, or "".
func (c *ContentTypes) PartNameForContentType(contentType string) string {
	if c == nil {
		return ""
	}
	for _, o := range c.Override {
		if o.ContentType == contentType {
			return o.PartName
		}
	}
	return ""
}
