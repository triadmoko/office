package ooxml

import "strings"

// NormalizePartName returns a part path with a single leading slash, as used in
// [Content_Types].xml PartName and OOXML conventions.
func NormalizePartName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "/")
	if !strings.HasPrefix(name, "/") {
		return "/" + name
	}
	return name
}

// ZipEntryName converts a package part name (e.g. "/word/document.xml") to the
// path stored in the ZIP central directory (usually without leading slash).
func ZipEntryName(partName string) string {
	p := NormalizePartName(partName)
	return strings.TrimPrefix(p, "/")
}
