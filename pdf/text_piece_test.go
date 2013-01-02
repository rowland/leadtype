// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "testing"

func textPieceTestText() *TextPiece {
	font := testTtfFonts("Arial")[0]
	return &TextPiece{
		Text:     "Lorem",
		Font:     font,
		FontSize: 10,
	}
}

func TestTextPiece_MatchesAttributes(t *testing.T) {
	font := testTtfFonts("Arial")[0]
	p1 := TextPiece{
		Text:     "Lorem",
		Font:     font,
		FontSize: 10,
	}
	p2 := p1
	check(t, p1.MatchesAttributes(&p2), "Attributes should match.")

	p2.Font = testTtfFonts("Arial")[0]
	check(t, p1.MatchesAttributes(&p2), "Attributes should match.")

	p2.FontSize = 12
	check(t, !p1.MatchesAttributes(&p2), "Attributes should not match.")

	p2 = p1
	p2.Color = Azure
	check(t, !p1.MatchesAttributes(&p2), "Attributes should not match.")

	p2 = p1
	p2.Underline = true
	check(t, !p1.MatchesAttributes(&p2), "Attributes should not match.")

	p2 = p1
	p2.LineThrough = true
	check(t, !p1.MatchesAttributes(&p2), "Attributes should not match.")
}

func TestTextPiece_IsNewLine(t *testing.T) {
	font := testTtfFonts("Arial")[0]

	empty := TextPiece{
		Text:     "",
		Font:     font,
		FontSize: 10,
	}
	check(t, !empty.IsNewLine(), "An empty string is not a newline.")

	newline := TextPiece{
		Text:     "\n",
		Font:     font,
		FontSize: 10,
	}
	check(t, newline.IsNewLine(), "It really is a newline.")

	nonNewline := TextPiece{
		Text:     "Lorem",
		Font:     font,
		FontSize: 10,
	}
	check(t, !nonNewline.IsNewLine(), "This isn't a newline.")
}

func TestTextPiece_IsWhiteSpace(t *testing.T) {
	font := testTtfFonts("Arial")[0]

	empty := TextPiece{
		Text:     "",
		Font:     font,
		FontSize: 10,
	}
	check(t, !empty.IsWhiteSpace(), "Empty string should not be considered whitespace.")

	singleWhite := TextPiece{
		Text:     " ",
		Font:     font,
		FontSize: 10,
	}
	check(t, singleWhite.IsWhiteSpace(), "A single space should be considered whitespace.")

	multiWhite := TextPiece{
		Text:     "  \t\n\v\f\r",
		Font:     font,
		FontSize: 10,
	}
	check(t, multiWhite.IsWhiteSpace(), "Multiple spaces should be considered whitespace.")

	startWhite := TextPiece{
		Text:     "  Lorem",
		Font:     font,
		FontSize: 10,
	}
	check(t, !startWhite.IsWhiteSpace(), "A piece that only starts with spaces should not be considered whitespace.")

	nonWhite := TextPiece{
		Text:     "Lorem",
		Font:     font,
		FontSize: 10,
	}
	check(t, !nonWhite.IsWhiteSpace(), "Piece contains no whitespace.")
}

func TestTextPiece_measure(t *testing.T) {
	piece := textPieceTestText()
	piece.measure(0, 0)
	expectNFdelta(t, "Ascent", 9.052734, piece.Ascent, 0.001)
	expectNFdelta(t, "Descent", -2.119141, piece.Descent, 0.001)
	expectNFdelta(t, "Height", 11.171875, piece.Height, 0.001)
	expectNFdelta(t, "UnderlinePosition", -1.059570, piece.UnderlinePosition, 0.001)
	expectNFdelta(t, "UnderlineThickness", 0.732422, piece.UnderlineThickness, 0.001)
	expectNFdelta(t, "Width", 28.344727, piece.Width, 0.001)
	expectNI(t, "Chars", 5, piece.Chars)
	expectNI(t, "Tokens", 1, piece.Tokens)
}

// 75.5 ns
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
		piece.measure(0, 0)
	}
}
