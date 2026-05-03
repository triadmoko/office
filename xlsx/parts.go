package xlsx

// PartBytes returns a copy of raw OPC part bytes captured at [Open], or nil for [NewWorkbook] workbooks.
func (w *Workbook) PartBytes() map[string][]byte {
	if w == nil || len(w.partData) == 0 {
		return nil
	}
	out := make(map[string][]byte, len(w.partData))
	for k, v := range w.partData {
		cp := make([]byte, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}
