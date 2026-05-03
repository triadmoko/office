package wml

import "strings"

// PlainText returns all logical run text in the document in reading order,
// including text inside tables and nested tables.
func (d *Document) PlainText() string {
	if d == nil {
		return ""
	}
	var b strings.Builder
	walkBodyBlocks(&d.Body, func(p *Paragraph) {
		for _, r := range p.Runs {
			b.WriteString(r.Text)
		}
	})
	return b.String()
}

func walkBodyBlocks(body *Body, fn func(*Paragraph)) {
	if body == nil {
		return
	}
	for _, bl := range body.Blocks {
		switch {
		case bl.Para != nil:
			fn(bl.Para)
		case bl.Table != nil:
			walkTable(bl.Table, fn)
		}
	}
}

func walkTable(t *Table, fn func(*Paragraph)) {
	if t == nil {
		return
	}
	for _, row := range t.Rows {
		if row == nil {
			continue
		}
		for _, cell := range row.Cells {
			if cell == nil {
				continue
			}
			for _, cb := range cell.Blocks {
				switch {
				case cb.Para != nil:
					fn(cb.Para)
				case cb.Table != nil:
					walkTable(cb.Table, fn)
				}
			}
		}
	}
}

// DirectParagraphs returns only top-level w:p in document body order (excludes table cell paragraphs).
func (d *Document) DirectParagraphs() []*Paragraph {
	if d == nil {
		return nil
	}
	var out []*Paragraph
	for _, bl := range d.Body.Blocks {
		if bl.Para != nil {
			out = append(out, bl.Para)
		}
	}
	return out
}
