package ooxml

import (
	"errors"
	"testing"
)

func TestNormalizePartName_valid(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{`word/document.xml`, `/word/document.xml`},
		{`/word/document.xml`, `/word/document.xml`},
		{`/word/./document.xml`, `/word/document.xml`},
		{`//word//document.xml`, `/word/document.xml`},
		{`\windows\path`, `/windows/path`},
		{`/`, `/`},
		{`/./`, `/`},
		{`/word`, `/word`},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := NormalizePartName(tc.in)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestNormalizePartName_traversal(t *testing.T) {
	for _, in := range []string{
		`../foo`,
		`/word/../etc/x`,
		`/..`,
		`/../foo`,
		`/word/../../x`,
	} {
		t.Run(in, func(t *testing.T) {
			_, err := NormalizePartName(in)
			if !errors.Is(err, ErrPathTraversal) {
				t.Fatalf("want ErrPathTraversal, got %v", err)
			}
		})
	}
}

func TestNormalizePartName_invalid(t *testing.T) {
	_, err := NormalizePartName("")
	if !errors.Is(err, ErrInvalidPartName) {
		t.Fatalf("empty: want ErrInvalidPartName, got %v", err)
	}
	_, err = NormalizePartName("   ")
	if !errors.Is(err, ErrInvalidPartName) {
		t.Fatalf("whitespace: want ErrInvalidPartName, got %v", err)
	}
	_, err = NormalizePartName("/foo\x00bar")
	if !errors.Is(err, ErrInvalidPartName) {
		t.Fatalf("nul: want ErrInvalidPartName, got %v", err)
	}
}

func TestZipEntryName(t *testing.T) {
	got, err := ZipEntryName("/word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if got != `word/document.xml` {
		t.Fatalf("got %q", got)
	}
	_, err = ZipEntryName("../x")
	if !errors.Is(err, ErrPathTraversal) {
		t.Fatalf("got %v", err)
	}
}

func TestRelationshipBaseDir(t *testing.T) {
	got, err := RelationshipBaseDir("/_rels/.rels")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/" {
		t.Fatalf("root rels base: got %q", got)
	}
	got, err = RelationshipBaseDir("/word/_rels/document.xml.rels")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/word" {
		t.Fatalf("word rels base: got %q want /word", got)
	}
	_, err = RelationshipBaseDir("/word/../_rels/.rels")
	if !errors.Is(err, ErrPathTraversal) {
		t.Fatalf("want ErrPathTraversal, got %v", err)
	}
}
