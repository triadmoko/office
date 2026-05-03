package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/triadmoko/office/docx"
)

// contoh DOCX tunggal: sectPr, footer (PAGE/NUMPAGES), styles, paragraf, run,
// page break, section break, numbering, tabel, Save.

func main() {
	out := flag.String("o", "office-sample.docx", "path file .docx keluaran")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Pembuat contoh DOCX (paket github.com/triadmoko/office/docx).\n\n")
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

	applySampleFooter(d)
	applyFirstSectionLayout(d)

	addCoverBlock(d)
	addHeadingRegistryDemo(d)
	addIndentAndAlignmentDemos(d)
	addPageBreakDemo(d)
	addRunFormattingDemo(d)
	addListDemos(d)
	addTableDemo(d)
	addSectionBreakDemos(d)
	addClosingParagraph(d)

	return d.SaveFile(path)
}

// --- Footer (MVP: hanya NewDocument + Save) ---

func applySampleFooter(d *docx.Document) {
	d.SetFooterPageNumber(true)
	d.SetFooterPageNumberTemplate(
		"No. " + docx.FooterPlaceholderPage +
			" / " + docx.FooterPlaceholderNumPages + " hal.",
	)
}

// --- Section pertama: A4, portrait, margin ---

func applyFirstSectionLayout(d *docx.Document) {
	sec := d.SectionAt(0)
	if sec == nil {
		return
	}
	sec.SetPageSize(docx.PageSizeA4)
	sec.SetOrientation(docx.Portrait)
	// ~2,5 cm atas/bawah, ~2 cm kiri/kanan (1440 twip ≈ 1 inch)
	sec.SetMargins(docx.Margins{
		Top: 1800, Bottom: 1800, Left: 1440, Right: 1440,
		Header: 720, Footer: 720, Gutter: 0,
	})
}

// --- Judul & catatan ---

func addCoverBlock(d *docx.Document) {
	title := d.Body().Paragraphs()[0]
	r := title.AppendRun("Contoh Office (DOCX)")
	r.SetBold(true)
	r.SetSize(36) // 18 pt

	note := d.Body().AppendParagraph()
	note.AppendRun("Catatan: footer kanan memakai template No. + PAGE + NUMPAGES. " +
		"Setelah pemecah halaman, Word memperbarui nomor per halaman.")
}

// --- Styles (registry) ---

func addHeadingRegistryDemo(d *docx.Document) {
	st := d.Styles()
	if st == nil {
		return
	}
	h := st.ByID("Heading1")
	if h == nil || h.Name() == "" {
		return
	}
	p := d.Body().AppendParagraph()
	p.AppendRun("Style Heading1 terdaftar sebagai: " + h.Name())
}

// --- Paragraf: indent, spacing, perataan ---

