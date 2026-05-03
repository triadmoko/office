package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/triadmoko/office/docx"
)

// contoh DOCX tunggal yang memakai fitur MVP: sectPr (ukuran halaman & margin),
// styles (baca registry), paragraf (indent/spacing/alignment), numbering (bullet + numbered),
// run (bold/italic/garis bawah/warna/sorot teks/coret/emphasis mark/superscript/subscript/ukuran),
// pemecah halaman (w:br type page), pemecah bagian (w:sectPr + SectionBreakConfig.Break: nextPage/continuous/…),
// tabel + border sel, serta Save (paket lengkap: document, styles, numbering, props).
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

	// --- Section (OFFICE-106 / API section.go): ukuran A4, margin, orientasi portrait ---
	sec := d.SectionAt(0)
	if sec != nil {
		sec.SetPageSize(docx.PageSizeA4)
		sec.SetOrientation(docx.Portrait)
		// margin ~2,5 cm atas/bawah, ~2 cm kiri/kanan (twips ≈ 567 per 1 mm tidak dipakai;
		// nilai umum Word: 1440 twip = 1 inch)
		sec.SetMargins(docx.Margins{
			Top: 1800, Bottom: 1800, Left: 1440, Right: 1440,
			Header: 720, Footer: 720, Gutter: 0,
		})
	}

	// Pakai paragraf kosong pertama dari EmptyDocument sebagai judul (tanpa paragraf kosong tambahan di awal).
	title := d.Body().Paragraphs()[0]
	rt := title.AppendRun("Contoh Office (DOCX)")
	rt.SetBold(true)
	rt.SetSize(36) // 18 pt

	// --- Styles: registry dari styles.xml in-memory (OFFICE-104) ---
	st := d.Styles()
	if st != nil {
		if h := st.ByID("Heading1"); h != nil && h.Name() != "" {
			p := d.Body().AppendParagraph()
			p.AppendRun("Paragraf setelah judul — style Heading1 terdaftar sebagai: " + h.Name())
		}
	}

	// --- Paragraf: w:ind + w:spacing (twip) ---
	pInd := d.Body().AppendParagraph()
	pInd.SetIndent(docx.Indent{Left: 720, FirstLine: 720}) // ~0,5 in; first line indent
	pInd.SetSpacing(docx.Spacing{After: 240})
	pInd.AppendRun("Paragraf dengan indent kiri dan baris pertama (w:ind), plus jarak setelah paragraf (w:spacing). Silahkan lihat di atas. Contoh paragraf ini. Paragraf ini juga memakai style Heading1. lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")

	// --- Perataan paragraf (w:jc) ---
	pAL := d.Body().AppendParagraph()
	pAL.SetAlignment(docx.AlignLeft)
	pAL.AppendRun("[Kiri] Teks ini rata kiri (default).")

	pAC := d.Body().AppendParagraph()
	pAC.SetAlignment(docx.AlignCenter)
	pAC.AppendRun("[Tengah] Judul atau baris singkat di tengah.")

	pAR := d.Body().AppendParagraph()
	pAR.SetAlignment(docx.AlignRight)
	pAR.AppendRun("[Kanan] Catatan atau tanggal di kanan.")

	pAJ := d.Body().AppendParagraph()
	pAJ.SetAlignment(docx.AlignJustify)
	pAJ.AppendRun("[Rata kiri-kanan] Paragraf panjang supaya justify terlihat: Lorem ipsum dolor sit amet, consectetur adipiscing elit. Integer vitae velit non ligula faucibus auctor. Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante etiam sit amet orci eget eros faucibus tincidunt.")

	// --- Halaman baru (w:br w:type="page") ---
	pPg := d.Body().AppendParagraph()
	pPg.AppendRun("Teks di halaman pertama (di bawah ini pemecah halaman).")
	pPg.AppendPageBreak()
	pPg2 := d.Body().AppendParagraph()
	pPg2.AppendRun("Teks ini dimulai di halaman baru (lihat di Word: tampilan halaman / cetak).")

	// --- Run formatting ---
	pFmt := d.Body().AppendParagraph()
	pFmt.AppendRun("Teks biasa, ")
	rBold := pFmt.AppendRun("tebal")
	rBold.SetBold(true)
	pFmt.AppendRun(", ")
	rIt := pFmt.AppendRun("miring")
	rIt.SetItalic(true)
	pFmt.AppendRun(", ")
	rU := pFmt.AppendRun("garis bawah")
	rU.SetUnderline(true)
	pFmt.AppendRun(", ")
	rC := pFmt.AppendRun("warna")
	rC.SetColor("C00000")
	rC.SetBold(true)
	pFmt.AppendRun(", ")
	rHi := pFmt.AppendRun("sorot")
	rHi.SetHighlight("yellow")
	pFmt.AppendRun(", ")
	rSt := pFmt.AppendRun("coret")
	rSt.SetStrike(true)
	pFmt.AppendRun(", ")
	// w:em (tanda penekanan Asia Timur, mis. titik di atas teks — terlihat jelas pada font CJK).
	rEm := pFmt.AppendRun("Hellow selamat siang")
	rEm.SetEmphasis("dot")
	pFmt.AppendRun(", ")
	pFmt.AppendRun("x")
	rSup := pFmt.AppendRun("2")
	rSup.SetSubSuperscript(docx.VertAlignSuperscript)
	pFmt.AppendRun(" / H")
	rSub := pFmt.AppendRun("2")
	rSub.SetSubSuperscript(docx.VertAlignSubscript)
	pFmt.AppendRun("O")
	pFmt.AppendRun(".")

	// --- Numbering: bullet + numbered (OFFICE-108) ---
	d.Body().AppendParagraph().AppendRun("Daftar bullet:")
	bl := d.Body().AppendList(docx.ListBullet)
	bl.AppendItem("Butir pertama")
	bl.AppendItem("Butir kedua")
	bl.AppendItem("Butir ketiga")

	d.Body().AppendParagraph().AppendRun("Daftar bernomor:")
	nl := d.Body().AppendList(docx.ListNumbered)
	nl.AppendItem("Langkah satu")
	nl.AppendItem("Langkah dua")

	// --- Table + borders (OFFICE-108) ---
	d.Body().AppendParagraph().AppendRun("Tabel ringkas (border + isi sel):")
	tbl := d.Body().AppendTable(3, 2) // lebar default = penuh area teks (lihat docx.Body.AppendTable)
	// Lebar kolom (twip) + tinggi baris header: w:tblGrid + w:trHeight.
	tbl.SetGridColWidths([]int64{2000, 2000})
	tbl.Rows()[0].SetHeight(500, docx.RowHeightAtLeast)
	tbl.SetBorder(docx.BorderTop|docx.BorderLeft|docx.BorderBottom|docx.BorderRight|docx.BorderInsideH|docx.BorderInsideV,
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
	tbl.Rows()[2].Cells()[0].Paragraphs()[0].AppendRun("Gabungan teks panjang di satu sel untuk uji wrap.")
	tbl.Rows()[2].Cells()[1].Paragraphs()[0].AppendRun("OK")

	// --- Orientasi per bagian: section baru (lanskap) setelah paragraf ini ---
	pSec := d.Body().AppendParagraph()
	pSec.AppendRun("Di bawah ini pemecah bagian: isi berikutnya memakai orientasi lanskap (cek Layout di Word).")
	pSec.SetSectionBreak(docx.SectionBreakConfig{
		PageKind: docx.PageSizeA4,
		Orient:   docx.Landscape,
		Break:    docx.SectionBreakNextPage, // atau docx.SectionBreakContinuous untuk tanpa halaman baru
	})
	d.Body().AppendParagraph().AppendRun("Teks di section kedua (landskap). Sesuaikan dengan d.SectionAt(i).SetOrientation(...) per indeks Sections().")

	d.Body().AppendParagraph().AppendRun("Akhir contoh — buka file di Word/LibreOffice untuk memeriksa tampilan.")

	return d.SaveFile(path)
}
