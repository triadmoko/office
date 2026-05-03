# Task Directory — Pure-Go Office Library

**Repo:** `github.com/triadmoko/office`
**Diturunkan dari:** [`../PRD.md`](../PRD.md)
**Tanggal:** 2026-05-04

Direktori ini berisi backlog detail gaya Jira yang dipecah per-epic untuk memudahkan navigasi.

## Index

| File | Isi | Tickets | Points |
|---|---|---|---|
| [`00-conventions.md`](./00-conventions.md) | Konvensi ticket, story points, sprint plan, epic overview | — | — |
| [`E01-opc-foundation.md`](./E01-opc-foundation.md) | OPC Foundation Hardening (Sprint 0) | 9 | 34 |
| [`E02-docx-mvp.md`](./E02-docx-mvp.md) | DOCX MVP — Read, Write, Round-Trip (Sprint 1–3) | 9 | 89 |
| [`E03-xlsx-mvp.md`](./E03-xlsx-mvp.md) | XLSX MVP — Read, Write, Stream (Sprint 4–5) | 10 | 71 |
| [`E04-pptx-mvp.md`](./E04-pptx-mvp.md) | PPTX MVP — Read, Write (Sprint 6–7) | 8 | 63 |
| [`E05-ci-cd.md`](./E05-ci-cd.md) | CI/CD & Quality Infrastructure | 6 | 21 |
| [`E06-documentation.md`](./E06-documentation.md) | Documentation & Examples (Sprint 8) | 4 | 13 |
| [`E07-cli.md`](./E07-cli.md) | CLI Tooling (Sprint 8) | 2 | 8 |
| [`E08-security.md`](./E08-security.md) | Security & Hardening | 7 | 21 |
| [`99-backlog-future.md`](./99-backlog-future.md) | Post-MVP & future work | 4 | 55 |
| [`definition-of-done.md`](./definition-of-done.md) | Global DoD + GitHub issue templates | — | — |

**Total komitmen v0.5 (MVP):** 320 story points / ~16 minggu (1 dev) atau ~8 minggu (2 dev paralel).

## Cara Membaca

1. Mulai dari [`00-conventions.md`](./00-conventions.md) untuk memahami konvensi.
2. Sprint 0 wajib dikerjakan dulu — lihat [`E01-opc-foundation.md`](./E01-opc-foundation.md), [`E05-ci-cd.md`](./E05-ci-cd.md) (sebagian), [`E08-security.md`](./E08-security.md) (sebagian).
3. Setelah itu, pilih path: DOCX-first (E02), XLSX-first (E03), atau PPTX-first (E04). Rekomendasi: **DOCX-first** karena fondasi sudah ada.

## Penomoran Ticket

| Range | Domain |
|---|---|
| OFFICE-001..099 | OPC Foundation (E01) |
| OFFICE-101..199 | DOCX (E02) |
| OFFICE-201..299 | XLSX (E03) |
| OFFICE-301..399 | PPTX (E04) |
| OFFICE-501..599 | CI/CD (E05) |
| OFFICE-601..699 | Documentation (E06) |
| OFFICE-701..799 | CLI (E07) |
| OFFICE-801..899 | Security (E08) |
| OFFICE-901..999 | Future / Backlog |

## Cross-Reference

- Source of truth produk: [`../PRD.md`](../PRD.md)
- Index ringkas (lama, akan di-deprecate): [`../BACKLOG.md`](../BACKLOG.md) — sekarang hanya index ke folder ini
