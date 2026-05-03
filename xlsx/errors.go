package xlsx

import "errors"

var (
	ErrNotXLSX         = errors.New("xlsx: not a SpreadsheetML package")
	ErrMissingMainPart = errors.New("xlsx: main workbook part not found")
	ErrNotImplemented  = errors.New("xlsx: not implemented")
)
