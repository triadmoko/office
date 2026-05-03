package xlsx

// Hyperlink is a cell hyperlink (read support may expand in future releases).
type Hyperlink struct {
	URL     string
	Display string
}

// Hyperlink returns the hyperlink for this cell when read from a file, or nil.
func (c *Cell) Hyperlink() *Hyperlink {
	return nil
}
