package xlsx

import (
	"fmt"
	"strings"
)

// StyleBuilder configures a cell style for [Sheet.SetCell].
type StyleBuilder struct {
	wb     *Workbook
	numFmt string
	bold   bool
	bg     string // ARGB without #
	regID  int    // assigned lazily, -1 = unset
}

func (wb *Workbook) NewStyle() *StyleBuilder {
	if wb == nil || !wb.fromNew {
		return &StyleBuilder{}
	}
	return &StyleBuilder{wb: wb, numFmt: "General", regID: -1}
}

// NumberFormat sets the number format code (e.g. "0.00", "m/d/yy").
func (b *StyleBuilder) NumberFormat(f string) *StyleBuilder {
	if b != nil {
		b.numFmt = f
	}
	return b
}

// Bold requests bold font (minimal serialization support).
func (b *StyleBuilder) Bold(v bool) *StyleBuilder {
	if b != nil {
		b.bold = v
	}
	return b
}

// Background sets solid fill RGB (with or without leading "#").
func (b *StyleBuilder) Background(rgb string) *StyleBuilder {
	if b != nil {
		b.bg = strings.TrimPrefix(strings.TrimSpace(rgb), "#")
	}
	return b
}

func (b *StyleBuilder) register() (int, error) {
	if b == nil || b.wb == nil || b.wb.styleReg == nil {
		return -1, fmt.Errorf("xlsx: style not bound to workbook")
	}
	if b.regID >= 0 {
		return b.regID, nil
	}
	id := b.wb.styleReg.add(styleSig{
		numFmt: b.numFmt,
		bold:   b.bold,
		bg:     b.bg,
	})
	b.regID = id
	return id, nil
}

type styleSig struct {
	numFmt string
	bold   bool
	bg     string
}

type styleRegistry struct {
	list []styleSig
	idx  map[string]int
}

func newStyleRegistry() *styleRegistry {
	r := &styleRegistry{idx: make(map[string]int)}
	r.add(styleSig{numFmt: "General"}) // index 0 default
	return r
}

func (r *styleRegistry) key(s styleSig) string {
	return s.numFmt + "\x00" + boolStr(s.bold) + "\x00" + s.bg
}

func boolStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (r *styleRegistry) add(s styleSig) int {
	k := r.key(s)
	if id, ok := r.idx[k]; ok {
		return id
	}
	id := len(r.list)
	r.list = append(r.list, s)
	r.idx[k] = id
	return id
}

func (r *styleRegistry) entries() []styleSig {
	return r.list
}
