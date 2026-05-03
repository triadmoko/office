package xlsx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/triadmoko/office/internal/ooxml"
	"github.com/triadmoko/office/internal/opcprops"
	"github.com/triadmoko/office/internal/sml"
)

const (
	ctWorksheet     = "application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"
	ctSharedStrings = "application/vnd.openxmlformats-officedocument.spreadsheetml.sharedStrings+xml"
	ctStyles        = "application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"
)

func (w *Workbook) saveNew(out io.Writer) error {
	if w == nil || !w.fromNew || len(w.newSheets) == 0 {
		return fmt.Errorf("xlsx: no sheets to save")
	}
	for _, sh := range w.newSheets {
		if sh.ws != nil && sh.ws.stream != nil && !sh.ws.stream.closed {
			return fmt.Errorf("xlsx: StreamWriter on sheet %q not flushed", sh.name)
		}
	}
	sst, sstIndex, sstXML := w.computeSharedStringPool()
	pw := ooxml.NewPackageWriter(out)
	if err := pw.AddRelationships("", newRootWorkbookRels()); err != nil {
		return err
	}
	if err := pw.AddRelationships("/xl/workbook.xml", w.workbookRels(sstIndex != nil)); err != nil {
		return err
	}
	now := time.Now().UTC()
	core := &opcprops.CoreProperties{Created: now, Modified: now}
	var coreBuf bytes.Buffer
	if _, err := core.WriteTo(&coreBuf); err != nil {
		return err
	}
	if err := pw.AddPartBytes("/docProps/core.xml", ooxml.CTCoreProps, coreBuf.Bytes()); err != nil {
		return err
	}
	app := &opcprops.AppProperties{}
	var appBuf bytes.Buffer
	if _, err := app.WriteTo(&appBuf); err != nil {
		return err
	}
	if err := pw.AddPartBytes("/docProps/app.xml", ooxml.CTAppProps, appBuf.Bytes()); err != nil {
		return err
	}
	if len(sst) > 0 {
		if err := pw.AddPartBytes("/xl/sharedStrings.xml", ctSharedStrings, sstXML); err != nil {
			return err
		}
	}
	if err := pw.AddPartBytes("/xl/styles.xml", ctStyles, w.buildStylesXML()); err != nil {
		return err
	}
	for _, sh := range w.newSheets {
		if sh.ws == nil {
			continue
		}
		body, rels := w.serializeWorksheet(sh, sstIndex)
		if err := pw.AddPartBytes(sh.part, ctWorksheet, body); err != nil {
			return err
		}
		if rels != nil && len(rels.Relationship) > 0 {
			if err := pw.AddRelationships(sh.part, rels); err != nil {
				return err
			}
		}
	}
	if err := pw.AddPartBytes("/xl/workbook.xml", ooxml.CTSpreadsheetMain, []byte(w.workbookXML())); err != nil {
		return err
	}
	return pw.Close()
}

func newRootWorkbookRels() *ooxml.Relationships {
	return &ooxml.Relationships{
		Relationship: []ooxml.Relationship{
			{ID: "rId1", Type: ooxml.NSRelOfficeDocument, Target: "xl/workbook.xml"},
			{ID: "rId2", Type: "http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties", Target: "docProps/core.xml"},
			{ID: "rId3", Type: "http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties", Target: "docProps/app.xml"},
		},
	}
}

func (w *Workbook) workbookRels(hasSST bool) *ooxml.Relationships {
	var rels []ooxml.Relationship
	rid := 1
	if hasSST {
		rels = append(rels, ooxml.Relationship{ID: fmt.Sprintf("rId%d", rid), Type: relTypeSharedStrings, Target: "sharedStrings.xml"})
		rid++
	}
	rels = append(rels, ooxml.Relationship{ID: fmt.Sprintf("rId%d", rid), Type: relTypeStyles, Target: "styles.xml"})
	rid++
	for i := range w.newSheets {
		rels = append(rels, ooxml.Relationship{
			ID:     fmt.Sprintf("rId%d", rid),
			Type:   relTypeWorksheet,
			Target: fmt.Sprintf("worksheets/sheet%d.xml", i+1),
		})
		rid++
	}
	return &ooxml.Relationships{Relationship: rels}
}

