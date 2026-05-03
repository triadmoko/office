package wml

import (
	"encoding/xml"
	"io"
)

// ParseNumbering parses word/numbering.xml.
func ParseNumbering(r io.Reader) (*Numbering, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	n := &Numbering{
		Abstract: make(map[int]*AbstractNum),
		Nums:     make(map[int]*NumDef),
	}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if !isWML(se.Name.Space, se.Name.Local) {
			continue
		}
		switch se.Name.Local {
		case "abstractNum":
			aid := intAttr(attrLocal(se.Attr, "abstractNumId"))
			an, err := parseAbstractNum(dec, se, aid)
			if err != nil {
				return nil, err
			}
			n.Abstract[aid] = an
		case "num":
			nd, err := parseNum(dec, se)
			if err != nil {
				return nil, err
			}
			if nd != nil {
				n.Nums[nd.NumID] = nd
			}
		default:
			if err := skipSubtree(dec, se); err != nil {
				return nil, err
			}
		}
	}
	// Link num levels from abstract definitions.
	for id, def := range n.Nums {
		if def == nil {
			continue
		}
		if ab := n.Abstract[def.AbstractID]; ab != nil && len(def.Levels) == 0 {
			def.Levels = ab.Levels
			n.Nums[id] = def
		}
	}
	return n, nil
}

func parseAbstractNum(dec *xml.Decoder, start xml.StartElement, id int) (*AbstractNum, error) {
	an := &AbstractNum{ID: id}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "lvl" {
				lvl, err := parseLvl(dec, t)
				if err != nil {
					return nil, err
				}
				for len(an.Levels) <= lvl.Ilvl {
					an.Levels = append(an.Levels, nil)
				}
				an.Levels[lvl.Ilvl] = lvl
				continue
			}
			if err := skipSubtree(dec, t); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "abstractNum" {
				return an, nil
			}
		}
	}
}

func parseLvl(dec *xml.Decoder, start xml.StartElement) (*NumLevel, error) {
	lvl := &NumLevel{Ilvl: intAttr(attrLocal(start.Attr, "ilvl"))}
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) {
				switch t.Name.Local {
				case "numFmt":
					lvl.Format = valAttr(t.Attr)
					if err := skipSubtree(dec, t); err != nil {
						return nil, err
					}
					continue
				case "lvlText":
					lvl.Text = valAttr(t.Attr)
					if err := skipSubtree(dec, t); err != nil {
						return nil, err
					}
					continue
				case "start":
					lvl.StartAt = intAttr(valAttr(t.Attr))
					if err := skipSubtree(dec, t); err != nil {
						return nil, err
					}
					continue
				case "lvlRestart":
					lvl.Restart = intAttr(valAttr(t.Attr))
					if err := skipSubtree(dec, t); err != nil {
						return nil, err
					}
					continue
				}
			}
			if err := skipSubtree(dec, t); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "lvl" {
				return lvl, nil
			}
		}
	}
}

func parseNum(dec *xml.Decoder, start xml.StartElement) (*NumDef, error) {
	def := &NumDef{}
	def.NumID = intAttr(attrLocal(start.Attr, "numId"))
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "abstractNumId" {
				def.AbstractID = intAttr(valAttr(t.Attr))
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
				continue
			}
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "lvlOverride" {
				if err := skipSubtree(dec, t); err != nil {
					return nil, err
				}
				continue
			}
			if err := skipSubtree(dec, t); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if isWML(t.Name.Space, t.Name.Local) && t.Name.Local == "num" {
				return def, nil
			}
		}
	}
}

// ByNumID returns numbering definition for w:numId (OFFICE-105).
func (n *Numbering) ByNumID(id int) *NumDef {
	if n == nil {
		return nil
	}
	return n.Nums[id]
}

// ResolveNumPr maps paragraph numPr to effective level definition.
func (n *Numbering) ResolveNumPr(np *NumPr) *NumLevel {
	if n == nil || np == nil {
		return nil
	}
	def := n.Nums[np.NumID]
	if def == nil || np.Ilvl < 0 || len(def.Levels) <= np.Ilvl {
		return nil
	}
	return def.Levels[np.Ilvl]
}
