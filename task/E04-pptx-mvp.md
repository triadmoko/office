# EPIC OFFICE-E04 — PPTX MVP

> **Goal:** PPTX read+write untuk slide, layout, master, theme, shape (text box + picture), MVP-tier text formatting.
> **Sprint:** 6–7
> **Total points:** 63

## Daftar Ticket

| ID | Title | Type | Sprint | Points | Priority |
|---|---|---|---|---|---|
| [OFFICE-301](#office-301) | Read — Presentation & slide enumeration | Story | 6 | 5 | P1 |
| [OFFICE-302](#office-302) | Read — Theme parsing | Story | 6 | 5 | P2 |
| [OFFICE-303](#office-303) | Read — Shape tree per slide | Story | 6 | 8 | P1 |
| [OFFICE-304](#office-304) | Read — DrawingML text body | Story | 6 | 5 | P1 |
| [OFFICE-305](#office-305) | Read — Speaker notes | Story | 6 | 2 | P2 |
| [OFFICE-306](#office-306) | Write — New presentation builder | Story | 7 | 13 | P1 |
| [OFFICE-307](#office-307) | Round-Trip | Story | 7 | 5 | P2 |
| [OFFICE-308](#office-308) | Write — Tabel sederhana di slide | Story | 7 | 5 | P3 |

---

## OFFICE-301

### [Story] PPTX Read — Presentation & slide enumeration

```
Type     : Story
Priority : P1
Points   : 5
Sprint   : 6
Epic     : OFFICE-E04
File     : internal/pml/presentation.go, pptx/presentation.go
Depends  : OFFICE-005
```

#### Acceptance Criteria
- [ ] **AC1**: `pptx.Open()` parse `ppt/presentation.xml` → list slide IDs.
- [ ] **AC2**: `Presentation.Slides() []*Slide` — slides in order.
- [ ] **AC3**: `Presentation.SlideSize() SlideSize` (Width/Height in EMU, e.g., 9144000×6858000 untuk 4:3).
- [ ] **AC4**: `Presentation.Layouts()`, `Presentation.Masters()` — registry.
- [ ] **AC5**: `Slide.Layout() *Layout`, `Layout.Master() *Master`.
- [ ] **AC6**: Test fixture: pptx 3-slide.

---

## OFFICE-302

### [Story] PPTX Read — Theme parsing

```
Type     : Story
Priority : P2
Points   : 5
Sprint   : 6
Epic     : OFFICE-E04
File     : internal/dml/theme.go (NEW)
Depends  : OFFICE-301
```

#### Acceptance Criteria
- [ ] **AC1**: Parse `ppt/theme/theme1.xml` → `Theme` struct.
- [ ] **AC2**: `Theme.ColorScheme()` — accent1–6, bg1–2, fg1–2, hyperlink, followedHyperlink (RGB or scheme).
- [ ] **AC3**: `Theme.FontScheme()` — major (heading) & minor (body) font.
- [ ] **AC4**: `Theme.FormatScheme()` — preserve raw (Tier 3).
- [ ] **AC5**: `Slide.Theme()` resolve via inheritance: Slide → Layout → Master → Theme.

---

## OFFICE-303

### [Story] PPTX Read — Shape tree per slide

```
Type     : Story
Priority : P1
Points   : 8
Sprint   : 6
Epic     : OFFICE-E04
File     : pptx/shape.go, internal/dml/shape.go (NEW)
Depends  : OFFICE-301
```

#### Acceptance Criteria
- [ ] **AC1**: `Slide.Shapes() []Shape` (interface).
- [ ] **AC2**: Implementations: `*TextShape` (sp dengan txBody), `*Picture` (pic), `*GraphicFrame` (table/chart placeholder).
- [ ] **AC3**: `Shape.ID()`, `Shape.Name()`, `Shape.Transform() Transform` (Offset, Extent, Rotation, FlipH, FlipV).
- [ ] **AC4**: `TextShape.TextBody() *TextBody` — paragraphs of DrawingML text.
- [ ] **AC5**: `Picture.Source()` mengembalikan `io.Reader` ke media bytes + content type.
- [ ] **AC6**: `GraphicFrame.Kind()` enum: `Table`, `Chart`, `SmartArt`, `Other`.
- [ ] **AC7**: Group shape (`p:grpSp`) → flatten dengan inherited transform (atau expose tree, decision time).
- [ ] **AC8**: Test fixture: slide dengan title text, body bullet text, picture, table.

---

## OFFICE-304

### [Story] PPTX Read — DrawingML text body

```
Type     : Story
Priority : P1
Points   : 5
Sprint   : 6
Epic     : OFFICE-E04
File     : internal/dml/text.go (NEW), pptx/text_body.go
Depends  : OFFICE-303
```

#### Acceptance Criteria
- [ ] **AC1**: `TextBody.Paragraphs() []*TextParagraph`.
- [ ] **AC2**: `TextParagraph.Runs() []*TextRun`.
- [ ] **AC3**: `TextRun.Text()`, `Font()`, `Size()`, `Color()`, `Bold()`, `Italic()`, `Underline()`.
- [ ] **AC4**: `TextParagraph.Level() int` (0–8 outline level untuk bullet indent).
- [ ] **AC5**: `TextParagraph.Bullet() Bullet` enum: `None`, `Char(rune)`, `AutoNum(format)`, `Picture(rId)`.
- [ ] **AC6**: `TextParagraph.Alignment()`.
- [ ] **AC7**: Inheritance: jika run kosong properties → ambil dari paragraph → list style → master.
- [ ] **AC8**: Test: extract text dari title slide + body slide dengan 5 bullet.

---

## OFFICE-305

### [Story] PPTX Read — Speaker notes

```
Type     : Story
Priority : P2
Points   : 2
Sprint   : 6
Epic     : OFFICE-E04
Depends  : OFFICE-303, OFFICE-304
```

#### Acceptance Criteria
- [ ] **AC1**: `Slide.Notes() string` — concatenated text from notes slide body placeholder.
- [ ] **AC2**: Empty string jika tidak ada notes.
- [ ] **AC3**: Test fixture: pptx dengan speaker notes 2 paragraph.

---

## OFFICE-306

### [Story] PPTX Write — New presentation builder

```
Type     : Story
Priority : P1
Points   : 13
Sprint   : 7
Epic     : OFFICE-E04
File     : pptx/presentation.go, pptx/slide.go
Depends  : OFFICE-003, OFFICE-301–305
```

#### User Story
> Sebagai dev, saya ingin generate PPTX dengan slide title + content + image dari Go code.

#### Acceptance Criteria
- [ ] **AC1**: `pptx.NewPresentation()` mengembalikan blank deck dengan default theme + 1 master + 1 layout (Title and Content).
- [ ] **AC2**: API:
  ```go
  p := pptx.NewPresentation()
  s := p.AddSlide("Title and Content")
  s.AddTextBox(pptx.Inches(1), pptx.Inches(1), pptx.Inches(8), pptx.Inches(1)).SetText("Hello")
  s.AddPicture("logo.png", pptx.Inches(1), pptx.Inches(3), pptx.Inches(2), pptx.Inches(2))
  p.Save(w)
  ```
- [ ] **AC3**: Default 9 layouts (Title Slide, Title and Content, Section Header, Two Content, Comparison, Title Only, Blank, Content with Caption, Picture with Caption).
- [ ] **AC4**: Default theme (kita generate sendiri, **JANGAN** copy dari MS) dengan 12-color scheme + Calibri.
- [ ] **AC5**: Generated file lulus matrix compat: PowerPoint 365, Keynote, LibreOffice Impress, Google Slides.
- [ ] **AC6**: Slide size default 16:9 (12192000×6858000 EMU).
- [ ] **AC7**: Test: 5 fixtures (blank, title only, title+content, picture slide, multi-slide).

#### Subtasks
- [ ] OFFICE-306.1 — Default theme generator (kita-buat-sendiri) (4 pts)
- [ ] OFFICE-306.2 — Default 9 slide layouts (3 pts)
- [ ] OFFICE-306.3 — Slide & shape serialization (3 pts)
- [ ] OFFICE-306.4 — Image embedding (PNG/JPEG) (3 pts)

---

## OFFICE-307

### [Story] PPTX Round-Trip

```
Type     : Story
Priority : P2
Points   : 5
Sprint   : 7
Epic     : OFFICE-E04
Depends  : OFFICE-301–306
```

#### Acceptance Criteria
- [ ] **AC1**: Open + save tanpa edit → semantic equality.
- [ ] **AC2**: Modify slide 0 title → only slide1.xml changed; rest preserved verbatim.
- [ ] **AC3**: Test corpus: 3 pptx kompleks dari PowerPoint 365 (chart, smartart, video — preserve as raw).

---

## OFFICE-308

### [Story] PPTX Write — Tabel sederhana di slide

```
Type     : Story
Priority : P3
Points   : 5
Sprint   : 7 (stretch)
Epic     : OFFICE-E04
Depends  : OFFICE-306
```

#### Acceptance Criteria
- [ ] **AC1**: API:
  ```go
  t := slide.AddTable(rows, cols, x, y, w, h)
  t.Cell(0, 0).SetText("Header")
  ```
- [ ] **AC2**: Borders default (1pt black).
- [ ] **AC3**: Cell alignment, fill color.
- [ ] **AC4**: Compat test pass.

---

**Navigasi:** [← E03 XLSX](./E03-xlsx-mvp.md) | [Index](./README.md) | [E05 CI/CD →](./E05-ci-cd.md)
