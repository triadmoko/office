package docx

import "github.com/triadmoko/office/internal/wml"

// ListKind selects bullet vs numbered list (OFFICE-108).
type ListKind int

const (
	ListBullet ListKind = iota
	ListNumbered
)

// List is a bulleted or numbered list under construction.
type List struct {
	doc   *Document
	numID int
	kind  ListKind
}

// AppendList starts a new list (adds numbering definitions when needed).
func (b Body) AppendList(kind ListKind) *List {
	if b.doc == nil {
		return nil
	}
	_, _ = b.doc.ensureLoaded()
	if b.doc.numbering == nil {
		b.doc.numbering = &wml.Numbering{
			Abstract: make(map[int]*wml.AbstractNum),
			Nums:     make(map[int]*wml.NumDef),
		}
	}
	n := b.doc.numbering
	switch kind {
	case ListNumbered:
		ensureAbstractNumbered(n, 2)
		ensureNum(n, 2, 2)
		return &List{doc: b.doc, numID: 2, kind: kind}
	default:
		ensureAbstractBullet(n, 1)
		ensureNum(n, 1, 1)
		return &List{doc: b.doc, numID: 1, kind: kind}
	}
}

func ensureAbstractBullet(num *wml.Numbering, aid int) {
	if num.Abstract[aid] != nil {
		return
	}
	lvl := &wml.NumLevel{
		Ilvl: 0, Format: "bullet", Text: "\u2022", StartAt: 1,
	}
	num.Abstract[aid] = &wml.AbstractNum{
		ID:     aid,
		Levels: []*wml.NumLevel{lvl},
	}
}

func ensureAbstractNumbered(num *wml.Numbering, aid int) {
	if num.Abstract[aid] != nil {
		return
	}
	lvl := &wml.NumLevel{
		Ilvl: 0, Format: "decimal", Text: "%1.", StartAt: 1,
	}
	num.Abstract[aid] = &wml.AbstractNum{
		ID:     aid,
		Levels: []*wml.NumLevel{lvl},
	}
}

func ensureNum(num *wml.Numbering, nid, abstractID int) {
	if num.Nums[nid] != nil {
		return
	}
	num.Nums[nid] = &wml.NumDef{
		NumID:      nid,
		AbstractID: abstractID,
	}
	// Link levels from abstract for marshal.
	if ab := num.Abstract[abstractID]; ab != nil {
		num.Nums[nid].Levels = ab.Levels
	}
}

// AppendItem adds a list item as a new paragraph.
func (l *List) AppendItem(text string) {
	if l == nil || l.doc == nil {
		return
	}
	p := Body{doc: l.doc}.AppendParagraph()
	p.applyListRef(l.numID, 0)
	p.AppendRun(text)
}
