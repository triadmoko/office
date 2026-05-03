package xlsx

import (
	"bytes"
	"fmt"

	"github.com/triadmoko/office/internal/sml"
)

const relTypeStyles = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles"

func (w *Workbook) ensureStyles() error {
	if w == nil {
		return ErrMissingMainPart
	}
	if w.fromNew {
		w.stylesOnce.Do(func() {
			w.styles = &sml.StylesTable{}
		})
		return w.stylesErr
	}
	if err := w.ensureMeta(); err != nil {
		return err
	}
	w.stylesOnce.Do(w.loadStyles)
	return w.stylesErr
}

func (w *Workbook) loadStyles() {
	path, ok, err := resolveRelatedPartByType(w.pkg, w.main, relTypeStyles)
	if err != nil {
		w.stylesErr = fmt.Errorf("xlsx: styles rel: %w", err)
		return
	}
	if !ok || !w.pkg.HasPart(path) {
		w.styles = &sml.StylesTable{}
		return
	}
	body, err := w.pkg.ReadFile(path)
	if err != nil {
		w.stylesErr = fmt.Errorf("xlsx: read styles: %w", err)
		return
	}
	st, err := sml.ParseStylesTable(bytes.NewReader(body))
	if err != nil {
		w.stylesErr = fmt.Errorf("xlsx: parse styles: %w", err)
		return
	}
	w.styles = st
}

func (w *Workbook) stylesTable() *sml.StylesTable {
	if w == nil {
		return nil
	}
	_ = w.ensureStyles()
	if w.styles == nil {
		return &sml.StylesTable{}
	}
	return w.styles
}
