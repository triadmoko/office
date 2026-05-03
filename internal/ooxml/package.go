package ooxml

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
)

// Package is an opened OOXML OPC package (ZIP + parsed metadata).
type Package struct {
	z     *zip.Reader
	ct    *ContentTypes
	files map[string]*zip.File // normalized "/part" -> zip file
}

// Open reads an OOXML package from a ZIP reader positioned at ra, size.
func Open(ra io.ReaderAt, size int64) (*Package, error) {
	z, err := zip.NewReader(ra, size)
	if err != nil {
		return nil, ErrInvalidArchive
	}
	p := &Package{z: z, files: make(map[string]*zip.File)}
	for _, f := range z.File {
		name := f.Name
		name = strings.TrimPrefix(name, "./")
		key := NormalizePartName("/" + name)
		p.files[key] = f
	}
	rc, err := p.OpenReader("[Content_Types].xml")
	if err != nil {
		return nil, ErrMissingContentTypes
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, ErrMissingContentTypes
	}
	ct, err := ParseContentTypes(bytes.NewReader(data))
	if err != nil {
		return nil, ErrMalformedContentTypes
	}
	p.ct = ct
	return p, nil
}

// ContentTypes returns parsed [Content_Types].xml.
func (p *Package) ContentTypes() *ContentTypes {
	if p == nil {
		return nil
	}
	return p.ct
}

// OpenReader returns a ReadCloser for part path (e.g. "/word/document.xml" or "word/document.xml").
func (p *Package) OpenReader(partName string) (io.ReadCloser, error) {
	if p == nil {
		return nil, ErrPartNotFound
	}
	key := NormalizePartName(partName)
	f := p.files[key]
	if f == nil {
		alt := "/" + strings.TrimPrefix(key, "/")
		if alt != key {
			f = p.files[alt]
		}
	}
	if f == nil {
		zname := ZipEntryName(partName)
		for k, zf := range p.files {
			if strings.TrimPrefix(k, "/") == zname {
				f = zf
				break
			}
		}
	}
	if f == nil {
		return nil, ErrPartNotFound
	}
	return f.Open()
}

// ReadFile reads an entire part into memory.
func (p *Package) ReadFile(partName string) ([]byte, error) {
	rc, err := p.OpenReader(partName)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// ZipReader exposes the underlying zip.Reader for advanced consumers (e.g. writers).
func (p *Package) ZipReader() *zip.Reader {
	if p == nil {
		return nil
	}
	return p.z
}

// FileNames returns all part paths with a leading slash, sorted for stability.
func (p *Package) FileNames() []string {
	if p == nil || p.z == nil {
		return nil
	}
	out := make([]string, 0, len(p.z.File))
	for _, f := range p.z.File {
		if f == nil {
			continue
		}
		name := strings.TrimPrefix(f.Name, "./")
		out = append(out, NormalizePartName("/"+name))
	}
	return out
}

// HasPart reports whether partName exists in the package.
func (p *Package) HasPart(partName string) bool {
	if p == nil {
		return false
	}
	key := NormalizePartName(partName)
	if p.files[key] != nil {
		return true
	}
	zname := ZipEntryName(partName)
	for k := range p.files {
		if strings.TrimPrefix(k, "/") == zname {
			return true
		}
	}
	return false
}

// RootRelationships parses _rels/.rels if present.
func (p *Package) RootRelationships() (*Relationships, error) {
	if p == nil {
		return nil, ErrMissingRelationships
	}
	rc, err := p.OpenReader("_rels/.rels")
	if err != nil {
		return nil, ErrMissingRelationships
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, ErrMissingRelationships
	}
	rels, err := ParseRelationships(bytes.NewReader(data))
	if err != nil {
		return nil, ErrMalformedRelationships
	}
	return rels, nil
}

// RelationshipBaseDir returns the OPC base directory for a .rels part
// (e.g. /_rels/.rels -> /, /word/_rels/document.xml.rels -> /word/).
func RelationshipBaseDir(relsPartPath string) string {
	p := NormalizePartName(relsPartPath)
	const seg = "/_rels/"
	i := strings.Index(p, seg)
	if i < 0 {
		return "/"
	}
	if i == 0 {
		return "/"
	}
	return NormalizePartName(p[:i] + "/")
}

// ResolveTarget resolves a relationship Target relative to the .rels part path relsPart.
func ResolveTarget(relsPart, target string) string {
	target = strings.TrimSpace(target)
	if strings.HasPrefix(target, "/") {
		return NormalizePartName(target)
	}
	base := RelationshipBaseDir(relsPart)
	return joinResolveOPC(base, target)
}

func joinResolveOPC(baseDir, rel string) string {
	base := NormalizePartName(baseDir)
	if base != "/" {
		base = strings.TrimSuffix(base, "/")
	}
	var segs []string
	if base != "" && base != "/" {
		s := strings.TrimPrefix(base, "/")
		if s != "" {
			segs = strings.Split(s, "/")
		}
	}
	for _, t := range strings.Split(rel, "/") {
		switch t {
		case "", ".":
			continue
		case "..":
			if len(segs) > 0 {
				segs = segs[:len(segs)-1]
			}
		default:
			segs = append(segs, t)
		}
	}
	if len(segs) == 0 {
		return "/"
	}
	return NormalizePartName("/" + strings.Join(segs, "/"))
}

// WalkParts calls fn for each zip file (skipping directory entries). If fn returns
// fs.SkipDir or an error, iteration stops.
func (p *Package) WalkParts(fn func(partName string, zf *zip.File) error) error {
	if p == nil || p.z == nil {
		return nil
	}
	for _, f := range p.z.File {
		if f.FileInfo().IsDir() {
			continue
		}
		name := strings.TrimPrefix(f.Name, "./")
		part := NormalizePartName("/" + name)
		if err := fn(part, f); err != nil {
			return err
		}
	}
	return nil
}
