# PRD: Pure-Go Office (DOCX / XLSX / PPTX) Library

**Repo:** `github.com/triadmoko/office`
**Tanggal:** 2026-05-04
**Status:** Draft v1
**Bahasa implementasi:** Go ≥ 1.23, **hanya** standard library (tanpa dependency eksternal).
**Lisensi:** MIT — Copyright (c) 2026 Triadmoko Denny Fatrosa.

---

## 1. Context & Problem Statement

Ekosistem Go saat ini memiliki beberapa pustaka untuk OOXML, namun masing-masing memiliki keterbatasan:

| Pustaka | Format | Catatan |
|---|---|---|
| `unidoc/unioffice` | DOCX/XLSX/PPTX | Lisensi komersial / AGPL, write fidelity baik tapi berat |
| `xuri/excelize` | XLSX | Hanya XLSX, sudah matang |
| `nguyenthenguyen/docx` | DOCX | Find-and-replace level, bukan true model |
| `johnfercher/maroto` | PDF, bukan DOCX | |

**Gap yang dijawab proyek ini:**
- **Pure-stdlib only** — tidak ada dependensi pihak ketiga (audit-friendly, lisensi bebas, supply-chain risk minimum).
- **Tiga format dalam satu tree** dengan fondasi OPC bersama (`internal/ooxml`).
- **Lisensi permisif** (MIT).
- **Idiomatic Go API** — bukan hasil port mekanis dari .NET OpenXML SDK.

**Problem yang dipecahkan untuk pengguna:**
1. Membuat dokumen kantoran dari layanan backend Go (laporan PDF→DOCX, ekspor data XLSX, slide otomatis).
2. Membaca/parsing dokumen yang di-upload user untuk ekstraksi data atau konversi.
3. Round-trip: buka dokumen, modifikasi terprogram, simpan kembali tanpa merusak konten lain.

**Outcome yang diinginkan:**
Sebuah pustaka Go yang dapat **menggantikan ~80% kebutuhan harian** untuk DOCX/XLSX/PPTX di backend service, dengan kualitas dan ergonomis API yang setara `excelize` untuk XLSX serta sebaik `python-docx` untuk DOCX.

---

## 2. Goals & Non-Goals

### Goals (apa yang akan dicapai)
- **Read & write** ketiga format di tier MVP yang berguna.
- **Round-trip safe** — buka file Office asli, ubah field tertentu, simpan; bagian yang tidak disentuh harus tetap valid.
- **Streaming-friendly** — tidak harus memuat seluruh dokumen ke memori untuk file besar.
- **Strict spec compliance** — file yang dihasilkan harus dapat dibuka oleh Microsoft Office, LibreOffice, Google Docs/Sheets/Slides, Apple iWork tanpa "repair" prompt.
- **Zero external dependencies** — `go.mod` tetap kosong (hanya stdlib).
- **API stabil** setelah v1.0 (semver).
- **Test coverage ≥ 80%** untuk paket inti.
- **Fuzz-tested** parser (anti-zip-bomb, anti-XXE, malformed XML).

### Non-Goals (eksplisit TIDAK akan dilakukan di scope ini)
- **Rendering visual** (DOCX → PNG/PDF). Pakai LibreOffice headless/external.
- **Macro / VBA execution**. Hanya preserve binary part jika round-trip.
- **DRM / IRM / encryption MS-OFFCRYPTO**. Mungkin di v2.
- **Format legacy biner** (`.doc`, `.xls`, `.ppt` Compound File Binary). Hanya OOXML (`.docx`, `.xlsx`, `.pptx`).
- **OpenDocument Format** (`.odt`, `.ods`, `.odp`). Berbeda spec.
- **Formula engine evaluator** untuk XLSX (kalkulasi). Hanya parse + serialize formula string.
- **Charts rendering**. Hanya struktur XML chart.
- **Spell-check, grammar, AI features**.

---

## 3. Target Users & Use Cases

### Persona 1: Backend Engineer
> *"Saya butuh men-generate laporan bulanan ratusan DOCX dengan tabel + chart dari data PostgreSQL."*

Membutuhkan: **Write-heavy API**, performant, deterministic output, template-friendly.

### Persona 2: Data Engineer
> *"User upload XLSX, saya parse setiap sheet jadi rows untuk diproses ETL."*

Membutuhkan: **Read-heavy API**, streaming row iterator, type-aware cell values, shared strings cache.

### Persona 3: SaaS Document Workflow
> *"User unggah DOCX template, sistem isi placeholder lalu kirim hasil. Format harus persis sama."*

Membutuhkan: **Round-trip fidelity**, content control / mail-merge field resolution.

### Persona 4: Compliance & Audit
> *"Saya butuh pustaka yang lisensinya bebas pakai komersial dan tidak menarik dependency yang berisiko."*

Membutuhkan: **Pure stdlib**, lisensi permisif, tanpa CGO.

