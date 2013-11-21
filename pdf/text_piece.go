// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
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
	ascent             float64
	descent            float64
	height             float64
	UnderlinePosition  float64
	UnderlineThickness float64
	width              float64
	chars              int
	CharSpacing        float64
	WordSpacing        float64
	NoBreak            bool
	pieces             []*TextPiece
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
	return &TextPiece{pieces: pieces}, nil
}

func (piece *TextPiece) Add(s string, fonts []*Font, fontSize float64, options Options) (*TextPiece, error) {
	p, err := NewTextPiece(s, fonts, fontSize, options)
	if err != nil {
		return piece, err
	}
	if len(piece.pieces) > 0 {
		piece.pieces = append(piece.pieces, p)
	} else {
		piece = &TextPiece{pieces: []*TextPiece{piece, p}}
	}
	return piece, nil
}

func (piece *TextPiece) Count() int {
	result := 1
	for _, p := range piece.pieces {
		result += p.Count()
	}
	return result
}

func (piece *TextPiece) IsLeaf() bool {
	return len(piece.pieces) == 0
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

func (piece *TextPiece) Len() int {
	result := 0
	for i := 0; i < len(piece.pieces); i++ {
		result += len(piece.pieces[i].Text)
	}
	return result
}

func (piece *TextPiece) MatchesAttributes(other *TextPiece) bool {
	return (piece.Font != nil && other.Font != nil) && (piece.Font == other.Font || piece.Font.Matches(other.Font)) &&
		piece.FontSize == other.FontSize &&
		piece.Color == other.Color &&
		piece.Underline == other.Underline &&
		piece.LineThrough == other.LineThrough &&
		piece.CharSpacing == other.CharSpacing &&
		piece.WordSpacing == other.WordSpacing
}

func (piece *TextPiece) measure() *TextPiece {
	piece.chars, piece.width = 0, 0
	metrics := piece.Font.metrics
	fsize := piece.FontSize / float64(metrics.UnitsPerEm())
	piece.ascent = float64(metrics.Ascent()) * fsize
	piece.descent = float64(metrics.Descent()) * fsize
	piece.height = float64(piece.Font.Height()) * fsize
	piece.UnderlinePosition = float64(metrics.UnderlinePosition()) * fsize
	piece.UnderlineThickness = float64(metrics.UnderlineThickness()) * fsize
	for _, rune := range piece.Text {
		piece.chars += 1
		runeWidth, _ := metrics.AdvanceWidth(rune)
		piece.width += (fsize * float64(runeWidth)) + piece.CharSpacing
		if unicode.IsSpace(rune) {
			piece.width += piece.WordSpacing
		}
	}
	return piece
}

func (piece *TextPiece) Merge() *TextPiece {
	if len(piece.pieces) == 0 {
		return piece
	}

	flattened := make([]*TextPiece, 0, len(piece.pieces))
	for _, p := range piece.pieces {
		pm := p.Merge()
		if pm.IsLeaf() || pm.NoBreak {
			flattened = append(flattened, pm)
		} else {
			flattened = append(flattened, pm.pieces...)
		}
	}

	mergedText := *piece
	mergedText.pieces = make([]*TextPiece, 0, len(piece.pieces))
	var last *TextPiece
	for _, p := range flattened {
		if last != nil && p.MatchesAttributes(last) {
			last.Text += p.Text
		} else {
			newPiece := *p
			mergedText.pieces = append(mergedText.pieces, &newPiece)
			last = &newPiece
		}
	}
	if len(mergedText.pieces) == 1 {
		return mergedText.pieces[0]
	}
	return &mergedText
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
					var newPieces []*TextPiece
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
			var newPieces []*TextPiece
			newPieces, err = newPiece.splitByFont(fonts[1:])
			pieces = append(pieces, newPieces...)
		}
	}
	return
}

func (piece *TextPiece) String() string {
	var buf bytes.Buffer
	piece.VisitAll(func(p *TextPiece) {
		buf.WriteString(p.Text)
	})
	return buf.String()
}

func (piece *TextPiece) VisitAll(fn func(*TextPiece)) {
	fn(piece)
	for _, p := range piece.pieces {
		p.VisitAll(fn)
	}
}
