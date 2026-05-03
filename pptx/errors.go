package pptx

import "errors"

var (
	ErrNotPPTX         = errors.New("pptx: not a PresentationML package")
	ErrMissingMainPart = errors.New("pptx: main presentation part not found")
	ErrNotImplemented  = errors.New("pptx: not implemented")
)
