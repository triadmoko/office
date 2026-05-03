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
