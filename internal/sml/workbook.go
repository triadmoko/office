package sml

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
)

// SheetDef is a raw <sheet> from workbook.xml before resolving the worksheet part path.
type SheetDef struct {
	Name    string
	SheetID int
	RID     string // r:id relationship id
	State   string // "", "hidden", "veryHidden"
}

// DefinedNameDef is a raw <definedName> from workbook.xml.
type DefinedNameDef struct {
	Name         string
	Formula      string // element text (Excel formula string)
	LocalSheetID *int   // localSheetId attribute when present (sheet-scoped name)
}

type workbookXML struct {
	XMLName      xml.Name         `xml:"workbook"`
	Sheets       sheetsXML        `xml:"sheets"`
	DefinedNames *definedNamesXML `xml:"definedNames"`
}

type sheetsXML struct {
	Sheet []sheetXML `xml:"sheet"`
}

type sheetXML struct {
	Name    string `xml:"name,attr"`
	SheetID string `xml:"sheetId,attr"`
	State   string `xml:"state,attr"`
	RID     string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
}

type definedNamesXML struct {
	DefinedName []definedNameXML `xml:"definedName"`
}

type definedNameXML struct {
	Name         string `xml:"name,attr"`
	LocalSheetID string `xml:"localSheetId,attr"`
	Inner        string `xml:",chardata"`
}

// ParseWorkbook parses xl/workbook.xml from r (main workbook part body).
func ParseWorkbook(r io.Reader) ([]SheetDef, []DefinedNameDef, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false
	var root workbookXML
	if err := dec.Decode(&root); err != nil {
		return nil, nil, err
	}
	outSheets := make([]SheetDef, 0, len(root.Sheets.Sheet))
	for _, s := range root.Sheets.Sheet {
		sid, _ := strconv.Atoi(strings.TrimSpace(s.SheetID))
		outSheets = append(outSheets, SheetDef{
			Name:    s.Name,
			SheetID: sid,
			RID:     strings.TrimSpace(s.RID),
			State:   strings.TrimSpace(s.State),
		})
	}
	var names []DefinedNameDef
	if root.DefinedNames != nil {
		for _, dn := range root.DefinedNames.DefinedName {
			var local *int
			if ls := strings.TrimSpace(dn.LocalSheetID); ls != "" {
				if v, err := strconv.Atoi(ls); err == nil {
					local = new(int)
					*local = v
				}
			}
			names = append(names, DefinedNameDef{
				Name:         strings.TrimSpace(dn.Name),
				Formula:      strings.TrimSpace(dn.Inner),
				LocalSheetID: local,
			})
		}
	}
	return outSheets, names, nil
}
