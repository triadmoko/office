# `docx` package guide (WordprocessingML)

The [`github.com/triadmoko/office/docx`](../docx/) package reads and writes a subset of **Office Open XML** for `.docx` files using **only the Go standard library** (no third-party modules in `go.mod`).

This document summarizes **what is implemented today** and provides **short examples per feature area**. A combined sample that exercises most of the API lives in [`cmd/office/main.go`](../cmd/office/main.go). Build and run:

```bash
go build -o office ./cmd/office
./office -o sample.docx
```

---

## Status overview

| Area | Read (`Open`) | Write (`NewDocument` + `Save` / `SaveFile`) | Notes |
|------|---------------|---------------------------------------------|--------|
| Paragraph & run (`w:p`, `w:r`, `w:t`) | Yes | Yes | Alignment, indent, spacing, style ID, numbering ref |
| Run: bold, italic, underline, strike, font, size, color, highlight, emphasis, sub/sup | Yes | Yes | |
| Page / column break in run | Yes | Yes | `AppendPageBreak`, `AppendColumnBreak` |
| Paragraph pagination properties | Yes | Yes | `keepNext`, `keepLines`, `widowControl`, `pageBreakBefore` |
| Section break (`w:sectPr` on paragraph) | Yes | Yes | `SetSectionBreak`, `SectionAt`, margins, orientation, A4/Letter |
| Page numbering per section | Yes | Yes | `SetPageNumberFormat`, `SetPageNumberStart` |
| Bullet & numbered lists | Yes | Yes | `Body.AppendList` |
| Tables | Yes | Yes | Borders, column widths, row height, header row, `cantSplit`, merge (read) |
| Styles (`styles.xml`) | Yes (partial) | Yes (default built-in) | `Styles().ByID`, `Style.Resolved()` |
| `numbering.xml` | Yes | Yes (for list builder) | `Document.Numbering()` |
| Header/footer with PAGE fields | Partial read | **`NewDocument` only** | `SetHeaderPageNumber`, `SetFooterPageNumber` + template placeholders |
| TOC (complex field) | — | Yes | `AppendTOCField` — evaluated in Microsoft Word after “Update field” |
| Images, comments, revision, macros | Bytes preserved on some paths | Not modeled | Round-trip for unparsed parts depends on scenario |

---

## 1. New document, save, and minimal write

```go
d := docx.NewDocument()
d.Body().Paragraphs()[0].AppendRun("Hello, world")
if err := d.SaveFile("out.docx"); err != nil {
	log.Fatal(err)
}
```

Single paragraph minimal package:

```go
var buf bytes.Buffer
if err := docx.WriteMinimal(&buf, "Hello, OOXML"); err != nil {
	log.Fatal(err)
}
```

---

## 2. Open a DOCX and plain text

```go
f, err := os.Open("input.docx")
if err != nil {
	log.Fatal(err)
}
defer f.Close()
st, _ := f.Stat()
d, err := docx.Open(f, st.Size())
if err != nil {
	log.Fatal(err)
}
text, err := d.PlainText()
if err != nil {
	log.Fatal(err)
}
fmt.Println(text)
```

---

## 3. Paragraph: alignment, indent, spacing

```go
p := d.Body().AppendParagraph()
p.SetAlignment(docx.AlignCenter)
p.SetIndent(docx.Indent{Left: 720, FirstLine: 720}) // twips (1/20 point)
p.SetSpacing(docx.Spacing{Before: 120, After: 240, Line: 360, LineRule: docx.LineRuleAuto})
p.AppendRun("Paragraph body")
```

---

## 4. Run: character formatting

```go
p := d.Body().AppendParagraph()
p.AppendRun("Plain, ")
r := p.AppendRun("bold red")
r.SetBold(true)
r.SetColor("C00000")
r.SetSize(24) // half-points
r.SetFont("Calibri")
r2 := p.AppendRun(" italic")
r2.SetItalic(true)
r3 := p.AppendRun(" underline")
r3.SetUnderline(true)
r4 := p.AppendRun(" strike")
r4.SetStrike(true)
r5 := p.AppendRun(" highlight")
r5.SetHighlight("yellow")
r6 := p.AppendRun("x")
r6.SetSubSuperscript(docx.VertAlignSuperscript)
```

---

## 5. Paragraph styles (Heading, etc.)

```go
h := d.Body().AppendParagraph()
h.SetStyleID("Heading1")
h.AppendRun("Chapter 1").SetBold(true)

if st := d.Styles(); st != nil {
	if s := st.ByID("Heading1"); s != nil {
		_ = s.Name()
		_ = s.Resolved() // effective format after BasedOn chain
	}
}
```

Word TOC entries usually rely on Heading styles and outline levels in OOXML; for TOC behavior see section 12.

