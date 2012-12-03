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

func TestTextPiece_measure(t *testing.T) {
	font := testTtfFonts("Arial")[0]
	piece := &TextPiece{
		Text:     "Lorem",
		Font:     font,
		FontSize: 10,
	}
	piece.measure(0, 0)
	expectNF(t, "Width", 58.05, piece.Width)
	expectNI(t, "Chars", 5, piece.Chars)
	expectNI(t, "Tokens", 1, piece.Tokens)
}

func TestTextPiece_Ascent(t *testing.T) {
	piece := textPieceTestText()
	expectF(t, 18.54, piece.Ascent())
}

func TestTextPieceDescent(t *testing.T) {
	piece := textPieceTestText()
	expectF(t, -4.34, piece.Descent())
}

func TestTextPiece_Height(t *testing.T) {
	piece := textPieceTestText()
	expectFdelta(t, 22.88, piece.Height(), 0.001)
}

// 313 ns
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
