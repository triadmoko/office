package xlsx

import "fmt"

// NewWorkbook returns an empty workbook for programmatic generation ([Workbook.Save]).
func NewWorkbook() *Workbook {
	return &Workbook{
		fromNew:  true,
		main:     "/xl/workbook.xml",
		styleReg: newStyleRegistry(),
		partData: nil,
	}
}

// AddSheet appends a worksheet and returns it.
func (w *Workbook) AddSheet(name string) *Sheet {
	if w == nil || !w.fromNew {
		return nil
	}
	id := len(w.newSheets) + 1
	part := fmt.Sprintf("/xl/worksheets/sheet%d.xml", id)
	ws := newWriteSheet(name, id)
	sh := &Sheet{
		wb:      w,
		name:    name,
		sheetID: id,
		rid:     fmt.Sprintf("rId%d", id+3),
		state:   SheetVisible,
		part:    part,
		ws:      ws,
	}
	w.newSheets = append(w.newSheets, sh)
	return sh
}
