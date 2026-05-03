package wml

// Document is the parsed main document part (word/document.xml body).
type Document struct {
	Body Body
}

// Body holds ordered block-level content in the document body.
type Body struct {
	// Blocks preserves document order (paragraphs, tables, unknown XML).
	Blocks []BodyBlock
	// SectPr is the final w:sectPr element of w:body (document-level section), if any.
	SectPr []byte
}

// BodyBlock is a union of top-level body elements.
type BodyBlock struct {
	Para    *Paragraph
	Table   *Table
	Unknown []byte // raw XML for unrecognized block-level elements
}

// Paragraph is one w:p (paragraph-level; may appear in body or table cell).
type Paragraph struct {
	Runs    []*Run
	Unknown []byte // raw XML inside w:p excluding parsed runs (for round-trip)

	// OFFICE-102 fields (filled when parser extended)
	PPr ParagraphProps
}

// ParagraphProps holds w:pPr-derived data (OFFICE-102).
type ParagraphProps struct {
	Alignment Alignment
	Indent    Indent
	Spacing   Spacing
	StyleID   string
	Numbering *NumPr
	RawPPr    []byte // unparsed w:pPr remainder for round-trip
	SectPr    []byte // w:sectPr inside last para of section (106)
}

// Alignment is w:jc.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignRight
	AlignCenter
	AlignJustify
	AlignDistribute
	AlignStart
	AlignEnd
)

// Indent is w:ind (twentieths of a point / dxa where applicable).
type Indent struct {
	Left, Right, FirstLine, Hanging int64 // 0 = unset
}

// LineRule is w:spacing lineRule.
type LineRule int

const (
	LineRuleUnset LineRule = iota
	LineRuleAuto
	LineRuleExact
	LineRuleAtLeast
)

// Spacing is w:spacing.
type Spacing struct {
	Before, After int64 // twips, 0 = unset
	Line          int64
	LineRule      LineRule
}

// NumPr is w:numPr (paragraph numbering reference).
type NumPr struct {
	NumID int
	Ilvl  int
}

// RunPart is one logical segment inside a run (text, special char, or raw XML).
type RunPart struct {
	Text       string
	Tab        bool
	Br         bool
	SoftHyphen bool
	Unknown    []byte
}

// Run is one w:r.
type Run struct {
	Parts   []RunPart // ordered; drives marshal round-trip
	Text    string    // cached logical text (tab/br/soft hyphen expanded)
	RPr     RunProps
	Unknown []byte // deprecated: aggregate unknown tail; prefer Parts
}

// RunProps is w:rPr character formatting.
type RunProps struct {
	Bold         bool
	Italic       bool
	Underline    bool
	Strike       bool
	VertAlign    VertAlignKind
	FontSizeHalf int    // w:sz half-points; 0 = unset
	Color        string // RRGGBB without #; empty = unset
	FontName     string // ascii or hAnsi from w:rFonts
	RawRPr       []byte // unparsed w:rPr for round-trip
}

// VertAlignKind is w:vertAlign.
type VertAlignKind int

const (
	VertAlignBaseline VertAlignKind = iota
	VertAlignSuperscript
	VertAlignSubscript
)

// Table is w:tbl.
type Table struct {
	Props   TableProps
	Rows    []*TableRow
	Unknown []byte
}

// TableRow is w:tr.
type TableRow struct {
	Cells []*TableCell
}

// TableCell is w:tc.
type TableCell struct {
	Blocks []BodyBlock // paragraphs + nested tables
	// OFFICE-103 cell properties
	TcPr TableCellProps
}

// TableCellProps holds w:tcPr (103); zero value = unset.
type TableCellProps struct {
	GridSpan int
	VMerge   VMergeKind
	Width    TableWidth
	Borders  *TcBorders
	Shading  *Shading
	RawTcPr  []byte
}

// VMergeKind is w:vMerge val.
type VMergeKind int

const (
	VMergeNone VMergeKind = iota
	VMergeRestart
	VMergeContinue
)

// TableWidth is tblW / tcW.
type TableWidth struct {
	Value int64
	Kind  WidthKind
}

// WidthKind is w:type on tblW/tcW.
type WidthKind int

const (
	WidthAuto WidthKind = iota
	WidthDxa
	WidthPct
)

// TcBorders groups cell borders (103).
type TcBorders struct {
	Top, Left, Bottom, Right, InsideH, InsideV *BorderDef
}

// BorderDef is one w:top etc.
type BorderDef struct {
	Val   string
	Color string
	Size  int // eighths of a point
	Space int
}

// Shading is w:shd.
type Shading struct {
	Fill  string
	Color string
	Val   string
}

// TableProps is w:tblPr / tblGrid (103).
type TableProps struct {
	Width TableWidth
	Raw   []byte
}
