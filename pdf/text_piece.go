// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "unicode"

type TextPiece struct {
	Text               string
	Font               *Font
	FontSize           float64
	Color              Color
	Underline          bool
	LineThrough        bool
	Ascent             float64
	Descent            float64
	Height             float64
	UnderlinePosition  float64
	UnderlineThickness float64
	Width              float64
	Chars              int
	Tokens             int
}

func (piece *TextPiece) IsNewLine() bool {
	return piece.Text == "\n"
}

func (piece *TextPiece) IsWhiteSpace() bool {
	for _, rune := range piece.Text {
		if !unicode.IsSpace(rune) {
			return false
		}
	}
	return len(piece.Text) > 0
}

func (piece *TextPiece) measure(charSpacing, wordSpacing float64) *TextPiece {
	piece.Chars, piece.Width = 0, 0
	metrics := piece.Font.metrics
	fsize := piece.FontSize / float64(metrics.UnitsPerEm())
	piece.Ascent = float64(metrics.Ascent()) * fsize
	piece.Descent = float64(metrics.Descent()) * fsize
	piece.Height = float64(piece.Font.Height()) * fsize
	piece.UnderlinePosition = float64(metrics.UnderlinePosition()) * fsize
	piece.UnderlineThickness = float64(metrics.UnderlineThickness()) * fsize
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
