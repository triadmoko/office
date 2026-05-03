# EPIC OFFICE-E01 — OPC Foundation Hardening

> **Goal:** Lengkapi `internal/ooxml` agar siap mendukung write + round-trip + secure parsing.
> **Sprint:** 0
> **Total points:** 34

## Daftar Ticket

| ID | Title | Type | Points | Priority |
|---|---|---|---|---|
| [OFFICE-001](#office-001) | Fix `ResolveTarget` ".." traversal silent-drop | Bug | 3 | P1 |
| [OFFICE-002](#office-002) | Tambah `ErrPathTraversal` + harden `NormalizePartName` | Task | 2 | P1 |
| [OFFICE-003](#office-003) | Implementasi Package Writer | Story | 8 | P0 |
| [OFFICE-004](#office-004) | Buat package `internal/xmlwriter` | Task | 5 | P1 |
| [OFFICE-005](#office-005) | Namespace constants & content-type registry | Task | 3 | P2 |
| [OFFICE-006](#office-006) | `internal/opcprops` (core.xml + app.xml) | Task | 5 | P1 |
| [OFFICE-007](#office-007) | Audit `archive/zip` untuk anti zip-bomb | Spike | 2 | P1 |
| [OFFICE-008](#office-008) | `OpenOptions` dengan zip-bomb protection | Story | 5 | P1 |
| [OFFICE-009](#office-009) | Fuzz test untuk parser ContentTypes & Relationships | Task | 3 | P2 |

---

## OFFICE-001

### [Bug] Fix `ResolveTarget` ".." traversal silent-drop

```
Type     : Bug
Priority : P1 (Critical — security-relevant)
Points   : 3
Sprint   : 0
Epic     : OFFICE-E01
File     : internal/ooxml/package.go (joinResolveOPC)
Tag      : security, oopc
```

#### Background
Function `joinResolveOPC` menangani segment `".."` dengan `segs = segs[:len(segs)-1]`. Jika sudah di root (`len(segs)==0`), traversal **silent drop** alih-alih error. Ini di-flag sebelumnya sebagai bug (claude-mem 697, 698).

#### Acceptance Criteria
- [ ] **AC1**: `ResolveTarget("/_rels/.rels", "../../etc/passwd")` mengembalikan error `ErrPathTraversal` (sentinel baru), **bukan** path yang silently di-clamp.
- [ ] **AC2**: `ResolveTarget("/word/_rels/document.xml.rels", "../document.xml")` mengembalikan `/word/document.xml` (kasus normal harus tetap jalan).
- [ ] **AC3**: `ResolveTarget("/word/_rels/document.xml.rels", "../../word/document.xml")` mengembalikan `/word/document.xml` (parent navigation legal).
- [ ] **AC4**: Test case existing `TestResolveTarget` di `content_types_test.go` tetap hijau.
- [ ] **AC5**: Tambahkan minimal 5 test case baru: traversal di luar package, multiple `..`, leading `./`, absolute path target, empty target.
- [ ] **AC6**: API berubah → `ResolveTarget(rels, target string) (string, error)`. Update semua call-site di `package.go`.
- [ ] **AC7**: Dokumentasikan perilaku baru di doc-comment fungsi.

#### Definition of Done
- Code review approved
- Coverage `joinResolveOPC` ≥ 95%
- `go test ./internal/ooxml/...` hijau
- `go vet ./...` clean

#### Technical Notes
- Tambah `ErrPathTraversal = errors.New("ooxml: target traverses out of package")` di `errors.go`.
- Pertimbangkan helper `validatePartName(name)` untuk re-use di package writer (cek absolute, traversal, encoding).

---

## OFFICE-002

### [Task] Tambah `ErrPathTraversal` sentinel + harden `NormalizePartName`

```
Type     : Task
Priority : P1
Points   : 2
Sprint   : 0
Epic     : OFFICE-E01
File     : internal/ooxml/path.go, errors.go
Depends  : —
Blocks   : OFFICE-001
```

#### Acceptance Criteria
- [ ] **AC1**: `NormalizePartName("../foo")` mengembalikan error `ErrPathTraversal`.
- [ ] **AC2**: `NormalizePartName("\\windows\\path")` di-normalize ke `/windows/path` (slash conversion tetap aman, bukan traversal).
- [ ] **AC3**: `NormalizePartName("/word/../etc/x")` → error.
- [ ] **AC4**: Empty string → error `ErrInvalidPartName`.
- [ ] **AC5**: Path dengan null byte (`\x00`) → error.
- [ ] **AC6**: Test coverage ≥ 90%.

---

## OFFICE-003

### [Story] Implementasi Package Writer (zip.Writer + serializer CT/Rels)

```
Type     : Story
Priority : P0 (Blocker untuk semua write)
Points   : 8
Sprint   : 0
Epic     : OFFICE-E01
File     : internal/ooxml/package_writer.go (NEW)
Depends  : OFFICE-001, OFFICE-002
Blocks   : OFFICE-E02 (DOCX write), OFFICE-E03, OFFICE-E04
```

#### User Story
> Sebagai developer paket `docx/xlsx/pptx`, saya butuh API untuk **menulis** OPC package (ZIP + Content Types + Relationships) sehingga bisa men-generate file Office baru tanpa mengulang infrastruktur ZIP.

#### Acceptance Criteria
- [ ] **AC1**: API public:
  ```go
  type PackageWriter struct { /* ... */ }
  func NewPackageWriter(w io.Writer) *PackageWriter
  func (pw *PackageWriter) AddPart(name, contentType string, body io.Reader) error
  func (pw *PackageWriter) AddPartBytes(name, contentType string, body []byte) error
  func (pw *PackageWriter) AddRelationships(partName string, rels *Relationships) error
  func (pw *PackageWriter) Close() error
  ```
- [ ] **AC2**: `Close()` otomatis emit `[Content_Types].xml` di akhir berdasarkan part yang ditambahkan + rules content-type per ekstensi.
- [ ] **AC3**: `Close()` otomatis emit `_rels/.rels` jika ada relationships level package.
- [ ] **AC4**: Output ZIP dapat dibuka kembali via `ooxml.Open()` tanpa error → round-trip integrity.
- [ ] **AC5**: ZIP entries dalam **urutan deterministik** (sort lexicographic), kecuali `[Content_Types].xml` selalu pertama (Office requirement).
- [ ] **AC6**: ZIP modification time = `time.Time{}` (epoch) untuk reproducibility.
- [ ] **AC7**: Reject `AddPart` dengan name invalid (call `NormalizePartName`).
- [ ] **AC8**: Reject double-add part name yang sama → `ErrDuplicatePart`.
- [ ] **AC9**: Test: build minimal docx → `unzip -t output.docx` exit 0.
- [ ] **AC10**: Test golden: hash SHA-256 output stabil antar run.

#### Subtasks
- [ ] OFFICE-003.1 — `MarshalXML` untuk `ContentTypes` (3 pts)
- [ ] OFFICE-003.2 — `MarshalXML` untuk `Relationships` (2 pts)
- [ ] OFFICE-003.3 — `PackageWriter` core dengan deterministic ZIP ordering (3 pts)

---

## OFFICE-004

### [Task] Buat package `internal/xmlwriter`

```
Type     : Task
Priority : P1
Points   : 5
Sprint   : 0
Epic     : OFFICE-E01
File     : internal/xmlwriter/writer.go (NEW)
Depends  : —
Blocks   : OFFICE-005, OFFICE-E02 (semua writer ML)
```

#### Background
`encoding/xml` standard library tidak mempertahankan namespace prefix yang tepat (`w:p`, `r:id`, dst). Office Word/Excel **gagal membuka** file jika prefix berbeda dari yang diharapkan. Kita butuh wrapper writer yang:
- Mengizinkan deklarasi namespace prefix global di root.
- Menjamin atribut `w:val`, `r:id` tetap pakai prefix benar.
- Menulis self-closing tag yang valid.
- Escape karakter sesuai XML 1.0.

#### Acceptance Criteria
- [ ] **AC1**: API:
  ```go
  type Writer struct { /* ... */ }
  func New(w io.Writer) *Writer
  func (w *Writer) StartElement(name xml.Name, attrs []xml.Attr) error
  func (w *Writer) EndElement() error
  func (w *Writer) CharData(s string) error
  func (w *Writer) DeclareNamespace(prefix, uri string)
  func (w *Writer) Close() error  // emit XML decl + final newline
  ```
- [ ] **AC2**: Output start dengan `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`.
- [ ] **AC3**: Namespace prefix dipertahankan persis sesuai deklarasi (case-sensitive: `w` ≠ `W`).
- [ ] **AC4**: Karakter `<`, `>`, `&`, `"`, `'`, `\t`, `\n`, `\r`, NUL → diescape sesuai XML 1.0 spec.
- [ ] **AC5**: Tag tanpa child → self-closing (`<w:tab/>`).
- [ ] **AC6**: Performance: append-only, tidak buffer keseluruhan dokumen di memori.
- [ ] **AC7**: Round-trip test: parse output dengan `encoding/xml` standar → semua tag/attr/text recovered.
- [ ] **AC8**: Test fuzz: random byte input ke `CharData` tidak menghasilkan output yang tidak valid XML.

#### Definition of Done
- Coverage ≥ 90%
- Documented dengan example di package doc comment

---

## OFFICE-005

### [Task] Tambah namespace constants & content-type registry

```
Type     : Task
Priority : P2
Points   : 3
Sprint   : 0
Epic     : OFFICE-E01
File     : internal/ooxml/namespaces.go
Depends  : —
```

#### Acceptance Criteria
- [ ] **AC1**: Tambah konstanta untuk seluruh NS yang dibutuhkan MVP:
  - WordprocessingML: `w`, `w14`, `w15`, `wp`, `wp14`
  - SpreadsheetML: `x`, `xr`, `xr2`, `xr3`, `mc`
  - PresentationML: `p`, `p14`, `p15`
  - DrawingML: `a`, `a14`, `r` (relationships in markup), `pic`
- [ ] **AC2**: Constant naming: `NS<Pascal>` untuk URI, `Prefix<Pascal>` untuk prefix.
- [ ] **AC3**: Content-type strings: tambah `CT<Format><Part>` untuk minimal: `styles`, `numbering`, `theme`, `settings`, `fontTable`, `webSettings`, `core`, `app`, `customXml`, `image/png`, `image/jpeg`.
- [ ] **AC4**: Map `ExtensionToContentType` untuk ekstensi default: `xml`, `rels`, `png`, `jpg`, `jpeg`, `gif`, `bmp`, `bin`, `vml`.
- [ ] **AC5**: Test sanity: setiap konstanta non-empty, NS URI valid (parseable as URL).

---

## OFFICE-006

### [Task] Buat package `internal/opcprops` (core.xml + app.xml)

```
Type     : Task
Priority : P1
Points   : 5
Sprint   : 0
Epic     : OFFICE-E01
File     : internal/opcprops/props.go (NEW)
Depends  : OFFICE-004, OFFICE-005
```

#### Acceptance Criteria
- [ ] **AC1**: Struct `CoreProperties` dengan field: Title, Subject, Creator, Keywords, Description, LastModifiedBy, Revision, Created (time.Time), Modified (time.Time), Category, ContentStatus, Language, Version.
- [ ] **AC2**: Struct `AppProperties` dengan field: Application, AppVersion, Company, Manager, DocSecurity, ScaleCrop, LinksUpToDate, SharedDoc, HyperlinksChanged.
- [ ] **AC3**: `ParseCore(io.Reader) (*CoreProperties, error)` — handle namespace `dc:`, `dcterms:`, `cp:`.
- [ ] **AC4**: `ParseApp(io.Reader) (*AppProperties, error)`.
- [ ] **AC5**: `(*CoreProperties).WriteTo(w io.Writer) (int64, error)` — emit valid XML dengan namespace `cp`, `dc`, `dcterms`, `xsi`.
- [ ] **AC6**: `(*AppProperties).WriteTo(w io.Writer) (int64, error)`.
- [ ] **AC7**: Round-trip test: Parse → Write → Parse menghasilkan struct yang sama (kecuali timestamp re-formatted).
- [ ] **AC8**: Default value untuk file baru: Application = "github.com/triadmoko/office", AppVersion = library version.

---

## OFFICE-007

### [Spike] Audit `archive/zip` untuk anti zip-bomb

```
Type     : Spike
Priority : P1
Points   : 2
Sprint   : 0
Epic     : OFFICE-E08 (Security)
Time-box : 1 hari
```

#### Goal
Memahami batasan `archive/zip` Go stdlib terkait:
1. Apakah ada built-in protection terhadap zip bomb?
2. Apakah `zip.File.Open()` membuka stream lazy atau decompress sekaligus?
3. Bagaimana cara membatasi total uncompressed size sebelum parse?

#### Deliverable
Tuliskan finding di `docs/security/zip-bomb-mitigation.md` (file baru). Termasuk:
- Default behavior Go zip
- Strategi limit per-entry: wrap reader dengan `io.LimitReader(rc, MaxPartBytes)`.
- Strategi limit total: track sum bytes read.
- Rekomendasi default: `MaxBytes=1GiB`, `MaxParts=10000`, `MaxPartBytes=256MiB`.

#### Acceptance Criteria
- [ ] Dokumen tertulis dengan referensi ke source Go stdlib
- [ ] PoC zip bomb file 10KB yang expand jadi 1GB → konfirmasi current `ooxml.Open()` rentan
- [ ] Rekomendasi konkret untuk OFFICE-008

---

## OFFICE-008

### [Story] Implementasi `OpenOptions` dengan zip-bomb protection

```
Type     : Story
Priority : P1
Points   : 5
Sprint   : 0
Epic     : OFFICE-E08
Depends  : OFFICE-007
```

#### Acceptance Criteria
- [ ] **AC1**: API:
  ```go
  type OpenOptions struct {
      MaxBytes     int64 // default 1 GiB; 0 = unlimited
      MaxParts     int   // default 10000
      MaxPartBytes int64 // default 256 MiB per part
  }
  func OpenWithOptions(r io.ReaderAt, size int64, opts OpenOptions) (*Package, error)
  ```
- [ ] **AC2**: `Open()` (existing) memanggil `OpenWithOptions` dengan default.
- [ ] **AC3**: ZIP dengan total uncompressed > MaxBytes → error `ErrPackageTooLarge` saat first read.
- [ ] **AC4**: ZIP dengan jumlah entry > MaxParts → error `ErrTooManyParts` saat init.
- [ ] **AC5**: Part individual yang decompress > MaxPartBytes → error saat read (via wrapped reader).
- [ ] **AC6**: Test dengan zip bomb fixture (PoC dari OFFICE-007) → harus rejected.
- [ ] **AC7**: Dokumentasikan default & override di doc.go.

---

## OFFICE-009

### [Task] Fuzz test untuk parser ContentTypes & Relationships

```
Type     : Task
Priority : P2
Points   : 3
Sprint   : 0
Epic     : OFFICE-E01
File     : internal/ooxml/fuzz_test.go (NEW)
```

#### Acceptance Criteria
- [ ] **AC1**: `FuzzParseContentTypes` di `internal/ooxml/` — seed dengan minimal 5 valid corpus + 5 malformed.
- [ ] **AC2**: `FuzzParseRelationships` — seed serupa.
- [ ] **AC3**: `FuzzResolveTarget(rels, target string)` — seed beragam path patterns.
- [ ] **AC4**: Run `go test -fuzz=. -fuzztime=30s` lokal tanpa panic atau infinite loop.
- [ ] **AC5**: Tambah ke CI: nightly job dengan `-fuzztime=5m` untuk setiap fuzz target.
- [ ] **AC6**: Setiap crash temuan → tambahkan ke regression test di test file biasa.

---

**Navigasi:** [← Index](./README.md) | [Konvensi](./00-conventions.md) | [E02 DOCX →](./E02-docx-mvp.md)
