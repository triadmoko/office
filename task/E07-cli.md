# EPIC OFFICE-E07 — CLI Tooling

> **Sprint:** 8
> **Total points:** 8

## Daftar Ticket

| ID | Title | Type | Sprint | Points | Priority |
|---|---|---|---|---|---|
| [OFFICE-701](#office-701) | Refactor CLI ke subcommand pattern | Story | 8 | 3 | P3 |
| [OFFICE-702](#office-702) | CLI: docx-to-text + xlsx-to-csv export | Story | 8 | 5 | P3 |

---

## OFFICE-701

### [Story] Refactor CLI ke subcommand pattern

```
Type     : Story
Priority : P3
Points   : 3
Sprint   : 8
Epic     : OFFICE-E07
File     : cmd/office/main.go
```

#### Acceptance Criteria
- [ ] **AC1**: Subcommands:
  - `office info <file>` — print metadata (format, slide count / sheet count / paragraph count, file size)
  - `office text <file>` — extract plain text dari format apa pun
  - `office new docx|xlsx|pptx <out>` — generate template
- [ ] **AC2**: Auto-detect format dari extension.
- [ ] **AC3**: Help: `office --help`, `office <cmd> --help`.
- [ ] **AC4**: Test E2E: setiap subcommand happy path + 1 error case.

---

## OFFICE-702

### [Story] CLI: docx-to-text + xlsx-to-csv export

```
Type     : Story
Priority : P3
Points   : 5
Sprint   : 8
Epic     : OFFICE-E07
Depends  : OFFICE-701
```

#### Acceptance Criteria
- [ ] **AC1**: `office text doc.docx` — concat semua paragraph text, satu paragraph per line.
- [ ] **AC2**: `office csv data.xlsx --sheet=Sheet1` — output CSV ke stdout (RFC 4180 compliant).
- [ ] **AC3**: Multi-sheet support: `--sheet=*` → multiple files in current dir.
- [ ] **AC4**: Date format: `--date-format=2006-01-02` (Go time layout).
- [ ] **AC5**: Test E2E.

---

**Navigasi:** [← E06 Documentation](./E06-documentation.md) | [Index](./README.md) | [E08 Security →](./E08-security.md)
