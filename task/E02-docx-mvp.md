# EPIC OFFICE-E02 — DOCX MVP

> **Goal:** DOCX read+write+round-trip dengan paragraph/run/table/style/numbering tier MVP.
> **Sprint:** 1–3
> **Total points:** 89

## Daftar Ticket

| ID | Title | Type | Sprint | Points | Priority |
|---|---|---|---|---|---|
| [OFFICE-101](#office-101) | Read — Domain Model paragraph/run/text | Story | 1 | 8 | P1 |
| [OFFICE-102](#office-102) | Read — Paragraph properties | Story | 1 | 5 | P1 |
| [OFFICE-103](#office-103) | Read — Tables | Story | 1 | 8 | P1 |
| [OFFICE-104](#office-104) | Read — Styles | Story | 1 | 5 | P2 |
| [OFFICE-105](#office-105) | Read — Numbering & Lists | Story | 1 | 5 | P2 |
| [OFFICE-106](#office-106) | Read — Section properties | Story | 1 | 3 | P2 |
| [OFFICE-107](#office-107) | Write — Builder for new document | Story | 2 | 13 | P1 |
| [OFFICE-108](#office-108) | Write — Numbering, table styles, page setup | Story | 2 | 8 | P2 |
| [OFFICE-109](#office-109) | Round-Trip — Open + Modify + Save fidelity | Story | 3 | 13 | P1 |

---

## OFFICE-101

### [Story] DOCX Read — Domain Model paragraph/run/text

```
Type     : Story
Priority : P1
Points   : 8
Sprint   : 1
Epic     : OFFICE-E02
File     : internal/wml/document.go, docx/paragraph.go, docx/run.go (NEW)
Depends  : OFFICE-005
```

#### User Story
> Sebagai backend developer, saya ingin membuka DOCX yang ada dan mengakses setiap paragraph + run dengan teks dan formatting karakter, agar bisa mengekstrak konten terstruktur untuk diolah.

#### Acceptance Criteria
- [ ] **AC1**: `Document.Body().Paragraphs()` mengembalikan `[]*Paragraph` dalam urutan dokumen.
- [ ] **AC2**: `Paragraph.Runs()` mengembalikan `[]*Run` dalam urutan; setiap Run mewakili 1 `<w:r>` element.
- [ ] **AC3**: `Run.Text()` mengembalikan teks dari semua `<w:t>` child (handle `xml:space="preserve"`).
- [ ] **AC4**: `Run.Bold()`, `Italic()`, `Underline()`, `Strike()`, `SubSuperscript()` mengembalikan boolean/enum dari `w:rPr`.
- [ ] **AC5**: `Run.FontSize() int` (half-points), `Run.Color() string` (hex `RRGGBB`), `Run.FontName() string`.
- [ ] **AC6**: Element yang tidak dikenal (Tier 2/3 features) tidak menyebabkan error — disimpan sebagai `Unknown []byte` raw untuk round-trip.
- [ ] **AC7**: Test fixture: minimal.docx + 1 docx kompleks dengan 5 paragraph berisi mix bold/italic/color.
- [ ] **AC8**: `PlainText()` legacy method tetap bekerja, sekarang implementasi via traversal model baru.
- [ ] **AC9**: Performance: parse 100-paragraph docx < 50 ms.

#### Edge Cases
- [ ] Paragraph kosong (`<w:p/>`)
- [ ] Run dengan multiple `<w:t>` (Word kadang split per karakter)
- [ ] Soft-hyphen (`<w:softHyphen/>`), tab (`<w:tab/>`), break (`<w:br/>`) — di-render sebagai `­`, `\t`, `\n` di Text.
- [ ] Smart-tag wrapper (Word legacy) — di-flatten transparently.

---

## OFFICE-102

### [Story] DOCX Read — Paragraph properties (alignment/indent/spacing)

```
Type     : Story
Priority : P1
Points   : 5
Sprint   : 1
Epic     : OFFICE-E02
File     : docx/paragraph.go
Depends  : OFFICE-101
```

#### Acceptance Criteria
- [ ] **AC1**: `Paragraph.Alignment() Alignment` enum: `Left`, `Right`, `Center`, `Justify`, `Distribute`, `Start`, `End`.
- [ ] **AC2**: `Paragraph.Indentation() Indent` struct: `Left`, `Right`, `FirstLine`, `Hanging` (twentieths of point).
- [ ] **AC3**: `Paragraph.Spacing() Spacing` struct: `Before`, `After`, `Line`, `LineRule` enum.
- [ ] **AC4**: `Paragraph.StyleID() string` — referensi ke style di styles.xml.
- [ ] **AC5**: `Paragraph.NumberingRef() *NumPr` (level + numId) atau nil.
- [ ] **AC6**: Default value yang masuk akal jika property tidak ada (alignment default `Left`).
- [ ] **AC7**: Test dengan fixture DOCX dari Word 365 dengan 4 alignment berbeda.

---

## OFFICE-103

### [Story] DOCX Read — Tables

```
Type     : Story
Priority : P1
Points   : 8
Sprint   : 1
Epic     : OFFICE-E02
File     : docx/table.go
Depends  : OFFICE-101
```

#### Acceptance Criteria
- [ ] **AC1**: `Body.Tables()` mengembalikan `[]*Table`.
- [ ] **AC2**: `Table.Rows()` → `[]*TableRow`; `TableRow.Cells()` → `[]*TableCell`.
- [ ] **AC3**: `TableCell.Paragraphs()` (cell bisa berisi multiple paragraph).
- [ ] **AC4**: Detect merge: `TableCell.GridSpan() int` (horizontal merge), `TableCell.VMerge() VMergeKind` (`None`, `Restart`, `Continue`).
- [ ] **AC5**: `Table.Width()` & `TableCell.Width()` dengan unit `WidthDxa`, `WidthPct`, `WidthAuto`.
- [ ] **AC6**: `TableCell.Borders()`, `TableCell.Shading()`.
- [ ] **AC7**: Nested table didukung (cell dapat berisi table lagi).
- [ ] **AC8**: Test fixture: 3x3 table simple, 3x3 table dengan merged cells (vertical & horizontal), nested table 1-level.

---

## OFFICE-104

### [Story] DOCX Read — Styles (styles.xml)

```
Type     : Story
Priority : P2
Points   : 5
Sprint   : 1
Epic     : OFFICE-E02
File     : internal/wml/styles.go (NEW), docx/style.go (NEW)
Depends  : OFFICE-101
```

#### Acceptance Criteria
- [ ] **AC1**: `Document.Styles() *Styles` — mengembalikan registry style.
- [ ] **AC2**: `Styles.ByID(id string) *Style` — lookup by `w:styleId`.
- [ ] **AC3**: Style memiliki: `Type` (paragraph/character/table/numbering), `Name`, `BasedOn`, `LinkedStyle`, `RPr`, `PPr`.
- [ ] **AC4**: `Style.Resolved() *ResolvedFormat` — flatten inheritance chain (BasedOn) jadi format final.
- [ ] **AC5**: Document defaults (`w:docDefaults`) di-resolve sebagai root style.
- [ ] **AC6**: Test: parse styles.xml dari fixture Word 365, verifikasi 5 built-in style (Normal, Heading1–4) ter-resolve dengan benar.

---

## OFFICE-105

### [Story] DOCX Read — Numbering & Lists

```
Type     : Story
Priority : P2
Points   : 5
Sprint   : 1
Epic     : OFFICE-E02
File     : internal/wml/numbering.go (NEW), docx/numbering.go (NEW)
Depends  : OFFICE-101, OFFICE-104
```

#### Acceptance Criteria
- [ ] **AC1**: `Document.Numbering() *Numbering`.
- [ ] **AC2**: `Numbering.ByNumID(id int) *NumDef`.
- [ ] **AC3**: `NumDef` punya `Levels []*NumLevel` (sampai 9 level untuk tier 2; MVP support level 0).
- [ ] **AC4**: `NumLevel` memiliki: `Format` (decimal, bullet, upperRoman, lowerLetter, dll), `Text` (template `%1.`), `RestartAt`, `StartAt`.
- [ ] **AC5**: Resolution chain: `<w:p>/<w:numPr>` → `numId` → `abstractNumId` → `lvl[i]`.
- [ ] **AC6**: Test fixture: docx dengan bulleted list (3 item) + numbered list (5 item).

---

## OFFICE-106

### [Story] DOCX Read — Section properties

```
Type     : Story
Priority : P2
Points   : 3
Sprint   : 1
Epic     : OFFICE-E02
File     : docx/section.go (NEW)
Depends  : OFFICE-101
```

#### Acceptance Criteria
- [ ] **AC1**: `Document.Sections() []*Section`.
- [ ] **AC2**: `Section.PageSize() PageSize` (W, H, Orientation `Portrait`/`Landscape`).
- [ ] **AC3**: `Section.Margins() Margins` (Top, Bottom, Left, Right, Header, Footer, Gutter).
- [ ] **AC4**: `Section.Columns() Columns` (Num, Sep, EqualWidth).
- [ ] **AC5**: Standard page sizes resolvable: Letter (12240×15840), A4 (11906×16838).
- [ ] **AC6**: Test: parse default Word section + custom landscape section.

---

## OFFICE-107

### [Story] DOCX Write — Builder for new document

```
Type     : Story
Priority : P1
Points   : 13
Sprint   : 2
Epic     : OFFICE-E02
File     : docx/document.go, docx/write.go
Depends  : OFFICE-003, OFFICE-004, OFFICE-006, OFFICE-101–106
```

#### User Story
> Sebagai backend developer, saya ingin membuat DOCX baru dari Go code dengan paragraph berisi text + formatting + tabel, agar bisa generate laporan otomatis.

#### Acceptance Criteria
- [ ] **AC1**: `docx.NewDocument()` mengembalikan `*Document` siap untuk Save.
- [ ] **AC2**: API builder lengkap (sesuai PRD §7.1):
  - `Body.AppendParagraph() *Paragraph`
  - `Paragraph.AppendRun(text string) *Run`
  - `Run.SetBold(true)`, `SetItalic`, `SetFont`, `SetSize`, `SetColor`
  - `Body.AppendTable(rows, cols int) *Table`
  - `Table.Cell(r, c).Paragraphs()...`
- [ ] **AC3**: `Document.Save(w io.Writer) error` menghasilkan file `.docx` yang:
  - Terbuka di MS Word 365 tanpa repair prompt
  - Terbuka di LibreOffice 7.6+ tanpa warning
  - Terbuka di Google Docs (upload + view) tanpa error
- [ ] **AC4**: Generated file punya `[Content_Types].xml`, `_rels/.rels`, `word/document.xml`, `word/styles.xml`, `docProps/core.xml`, `docProps/app.xml`.
- [ ] **AC5**: Default styles minimal: Normal, Heading1–3, ListParagraph.
- [ ] **AC6**: `core.xml` default value: Creator = "github.com/triadmoko/office", Created = now().
- [ ] **AC7**: `Document.SaveFile(path string) error` convenience.
- [ ] **AC8**: Round-trip test: build → save → open → readback → text match.
- [ ] **AC9**: 5 sample fixture: hello world, formatted runs, 3x3 table, bulleted list, multi-paragraph.

#### Definition of Done
- All 5 fixtures lulus matrix compatibility (Word/LibreOffice/Google Docs)
- Coverage `docx/` ≥ 70%
- Benchmark: write 1000-paragraph docx < 200 ms

#### Subtasks
- [ ] OFFICE-107.1 — `Document.MarshalToWML()` ke struct `internal/wml.Document` (3 pts)
- [ ] OFFICE-107.2 — Default styles.xml generator (3 pts)
- [ ] OFFICE-107.3 — Hooks ke `PackageWriter` + relationships chain (3 pts)
- [ ] OFFICE-107.4 — Test fixtures + compatibility verification (4 pts)

---

## OFFICE-108

### [Story] DOCX Write — Numbering, table styles, page setup

```
Type     : Story
Priority : P2
Points   : 8
Sprint   : 2
Epic     : OFFICE-E02
Depends  : OFFICE-107
```

#### Acceptance Criteria
- [ ] **AC1**: API builder list:
  ```go
  list := body.AppendList(ListBullet)  // atau ListNumbered
  list.AppendItem("First")
  list.AppendItem("Second")
  ```
- [ ] **AC2**: Generated `numbering.xml` valid + referenced di `[Content_Types]` & rels.
- [ ] **AC3**: API table border:
  ```go
  table.SetBorder(BorderAll, BorderStyle{Color: "000000", Size: 4, Kind: BorderSingle})
  ```
- [ ] **AC4**: API page setup:
  ```go
  doc.SectionAt(0).SetPageSize(PageSizeA4)
  doc.SectionAt(0).SetOrientation(Landscape)
  doc.SectionAt(0).SetMargins(Margins{Top: 1440, ...}) // twentieths
  ```
- [ ] **AC5**: Round-trip test untuk setiap fitur.
- [ ] **AC6**: Compatibility test pass.

---

## OFFICE-109

### [Story] DOCX Round-Trip — Open + Modify + Save fidelity

```
Type     : Story
Priority : P1
Points   : 13
Sprint   : 3
Epic     : OFFICE-E02
File     : docx/document.go (modifikasi)
Depends  : OFFICE-101–106, OFFICE-107
```

#### User Story
> Sebagai SaaS dev, saya ingin membuka DOCX template yang dikirim user, mengisi placeholder, dan menyimpan kembali — di mana semua bagian yang tidak saya sentuh (header/footer/comments/tracked changes/macros) tetap utuh.

#### Acceptance Criteria
- [ ] **AC1**: `Open()` menyimpan referensi raw bytes untuk part yang tidak di-parse (header, footer, footnotes, vbaProject, customXml, theme, fontTable, settings, webSettings).
- [ ] **AC2**: `Save()` setelah `Open()` (tanpa modifikasi) menghasilkan file dengan:
  - Semua part dipertahankan (sama list)
  - Part yang tidak dipahami → byte-identical
  - Part `document.xml`, `styles.xml`, `numbering.xml` → re-serialized (bisa berbeda byte tapi semantik sama)
- [ ] **AC3**: `Open() → modify paragraph 0 → Save()` menghasilkan file di mana hanya paragraph 0 yang berubah; rest tetap.
- [ ] **AC4**: Test corpus: 3 docx kompleks dari Word 365 (bisa berisi header, footer, comments, image) — verifikasi round-trip.
- [ ] **AC5**: `[Content_Types].xml` regenerated correctly (tidak kehilangan override custom).
- [ ] **AC6**: Relationships antar part tidak rusak setelah save.

#### Edge Cases
- [ ] Docx dengan macro (`vbaProject.bin`) — preserve verbatim, output tetap bisa dibuka di Word dengan macro intact.
- [ ] Docx dengan tracked changes — preserve verbatim.
- [ ] Docx dengan embedded image — preserve verbatim, relationship tidak putus.

---

**Navigasi:** [← E01 OPC](./E01-opc-foundation.md) | [Index](./README.md) | [E03 XLSX →](./E03-xlsx-mvp.md)
