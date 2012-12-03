// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "unicode"

type TextPiece struct {
	Text        string
	Font        *Font
	FontSize    float64
	Color       Color
	Underline   bool
	LineThrough bool
	Ascent      float64
	Descent     float64
	Height      float64
	Width       float64
	Chars       int
	Tokens      int
}

func (piece *TextPiece) measure(charSpacing, wordSpacing float64) *TextPiece {
	piece.Chars, piece.Width = 0, 0
	fsize := piece.FontSize / float64(piece.Font.metrics.UnitsPerEm())
	metrics := piece.Font.metrics
	piece.Ascent = float64(metrics.Ascent()) * piece.FontSize / float64(metrics.UnitsPerEm())
	piece.Descent = float64(metrics.Descent()) * piece.FontSize / float64(metrics.UnitsPerEm())
	piece.Height = float64(piece.Font.Height()) * piece.FontSize / float64(metrics.UnitsPerEm())
	for _, rune := range piece.Text {
		piece.Chars += 1
		runeWidth, _ := metrics.AdvanceWidth(rune)
		piece.Width += (fsize * float64(runeWidth)) + charSpacing
		if unicode.IsSpace(rune) {
			piece.Width += wordSpacing
		}
	}
	piece.Tokens = 1
	return piece
}
