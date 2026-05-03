package xlsx

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/triadmoko/office/internal/ooxml"
	"github.com/triadmoko/office/internal/sml"
)

// streamSheetBuf buffers streamed sheetData XML and optional hyperlink rels.
type streamSheetBuf struct {
	fragment bytes.Buffer
	rels     ooxml.Relationships
	hlinkSeq int
	closed   bool
}

// StreamWriter writes rows in streaming mode (inline strings, low memory).
type StreamWriter struct {
	sh *Sheet
}

// StreamWriter starts streaming output for this sheet. Random SetCell is disabled afterward.
func (s *Sheet) StreamWriter() (*StreamWriter, error) {
	ws, err := s.mustWrite()
	if err != nil {
		return nil, err
	}
	if ws.stream != nil {
		return nil, fmt.Errorf("xlsx: StreamWriter already active")
	}
	ws.stream = &streamSheetBuf{}
	ws.stream.fragment.WriteString(`<sheetData>`)
	ws.locked = true
	return &StreamWriter{sh: s}, nil
}

// WriteRow writes one row with 1-based column order starting at column 1.
func (sw *StreamWriter) WriteRow(row int, values ...any) error {
	if sw == nil || sw.sh == nil || sw.sh.ws == nil || sw.sh.ws.stream == nil {
		return fmt.Errorf("xlsx: invalid StreamWriter")
	}
	b := &sw.sh.ws.stream.fragment
	b.WriteString(fmt.Sprintf(`<row r="%d">`, row))
	for i, v := range values {
		addr := sml.IndexesToCellRef(i+1, row)
		b.WriteString(`<c r="`)
		b.WriteString(addr)
		b.WriteString(`" t="inlineStr"><is><t>`)
		xmlEscapeStringBuilder(b, streamCellString(v))
		b.WriteString(`</t></is></c>`)
	}
	b.WriteString(`</row>`)
	return nil
}

func streamCellString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case int:
		return strconv.Itoa(x)
	case bool:
		if x {
			return "TRUE"
		}
		return "FALSE"
	case time.Time:
		return x.Format(time.RFC3339)
	default:
		return fmt.Sprint(x)
	}
}

func xmlEscapeStringBuilder(b *bytes.Buffer, s string) {
	_ = xml.EscapeText(b, []byte(s))
}

// Flush closes the sheetData element. Call before Save.
func (sw *StreamWriter) Flush() error {
	if sw == nil || sw.sh == nil || sw.sh.ws == nil || sw.sh.ws.stream == nil {
		return nil
	}
	if !sw.sh.ws.stream.closed {
		sw.sh.ws.stream.fragment.WriteString(`</sheetData>`)
		sw.sh.ws.stream.closed = true
	}
	return nil
}
