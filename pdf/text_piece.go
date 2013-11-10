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
	CharSpacing        float64
	WordSpacing        float64
}

func (piece *TextPiece) MatchesAttributes(other *TextPiece) bool {
	return (piece.Font == other.Font || piece.Font.Matches(other.Font)) &&
		piece.FontSize == other.FontSize &&
		piece.Color == other.Color &&
		piece.Underline == other.Underline &&
		piece.LineThrough == other.LineThrough &&
		piece.CharSpacing == other.CharSpacing &&
		piece.WordSpacing == other.WordSpacing
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

func (piece *TextPiece) measure() *TextPiece {
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
		piece.Width += (fsize * float64(runeWidth)) + piece.CharSpacing
		if unicode.IsSpace(rune) {
			piece.Width += piece.WordSpacing
		}
	}
	piece.Tokens = 1
	return piece
}
