package docx

import "github.com/triadmoko/office/internal/wml"

// Table is a w:tbl.
type Table struct {
	t   *wml.Table
	doc *Document
}

// Rows returns table rows.
func (t *Table) Rows() []*TableRow {
	if t == nil || t.t == nil {
		return nil
	}
	out := make([]*TableRow, len(t.t.Rows))
	for i, r := range t.t.Rows {
		out[i] = &TableRow{r: r, doc: t.doc}
	}
	return out
}

// Width returns table width from tblPr if set.
func (t *Table) Width() TableWidth {
	if t == nil || t.t == nil {
		return TableWidth{}
	}
	return fromWMLTableWidth(t.t.Props.Width)
}

// SetWidth sets w:tblW for the whole table.
// WidthDxa: Value is in twips (twentieths of a point).
// WidthPct: Value is in fiftyths of a percent per ECMA-376 (5000 = 100% of usable page width).
func (t *Table) SetWidth(w TableWidth) {
	if t == nil || t.t == nil {
		return
	}
	t.t.Props.Width = toWMLTableWidth(w)
}

// GridColWidths returns a copy of w:tblGrid column widths in twips (dxa), or nil if unset.
func (t *Table) GridColWidths() []int64 {
	if t == nil || t.t == nil || len(t.t.Props.GridColWidths) == 0 {
		return nil
	}
	out := make([]int64, len(t.t.Props.GridColWidths))
	copy(out, t.t.Props.GridColWidths)
	return out
}

// SetGridColWidths sets w:tblGrid: one width in twips (dxa) per logical column.
// Pass nil or empty slice to clear.
func (t *Table) SetGridColWidths(widths []int64) {
	if t == nil || t.t == nil {
		return
	}
	if len(widths) == 0 {
		t.t.Props.GridColWidths = nil
		return
	}
	t.t.Props.GridColWidths = append([]int64(nil), widths...)
}

// SetBorder applies the same border style to selected sides on every cell (MVP).
func (t *Table) SetBorder(mask BorderMask, st BorderStyle) {
	if t == nil || t.t == nil {
		return
	}
	bd := &wml.BorderDef{Val: string(st.Kind), Color: st.Color, Size: st.Size}
	for _, row := range t.t.Rows {
		if row == nil {
			continue
		}
		for _, cell := range row.Cells {
			if cell == nil {
				continue
			}
			if cell.TcPr.Borders == nil {
				cell.TcPr.Borders = &wml.TcBorders{}
			}
			b := cell.TcPr.Borders
			if mask&BorderTop != 0 {
				b.Top = cloneBorder(bd)
			}
			if mask&BorderLeft != 0 {
				b.Left = cloneBorder(bd)
			}
			if mask&BorderBottom != 0 {
				b.Bottom = cloneBorder(bd)
			}
			if mask&BorderRight != 0 {
				b.Right = cloneBorder(bd)
			}
			if mask&BorderInsideH != 0 {
				b.InsideH = cloneBorder(bd)
			}
			if mask&BorderInsideV != 0 {
				b.InsideV = cloneBorder(bd)
			}
		}
	}
}

func cloneBorder(b *wml.BorderDef) *wml.BorderDef {
	if b == nil {
		return nil
	}
	c := *b
	return &c
}

// TableRow is w:tr.
type TableRow struct {
	r   *wml.TableRow
	doc *Document
}

// Cells returns row cells.
func (tr *TableRow) Cells() []*TableCell {
	if tr == nil || tr.r == nil {
		return nil
	}
	out := make([]*TableCell, len(tr.r.Cells))
	for i, c := range tr.r.Cells {
		out[i] = &TableCell{c: c, doc: tr.doc}
	}
	return out
}

// SetHeight sets w:trPr/w:trHeight. twips is w:val. If twips > 0 and rule is [RowHeightUnset], AtLeast is used.
func (tr *TableRow) SetHeight(twips int64, rule RowHeightRule) {
	if tr == nil || tr.r == nil {
		return
	}
	tr.r.HeightVal = twips
	h := toWMLTrHeightRule(rule)
	if twips > 0 && h == wml.TrHeightUnset {
		h = wml.TrHeightAtLeast
	}
	tr.r.HeightRule = h
}

