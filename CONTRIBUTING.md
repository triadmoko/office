# Contributing to `office`

Thank you for considering a contribution. This guide explains **how the repository is organized**, **what we optimize for**, and **how to get a change merged** with minimal back-and-forth.

If anything here is unclear or outdated, opening an issue (or a small PR to fix this file) is welcome.

---

## Table of contents

1. [Project scope and principles](#project-scope-and-principles)
2. [Ways to contribute](#ways-to-contribute)
3. [Development environment](#development-environment)
4. [Repository layout and import rules](#repository-layout-and-import-rules)
5. [Coding guidelines](#coding-guidelines)
6. [OOXML and interoperability notes](#ooxml-and-interoperability-notes)
7. [Testing](#testing)
8. [Documentation and examples](#documentation-and-examples)
9. [Git workflow: branches, commits, pull requests](#git-workflow-branches-commits-pull-requests)
10. [Review expectations](#review-expectations)
11. [Security](#security)
12. [Releases and versioning](#releases-and-versioning)
13. [Community and conduct](#community-and-conduct)
14. [License](#license)
15. [Repository automation and docs](#repository-automation-and-docs)

### Repository automation and docs

- **GitHub Actions**: [`.github/workflows/ci.yml`](.github/workflows/ci.yml)
- **Issue templates**: [`.github/ISSUE_TEMPLATE/`](.github/ISSUE_TEMPLATE/)
- **PR template**: [`.github/pull_request_template.md`](.github/pull_request_template.md)
- **Dependabot (Actions)**: [`.github/dependabot.yml`](.github/dependabot.yml)
- **Extra docs**: [`docs/`](docs/) (architecture, formats, development)

---

## Project scope and principles

### Goals

- Provide **Go libraries** for **Office Open XML** (`.docx`, `.xlsx`, `.pptx`) built on **OPC**: ZIP containers, XML parts, and relationships.
- Keep the default policy: **only the Go standard library** in [`go.mod`](go.mod) (no third-party modules) unless the project explicitly revises that policy in a tracked decision (issue + maintainer agreement).

### Non-goals (for now)

- Replacing full desktop Office or every ECMA-376 edge case on day one.
- Binding to native Office binaries or COM automation.
- Shipping large binary corpora in the repo without justification.

### Quality bar

- Prefer **small, reviewable PRs** with **tests** over sweeping refactors.
- Prefer **correctness for a documented subset** over pretending to support every feature flag Word/Excel/PowerPoint accepts.

---

## Ways to contribute

You can help without writing production code:

- **Report bugs** with reproducible inputs (small files, hex dumps avoided when a `.docx` zip suffices).
- **Improve docs** ([`README.md`](README.md), package doc comments, this file).
- **Add tests** that lock in behavior before a larger refactor.
- **Review PRs**: typos, API clarity, test gaps, and spec references are all valuable.

For code changes, read [Repository layout](#repository-layout-and-import-rules) first so new code lands in the right package.

---

## Development environment

### Required

- **Go 1.23 or newer** (see [`go.mod`](go.mod)).

### Recommended checks (from repository root)

```bash
gofmt -w .
go test ./...
go vet ./...
go build -o office ./cmd/office
```

Optional but useful before pushing:

```bash
go test -race ./...
```

### Editor

Any editor is fine. If you use VS Code / Cursor, enabling format-on-save with the official Go extension helps keep `gofmt` consistent.

---

## Repository layout and import rules

### Layout

| Area | Path | Purpose |
|------|------|--------|
| Shared OPC / ZIP / `[Content_Types].xml` / `.rels` | [`internal/ooxml`](internal/ooxml) | Building blocks for all formats. |
| WordprocessingML | [`docx`](docx) | Public API for `.docx`. |
| SpreadsheetML | [`xlsx`](xlsx) | Public API for `.xlsx`. |
| PresentationML | [`pptx`](pptx) | Public API for `.pptx`. |
| CLI demo | [`cmd/office`](cmd/office) | Small stdlib-only demo binary. |

### Import rules

- **`internal/ooxml` must not import** `docx`, `xlsx`, or `pptx` (avoid cycles and layer violations).
- **Format packages** (`docx`, `xlsx`, `pptx`) **may import** `internal/ooxml`.
- Treat **`internal/ooxml` as internal**: it is not committed to semantic-versioning stability the same way as root-level public packages; still avoid breaking it gratuitously if other packages depend on it.

### Where to put new logic

- **Cross-format ZIP/XML/relationship behavior** → extend `internal/ooxml`.
- **Format-specific semantics** (e.g. workbook sheets, presentation slides) → the corresponding format package.
- **Demos and flags** → `cmd/office`, keeping the CLI thin.

---

## Coding guidelines

### General Go style

- Run **`gofmt`** on changed files; do not fight Go’s formatting conventions.
- Use **idiomatic error handling**: wrap with context where it helps (`fmt.Errorf("docx: …: %w", err)`), avoid silent swallowing.
- **Exported symbols** need **clear doc comments** starting with the name (`// Document …`).
- Keep **exported API surfaces small**: prefer unexported helpers until a use case needs stability.

### Dependencies

- **Do not add** new `require` entries to [`go.mod`](go.mod) for third-party libraries unless the project maintainers have agreed to change the stdlib-only policy.
- The standard library already provides `archive/zip`, `encoding/xml`, `bytes`, `io`, `path`/`strings`, etc.—prefer composition of those.

### Performance

- Avoid unnecessary full-buffer reads of large parts when streaming is possible.
- Add **benchmarks** only when you are optimizing a hot path or proving a regression; include what you measured in the PR text.

### API changes

- Breaking changes to exported APIs should be **called out explicitly** in the PR description and ideally discussed in an issue first.
- If you must deprecate something, use a **comment-based deprecation** and a migration path in the same or follow-up PR.

---

## OOXML and interoperability notes

- Real files from Microsoft Office can be **strict** about namespaces, ordering, and relationship targets. When adding writers, **test round-trips** and, when feasible, open output in Word/Excel/PowerPoint.
- Prefer **explicit namespace URIs** matching ECMA-376 / ISO 29500 over guessing `Local` element names without `Space`.
- When supporting only a **subset** of a feature, document limitations in:
  - the **package doc comment**, and/or
  - the **function doc comment** near the entry point (`Open`, `Write`, etc.).

---

## Testing

### Expectations

- New behavior should come with **tests** in the same package (`*_test.go`) or `testdata` where binary inputs are required.
- **Table-driven tests** are encouraged when they improve clarity; avoid tables so large they obscure intent.
- Name tests so failures read well: `TestOpen_rejects_truncated_zip`, not `Test1`.

### Commands

```bash
# all packages
go test ./...

# single package while iterating
go test ./docx -count=1 -v

# race detector (slower)
go test -race ./...
```

### Fixtures

- Keep **`testdata`** files **small** and **purposeful** (one concern per fixture when possible).
- If a test skips without a fixture, explain why (`t.Skip` message).
- Avoid committing **secrets**, **PII**, or **huge** documents.

### Coverage (optional)

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Use coverage to find obvious gaps; do not chase 100% at the expense of meaningful assertions.

---

## Documentation and examples

- Update [`README.md`](README.md) when user-visible behavior or installation instructions change.
- Use **`example_test.go`** (package `docx_test`, etc.) for examples that appear in pkg.go.dev; keep them **buildable** and **focused**.
- Link to **spec sections** (ECMA-376 / ISO 29500) in comments when behavior is non-obvious.

---

## Git workflow: branches, commits, pull requests

### Fork and branch

1. **Fork** the repository on GitHub (if you lack direct push access).
2. Create a **feature branch** off `main` (or the default branch):

   ```bash
   git checkout -b docx-fix-paragraph-spacing
   ```

### Branch naming (suggestions)

Use lowercase and hyphens: `docx/add-table-read`, `ooxml/fix-rels-resolve`, `docs/contributing-typo`.

### Commits

- Prefer **logical commits**: each commit should build and tests should pass when feasible.
- **Commit messages** in imperative mood work well: `Add xlsx sheet name parser`, `Fix rels target for root .rels`.
- If an issue exists, reference it: `Fixes #123` or `Refs #123` (use `Fixes` only when the PR fully closes the issue).

### Opening a PR

Include in the description:

1. **Motivation** — why is this change needed?
2. **What changed** — high-level bullet list.
3. **How to verify** — commands you ran (`go test ./...`, manual steps with a sample file).
4. **Risk / follow-ups** — known limitations or TODOs you intentionally left.

### PR size

- Smaller PRs merge faster. If work is large, consider **stacking** or **splitting**:
  - PR A: internal helper + tests  
  - PR B: wire helper into `docx` API  

If you must do a large change, a **draft PR** early helps align direction.

---

## Review expectations

- Maintainers may request **tests**, **naming tweaks**, or **API simplifications**—this is normal; iterate in the same PR when practical.
- **Force-pushes** are acceptable on your branch while the PR is open if you are cleaning history; say so in a comment after a rebase.
- If a PR stalls, a polite **ping after several days** is fine.

---

## Security

If you believe you found a **security vulnerability** (e.g. unbounded memory use on untrusted input, path traversal when unpacking ZIPs), please **do not** file a public issue with exploit details until maintainers can respond.

Instead, use **private GitHub Security Advisories** for this repository (if enabled), or contact maintainers through a **private channel** they publish in the README or org profile. If no channel exists yet, open a **minimal** public issue asking where to report security problems.

---

## Releases and versioning

Module path: **`github.com/triadmoko/office`**. Tags like **`v0.1.0`** follow Go module versioning.

- **Maintainers** cut releases; contributors do not need to bump versions in every PR.
- If your change is user-visible, mention in the PR whether it should appear in **release notes** (feature / fix / breaking).

---

## Community and conduct

- **Be respectful**, assume good intent, and stay on-topic.
- Critique **code and ideas**, not people.
- Harassment, discrimination, and sustained disruption are not acceptable.

### Code of conduct (short form)

This project aims for a **professional, welcoming** environment. If you see behavior that violates that spirit, report it to the maintainers.

---

## License

By contributing, you agree your contributions are licensed under the **same terms as the project**: the [MIT License](LICENSE).
