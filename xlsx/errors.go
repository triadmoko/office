package xlsx

import "errors"

var (
	ErrNotXLSX                = errors.New("xlsx: not a SpreadsheetML package")
	ErrMissingMainPart        = errors.New("xlsx: main workbook part not found")
	ErrNotImplemented         = errors.New("xlsx: not implemented")
	ErrSharedStringOutOfRange = errors.New("xlsx: shared string index out of range")
	ErrSheetTooLargeRandom    = errors.New("xlsx: worksheet too large for random access; use streaming Rows()")
	ErrInvalidCellRef         = errors.New("xlsx: invalid cell reference")
	ErrInvalidCellRange       = errors.New("xlsx: invalid cell range")
	ErrReadOnlySheet          = errors.New("xlsx: sheet is read-only")
	ErrSheetStreamLocked      = errors.New("xlsx: sheet locked after StreamWriter use")
)
