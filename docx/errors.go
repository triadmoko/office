package docx

import "errors"

var (
	ErrNotDOCX         = errors.New("docx: not a WordprocessingML package")
	ErrMissingMainPart = errors.New("docx: main document part not found")
	ErrMalformedBody   = errors.New("docx: document body could not be read")
)
