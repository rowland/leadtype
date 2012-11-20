// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "testing"

func testTtfFonts(families ...string) (fonts []*Font) {
	fc, err := NewTtfFontCollection("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	for _, family := range families {
		metrics, err := fc.Select(family, "", "", nil)
		if err != nil {
			panic(err)
		}
		font := &Font{family: family, metrics: metrics}
		fonts = append(fonts, font)
	}
	return
}

func testAfmFonts(families ...string) (fonts []*Font) {
	fc, err := NewAfmFontCollection("../afm/data/fonts/*.afm")
	if err != nil {
		panic(err)
	}
	for _, family := range families {
		metrics, err := fc.Select(family, "", "", nil)
		if err != nil {
			panic(err)
		}
		font := &Font{family: family, metrics: metrics}
		fonts = append(fonts, font)
	}
	return
}

func TestNewRichText_English(t *testing.T) {
	fonts := testTtfFonts("Arial")
	rt, err := NewRichText("abc", fonts)
	if err != nil {
		t.Fatal(err)
	}
	checkFatal(t, len(rt) == 1, "length of rt should be 1")
	expectS(t, "abc", rt[0].Text)
	check(t, rt[0].Font == fonts[0], "Should be tagged with Arial font.")
}

func TestNewRichText_EnglishAndChinese_Fail(t *testing.T) {
	fonts := testTtfFonts("Arial")
	_, err := NewRichText("abc所有测", fonts)
	check(t, err != nil, "NewRichText should fail with Chinese text and only Arial.")
	expectS(t, "No font found for 所有测.", err.Error())
}

func TestNewRichText_EnglishAndChinese_Pass(t *testing.T) {
	fonts := testTtfFonts("Arial", "STSong")
	rt, err := NewRichText("abc所有测def", fonts)
	if err != nil {
		t.Fatal(err)
	}
	checkFatal(t, len(rt) == 3, "length of rt should be 3")
	expectS(t, "abc", rt[0].Text)
	expectS(t, "所有测", rt[1].Text)
	expectS(t, "def", rt[2].Text)
	check(t, rt[0].Font == fonts[0], "abc should be tagged with Arial font.")
	check(t, rt[1].Font == fonts[1], "Chinese should be tagged with STSong font.")
	check(t, rt[2].Font == fonts[0], "def should be tagged with Arial font.")
}

func TestNewRichText_ChineseAndEnglish(t *testing.T) {
	fonts := testTtfFonts("Arial", "STSong")
	rt, err := NewRichText("所有测abc", fonts)
	if err != nil {
		t.Fatal(err)
	}
	checkFatal(t, len(rt) == 2, "length of rt should be 2")
	expectS(t, "所有测", rt[0].Text)
	expectS(t, "abc", rt[1].Text)
	check(t, rt[0].Font == fonts[1], "Should be tagged with Arial font.")
	check(t, rt[1].Font == fonts[0], "Should be tagged with STSong font.")
}

func TestNewRichText_EnglishRussianAndChineseLanguages(t *testing.T) {
	afmFonts := testAfmFonts("Helvetica")
	ttfFonts := testTtfFonts("Arial", "STSong")
	fonts := append(afmFonts, ttfFonts...)
	text := "Here is some Russian, Неприкосновенность, and some Chinese, 表明你已明确同意你的回答接受评估."
	rt, err := NewRichText(text, fonts)
	if err != nil {
		t.Fatal(err)
	}
	checkFatal(t, len(rt) == 5, "length of rt should be 5")
	expectS(t, "Here is some Russian, ", rt[0].Text)
	expectS(t, "Неприкосновенность", rt[1].Text)
	expectS(t, ", and some Chinese, ", rt[2].Text)
	expectS(t, "表明你已明确同意你的回答接受评估", rt[3].Text)
	expectS(t, ".", rt[4].Text)
	check(t, rt[0].Font == fonts[0], "Should be tagged with Helvetica font.")
	check(t, rt[1].Font == fonts[1], "Should be tagged with Arial font.")
	check(t, rt[2].Font == fonts[0], "Should be tagged with Helvetica font.")
	check(t, rt[3].Font == fonts[2], "Should be tagged with STSong font.")
	check(t, rt[4].Font == fonts[0], "Should be tagged with Helvetica font.")
}

// With Chinese font first in list, Arial is not called upon for English.
func TestNewRichText_ChineseAndEnglish_Reversed(t *testing.T) {
	fonts := testTtfFonts("STSong", "Arial")
	rt, err := NewRichText("所有测abc", fonts)
	if err != nil {
		t.Fatal(err)
	}
	checkFatal(t, len(rt) == 1, "length of rt should be 1")
	expectS(t, "所有测abc", rt[0].Text)
	check(t, rt[0].Font == fonts[0], "Should be tagged with Arial font.")
}

// 10,145 ns
func BenchmarkNewRichText(b *testing.B) {
	b.StopTimer()
	afmFonts := testAfmFonts("Helvetica")
	ttfFonts := testTtfFonts("Arial", "STSong")
	fonts := append(afmFonts, ttfFonts...)
	text := "Here is some Russian, Неприкосновенность, and some Chinese, 表明你已明确同意你的回答接受评估."
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := NewRichText(text, fonts)
		if err != nil {
			b.Fatal(err)
		}
	}
}
