package docx

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteOpenPlainTextRoundTrip(t *testing.T) {
	const want = "café <office> & \"quotes\""
	var buf bytes.Buffer
	if err := WriteMinimal(&buf, want); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	d, err := Open(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if d.MainPart() != "/word/document.xml" {
		t.Fatalf("main part: %q", d.MainPart())
	}
	got, err := d.PlainText()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("plain text: got %q want %q", got, want)
	}
	var out bytes.Buffer
	if err := d.Write(&out); err != nil {
		t.Fatal(err)
	}
	d2, err := Open(bytes.NewReader(out.Bytes()), int64(out.Len()))
	if err != nil {
		t.Fatal(err)
	}
	got2, err := d2.PlainText()
	if err != nil {
		t.Fatal(err)
	}
	if got2 != want {
		t.Fatalf("round-trip write: got %q", got2)
	}
}

func TestReadTestdataFixture(t *testing.T) {
	path := filepath.Join("testdata", "minimal.docx")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Skip("testdata fixture:", err)
	}
	d, err := Open(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	txt, err := d.PlainText()
	if err != nil {
		t.Fatal(err)
	}
	if txt != "fixture" {
		t.Fatalf("fixture text: got %q", txt)
	}
}
