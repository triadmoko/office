package docx

import "github.com/triadmoko/office/internal/wml"

// Numbering is word/numbering.xml (lists).
type Numbering struct {
	x *wml.Numbering
}

// Numbering returns the document numbering model or nil.
func (d *Document) Numbering() *Numbering {
	if d == nil {
		return nil
	}
	if _, err := d.ensureLoaded(); err != nil || d.numbering == nil {
		return nil
	}
	return &Numbering{x: d.numbering}
}

// ByNumID returns a concrete numbering instance (w:num).
func (n *Numbering) ByNumID(id int) *NumDef {
	if n == nil || n.x == nil {
		return nil
	}
	d := n.x.ByNumID(id)
	if d == nil {
		return nil
	}
	return &NumDef{d: d}
}

// NumDef is one w:num with resolved levels when available.
type NumDef struct {
	d *wml.NumDef
}

// Levels returns up to 9 list levels (MVP may only populate level 0).
func (nd *NumDef) Levels() []*NumLevel {
	if nd == nil || nd.d == nil {
		return nil
	}
	var out []*NumLevel
	for _, lv := range nd.d.Levels {
		if lv == nil {
			continue
		}
		out = append(out, &NumLevel{x: lv})
	}
	return out
}

// NumLevel is one w:lvl.
type NumLevel struct {
	x *wml.NumLevel
}

// Format returns w:numFmt (decimal, bullet, upperRoman, …).
func (nl *NumLevel) Format() string {
	if nl == nil || nl.x == nil {
		return ""
	}
	return nl.x.Format
}

// Text returns w:lvlText template (e.g. "%1.").
func (nl *NumLevel) Text() string {
	if nl == nil || nl.x == nil {
		return ""
	}
	return nl.x.Text
}

// RestartAt returns w:lvlRestart when set.
func (nl *NumLevel) RestartAt() int {
	if nl == nil || nl.x == nil {
		return 0
	}
	return nl.x.Restart
}

// StartAt returns w:start.
func (nl *NumLevel) StartAt() int {
	if nl == nil || nl.x == nil {
		return 0
	}
	return nl.x.StartAt
}
