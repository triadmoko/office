# Format support

This document tracks **what works today** versus **planned** work. The project goal is faithful, incremental OOXML support while staying **stdlib-only**.

## Summary

| Format | Extension | Package | Read | Write | Notes |
|--------|-----------|---------|------|-------|--------|
| Word | `.docx` | `docx` | Open package; **plain text** from `w:t` | **Minimal** one-paragraph document | Not a full Word feature set (styles, tables, images, etc.). |
| Excel | `.xlsx` | `xlsx` | Open + **validate** main workbook part | **Not implemented** | Cell/sheet APIs to be added incrementally. |
| PowerPoint | `.pptx` | `pptx` | Open + **validate** main presentation part | **Not implemented** | Slides/masters to be added incrementally. |

## DOCX (WordprocessingML)

### Supported (subset)

- Opening a package and locating the **main document** part (content type or root relationship).
- **`PlainText()`**: concatenates text in `w:t` elements (common namespace; empty namespace tolerated for `t` in some files).
- **`WriteMinimal`**: writes a tiny valid package (single run/paragraph).
- **`Document.Write`**: round-trip via re-serialization through `PlainText` + `WriteMinimal` (loses structure beyond plain text).

### Not supported yet (examples)

- Tables, images, headers/footers, styles, numbering, revision markup, fields, etc.

## XLSX (SpreadsheetML)

- **`xlsx.Open`**: validates OPC + main workbook part presence.
- **`Workbook.Write`**: returns `xlsx.ErrNotImplemented`.

## PPTX (PresentationML)

- **`pptx.Open`**: validates OPC + main presentation part presence.
- **`Presentation.Write`**: returns `pptx.ErrNotImplemented`.

## Interoperability

Real files produced by Microsoft Office may include parts and extensions beyond this library’s current parsers. When adding features, prefer **small fixtures** and **round-trip tests**; validate with Office apps when feasible.
