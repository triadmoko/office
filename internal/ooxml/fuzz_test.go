package ooxml

import (
	"strings"
	"testing"
)

// FuzzParseContentTypes ensures ParseContentTypes never panics and that any
// successfully parsed PartName values are free of ".." traversal segments.
func FuzzParseContentTypes(f *testing.F) {
	// Valid seeds
	f.Add(`<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="xml" ContentType="application/xml"/></Types>`)
	f.Add(`<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/></Types>`)
	f.Add(`<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"></Types>`)
	f.Add(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/></Types>`)
	f.Add(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="png" ContentType="image/png"/><Default Extension="jpeg" ContentType="image/jpeg"/></Types>`)
	// Malformed seeds
	f.Add(`not xml at all`)
	f.Add(``)
	f.Add(`<Types`)
	f.Add(`<?xml version="1.0"?><Types xmlns="x"><Override PartName="/word/../doc.xml" ContentType="text/plain"/></Types>`)
	f.Add(`<?xml version="1.0"?><Types xmlns="x"><Override PartName="\x00" ContentType="text/plain"/></Types>`)

	f.Fuzz(func(t *testing.T, data string) {
		ct, err := ParseContentTypes(strings.NewReader(data))
		if err != nil {
			return // error is acceptable
		}
		// Invariant: no successfully parsed PartName should contain a ".." segment.
		for _, o := range ct.Override {
			for _, seg := range strings.Split(o.PartName, "/") {
				if seg == ".." {
					t.Errorf("traversal escaped ParseContentTypes: PartName=%q", o.PartName)
				}
			}
		}
	})
}

// FuzzParseRelationships ensures ParseRelationships never panics on arbitrary input.
func FuzzParseRelationships(f *testing.F) {
	// Valid seeds
	f.Add(`<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`)
	f.Add(`<?xml version="1.0"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`)
	f.Add(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://example.com/t" Target="/abs/path" TargetMode="External"/></Relationships>`)
	f.Add(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId2" Type="http://example.com/t" Target="../relative.xml"/></Relationships>`)
	f.Add(`<?xml version="1.0"?><Relationships xmlns="x"><Relationship Id="" Type="" Target=""/></Relationships>`)
	// Malformed seeds
	f.Add(``)
	f.Add(`not xml`)
	f.Add(`<Relationships`)
	f.Add(`<?xml?><Relationships xmlns="x"><Relationship Id="rId1"/></Relationships>`)
	f.Add("\x00\x01\x02\x03")

	f.Fuzz(func(t *testing.T, data string) {
		// Must not panic.
		_, _ = ParseRelationships(strings.NewReader(data))
	})
}

// FuzzResolveTarget ensures ResolveTarget is panic-free and that any non-error
// result satisfies basic invariants: starts with "/" and has no ".." segments.
func FuzzResolveTarget(f *testing.F) {
	// Seed corpus: (relsPart, target) pairs
	f.Add("/_rels/.rels", "word/document.xml")
	f.Add("/_rels/.rels", "../../etc/passwd")
	f.Add("/word/_rels/document.xml.rels", "../document.xml")
	f.Add("/word/_rels/document.xml.rels", "document.xml")
	f.Add("/_rels/.rels", "/ppt/presentation.xml")
	f.Add("/_rels/.rels", "")
	f.Add("/_rels/.rels", "./word/document.xml")
	f.Add("/a/b/c/_rels/d.rels", "../../../../escape")
	f.Add("", "")
	f.Add("not-a-rels-path", "target.xml")

	f.Fuzz(func(t *testing.T, relsPart, target string) {
		result, err := ResolveTarget(relsPart, target)
		if err != nil {
			return // error is acceptable
		}
		// Invariant 1: result must start with "/".
		if !strings.HasPrefix(result, "/") {
			t.Errorf("ResolveTarget(%q, %q) = %q: does not start with /", relsPart, target, result)
		}
		// Invariant 2: result must not contain ".." segments.
		for _, seg := range strings.Split(result, "/") {
			if seg == ".." {
				t.Errorf("ResolveTarget(%q, %q) = %q: contains traversal segment", relsPart, target, result)
			}
		}
	})
}
