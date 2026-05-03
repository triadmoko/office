package docx

// PlainText returns concatenated logical text from the parsed document model
// (including text inside tables), preserving w:t, tab, break, and soft hyphen semantics.
func (d *Document) PlainText() (string, error) {
	if d == nil {
		return "", ErrMissingMainPart
	}
	m, err := d.ensureLoaded()
	if err != nil {
		return "", err
	}
	if m == nil {
		return "", ErrMissingMainPart
	}
	return m.PlainText(), nil
}
