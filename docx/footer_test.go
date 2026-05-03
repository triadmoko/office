package docx

import (
	"strings"
	"testing"
)

func TestParseFooterLayout(t *testing.T) {
	tests := []struct {
		in   string
		want string // kinds: T=text, P=PAGE, N=NUMPAGES
	}{
		{"No. {{PAGE}}", "T:No. |P"},
		{"{{PAGE}}", "P"},
		{"Page {{PAGE}} of {{NUMPAGES}}", "T:Page |P|T: of |N"},
		{"a{{{{PAGE}}", "T:a|T:{{|P"},
		{"{{unknown}}{{PAGE}}", "T:{{|T:unknown}}|P"},
	}
	for _, tc := range tests {
		segs := parseFooterLayout(tc.in)
		var parts []string
		for _, s := range segs {
			switch s.kind {
			case footerSegText:
				parts = append(parts, "T:"+s.text)
			case footerSegPage:
				parts = append(parts, "P")
			case footerSegNumPages:
				parts = append(parts, "N")
			}
		}
		got := strings.Join(parts, "|")
		if got != tc.want {
			t.Errorf("parseFooterLayout(%q) => %q want %q", tc.in, got, tc.want)
		}
	}
}

func TestMarshalFooterPageXMLCustomPrefix(t *testing.T) {
	xml := string(marshalFooterPageXML("No. {{PAGE}}"))
	if !strings.Contains(xml, ">No. </w:t>") && !strings.Contains(xml, ">No.</w:t>") {
		// preserve space may attach to token
		if !strings.Contains(xml, "No.") {
			t.Fatalf("missing literal No.: %s", xml)
		}
	}
	if !strings.Contains(xml, "instrText") || !strings.Contains(xml, " PAGE ") {
		t.Fatalf("missing PAGE field: %s", xml)
	}
}

func TestMarshalFooterPageXMLDefault(t *testing.T) {
	xml := string(marshalFooterPageXML(""))
	if !strings.Contains(xml, "Hal.") || !strings.Contains(xml, " PAGE ") {
		t.Fatalf("default footer: %s", xml)
	}
}
