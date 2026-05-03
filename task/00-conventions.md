# Konvensi & Sprint Plan

## Format Ticket

```
ID         : OFFICE-XXX
Type       : Epic | Story | Task | Bug | Spike
Priority   : P0 (Blocker) | P1 (Critical) | P2 (Major) | P3 (Minor)
Points     : 1 | 2 | 3 | 5 | 8 | 13 (Fibonacci)
Sprint     : Sprint number atau Backlog
Status     : To Do | In Progress | Review | Done
```

## Story Points Reference

| Points | Effort |
|---|---|
| 1 | Trivial (< 1 jam) |
| 2 | Small (1–3 jam) |
| 3 | Medium (1 hari) |
| 5 | Large (2–3 hari) |
| 8 | Very large (1 minggu) |
| 13 | Epic-sized (perlu di-split) |

## Sprint Plan (Indikatif)

| Sprint | Tema | Durasi | Output |
|---|---|---|---|
| **Sprint 0** | Foundation hardening | 2 minggu | Bug fix, package writer, fuzz, CI |
| **Sprint 1** | DOCX MVP — Read | 2 minggu | Parse paragraph/run/table → struct |
| **Sprint 2** | DOCX MVP — Write | 2 minggu | Build new DOCX with formatting |
| **Sprint 3** | DOCX MVP — Round-trip | 1 minggu | Open + modify + save fidelity |
| **Sprint 4** | XLSX MVP — Read | 2 minggu | Sheet iterator, cell types, shared strings |
| **Sprint 5** | XLSX MVP — Write | 2 minggu | Builder + StreamWriter |
| **Sprint 6** | PPTX MVP — Read | 2 minggu | Slide/shape/text body parser |
| **Sprint 7** | PPTX MVP — Write | 2 minggu | New presentation + slides |
| **Sprint 8** | Polish v0.5 release | 1 minggu | Docs, examples, perf benchmark |

**Total target v0.5 (MVP semua format):** ~14 minggu (3.5 bulan).

## Epic Overview

| Epic ID | Title | Sprint | Points | File |
|---|---|---|---|---|
| **OFFICE-E01** | OPC Foundation Hardening | Sprint 0 | 34 | [E01](./E01-opc-foundation.md) |
| **OFFICE-E02** | DOCX MVP | Sprint 1–3 | 89 | [E02](./E02-docx-mvp.md) |
| **OFFICE-E03** | XLSX MVP | Sprint 4–5 | 71 | [E03](./E03-xlsx-mvp.md) |
| **OFFICE-E04** | PPTX MVP | Sprint 6–7 | 63 | [E04](./E04-pptx-mvp.md) |
| **OFFICE-E05** | CI/CD & Quality | Sprint 0–8 | 21 | [E05](./E05-ci-cd.md) |
| **OFFICE-E06** | Documentation & Examples | Sprint 8 | 13 | [E06](./E06-documentation.md) |
| **OFFICE-E07** | CLI Tooling | Sprint 8 | 8 | [E07](./E07-cli.md) |
| **OFFICE-E08** | Security & Hardening | Sprint 0, 2, 4 | 21 | [E08](./E08-security.md) |
| **OFFICE-E99** | Post-MVP / Backlog | — | 55+ | [99](./99-backlog-future.md) |

**Total estimasi v0.5:** ~320 story points.

## Velocity Assumption

- 1 developer fokus penuh ≈ **20 pts/sprint** → 16 sprint = 4 bulan kalender
- 2 developer paralel ≈ **40 pts/sprint** → 8 sprint = 2 bulan kalender

## Sprint 0 Wajib

Sebelum Sprint 1 dimulai, **wajib** menyelesaikan ticket berikut karena memblokir feature work:

1. [OFFICE-001](./E01-opc-foundation.md#office-001) — Bug fix `ResolveTarget` ".." traversal
2. [OFFICE-002](./E01-opc-foundation.md#office-002) — Harden `NormalizePartName`
3. [OFFICE-003](./E01-opc-foundation.md#office-003) — Package Writer (blocker semua write)
4. [OFFICE-004](./E01-opc-foundation.md#office-004) — `internal/xmlwriter`
5. [OFFICE-005](./E01-opc-foundation.md#office-005) — Namespace constants & content-type registry
6. [OFFICE-006](./E01-opc-foundation.md#office-006) — `internal/opcprops` (core.xml + app.xml)
7. [OFFICE-501](./E05-ci-cd.md#office-501) — GitHub Actions CI
8. [OFFICE-503](./E05-ci-cd.md#office-503) — Golden test infrastructure
9. [OFFICE-802](./E08-security.md#office-802) — XML billion-laughs guard
10. [OFFICE-806](./E08-security.md#office-806) — Threat model

**Sprint 0 estimated: ~25 pts** (1 dev penuh, 1.5 minggu).
