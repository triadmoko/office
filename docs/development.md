# Development

## Prerequisites

- **Go 1.23+** (see `go.mod`).

## Common commands

```bash
gofmt -w .
go test ./...
go vet ./...
go build ./...
go build -o office ./cmd/office
```

Race detector (slower):

```bash
go test -race ./...
```

## Continuous integration

GitHub Actions workflow: [`.github/workflows/ci.yml`](../.github/workflows/ci.yml)

On each push and pull request to `main` / `master`, CI runs:

- `go mod tidy` (must produce no uncommitted `go.mod` / `go.sum` changes)
- `go vet ./...`
- `go test ./...`
- `go build ./...` and `go build ./cmd/office`

Matrix: **Go 1.23.x** and **stable** on **ubuntu-latest**.

## Fuzzing (OPC / OOXML)

`internal/ooxml` defines fuzz targets (`FuzzParseContentTypes`, `FuzzParseRelationships`, `FuzzResolveTarget`) in [`internal/ooxml/fuzz_test.go`](../internal/ooxml/fuzz_test.go). Quick local example:

```bash
go test ./internal/ooxml -fuzz=FuzzParseContentTypes -fuzztime=30s
```

Workflow GitHub Actions untuk fuzz **nightly** direncanakan di tiket **OFFICE-502** ([`task/E05-ci-cd.md`](../task/E05-ci-cd.md)).

## Large packages (`OpenWithOptions`)

Untuk file ZIP besar atau batas kustom, gunakan `ooxml.OpenWithOptions` (lihat [`docs/security/zip-bomb-mitigation.md`](security/zip-bomb-mitigation.md)). `ooxml.Open` memakai default aman (total, jumlah part, dan ukuran per-part).

## Dependency policy

**No third-party modules** in `go.mod` unless the project explicitly changes policy (see [CONTRIBUTING.md](../CONTRIBUTING.md)).

## Issue and PR templates

- Issue forms: [`.github/ISSUE_TEMPLATE/`](../.github/ISSUE_TEMPLATE/)
- Pull request template: [`.github/pull_request_template.md`](../.github/pull_request_template.md)

## Dependabot

[`.github/dependabot.yml`](../.github/dependabot.yml) updates **GitHub Actions** dependencies on a monthly schedule.

## Code of conduct / security

- [CODE_OF_CONDUCT.md](../CODE_OF_CONDUCT.md)
- [SECURITY.md](../SECURITY.md)
