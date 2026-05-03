# Definition of Done & Issue Templates

## Definition of Done (Global)

Setiap ticket harus memenuhi:

1. **Code merged ke main** via PR yang di-review minimal 1 reviewer (atau self-merged untuk solo dev — tetap pakai PR untuk audit trail).
2. **Test hijau** — `go test -race ./...` clean di CI.
3. **`go vet ./...` clean** — no warnings.
4. **Coverage tidak turun** dari baseline (≥ 80% internal, ≥ 70% public).
5. **Documentation updated** — godoc comments untuk public API, README/PRD update jika scope berubah.
6. **CHANGELOG entry** — kategori sesuai (Added/Changed/Fixed/Security).
7. **Backwards compatibility** — pre-1.0 boleh breaking; sesudah 1.0 wajib semver.
8. **Acceptance criteria semua checked** atau dengan justifikasi tertulis di PR jika dikecualikan.

## Per-Ticket-Type Checklist

### Story / Task
- [ ] Semua AC dicentang
- [ ] Unit test untuk happy path + edge cases
- [ ] Godoc comment untuk public API
- [ ] CHANGELOG.md updated
- [ ] PR linked ke ticket ID di title (`OFFICE-XXX:`)

### Bug
- [ ] Regression test ditulis SEBELUM fix (TDD)
- [ ] Root cause didokumentasikan di PR description
- [ ] Test case existing yang related tidak hilang

### Spike
- [ ] Time-box dihormati (max 2 hari)
- [ ] Output: dokumen tertulis di `docs/`
- [ ] Follow-up ticket dibuat berdasarkan finding

### Security ticket (E08)
- [ ] Threat model updated jika ada vector baru
- [ ] PoC test case (negative test) ditambahkan
- [ ] Tidak ada secret di test fixture

## GitHub Issue Templates

Untuk konsistensi tracking, berikut template untuk file `.github/ISSUE_TEMPLATE/`:

### `.github/ISSUE_TEMPLATE/feature.yml`

```yaml
name: Feature Request
description: Propose a new feature or enhancement
title: "[FEATURE] "
labels: ["enhancement"]
body:
  - type: input
    id: ticket
    attributes:
      label: Ticket ID
      placeholder: "OFFICE-XXX"
    validations:
      required: false
  - type: dropdown
    id: type
    attributes:
      label: Type
      options:
        - Story
        - Task
        - Spike
    validations:
      required: true
  - type: dropdown
    id: priority
    attributes:
      label: Priority
      options:
        - P0 (Blocker)
        - P1 (Critical)
        - P2 (Major)
        - P3 (Minor)
    validations:
      required: true
  - type: textarea
    id: user-story
    attributes:
      label: User Story
      description: "As a... I want... so that..."
      placeholder: "As a backend developer, I want to ..."
    validations:
      required: true
  - type: textarea
    id: ac
    attributes:
      label: Acceptance Criteria
      description: Use checklist format
      value: |
        - [ ] AC1: ...
        - [ ] AC2: ...
    validations:
      required: true
  - type: textarea
    id: notes
    attributes:
      label: Technical Notes
      description: Architecture, dependencies, risks
```

### `.github/ISSUE_TEMPLATE/bug.yml`

```yaml
name: Bug Report
description: Report a bug or unexpected behavior
title: "[BUG] "
labels: ["bug"]
body:
  - type: textarea
    id: repro
    attributes:
      label: Reproduction Steps
      placeholder: |
        1. Open file ...
        2. Call function ...
        3. Observe ...
    validations:
      required: true
  - type: textarea
    id: expected
    attributes:
      label: Expected Behavior
    validations:
      required: true
  - type: textarea
    id: actual
    attributes:
      label: Actual Behavior
    validations:
      required: true
  - type: input
    id: file
    attributes:
      label: Sample File
      description: Link to anonymized sample, or attach file
  - type: input
    id: version
    attributes:
      label: Library Version
      placeholder: "v0.1.0"
    validations:
      required: true
  - type: input
    id: go-version
    attributes:
      label: Go Version
      placeholder: "1.23.4"
    validations:
      required: true
  - type: dropdown
    id: os
    attributes:
      label: OS
      options:
        - Linux
        - macOS
        - Windows
        - Other
```

### `.github/ISSUE_TEMPLATE/security.yml`

> **Catatan:** Security issue **JANGAN** di-report via public GitHub issue. Lihat `SECURITY.md` untuk private disclosure channel.

```yaml
name: Security Disclosure (Public)
description: Only for security issues that are already public/acknowledged
title: "[SECURITY] "
labels: ["security"]
body:
  - type: markdown
    attributes:
      value: |
        ⚠️ **DO NOT** report unpatched vulnerabilities here. Use private channel via SECURITY.md.
  - type: input
    id: cve
    attributes:
      label: CVE ID (if assigned)
  - type: textarea
    id: description
    attributes:
      label: Description
    validations:
      required: true
```

## Pull Request Template

`.github/pull_request_template.md`:

```markdown
## Summary
<!-- Apa yang berubah dan kenapa -->

## Ticket
Closes OFFICE-XXX

## Acceptance Criteria
<!-- Centang AC yang relevan dari ticket -->
- [ ] AC1: ...
- [ ] AC2: ...

## Test Plan
- [ ] Unit test ditambahkan/diupdate
- [ ] `go test -race ./...` hijau
- [ ] `go vet ./...` clean
- [ ] Compatibility test pass (untuk writer changes)

## Breaking Changes
<!-- "None" atau jelaskan -->

## CHANGELOG
<!-- Salin entry yang akan masuk CHANGELOG.md -->

## Reviewer Checklist
- [ ] AC semua tercapai
- [ ] No new dependencies di go.mod
- [ ] Doc-comment public API ditulis
- [ ] Memory budget tidak melonjak (lihat benchmark)
```

## Branching Convention

- `main` — protected, hanya merge via PR
- `feature/OFFICE-XXX-short-description` — feature branch
- `bug/OFFICE-XXX-short-description` — bug fix
- `release/v0.X.Y` — release branch
- `hotfix/OFFICE-XXX-...` — urgent fix off main

## Commit Convention

Format: `<type>(<scope>): <subject>`

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `chore`, `ci`, `build`.

Scopes: `docx`, `xlsx`, `pptx`, `ooxml`, `wml`, `sml`, `pml`, `dml`, `cli`, `ci`, `docs`.

Examples:
```
feat(docx): add paragraph alignment support (OFFICE-102)
fix(ooxml): reject path traversal in ResolveTarget (OFFICE-001)
test(xlsx): add streaming iterator benchmark (OFFICE-203)
docs(prd): update tier matrix after MVP review
```

---

**Navigasi:** [← Backlog Future](./99-backlog-future.md) | [Index](./README.md)
