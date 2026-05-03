// Package xmlwriter provides a streaming XML writer that preserves namespace prefixes.
//
// Unlike [encoding/xml], this writer lets callers declare explicit prefixes (e.g. "w", "r")
// and guarantees those exact prefixes appear in the output — which is required by
// Microsoft Office applications when opening OOXML files.
//
// Example:
//
//	w := xmlwriter.New(out)
//	w.DeclareNamespace("w", "http://schemas.openxmlformats.org/wordprocessingml/2006/main")
//	w.StartElement(xml.Name{Space: "http://schemas.openxmlformats.org/wordprocessingml/2006/main", Local: "document"}, nil)
//	w.EndElement()
//	w.Close()
package xmlwriter

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

// Writer is a streaming, append-only XML writer with explicit namespace prefix control.
type Writer struct {
	w           io.Writer
	nsMap       map[string]string // namespace URI → prefix
	stack       []frame           // open element stack
	pendingOpen bool              // true when StartElement emitted tag but not yet '>'
	headerDone  bool              // true after XML declaration written
	err         error             // sticky error
}

type frame struct {
	qualified string // qualified tag name (e.g. "w:document")
}

// New creates a Writer that writes to w.
func New(w io.Writer) *Writer {
	return &Writer{
		w:     w,
		nsMap: make(map[string]string),
	}
}

// DeclareNamespace registers a namespace prefix mapping. Must be called before StartElement.
func (w *Writer) DeclareNamespace(prefix, uri string) {
	w.nsMap[uri] = prefix
}

// StartElement writes an opening tag. The actual '>' is deferred until the next call,
// allowing EndElement to detect a childless element and emit '/>' instead.
func (w *Writer) StartElement(name xml.Name, attrs []xml.Attr) error {
	if w.err != nil {
		return w.err
	}
	if err := w.flushHeader(); err != nil {
		return err
	}
	if err := w.flushPending(false); err != nil {
		return err
	}
	qname := w.qualifiedName(name)
	var sb strings.Builder
	sb.WriteByte('<')
	sb.WriteString(qname)
	for _, a := range attrs {
		sb.WriteByte(' ')
		sb.WriteString(w.qualifiedName(a.Name))
		sb.WriteString(`="`)
		sb.WriteString(escapeAttr(a.Value))
		sb.WriteByte('"')
	}
	if _, err := io.WriteString(w.w, sb.String()); err != nil {
		w.err = err
		return err
	}
	w.stack = append(w.stack, frame{qualified: qname})
	w.pendingOpen = true
	return nil
}

// EndElement closes the most recently opened element.
// If no child content was written, emits a self-closing tag (e.g. <w:tab/>).
func (w *Writer) EndElement() error {
	if w.err != nil {
		return w.err
	}
	if len(w.stack) == 0 {
		w.err = errors.New("xmlwriter: EndElement called with no open element")
		return w.err
	}
	top := w.stack[len(w.stack)-1]
	w.stack = w.stack[:len(w.stack)-1]
	var s string
	if w.pendingOpen {
		// No children written — emit self-closing tag.
		s = "/>"
		w.pendingOpen = false
	} else {
		s = "</" + top.qualified + ">"
	}
	if _, err := io.WriteString(w.w, s); err != nil {
		w.err = err
		return err
	}
	return nil
}

// CharData writes escaped text content. Flushes any pending '>' first.
func (w *Writer) CharData(s string) error {
	if w.err != nil {
		return w.err
	}
	if err := w.flushHeader(); err != nil {
		return err
	}
	if err := w.flushPending(false); err != nil {
		return err
	}
	if _, err := io.WriteString(w.w, escapeCharData(s)); err != nil {
		w.err = err
		return err
	}
	return nil
}

// Close validates that all opened elements have been closed.
func (w *Writer) Close() error {
	if w.err != nil {
		return w.err
	}
	if len(w.stack) != 0 {
		w.err = errors.New("xmlwriter: unclosed elements at Close")
		return w.err
	}
	return nil
}

// flushHeader emits the XML declaration on first write.
func (w *Writer) flushHeader() error {
	if w.headerDone {
		return nil
	}
	w.headerDone = true
	_, err := io.WriteString(w.w, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	if err != nil {
		w.err = err
	}
	return err
}

// flushPending closes any pending open tag with '>'.
func (w *Writer) flushPending(selfClose bool) error {
	if !w.pendingOpen {
		return nil
	}
	w.pendingOpen = false
	ch := ">"
	if selfClose {
		ch = "/>"
		w.stack = w.stack[:len(w.stack)-1]
	}
	if _, err := io.WriteString(w.w, ch); err != nil {
		w.err = err
		return err
	}
	return nil
}

// qualifiedName returns "prefix:local" if a prefix is registered, else just "local".
func (w *Writer) qualifiedName(name xml.Name) string {
	if name.Space == "" {
		return name.Local
	}
	if prefix, ok := w.nsMap[name.Space]; ok && prefix != "" {
		return prefix + ":" + name.Local
	}
	return name.Local
}

// escapeCharData escapes text for XML character data (element content).
func escapeCharData(s string) string {
	// Strip NUL bytes — invalid in XML 1.0.
	s = strings.ReplaceAll(s, "\x00", "")
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			sb.WriteString("&amp;")
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// escapeAttr escapes text for use inside a double-quoted attribute value.
func escapeAttr(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	var sb strings.Builder
	sb.Grow(len(s))
	for _, r := range s {
		switch r {
		case '&':
			sb.WriteString("&amp;")
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		case '"':
			sb.WriteString("&quot;")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
