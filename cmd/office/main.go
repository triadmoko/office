package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/triadmoko/office/docx"
)

// contoh DOCX dalam satu fungsi: sampul + TOC + pemecah bagian (next page) + isi.
// Daftar isi statis (bukan bidang TOC Word). Header + footer: bidang PAGE/NUMPAGES (NewDocument+Save).

func main() {
	out := flag.String("o", "office-sample.docx", "path file .docx keluaran")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Pembuat contoh DOCX (github.com/triadmoko/office/docx).\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nContoh: %s -o laporan.docx\n", os.Args[0])
	}
	flag.Parse()

	if err := writeSampleDocx(*out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote", *out)
}

func writeSampleDocx(path string) error {
	d := docx.NewDocument()
	// Saat membuka DOCX dari Word lalu Save lagi: hilangkan jejak layout w:lastRenderedPageBreak (bukan instruksi page break).
	d.SetStripLayoutHints(true)

	d.SetFooterPageNumber(true)
	d.SetFooterPageNumberTemplate(
		"No. " + docx.FooterPlaceholderPage +
			" / " + docx.FooterPlaceholderNumPages + " hal.",
	)
	d.SetHeaderPageNumber(true)
	d.SetHeaderPageNumberTemplate(
		"Laporan · p. " + docx.HeaderPlaceholderPage,
	)
	if sec := d.SectionAt(0); sec != nil {
		sec.SetPageSize(docx.PageSizeA4)
		sec.SetOrientation(docx.Portrait)
		sec.SetMargins(docx.Margins{
			Top: 1800, Bottom: 1800, Left: 1440, Right: 1440,
			Header: 720, Footer: 720, Gutter: 0,
		})
	}

	// ----- Sampul -----
	title := d.Body().Paragraphs()[0]
	title.SetAlignment(docx.AlignCenter)
	title.SetSpacing(docx.Spacing{After: 360})
	rt := title.AppendRun("LAPORAN CONTOH")
	rt.SetBold(true)
	rt.SetSize(56)

	sub := d.Body().AppendParagraph()
	sub.SetAlignment(docx.AlignCenter)
	sub.SetSpacing(docx.Spacing{After: 720})
	rs := sub.AppendRun("Paket github.com/triadmoko/office/docx")
	rs.SetSize(28)
	for i := 0; i < 8; i++ {
		sp := d.Body().AppendParagraph()
		sp.SetAlignment(docx.AlignCenter)
		sp.AppendRun("\u00a0")
	}
	meta := d.Body().AppendParagraph()
	meta.SetAlignment(docx.AlignCenter)
	meta.AppendRun("Disusun untuk demonstrasi API")
	ver := d.Body().AppendParagraph()
	ver.SetAlignment(docx.AlignCenter)
	ver.AppendRun("Setelah TOC: SetSectionBreak(nextPage) memisahkan bagian depan dan isi utama.")

	pCoverEnd := d.Body().AppendParagraph()
	pCoverEnd.SetAlignment(docx.AlignCenter)
	pCoverEnd.AppendRun("Halaman berikutnya: daftar isi.")
	pCoverEnd.AppendPageBreak()

	// ----- Daftar isi (manual) -----
	hTOC := d.Body().AppendParagraph()
	hTOC.SetSpacing(docx.Spacing{After: 240})
	hTOCr := hTOC.AppendRun("Daftar isi")
	hTOCr.SetBold(true)
	hTOCr.SetSize(32)
	pTOCIntro := d.Body().AppendParagraph()
	pTOCIntro.SetSpacing(docx.Spacing{After: 180})
	pTOCIntro.AppendRun("Entri manual (titik). TOC otomatis Word = Tier 3.")

	tocEntries := []struct{ title, page string }{
		{"1. Pendahuluan", "3"},
		{"2. Format paragraf & perataan", "3"},
		{"3. Format run (tebal, warna, …)", "4"},
		{"4. Daftar", "4"},
		{"5. Tabel", "4"},
		{"6. Pagination & pemisah OOXML", "6"},
	}
	const tocWidth = 62
	for _, e := range tocEntries {
		p := d.Body().AppendParagraph()
		p.SetSpacing(docx.Spacing{After: 120})
		dots := tocWidth - len(e.title) - len(e.page)
		if dots < 3 {
			dots = 3
		}
		line := e.title + " " + strings.Repeat(".", dots) + " " + e.page
		r := p.AppendRun(line)
		r.SetSize(22)
	}

	// ----- Pemecah bagian: akhir front matter → section baru (isi utama), next page, portrait -----
	pSect := d.Body().AppendParagraph()
	pSect.SetSpacing(docx.Spacing{Before: 240, After: 120})
	pSect.AppendRun("Di bawah ini w:sectPr + w:type nextPage: section baru dimulai setelah paragraf ini (biasanya halaman baru).")
	pSect.SetSectionBreak(docx.SectionBreakConfig{
		PageKind: docx.PageSizeA4,
		Orient:   docx.Portrait,
		Break:    docx.SectionBreakNextPage,
	})

	if sec := d.SectionAt(1); sec != nil {
		sec.SetPageSize(docx.PageSizeA4)
		sec.SetOrientation(docx.Portrait)
		sec.SetMargins(docx.Margins{
			Top: 1800, Bottom: 1800, Left: 1440, Right: 1440,
			Header: 720, Footer: 720, Gutter: 0,
		})
		// Penomoran halaman isi utama: mulai dari 1, format desimal (w:pgNumType di w:sectPr bagian ini).
		sec.SetPageNumberFormat(docx.PageNumberFormatDecimal)
		mainStart := 1
		sec.SetPageNumberStart(&mainStart)
	}

	// ----- Isi utama -----
	h1 := d.Body().AppendParagraph()
	h1.SetSpacing(docx.Spacing{Before: 120, After: 200})
	h1r := h1.AppendRun("1. Pendahuluan")
	h1r.SetBold(true)
	h1r.SetSize(28)
	pIntro := d.Body().AppendParagraph()
	pIntro.SetSpacing(docx.Spacing{After: 200})
	pIntro.AppendRun("Ini isi utama setelah section break. Di bawah: demo paragraf, run, daftar, tabel.")

	if st := d.Styles(); st != nil {
		if h := st.ByID("Heading1"); h != nil && h.Name() != "" {
			d.Body().AppendParagraph().AppendRun("Heading1 terdaftar sebagai: " + h.Name())
		}
	}

	h2 := d.Body().AppendParagraph()
	h2.SetSpacing(docx.Spacing{Before: 240, After: 160})
	h2r := h2.AppendRun("2. Format paragraf")
	h2r.SetBold(true)
	h2r.SetSize(28)
	pInd := d.Body().AppendParagraph()
	pInd.SetIndent(docx.Indent{Left: 720, FirstLine: 720})
	pInd.SetSpacing(docx.Spacing{After: 240})
	pInd.AppendRun("Indent + jarak setelah paragraf. Lorem ipsum dolor sit amet.")

	pAL := d.Body().AppendParagraph()
	pAL.SetAlignment(docx.AlignLeft)
	pAL.AppendRun("[Kiri] Teks rata kiri.")

	pAC := d.Body().AppendParagraph()
	pAC.SetAlignment(docx.AlignCenter)
	pAC.AppendRun("[Tengah] Judul singkat.")

	pAR := d.Body().AppendParagraph()
	pAR.SetAlignment(docx.AlignRight)
	pAR.AppendRun("[Kanan] Catatan di kanan.")

	pAJ := d.Body().AppendParagraph()
	pAJ.SetAlignment(docx.AlignJustify)
	pAJ.AppendRun("[Justify] Paragraf panjang: Lorem ipsum dolor sit amet, consectetur adipiscing elit.")

	pPg := d.Body().AppendParagraph()
	pPg.SetSpacing(docx.Spacing{Before: 120, After: 120})
	pPg.AppendRun("Page break berikutnya untuk demo run.")
	pPg.AppendPageBreak()

	h3 := d.Body().AppendParagraph()
	h3.SetSpacing(docx.Spacing{After: 160})
	h3r := h3.AppendRun("3. Format run")
	h3r.SetBold(true)
	h3r.SetSize(28)
	pRun := d.Body().AppendParagraph()
	pRun.AppendRun("Teks biasa, ")
	rBold := pRun.AppendRun("tebal")
	rBold.SetBold(true)
	pRun.AppendRun(", ")
	rIt := pRun.AppendRun("miring")
	rIt.SetItalic(true)
	pRun.AppendRun(", ")
	rU := pRun.AppendRun("garis bawah")
	rU.SetUnderline(true)
	pRun.AppendRun(", ")
	rC := pRun.AppendRun("warna")
	rC.SetColor("C00000")
	rC.SetBold(true)
	pRun.AppendRun(", ")
	rHi := pRun.AppendRun("sorot")
	rHi.SetHighlight("yellow")
	pRun.AppendRun(", ")
	rSt := pRun.AppendRun("coret")
	rSt.SetStrike(true)
	pRun.AppendRun(", ")
	rEm := pRun.AppendRun("emphasis dot")
	rEm.SetEmphasis("dot")
	pRun.AppendRun(", x")
	rSup := pRun.AppendRun("2")
	rSup.SetSubSuperscript(docx.VertAlignSuperscript)
	pRun.AppendRun(" / H")
	rSub := pRun.AppendRun("2")
	rSub.SetSubSuperscript(docx.VertAlignSubscript)
	pRun.AppendRun("O.")

	h4 := d.Body().AppendParagraph()
	h4.SetSpacing(docx.Spacing{Before: 240, After: 160})
	h4r := h4.AppendRun("4. Daftar")
	h4r.SetBold(true)
	h4r.SetSize(28)
	d.Body().AppendParagraph().AppendRun("Bullet:")
	bl := d.Body().AppendList(docx.ListBullet)
	bl.AppendItem("Butir pertama")
	bl.AppendItem("Butir kedua")
	bl.AppendItem("Butir ketiga")
	d.Body().AppendParagraph().AppendRun("Bernomor:")
	nl := d.Body().AppendList(docx.ListNumbered)
	nl.AppendItem("Langkah satu")
	nl.AppendItem("Langkah dua")

	h5 := d.Body().AppendParagraph()
	h5.SetSpacing(docx.Spacing{Before: 240, After: 160})
	h5r := h5.AppendRun("5. Tabel")
	h5r.SetBold(true)
	h5r.SetSize(28)
	d.Body().AppendParagraph().AppendRun("Tabel dengan border:")
	tbl := d.Body().AppendTable(3, 2)
	tbl.SetGridColWidths([]int64{2000, 2000})
	tbl.Rows()[0].SetHeight(500, docx.RowHeightAtLeast)
	tbl.Rows()[0].SetRepeatAsHeaderRow(true) // w:tblHeader — baris judul diulang di atas setiap halaman (perilaku Word)
	tbl.Rows()[1].SetCantSplit(true)       // w:cantSplit — baris ini tidak boleh terpotong antar halaman
	tbl.SetBorder(
		docx.BorderTop|docx.BorderLeft|docx.BorderBottom|docx.BorderRight|docx.BorderInsideH|docx.BorderInsideV,
		docx.BorderStyle{Color: "4472C4", Size: 8, Kind: docx.BorderSingle},
	)
	headers := []string{"Kolom A", "Kolom B"}
	for c, hn := range headers {
		cell := tbl.Rows()[0].Cells()[c]
		rh := cell.Paragraphs()[0].AppendRun(hn)
		rh.SetBold(true)
	}
	tbl.Rows()[1].Cells()[0].Paragraphs()[0].AppendRun("Sel 2,1")
	tbl.Rows()[1].Cells()[1].Paragraphs()[0].AppendRun("Sel 2,2")
	tbl.Rows()[2].Cells()[0].Paragraphs()[0].AppendRun("Teks panjang (wrap).")
	tbl.Rows()[2].Cells()[1].Paragraphs()[0].AppendRun("OK")

	// ----- Pagination & marka OOXML (w:pPr, w:br, w:trPr) -----
	h6 := d.Body().AppendParagraph()
	h6.SetSpacing(docx.Spacing{Before: 240, After: 160})
	h6r := h6.AppendRun("6. Pagination & pemisah OOXML")
	h6r.SetBold(true)
	h6r.SetSize(28)

	d.Body().AppendParagraph().AppendRun("Paragraf berikut memakai keepNext: paragraf ini tetap bersama paragraf berikutnya di halaman yang sama bila memungkinkan.")
	pKN := d.Body().AppendParagraph()
	pKN.SetKeepNext(true)
	pKN.SetSpacing(docx.Spacing{After: 60})
	pKN.AppendRun("[keepNext] Judul kecil yang dijaga bersama blok berikut.")
	pKNFollow := d.Body().AppendParagraph()
	pKNFollow.SetSpacing(docx.Spacing{After: 200})
	pKNFollow.AppendRun("[ikut keepNext] Blok narasi yang ingin tidak terpisah halaman dari paragraf di atas.")

	pPBB := d.Body().AppendParagraph()
	pPBB.SetPageBreakBefore(true)
	pPBB.SetSpacing(docx.Spacing{After: 120})
	pPBB.AppendRun("[pageBreakBefore] Paragraf ini meminta halaman baru sebelum dirinya (w:pageBreakBefore).")

	pKL := d.Body().AppendParagraph()
	pKL.SetKeepLines(true)
	on := true
	pKL.SetWidowControl(&on)
	pKL.SetSpacing(docx.Spacing{After: 120})
	pKL.AppendRun("[keepLines + widowControl] Mengurangi janda/yatim: seluruh baris paragraf dijaga, kontrol widow/line orphan aktif.")

	pCol := d.Body().AppendParagraph()
	pCol.SetSpacing(docx.Spacing{After: 120})
	pCol.AppendRun("Dalam layout multi-kolom Word, ")
	pCol.AppendColumnBreak() // w:br w:type="column"
	pCol.AppendRun("teks setelah ini lanjut ke kolom berikutnya.")

	d.Body().AppendParagraph().AppendRun("(Catatan: AppendPageBreak = w:br w:type page sudah dipakai di atas menuju bagian format run.)")
	d.Body().AppendParagraph().AppendRun(
		"API section: SetPageNumberStart(&n) + SetPageNumberFormat (decimal, lowerRoman, upperRoman, lowerLetter, …) mengatur w:pgNumType; footer {{PAGE}} mengikuti bagian aktif.",
	)

	pEnd := d.Body().AppendParagraph()
	pEnd.SetSpacing(docx.Spacing{Before: 360})
	pEnd.AppendRun("Selesai — periksa struktur: sampul + TOC (section pertama), lalu section break next page, lalu isi, pagination OOXML, SetStripLayoutHints untuk save bersih dari lastRenderedPageBreak.")

	return d.SaveFile(path)
}
