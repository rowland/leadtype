// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf_fonts

import (
	"testing"

	"github.com/rowland/leadtype/ttf"
)

type ttfFontSelection struct {
	family         string
	weight         string
	style          string
	ranges         []string
	postscriptName string
}

var testTtfSelectData = []ttfFontSelection{
	{"Arial", "", "", nil, "ArialMT"},
	{"Arial", "", "Italic", nil, "Arial-ItalicMT"},
	{"Arial", "Bold", "", nil, "Arial-BoldMT"},
	{"Arial", "Bold", "Italic", nil, "Arial-BoldItalicMT"},
	{"arial", "bold", "italic", nil, "Arial-BoldItalicMT"},

	{"Courier New", "", "", nil, "CourierNewPSMT"},
	{"Courier New", "", "Italic", nil, "CourierNewPS-ItalicMT"},
	{"Courier New", "Bold", "", nil, "CourierNewPS-BoldMT"},
	{"Courier New", "Bold", "Italic", nil, "CourierNewPS-BoldItalicMT"},
	{"courier new", "bold", "italic", nil, "CourierNewPS-BoldItalicMT"},

	{"Times New Roman", "", "", nil, "TimesNewRomanPSMT"},
	{"Times New Roman", "", "Italic", nil, "TimesNewRomanPS-ItalicMT"},
	{"Times New Roman", "Bold", "", nil, "TimesNewRomanPS-BoldMT"},
	{"Times New Roman", "Bold", "Italic", nil, "TimesNewRomanPS-BoldItalicMT"},
	{"times new roman", "bold", "italic", nil, "TimesNewRomanPS-BoldItalicMT"},

	{"Arial Unicode MS", "", "", []string{"CJK Unified Ideographs"}, "ArialUnicodeMS"},
	{"AppleMyungjo", "Regular", "", nil, "AppleMyungjo"},
}

func TestTtfFonts(t *testing.T) {
	var fc TtfFonts

	if err := fc.Add("/Library/Fonts/*.ttf"); err != nil {
		t.Skip("no TTF fonts available at /Library/Fonts")
	}
	if fc.Len() == 0 {
		t.Skip("no TTF fonts available at /Library/Fonts")
	}

	if expected, actual := 85, fc.Len(); actual != expected {
		t.Skipf("expected %v fonts, got %v — skipping (system font set differs)", expected, actual)
	}
	for _, fs := range testTtfSelectData {
		f, err := fc.Select(fs.family, fs.weight, fs.style, fs.ranges)
		if err == nil {
			if f.PostScriptName() != fs.postscriptName {
				t.Errorf("%s: expected %v, got %v", fs.postscriptName, fs.postscriptName, f.PostScriptName())
			}
		} else {
			t.Error(err)
		}
	}
	bogusFont, err2 := fc.Select("Bogus", "Regular", "", nil)
	if bogusFont != nil {
		t.Errorf("%s: expected nil, got %v", "Bogus Select Font", bogusFont)
	}
	const expectedError = "Font Bogus Regular not found"
	if err2.Error() != expectedError {
		t.Errorf("%s: expected %v, got %v", "Bogus Select Error", expectedError, err2.Error())
	}
}

// 81,980,000 ns
// 45,763,220 ns
// 44,562,080 ns
// 37,788,928 ns go1.1.1
// 33,604,903 ns go1.1.2
// 29,678,826 ns go1.2.1
// 33,674,155 ns go1.4.2
// 13,663,640 ns go1.7.3 mbp
func BenchmarkTtfFontCollection_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var fonts TtfFonts
		fonts.Add("/Library/Fonts/*.ttf")
	}
}

// 3,132 ns
// 1,938 ns go1.1.1
// 1,769 ns go1.1.2
// 1,610 ns go1.2.1
// 1,855 ns go1.4.2
//	729 ns go1.7.3 mbp

func BenchmarkTtfFontCollection_Select(b *testing.B) {
	b.StopTimer()
	var fonts TtfFonts
	if err := fonts.Add("/Library/Fonts/*.ttf"); err != nil {
		b.Fatal(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		fonts.Select("Times New Roman", "Bold", "Italic", nil)
	}
}

func TestNewFromSystemFonts_CachesInventory(t *testing.T) {
	ClearCache()
	originalLoader := loadSystemFontInfos
	defer func() {
		loadSystemFontInfos = originalLoader
		ClearCache()
	}()

	calls := 0
	info := &ttf.FontInfo{}
	loadSystemFontInfos = func() ([]*ttf.FontInfo, error) {
		calls++
		return []*ttf.FontInfo{info}, nil
	}

	fc1, err := NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}
	fc2, err := NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}

	if calls != 1 {
		t.Fatalf("expected one system font scan, got %d", calls)
	}
	if len(fc1.FontInfos) != 1 || fc1.FontInfos[0] != info {
		t.Fatalf("unexpected first font inventory: %#v", fc1.FontInfos)
	}
	if len(fc2.FontInfos) != 1 || fc2.FontInfos[0] != info {
		t.Fatalf("unexpected second font inventory: %#v", fc2.FontInfos)
	}

	fc1.FontInfos[0] = nil
	if fc2.FontInfos[0] != info {
		t.Fatal("expected each TtfFonts instance to receive its own FontInfos slice")
	}
}

func TestClearCache_ForcesRescan(t *testing.T) {
	ClearCache()
	originalLoader := loadSystemFontInfos
	defer func() {
		loadSystemFontInfos = originalLoader
		ClearCache()
	}()

	calls := 0
	first := &ttf.FontInfo{}
	second := &ttf.FontInfo{}
	loadSystemFontInfos = func() ([]*ttf.FontInfo, error) {
		calls++
		if calls == 1 {
			return []*ttf.FontInfo{first}, nil
		}
		return []*ttf.FontInfo{second}, nil
	}

	fc1, err := NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}
	ClearCache()
	fc2, err := NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}

	if calls != 2 {
		t.Fatalf("expected ClearCache to force a rescan, got %d scans", calls)
	}
	if fc1.FontInfos[0] != first {
		t.Fatal("expected first cached inventory before ClearCache")
	}
	if fc2.FontInfos[0] != second {
		t.Fatal("expected fresh inventory after ClearCache")
	}
}
