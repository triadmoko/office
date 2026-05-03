# EPIC OFFICE-E08 — Security & Hardening

> **Sprint:** 0, 2, 4
> **Total points:** 21

## Daftar Ticket

| ID | Title | Type | Sprint | Points | Priority |
|---|---|---|---|---|---|
| OFFICE-801 | Anti zip-bomb di Open path | — | 0 | — | P1 |
| [OFFICE-802](#office-802) | XML billion-laughs / entity expansion guard | Task | 0 | 3 | P1 |
| [OFFICE-803](#office-803) | Path validation di PackageWriter | Task | 0 | 2 | P1 |
| [OFFICE-804](#office-804) | Memory budget guard di parsers | Task | 4 | 3 | P2 |
| [OFFICE-805](#office-805) | Fuzz worksheet tokenizer | Task | 4 | 3 | P2 |
| [OFFICE-806](#office-806) | Threat model dokumentasi | Spike | 0 | 2 | P2 |
| [OFFICE-807](#office-807) | SECURITY.md + responsible disclosure | Task | 0 | 1 | P3 |

---

## OFFICE-801

> **Sudah ditrack di** [`OFFICE-007`](./E01-opc-foundation.md#office-007) **+** [`OFFICE-008`](./E01-opc-foundation.md#office-008). Cross-reference here.

---

## OFFICE-802

### [Task] XML billion-laughs / entity expansion guard

```
Type     : Task
Priority : P1
Points   : 3
Sprint   : 0
Epic     : OFFICE-E08
```

#### Background
`encoding/xml` Go default tidak resolve external entity, tapi bisa expand internal entities dalam jumlah besar (DTD-defined). Verifikasi dan tambah hard limit.

#### Acceptance Criteria
- [ ] **AC1**: Test PoC: file dengan DTD `<!ENTITY x "AAA..." >` repeat → confirm Go behavior.
- [ ] **AC2**: Jika rentan, wrap decoder dengan `decoder.Strict = true; decoder.Entity = nil` (atau custom Token reader yang reject DOCTYPE).
- [ ] **AC3**: Reject DOCTYPE entirely di parsers OOXML (Office files seharusnya tidak punya DOCTYPE).
- [ ] **AC4**: Sentinel `ErrDoctypeForbidden`.
- [ ] **AC5**: Test: Open file dengan DOCTYPE → error.

---

## OFFICE-803

### [Task] Path validation di PackageWriter

```
Type     : Task
Priority : P1
Points   : 2
Sprint   : 0
Epic     : OFFICE-E08
Depends  : OFFICE-003
```

#### Acceptance Criteria
- [ ] **AC1**: `PackageWriter.AddPart` reject name dengan: absolute Windows path (`C:\...`), backslash setelah normalize, leading `..`, NUL byte, length > 255.
- [ ] **AC2**: ZIP entry name tidak boleh sama secara case-insensitive (Windows constraint).
- [ ] **AC3**: Test untuk masing-masing reject case.

---

## OFFICE-804

### [Task] Memory budget guard di parsers

```
Type     : Task
Priority : P2
Points   : 3
Sprint   : 4
Epic     : OFFICE-E08
```

#### Acceptance Criteria
- [ ] **AC1**: SharedStrings parser: jika count > 10M strings → return error (configurable via `OpenOptions.MaxSharedStrings`).
- [ ] **AC2**: Worksheet parser (random-access mode): jika cell count > 10M → error, sarankan stream mode.
- [ ] **AC3**: DOCX paragraph count > 100k → warning log (atau error?, decision time).
- [ ] **AC4**: Test PoC dengan synthetic large file.

---

## OFFICE-805

### [Task] Fuzz worksheet tokenizer

```
Type     : Task
Priority : P2
Points   : 3
Sprint   : 4
Epic     : OFFICE-E08
Depends  : OFFICE-203
```

#### Acceptance Criteria
- [ ] **AC1**: `FuzzSheetReader` di `internal/sml/`.
- [ ] **AC2**: Run 30 menit lokal tanpa panic.
- [ ] **AC3**: Termasuk di nightly CI fuzz job.

---

## OFFICE-806

### [Spike] Threat model dokumentasi

```
Type     : Spike
Priority : P2
Points   : 2
Sprint   : 0
Time-box : 1 hari
```

#### Deliverable
`docs/security/THREAT_MODEL.md` mencakup:
- Trust boundaries (untrusted file from upload)
- Attack vectors (zip bomb, billion laughs, path traversal, OOM, infinite loop in XML)
- Mitigations
- Out of scope (timing attacks, side channels)

#### Acceptance Criteria
- [ ] **AC1**: Dokumen committed
- [ ] **AC2**: Referenced di README + SECURITY.md
- [ ] **AC3**: Tracked threats dipetakan ke ticket terkait

---

## OFFICE-807

### [Task] SECURITY.md + responsible disclosure

```
Type     : Task
Priority : P3
Points   : 1
Sprint   : 0
Epic     : OFFICE-E08
```

#### Acceptance Criteria
- [ ] **AC1**: SECURITY.md mengikuti format GitHub default.
- [ ] **AC2**: Email atau PGP key untuk private report.
- [ ] **AC3**: SLA respons: 7 hari acknowledge, 90 hari fix sebelum disclose.

---

**Navigasi:** [← E07 CLI](./E07-cli.md) | [Index](./README.md) | [Backlog Future →](./99-backlog-future.md)
