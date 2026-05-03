package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/triadmoko/office/docx"
)

func main() {
	out := flag.String("write-docx", "", "write a minimal .docx with given text to path")
	text := flag.String("text", "Hello from office", "text for -write-docx")
	flag.Parse()

	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		if err := docx.WriteMinimal(f, *text); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("wrote", *out)
		return
	}

	fmt.Println("office — OOXML tooling (pure Go stdlib)")
	fmt.Println("Usage: office -write-docx out.docx [-text \"...\"]")
}