---

## 6. Lists (bullet & numbered)

```go
bl := d.Body().AppendList(docx.ListBullet)
bl.AppendItem("First bullet")
bl.AppendItem("Second bullet")

nl := d.Body().AppendList(docx.ListNumbered)
nl.AppendItem("Step one")
nl.AppendItem("Step two")
```

---

## 7. Tables

```go
tbl := d.Body().AppendTable(3, 2)
tbl.SetGridColWidths([]int64{3000, 3000})
tbl.SetBorder(
	docx.BorderTop|docx.BorderLeft|docx.BorderBottom|docx.BorderRight|docx.BorderInsideH|docx.BorderInsideV,
	docx.BorderStyle{Color: "000000", Size: 4, Kind: docx.BorderSingle},
)
tbl.Rows()[0].SetRepeatAsHeaderRow(true)
tbl.Rows()[0].SetHeight(400, docx.RowHeightAtLeast)
tbl.Rows()[1].SetCantSplit(true)
tbl.Rows()[1].Cells()[0].Paragraphs()[0].AppendRun("Cell A")
```

---

## 8. Section, page size, margins, orientation

```go
if sec := d.SectionAt(0); sec != nil {
	sec.SetPageSize(docx.PageSizeA4)
	sec.SetOrientation(docx.Portrait)
	sec.SetMargins(docx.Margins{
		Top: 1800, Bottom: 1800, Left: 1440, Right: 1440,
		Header: 720, Footer: 720, Gutter: 0,
	})
}
```

Section break after a paragraph (e.g. next page):

```go
p := d.Body().AppendParagraph()
p.AppendRun("End of front matter")
p.SetSectionBreak(docx.SectionBreakConfig{
	PageKind: docx.PageSizeA4,
	Orient:   docx.Portrait,
	Break:    docx.SectionBreakNextPage,
})
```

---

## 9. Page number format in a section

```go
if sec := d.SectionAt(1); sec != nil {
	sec.SetPageNumberFormat(docx.PageNumberFormatDecimal)
	start := 1
	sec.SetPageNumberStart(&start)
}
```

---

## 10. Header & footer with page numbers (`NewDocument` only)

Text placeholders: `docx.FooterPlaceholderPage` (`{{PAGE}}`) and `docx.FooterPlaceholderNumPages` (`{{NUMPAGES}}`).

```go
d.SetFooterPageNumber(true)
d.SetFooterPageNumberTemplate("p. " + docx.FooterPlaceholderPage + " of " + docx.FooterPlaceholderNumPages)

d.SetHeaderPageNumber(true)
d.SetHeaderPageNumberTemplate("Report · p. " + docx.HeaderPlaceholderPage)
```

Calling these setters on a document opened with `Open` then saving may return an error (see `docx.ErrFooterPageNumberOpenDoc` and the header equivalent).

---

## 11. Pagination & layout hints

```go
p := d.Body().AppendParagraph()
p.AppendRun("Continue on next page")
p.AppendPageBreak()

p2 := d.Body().AppendParagraph()
p2.AppendRun("Column one")
p2.AppendColumnBreak()
p2.AppendRun("Column two")

p3 := d.Body().AppendParagraph()
p3.SetKeepNext(true)
p3.SetPageBreakBefore(true)
p3.SetKeepLines(true)
on := true
p3.SetWidowControl(&on)
```

Omit Word layout hints (`w:lastRenderedPageBreak`) when saving:

```go
d.SetStripLayoutHints(true)
```

---

## 12. Table of contents (TOC) field for Word

The library only emits the complex-field OOXML; **Microsoft Word** fills entries after the user updates fields. WPS / LibreOffice often show a placeholder only.

```go
p := d.Body().AppendParagraph()
p.AppendTOCField(docx.TOCFieldOptions{
	OutlineLevels:              "1-3",
	Hyperlinks:                 true,
	OmitPageNumbersInWebLayout: true,
	// UseAppliedOutlineLevels: true → outline from w:outlineLvl only, not from styles
})
```

---

## 13. Other useful APIs

- `d.Save(w)` — write to `io.Writer`.
- `d.MainPart()` — main document part path, e.g. `/word/document.xml`.
- `d.Package()` — low-level OPC package access (not a stability guarantee; prefer for debugging).
- `d.PartBytes()` — part snapshot when opened (useful for round-trip).

---

## Limitations (short)

- No high-level model for images, shapes, or charts.
- TOC and PAGE/NUMPAGES are **fields**; final rendering depends on the host app (Word is the most complete).
- Automatic header/footer generation is not supported for `Open` → edit → `Save`.

For ZIP/OPC architecture and safe package opening, see [architecture.md](architecture.md) and [security/zip-bomb-mitigation.md](security/zip-bomb-mitigation.md).