---

## 4. Microsoft Office Feature Inventory (Referensi Lengkap)

> Ini adalah **inventory eksplisit fitur Microsoft Office** yang tersedia di Microsoft 365 (versi 2024-2026). Digunakan sebagai referensi untuk memetakan **In-Scope / Out-of-Scope** untuk pustaka kita.
> Spec resmi: **ECMA-376 5th Edition** + **MS-OE376 / MS-DOCX / MS-XLSX / MS-PPTX** open spec.

### 4.1 Microsoft Word (DOCX → WordprocessingML / WML)

#### 4.1.1 Konten & Struktur
| Fitur | Bagian Spec | Tier |
|---|---|---|
| Paragraph (`w:p`), Run (`w:r`), Text (`w:t`) | §17.3, §17.4 | **MVP** |
| Section (`w:sectPr`): margin, ukuran kertas, orientasi, kolom | §17.6 | **MVP** |
| Page break, line break, column break | §17.3.3 | **MVP** |
| Header & Footer (per section, first/odd/even) | §17.10 | **Tier 2** |
| Footnote & Endnote | §17.11 | **Tier 3** |
| Comments (`w:comment`) | §17.13 | **Tier 3** |
| Track Changes (revisions: `w:ins`, `w:del`, `w:moveFrom/To`) | §17.13 | **Tier 3** |
| Bookmarks, Hyperlinks, Cross-references | §17.13 | **Tier 2** |
| TOC (Table of Contents) sebagai field | §17.16 | **Tier 3** |
| Document Protection / Form Protection | §17.15 | **Out** |
| Custom XML data binding | §17.17 | **Out** |
| Content Controls (`w:sdt`) — Rich/Plain/Date/Combo/Picture | §17.5 | **Tier 2** |
| Mail Merge fields | §17.16 | **Tier 3** |

#### 4.1.2 Formatting Karakter (Run Properties `w:rPr`)
| Fitur | Tier |
|---|---|
| Bold, Italic, Underline (single/double/dotted/wave/...), Strikethrough, Double-strike | **MVP** |
| Font family (ascii/hAnsi/eastAsia/cs/complex script) | **MVP** |
| Font size (`w:sz`), Color (`w:color`), Highlight (`w:highlight`), Shading | **MVP** |
| Subscript / Superscript / Vertical alignment | **MVP** |
| All caps / Small caps / Hidden / Vanish | **Tier 2** |
| Character spacing, kerning, scaling, position | **Tier 2** |
| Text effects (glow, shadow, reflection, 3D) — DrawingML | **Tier 3** |
| Text border, emphasis mark | **Tier 3** |
| Language tag (lang) | **Tier 2** |

#### 4.1.3 Formatting Paragraf (`w:pPr`)
| Fitur | Tier |
|---|---|
| Alignment (left/right/center/justify/distribute) | **MVP** |
| Indentation (left/right/firstLine/hanging) | **MVP** |
| Spacing (before/after/line, lineRule auto/exact/atLeast) | **MVP** |
| Borders (top/bottom/left/right/between/bar) | **Tier 2** |
| Shading | **Tier 2** |
| Tabs (custom tab stops dengan leader) | **Tier 2** |
| Page break before, keep with next, keep lines together, widow/orphan | **Tier 2** |
| Outline level | **Tier 2** |
| Numbering reference (`w:numPr`) | **MVP** |
| Frame properties | **Out** |

#### 4.1.4 Tabel (`w:tbl`)
| Fitur | Tier |
|---|---|
| Tabel dasar: row, cell, grid | **MVP** |
| Cell width (auto/dxa/pct), table width | **MVP** |
| Borders (table/cell level), shading | **MVP** |
| Vertical merge (vMerge), horizontal merge (gridSpan) | **MVP** |
| Cell margin, alignment (vAlign) | **MVP** |
| Header row (tblHeader: repeat at top of page) | **Tier 2** |
| Table style (referensi ke styles.xml table style) | **Tier 2** |
| Conditional formatting (firstRow, lastRow, banding) | **Tier 2** |
| Text direction in cell, table positioning | **Tier 3** |
| Nested tables | **Tier 2** |

#### 4.1.5 List & Numbering (`numbering.xml`)
| Fitur | Tier |
|---|---|
| Bulleted list, numbered list (decimal/upper-roman/lower-letter/...) | **MVP** |
| Multi-level list (sampai 9 level) | **Tier 2** |
| Custom bullet character / picture bullet | **Tier 2** |
| Numbering restart, override, continue | **Tier 2** |

#### 4.1.6 Styles (`styles.xml`)
| Fitur | Tier |
|---|---|
| Paragraph style, character style, linked style | **MVP** |
| Style inheritance (`w:basedOn`) | **MVP** |
| Default style (`w:default`), document defaults | **MVP** |
| Table style | **Tier 2** |
| Numbering style | **Tier 2** |
| Latent styles | **Tier 3** |

