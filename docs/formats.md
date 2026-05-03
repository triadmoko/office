# Format support

This document tracks **what works today** versus **planned** work. The project goal is faithful, incremental OOXML support while staying **stdlib-only**.

## Summary

| Format | Extension | Package | Read | Write | Notes |
|--------|-----------|---------|------|-------|--------|
| Word | `.docx` | `docx` | Paragraph/run/table/list/section/styles/numbering model + `PlainText()` | `NewDocument` + builder + `Save`/`SaveFile`; `WriteMinimal` | See [docx.md](docx.md) for the feature matrix and examples. Not a full Word feature set (images, comments, etc.). |
| Excel | `.xlsx` | `xlsx` | Open + **validate** main workbook part | **Not implemented** | Cell/sheet APIs to be added incrementally. |
| PowerPoint | `.pptx` | `pptx` | Open + **validate** main presentation part | **Not implemented** | Slides/masters to be added incrementally. |

## DOCX (WordprocessingML)

Up-to-date summary and per-feature examples: **[docx.md](docx.md)**.

### Supported today (high level)

- **`Open` / `NewDocument`**: main parts `document.xml`, `styles.xml`, `numbering.xml` as supported by the current parser.
- **Body**: paragraphs, runs, tables, bullet/numbered lists; paragraph properties (alignment, indent, spacing, style, numbering ref); pagination (`pageBreak`, `columnBreak`, `keepNext`, `keepLines`, `widowControl`, `pageBreakBefore`).
- **Sections**: page size (A4/Letter), orientation, margins, break kind (`nextPage`, `continuous`, …), page number format & start per section.
- **Tables**: borders, grid/column widths, row height, header row, `cantSplit`, merge (read), nested tables (read).
- **Styles & numbering**: read style registry and numbering definitions; write default built-ins for new documents.
- **Header/footer**: PAGE/NUMPAGES fields for documents from **`NewDocument`** only (`SetHeaderPageNumber` / `SetFooterPageNumber` + template).
- **Fields**: TOC complex field (`AppendTOCField`) — entries filled after refresh in Microsoft Word.
- **`PlainText()`**: aggregate text from `w:t` (quick summary).
- **`WriteMinimal`**: minimal single paragraph/run (smoke test / hello world).
- **`Save` / `SaveFile`**: re-serialize; round-trip retains unparsed parts per current implementation (`PartBytes`, etc.).

### Not supported or limited

- Images, shapes, charts, comments, revision markup, macros (partial preservation only, file-dependent).
- Full custom header/footer on `Open` → edit → `Save` (page-number setters are `NewDocument`-only).
- TOC/PAGE field evaluation beyond host-app behavior (not computed in Go).

## XLSX (SpreadsheetML)

- **`xlsx.Open`**: validates OPC + main workbook part presence.
- **`Workbook.Write`**: returns `xlsx.ErrNotImplemented`.

## PPTX (PresentationML)

- **`pptx.Open`**: validates OPC + main presentation part presence.
- **`Presentation.Write`**: returns `pptx.ErrNotImplemented`.

## Interoperability

Real files produced by Microsoft Office may include parts and extensions beyond this library’s current parsers. When adding features, prefer **small fixtures** and **round-trip tests**; validate with Office apps when feasible.
