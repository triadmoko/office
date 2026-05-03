package sml

import (
	"strings"
	"testing"
)

func TestParseSharedStringsPlainAndRich(t *testing.T) {
	const xml = `<?xml version="1.0" encoding="UTF-8"?>
<sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" count="3" uniqueCount="3">
<si><t>hello</t></si>
<si><r><t>ab</t></r><r><t>cd</t></r></si>
<si><t xml:space="preserve">  spaced  </t></si>
</sst>`
	out, err := ParseSharedStrings(strings.NewReader(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 {
		t.Fatalf("len %d %#v", len(out), out)
	}
	if out[0] != "hello" {
		t.Fatalf("0: %q", out[0])
	}
	if out[1] != "abcd" {
		t.Fatalf("1: %q", out[1])
	}
	if out[2] != "  spaced  " {
		t.Fatalf("2: %q", out[2])
	}
}

func TestParseSharedStringsLargeStreaming(t *testing.T) {
	if testing.Short() {
		t.Skip("large shared strings")
	}
	var b strings.Builder
	b.WriteString(`<?xml version="1.0"?><sst xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">`)
	n := 100_000
	for i := 0; i < n; i++ {
		b.WriteString("<si><t>")
		b.WriteString(strings.Repeat("x", 8))
		b.WriteString("</t></si>")
	}
	b.WriteString("</sst>")
	out, err := ParseSharedStrings(strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != n {
		t.Fatalf("len %d", len(out))
	}
}
