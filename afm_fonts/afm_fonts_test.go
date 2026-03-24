// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package afm_fonts

import (
	"strings"
	"testing"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
)

type afmFontSelection struct {
	family         string
	weight         string
	style          string
	ranges         []string
	postscriptName string
}

var testAfmSelectData = []afmFontSelection{
	{"Helvetica", "", "", nil, "Helvetica"},
	{"Helvetica", "", "Oblique", nil, "Helvetica-Oblique"},
	{"Helvetica", "", "oblique", nil, "Helvetica-Oblique"},
	{"Helvetica", "", "Italic", nil, "Helvetica-Oblique"},
	{"Helvetica", "", "italic", nil, "Helvetica-Oblique"},
	{"Helvetica", "Bold", "", nil, "Helvetica-Bold"},
	{"Helvetica", "bold", "", nil, "Helvetica-Bold"},
	{"Helvetica", "Bold", "Italic", nil, "Helvetica-BoldOblique"},
	{"helvetica", "bold", "italic", nil, "Helvetica-BoldOblique"},

	{"Courier", "", "", nil, "Courier"},
	{"Courier", "", "Oblique", nil, "Courier-Oblique"},
	{"Courier", "", "oblique", nil, "Courier-Oblique"},
	{"Courier", "", "Italic", nil, "Courier-Oblique"},
	{"Courier", "", "italic", nil, "Courier-Oblique"},
	{"Courier", "Bold", "", nil, "Courier-Bold"},
	{"Courier", "bold", "", nil, "Courier-Bold"},
	{"Courier", "Bold", "Italic", nil, "Courier-BoldOblique"},
	{"courier", "bold", "italic", nil, "Courier-BoldOblique"},

	{"Times", "", "", nil, "Times-Roman"},
	{"Times", "", "Italic", nil, "Times-Italic"},
	{"Times", "", "italic", nil, "Times-Italic"},
	{"Times", "", "Oblique", nil, "Times-Italic"},
	{"Times", "", "oblique", nil, "Times-Italic"},
	{"Times", "Bold", "", nil, "Times-Bold"},
	{"Times", "bold", "", nil, "Times-Bold"},
	{"Times", "Bold", "Italic", nil, "Times-BoldItalic"},
	{"times", "bold", "italic", nil, "Times-BoldItalic"},

	{"ZapfDingbats", "", "", nil, "ZapfDingbats"},
	{"zapfdingbats", "", "", nil, "ZapfDingbats"},
	{"Symbol", "", "", nil, "Symbol"},
	{"symbol", "", "", nil, "Symbol"},
}

func testAfmFonts(families ...string) (fonts []*font.Font) {
	fc, err := Default()
	if err != nil {
		panic(err)
	}
	for _, family := range families {
		f, err := font.New(family, options.Options{}, font.FontSources{fc})
		if err != nil {
			panic(err)
		}
		fonts = append(fonts, f)
	}
	return
}

func assertSelectCases(t *testing.T, fc *AfmFonts) {
	t.Helper()
	for _, fs := range testAfmSelectData {
		f, err := fc.Select(fs.family, fs.weight, fs.style, fs.ranges)
		if err == nil {
			if f.PostScriptName() != fs.postscriptName {
				t.Errorf("%s: expected %s, got %s", "PostScriptName", fs.postscriptName, f.PostScriptName())
			}
		} else {
			t.Error(err)
		}
	}
}

func assertBogusSelect(t *testing.T, fc *AfmFonts) {
	t.Helper()

	bogusFont, err2 := fc.Select("Bogus", "Medium", "", nil)
	if bogusFont != nil {
		t.Errorf("bogusFont: expected nil, got %v", bogusFont)
	}
	const expectedError = "Font 'Bogus Medium ' not found"
	if err2.Error() != expectedError {
		t.Errorf("%s: expected %s, got %s", "Bogus Select Error", expectedError, err2.Error())
	}
}

func TestDefault(t *testing.T) {
	fc, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if fc == nil {
		t.Fatal("Default returned nil collection")
	}
	if fc.Len() != 14 {
		t.Fatalf("Len: expected %d, got %d", 14, fc.Len())
	}
	for _, fi := range fc.FontInfos {
		if !strings.HasPrefix(fi.Filename(), "data/fonts/") {
			t.Fatalf("embedded filename should be rooted at data/fonts/, got %q", fi.Filename())
		}
	}
	assertSelectCases(t, fc)
	assertBogusSelect(t, fc)
}

func TestAfmFontCollection_Add(t *testing.T) {
	var fc AfmFonts

	if err := fc.Add("../afm/data/fonts/*.afm"); err != nil {
		t.Fatal(err)
	}

	expected, actual := 77, fc.Len()
	expected = 14
	if actual != expected {
		t.Errorf("%s: expected %d, got %d", "Len", expected, actual)
	}
	assertSelectCases(t, &fc)
	assertBogusSelect(t, &fc)
}

func TestAfmFonts_AddDefault(t *testing.T) {
	var fc AfmFonts
	if err := fc.AddDefault(); err != nil {
		t.Fatal(err)
	}
	if fc.Len() != 14 {
		t.Fatalf("Len: expected %d, got %d", 14, fc.Len())
	}
}

func TestFamilies(t *testing.T) {
	fonts := Families("Helvetica", "Times")
	if len(fonts) != 2 {
		t.Fatalf("Families length: expected %d, got %d", 2, len(fonts))
	}
	if fonts[0].PostScriptName() != "Helvetica" {
		t.Fatalf("Families[0]: expected %q, got %q", "Helvetica", fonts[0].PostScriptName())
	}
	if fonts[1].PostScriptName() != "Times-Roman" {
		t.Fatalf("Families[1]: expected %q, got %q", "Times-Roman", fonts[1].PostScriptName())
	}
}

// 81,980,000 ns
// 45,763,220 ns
// 13,754,000 ns
// 12,035,539 ns go1.1.1
// 11,984,518 ns go1.1.2
// 10,951,804 ns go1.2.1
// 13,275,334 ns go1.4.2
//	4,440,591 ns go1.7.3 mbp
//  1,460,074 ns go1.25.5 M1

func BenchmarkAfmFontCollection_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var fc AfmFonts
		if err := fc.AddDefault(); err != nil {
			b.Fatal(err)
		}
		if fc.Len() != 14 {
			b.Fatalf("Unexpected number of fonts found: %d.", fc.Len())
		}
	}
}

// 44,797 ns
// 40,640 ns go1.1.1
// 41,299 ns go1.1.2
// 39,918 ns go1.2.1
// 99,394 ns go1.4.2
// 31,638 ns go1.7.3 mbp
//  7,312 ns go1.25.5 M1

func BenchmarkAfmFontCollection_Select(b *testing.B) {
	b.StopTimer()
	fc, err := Default()
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		fc.Select("Times", "Bold", "Italic", nil)
	}
}
