package wml

// Page dimensions in twentieths of a point (twips), per ECMA-376.
const (
	PageLetterW = 12240
	PageLetterH = 15840
	PageA4W     = 11906
	PageA4H     = 16838
)

// Orientation is w:orient on pgSz.
type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

// PageSize is w:pgSz.
type PageSize struct {
	Width, Height int64
	Orient        Orientation
}

// Margins is w:pgMar (twips).
type Margins struct {
	Top, Bottom, Left, Right int64
	Header, Footer, Gutter   int64
}

// Columns is w:cols.
type Columns struct {
	Num        int
	Sep        bool
	EqualWidth bool
}

// Section describes one document section (w:sectPr).
type Section struct {
	PageSize PageSize
	Margins  Margins
	Columns  Columns
	// TypeVal is w:type/@w:val (ST_SectionMark): nextPage, continuous, nextColumn, evenPage, oddPage; empty = omit.
	TypeVal string
	// PageNumFmt is w:pgNumType/@w:fmt (ST_NumberFormat): decimal, upperRoman, lowerRoman, upperLetter, lowerLetter, …; empty = omit.
	PageNumFmt string
	// PageNumStartSet + PageNumStart map w:pgNumType/@w:start (first page number of this section for PAGE field).
	PageNumStartSet bool
	PageNumStart    int
	Raw             []byte
}