#### 4.1.7 Drawing & Media (DrawingML — `dml`)
| Fitur | Tier |
|---|---|
| Inline image (PNG/JPEG/GIF) sebagai `w:drawing/wp:inline` | **Tier 2** |
| Anchored image (floating, with wrap) | **Tier 3** |
| Shapes (rect, ellipse, line, callout, autoShape) | **Tier 3** |
| SmartArt | **Out** |
| Chart (referensi ke `chart1.xml`) | **Tier 3** |
| Diagram | **Out** |
| Embedded OLE object | **Out** |
| Equation (OMML / Office MathML) | **Tier 3** |

#### 4.1.8 Fields & Automation
| Fitur | Tier |
|---|---|
| Simple field (`w:fldSimple`) | **Tier 2** |
| Complex field (`w:fldChar` begin/separate/end) | **Tier 2** |
| Built-in fields: PAGE, NUMPAGES, DATE, TIME, FILENAME, AUTHOR | **Tier 2** |
| TOC field, INDEX field | **Tier 3** |
| HYPERLINK field | **Tier 2** |
| MERGEFIELD | **Tier 3** |

#### 4.1.9 Lain-lain
| Fitur | Tier |
|---|---|
| `core.xml` (Dublin Core metadata: title/creator/subject/...) | **MVP** |
| `app.xml` (application properties) | **MVP** |
| `custom.xml` (custom document properties) | **Tier 2** |
| `theme1.xml` (color/font/effect scheme) | **Tier 2** |
| Web settings, font table | **Tier 2** |
| Glossary document (Quick Parts / Building Blocks) | **Out** |
| VBA macros (`vbaProject.bin`) | **Out** (preserve only) |
| Encryption / Password protection | **Out** |

---

### 4.2 Microsoft Excel (XLSX → SpreadsheetML / SML)

#### 4.2.1 Workbook & Worksheet
| Fitur | Tier |
|---|---|
| Workbook (`workbook.xml`): sheets list, defined names, calculation properties | **MVP** |
| Worksheet (`sheet1.xml`): rows, cells, dimensions | **MVP** |
| Multiple sheets, sheet visibility (visible/hidden/veryHidden) | **MVP** |
| Sheet tab color, sheet ordering | **Tier 2** |
| Frozen panes, split panes | **Tier 2** |
| Print area, print titles, page setup | **Tier 2** |
| Sheet protection (read-only) | **Tier 2** |

#### 4.2.2 Cell & Data
| Fitur | Tier |
|---|---|
| Cell types: number (`n`), string inline (`inlineStr`), shared string (`s`), boolean (`b`), error (`e`), formula | **MVP** |
| Shared Strings Table (`sharedStrings.xml`) | **MVP** |
| Date/time as serial number + number format | **MVP** |
| Rich text run di dalam shared string | **Tier 2** |
| Formulas (string saja, tidak dievaluasi): A1 & R1C1 reference, range, function names | **MVP** |
| Array formulas, dynamic arrays (CSE) | **Tier 2** |
| Shared formulas (`t="shared"`) | **Tier 2** |
| Cell comments / Notes (`comments1.xml`) | **Tier 2** |
| Threaded comments (modern) | **Tier 3** |
| Hyperlinks | **MVP** |
| Data validation (list/whole/decimal/date/time/textLength/custom) | **Tier 2** |
| Linked data types (Stocks, Geography) | **Out** |

#### 4.2.3 Format & Style (`styles.xml`)
| Fitur | Tier |
|---|---|
| Number formats (built-in + custom) | **MVP** |
| Fonts (name/size/color/bold/italic/underline/strike) | **MVP** |
| Fills (solid, pattern, gradient) | **MVP** |
| Borders (all 6 sides + diagonal) | **MVP** |
| Cell alignment (horizontal/vertical/wrapText/indent/rotation/shrinkToFit) | **MVP** |
| Cell styles (`cellXfs`, `cellStyleXfs`, `cellStyles`) | **MVP** |
| Named cell styles (Good/Bad/Neutral, dll) | **Tier 2** |
| Differential formatting (`dxf`) untuk conditional formatting | **Tier 2** |
| Theme colors & tints | **Tier 2** |

#### 4.2.4 Layout
| Fitur | Tier |
|---|---|
| Merged cells | **MVP** |
| Row height, column width, custom width | **MVP** |
| Hidden rows/columns | **MVP** |
| Outline (group/ungroup), level | **Tier 2** |
| AutoFilter | **Tier 2** |
| Sort state | **Tier 3** |

#### 4.2.5 Tables (ListObjects)
| Fitur | Tier |
|---|---|
| Excel Table (`table1.xml`): range, name, columns, totalsRow | **Tier 2** |
| Table style reference | **Tier 2** |
| Calculated column formula | **Tier 3** |

