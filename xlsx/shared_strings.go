package xlsx

import (
	"bytes"
	"fmt"

	"github.com/triadmoko/office/internal/sml"
)

func (w *Workbook) ensureSharedStrings() error {
	if w == nil {
		return ErrMissingMainPart
	}
	if w.fromNew {
		return nil
	}
	if err := w.ensureMeta(); err != nil {
		return err
	}
	w.sharedOnce.Do(w.loadSharedStrings)
	return w.sharedErr
}

func (w *Workbook) loadSharedStrings() {
	path, ok, err := resolveRelatedPartByType(w.pkg, w.main, relTypeSharedStrings)
	if err != nil {
		w.sharedErr = fmt.Errorf("xlsx: shared strings rel: %w", err)
		return
	}
	if !ok {
		w.sharedStrings = []string{}
		return
	}
	if !w.pkg.HasPart(path) {
		w.sharedErr = fmt.Errorf("xlsx: shared strings part missing %q", path)
		return
	}
	body, err := w.pkg.ReadFile(path)
	if err != nil {
		w.sharedErr = fmt.Errorf("xlsx: read shared strings: %w", err)
		return
	}
	ss, err := sml.ParseSharedStrings(bytes.NewReader(body))
	if err != nil {
		w.sharedErr = fmt.Errorf("xlsx: parse sharedStrings.xml: %w", err)
		return
	}
	w.sharedStrings = ss
}

// SharedString returns the string at index in xl/sharedStrings.xml (lazy-loaded).
func (w *Workbook) SharedString(idx int) (string, error) {
	if w == nil {
		return "", ErrMissingMainPart
	}
	if err := w.ensureSharedStrings(); err != nil {
		return "", err
	}
	if idx < 0 || idx >= len(w.sharedStrings) {
		return "", ErrSharedStringOutOfRange
	}
	return w.sharedStrings[idx], nil
}
