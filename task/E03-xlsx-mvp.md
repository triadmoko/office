# EPIC OFFICE-E03 — XLSX MVP

> **Goal:** XLSX read (streaming + random) + write (builder + StreamWriter) untuk worksheet, cell, shared strings, styles, formula string.
> **Sprint:** 4–5
> **Total points:** 71

## Daftar Ticket

| ID | Title | Type | Sprint | Points | Priority |
|---|---|---|---|---|---|
| [OFFICE-201](#office-201) | Read — Workbook & sheet enumeration | Story | 4 | 5 | P1 |
| [OFFICE-202](#office-202) | Read — Shared strings table | Story | 4 | 5 | P1 |
| [OFFICE-203](#office-203) | Read — Streaming row iterator | Story | 4 | 8 | P1 |
| [OFFICE-204](#office-204) | Read — Random-access cell API | Story | 4 | 5 | P2 |
| [OFFICE-205](#office-205) | Read — Styles & number formats | Story | 4 | 5 | P2 |
| [OFFICE-206](#office-206) | Read — Merged cells, hidden, freeze | Story | 4 | 3 | P2 |
| [OFFICE-207](#office-207) | Write — Workbook builder | Story | 5 | 8 | P1 |
| [OFFICE-208](#office-208) | Write — StreamWriter for large datasets | Story | 5 | 8 | P1 |
| [OFFICE-209](#office-209) | Read+Write — Hyperlinks | Story | 5 | 3 | P2 |
| [OFFICE-210](#office-210) | Round-Trip — Open + Modify + Save | Story | 5 | 8 | P1 |

---

## OFFICE-201

### [Story] XLSX Read — Workbook & sheet enumeration

```
Type     : Story
Priority : P1
Points   : 5
Sprint   : 4
Epic     : OFFICE-E03
File     : internal/sml/workbook.go, xlsx/workbook.go
Depends  : OFFICE-005
```

#### Acceptance Criteria
- [ ] **AC1**: `xlsx.Open()` parse `xl/workbook.xml` → list of sheets.
- [ ] **AC2**: `Workbook.Sheets() []*Sheet` — name, sheetId, rId, state (visible/hidden/veryHidden).
- [ ] **AC3**: `Workbook.SheetByName(name string) *Sheet` (case-insensitive lookup ala Excel).
- [ ] **AC4**: `Workbook.DefinedNames() []*DefinedName` — name + formula + scope.
- [ ] **AC5**: Test fixture: workbook 3 sheets, 1 hidden, 1 named range.

---

## OFFICE-202

### [Story] XLSX Read — Shared strings table

```
Type     : Story
Priority : P1
Points   : 5
Sprint   : 4
Epic     : OFFICE-E03
File     : internal/sml/shared_strings.go
Depends  : OFFICE-201
```

#### Acceptance Criteria
- [ ] **AC1**: Lazy-load `xl/sharedStrings.xml` saat first cell access.
- [ ] **AC2**: API: `Workbook.SharedString(idx int) string`.
- [ ] **AC3**: Handle inline rich text → join runs jadi plain string (rich text Tier 2).
- [ ] **AC4**: `xml:space="preserve"` dipertahankan.
- [ ] **AC5**: Streaming parse (per-`<si>`, tidak load full XML jika file besar).
- [ ] **AC6**: Test: 100k unique strings → memory < 100 MB, time < 1 s.

---

## OFFICE-203

### [Story] XLSX Read — Streaming row iterator

```
Type     : Story
Priority : P1
Points   : 8
Sprint   : 4
Epic     : OFFICE-E03
File     : xlsx/sheet.go
Depends  : OFFICE-201, OFFICE-202
```

#### User Story
> Sebagai data engineer, saya ingin iterate baris XLSX besar (1M rows) tanpa memuat seluruh sheet ke memori, agar bisa di-stream ke pipeline ETL.

#### Acceptance Criteria
- [ ] **AC1**: API:
  ```go
  rows := sheet.Rows()
  for rows.Next() {
      row := rows.Row()
      for _, c := range row.Cells() {
          fmt.Println(c.Address(), c.Value())
      }
  }
  if err := rows.Err(); err != nil { ... }
  ```
- [ ] **AC2**: `Cell.Value() any` mengembalikan tipe sesuai cell:
  - `string` untuk shared/inline string
  - `float64` untuk number
  - `bool` untuk boolean
  - `time.Time` jika number format adalah date (auto-detect: format index 14–22, 45–47, atau format string mengandung `y/m/d/h/s`)
  - `error` untuk error type (#N/A, #DIV/0!, dll)
- [ ] **AC3**: `Cell.Type() CellType` enum eksplisit: `Number`, `String`, `Boolean`, `Date`, `Formula`, `Error`, `Empty`.
- [ ] **AC4**: `Cell.Address() string` (e.g., "A1"), `Cell.Row() int`, `Cell.Col() int`.
- [ ] **AC5**: `Cell.RawValue() string` raw dari `<v>` element.
- [ ] **AC6**: `Cell.Formula() string` jika cell punya `<f>`, else "".
- [ ] **AC7**: Performance: iterate 100k rows × 10 cells < 2 detik, memory < 50 MB peak.
- [ ] **AC8**: Iterator handle sparse rows (skip empty rows, expose row.Index()).
- [ ] **AC9**: Test fixture: 1k rows × 5 cols mixed types + dates + formulas.

---

## OFFICE-204

### [Story] XLSX Read — Random-access cell API

```
Type     : Story
Priority : P2
Points   : 5
Sprint   : 4
Epic     : OFFICE-E03
File     : xlsx/sheet.go
Depends  : OFFICE-203
```

#### Acceptance Criteria
- [ ] **AC1**: `sheet.Cell("A1") *Cell` — load lazy (parse sheet on first call, cache in memory).
- [ ] **AC2**: `sheet.CellAt(row, col int) *Cell` — 1-based indexing seperti Excel.
- [ ] **AC3**: Saat sheet di-load random-access, simpan map cells; subsequent calls O(1).
- [ ] **AC4**: Memory budget: load full sheet via random-access dilarang jika sheet > 100 MB (return error, sarankan streaming).
- [ ] **AC5**: Address parser: support absolute (`$A$1`), relative (`A1`), range (`A1:B10`).
- [ ] **AC6**: Test: random access mendapat hasil yang sama dengan streaming iterator.

---

## OFFICE-205

### [Story] XLSX Read — Styles & number formats

```
Type     : Story
Priority : P2
Points   : 5
Sprint   : 4
Epic     : OFFICE-E03
File     : internal/sml/styles.go, xlsx/style.go
Depends  : OFFICE-201
```

#### Acceptance Criteria
- [ ] **AC1**: Parse `xl/styles.xml`: `numFmts`, `fonts`, `fills`, `borders`, `cellXfs`.
- [ ] **AC2**: `Cell.Style() *Style` — referensi.
- [ ] **AC3**: `Style.NumberFormat() string` — built-in (1–47) atau custom string.
- [ ] **AC4**: `Style.Font()`, `Fill()`, `Border()`, `Alignment()`.
- [ ] **AC5**: Number format detection untuk date (digunakan AC2 di OFFICE-203).
- [ ] **AC6**: Test: parse styles dari Excel 365 fixture, verify 5 sample cells dapat style yang benar.

---

## OFFICE-206

### [Story] XLSX Read — Merged cells, hidden rows/cols, freeze pane

```
Type     : Story
Priority : P2
Points   : 3
Sprint   : 4
Epic     : OFFICE-E03
Depends  : OFFICE-203
```

#### Acceptance Criteria
- [ ] **AC1**: `Sheet.MergedRanges() []Range` — list of merged regions.
- [ ] **AC2**: `Sheet.HiddenRows() []int`, `HiddenCols() []int`.
- [ ] **AC3**: `Sheet.FreezePane() *FreezePane` (Row, Col atau nil).
- [ ] **AC4**: `Sheet.RowHeight(row int) float64` (default + custom).
- [ ] **AC5**: `Sheet.ColumnWidth(col int) float64`.
- [ ] **AC6**: Test fixture: workbook dengan 2 merged regions, freeze pane row 1, hidden col B.

---

## OFFICE-207

### [Story] XLSX Write — Workbook builder

```
Type     : Story
Priority : P1
Points   : 8
Sprint   : 5
Epic     : OFFICE-E03
File     : xlsx/workbook.go, xlsx/sheet.go
Depends  : OFFICE-003, OFFICE-201–206
```

#### User Story
> Sebagai backend dev, saya ingin men-generate XLSX dengan multiple sheet berisi data tabular + formatting + formula sederhana.

#### Acceptance Criteria
- [ ] **AC1**: `xlsx.NewWorkbook()`, `wb.AddSheet(name)`, `sheet.SetCell("A1", "Hello")`.
- [ ] **AC2**: `sheet.SetCell` polymorphic: accept `string`, `float64`, `int`, `bool`, `time.Time`, `*Cell` (with style).
- [ ] **AC3**: `sheet.SetFormula("A3", "=SUM(A1:A2)")` — formula disimpan sebagai string, tidak dievaluasi.
- [ ] **AC4**: Auto-generate sharedStrings untuk repeated strings (threshold: muncul ≥ 2x).
- [ ] **AC5**: Style API:
  ```go
  s := wb.NewStyle().Bold(true).NumberFormat("0.00").Background("FFFF00")
  sheet.SetCell("A1", 1234.56, s)
  ```
- [ ] **AC6**: `Workbook.Save()` menghasilkan XLSX yang:
  - Terbuka di Excel 365 tanpa repair
  - Terbuka di LibreOffice & Google Sheets
  - Lolos `unzip -t`
- [ ] **AC7**: Generated file: `xl/workbook.xml`, `xl/sharedStrings.xml` (jika ada string), `xl/styles.xml`, `xl/worksheets/sheet*.xml`, `_rels`, `[Content_Types].xml`, `docProps/*.xml`.
- [ ] **AC8**: Test: 5 fixtures (hello world, 100-row data, formula sum, formatted cells, dates).

#### Subtasks
- [ ] OFFICE-207.1 — Sheet XML serializer (3 pts)
- [ ] OFFICE-207.2 — Style registry + dedup (2 pts)
- [ ] OFFICE-207.3 — Shared strings auto-pool (2 pts)
- [ ] OFFICE-207.4 — Compatibility tests (1 pt)

---

## OFFICE-208

### [Story] XLSX Write — StreamWriter for large datasets

```
Type     : Story
Priority : P1
Points   : 8
Sprint   : 5
Epic     : OFFICE-E03
File     : xlsx/stream_writer.go (NEW)
Depends  : OFFICE-207
```

#### User Story
> Sebagai data dev, saya ingin export 1M baris ke XLSX tanpa OOM, dengan throughput tinggi.

#### Acceptance Criteria
- [ ] **AC1**: API:
  ```go
  sw := sheet.StreamWriter()
  sw.WriteRow(1, "Alice", 100.0, time.Now())
  sw.WriteRow(2, "Bob", 200.0, time.Now())
  sw.Flush()
  wb.Save(w)
  ```
- [ ] **AC2**: Memory: stream 1M rows × 10 cols → peak < 100 MB.
- [ ] **AC3**: Throughput: ≥ 100k rows/detik.
- [ ] **AC4**: Sheet yang ditulis via StreamWriter: cell tidak bisa di-edit lagi via random access (locked-write mode).
- [ ] **AC5**: Output buffered ke temp file lalu di-include ke ZIP saat Save (atau langsung stream ke ZIP entry).
- [ ] **AC6**: Auto shared-strings disabled di stream mode (gunakan inline string `t="inlineStr"`) — atau implement disk-backed pool.
- [ ] **AC7**: Benchmark dengan `go test -bench=BenchmarkStreamWriter1M`.

---

## OFFICE-209

### [Story] XLSX Read+Write — Hyperlinks

```
Type     : Story
Priority : P2
Points   : 3
Sprint   : 5
Epic     : OFFICE-E03
Depends  : OFFICE-203, OFFICE-207
```

#### Acceptance Criteria
- [ ] **AC1**: `Cell.Hyperlink() *Hyperlink` (read).
- [ ] **AC2**: `Cell.SetHyperlink(url, display string)` (write).
- [ ] **AC3**: Internal hyperlink (sheet reference) didukung: `SetHyperlink("Sheet2!A1", "Click")`.
- [ ] **AC4**: Generated file: relationships entry untuk external URL, plus `<hyperlinks>` di sheet XML.
- [ ] **AC5**: Test: external URL + sheet-internal + email mailto.

---

## OFFICE-210

### [Story] XLSX Round-Trip — Open + Modify + Save

```
Type     : Story
Priority : P1
Points   : 8
Sprint   : 5
Epic     : OFFICE-E03
Depends  : OFFICE-201–207
```

#### Acceptance Criteria
- [ ] **AC1**: `xlsx.Open()` simpan raw bytes untuk part tidak dipahami: `calcChain.xml`, `theme1.xml`, custom xml, vbaProject.
- [ ] **AC2**: Round-trip simple: open → save tanpa edit → file valid (semantic equality, bukan byte-identical).
- [ ] **AC3**: Modify single cell value → only `sheet1.xml` regenerated; theme, calcChain, vbaProject preserved verbatim.
- [ ] **AC4**: Test corpus: 3 xlsx kompleks dari Excel 365 (chart, pivot via preserve, formula).

---

**Navigasi:** [← E02 DOCX](./E02-docx-mvp.md) | [Index](./README.md) | [E04 PPTX →](./E04-pptx-mvp.md)
