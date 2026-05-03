package wml

// Numbering is parsed word/numbering.xml.
type Numbering struct {
	Abstract map[int]*AbstractNum
	Nums     map[int]*NumDef
}

// AbstractNum is w:abstractNum (keyed by abstractNumId).
type AbstractNum struct {
	ID     int
	Levels []*NumLevel // index = ilvl
}

// NumDef maps w:numId to abstract numbering + level overrides.
type NumDef struct {
	NumID      int
	AbstractID int
	Levels     []*NumLevel // copy from abstract, can override per num
}

// NumLevel is one w:lvl.
type NumLevel struct {
	Ilvl    int
	Format  string // numFmt val
	Text    string // lvlText %1.
	Restart int    // startOverride or lvlRestart
	StartAt int    // start val
	Raw     []byte
}