#### 4.2.6 Conditional Formatting
| Fitur | Tier |
|---|---|
| Cell value rules (greaterThan, between, equal, ...) | **Tier 2** |
| Color scale (2-color, 3-color) | **Tier 2** |
| Data bar | **Tier 2** |
| Icon set | **Tier 3** |
| Top/bottom rules, duplicates, formula-based | **Tier 2** |

#### 4.2.7 Charts & Visuals
| Fitur | Tier |
|---|---|
| Chart parts (`chart1.xml` di drawings) — column, bar, line, pie, doughnut, area, scatter, bubble | **Tier 3** |
| Combo charts | **Tier 3** |
| Sparklines | **Tier 3** |
| Maps, treemap, sunburst, waterfall, funnel, histogram, box-and-whisker | **Out** |
| 3D charts | **Out** |

#### 4.2.8 Pivot & Advanced
| Fitur | Tier |
|---|---|
| Pivot Table (`pivotTable1.xml` + `pivotCacheDefinition`) | **Out** |
| Pivot Chart | **Out** |
| Slicers, Timeline | **Out** |
| Power Query (`queryTable`, M code) | **Out** |
| Data Model / Power Pivot (analysis services) | **Out** |
| External connections | **Out** |

#### 4.2.9 Drawing & Media
| Fitur | Tier |
|---|---|
| Images embedded di sheet (`drawing1.xml` + media) | **Tier 3** |
| Shapes | **Tier 3** |
| SmartArt | **Out** |
| 3D models, Icons, Stock images | **Out** |

#### 4.2.10 Lain-lain
| Fitur | Tier |
|---|---|
| Defined Names (workbook & sheet scope) | **MVP** |
| Calculation chain (`calcChain.xml`) | **MVP** (preserve) |
| `core.xml` / `app.xml` metadata | **MVP** |
| Workbook protection | **Tier 2** |
| Workbook views, sheet views (zoom, gridlines, ruler) | **Tier 2** |
| External references (`externalLink`) | **Out** |
| Custom XML | **Tier 2** |
| VBA macros (`xlsm`) | **Out** (preserve only) |
| Encryption | **Out** |

---

### 4.3 Microsoft PowerPoint (PPTX → PresentationML / PML)

#### 4.3.1 Struktur Presentasi
| Fitur | Tier |
|---|---|
| Presentation (`presentation.xml`) — slide list, slide size, default text style | **MVP** |
| Slide (`slide1.xml`) | **MVP** |
| Slide Layout (`slideLayout*.xml`) | **MVP** |
| Slide Master (`slideMaster*.xml`) | **MVP** |
| Notes Slide (`notesSlide*.xml`) | **Tier 2** |
| Notes Master, Handout Master | **Tier 3** |
| Sections (grouping slides) | **Tier 2** |
| Slide size (standard 4:3, widescreen 16:9, custom) | **MVP** |

#### 4.3.2 Theme & Master
| Fitur | Tier |
|---|---|
| Theme (`theme1.xml`): color scheme, font scheme, format scheme | **MVP** |
| Background fill (solid/gradient/picture) | **MVP** |
| Master placeholders (title, body, date, footer, slideNumber) | **MVP** |
| Layout placeholders | **MVP** |
| Color scheme variants | **Tier 2** |

#### 4.3.3 Shapes (`p:sp` / `p:pic` / `p:graphicFrame`)
| Fitur | Tier |
|---|---|
| Text box (placeholder & non-placeholder) | **MVP** |
| AutoShapes (rectangle, ellipse, arrow, callout, ...) | **Tier 2** |
| Lines, connectors | **Tier 2** |
| Freeform / custom geometry | **Tier 3** |
| Shape grouping | **Tier 2** |
| Shape style (presetGeometry, fill, line, effect) | **MVP** |
| Shape transform (offset, extent, rotation, flip) | **MVP** |

#### 4.3.4 Text in Shape (`a:txBody`)
| Fitur | Tier |
|---|---|
| Paragraphs & runs (DrawingML text) | **MVP** |
| Run formatting (font, size, color, bold, italic, underline) | **MVP** |
| Paragraph formatting (alignment, indent, spacing, level) | **MVP** |
| Bullets (character, autoNum, picture, none) | **MVP** |
| Hyperlinks (action, click) | **Tier 2** |
| Field (slide number, date, footer) | **Tier 2** |
| Body autofit (none/normal/shape) | **Tier 2** |

#### 4.3.5 Picture & Media
| Fitur | Tier |
|---|---|
| Picture insertion (PNG/JPEG/GIF/SVG) | **MVP** |
| Picture cropping, transparency | **Tier 2** |
| Video (mp4, wmv) embedded/linked | **Tier 3** |
| Audio embedded/linked | **Tier 3** |
| Screen recording | **Out** |
| 3D Model (`gltf`/`glb`) | **Out** |

