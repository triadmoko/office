package docx

import "errors"

var (
	ErrNotDOCX         = errors.New("docx: not a WordprocessingML package")
	ErrMissingMainPart = errors.New("docx: main document part not found")
	ErrMalformedBody   = errors.New("docx: document body could not be read")
	// ErrFooterPageNumberOpenDoc is returned when SetFooterPageNumber(true) is used on a document opened from disk (MVP: only NewDocument+Save).
	ErrFooterPageNumberOpenDoc = errors.New("docx: footer page number is only supported for NewDocument until package merge is implemented")
	// ErrHeaderPageNumberOpenDoc is returned when SetHeaderPageNumber(true) is used on a document opened from disk (MVP: only NewDocument+Save).
	ErrHeaderPageNumberOpenDoc = errors.New("docx: header page number is only supported for NewDocument until package merge is implemented")
)
