package ooxml

import "errors"

var (
	// ErrInvalidArchive indicates the reader is not a valid OOXML ZIP package.
	ErrInvalidArchive = errors.New("ooxml: invalid zip archive")
	// ErrMissingContentTypes indicates [Content_Types].xml is missing.
	ErrMissingContentTypes = errors.New("ooxml: missing [Content_Types].xml")
	// ErrMalformedContentTypes indicates [Content_Types].xml could not be parsed.
	ErrMalformedContentTypes = errors.New("ooxml: malformed [Content_Types].xml")
	// ErrMissingRelationships indicates a required .rels file is missing.
	ErrMissingRelationships = errors.New("ooxml: missing relationships file")
	// ErrMalformedRelationships indicates a .rels file could not be parsed.
	ErrMalformedRelationships = errors.New("ooxml: malformed relationships")
	// ErrPartNotFound indicates a requested part path is not in the package.
	ErrPartNotFound = errors.New("ooxml: part not found")
	// ErrInvalidPartName indicates an empty or malformed OPC part name.
	ErrInvalidPartName = errors.New("ooxml: invalid part name")
	// ErrPathTraversal indicates a part path contains a ".." segment or otherwise escapes package root.
	ErrPathTraversal = errors.New("ooxml: path traversal in part name")
)