#### 4.3.6 Tabel & Chart
| Fitur | Tier |
|---|---|
| Table (`a:tbl`) — rows/cells/borders/fill | **Tier 2** |
| Chart (`chart1.xml` di slide drawings) | **Tier 3** |
| SmartArt | **Out** |

#### 4.3.7 Animation & Transition
| Fitur | Tier |
|---|---|
| Slide transition (fade, push, wipe, ...) | **Tier 2** |
| Animation timing tree (`p:timing`) — entrance, exit, emphasis | **Tier 3** |
| Motion paths | **Out** |
| Morph transition | **Out** |
| Trigger animations (on click) | **Tier 3** |

#### 4.3.8 Lain-lain
| Fitur | Tier |
|---|---|
| Speaker Notes (text di notes slide) | **Tier 2** |
| Comments | **Tier 3** |
| Hyperlinks (slide-to-slide, URL, file) | **Tier 2** |
| Slide show settings (loop, kiosk, custom show) | **Tier 3** |
| Custom shows | **Tier 3** |
| Coauthoring metadata | **Out** |
| Macros, encryption | **Out** |

---

## 5. Tier Roadmap (Mapping Fitur → Versi)

| Tier | Versi | Target | Cakupan |
|---|---|---|---|
| **MVP** | v0.1 – v0.5 | DOCX core, XLSX core, PPTX core | Read+write fitur paling umum (lihat tabel) |
| **Tier 2** | v0.6 – v0.9 | Production-ready 80% use cases | Header/footer, table style, chart placeholder, conditional format, etc. |
| **Tier 3** | v1.0 – v1.5 | Power-user features | Comments, track changes, footnotes, animation, pivot table read |
| **Out** | — | Tidak akan didukung | Macros, encryption, SmartArt, ODF, formula eval, rendering |

---

## 6. Architecture & Package Layout

### 6.1 Struktur paket yang diusulkan

```
github.com/triadmoko/office/
├── go.mod                          # tetap kosong (stdlib only)
├── docx/                           # Public API: WordprocessingML
│   ├── doc.go
│   ├── document.go                 # Document struct, Open, Save
│   ├── paragraph.go                # Paragraph, Run, Text
│   ├── table.go                    # Table, Row, Cell
│   ├── style.go                    # Style references
│   ├── numbering.go                # Lists
│   ├── section.go                  # Section, page setup
│   ├── header_footer.go
│   ├── image.go
│   └── *_test.go
├── xlsx/                           # Public API: SpreadsheetML
│   ├── doc.go
│   ├── workbook.go                 # Workbook, Open, Save
│   ├── sheet.go                    # Sheet iterator (read), builder (write)
│   ├── cell.go                     # Cell value, type, format
│   ├── formula.go                  # Formula string handling
│   ├── style.go                    # Number format, font, fill, border
│   ├── shared_strings.go
│   ├── defined_name.go
│   └── *_test.go
├── pptx/                           # Public API: PresentationML
│   ├── doc.go
│   ├── presentation.go
│   ├── slide.go
│   ├── layout.go
│   ├── master.go
│   ├── shape.go
│   ├── text_body.go
│   └── *_test.go
├── internal/
│   ├── ooxml/                      # OPC layer (sudah ada, perlu write support)
│   │   ├── package.go              # zip.Reader (existing)
│   │   ├── package_writer.go       # NEW: zip.Writer + serialize CT/Rels
│   │   ├── content_types.go
│   │   ├── rels.go
│   │   ├── path.go
│   │   ├── namespaces.go
│   │   └── errors.go
│   ├── xmlwriter/                  # NEW: XML serializer dengan namespace handling
│   │   └── writer.go
│   ├── wml/                        # NEW: WordprocessingML schema types
│   │   ├── document.go
│   │   ├── styles.go
│   │   ├── numbering.go
│   │   └── shared.go
│   ├── sml/                        # NEW: SpreadsheetML schema types
│   │   ├── workbook.go
│   │   ├── sheet.go
│   │   └── styles.go
│   ├── pml/                        # NEW: PresentationML schema types
│   │   ├── presentation.go
│   │   ├── slide.go
│   │   └── theme.go
│   ├── dml/                        # NEW: DrawingML (shared by docx/xlsx/pptx)
│   │   ├── color.go
│   │   ├── theme.go
│   │   ├── shape.go
│   │   └── text.go
│   ├── opcprops/                   # NEW: core.xml / app.xml / custom.xml
│   │   └── props.go
│   └── tokenpool/                  # OPTIONAL: pool xml.Token & buffer reuse
│       └── pool.go
├── cmd/
│   └── office/                     # CLI tool (existing, akan diperluas)
│       └── main.go
└── docs/
    ├── ARCHITECTURE.md
    ├── DOCX.md
    ├── XLSX.md
    └── PPTX.md
```

