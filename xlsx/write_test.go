package xlsx

import (
	"bytes"
	"testing"
)

func TestNewWorkbookSaveAndRoundTrip(t *testing.T) {
	wb := NewWorkbook()
	sh := wb.AddSheet("S1")
	if err := sh.SetCell("A1", "hello"); err != nil {
		t.Fatal(err)
	}
	if err := sh.SetCell("A2", 3.14); err != nil {
		t.Fatal(err)
	}
	if err := sh.SetFormula("A3", "=SUM(A1:A2)"); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := wb.Save(&buf); err != nil {
		t.Fatal(err)
	}
	rb, err := Open(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := rb.Save(&out); err != nil {
		t.Fatal(err)
	}
}
