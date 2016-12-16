// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
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

	{"Zapf Chancery", "Medium", "Italic", nil, "ZapfChancery-MediumItalic"},
	{"zapf chancery", "medium", "italic", nil, "ZapfChancery-MediumItalic"},
}

func testAfmFonts(families ...string) (fonts []*font.Font) {
	fc, err := NewAfmFontCollection("../afm/data/fonts/*.afm")
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

func TestAfmFontCollection(t *testing.T) {
	var fc AfmFontCollection

	if err := fc.Add("../afm/data/fonts/*.afm"); err != nil {
		t.Fatal(err)
	}

	expectNI(t, "Len", 77, fc.Len())
	for _, fs := range testAfmSelectData {
		f, err := fc.Select(fs.family, fs.weight, fs.style, fs.ranges)
		if err == nil {
			expectNS(t, fs.postscriptName, fs.postscriptName, f.PostScriptName())
		} else {
			t.Error(err)
		}
	}
	bogusFont, err2 := fc.Select("Bogus", "Medium", "", nil)
	expect(t, "Bogus Select Font", bogusFont == nil)
	expectNS(t, "Bogus Select Error", "Font 'Bogus Medium ' not found", err2.Error())
}

// 81,980,000 ns
// 45,763,220 ns
// 13,754,000 ns
// 12,035,539 ns go1.1.1
// 11,984,518 ns go1.1.2
// 10,951,804 ns go1.2.1
// 13,275,334 ns go1.4.2
//  4,440,591 ns go1.7.3 mbp
func BenchmarkAfmFontCollection_Add(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var fc AfmFontCollection
		if err := fc.Add("../afm/data/fonts/*.afm"); err != nil {
			b.Fatal(err)
		}
		if fc.Len() != 77 {
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
func BenchmarkAfmFontCollection_Select(b *testing.B) {
	b.StopTimer()
	var fc AfmFontCollection
	if err := fc.Add("../afm/data/fonts/*.afm"); err != nil {
		b.Fatal(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		fc.Select("Times", "Bold", "Italic", nil)
	}
}
