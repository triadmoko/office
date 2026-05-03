package ooxml

// OPC package namespace URIs.
const (
	NSContentTypes  = "http://schemas.openxmlformats.org/package/2006/content-types"
	NSRelationships = "http://schemas.openxmlformats.org/package/2006/relationships"
)

// Relationship type URIs.
const (
	NSRelOfficeDocument = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument"
)

// WordprocessingML namespace URIs and prefixes.
const (
	NSWordprocessingML      = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	NSWordprocessingML14    = "http://schemas.microsoft.com/office/word/2010/wordml"
	NSWordprocessingML15    = "http://schemas.microsoft.com/office/word/2012/wordml"
	NSWordprocessingDrawing = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
	NSWordprocessingDrawing14 = "http://schemas.microsoft.com/office/word/2010/wordprocessingDrawing"

	PrefixWordprocessingML        = "w"
	PrefixWordprocessingML14      = "w14"
	PrefixWordprocessingML15      = "w15"
	PrefixWordprocessingDrawing   = "wp"
	PrefixWordprocessingDrawing14 = "wp14"
)

// SpreadsheetML namespace URIs and prefixes.
const (
	NSSpreadsheetML     = "http://schemas.openxmlformats.org/spreadsheetml/2006/main"
	NSSpreadsheetMLRev  = "http://schemas.microsoft.com/office/spreadsheetml/2014/revision"
	NSSpreadsheetMLRev2 = "http://schemas.microsoft.com/office/spreadsheetml/2015/02/main"
	NSSpreadsheetMLRev3 = "http://schemas.microsoft.com/office/spreadsheetml/2016/02/main"
	NSMarkupCompat      = "http://schemas.openxmlformats.org/markup-compatibility/2006"

	PrefixSpreadsheetML  = "x"
	PrefixSpreadsheetRev  = "xr"
	PrefixSpreadsheetRev2 = "xr2"
	PrefixSpreadsheetRev3 = "xr3"
	PrefixMarkupCompat   = "mc"
)

// PresentationML namespace URIs and prefixes.
const (
	NSPresentationML   = "http://schemas.openxmlformats.org/presentationml/2006/main"
	NSPresentationML14 = "http://schemas.microsoft.com/office/powerpoint/2010/main"
	NSPresentationML15 = "http://schemas.microsoft.com/office/powerpoint/2012/main"

	PrefixPresentationML   = "p"
	PrefixPresentationML14 = "p14"
	PrefixPresentationML15 = "p15"
)

// DrawingML namespace URIs and prefixes.
const (
	NSDrawingML     = "http://schemas.openxmlformats.org/drawingml/2006/main"
	NSDrawingML14   = "http://schemas.microsoft.com/office/drawing/2010/main"
	NSPicture       = "http://schemas.openxmlformats.org/drawingml/2006/picture"
	NSRelMarkup     = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"

	PrefixDrawingML     = "a"
	PrefixDrawingML14   = "a14"
	PrefixPicture       = "pic"
	PrefixRelMarkup     = "r"
)

// ECMA-376 content types.
const (
	// Package-level
	CTRelsXML  = "application/vnd.openxmlformats-package.relationships+xml"
	CTCoreProps = "application/vnd.openxmlformats-package.core-properties+xml"

	// WordprocessingML
	CTWordDocumentMain = "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"
	CTWordStyles       = "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"
	CTWordNumbering    = "application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"
	CTWordSettings     = "application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"
	CTWordFontTable    = "application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"
	CTWordWebSettings  = "application/vnd.openxmlformats-officedocument.wordprocessingml.webSettings+xml"

	// SpreadsheetML
	CTSpreadsheetMain = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"

	// PresentationML
	CTPresentationMain = "application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"

	// Shared
	CTTheme    = "application/vnd.openxmlformats-officedocument.theme+xml"
	CTAppProps = "application/vnd.openxmlformats-officedocument.extended-properties+xml"

	// Images
	CTImagePNG  = "image/png"
	CTImageJPEG = "image/jpeg"
	CTImageGIF  = "image/gif"
	CTImageBMP  = "image/bmp"
)

// ExtensionToContentType maps common file extensions to their default OPC content types.
// Used by PackageWriter to auto-populate [Content_Types].xml Default entries.
var ExtensionToContentType = map[string]string{
	"xml":  "application/xml",
	"rels": CTRelsXML,
	"png":  CTImagePNG,
	"jpg":  CTImageJPEG,
	"jpeg": CTImageJPEG,
	"gif":  CTImageGIF,
	"bmp":  CTImageBMP,
	"bin":  "application/octet-stream",
	"vml":  "application/vnd.openxmlformats-officedocument.vmlDrawing",
}
