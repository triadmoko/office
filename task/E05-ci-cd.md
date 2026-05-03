# EPIC OFFICE-E05 — CI/CD & Quality Infrastructure

> **Goal:** GitHub Actions CI dengan multi-OS, fuzz, golden tests, compat matrix.
> **Sprint:** 0 (setup) + ongoing
> **Total points:** 21

## Daftar Ticket

| ID | Title | Type | Sprint | Points | Priority |
|---|---|---|---|---|---|
| [OFFICE-501](#office-501) | GitHub Actions: build & test matrix | Task | 0 | 3 | P1 |
| [OFFICE-502](#office-502) | GitHub Actions: fuzz nightly | Task | 0 | 3 | P2 |
| [OFFICE-503](#office-503) | Golden test infrastructure | Task | 0 | 5 | P1 |
| [OFFICE-504](#office-504) | Compatibility test harness | Task | 5 | 5 | P2 |
| [OFFICE-505](#office-505) | Pre-commit hooks (gofmt, govet) | Task | 0 | 2 | P3 |
| [OFFICE-506](#office-506) | Benchmark harness | Task | 5 | 3 | P2 |

---

## OFFICE-501

### [Task] GitHub Actions: build & test matrix

```
Type     : Task
Priority : P1
Points   : 3
Sprint   : 0
Epic     : OFFICE-E05
File     : .github/workflows/ci.yml (NEW)
```

#### Acceptance Criteria
- [ ] **AC1**: Trigger: push ke main, PR ke main, schedule nightly.
- [ ] **AC2**: Matrix: Go 1.23, 1.24 × ubuntu-latest, macos-latest, windows-latest = 6 jobs.
- [ ] **AC3**: Steps: checkout, setup-go, `go mod download` (no-op stdlib only), `go vet ./...`, `go test -race ./...`, `go test -cover ./...`.
- [ ] **AC4**: Coverage upload ke Codecov / GitHub artifact.
- [ ] **AC5**: Status badge ditambahkan ke README.

---

## OFFICE-502

### [Task] GitHub Actions: fuzz nightly

```
Type     : Task
Priority : P2
Points   : 3
Sprint   : 0
Epic     : OFFICE-E05
File     : .github/workflows/fuzz.yml (NEW)
Depends  : OFFICE-009
```

#### Acceptance Criteria
- [ ] **AC1**: Schedule: cron daily 03:00 UTC.
- [ ] **AC2**: Run: setiap fuzz target di `internal/ooxml`, `xlsx`, `docx`, `pptx` dengan `-fuzztime=5m`.
- [ ] **AC3**: Crash file → upload sebagai artifact + auto-create issue.
- [ ] **AC4**: Tidak fail PR jika fuzz menemukan crash baru — hanya open issue.

---

## OFFICE-503

### [Task] Golden test infrastructure

```
Type     : Task
Priority : P1
Points   : 5
Sprint   : 0
Epic     : OFFICE-E05
File     : internal/testutil/golden.go (NEW)
```

#### Acceptance Criteria
- [ ] **AC1**: Helper `AssertGoldenZip(t, gotBytes, "testdata/golden/foo.docx")` — bandingkan ZIP secara semantik (sort entries + diff XML kanonikal).
- [ ] **AC2**: `-update` flag untuk regenerate golden files: `go test ./... -update`.
- [ ] **AC3**: Diff output human-readable saat fail (per-part diff).
- [ ] **AC4**: Handle XML namespace prefix variation (e.g., `xmlns:w` urutan).
- [ ] **AC5**: Test sendiri: helper digunakan di minimal 3 test, semua hijau.

---

## OFFICE-504

### [Task] Compatibility test harness

```
Type     : Task
Priority : P2
Points   : 5
Sprint   : 5 (setelah ada writer 2 format)
Epic     : OFFICE-E05
File     : .github/workflows/compat.yml (NEW)
```

#### Acceptance Criteria
- [ ] **AC1**: Job di CI yang install LibreOffice headless di runner.
- [ ] **AC2**: Untuk setiap golden file generated, jalankan `soffice --headless --convert-to pdf` → expect exit 0 dan PDF non-empty.
- [ ] **AC3**: Untuk MS Office: tidak realistis di CI; substitute dengan fixture upload manual ke release notes (maintainer responsibility).
- [ ] **AC4**: Output: matrix table di README otomatis update via badge / generated section.

---

## OFFICE-505

### [Task] Pre-commit hooks (gofmt, govet, unused-import)

```
Type     : Task
Priority : P3
Points   : 2
Sprint   : 0
Epic     : OFFICE-E05
File     : .githooks/pre-commit (NEW)
```

#### Acceptance Criteria
- [ ] **AC1**: Hook menjalankan `gofmt -l` (fail jika ada output).
- [ ] **AC2**: `go vet ./...` clean.
- [ ] **AC3**: Doc README cara install: `git config core.hooksPath .githooks`.

---

## OFFICE-506

### [Task] Benchmark harness

```
Type     : Task
Priority : P2
Points   : 3
Sprint   : 5
Epic     : OFFICE-E05
File     : */bench_test.go (NEW)
```

#### Acceptance Criteria
- [ ] **AC1**: Benchmark targets:
  - `BenchmarkDocxOpen100p` — open 100-paragraph docx
  - `BenchmarkDocxWrite1000p` — write 1000-paragraph
  - `BenchmarkXlsxStreamRead100k` — iterate 100k row
  - `BenchmarkXlsxStreamWrite1M` — write 1M row
  - `BenchmarkPptxOpen20s` — open 20-slide
- [ ] **AC2**: CI nightly run benchmarks → upload to perf tracker (Bencher.dev gratis untuk OSS).
- [ ] **AC3**: Performance regression > 20% → CI alert (issue auto-created).

---

**Navigasi:** [← E04 PPTX](./E04-pptx-mvp.md) | [Index](./README.md) | [E06 Documentation →](./E06-documentation.md)
