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

func (piece *TextPiece) measure(charSpacing, wordSpacing float64) *TextPiece {
	piece.Width = 0
	fsize := piece.FontSize * 0.001
	metrics := piece.Font.metrics
	for _, rune := range piece.Text {
		runeWidth, _ := metrics.AdvanceWidth(rune)
		piece.Width += (fsize * float64(runeWidth)) + charSpacing
		if unicode.IsSpace(rune) {
			piece.Width += wordSpacing
		}
	}
	piece.Chars = len(piece.Text)
	piece.Tokens = 1
	return piece
}
