package ooxml

import (
	"encoding/xml"
	"io"
)

// Relationships is a package or part-level .rels file.
type Relationships struct {
	XMLName      xml.Name         `xml:"http://schemas.openxmlformats.org/package/2006/relationships Relationships"`
	Relationship []Relationship `xml:"Relationship"`
}

// Relationship is a single Relationship element.
type Relationship struct {
	ID     string `xml:"Id,attr"`
	Type   string `xml:"Type,attr"`
	Target string `xml:"Target,attr"`
	TargetMode string `xml:"TargetMode,attr"`
}

// ParseRelationships decodes a .rels file from r.
func ParseRelationships(r io.Reader) (*Relationships, error) {
	var rels Relationships
	dec := xml.NewDecoder(r)
	dec.Strict = false
	if err := dec.Decode(&rels); err != nil {
		return nil, err
	}
	return &rels, nil
}

// ByType returns the first relationship with the given Type URI, or nil.
func (r *Relationships) ByType(typeURI string) *Relationship {
	if r == nil {
		return nil
	}
	for i := range r.Relationship {
		if r.Relationship[i].Type == typeURI {
			return &r.Relationship[i]
		}
	}
	return nil
}
