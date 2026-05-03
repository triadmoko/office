# EPIC OFFICE-E06 — Documentation & Examples

> **Sprint:** 8
> **Total points:** 13

## Daftar Ticket

| ID | Title | Type | Sprint | Points | Priority |
|---|---|---|---|---|---|
| [OFFICE-601](#office-601) | README.md komprehensif | Task | 8 | 3 | P1 |
| [OFFICE-602](#office-602) | Per-package documentation | Task | 8 | 3 | P2 |
| [OFFICE-603](#office-603) | Examples folder | Task | 8 | 5 | P2 |
| [OFFICE-604](#office-604) | CHANGELOG.md + RELEASING.md | Task | 8 | 2 | P3 |

---

## OFFICE-601

### [Task] README.md komprehensif

```
Type     : Task
Priority : P1
Points   : 3
Sprint   : 8
Epic     : OFFICE-E06
File     : README.md (NEW)
```

#### Acceptance Criteria
- [ ] **AC1**: Sections: Features, Status (tier matrix), Install, Quick Start (5 example), Comparison vs alternatives, License, Contributing.
- [ ] **AC2**: Quick start 5 contoh: read docx text, write docx, read xlsx, write xlsx with format, write pptx with text.
- [ ] **AC3**: Badges: Go report card, coverage, CI, Go version, license.
- [ ] **AC4**: Link ke PRD.md, BACKLOG.md, ARCHITECTURE.md.

---

## OFFICE-602

### [Task] Per-package documentation

```
Type     : Task
Priority : P2
Points   : 3
Sprint   : 8
Epic     : OFFICE-E06
File     : docs/DOCX.md, docs/XLSX.md, docs/PPTX.md, docs/ARCHITECTURE.md
```

#### Acceptance Criteria
- [ ] **AC1**: Setiap dokumen format punya: Public API reference (godoc-style), Tier matrix (apa yang supported), 3 worked examples, Limitations.
- [ ] **AC2**: ARCHITECTURE.md: layered diagram, package responsibility, write strategy, round-trip strategy.
- [ ] **AC3**: Indonesian + English (atau English-only — keputusan owner).

---

## OFFICE-603

### [Task] Examples folder

```
Type     : Task
Priority : P2
Points   : 5
Sprint   : 8
Epic     : OFFICE-E06
File     : examples/ (NEW)
```

#### Acceptance Criteria
- [ ] **AC1**: 6 example program runnable:
  - `examples/docx-hello/`
  - `examples/docx-report/` — generate report dengan tabel
  - `examples/xlsx-export/` — export data ke spreadsheet
  - `examples/xlsx-stream/` — stream 100k rows
  - `examples/pptx-deck/` — generate deck dari template data
  - `examples/extract-text/` — universal text extractor
- [ ] **AC2**: Setiap example punya `go run .` yang berhasil dengan output di `out/`.
- [ ] **AC3**: README per example.

---

## OFFICE-604

### [Task] CHANGELOG.md + RELEASING.md

```
Type     : Task
Priority : P3
Points   : 2
Sprint   : 8
Epic     : OFFICE-E06
```

#### Acceptance Criteria
- [ ] **AC1**: CHANGELOG.md mengikuti format Keep a Changelog (Added/Changed/Deprecated/Removed/Fixed/Security).
- [ ] **AC2**: RELEASING.md menjelaskan langkah release: tag, changelog update, GitHub release, pkg.go.dev refresh.
- [ ] **AC3**: Setiap PR wajib update CHANGELOG (enforced via PR template).

---

**Navigasi:** [← E05 CI/CD](./E05-ci-cd.md) | [Index](./README.md) | [E07 CLI →](./E07-cli.md)
