package ooxml

// XML namespace URIs used across OOXML packages.
const (
	NSContentTypes        = "http://schemas.openxmlformats.org/package/2006/content-types"
	NSRelationships       = "http://schemas.openxmlformats.org/package/2006/relationships"
	NSRelOfficeDocument   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
	NSWordprocessingMain  = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
)

// ECMA-376 content types for package validation.
const (
	CTWordDocumentMain = "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"
	CTSpreadsheetMain  = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"
	CTPresentationMain = "application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"
	CTRelsXML          = "application/vnd.openxmlformats-package.relationships+xml"
)
