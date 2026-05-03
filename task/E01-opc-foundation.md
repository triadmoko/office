# EPIC OFFICE-E01 — OPC Foundation Hardening

> **Goal:** Lengkapi `internal/ooxml` agar siap mendukung write + round-trip + secure parsing.
> **Sprint:** 0
> **Total points:** 34
> **Status:** **Selesai** (2026-05). Kode utama: [`internal/ooxml`](../internal/ooxml/), [`internal/xmlwriter`](../internal/xmlwriter/), [`internal/opcprops`](../internal/opcprops/). **Tindak lanjut terpisah:** fuzz **nightly** di CI masih pada tiket [OFFICE-502](./E05-ci-cd.md#office-502) (sasaran fuzz sudah ada di `internal/ooxml/fuzz_test.go`).

## Daftar Ticket

| ID | Title | Type | Points | Priority | Status |
|---|---|---|---|---|---|
| [OFFICE-001](#office-001) | Fix `ResolveTarget` ".." traversal silent-drop | Bug | 3 | P1 | Done |
| [OFFICE-002](#office-002) | Tambah `ErrPathTraversal` + harden `NormalizePartName` | Task | 2 | P1 | Done |
| [OFFICE-003](#office-003) | Implementasi Package Writer | Story | 8 | P0 | Done |
| [OFFICE-004](#office-004) | Buat package `internal/xmlwriter` | Task | 5 | P1 | Done |
| [OFFICE-005](#office-005) | Namespace constants & content-type registry | Task | 3 | P2 | Done |
| [OFFICE-006](#office-006) | `internal/opcprops` (core.xml + app.xml) | Task | 5 | P1 | Done |
| [OFFICE-007](#office-007) | Audit `archive/zip` untuk anti zip-bomb | Spike | 2 | P1 | Done |
| [OFFICE-008](#office-008) | `OpenOptions` dengan zip-bomb protection | Story | 5 | P1 | Done |
| [OFFICE-009](#office-009) | Fuzz test untuk parser ContentTypes & Relationships | Task | 3 | P2 | Done* |

\*OFFICE-009 **AC5** (workflow fuzz terjadwal di CI) dilacak sebagai [OFFICE-502](./E05-ci-cd.md#office-502); target fuzz sudah ada di repositori.

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
- [x] **AC1**: `ResolveTarget("/_rels/.rels", "../../etc/passwd")` mengembalikan error `ErrPathTraversal` (sentinel baru), **bukan** path yang silently di-clamp.
- [x] **AC2**: `ResolveTarget("/word/_rels/document.xml.rels", "../document.xml")` mengembalikan `/word/document.xml` (kasus normal harus tetap jalan).
- [x] **AC3**: `ResolveTarget("/word/_rels/document.xml.rels", "../../word/document.xml")` mengembalikan `/word/document.xml` (parent navigation legal).
- [x] **AC4**: Test case existing `TestResolveTarget` di `content_types_test.go` tetap hijau.
- [x] **AC5**: Tambahkan minimal 5 test case baru: traversal di luar package, multiple `..`, leading `./`, absolute path target, empty target.
- [x] **AC6**: API berubah → `ResolveTarget(rels, target string) (string, error)`. Update semua call-site di `package.go`.
- [x] **AC7**: Dokumentasikan perilaku baru di doc-comment fungsi.

#### Definition of Done
- [x] Code review approved *(proses tim)*
- [x] Coverage `joinResolveOPC` ≥ 95% *(≈95% di `package.go`)*
- [x] `go test ./internal/ooxml/...` hijau
- [x] `go vet ./...` clean

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
- [x] **AC1**: `NormalizePartName("../foo")` mengembalikan error `ErrPathTraversal`.
- [x] **AC2**: `NormalizePartName("\\windows\\path")` di-normalize ke `/windows/path` (slash conversion tetap aman, bukan traversal).
- [x] **AC3**: `NormalizePartName("/word/../etc/x")` → error.
- [x] **AC4**: Empty string → error `ErrInvalidPartName`.
- [x] **AC5**: Path dengan null byte (`\x00`) → error.
- [x] **AC6**: Test coverage ≥ 90%.

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
- [x] **AC1**: API public:
  ```go
  type PackageWriter struct { /* ... */ }
  func NewPackageWriter(w io.Writer) *PackageWriter
  func (pw *PackageWriter) AddPart(name, contentType string, body io.Reader) error
  func (pw *PackageWriter) AddPartBytes(name, contentType string, body []byte) error
  func (pw *PackageWriter) AddRelationships(partName string, rels *Relationships) error
  func (pw *PackageWriter) Close() error
  ```
- [x] **AC2**: `Close()` otomatis emit `[Content_Types].xml` di akhir berdasarkan part yang ditambahkan + rules content-type per ekstensi.
- [x] **AC3**: `Close()` otomatis emit `_rels/.rels` jika ada relationships level package.
- [x] **AC4**: Output ZIP dapat dibuka kembali via `ooxml.Open()` tanpa error → round-trip integrity.
- [x] **AC5**: ZIP entries dalam **urutan deterministik** (sort lexicographic), kecuali `[Content_Types].xml` selalu pertama (Office requirement).
- [x] **AC6**: ZIP modification time = `time.Time{}` (epoch) untuk reproducibility.
- [x] **AC7**: Reject `AddPart` dengan name invalid (call `NormalizePartName`).
- [x] **AC8**: Reject double-add part name yang sama → `ErrDuplicatePart`.
- [x] **AC9**: Test: build minimal docx → `unzip -t output.docx` exit 0.
- [x] **AC10**: Test golden: hash SHA-256 output stabil antar run.

#### Subtasks
- [x] OFFICE-003.1 — `MarshalXML` untuk `ContentTypes` (3 pts)
- [x] OFFICE-003.2 — `MarshalXML` untuk `Relationships` (2 pts)
- [x] OFFICE-003.3 — `PackageWriter` core dengan deterministic ZIP ordering (3 pts)

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
- [x] **AC1**: API:
  ```go
  type Writer struct { /* ... */ }
  func New(w io.Writer) *Writer
  func (w *Writer) StartElement(name xml.Name, attrs []xml.Attr) error
  func (w *Writer) EndElement() error
  func (w *Writer) CharData(s string) error
  func (w *Writer) DeclareNamespace(prefix, uri string)
  func (w *Writer) Close() error  // emit XML decl + final newline
  ```
- [x] **AC2**: Output start dengan `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`.
- [x] **AC3**: Namespace prefix dipertahankan persis sesuai deklarasi (case-sensitive: `w` ≠ `W`).
- [x] **AC4**: Karakter `<`, `>`, `&`, `"`, `'`, `\t`, `\n`, `\r`, NUL → diescape sesuai XML 1.0 spec.
- [x] **AC5**: Tag tanpa child → self-closing (`<w:tab/>`).
- [x] **AC6**: Performance: append-only, tidak buffer keseluruhan dokumen di memori.
- [x] **AC7**: Round-trip test: parse output dengan `encoding/xml` standar → semua tag/attr/text recovered.
- [x] **AC8**: Test fuzz: random byte input ke `CharData` tidak menghasilkan output yang tidak valid XML.

#### Definition of Done
- [x] Coverage ≥ 90%
- [x] Documented dengan example di package doc comment

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
- [x] **AC1**: Tambah konstanta untuk seluruh NS yang dibutuhkan MVP:
  - WordprocessingML: `w`, `w14`, `w15`, `wp`, `wp14`
  - SpreadsheetML: `x`, `xr`, `xr2`, `xr3`, `mc`
  - PresentationML: `p`, `p14`, `p15`
  - DrawingML: `a`, `a14`, `r` (relationships in markup), `pic`
- [x] **AC2**: Constant naming: `NS<Pascal>` untuk URI, `Prefix<Pascal>` untuk prefix.
- [x] **AC3**: Content-type strings: tambah `CT<Format><Part>` untuk minimal: `styles`, `numbering`, `theme`, `settings`, `fontTable`, `webSettings`, `core`, `app`, `customXml`, `image/png`, `image/jpeg`.
- [x] **AC4**: Map `ExtensionToContentType` untuk ekstensi default: `xml`, `rels`, `png`, `jpg`, `jpeg`, `gif`, `bmp`, `bin`, `vml`.
- [x] **AC5**: Test sanity: setiap konstanta non-empty, NS URI valid (parseable as URL).

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
- [x] **AC1**: Struct `CoreProperties` dengan field: Title, Subject, Creator, Keywords, Description, LastModifiedBy, Revision, Created (time.Time), Modified (time.Time), Category, ContentStatus, Language, Version.
- [x] **AC2**: Struct `AppProperties` dengan field: Application, AppVersion, Company, Manager, DocSecurity, ScaleCrop, LinksUpToDate, SharedDoc, HyperlinksChanged.
- [x] **AC3**: `ParseCore(io.Reader) (*CoreProperties, error)` — handle namespace `dc:`, `dcterms:`, `cp:`.
- [x] **AC4**: `ParseApp(io.Reader) (*AppProperties, error)`.
- [x] **AC5**: `(*CoreProperties).WriteTo(w io.Writer) (int64, error)` — emit valid XML dengan namespace `cp`, `dc`, `dcterms`, `xsi`.
- [x] **AC6**: `(*AppProperties).WriteTo(w io.Writer) (int64, error)`.
- [x] **AC7**: Round-trip test: Parse → Write → Parse menghasilkan struct yang sama (kecuali timestamp re-formatted).
- [x] **AC8**: Default value untuk file baru: Application = "github.com/triadmoko/office", AppVersion = library version.

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
- [x] Dokumen tertulis dengan referensi ke source Go stdlib
- [x] PoC zip bomb file 10KB yang expand jadi 1GB → konfirmasi current `ooxml.Open()` rentan
- [x] Rekomendasi konkret untuk OFFICE-008

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
- [x] **AC1**: API:
  ```go
  type OpenOptions struct {
      MaxBytes     int64 // default 1 GiB; 0 = unlimited
      MaxParts     int   // default 10000
      MaxPartBytes int64 // default 256 MiB per part
  }
  func OpenWithOptions(r io.ReaderAt, size int64, opts OpenOptions) (*Package, error)
  ```
- [x] **AC2**: `Open()` (existing) memanggil `OpenWithOptions` dengan default.
- [x] **AC3**: ZIP dengan total uncompressed > MaxBytes → error `ErrPackageTooLarge` saat first read.
- [x] **AC4**: ZIP dengan jumlah entry > MaxParts → error `ErrTooManyParts` saat init.
- [x] **AC5**: Part individual yang decompress > MaxPartBytes → error saat read (via wrapped reader).
- [x] **AC6**: Test dengan zip bomb fixture (PoC dari OFFICE-007) → harus rejected.
- [x] **AC7**: Dokumentasikan default & override di doc.go.

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
- [x] **AC1**: `FuzzParseContentTypes` di `internal/ooxml/` — seed dengan minimal 5 valid corpus + 5 malformed.
- [x] **AC2**: `FuzzParseRelationships` — seed serupa.
- [x] **AC3**: `FuzzResolveTarget(rels, target string)` — seed beragam path patterns.
- [x] **AC4**: Run `go test -fuzz=. -fuzztime=30s` lokal tanpa panic atau infinite loop.
- [ ] **AC5**: Tambah ke CI: nightly job dengan `-fuzztime=5m` untuk setiap fuzz target. *(Belum ada workflow terjadwal; dilanjutkan lewat [OFFICE-502](./E05-ci-cd.md#office-502). Target fuzz `FuzzParseContentTypes`, `FuzzParseRelationships`, `FuzzResolveTarget` sudah di `internal/ooxml/fuzz_test.go`.)*
- [x] **AC6**: Setiap crash temuan → tambahkan ke regression test di test file biasa.

---

**Navigasi:** [← Index](./README.md) | [Konvensi](./00-conventions.md) | [E02 DOCX →](./E02-docx-mvp.md)
