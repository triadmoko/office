package docx

import (
	"github.com/triadmoko/office/internal/wml"
)

// Body is a view of the document main body for reading and building.
type Body struct {
	doc *Document
}

// Body returns the document body.
func (d *Document) Body() Body {
	return Body{doc: d}
}

// Paragraphs returns top-level w:p elements in body order (not including paragraphs inside table cells).
func (b Body) Paragraphs() []*Paragraph {
	if b.doc == nil {
		return nil
	}
	m, err := b.doc.ensureLoaded()
	if err != nil || m == nil {
		return nil
	}
	list := m.DirectParagraphs()
	out := make([]*Paragraph, len(list))
	for i, p := range list {
		out[i] = &Paragraph{x: p, doc: b.doc}
	}
	return out
}

// Tables returns top-level w:tbl elements in body order.
func (b Body) Tables() []*Table {
	if b.doc == nil {
		return nil
	}
	m, err := b.doc.ensureLoaded()
	if err != nil || m == nil {
		return nil
	}
	var out []*Table
	for _, bl := range m.Body.Blocks {
		if bl.Table != nil {
			out = append(out, &Table{t: bl.Table, doc: b.doc})
		}
	}
	return out
}

// AppendParagraph adds an empty paragraph at the end of the body (builder API).
func (b Body) AppendParagraph() *Paragraph {
	if b.doc == nil {
		return nil
	}
	m, _ := b.doc.ensureLoaded()
	p := &wml.Paragraph{}
	m.Body.Blocks = append(m.Body.Blocks, wml.BodyBlock{Para: p})
	return &Paragraph{x: p, doc: b.doc}
}

// AppendTable adds a rows×cols table with one empty paragraph per cell.
// The table width defaults to the full text area (100% pct); override with [Table.SetWidth] if needed.
func (b Body) AppendTable(rows, cols int) *Table {
	if b.doc == nil || rows < 1 || cols < 1 {
		return nil
	}
	m, _ := b.doc.ensureLoaded()
	tbl := &wml.Table{Rows: make([]*wml.TableRow, rows)}
	// Default: lebar tabel = 100% area teks (w:tblW type pct; 5000 = 100% per ECMA-376).
	tbl.Props.Width = wml.TableWidth{Value: 5000, Kind: wml.WidthPct}
	for r := 0; r < rows; r++ {
		row := &wml.TableRow{Cells: make([]*wml.TableCell, cols)}
		for c := 0; c < cols; c++ {
			row.Cells[c] = &wml.TableCell{
				Blocks: []wml.BodyBlock{{Para: &wml.Paragraph{}}},
			}
		}
		tbl.Rows[r] = row
	}
	m.Body.Blocks = append(m.Body.Blocks, wml.BodyBlock{Table: tbl})
	return &Table{t: tbl, doc: b.doc}
}
