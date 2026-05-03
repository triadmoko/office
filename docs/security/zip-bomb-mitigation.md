# Zip-Bomb Mitigation in github.com/triadmoko/office

## Background

OOXML files (.docx, .xlsx, .pptx) are ZIP archives. A malicious actor can craft a ZIP file where
a small compressed payload decompresses to gigabytes of data — a classic zip-bomb attack. Without
protection, a single `io.ReadAll` on an OOXML part could exhaust all available memory and crash
the process.

---

## Go stdlib `archive/zip` Behavior

### Lazy Decompression

`zip.NewReader(r, size)` reads only the ZIP central directory (the end-of-archive metadata). It
does **not** decompress any entry content at open time. Memory usage at this stage is proportional
to the number of entries, not their uncompressed sizes.

`zip.File.Open()` returns an `io.ReadCloser` that decompresses lazily on each `Read()` call.
This means the vulnerability is triggered only when code reads the entry content — for example,
via `io.ReadAll(rc)`.

### `UncompressedSize64` Header

Each ZIP central directory entry contains `UncompressedSize64`, which reports the expected
uncompressed size. **This field is not authenticated**: a malicious ZIP can report `0` while
actually expanding to gigabytes. It can also report an honest size that is legitimately large.

Pre-checking `UncompressedSize64` provides a fast, low-cost guard against straightforward bombs,
but is insufficient alone.

### No Built-in Limits

Go's `archive/zip` has no built-in protection against zip bombs. There is no `MaxUncompressedSize`
option. Callers are responsible for enforcing limits.

---

## Vulnerability in Current `ooxml.Open()`

The existing implementation contains an unbounded read in `Open()`:

```go
// internal/ooxml/package.go (before OFFICE-008)
data, err := io.ReadAll(rc) // rc is the [Content_Types].xml entry — no size limit
```

A malicious OOXML file with a compressed `[Content_Types].xml` that expands to 10 GiB would
cause the process to allocate 10 GiB of memory before returning an error (or OOM-crashing).
The same vulnerability exists in `RootRelationships()` and any code that calls `ReadFile()`.

---

## Mitigation Strategies

### Strategy 1: Per-Entry Limit via `io.LimitReader`

Wrap each `rc` with `io.LimitReader(rc, maxBytes)`. If the entry exceeds `maxBytes`, `Read()`
returns `io.EOF` after `maxBytes` bytes. The caller must detect the truncation and return an error.

```go
const MaxPartBytes = 256 << 20 // 256 MiB

rc, err := f.Open()
limited := io.LimitReader(rc, MaxPartBytes+1)
data, err := io.ReadAll(limited)
if int64(len(data)) > MaxPartBytes {
    return nil, ErrPackageTooLarge
}
```

Using `MaxPartBytes+1` ensures that an exactly-at-limit read is not rejected, but a one-byte
overage is caught.

A better approach is a `limitedReadCloser` that returns `ErrPackageTooLarge` directly from `Read()`:

```go
type limitedReadCloser struct {
    rc    io.ReadCloser
    read  int64
    limit int64
}

func (l *limitedReadCloser) Read(p []byte) (int, error) {
    n, err := l.rc.Read(p)
    l.read += int64(n)
    if l.read > l.limit {
        return n, ErrPackageTooLarge
    }
    return n, err
}
```

This surfaces the error immediately from the first overflowing `Read()` call.

### Strategy 2: Total Package Size Guard

Track the sum of `UncompressedSize64` across all entries as a pre-check:

```go
var totalBytes int64
for _, f := range z.File {
    totalBytes += int64(f.UncompressedSize64)
    if totalBytes > MaxBytes {
        return nil, ErrPackageTooLarge
    }
}
```

Since `UncompressedSize64` can be spoofed, this is a fast heuristic, not a guarantee. The
per-entry `limitedReadCloser` provides the hard guarantee at read time.

### Strategy 3: Entry Count Limit

Reject archives with an excessive number of ZIP entries before any content is read:

```go
if len(z.File) > MaxParts {
    return nil, ErrTooManyParts
}
```

This protects against zip-slip-style attacks and pathological archives with millions of tiny entries
that exhaust file descriptor limits or parse time.

---

## Recommended Default Limits

| Limit | Default | Rationale |
|-------|---------|-----------|
| `MaxBytes` | 1 GiB (`1 << 30`) | Realistic upper bound for Office documents |
| `MaxParts` | 10 000 | ECMA-376 documents rarely exceed a few hundred parts |
| `MaxPartBytes` | 256 MiB (`256 << 20`) | Largest realistic single part (embedded media) |

All limits can be overridden via `OpenOptions` for legitimate use cases (e.g., large embedded video).
Setting a limit to `0` disables it.

---

## Implementation Reference

The mitigations above are implemented in `internal/ooxml/package.go` via:

- `OpenOptions` struct — configurable limits
- `OpenWithOptions(r, size, opts)` — enforces limits at open time
- `Open(r, size)` — calls `OpenWithOptions` with the default limits above
- `limitedReadCloser` — enforces `MaxPartBytes` at read time in `OpenReader()`

See also: `ErrPackageTooLarge`, `ErrTooManyParts` in `internal/ooxml/errors.go`.

---

## References

- [ECMA-376 Part 2 §13 — Open Packaging Conventions](https://ecma-international.org/publications-and-standards/standards/ecma-376/)
- [Go `archive/zip` package source](https://pkg.go.dev/archive/zip)
- [CWE-409: Improper Handling of Highly Compressed Data](https://cwe.mitre.org/data/definitions/409.html)
