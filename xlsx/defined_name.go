package xlsx

// DefinedName is a workbook or sheet-scoped defined name from workbook.xml.
type DefinedName struct {
	name         string
	formula      string
	localSheetID *int
}

// Name returns the defined name identifier.
func (d *DefinedName) Name() string {
	if d == nil {
		return ""
	}
	return d.name
}

// Formula returns the stored formula string (e.g. "Sheet1!$A$1" or "=A1").
func (d *DefinedName) Formula() string {
	if d == nil {
		return ""
	}
	return d.formula
}

// LocalSheetID returns the 0-based sheet index when the name is sheet-scoped.
func (d *DefinedName) LocalSheetID() (index int, ok bool) {
	if d == nil || d.localSheetID == nil {
		return 0, false
	}
	return *d.localSheetID, true
}