**Rasional pemisahan paket:**
- `internal/wml|sml|pml|dml` — schema layer; bisa di-share antar format (DrawingML dipakai semuanya).
- Public packages `docx|xlsx|pptx` — domain-friendly API, tidak mengekspos struct XML mentah.
- `internal/ooxml` — OPC layer (sudah solid untuk read, perlu add write).
- Tidak ada paket eksternal (`go.mod` tetap kosong).

### 6.2 Layered Architecture

```
┌───────────────────────────────────────────────────┐
│ Layer 4: Public API (docx/xlsx/pptx)              │  ← Domain types: Paragraph, Cell, Slide
├───────────────────────────────────────────────────┤
│ Layer 3: Schema Mapping (internal/wml/sml/pml)    │  ← Struct-tagged untuk encoding/xml
├───────────────────────────────────────────────────┤
│ Layer 2: OPC Package (internal/ooxml)             │  ← Parts, Rels, Content Types
├───────────────────────────────────────────────────┤
│ Layer 1: ZIP + XML (archive/zip, encoding/xml)    │  ← stdlib only
└───────────────────────────────────────────────────┘
```

### 6.3 Reading Strategy

- **Streaming-first** untuk worksheet besar: gunakan `xml.Decoder.Token()` daripada `xml.Unmarshal` penuh. API: `sheet.Rows()` mengembalikan iterator.
- **Eager untuk metadata**: styles/sharedStrings/relationships dimuat penuh saat Open (relatif kecil).
- **Lazy untuk part yang berat**: media, drawing, chart hanya dimuat saat diakses.

### 6.4 Writing Strategy

- **Builder pattern** untuk objek baru (`docx.NewDocument()`, `xlsx.NewWorkbook()`).
- **Round-trip preserve**: jika dibuka dari file existing, bagian yang tidak dimodifikasi disimpan **byte-identical** (kecuali XML yang sengaja di-touch).
- **Deterministic output** — urutan ZIP entries stabil (untuk hashable/reproducible builds).
- **Pre-flight validation** opsional: `doc.Validate() error` mengecek schema constraints sebelum save.

### 6.5 Round-Trip Fidelity Strategy

Tiga mode operasi:
1. **Create from scratch** — full control, output kita sepenuhnya.
2. **Open + read-only** — parse selektif, sisanya diabaikan.
3. **Open + modify + save** — parts yang dipahami → diserialisasi ulang. Parts yang tidak dipahami (mis. `vbaProject.bin`, custom XML, charts) → **disalin verbatim** dari ZIP entry asli ke ZIP output.

Implementasi: `Package` menyimpan referensi `[]byte` mentah untuk part "unknown", dan `WriteTo(w)` akan menulisnya kembali tanpa parsing.

---

## 7. Public API — Sketsa per Format

### 7.1 DOCX

```go
package docx

// Open / Create
func Open(r io.ReaderAt, size int64) (*Document, error)
func OpenFile(path string) (*Document, error)
func NewDocument() *Document

// Save
func (d *Document) Save(w io.Writer) error
func (d *Document) SaveFile(path string) error

// Body access
func (d *Document) Body() *Body
func (b *Body) Paragraphs() []*Paragraph
func (b *Body) AppendParagraph() *Paragraph
func (b *Body) AppendTable(rows, cols int) *Table

// Paragraph
func (p *Paragraph) Runs() []*Run
func (p *Paragraph) AppendRun(text string) *Run
func (p *Paragraph) SetAlignment(a Alignment)
func (p *Paragraph) SetStyle(styleID string)

// Run
func (r *Run) Text() string
func (r *Run) SetText(s string)
func (r *Run) SetBold(b bool)
func (r *Run) SetItalic(b bool)
func (r *Run) SetFont(name string)
func (r *Run) SetSize(halfPoints int)
func (r *Run) SetColor(hex string)

// Table
func (t *Table) Row(i int) *TableRow
func (t *Table) Cell(row, col int) *TableCell
func (c *TableCell) Paragraphs() []*Paragraph
```

### 7.2 XLSX

```go
package xlsx

func Open(r io.ReaderAt, size int64) (*Workbook, error)
func NewWorkbook() *Workbook

func (wb *Workbook) Save(w io.Writer) error
func (wb *Workbook) Sheets() []*Sheet
func (wb *Workbook) AddSheet(name string) *Sheet
func (wb *Workbook) SheetByName(name string) *Sheet

// Streaming read (memory-efficient)
func (s *Sheet) Rows() *RowIterator
func (it *RowIterator) Next() bool
func (it *RowIterator) Row() *Row
func (r *Row) Cell(col int) *Cell

// Random access (loads sheet)
func (s *Sheet) Cell(addr string) *Cell  // "A1" notation
func (s *Sheet) SetCell(addr string, v any)

// Cell value
func (c *Cell) Value() any        // returns float64, string, bool, time.Time
func (c *Cell) String() string
func (c *Cell) Float() (float64, error)
func (c *Cell) Time() (time.Time, error)
func (c *Cell) Formula() string
func (c *Cell) SetValue(v any)
func (c *Cell) SetFormula(f string)
func (c *Cell) SetNumberFormat(fmt string)
func (c *Cell) SetStyle(s *Style)

// Streaming write (memory-efficient untuk file besar)
func (s *Sheet) StreamWriter() *StreamWriter
func (sw *StreamWriter) WriteRow(values ...any) error
func (sw *StreamWriter) Flush() error
```

