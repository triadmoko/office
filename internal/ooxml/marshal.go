package ooxml

import (
	"bytes"
	"sort"
	"strings"
)

// marshalContentTypes serializes ct to [Content_Types].xml bytes with deterministic ordering.
func marshalContentTypes(ct *ContentTypes) ([]byte, error) {
	defaults := make([]CTDefault, len(ct.Default))
	copy(defaults, ct.Default)
	sort.Slice(defaults, func(i, j int) bool {
		return defaults[i].Extension < defaults[j].Extension
	})

	overrides := make([]CTOverride, len(ct.Override))
	copy(overrides, ct.Override)
	sort.Slice(overrides, func(i, j int) bool {
		return overrides[i].PartName < overrides[j].PartName
	})

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	for _, d := range defaults {
		buf.WriteString(`<Default Extension="`)
		buf.WriteString(xmlEscapeAttr(d.Extension))
		buf.WriteString(`" ContentType="`)
		buf.WriteString(xmlEscapeAttr(d.ContentType))
		buf.WriteString(`"/>`)
	}
	for _, o := range overrides {
		buf.WriteString(`<Override PartName="`)
		buf.WriteString(xmlEscapeAttr(o.PartName))
		buf.WriteString(`" ContentType="`)
		buf.WriteString(xmlEscapeAttr(o.ContentType))
		buf.WriteString(`"/>`)
	}
	buf.WriteString(`</Types>`)
	return buf.Bytes(), nil
}

// marshalRelationships serializes r to .rels XML bytes with deterministic ordering.
func marshalRelationships(r *Relationships) ([]byte, error) {
	rels := make([]Relationship, len(r.Relationship))
	copy(rels, r.Relationship)
	sort.Slice(rels, func(i, j int) bool {
		return rels[i].ID < rels[j].ID
	})

	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	for _, rel := range rels {
		buf.WriteString(`<Relationship Id="`)
		buf.WriteString(xmlEscapeAttr(rel.ID))
		buf.WriteString(`" Type="`)
		buf.WriteString(xmlEscapeAttr(rel.Type))
		buf.WriteString(`" Target="`)
		buf.WriteString(xmlEscapeAttr(rel.Target))
		buf.WriteByte('"')
		if rel.TargetMode != "" {
			buf.WriteString(` TargetMode="`)
			buf.WriteString(xmlEscapeAttr(rel.TargetMode))
			buf.WriteByte('"')
		}
		buf.WriteString(`/>`)
	}
	buf.WriteString(`</Relationships>`)
	return buf.Bytes(), nil
}

// xmlEscapeAttr escapes s for use inside a double-quoted XML attribute value.
func xmlEscapeAttr(s string) string {
	if !strings.ContainsAny(s, `&<>"`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
