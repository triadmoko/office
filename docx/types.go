package docx

import "github.com/triadmoko/office/internal/wml"

// Alignment is paragraph alignment (w:jc).
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

// Indent is paragraph indentation in twentieths of a point (twips).
type Indent struct {
	Left, Right, FirstLine, Hanging int64
}

// LineRule is w:spacing/@w:lineRule.
type LineRule int

const (
	LineRuleUnset LineRule = iota
	LineRuleAuto
	LineRuleExact
	LineRuleAtLeast
)

// Spacing is w:spacing.
type Spacing struct {
	Before, After, Line int64
	LineRule            LineRule
}

// NumPr is paragraph numbering (w:numPr).
type NumPr struct {
	NumID int
	Ilvl  int
}

// VertAlign is run vertical alignment (w:vertAlign).
type VertAlign int

const (
	VertAlignBaseline VertAlign = iota
	VertAlignSuperscript
	VertAlignSubscript
)

// VMergeKind is table cell vertical merge state.
type VMergeKind int

const (
	VMergeNone VMergeKind = iota
	VMergeRestart
	VMergeContinue
)

// WidthKind is tblW/tcW type.
type WidthKind int

const (
	WidthAuto WidthKind = iota
	WidthDxa
	WidthPct
)

// TableWidth is cell or table width.
type TableWidth struct {
	Value int64
	Kind  WidthKind
}

// BorderKind is w:val for borders.
type BorderKind string

const (
	BorderSingle BorderKind = "single"
	BorderNone   BorderKind = "none"
	BorderThick  BorderKind = "thick"
)

// BorderStyle describes a table border side.
type BorderStyle struct {
	Color string
	Size  int // eighths of a point
	Kind  BorderKind
}

// BorderMask selects which sides to apply (OFFICE-108).
type BorderMask int

const (
	BorderTop BorderMask = 1 << iota
	BorderLeft
	BorderBottom
	BorderRight
	BorderInsideH
	BorderInsideV
)

// BorderAll applies borders on all cell sides (table API).
const BorderAll = BorderTop | BorderLeft | BorderBottom | BorderRight | BorderInsideH | BorderInsideV

// PageSizeKind selects standard page dimensions.
type PageSizeKind int

const (
	PageSizeUnset PageSizeKind = iota
	PageSizeA4
	PageSizeLetter
)

func fromWMLAlignment(a wml.Alignment) Alignment {
	switch a {
	case wml.AlignRight:
		return AlignRight
	case wml.AlignCenter:
		return AlignCenter
	case wml.AlignJustify:
		return AlignJustify
	case wml.AlignDistribute:
		return AlignDistribute
	case wml.AlignStart:
		return AlignStart
	case wml.AlignEnd:
		return AlignEnd
	default:
		return AlignLeft
	}
}

func toWMLAlignment(a Alignment) wml.Alignment {
	switch a {
	case AlignRight:
		return wml.AlignRight
	case AlignCenter:
		return wml.AlignCenter
	case AlignJustify:
		return wml.AlignJustify
	case AlignDistribute:
		return wml.AlignDistribute
	case AlignStart:
		return wml.AlignStart
	case AlignEnd:
		return wml.AlignEnd
	default:
		return wml.AlignLeft
	}
}

func fromWMLIndent(i wml.Indent) Indent {
	return Indent{Left: i.Left, Right: i.Right, FirstLine: i.FirstLine, Hanging: i.Hanging}
}

func toWMLIndent(i Indent) wml.Indent {
	return wml.Indent{Left: i.Left, Right: i.Right, FirstLine: i.FirstLine, Hanging: i.Hanging}
}

func fromWMLSpacing(s wml.Spacing) Spacing {
	return Spacing{
		Before: s.Before, After: s.After, Line: s.Line,
		LineRule: fromWMLLineRule(s.LineRule),
	}
}

func toWMLSpacing(s Spacing) wml.Spacing {
	return wml.Spacing{
		Before: s.Before, After: s.After, Line: s.Line,
		LineRule: toWMLLineRule(s.LineRule),
	}
}

func fromWMLLineRule(r wml.LineRule) LineRule {
	switch r {
	case wml.LineRuleAuto:
		return LineRuleAuto
	case wml.LineRuleExact:
		return LineRuleExact
	case wml.LineRuleAtLeast:
		return LineRuleAtLeast
	default:
		return LineRuleUnset
	}
}

func toWMLLineRule(r LineRule) wml.LineRule {
	switch r {
	case LineRuleAuto:
		return wml.LineRuleAuto
	case LineRuleExact:
		return wml.LineRuleExact
	case LineRuleAtLeast:
		return wml.LineRuleAtLeast
	default:
		return wml.LineRuleUnset
	}
}

func fromWMLNum(np *wml.NumPr) *NumPr {
	if np == nil {
		return nil
	}
	return &NumPr{NumID: np.NumID, Ilvl: np.Ilvl}
}

func toWMLNum(np *NumPr) *wml.NumPr {
	if np == nil {
		return nil
	}
	return &wml.NumPr{NumID: np.NumID, Ilvl: np.Ilvl}
}

func fromWMLVert(v wml.VertAlignKind) VertAlign {
	switch v {
	case wml.VertAlignSuperscript:
		return VertAlignSuperscript
	case wml.VertAlignSubscript:
		return VertAlignSubscript
	default:
		return VertAlignBaseline
	}
}

func toWMLVert(v VertAlign) wml.VertAlignKind {
	switch v {
	case VertAlignSuperscript:
		return wml.VertAlignSuperscript
	case VertAlignSubscript:
		return wml.VertAlignSubscript
	default:
		return wml.VertAlignBaseline
	}
}

func fromWMLVMerge(v wml.VMergeKind) VMergeKind {
	switch v {
	case wml.VMergeRestart:
		return VMergeRestart
	case wml.VMergeContinue:
		return VMergeContinue
	default:
		return VMergeNone
	}
}

func toWMLVMerge(v VMergeKind) wml.VMergeKind {
	switch v {
	case VMergeRestart:
		return wml.VMergeRestart
	case VMergeContinue:
		return wml.VMergeContinue
	default:
		return wml.VMergeNone
	}
}

func fromWMLTableWidth(w wml.TableWidth) TableWidth {
	return TableWidth{Value: w.Value, Kind: fromWMLWidthKind(w.Kind)}
}

func toWMLTableWidth(w TableWidth) wml.TableWidth {
	return wml.TableWidth{Value: w.Value, Kind: toWMLWidthKind(w.Kind)}
}

func fromWMLWidthKind(k wml.WidthKind) WidthKind {
	switch k {
	case wml.WidthDxa:
		return WidthDxa
	case wml.WidthPct:
		return WidthPct
	default:
		return WidthAuto
	}
}

func toWMLWidthKind(k WidthKind) wml.WidthKind {
	switch k {
	case WidthDxa:
		return wml.WidthDxa
	case WidthPct:
		return wml.WidthPct
	default:
		return wml.WidthAuto
	}
}
