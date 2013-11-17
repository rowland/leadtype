// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
	"unicode"
)

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
	Pieces             []*TextPiece
}

func NewTextPiece(s string, fonts []*Font, fontSize float64, options Options) (*TextPiece, error) {
	piece := &TextPiece{
		Text:        s,
		FontSize:    fontSize,
		Color:       options.ColorDefault("color", Black),
		Underline:   options.BoolDefault("underline", false),
		LineThrough: options.BoolDefault("line_through", false),
		CharSpacing: options.FloatDefault("char_spacing", 0),
		WordSpacing: options.FloatDefault("word_spacing", 0),
	}
	pieces, err := piece.splitByFont(fonts)
	if err != nil {
		return nil, err
	}
	if len(pieces) == 1 {
		return pieces[0], nil
	}
	return &TextPiece{Pieces: pieces}, nil
}

func (richText *TextPiece) Add(s string, fonts []*Font, fontSize float64, options Options) (*TextPiece, error) {
	piece, err := NewTextPiece(s, fonts, fontSize, options)
	if err != nil {
		return richText, err
	}
	if len(richText.Pieces) > 0 {
		richText.Pieces = append(richText.Pieces, piece)
	} else {
		richText = &TextPiece{Pieces: []*TextPiece{richText, piece}}
	}
	return richText, nil
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

func (piece *TextPiece) MatchesAttributes(other *TextPiece) bool {
	return (piece.Font == other.Font || piece.Font.Matches(other.Font)) &&
		piece.FontSize == other.FontSize &&
		piece.Color == other.Color &&
		piece.Underline == other.Underline &&
		piece.LineThrough == other.LineThrough &&
		piece.CharSpacing == other.CharSpacing &&
		piece.WordSpacing == other.WordSpacing
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

func (piece *TextPiece) splitByFont(fonts []*Font) (pieces []*TextPiece, err error) {
	if len(fonts) == 0 {
		return nil, fmt.Errorf("No font found for %s.", piece.Text)
	}
	font := fonts[0]
	start := 0
	inFont := false
	for index, rune := range piece.Text {
		runeInFont := font.HasRune(rune)
		if runeInFont != inFont {
			if index > start {
				newPiece := *piece
				newPiece.Text = piece.Text[start:index]
				if inFont {
					newPiece.Font = font
					newPiece.measure()
					pieces = append(pieces, &newPiece)
				} else {
					var newPieces RichText
					newPieces, err = newPiece.splitByFont(fonts[1:])
					pieces = append(pieces, newPieces...)
				}
			}
			inFont = runeInFont
			start = index
		}
	}
	if len(piece.Text) > start {
		newPiece := *piece
		newPiece.Text = piece.Text[start:]
		if inFont {
			newPiece.Font = font
			newPiece.measure()
			pieces = append(pieces, &newPiece)
		} else {
			var newPieces RichText
			newPieces, err = newPiece.splitByFont(fonts[1:])
			pieces = append(pieces, newPieces...)
		}
	}
	return
}
