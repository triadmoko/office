package ooxml

import (
	"net/url"
	"reflect"
	"testing"
)

func TestNamespaceConstantsNonEmpty(t *testing.T) {
	uris := []struct {
		name string
		val  string
	}{
		{"NSContentTypes", NSContentTypes},
		{"NSRelationships", NSRelationships},
		{"NSRelOfficeDocument", NSRelOfficeDocument},
		{"NSWordprocessingML", NSWordprocessingML},
		{"NSWordprocessingML14", NSWordprocessingML14},
		{"NSWordprocessingML15", NSWordprocessingML15},
		{"NSWordprocessingDrawing", NSWordprocessingDrawing},
		{"NSWordprocessingDrawing14", NSWordprocessingDrawing14},
		{"NSSpreadsheetML", NSSpreadsheetML},
		{"NSSpreadsheetMLRev", NSSpreadsheetMLRev},
		{"NSSpreadsheetMLRev2", NSSpreadsheetMLRev2},
		{"NSSpreadsheetMLRev3", NSSpreadsheetMLRev3},
		{"NSMarkupCompat", NSMarkupCompat},
		{"NSPresentationML", NSPresentationML},
		{"NSPresentationML14", NSPresentationML14},
		{"NSPresentationML15", NSPresentationML15},
		{"NSDrawingML", NSDrawingML},
		{"NSDrawingML14", NSDrawingML14},
		{"NSPicture", NSPicture},
		{"NSRelMarkup", NSRelMarkup},
	}
	for _, tc := range uris {
		if tc.val == "" {
			t.Errorf("namespace constant %s is empty", tc.name)
		}
		if _, err := url.Parse(tc.val); err != nil {
			t.Errorf("namespace constant %s = %q is not a valid URI: %v", tc.name, tc.val, err)
		}
	}
}

func TestPrefixConstantsNonEmpty(t *testing.T) {
	prefixes := []struct {
		name string
		val  string
	}{
		{"PrefixWordprocessingML", PrefixWordprocessingML},
		{"PrefixWordprocessingML14", PrefixWordprocessingML14},
		{"PrefixWordprocessingML15", PrefixWordprocessingML15},
		{"PrefixWordprocessingDrawing", PrefixWordprocessingDrawing},
		{"PrefixWordprocessingDrawing14", PrefixWordprocessingDrawing14},
		{"PrefixSpreadsheetML", PrefixSpreadsheetML},
		{"PrefixSpreadsheetRev", PrefixSpreadsheetRev},
		{"PrefixSpreadsheetRev2", PrefixSpreadsheetRev2},
		{"PrefixSpreadsheetRev3", PrefixSpreadsheetRev3},
		{"PrefixMarkupCompat", PrefixMarkupCompat},
		{"PrefixPresentationML", PrefixPresentationML},
		{"PrefixPresentationML14", PrefixPresentationML14},
		{"PrefixPresentationML15", PrefixPresentationML15},
		{"PrefixDrawingML", PrefixDrawingML},
		{"PrefixDrawingML14", PrefixDrawingML14},
		{"PrefixPicture", PrefixPicture},
		{"PrefixRelMarkup", PrefixRelMarkup},
	}
	for _, tc := range prefixes {
		if tc.val == "" {
			t.Errorf("prefix constant %s is empty", tc.name)
		}
	}
}

func TestContentTypeConstantsNonEmpty(t *testing.T) {
	cts := []struct {
		name string
		val  string
	}{
		{"CTRelsXML", CTRelsXML},
		{"CTCoreProps", CTCoreProps},
		{"CTWordDocumentMain", CTWordDocumentMain},
		{"CTWordStyles", CTWordStyles},
		{"CTWordNumbering", CTWordNumbering},
		{"CTWordSettings", CTWordSettings},
		{"CTWordFontTable", CTWordFontTable},
		{"CTWordWebSettings", CTWordWebSettings},
		{"CTWordFooter", CTWordFooter},
		{"CTSpreadsheetMain", CTSpreadsheetMain},
		{"CTPresentationMain", CTPresentationMain},
		{"CTTheme", CTTheme},
		{"CTAppProps", CTAppProps},
		{"CTImagePNG", CTImagePNG},
		{"CTImageJPEG", CTImageJPEG},
		{"CTImageGIF", CTImageGIF},
		{"CTImageBMP", CTImageBMP},
	}
	for _, tc := range cts {
		if tc.val == "" {
			t.Errorf("content type constant %s is empty", tc.name)
		}
	}
}

func TestExtensionToContentType(t *testing.T) {
	required := []string{"xml", "rels", "png", "jpg", "jpeg", "gif", "bmp", "bin", "vml"}
	for _, ext := range required {
		if ct, ok := ExtensionToContentType[ext]; !ok || ct == "" {
			t.Errorf("ExtensionToContentType missing or empty entry for %q", ext)
		}
	}
	// No duplicate values for the same extension (map key uniqueness is guaranteed by Go).
	_ = reflect.TypeFor[map[string]string]() // sanity
}
