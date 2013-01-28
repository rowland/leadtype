// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "testing"
// import "fmt"

func TestNewRichText_English(t *testing.T) {
	fonts := testTtfFonts("Arial")
	rt, err := NewRichText("abc", fonts, 10, Options{"color": Green, "underline": true, "line_through": true})
	if err != nil {
		t.Fatal(err)
	}
	checkFatal(t, len(rt) == 1, "length of rt should be 1")
	expectS(t, "abc", rt[0].Text)
	expectF(t, 10, rt[0].FontSize)
	check(t, rt[0].Color == Green, "Color should be green.")
	check(t, rt[0].Underline, "Underline should be true.")
	check(t, rt[0].LineThrough, "LineThrough should be true.")
	check(t, rt[0].Font == fonts[0], "Should be tagged with Arial font.")
}

func TestNewRichText_EnglishAndChinese_Fail(t *testing.T) {
	fonts := testTtfFonts("Arial")
	_, err := NewRichText("abc所有测", fonts, 10, Options{})
	check(t, err != nil, "NewRichText should fail with Chinese text and only Arial.")
	expectS(t, "No font found for 所有测.", err.Error())
}

func TestNewRichText_EnglishAndChinese_Pass(t *testing.T) {
	fonts := testTtfFonts("Arial", "STSong")
	rt, err := NewRichText("abc所有测def", fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	checkFatal(t, len(rt) == 3, "length of rt should be 3")

	expectS(t, "abc", rt[0].Text)
	expectF(t, 10, rt[0].FontSize)
	check(t, rt[0].Color == Black, "Color should be (default) black.")
	check(t, !rt[0].Underline, "Underline should be (default) false.")
	check(t, !rt[0].LineThrough, "LineThrough should be (default) false.")

	expectS(t, "所有测", rt[1].Text)
	expectF(t, 10, rt[1].FontSize)
	check(t, rt[1].Color == Black, "Color should be (default) black.")
	check(t, !rt[1].Underline, "Underline should be (default) false.")
	check(t, !rt[1].LineThrough, "LineThrough should be (default) false.")

	expectS(t, "def", rt[2].Text)
	expectF(t, 10, rt[2].FontSize)
	check(t, rt[2].Color == Black, "Color should be (default) black.")
	check(t, !rt[2].Underline, "Underline should be (default) false.")
	check(t, !rt[2].LineThrough, "LineThrough should be (default) false.")

	check(t, rt[0].Font == fonts[0], "abc should be tagged with Arial font.")
	check(t, rt[1].Font == fonts[1], "Chinese should be tagged with STSong font.")
	check(t, rt[2].Font == fonts[0], "def should be tagged with Arial font.")
}

func TestNewRichText_ChineseAndEnglish(t *testing.T) {
	fonts := testTtfFonts("Arial", "STSong")
	rt, err := NewRichText("所有测abc", fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	checkFatal(t, len(rt) == 2, "length of rt should be 2")
	expectS(t, "所有测", rt[0].Text)
	expectF(t, 10, rt[0].FontSize)
	expectS(t, "abc", rt[1].Text)
	expectF(t, 10, rt[1].FontSize)
	check(t, rt[0].Font == fonts[1], "Should be tagged with Arial font.")
	check(t, rt[1].Font == fonts[0], "Should be tagged with STSong font.")
}

func TestNewRichText_EnglishRussianAndChineseLanguages(t *testing.T) {
	afmFonts := testAfmFonts("Helvetica")
	ttfFonts := testTtfFonts("Arial", "STSong")
	fonts := append(afmFonts, ttfFonts...)
	text := "Here is some Russian, Неприкосновенность, and some Chinese, 表明你已明确同意你的回答接受评估."
	rt, err := NewRichText(text, fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	// for i, p := range rt {
	// 	fmt.Println(i, p.Text)
	// }
	checkFatal(t, len(rt) == 19, "length of rt should be 19")
	expectS(t, "Here", rt[0].Text)
	expectF(t, 10, rt[0].FontSize)
	expectS(t, "Неприкосновенность", rt[8].Text)
	expectF(t, 10, rt[8].FontSize)
	expectS(t, ",", rt[9].Text)
	expectF(t, 10, rt[9].FontSize)
	expectS(t, "表明你已明确同意你的回答接受评估", rt[17].Text)
	expectF(t, 10, rt[17].FontSize)
	expectS(t, ".", rt[18].Text)
	expectF(t, 10, rt[18].FontSize)
	check(t, rt[0].Font == fonts[0], "Should be tagged with Helvetica font.")
	check(t, rt[8].Font == fonts[1], "Should be tagged with Arial font.")
	check(t, rt[11].Font == fonts[0], "Should be tagged with Helvetica font.")
	check(t, rt[17].Font == fonts[2], "Should be tagged with STSong font.")
	check(t, rt[18].Font == fonts[0], "Should be tagged with Helvetica font.")
}

// With Chinese font first in list, Arial is not called upon for English.
func TestNewRichText_ChineseAndEnglish_Reversed(t *testing.T) {
	fonts := testTtfFonts("STSong", "Arial")
	rt, err := NewRichText("所有测abc", fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	checkFatal(t, len(rt) == 1, "length of rt should be 1")
	expectS(t, "所有测abc", rt[0].Text)
	expectF(t, 10, rt[0].FontSize)
	check(t, rt[0].Font == fonts[0], "Should be tagged with Arial font.")
}

// 10,145 ns
// 10,987 ns after adding additional text attributes
// 55,649 ns after word breaking
// 76,497 ns after measuring
func BenchmarkNewRichText(b *testing.B) {
	b.StopTimer()
	afmFonts := testAfmFonts("Helvetica")
	ttfFonts := testTtfFonts("Arial", "STSong")
	fonts := append(afmFonts, ttfFonts...)
	text := "Here is some Russian, Неприкосновенность, and some Chinese, 表明你已明确同意你的回答接受评估."
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := NewRichText(text, fonts, 10, Options{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestTokenRegexp_Latin(t *testing.T) {
	const textToBreak = "Lorem ipsum dolor sit amet, consectetur adipisicing elit, \n" +
		"sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n"
	m := TokenRegexp.FindAllString(textToBreak, -1)
	expectS(t, "Lorem", m[0])
	expectS(t, " ", m[1])
	expectS(t, "ipsum", m[2])
	expectS(t, "amet,", m[8])
	expectS(t, "elit,", m[14])
	expectS(t, " ", m[15])
	expectS(t, "\n", m[16])
	expectS(t, "aliqua.", m[37])
	expectS(t, "\n", m[38])
}

func TestTokenRegexp_Mixed(t *testing.T) {
	const textToBreak = "Here is some Russian, Неприкосновенность, and some Chinese, 表明你已明确同意你的回答接受评估."
	m := TokenRegexp.FindAllString(textToBreak, -1)
	expectS(t, "Here", m[0])
	expectS(t, " ", m[1])
	expectS(t, "Неприкосновенность,", m[8])
	expectS(t, "表明你已明确同意你的回答接受评估.", m[16])
}

func TestRichText_Merge(t *testing.T) {
	afmFonts := testAfmFonts("Helvetica")
	ttfFonts := testTtfFonts("Arial", "STSong")
	fonts := append(afmFonts, ttfFonts...)
	text := "Here is some Russian, Неприкосновенность, and some Chinese, 表明你已明确同意你的回答接受评估."
	original, err := NewRichText(text, fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	piece0 := *original[0]
	merged := original.Merge()
	// for i, p := range merged {
	// 	fmt.Println(i, p.Text)
	// }
	checkFatal(t, len(merged) == 5, "length of merged should be 5")

	expectS(t, "Here is some Russian, ", merged[0].Text)
	expectF(t, 10, merged[0].FontSize)
	check(t, merged[0].Font == fonts[0], "Should be tagged with Helvetica font.")

	expectS(t, "Неприкосновенность", merged[1].Text)
	expectF(t, 10, merged[1].FontSize)
	check(t, merged[1].Font == fonts[1], "Should be tagged with Arial font.")

	expectS(t, ", and some Chinese, ", merged[2].Text)
	expectF(t, 10, merged[2].FontSize)
	check(t, merged[2].Font == fonts[0], "Should be tagged with Helvetica font.")

	expectS(t, "表明你已明确同意你的回答接受评估", merged[3].Text)
	expectF(t, 10, merged[3].FontSize)
	check(t, merged[3].Font == fonts[2], "Should be tagged with STSong font.")

	expectS(t, ".", merged[4].Text)
	expectF(t, 10, merged[4].FontSize)
	check(t, merged[4].Font == fonts[0], "Should be tagged with Helvetica font.")

	check(t, *original[0] == piece0, "Original should be unchanged.")
}
