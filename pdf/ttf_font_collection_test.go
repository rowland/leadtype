// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"testing"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
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

func testTtfFonts(families ...string) (fonts []*font.Font) {
	fc, err := NewTtfFontCollection("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	for _, family := range families {
		if family == "" {
			fonts = append(fonts, nil)
			continue
		}
		f, err := font.New(family, options.Options{}, font.FontSources{fc})
		if err != nil {
			panic(err)
		}
		fonts = append(fonts, f)
	}
	return
}

func TestTtfFontCollection(t *testing.T) {
	var fc TtfFontCollection

	if err := fc.Add("/Library/Fonts/*.ttf"); err != nil {
		t.Error(err)
	}

	expectNI(t, "Len", 85, fc.Len())
	for _, fs := range testTtfSelectData {
		f, err := fc.Select(fs.family, fs.weight, fs.style, fs.ranges)
		if err == nil {
			expectNS(t, fs.postscriptName, fs.postscriptName, f.PostScriptName())
		} else {
			t.Error(err)
		}
	}
	bogusFont, err2 := fc.Select("Bogus", "Regular", "", nil)
	expect(t, "Bogus Select Font", bogusFont == nil)
	expectNS(t, "Bogus Select Error", "Font Bogus Regular not found", err2.Error())
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
		var fc TtfFontCollection
		fc.Add("/Library/Fonts/*.ttf")
	}
}

// 3,132 ns
// 1,938 ns go1.1.1
// 1,769 ns go1.1.2
// 1,610 ns go1.2.1
// 1,855 ns go1.4.2
//   729 ns go1.7.3 mbp
func BenchmarkTtfFontCollection_Select(b *testing.B) {
	b.StopTimer()
	var fc TtfFontCollection
	if err := fc.Add("/Library/Fonts/*.ttf"); err != nil {
		b.Fatal(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		fc.Select("Times New Roman", "Bold", "Italic", nil)
	}
}
