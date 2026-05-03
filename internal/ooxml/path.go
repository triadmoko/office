package ooxml

import "strings"

// NormalizePartName returns a canonical package part path: a single leading slash,
// no "." or ".." path segments (any ".." is rejected with [ErrPathTraversal]).
func NormalizePartName(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "/")
	if name == "" {
		return "", ErrInvalidPartName
	}
	if strings.Contains(name, "\x00") {
		return "", ErrInvalidPartName
	}
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	var segs []string
	for _, s := range strings.Split(name, "/") {
		switch s {
		case "", ".":
			continue
		case "..":
			return "", ErrPathTraversal
		default:
			segs = append(segs, s)
		}
	}
	if len(segs) == 0 {
		return "/", nil
	}
	return "/" + strings.Join(segs, "/"), nil
}

// ZipEntryName converts a package part name (e.g. "/word/document.xml") to the
// path stored in the ZIP central directory (without leading slash).
func ZipEntryName(partName string) (string, error) {
	p, err := NormalizePartName(partName)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(p, "/"), nil
}
