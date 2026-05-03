# BACKLOG — Pure-Go Office Library

> **Backlog detail telah dipecah per-epic ke direktori [`task/`](./task/) untuk kemudahan navigasi.**

## Quick Links

- 📋 [`task/README.md`](./task/README.md) — Index lengkap
- 📐 [`task/00-conventions.md`](./task/00-conventions.md) — Konvensi & sprint plan
- ✅ [`task/definition-of-done.md`](./task/definition-of-done.md) — DoD & GitHub templates

## Epic Index

| Epic | File | Sprint | Points |
|---|---|---|---|
| **E01** OPC Foundation Hardening | [`task/E01-opc-foundation.md`](./task/E01-opc-foundation.md) | 0 | 34 |
| **E02** DOCX MVP | [`task/E02-docx-mvp.md`](./task/E02-docx-mvp.md) | 1–3 | 89 |
| **E03** XLSX MVP | [`task/E03-xlsx-mvp.md`](./task/E03-xlsx-mvp.md) | 4–5 | 71 |
| **E04** PPTX MVP | [`task/E04-pptx-mvp.md`](./task/E04-pptx-mvp.md) | 6–7 | 63 |
| **E05** CI/CD & Quality | [`task/E05-ci-cd.md`](./task/E05-ci-cd.md) | 0–8 | 21 |
| **E06** Documentation & Examples | [`task/E06-documentation.md`](./task/E06-documentation.md) | 8 | 13 |
| **E07** CLI Tooling | [`task/E07-cli.md`](./task/E07-cli.md) | 8 | 8 |
| **E08** Security & Hardening | [`task/E08-security.md`](./task/E08-security.md) | 0, 2, 4 | 21 |
| **E99** Post-MVP / Future | [`task/99-backlog-future.md`](./task/99-backlog-future.md) | Backlog | 55+ |

**Total komitmen v0.5 (MVP):** ~320 story points, 8 sprint = ~16 minggu (1 dev) atau ~8 minggu (2 dev paralel).

## Source of Truth

- 📘 Produk vision & feature inventory: [`PRD.md`](./PRD.md)
- 📋 Backlog detail per-epic: [`task/`](./task/)

## Sprint 0 Wajib (Blocker)

Sebelum mulai feature work apa pun:

1. [`OFFICE-001`](./task/E01-opc-foundation.md#office-001) — Bug fix `ResolveTarget` ".." traversal
2. [`OFFICE-003`](./task/E01-opc-foundation.md#office-003) — Package Writer (blocker semua write)
3. [`OFFICE-004`](./task/E01-opc-foundation.md#office-004) — `internal/xmlwriter`
4. [`OFFICE-006`](./task/E01-opc-foundation.md#office-006) — `internal/opcprops`
5. [`OFFICE-501`](./task/E05-ci-cd.md#office-501) — GitHub Actions CI
6. [`OFFICE-503`](./task/E05-ci-cd.md#office-503) — Golden test infrastructure
7. [`OFFICE-802`](./task/E08-security.md#office-802) — XML billion-laughs guard
