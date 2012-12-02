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
	Width       float64
	Chars       int
	Tokens      int
}

func (piece *TextPiece) Ascent() float64 {
	return 0.001 * float64(piece.Font.Ascent()) * piece.FontSize
}

func (piece *TextPiece) Descent() float64 {
	return 0.001 * float64(piece.Font.Descent()) * piece.FontSize
}

func (piece *TextPiece) Height() float64 {
	return 0.001 * float64(piece.Font.Height()) * piece.FontSize
}

func (piece *TextPiece) measure(charSpacing, wordSpacing float64) *TextPiece {
	piece.Chars, piece.Width = 0, 0
	fsize := piece.FontSize * 0.001
	metrics := piece.Font.metrics
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
