// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "testing"

// import "fmt"

func textPieceTestText() *TextPiece {
	font := testTtfFonts("Arial")[0]
	return &TextPiece{
		Text:     "Lorem",
		Font:     font,
		FontSize: 10,
	}
}

func TestNewTextPiece_English(t *testing.T) {
	st := SuperTest{t}
	fonts := testTtfFonts("Arial")
	rt, err := NewTextPiece("abc", fonts, 10, Options{"color": Green, "underline": true, "line_through": true})
	if err != nil {
		t.Fatal(err)
	}
	st.Equal("abc", rt.Text)
	st.Equal(10.0, rt.FontSize)
	st.Equal(Green, rt.Color)
	st.True(rt.Underline)
	st.True(rt.LineThrough)
	st.Equal(fonts[0], rt.Font, "Should be tagged with Arial font.")
}

func TestNewTextPiece_EnglishAndChinese_Fail(t *testing.T) {
	st := SuperTest{t}
	fonts := testTtfFonts("Arial")
	_, err := NewTextPiece("abc所有测", fonts, 10, Options{})
	st.False(err == nil, "NewRichText should fail with Chinese text and only Arial.")
	st.Equal("No font found for 所有测.", err.Error())
}

func TestNewTextPiece_EnglishAndChinese_Pass(t *testing.T) {
	st := SuperTest{t}
	fonts := testTtfFonts("Arial", "STSong")
	rt, err := NewTextPiece("abc所有测def", fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	st.Equal(3, len(rt.Pieces))

	st.Equal("abc", rt.Pieces[0].Text)
	st.Equal(10.0, rt.Pieces[0].FontSize)
	st.Equal(Black, rt.Pieces[0].Color, "Color should be (default) black.")
	st.False(rt.Pieces[0].Underline, "Underline should be (default) false.")
	st.False(rt.Pieces[0].LineThrough, "LineThrough should be (default) false.")

	st.Equal("所有测", rt.Pieces[1].Text)
	st.Equal(10.0, rt.Pieces[1].FontSize)
	st.Equal(Black, rt.Pieces[1].Color, "Color should be (default) black.")
	st.False(rt.Pieces[1].Underline, "Underline should be (default) false.")
	st.False(rt.Pieces[1].LineThrough, "LineThrough should be (default) false.")

	st.Equal("def", rt.Pieces[2].Text)
	st.Equal(10.0, rt.Pieces[2].FontSize)
	st.Equal(Black, rt.Pieces[2].Color, "Color should be (default) black.")
	st.False(rt.Pieces[2].Underline, "Underline should be (default) false.")
	st.False(rt.Pieces[2].LineThrough, "LineThrough should be (default) false.")

	st.Equal(fonts[0], rt.Pieces[0].Font, "abc should be tagged with Arial font.")
	st.Equal(fonts[1], rt.Pieces[1].Font, "Chinese should be tagged with STSong font.")
	st.Equal(fonts[0], rt.Pieces[2].Font, "def should be tagged with Arial font.")
}

func TestNewTextPiece_ChineseAndEnglish(t *testing.T) {
	st := SuperTest{t}
	fonts := testTtfFonts("Arial", "STSong")
	rt, err := NewTextPiece("所有测abc", fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	st.MustEqual(2, len(rt.Pieces))
	st.Equal("所有测", rt.Pieces[0].Text)
	st.Equal(10.0, rt.Pieces[0].FontSize)
	st.Equal("abc", rt.Pieces[1].Text)
	st.Equal(10.0, rt.Pieces[1].FontSize)
	st.Equal(fonts[1], rt.Pieces[0].Font, "Should be tagged with Arial font.")
	st.Equal(fonts[0], rt.Pieces[1].Font, "Should be tagged with STSong font.")
}

func TestNewTextPiece_EnglishRussianAndChineseLanguages(t *testing.T) {
	st := SuperTest{t}
	afmFonts := testAfmFonts("Helvetica")
	ttfFonts := testTtfFonts("Arial", "STSong")
	fonts := append(afmFonts, ttfFonts...)
	text := "Here is some Russian, Неприкосновенность, and some Chinese, 表明你已明确同意你的回答接受评估."
	rt, err := NewTextPiece(text, fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	st.MustEqual(5, len(rt.Pieces))
	st.Equal("Here is some Russian, ", rt.Pieces[0].Text)
	st.Equal(10.0, rt.Pieces[0].FontSize)
	st.Equal("Неприкосновенность", rt.Pieces[1].Text)
	st.Equal(10.0, rt.Pieces[1].FontSize)
	st.Equal(", and some Chinese, ", rt.Pieces[2].Text)
	st.Equal(10.0, rt.Pieces[2].FontSize)
	st.Equal("表明你已明确同意你的回答接受评估", rt.Pieces[3].Text)
	st.Equal(10.0, rt.Pieces[3].FontSize)
	st.Equal(".", rt.Pieces[4].Text)
	st.Equal(10.0, rt.Pieces[4].FontSize)
	st.Equal(fonts[0], rt.Pieces[0].Font, "Should be tagged with Helvetica font.")
	st.Equal(fonts[1], rt.Pieces[1].Font, "Should be tagged with Arial font.")
	st.Equal(fonts[0], rt.Pieces[2].Font, "Should be tagged with Helvetica font.")
	st.Equal(fonts[2], rt.Pieces[3].Font, "Should be tagged with STSong font.")
	st.Equal(fonts[0], rt.Pieces[4].Font, "Should be tagged with Helvetica font.")
}

// With Chinese font first in list, Arial is not called upon for English.
func TestNewTextPiece_ChineseAndEnglish_Reversed(t *testing.T) {
	st := SuperTest{t}
	fonts := testTtfFonts("STSong", "Arial")
	rt, err := NewTextPiece("所有测abc", fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	st.MustEqual(0, len(rt.Pieces))
	st.Equal("所有测abc", rt.Text)
	st.Equal(10.0, rt.FontSize)
	st.Equal(fonts[0], rt.Font, "Should be tagged with Arial font.")
}

func TestTextPiece_MatchesAttributes(t *testing.T) {
	st := SuperTest{t}

	font := testTtfFonts("Arial")[0]
	p1 := TextPiece{
		Text:     "Lorem",
		Font:     font,
		FontSize: 10,
	}
	p2 := p1
	st.True(p1.MatchesAttributes(&p2), "Attributes should match.")

	p2.Font = testTtfFonts("Arial")[0]
	st.True(p1.MatchesAttributes(&p2), "Attributes should match.")

	p2.FontSize = 12
	st.False(p1.MatchesAttributes(&p2), "Attributes should not match.")

	p2 = p1
	p2.Color = Azure
	st.False(p1.MatchesAttributes(&p2), "Attributes should not match.")

	p2 = p1
	p2.Underline = true
	st.False(p1.MatchesAttributes(&p2), "Attributes should not match.")

	p2 = p1
	p2.LineThrough = true
	st.False(p1.MatchesAttributes(&p2), "Attributes should not match.")

	p2 = p1
	p2.CharSpacing = 1
	st.False(p1.MatchesAttributes(&p2), "Attributes should not match.")

	p2 = p1
	p2.WordSpacing = 1
	st.False(p1.MatchesAttributes(&p2), "Attributes should not match.")
}

func TestTextPiece_IsNewLine(t *testing.T) {
	st := SuperTest{t}
	font := testTtfFonts("Arial")[0]

	empty := TextPiece{
		Text:     "",
		Font:     font,
		FontSize: 10,
	}
	st.False(empty.IsNewLine(), "An empty string is not a newline.")

	newline := TextPiece{
		Text:     "\n",
		Font:     font,
		FontSize: 10,
	}
	st.True(newline.IsNewLine(), "It really is a newline.")

	nonNewline := TextPiece{
		Text:     "Lorem",
		Font:     font,
		FontSize: 10,
	}
	st.False(nonNewline.IsNewLine(), "This isn't a newline.")
}

func TestTextPiece_IsWhiteSpace(t *testing.T) {
	st := SuperTest{t}
	font := testTtfFonts("Arial")[0]

	empty := TextPiece{
		Text:     "",
		Font:     font,
		FontSize: 10,
	}
	st.False(empty.IsWhiteSpace(), "Empty string should not be considered whitespace.")

	singleWhite := TextPiece{
		Text:     " ",
		Font:     font,
		FontSize: 10,
	}
	st.True(singleWhite.IsWhiteSpace(), "A single space should be considered whitespace.")

	multiWhite := TextPiece{
		Text:     "  \t\n\v\f\r",
		Font:     font,
		FontSize: 10,
	}
	st.True(multiWhite.IsWhiteSpace(), "Multiple spaces should be considered whitespace.")

	startWhite := TextPiece{
		Text:     "  Lorem",
		Font:     font,
		FontSize: 10,
	}
	st.False(startWhite.IsWhiteSpace(), "A piece that only starts with spaces should not be considered whitespace.")

	nonWhite := TextPiece{
		Text:     "Lorem",
		Font:     font,
		FontSize: 10,
	}
	st.False(nonWhite.IsWhiteSpace(), "Piece contains no whitespace.")
}

func TestTextPiece_Len(t *testing.T) {
	st := SuperTest{t}
	rt := &TextPiece{
		Pieces: []*TextPiece{
			{Text: "Hello"},
			{Text: " "},
			{Text: "World"},
		},
	}
	st.Equal(11, rt.Len())
}

func TestTextPiece_measure(t *testing.T) {
	piece := textPieceTestText()
	piece.measure()
	expectNFdelta(t, "Ascent", 9.052734, piece.Ascent, 0.001)
	expectNFdelta(t, "Descent", -2.119141, piece.Descent, 0.001)
	expectNFdelta(t, "Height", 11.171875, piece.Height, 0.001)
	expectNFdelta(t, "UnderlinePosition", -1.059570, piece.UnderlinePosition, 0.001)
	expectNFdelta(t, "UnderlineThickness", 0.732422, piece.UnderlineThickness, 0.001)
	expectNFdelta(t, "Width", 28.344727, piece.Width, 0.001)
	expectNI(t, "Chars", 5, piece.Chars)
	expectNI(t, "Tokens", 1, piece.Tokens)
}

func TestTextPiece_Merge(t *testing.T) {
	st := SuperTest{t}
	afmFonts := testAfmFonts("Helvetica")
	ttfFonts := testTtfFonts("Arial", "STSong")
	fonts := append(afmFonts, ttfFonts...)
	text := "Here is some "
	text1 := "Russian, Неприкосновенность, "
	text2 := "and some Chinese, 表明你已明确同意你的回答接受评估."
	original, err := NewTextPiece(text, fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.Add(text1, fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	original, err = original.Add(text2, fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}

	// fmt.Println("Original pieces: ", original.Count())
	// i := 0
	// original.VisitAll(func(p *TextPiece) {
	// 	i++
	// 	fmt.Println(i, p.Text, len(p.Pieces))
	// })

	// Node
	//     Leaf: Here is some  0
	//     Node
	// 	       Leaf: Russian,
	//         Leaf: Неприкосновенность
	//         Leaf: ,
	//     Node
	//         Leaf: and some Chinese,
	//         Leaf: 表明你已明确同意你的回答接受评估
	//         Leaf: .

	piece0 := *original.Pieces[0]
	merged := original.Merge()

	// fmt.Println("Merged pieces: ", merged.Count())
	// i = 0
	// merged.VisitAll(func(p *TextPiece) {
	// 	i++
	// 	fmt.Println(i, p.Text, len(p.Pieces))
	// })

	// Node
	// Leaf: Here is some Russian,
	// Leaf: Неприкосновенность
	// Leaf: , and some Chinese,
	// Leaf: 表明你已明确同意你的回答接受评估
	// Leaf: .

	st.Equal(text+text1+text2, merged.String())
	st.MustEqual(5, len(merged.Pieces))

	st.Equal("Here is some Russian, ", merged.Pieces[0].Text)
	st.Equal(10.0, merged.Pieces[0].FontSize)
	st.Equal(fonts[0], merged.Pieces[0].Font, "Should be tagged with Helvetica font.")

	st.Equal("Неприкосновенность", merged.Pieces[1].Text)
	st.Equal(10.0, merged.Pieces[1].FontSize)
	st.Equal(fonts[1], merged.Pieces[1].Font, "Should be tagged with Arial font.")

	st.Equal(", and some Chinese, ", merged.Pieces[2].Text)
	st.Equal(10.0, merged.Pieces[2].FontSize)
	st.Equal(fonts[0], merged.Pieces[2].Font, "Should be tagged with Helvetica font.")

	st.Equal("表明你已明确同意你的回答接受评估", merged.Pieces[3].Text)
	st.Equal(10.0, merged.Pieces[3].FontSize)
	st.Equal(fonts[2], merged.Pieces[3].Font, "Should be tagged with STSong font.")

	st.Equal(".", merged.Pieces[4].Text)
	st.Equal(10.0, merged.Pieces[4].FontSize)
	st.Equal(fonts[0], merged.Pieces[4].Font, "Should be tagged with Helvetica font.")

	st.Equal(piece0.Text, original.Pieces[0].Text, "Original should be unchanged.")
}

func TestTextPiece_String(t *testing.T) {
	st := SuperTest{t}
	afmFonts := testAfmFonts("Helvetica")
	ttfFonts := testTtfFonts("Arial", "STSong")
	fonts := append(afmFonts, ttfFonts...)
	text := "Here is some Russian, Неприкосновенность, and some Chinese, 表明你已明确同意你的回答接受评估."
	rt, err := NewTextPiece(text, fonts, 10, Options{})
	if err != nil {
		t.Fatal(err)
	}
	st.Equal(text, rt.String())
}

// 14,930 ns go1.1.2
func BenchmarkNewTextPiece(b *testing.B) {
	b.StopTimer()
	afmFonts := testAfmFonts("Helvetica")
	ttfFonts := testTtfFonts("Arial", "STSong")
	fonts := append(afmFonts, ttfFonts...)
	text := "Here is some Russian, Неприкосновенность, and some Chinese, 表明你已明确同意你的回答接受评估."
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := NewTextPiece(text, fonts, 10, Options{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// 75.5 ns
// 76.9 ns go1.1.1
func BenchmarkTextPiece_IsWhiteSpace(b *testing.B) {
	b.StopTimer()
	font := testTtfFonts("Arial")[0]
	piece := &TextPiece{
		Text:     "  \t\n\v\f\r",
		Font:     font,
		FontSize: 10,
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		piece.IsWhiteSpace()
	}
}

// 345 ns
// 301 ns go1.1.1
func BenchmarkTextPiece_measure(b *testing.B) {
	b.StopTimer()
	font := testTtfFonts("Arial")[0]
	piece := &TextPiece{
		Text:     "Lorem",
		Font:     font,
		FontSize: 10,
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		piece.measure()
	}
}