### 7.3 PPTX

```go
package pptx

func Open(r io.ReaderAt, size int64) (*Presentation, error)
func NewPresentation() *Presentation

func (p *Presentation) Save(w io.Writer) error
func (p *Presentation) Slides() []*Slide
func (p *Presentation) AddSlide(layoutName string) *Slide

func (s *Slide) Shapes() []Shape
func (s *Slide) AddTextBox(x, y, w, h Emu) *TextBox
func (s *Slide) AddPicture(path string, x, y, w, h Emu) (*Picture, error)
func (s *Slide) Notes() string
func (s *Slide) SetNotes(text string)

// Text body (shared dengan docx via internal/dml)
func (tb *TextBox) Paragraphs() []*Paragraph
func (tb *TextBox) AppendParagraph() *Paragraph

// Units helper
type Emu int64                         // English Metric Unit
func Inches(v float64) Emu
func Points(v float64) Emu
func Cm(v float64) Emu
```

---

## 8. Quality, Security, & Performance Requirements

### 8.1 Security
- **Anti zip-bomb**: limit total uncompressed size & per-entry size. Default cap: 1 GiB total, 256 MiB per part. Configurable via `OpenOptions{MaxBytes, MaxParts}`.
- **Anti XXE**: `xml.Decoder` di Go tidak resolve external entity by default — pastikan tidak ada custom entity expansion.
- **Anti path traversal**: ZIP entry name yang mengandung `..` atau absolute path harus ditolak. (Catatan: bug di `internal/ooxml.ResolveTarget` perlu diperbaiki — `..` sekarang silently drop, harus error pada traversal di luar package.)
- **Bounded memory** untuk parser: streaming decoder untuk part > 10 MiB.
- **Fuzz testing** (`testing/fuzz` di Go 1.18+): fuzz parser content types, rels, dan worksheet tokenizer.

### 8.2 Performance
- Parse 100k-row XLSX < 2 detik dengan streaming iterator (target sebanding `excelize` streaming).
- Generate 10k-row XLSX < 1 detik via `StreamWriter`.
- Open 100-page DOCX < 200 ms.
- Memory: streaming mode tidak boleh allocate > O(rows_buffered).

### 8.3 Compatibility Testing
Setiap PR yang menyentuh writer harus lulus matriks:
- MS Office 365 (Windows) — buka tanpa repair prompt
- MS Office 365 (macOS)
- LibreOffice ≥ 7.6
- Google Docs / Sheets / Slides (upload + download)
- Apple Pages / Numbers / Keynote
- WPS Office (penting di pasar Asia)

CI artifact: corpus file `testdata/golden/*.{docx,xlsx,pptx}` + script smoke test via `unzip -t` & schema-validate via `xmllint --schema` (opsional, requires libxml2 di runner).

### 8.4 Testing Strategy
- **Unit test** per file (.go ↔ _test.go).
- **Golden file test** — bandingkan output dengan fixture deterministik.
- **Round-trip test** — `Open → Save → Open` harus stabil (semantic equality).
- **Cross-tool test** — produce file, parse via `xlsx2csv`/`docx2txt` di CI sebagai independent verifier.
- **Fuzz test** untuk parser.
- **Coverage gate**: ≥ 80% di `internal/ooxml`, ≥ 70% di paket format.

---

## 9. Migration & Versioning

- **Versi pra-1.0** (`v0.x`): API boleh breaking-change antar minor. Setiap breaking change didokumentasikan di CHANGELOG.
- **Versi 1.0**: API stabil, semver strict. Breaking change → v2 module path.
- **Deprecation policy**: tandai dengan `// Deprecated:` doc-comment, hapus minimal 2 minor versi setelah deprecation.

---

## 10. Risiko & Mitigasi

| Risiko | Dampak | Mitigasi |
|---|---|---|
| Spec ECMA-376 sangat luas | Scope creep, never-ending | Tier roadmap ketat; tolak fitur Tier 3+ sampai MVP solid |
| Microsoft Office punya banyak undocumented quirks | File yang valid spec tetap ditolak Office | CI matrix testing di Office asli, golden files dari Office |
| Pure stdlib XML parser kurang performant untuk file besar | XLSX 1M row lambat | Streaming decoder + token pooling; benchmark di awal |
| Round-trip fidelity sulit | Modifikasi rusakkan format | Strategy: copy-verbatim untuk part yang tidak dipahami |
| Bug di `ResolveTarget` (sudah teridentifikasi) | Path traversal salah resolve, potential security | **Fix di Sprint 1** sebagai prerequisite |
| Lisensi font / theme MS yang ter-embed | Pelanggaran TOS jika kita ship sample MS theme | Theme generator kita sendiri untuk fixtures, jangan copy `theme1.xml` Microsoft |

