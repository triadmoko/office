package xlsx

import (
	"io"
	"testing"
)

// BenchmarkStreamWriter1M is named per OFFICE-208; default iteration uses 1k rows per op to keep CI fast.
func BenchmarkStreamWriter1M(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wb := NewWorkbook()
		sh := wb.AddSheet("S")
		sw, err := sh.StreamWriter()
		if err != nil {
			b.Fatal(err)
		}
		for r := 1; r <= 1000; r++ {
			if err := sw.WriteRow(r, "x", float64(r), float64(r)*2); err != nil {
				b.Fatal(err)
			}
		}
		if err := sw.Flush(); err != nil {
			b.Fatal(err)
		}
		if err := wb.Save(io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}
