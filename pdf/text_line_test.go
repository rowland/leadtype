// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "testing"

func textLineTestText() RichText {
	fonts := testTtfFonts("Arial", "STSong")
	rt, err := NewRichText("abc所有测def", fonts, 10, Options{})
	if err != nil {
		panic(err)
	}
	return rt
}

func TestTextLine_Ascent(t *testing.T) {
	var line TextLine
	expectF(t, 0, line.Ascent())
	line = TextLine{RichText: textLineTestText()}
	expectFdelta(t, 9.052734, line.Ascent(), 0.001)
}

func TestTextLine_Chars(t *testing.T) {
	var line TextLine
	expectI(t, 0, line.Chars())
	line = TextLine{RichText: textLineTestText()}
	expectI(t, 9, line.Chars())
}

func TestTextLine_Descent(t *testing.T) {
	var line TextLine
	expectF(t, 0, line.Descent())
	line = TextLine{RichText: textLineTestText()}
	expectFdelta(t, -2.119141, line.Descent(), 0.001)
}

func TestTextLine_Height(t *testing.T) {
	var line TextLine
	expectF(t, 0, line.Height())
	line = TextLine{RichText: textLineTestText()}
	expectFdelta(t, 11.171875, line.Height(), 0.001)
}

func TestTextLine_Tokens(t *testing.T) {
	var line TextLine
	expectI(t, 0, line.Tokens())
	line = TextLine{RichText: textLineTestText()}
	expectI(t, 3, line.Tokens())
}

func TestTextLine_Width(t *testing.T) {
	var line TextLine
	expectF(t, 0, line.Width())
	line = TextLine{RichText: textLineTestText()}
	expectFdelta(t, 60.024414, line.Width(), 0.001)
}

func textLineBenchmarkText() RichText {
	const lorem = "Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	fonts := testTtfFonts("Arial")
	rt, err := NewRichText(lorem, fonts, 10, Options{})
	if err != nil {
		panic(err)
	}
	return rt
}

// 201 ns
func BenchmarkTextLine_Ascent(b *testing.B) {
	b.StopTimer()
	text := textLineBenchmarkText()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		line := TextLine{RichText: text}
		line.Ascent()
	}
}

// 221 ns
func BenchmarkTextLine_Chars(b *testing.B) {
	b.StopTimer()
	text := textLineBenchmarkText()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		line := TextLine{RichText: text}
		line.Chars()
	}
}

// 200 ns
func BenchmarkTextLine_Descent(b *testing.B) {
	b.StopTimer()
	text := textLineBenchmarkText()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		line := TextLine{RichText: text}
		line.Descent()
	}
}

// 198 ns
func BenchmarkTextLine_Height(b *testing.B) {
	b.StopTimer()
	text := textLineBenchmarkText()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		line := TextLine{RichText: text}
		line.Height()
	}
}

// 221 ns
func BenchmarkTextLine_Tokens(b *testing.B) {
	b.StopTimer()
	text := textLineBenchmarkText()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		line := TextLine{RichText: text}
		line.Tokens()
	}
}

// 247 ns
func BenchmarkTextLine_Width(b *testing.B) {
	b.StopTimer()
	text := textLineBenchmarkText()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		line := TextLine{RichText: text}
		line.Width()
	}
}
