# office

Open-source **Office Open XML (OOXML)** tooling in **pure Go** (standard library only: no third-party modules in `go.mod`). The goal is to read and write formats such as **`.docx`**, **`.xlsx`**, and **`.pptx`** by working directly with the OPC ZIP container and XML parts.

## Status

| Format | Package   | Current capability |
|--------|-----------|--------------------|
| Word   | `docx`    | Open packages, read plain text from `w:t`, write a minimal one-paragraph document, basic round-trip |
| Excel  | `xlsx`    | Open and validate a workbook package; `Write` not implemented yet |
| PowerPoint | `pptx` | Open and validate a presentation package; `Write` not implemented yet |

Shared ZIP / `[Content_Types].xml` / relationship helpers live in `internal/ooxml` (not a stable public API—import the top-level format packages instead).

## Requirements

- Go **1.23** or newer

## Install

```bash
go get github.com/triadmoko/office@latest
```

## Library usage

### DOCX: minimal write and read

```go
package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/triadmoko/office/docx"
)

func main() {
	var buf bytes.Buffer
	if err := docx.WriteMinimal(&buf, "Hello, OOXML"); err != nil {
		log.Fatal(err)
	}
	d, err := docx.Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		log.Fatal(err)
	}
	text, err := d.PlainText()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(text) // Hello, OOXML
}
```

### XLSX / PPTX: validate package

```go
wb, err := xlsx.Open(ra, size)
if err != nil {
	// not a valid .xlsx main part / OPC package
}
_ = wb.MainPart() // e.g. /xl/workbook.xml
```

```go
pr, err := pptx.Open(ra, size)
if err != nil {
	// not a valid .pptx main part / OPC package
}
_ = pr.MainPart() // e.g. /ppt/presentation.xml
```

## CLI (`cmd/office`)

Build and run:

```bash
go build -o office ./cmd/office
./office -write-docx out.docx -text "Hello from office"
```

With no `-write-docx` flag, the binary prints a short usage line.

## Project layout

```
.
├── .github/          # issue forms, PR template, CI, dependabot
├── cmd/office/       # small demo CLI
├── docx/             # WordprocessingML
├── docs/             # architecture & format documentation
├── pptx/             # PresentationML (skeleton)
├── xlsx/             # SpreadsheetML (skeleton)
└── internal/ooxml/   # OPC: ZIP, content types, relationships
```

## Documentation

- [docs/README.md](docs/README.md) — documentation index
- [docs/architecture.md](docs/architecture.md) — OPC layers and package boundaries
- [docs/formats.md](docs/formats.md) — `.docx` / `.xlsx` / `.pptx` support matrix
- [docs/development.md](docs/development.md) — local commands and CI

## Tests

```bash
go test ./...
```

## Roadmap

Higher-level features (full styles, tables, formulas, charts, slide masters, strict conformance with Word/Excel/PowerPoint) will be added incrementally while keeping the **stdlib-only** constraint unless the project explicitly changes that policy.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issue and pull request templates live under [`.github/`](.github/).

## Code of conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## Security

See [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE). SPDX-License-Identifier: MIT.
