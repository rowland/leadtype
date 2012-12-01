// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "testing"

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
