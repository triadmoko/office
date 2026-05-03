package docx

import (
	"strings"

	"github.com/triadmoko/office/internal/wml"
)

// TOCFieldOptions configures a Word complex field TOC (Table of Contents).
// The library only emits OOXML; Microsoft Word evaluates the field and fills entries
// when the user updates fields (e.g. right-click → Update Field).
// WPS Writer, LibreOffice, and many viewers often leave TOC fields unevaluated (placeholder only).
type TOCFieldOptions struct {
	// OutlineLevels is the \\o switch range, e.g. "1-3" for heading levels 1–3. Empty means "1-3".
	OutlineLevels string
	// Hyperlinks adds \\h so Word uses hyperlinks for TOC entries when the field is updated.
	Hyperlinks bool
	// OmitPageNumbersInWebLayout adds \\z (hide page numbers in Web Layout).
	OmitPageNumbersInWebLayout bool
	// UseAppliedOutlineLevels adds \\u (paragraph outline from w:pPr/w:outlineLvl only, not from styles).
	// Leave false for typical Heading1–3 via w:pStyle; \\u with style-only headings often yields an empty TOC.
	UseAppliedOutlineLevels bool
	// Placeholder is shown between the field separator and end until Word refreshes the TOC.
	// Empty uses a short Indonesian hint.
	Placeholder string
}

func appendRunUnknownPara(p *wml.Paragraph, innerXML string) {
	r := &wml.Run{Parts: []wml.RunPart{{Unknown: []byte(innerXML)}}}
	r.RebuildText()
	p.Runs = append(p.Runs, r)
}

func buildTOCInstruction(lev string, hyperlinks, omitWeb, useApplied bool) string {
	lev = strings.TrimSpace(lev)
	if lev == "" {
		lev = "1-3"
	}
	lev = strings.ReplaceAll(lev, `"`, `\"`)

	var b strings.Builder
	b.WriteString(" TOC ")
	b.WriteString(`\o "`)
	b.WriteString(lev)
	b.WriteString(`"`)
	if hyperlinks {
		b.WriteString(` \h`)
	}
	if omitWeb {
		b.WriteString(` \z`)
	}
	if useApplied {
		b.WriteString(` \u`)
	}
	// Word expects leading/trailing spaces inside instrText for many fields.
	return " " + strings.TrimSpace(b.String()) + " "
}

// AppendTOCField appends a Word TOC complex field (w:fldChar / w:instrText) to this paragraph.
// Defaults when all switch booleans are false: \\h and \\z on, \\u off (\\u skips style-based outline). OutlineLevels default "1-3".
// For a readable outline in WPS/LibreOffice without field support, emit a plain-text list separately (see sample in cmd/office).
func (p *Paragraph) AppendTOCField(opts TOCFieldOptions) {
	if p == nil || p.x == nil {
		return
	}
	lev := opts.OutlineLevels
	if lev == "" {
		lev = "1-3"
	}
	hyper := opts.Hyperlinks
	omitWeb := opts.OmitPageNumbersInWebLayout
	useApplied := opts.UseAppliedOutlineLevels
	if !hyper && !omitWeb && !useApplied {
		hyper, omitWeb, useApplied = true, true, false
	}
	ph := opts.Placeholder
	if ph == "" {
		ph = "Klik kanan → Perbarui bidang (Update Field) untuk mengisi daftar isi."
	}

	instr := buildTOCInstruction(lev, hyper, omitWeb, useApplied)

	// w:dirty hints consumers that the field result may be stale (Word clears after refresh).
	appendRunUnknownPara(p.x, `<w:fldChar w:fldCharType="begin" w:dirty="true"/>`)
	appendRunUnknownPara(p.x, `<w:instrText xml:space="preserve">`+escapeCharData(instr)+`</w:instrText>`)
	appendRunUnknownPara(p.x, `<w:fldChar w:fldCharType="separate"/>`)

	rPh := &wml.Run{Parts: []wml.RunPart{{Text: ph}}}
	rPh.RebuildText()
	p.x.Runs = append(p.x.Runs, rPh)

	appendRunUnknownPara(p.x, `<w:fldChar w:fldCharType="end"/>`)
}