func addIndentAndAlignmentDemos(d *docx.Document) {
	pInd := d.Body().AppendParagraph()
	pInd.SetIndent(docx.Indent{Left: 720, FirstLine: 720})
	pInd.SetSpacing(docx.Spacing{After: 240})
	pInd.AppendRun("Paragraf dengan indent kiri + baris pertama (w:ind) dan jarak setelah paragraf (w:spacing). " +
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit.")

	pAL := d.Body().AppendParagraph()
	pAL.SetAlignment(docx.AlignLeft)
	pAL.AppendRun("[Kiri] Teks rata kiri (default).")

	pAC := d.Body().AppendParagraph()
	pAC.SetAlignment(docx.AlignCenter)
	pAC.AppendRun("[Tengah] Judul atau baris singkat di tengah.")

	pAR := d.Body().AppendParagraph()
	pAR.SetAlignment(docx.AlignRight)
	pAR.AppendRun("[Kanan] Catatan atau tanggal di kanan.")

	pAJ := d.Body().AppendParagraph()
	pAJ.SetAlignment(docx.AlignJustify)
	pAJ.AppendRun("[Justify] Paragraf panjang: Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
		"Integer vitae velit non ligula faucibus auctor. Donec vitae sapien ut libero venenatis faucibus.")
}

// --- Pemecah halaman ---

func addPageBreakDemo(d *docx.Document) {
	p1 := d.Body().AppendParagraph()
	p1.AppendRun("Halaman 1 — di bawah ini page break.")
	p1.AppendPageBreak()

	p2 := d.Body().AppendParagraph()
	p2.AppendRun("Halaman 2 — cek footer menampilkan nomor 2.")
}

// --- Run: bold, italic, underline, warna, highlight, strike, em, super/sub ---

func addRunFormattingDemo(d *docx.Document) {
	p := d.Body().AppendParagraph()
	p.AppendRun("Teks biasa, ")
	rBold := p.AppendRun("tebal")
	rBold.SetBold(true)
	p.AppendRun(", ")
	rIt := p.AppendRun("miring")
	rIt.SetItalic(true)
	p.AppendRun(", ")
	rU := p.AppendRun("garis bawah")
	rU.SetUnderline(true)
	p.AppendRun(", ")
	rC := p.AppendRun("warna")
	rC.SetColor("C00000")
	rC.SetBold(true)
	p.AppendRun(", ")
	rHi := p.AppendRun("sorot")
	rHi.SetHighlight("yellow")
	p.AppendRun(", ")
	rSt := p.AppendRun("coret")
	rSt.SetStrike(true)
	p.AppendRun(", ")
	rEm := p.AppendRun("Hellow selamat siang")
	rEm.SetEmphasis("dot")
	p.AppendRun(", ")
	p.AppendRun("x")
	rSup := p.AppendRun("2")
	rSup.SetSubSuperscript(docx.VertAlignSuperscript)
	p.AppendRun(" / H")
	rSub := p.AppendRun("2")
	rSub.SetSubSuperscript(docx.VertAlignSubscript)
	p.AppendRun("O.")
}

// --- Daftar ---

func addListDemos(d *docx.Document) {
	d.Body().AppendParagraph().AppendRun("Daftar bullet:")
	bl := d.Body().AppendList(docx.ListBullet)
	bl.AppendItem("Butir pertama")
	bl.AppendItem("Butir kedua")
	bl.AppendItem("Butir ketiga")

	d.Body().AppendParagraph().AppendRun("Daftar bernomor:")
	nl := d.Body().AppendList(docx.ListNumbered)
	nl.AppendItem("Langkah satu")
	nl.AppendItem("Langkah dua")
}

// --- Tabel ---

func addTableDemo(d *docx.Document) {
	d.Body().AppendParagraph().AppendRun("Tabel (border + isi sel):")
	tbl := d.Body().AppendTable(3, 2)
	tbl.SetGridColWidths([]int64{2000, 2000})
	tbl.Rows()[0].SetHeight(500, docx.RowHeightAtLeast)
	tbl.SetBorder(
		docx.BorderTop|docx.BorderLeft|docx.BorderBottom|docx.BorderRight|docx.BorderInsideH|docx.BorderInsideV,
		docx.BorderStyle{Color: "4472C4", Size: 8, Kind: docx.BorderSingle},
	)
	headers := []string{"Kolom A", "Kolom B"}
	for c, h := range headers {
		cell := tbl.Rows()[0].Cells()[c]
		rh := cell.Paragraphs()[0].AppendRun(h)
		rh.SetBold(true)
	}
	tbl.Rows()[1].Cells()[0].Paragraphs()[0].AppendRun("Sel baris 2, kolom 1")
	tbl.Rows()[1].Cells()[1].Paragraphs()[0].AppendRun("Sel baris 2, kolom 2")
	tbl.Rows()[2].Cells()[0].Paragraphs()[0].AppendRun("Teks panjang di satu sel (uji wrap).")
	tbl.Rows()[2].Cells()[1].Paragraphs()[0].AppendRun("OK")
}

// --- Pemecah bagian: lanskap lalu portrait (continuous) ---

func addSectionBreakDemos(d *docx.Document) {
	pLandscape := d.Body().AppendParagraph()
	pLandscape.AppendRun("Pemecah bagian berikut: section berikutnya lanskap.")
	pLandscape.SetSectionBreak(docx.SectionBreakConfig{
		PageKind: docx.PageSizeA4,
		Orient:   docx.Landscape,
		Break:    docx.SectionBreakNextPage,
	})
	d.Body().AppendParagraph().AppendRun(
		"Isi di section lanskap. Atur lewat d.SectionAt(i) sesuai indeks bagian.")

	pPortrait := d.Body().AppendParagraph()
	pPortrait.AppendRun("Pemecah bagian continuous: orientasi portrait lagi (tanpa halaman baru wajib).")
	pPortrait.SetSectionBreak(docx.SectionBreakConfig{
		PageKind: docx.PageSizeA4,
		Orient:   docx.Portrait,
		Break:    docx.SectionBreakContinuous,
	})
	d.Body().AppendParagraph().AppendRun(
		"Isi setelah section portrait (continuous).")
}

func addClosingParagraph(d *docx.Document) {
	d.Body().AppendParagraph().AppendRun("Akhir contoh — buka di Word/LibreOffice untuk memeriksa tampilan.")
}
