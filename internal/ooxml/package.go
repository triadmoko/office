package ooxml

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
)

// OpenOptions configures zip-bomb protection thresholds for [OpenWithOptions].
// A zero value for any field means "no limit" — use [Open] to get safe defaults.
type OpenOptions struct {
	MaxBytes     int64 // max total uncompressed bytes across all parts; 0 = unlimited
	MaxParts     int   // max number of ZIP entries; 0 = unlimited
	MaxPartBytes int64 // max uncompressed bytes per individual part; 0 = unlimited
}

// defaultOpenOptions are the limits applied by [Open].
var defaultOpenOptions = OpenOptions{
	MaxBytes:     1 << 30,        // 1 GiB
	MaxParts:     10_000,
	MaxPartBytes: 256 << 20,      // 256 MiB
}

// Package is an opened OOXML OPC package (ZIP + parsed metadata).
type Package struct {
	z     *zip.Reader
	ct    *ContentTypes
	files map[string]*zip.File // normalized "/part" -> zip file
	opts  OpenOptions          // size limits for subsequent reads
}

// Open reads an OOXML package with safe default size limits (1 GiB total, 256 MiB per part,
// 10 000 parts). For custom limits use [OpenWithOptions].
func Open(ra io.ReaderAt, size int64) (*Package, error) {
	return OpenWithOptions(ra, size, defaultOpenOptions)
}

// OpenWithOptions reads an OOXML package applying the given zip-bomb protection limits.
// Zero values in opts mean "no limit" for that dimension.
func OpenWithOptions(ra io.ReaderAt, size int64, opts OpenOptions) (*Package, error) {
	z, err := zip.NewReader(ra, size)
	if err != nil {
		return nil, ErrInvalidArchive
	}
	if opts.MaxParts > 0 && len(z.File) > opts.MaxParts {
		return nil, ErrTooManyParts
	}
	p := &Package{z: z, files: make(map[string]*zip.File), opts: opts}
	var totalUncompressed int64
	for _, f := range z.File {
		name := strings.TrimPrefix(f.Name, "./")
		key, err := NormalizePartName("/" + name)
		if err != nil {
			return nil, err
		}
		p.files[key] = f
		// Pre-check using ZIP header (not authenticated — enforced again at read time).
		if opts.MaxPartBytes > 0 && int64(f.UncompressedSize64) > opts.MaxPartBytes {
			return nil, ErrPackageTooLarge
		}
		totalUncompressed += int64(f.UncompressedSize64)
		if opts.MaxBytes > 0 && totalUncompressed > opts.MaxBytes {
			return nil, ErrPackageTooLarge
		}
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
	key, err := NormalizePartName(partName)
	if err != nil {
		return nil, err
	}
	f := p.files[key]
	if f == nil {
		alt := "/" + strings.TrimPrefix(key, "/")
		if alt != key {
			f = p.files[alt]
		}
	}
	if f == nil {
		zname, err := ZipEntryName(partName)
		if err != nil {
			return nil, err
		}
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
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	if p.opts.MaxPartBytes > 0 {
		return &limitedReadCloser{rc: rc, limit: p.opts.MaxPartBytes}, nil
	}
	return rc, nil
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
func (p *Package) FileNames() ([]string, error) {
	if p == nil || p.z == nil {
		return nil, nil
	}
	out := make([]string, 0, len(p.z.File))
	for _, f := range p.z.File {
		if f == nil {
			continue
		}
		name := strings.TrimPrefix(f.Name, "./")
		part, err := NormalizePartName("/" + name)
		if err != nil {
			return nil, err
		}
		out = append(out, part)
	}
	return out, nil
}

// HasPart reports whether partName exists in the package.
// Invalid partName values (see [NormalizePartName]) are treated as not present.
func (p *Package) HasPart(partName string) bool {
	if p == nil {
		return false
	}
	key, err := NormalizePartName(partName)
	if err != nil {
		return false
	}
	if p.files[key] != nil {
		return true
	}
	zname, err := ZipEntryName(partName)
	if err != nil {
		return false
	}
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
// (e.g. /_rels/.rels -> /, /word/_rels/document.xml.rels -> /word).
func RelationshipBaseDir(relsPartPath string) (string, error) {
	p, err := NormalizePartName(relsPartPath)
	if err != nil {
		return "", err
	}
	const seg = "/_rels/"
	i := strings.Index(p, seg)
	if i < 0 {
		return "/", nil
	}
	if i == 0 {
		return "/", nil
	}
	return NormalizePartName(p[:i] + "/")
}

// ResolveTarget resolves a relationship Target relative to the .rels part path relsPart.
func ResolveTarget(relsPart, target string) (string, error) {
	target = strings.TrimSpace(target)
	if strings.HasPrefix(target, "/") {
		return NormalizePartName(target)
	}
	base, err := RelationshipBaseDir(relsPart)
	if err != nil {
		return "", err
	}
	return joinResolveOPC(base, target)
}

func joinResolveOPC(baseDir, rel string) (string, error) {
	base, err := NormalizePartName(baseDir)
	if err != nil {
		return "", err
	}
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
			if len(segs) == 0 {
				return "", ErrPathTraversal
			}
			segs = segs[:len(segs)-1]
		default:
			segs = append(segs, t)
		}
	}
	if len(segs) == 0 {
		return "/", nil
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
		part, err := NormalizePartName("/" + name)
		if err != nil {
			return err
		}
		if err := fn(part, f); err != nil {
			return err
		}
	}
	return nil
}

// limitedReadCloser enforces a per-part byte limit enforced during Read calls.
type limitedReadCloser struct {
	rc    io.ReadCloser
	read  int64
	limit int64
}

func (l *limitedReadCloser) Read(p []byte) (int, error) {
	n, err := l.rc.Read(p)
	l.read += int64(n)
	if l.read > l.limit {
		return n, ErrPackageTooLarge
	}
	return n, err
}

func (l *limitedReadCloser) Close() error { return l.rc.Close() }
