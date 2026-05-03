package ooxml

import (
	"archive/zip"
	"bytes"
	"errors"
	"strings"
	"testing"
)

// buildZipWithEntries creates a minimal but valid OOXML ZIP with n identical XML entries.
func buildZipWithEntries(t *testing.T, n int) []byte {
	t.Helper()
	const ct = `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="xml" ContentType="application/xml"/></Types>`
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	writeZip := func(name, body string) {
		t.Helper()
		f, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	writeZip("[Content_Types].xml", ct)
	for i := range n {
		writeZip(strings.Repeat("a", 10)+".xml", "<x/>")
		_ = i
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// buildZipWithBodySized creates a ZIP whose [Content_Types].xml has an honest body of bodySize bytes.
// The ZIP central directory will correctly report UncompressedSize64 = bodySize.
func buildZipWithBodySized(t *testing.T, bodySize int) []byte {
	t.Helper()
	const prefix = `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`
	const suffix = `</Types>`
	// Pad with XML comments to reach bodySize.
	padding := bodySize - len(prefix) - len(suffix)
	padding = max(padding, 0)
	body := prefix + strings.Repeat(" ", padding) + suffix

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create("[Content_Types].xml")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte(body))
	zw.Close()
	return buf.Bytes()
}

func TestOpenWithOptionsMaxParts(t *testing.T) {
	data := buildZipWithEntries(t, 5) // 5 extra entries + [Content_Types].xml = 6 total
	opts := OpenOptions{MaxParts: 3}
	_, err := OpenWithOptions(bytes.NewReader(data), int64(len(data)), opts)
	if !errors.Is(err, ErrTooManyParts) {
		t.Errorf("expected ErrTooManyParts, got %v", err)
	}
}

func TestOpenWithOptionsMaxPartsOK(t *testing.T) {
	data := buildZipWithEntries(t, 2) // 3 total entries — within limit
	opts := OpenOptions{MaxParts: 10}
	_, err := OpenWithOptions(bytes.NewReader(data), int64(len(data)), opts)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenWithOptionsMaxPartBytesHeader(t *testing.T) {
	// Build a ZIP where [Content_Types].xml is ~2000 bytes. The header will honestly
	// report UncompressedSize64 ≈ 2000, which exceeds our MaxPartBytes of 500.
	data := buildZipWithBodySized(t, 2000)
	opts := OpenOptions{MaxPartBytes: 500}
	_, err := OpenWithOptions(bytes.NewReader(data), int64(len(data)), opts)
	if !errors.Is(err, ErrPackageTooLarge) {
		t.Errorf("expected ErrPackageTooLarge from header pre-check, got %v", err)
	}
}

func TestOpenWithOptionsZeroMeansUnlimited(t *testing.T) {
	// 6 entries, but MaxParts=0 means unlimited.
	data := buildZipWithEntries(t, 5)
	opts := OpenOptions{MaxParts: 0}
	_, err := OpenWithOptions(bytes.NewReader(data), int64(len(data)), opts)
	if err != nil {
		t.Errorf("MaxParts=0 should be unlimited, got %v", err)
	}
}

func TestOpenDefaultCallThrough(t *testing.T) {
	// Open() should work identically to OpenWithOptions with defaultOpenOptions for a valid file.
	var buf bytes.Buffer
	pw := NewPackageWriter(&buf)
	pw.AddPartBytes("/word/document.xml", CTWordDocumentMain, []byte("<doc/>"))
	pw.AddRelationships("", &Relationships{
		Relationship: []Relationship{{
			ID:     "rId1",
			Type:   NSRelOfficeDocument,
			Target: "word/document.xml",
		}},
	})
	pw.Close()
	data := buf.Bytes()

	pkg1, err1 := Open(bytes.NewReader(data), int64(len(data)))
	pkg2, err2 := OpenWithOptions(bytes.NewReader(data), int64(len(data)), defaultOpenOptions)
	if err1 != nil || err2 != nil {
		t.Fatalf("Open=%v, OpenWithOptions=%v", err1, err2)
	}
	if pkg1.HasPart("/word/document.xml") != pkg2.HasPart("/word/document.xml") {
		t.Error("Open and OpenWithOptions return different HasPart results")
	}
}

func TestOpenWithOptionsReadLimitEnforced(t *testing.T) {
	// Build a package where the [Content_Types].xml body is larger than MaxPartBytes.
	// The header is honest (small), but MaxPartBytes is set very small so the read limit triggers.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Write a [Content_Types].xml with 1000 bytes of valid XML.
	ctBody := `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
		strings.Repeat(`<Default Extension="xml" ContentType="application/xml"/>`, 10) +
		`</Types>`
	f, _ := zw.Create("[Content_Types].xml")
	f.Write([]byte(ctBody))
	zw.Close()

	data := buf.Bytes()
	// Set MaxPartBytes to 50 bytes — smaller than ctBody — so read limit fires.
	opts := OpenOptions{MaxPartBytes: 50}
	_, err := OpenWithOptions(bytes.NewReader(data), int64(len(data)), opts)
	if !errors.Is(err, ErrPackageTooLarge) {
		t.Errorf("expected ErrPackageTooLarge from read limit, got %v", err)
	}
}