// Height returns row height in twips and rule.
func (tr *TableRow) Height() (twips int64, rule RowHeightRule) {
	if tr == nil || tr.r == nil {
		return 0, RowHeightUnset
	}
	return tr.r.HeightVal, fromWMLTrHeightRule(tr.r.HeightRule)
}

// CantSplit reports w:cantSplit (row must not break across pages).
func (tr *TableRow) CantSplit() bool {
	if tr == nil || tr.r == nil {
		return false
	}
	return tr.r.CantSplit
}

// SetCantSplit sets w:cantSplit.
func (tr *TableRow) SetCantSplit(v bool) {
	if tr == nil || tr.r == nil {
		return
	}
	tr.r.CantSplit = v
}

// RepeatAsHeaderRow reports w:tblHeader (repeat row at top of each page).
func (tr *TableRow) RepeatAsHeaderRow() bool {
	if tr == nil || tr.r == nil {
		return false
	}
	return tr.r.TblHeader
}

// SetRepeatAsHeaderRow sets w:tblHeader.
func (tr *TableRow) SetRepeatAsHeaderRow(v bool) {
	if tr == nil || tr.r == nil {
		return
	}
	tr.r.TblHeader = v
}

// TableCell is w:tc.
type TableCell struct {
	c   *wml.TableCell
	doc *Document
}

// Tables returns nested w:tbl inside the cell.
func (tc *TableCell) Tables() []*Table {
	if tc == nil || tc.c == nil {
		return nil
	}
	var out []*Table
	for _, bl := range tc.c.Blocks {
		if bl.Table != nil {
			out = append(out, &Table{t: bl.Table, doc: tc.doc})
		}
	}
	return out
}

// Paragraphs returns paragraphs directly contained in the cell (not nested tables).
func (tc *TableCell) Paragraphs() []*Paragraph {
	if tc == nil || tc.c == nil {
		return nil
	}
	var out []*Paragraph
	for _, bl := range tc.c.Blocks {
		if bl.Para != nil {
			out = append(out, &Paragraph{x: bl.Para, doc: tc.doc})
		}
	}
	return out
}

// GridSpan returns w:gridSpan (1 if unset).
func (tc *TableCell) GridSpan() int {
	if tc == nil || tc.c == nil || tc.c.TcPr.GridSpan == 0 {
		return 1
	}
	return tc.c.TcPr.GridSpan
}

// VMerge returns vertical merge state.
func (tc *TableCell) VMerge() VMergeKind {
	if tc == nil || tc.c == nil {
		return VMergeNone
	}
	return fromWMLVMerge(tc.c.TcPr.VMerge)
}

// Width returns cell width.
func (tc *TableCell) Width() TableWidth {
	if tc == nil || tc.c == nil {
		return TableWidth{}
	}
	return fromWMLTableWidth(tc.c.TcPr.Width)
}

// SetWidth sets w:tcW (same units as [Table.SetWidth]).
func (tc *TableCell) SetWidth(w TableWidth) {
	if tc == nil || tc.c == nil {
		return
	}
	tc.c.TcPr.Width = toWMLTableWidth(w)
}

// Borders returns cell borders or nil.
func (tc *TableCell) Borders() *TcBordersView {
	if tc == nil || tc.c == nil || tc.c.TcPr.Borders == nil {
		return nil
	}
	return &TcBordersView{b: tc.c.TcPr.Borders}
}

// Shading returns cell shading or nil.
func (tc *TableCell) Shading() *ShadingView {
	if tc == nil || tc.c == nil || tc.c.TcPr.Shading == nil {
		return nil
	}
	return &ShadingView{s: tc.c.TcPr.Shading}
}

// TcBordersView wraps cell borders for reading.
type TcBordersView struct {
	b *wml.TcBorders
}

// ShadingView wraps shading.
type ShadingView struct {
	s *wml.Shading
}

// Fill returns w:shd fill color.
func (s *ShadingView) Fill() string {
	if s == nil || s.s == nil {
		return ""
	}
	return s.s.Fill
}