func (w *Workbook) workbookXML() string {
	_, sheetBaseRID := w.workbookSheetRIDBase()
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets>`)
	for i, sh := range w.newSheets {
		rid := sheetBaseRID + i
		b.WriteString(fmt.Sprintf(`<sheet name="%s" sheetId="%d" r:id="rId%d"/>`, xmlEscapeAttr(sh.name), i+1, rid))
	}
	b.WriteString(`</sheets></workbook>`)
	return b.String()
}

func (w *Workbook) workbookSheetRIDBase() (hasSST bool, firstSheetRID int) {
	hasSST = w.hasPooledStrings()
	rid := 1
	if hasSST {
		rid++
	}
	rid++ // styles
	return hasSST, rid
}

func (w *Workbook) hasPooledStrings() bool {
	_, idx, _ := w.computeSharedStringPool()
	return len(idx) > 0
}

func (w *Workbook) computeSharedStringPool() (sst []string, index map[string]int, xmlBytes []byte) {
	counts := make(map[string]int)
	var order []string
	for _, sh := range w.newSheets {
		if sh.ws == nil {
			continue
		}
		for _, c := range sh.ws.cells {
			if c == nil || c.formula != "" {
				continue
			}
			if s, ok := c.val.(string); ok && s != "" {
				counts[s]++
				if counts[s] == 1 {
					order = append(order, s)
				}
			}
		}
	}
	for _, s := range order {
		if counts[s] >= 2 {
			sst = append(sst, s)
		}
	}
	if len(sst) == 0 {
		return nil, nil, nil
	}
	index = make(map[string]int, len(sst))
	for i, s := range sst {
		index[s] = i
	}
	var buf bytes.Buffer
	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	buf.WriteString(`<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="`)
	buf.WriteString(strconv.Itoa(len(sst)))
	buf.WriteString(`" uniqueCount="`)
	buf.WriteString(strconv.Itoa(len(sst)))
	buf.WriteString(`">`)
	for _, s := range sst {
		buf.WriteString("<si><t>")
		xmlEscape(&buf, s)
		buf.WriteString("</t></si>")
	}
	buf.WriteString("</sst>")
	return sst, index, buf.Bytes()
}

func (w *Workbook) buildStylesXML() []byte {
	reg := w.styleReg
	if reg == nil {
		reg = newStyleRegistry()
	}
	var numFmt strings.Builder
	nextID := 164
	fmtToID := map[string]int{"General": 0, "": 0}
	var xfs strings.Builder
	xfs.WriteString(`<cellXfs count="`)
	xfs.WriteString(strconv.Itoa(len(reg.entries())))
	xfs.WriteString(`">`)
	for _, e := range reg.entries() {
		nfid := 0
		if e.numFmt != "" && e.numFmt != "General" {
			if id, ok := fmtToID[e.numFmt]; ok {
				nfid = id
			} else {
				nfid = nextID
				fmtToID[e.numFmt] = nextID
				numFmt.WriteString(fmt.Sprintf(`<numFmt numFmtId="%d" formatCode="%s"/>`, nextID, xmlEscapeAttr(e.numFmt)))
				nextID++
			}
		}
		fontID := 0
		if e.bold {
			fontID = 1
		}
		fillID := 0
		if e.bg != "" {
			fillID = 1
		}
		xfs.WriteString(fmt.Sprintf(`<xf numFmtId="%d" fontId="%d" fillId="%d" borderId="0" xfId="0" applyNumberFormat="1" applyFont="1" applyFill="1"/>`, nfid, fontID, fillID))
	}
	xfs.WriteString(`</cellXfs>`)
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	nfc := strings.Count(numFmt.String(), "<numFmt")
	if nfc > 0 {
		b.WriteString(`<numFmts count="`)
		b.WriteString(strconv.Itoa(nfc))
		b.WriteString(`">`)
		b.WriteString(numFmt.String())
		b.WriteString(`</numFmts>`)
	}
	b.WriteString(`<fonts count="2"><font/><font><b/></font></fonts>`)
	b.WriteString(`<fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FF`)
	b.WriteString(pad6(effectiveFillRGB(w.styleReg)))
	b.WriteString(`"/></patternFill></fill></fills>`)
	b.WriteString(`<borders count="1"><border/></borders>`)
	b.WriteString(`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>`)
	b.WriteString(xfs.String())
	b.WriteString(`</styleSheet>`)
	return []byte(b.String())
}

func effectiveFillRGB(reg *styleRegistry) string {
	if reg == nil {
		return "FFFF00"
	}
	for _, e := range reg.entries() {
		if e.bg != "" {
			s := strings.ToUpper(strings.TrimSpace(e.bg))
			if len(s) >= 6 {
				return s[len(s)-6:]
			}
			return pad6(s)
		}
	}
	return "FFFF00"
}

func pad6(s string) string {
	for len(s) < 6 {
		s = "0" + s
	}
	if len(s) > 6 {
		return s[len(s)-6:]
	}
	return s
}

