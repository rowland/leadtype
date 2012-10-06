// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import "testing"

type fontSelection struct {
	family         string
	weight         string
	style          string
	ranges         []string
	postscriptName string
}

var testSelectData = []fontSelection{
	{"Arial", "", "", nil, "ArialMT"},
	{"Arial", "", "Italic", nil, "Arial-ItalicMT"},
	{"Arial", "Bold", "", nil, "Arial-BoldMT"},
	{"Arial", "Bold", "Italic", nil, "Arial-BoldItalicMT"},

	{"Courier New", "", "", nil, "CourierNewPSMT"},
	{"Courier New", "", "Italic", nil, "CourierNewPS-ItalicMT"},
	{"Courier New", "Bold", "", nil, "CourierNewPS-BoldMT"},
	{"Courier New", "Bold", "Italic", nil, "CourierNewPS-BoldItalicMT"},

	{"Times New Roman", "", "", nil, "TimesNewRomanPSMT"},
	{"Times New Roman", "", "Italic", nil, "TimesNewRomanPS-ItalicMT"},
	{"Times New Roman", "Bold", "", nil, "TimesNewRomanPS-BoldMT"},
	{"Times New Roman", "Bold", "Italic", nil, "TimesNewRomanPS-BoldItalicMT"},

	{"Arial Unicode MS", "", "", []string{"CJK Unified Ideographs"}, "ArialUnicodeMS"},
	{"Myanmar Sangam MN", "", "", []string{"Myanmar"}, "MyanmarSangamMN"},
}

func TestFontCollection(t *testing.T) {
	var fc FontCollection

	if err := fc.Add("/Library/Fonts/*.ttf"); err != nil {
		t.Error(err)
	}

	expectI(t, "Len", 114, fc.Len())
	for _, fs := range testSelectData {
		f, err := fc.Select(fs.family, fs.weight, fs.style, fs.ranges)
		if err == nil {
			expectS(t, fs.postscriptName, fs.postscriptName, f.PostScriptName())
		} else {
			t.Error(err)
		}
	}
	bogusFont, err2 := fc.Select("Bogus", "Regular", "", nil)
	expect(t, "Bogus Select Font", bogusFont == nil)
	expectS(t, "Bogus Select Error", "Font Bogus Regular not found", err2.Error())
}

// 81,980,000 ns
// 45,763,220 ns
func BenchmarkFontCollection_Add(b *testing.B) {
	var fc FontCollection

	for i := 0; i < b.N; i++ {
		fc.Add("/Library/Fonts/*.ttf")
	}
}