---

## 11. Decided & Open Questions

### Decided
- **Lisensi:** MIT (file `LICENSE` ada di root, Copyright (c) 2026 Triadmoko Denny Fatrosa).
- **Module path:** `github.com/triadmoko/office` (sudah di `go.mod`).
- **Pure stdlib only:** dikonfirmasi sebagai constraint utama.

### Open (dapat ditentukan sambil jalan)
1. **Format prioritas v0.1**: DOCX, XLSX, atau PPTX dulu yang masuk MVP? (rekomendasi: DOCX karena fondasinya sudah ada)
2. **API untuk error**: error sentinel (`errors.Is`) saja, atau structured error type juga?
3. **Float vs Decimal di XLSX**: pakai `float64` (mudah, lossy) atau bikin tipe `Decimal` sendiri (akurat, kompleks)?
4. **`time.Time` di XLSX**: convert serial number ↔ time.Time otomatis, atau biarkan user yang convert?
5. **`bytes.Reader`-based open** atau hanya `io.ReaderAt`?

---

## 12. Sprint 0 — Prerequisites (sebelum mulai Tier MVP)

Tugas wajib sebelum membangun fitur baru:

1. Fix bug di `internal/ooxml.ResolveTarget` (".." traversal salah). File: `internal/ooxml/package.go`. Sudah teridentifikasi di session sebelumnya.
2. Tambahkan **package writer** di `internal/ooxml` (zip.Writer + serializer CT/Rels). File baru: `internal/ooxml/package_writer.go`.
3. Tambahkan **`internal/xmlwriter`** untuk emit XML dengan namespace prefix yang benar (Office punya konvensi prefix `w:`, `r:`, dst yang harus dipertahankan agar diterima Office).
4. Tambahkan **opc properties parser/writer** (`internal/opcprops`) untuk core.xml & app.xml. Wajib agar file diterima Office.
5. **Fuzz harness** untuk parser content-types, rels, sheet tokenizer.
6. Tambah **golden test infrastructure** (helper untuk diff DOCX zip-aware).
7. **CI**: GitHub Actions matrix Go 1.23 / 1.24 di linux/macos/windows, plus job `compat-test` yang upload artifact ke release dengan `unzip -t` validation.
8. Tetapkan **lisensi** + `LICENSE` file + header `// SPDX-License-Identifier: MIT`.

**Estimasi Sprint 0:** 1–2 minggu.

---

## 13. Verifikasi PRD Ini

Cara memvalidasi PRD tepat sebelum eksekusi:
- [x] Lisensi terkonfirmasi: MIT
- [ ] Pengguna mengonfirmasi prioritas format v0.1 (Q11.1)
- [ ] Pengguna mengonfirmasi cakupan MVP (Tier MVP table di §4 disetujui atau direvisi)
- [ ] Tinjau `internal/ooxml/package.go` + reproduksi bug `ResolveTarget` ".." sebagai test case yang gagal
- [ ] Konfirmasi fixture testdata yang akan dipakai (produce sendiri atau copy file Office user yang anonim?)

---

## 14. Critical Files (untuk referensi implementasi)

| File | Status | Aksi |
|---|---|---|
| `internal/ooxml/package.go` | Existing, ada bug `..` traversal | Perbaiki + tambah writer counterpart |
| `internal/ooxml/content_types.go` | Existing, read-only | Tambah `MarshalXML` / writer |
| `internal/ooxml/rels.go` | Existing, read-only | Tambah writer |
| `internal/ooxml/path.go` | Existing | Audit + perketat validasi traversal |
| `internal/ooxml/namespaces.go` | Existing minimal | Perluas: tambah NS untuk SML/PML/DML |
| `docx/open.go`, `docx/write.go`, `docx/body.go` | Stub level | Ganti dengan model paragraph/run + writer |
| `xlsx/open.go` | Stub | Tambah sheet iterator, cell, sharedStrings |
| `xlsx/doc.go` | Stub | API seperti §7.2 |
| `pptx/open.go` | Stub | Tambah slide/shape API §7.3 |
| `cmd/office/main.go` | CLI minimal | Perluas: export, import, info commands |

---

**Catatan akhir:** PRD ini bersifat hidup. Setiap selesai milestone, revisi PRD untuk mencerminkan pembelajaran (misalnya fitur yang ternyata trivial → naik tier, atau yang ternyata kompleks → turun tier).