func (w *Workbook) serializeWorksheet(sh *Sheet, sstIndex map[string]int) ([]byte, *ooxml.Relationships) {
	if sh.ws != nil && sh.ws.stream != nil && sh.ws.stream.closed && sh.ws.stream.fragment.Len() > 0 {
		var b bytes.Buffer
		b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
		b.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
		b.Write(sh.ws.stream.fragment.Bytes())
		b.WriteString(`</worksheet>`)
		if len(sh.ws.stream.rels.Relationship) == 0 {
			return b.Bytes(), nil
		}
		return b.Bytes(), &sh.ws.stream.rels
	}
	addrs := make([]string, 0, len(sh.ws.cells))
	for a := range sh.ws.cells {
		addrs = append(addrs, a)
	}
	sort.Slice(addrs, func(i, j int) bool {
		c1, r1, _ := sml.CellRefToIndexes(addrs[i])
		c2, r2, _ := sml.CellRefToIndexes(addrs[j])
		if r1 != r2 {
			return r1 < r2
		}
		return c1 < c2
	})
	rows := groupWriteRows(addrs)
	var body bytes.Buffer
	body.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	body.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	var hlinks strings.Builder
	var rels ooxml.Relationships
	hid := 1
	for _, rn := range rows {
		body.WriteString(fmt.Sprintf(`<row r="%d">`, rn.rowNum))
		for _, addr := range rn.addrs {
			c := sh.ws.cells[addr]
			if c == nil {
				continue
			}
			if c.hURL != "" {
				rid := fmt.Sprintf("rId%d", hid)
				hid++
				hlinks.WriteString(fmt.Sprintf(`<hyperlink ref="%s" r:id="%s"/>`, addr, rid))
				isExt := strings.HasPrefix(c.hURL, "http://") || strings.HasPrefix(c.hURL, "https://") || strings.HasPrefix(c.hURL, "mailto:")
				if isExt {
					rels.Relationship = append(rels.Relationship, ooxml.Relationship{
						ID:         rid,
						Type:       "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink",
						Target:     c.hURL,
						TargetMode: "External",
					})
				} else {
					rels.Relationship = append(rels.Relationship, ooxml.Relationship{
						ID:     rid,
						Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink",
						Target: c.hURL,
					})
				}
			}
			body.WriteString(emitCellXML(addr, c, sstIndex))
		}
		body.WriteString(`</row>`)
	}
	body.WriteString(`</sheetData>`)
	if hlinks.Len() > 0 {
		body.WriteString(`<hyperlinks>`)
		body.WriteString(hlinks.String())
		body.WriteString(`</hyperlinks>`)
	}
	body.WriteString(`</worksheet>`)
	if len(rels.Relationship) == 0 {
		return body.Bytes(), nil
	}
	return body.Bytes(), &rels
}

type writeRowGroup struct {
	rowNum int
	addrs  []string
}

func groupWriteRows(sortedAddrs []string) []writeRowGroup {
	var out []writeRowGroup
	var cur *writeRowGroup
	for _, addr := range sortedAddrs {
		_, r, _ := sml.CellRefToIndexes(addr)
		if cur == nil || cur.rowNum != r {
			if cur != nil {
				out = append(out, *cur)
			}
			cur = &writeRowGroup{rowNum: r, addrs: []string{addr}}
		} else {
			cur.addrs = append(cur.addrs, addr)
		}
	}
	if cur != nil {
		out = append(out, *cur)
	}
	return out
}

func emitCellXML(addr string, c *writeCell, sst map[string]int) string {
	var b strings.Builder
	b.WriteString(`<c r="`)
	b.WriteString(addr)
	b.WriteByte('"')
	if c.styleID >= 0 {
		fmt.Fprintf(&b, ` s="%d"`, c.styleID)
	}
	if c.formula != "" {
		b.WriteString(`><f>`)
		xmlEscape(&b, c.formula)
		b.WriteString(`</f>`)
		writeCellValue(&b, c, sst, true)
		b.WriteString(`</c>`)
		return b.String()
	}
	b.WriteByte('>')
	writeCellValue(&b, c, sst, false)
	b.WriteString(`</c>`)
	return b.String()
}

func writeCellValue(b *strings.Builder, c *writeCell, sst map[string]int, afterF bool) {
	switch v := c.val.(type) {
	case string:
		if ix, ok := sst[v]; ok {
			if !afterF {
				b.WriteString(` t="s"`)
			}
			b.WriteString(`<v>`)
			b.WriteString(strconv.Itoa(ix))
			b.WriteString(`</v>`)
			return
		}
		if !afterF {
			b.WriteString(` t="inlineStr"`)
		}
		b.WriteString(`<is><t>`)
		xmlEscape(b, v)
		b.WriteString(`</t></is>`)
	case bool:
		if !afterF {
			b.WriteString(` t="b"`)
		}
		if v {
			b.WriteString(`<v>1</v>`)
		} else {
			b.WriteString(`<v>0</v>`)
		}
	case time.Time:
		if !afterF {
			b.WriteString(` t="n"`)
		}
		b.WriteString(`<v>`)
		b.WriteString(floatStr(timeToExcelSerial(v)))
		b.WriteString(`</v>`)
	case float64:
		if !afterF {
			b.WriteString(` t="n"`)
		}
		b.WriteString(`<v>`)
		b.WriteString(floatStr(v))
		b.WriteString(`</v>`)
	case int:
		if !afterF {
			b.WriteString(` t="n"`)
		}
		b.WriteString(`<v>`)
		b.WriteString(floatStr(float64(v)))
		b.WriteString(`</v>`)
	default:
		if c.val == nil {
			return
		}
		if !afterF {
			b.WriteString(` t="inlineStr"`)
		}
		b.WriteString(`<is><t>`)
		xmlEscape(b, fmt.Sprint(c.val))
		b.WriteString(`</t></is>`)
	}
}

func xmlEscape(w io.Writer, s string) {
	_ = xml.EscapeText(w, []byte(s))
}

func xmlEscapeAttr(s string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}
