package ooxml

import (
	"archive/zip"
	"bytes"
	"io"
	"sort"
	"strings"
	"time"
)

// zipEpoch is the earliest representable MS-DOS time in ZIP archives (1980-01-01).
// Using this gives reproducible output without relying on the current clock.
var zipEpoch = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

// PackageWriter builds an OPC ZIP package with deterministic output.
// [Content_Types].xml is always the first ZIP entry. All other entries
// are written in lexicographic order. ZIP modification times are set to
// the zero time for reproducible builds.
type PackageWriter struct {
	zw     *zip.Writer
	parts  map[string][]byte         // normalized part name → buffered body
	ct     *ContentTypes             // accumulates Override entries via AddPart
	rels   map[string]*Relationships // normalized rels part name → Relationships
	closed bool
}

// NewPackageWriter returns a PackageWriter that writes to w.
func NewPackageWriter(w io.Writer) *PackageWriter {
	return &PackageWriter{
		zw:    zip.NewWriter(w),
		parts: make(map[string][]byte),
		ct:    &ContentTypes{},
		rels:  make(map[string]*Relationships),
	}
}

// AddPart validates name, buffers body, and registers contentType as an Override entry
// in [Content_Types].xml. Returns [ErrPathTraversal] or [ErrInvalidPartName] for
// invalid names, and [ErrDuplicatePart] if the name was already added.
func (pw *PackageWriter) AddPart(name, contentType string, body io.Reader) error {
	norm, err := NormalizePartName(name)
	if err != nil {
		return err
	}
	if _, exists := pw.parts[norm]; exists {
		return ErrDuplicatePart
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	pw.parts[norm] = data
	if contentType != "" {
		pw.ct.Override = append(pw.ct.Override, CTOverride{
			PartName:    norm,
			ContentType: contentType,
		})
	}
	return nil
}

// AddPartBytes is a convenience wrapper for [AddPart] with a []byte body.
func (pw *PackageWriter) AddPartBytes(name, contentType string, body []byte) error {
	return pw.AddPart(name, contentType, bytes.NewReader(body))
}

// AddRelationships stores a Relationships set for the given part.
// Pass partName="" or "/" for the package-level relationships file (_rels/.rels).
// For a part like "/word/document.xml" the rels file will be /word/_rels/document.xml.rels.
func (pw *PackageWriter) AddRelationships(partName string, rels *Relationships) error {
	relsPath := partToRelsPath(partName)
	pw.rels[relsPath] = rels
	return nil
}

// Close finalizes the package:
//  1. Populate [Content_Types].xml Default entries from part extensions.
//  2. Write [Content_Types].xml as the first ZIP entry.
//  3. Write all .rels files in lexicographic order.
//  4. Write all other parts in lexicographic order.
//  5. Close the underlying zip.Writer.
func (pw *PackageWriter) Close() error {
	if pw.closed {
		return ErrInvalidArchive // writer already closed
	}
	pw.closed = true

	pw.populateDefaults()

	ctXML, err := marshalContentTypes(pw.ct)
	if err != nil {
		return err
	}
	if err := pw.writeEntry("[Content_Types].xml", ctXML); err != nil {
		return err
	}

	// Write .rels entries sorted.
	relsPaths := make([]string, 0, len(pw.rels))
	for k := range pw.rels {
		relsPaths = append(relsPaths, k)
	}
	sort.Strings(relsPaths)
	for _, rp := range relsPaths {
		data, err := marshalRelationships(pw.rels[rp])
		if err != nil {
			return err
		}
		zipName, err := ZipEntryName(rp)
		if err != nil {
			return err
		}
		if err := pw.writeEntry(zipName, data); err != nil {
			return err
		}
	}

	// Write all other parts sorted.
	partNames := make([]string, 0, len(pw.parts))
	for k := range pw.parts {
		partNames = append(partNames, k)
	}
	sort.Strings(partNames)
	for _, pn := range partNames {
		zipName, err := ZipEntryName(pn)
		if err != nil {
			return err
		}
		if err := pw.writeEntry(zipName, pw.parts[pn]); err != nil {
			return err
		}
	}

	return pw.zw.Close()
}

func (pw *PackageWriter) writeEntry(name string, data []byte) error {
	hdr := &zip.FileHeader{
		Name:     name,
		Method:   zip.Deflate,
		Modified: zipEpoch, // fixed date for reproducible output
	}
	w, err := pw.zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// populateDefaults adds Default content-type entries for file extensions seen in parts,
// using [ExtensionToContentType] as the registry. Always ensures "rels" has a Default entry.
func (pw *PackageWriter) populateDefaults() {
	seen := make(map[string]bool, len(pw.ct.Default))
	for _, d := range pw.ct.Default {
		seen[d.Extension] = true
	}
	// Ensure rels Default entry is always present.
	if !seen["rels"] {
		pw.ct.Default = append(pw.ct.Default, CTDefault{
			Extension:   "rels",
			ContentType: CTRelsXML,
		})
		seen["rels"] = true
	}
	// Scan part extensions.
	for partName := range pw.parts {
		ext := extensionOf(partName)
		if ext == "" || seen[ext] {
			continue
		}
		if ct, ok := ExtensionToContentType[ext]; ok {
			pw.ct.Default = append(pw.ct.Default, CTDefault{
				Extension:   ext,
				ContentType: ct,
			})
			seen[ext] = true
		}
	}
}

// extensionOf returns the lowercase file extension of a normalized part name (without the dot).
func extensionOf(partName string) string {
	dot := strings.LastIndex(partName, ".")
	if dot < 0 {
		return ""
	}
	slash := strings.LastIndex(partName, "/")
	if slash > dot {
		return "" // dot is in a directory segment, not the file name
	}
	return strings.ToLower(partName[dot+1:])
}

// partToRelsPath converts a part name to its corresponding .rels part path.
func partToRelsPath(partName string) string {
	if partName == "" || partName == "/" {
		return "/_rels/.rels"
	}
	i := strings.LastIndex(partName, "/")
	dir := partName[:i]
	file := partName[i+1:]
	if dir == "" {
		return "/_rels/" + file + ".rels"
	}
	return dir + "/_rels/" + file + ".rels"
}
