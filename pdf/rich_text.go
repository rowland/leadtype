// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
)

type RichText []*TextPiece

func NewRichText(s string, fonts []*Font, fontSize float64, options Options) (RichText, error) {
	var richText RichText
	return richText.Add(s, fonts, fontSize, options)
}

func (richText RichText) Add(s string, fonts []*Font, fontSize float64, options Options) (RichText, error) {
	piece := &TextPiece{
		Text:        s,
		FontSize:    fontSize,
		Color:       options.ColorDefault("color", Black),
		Underline:   options.BoolDefault("underline", false),
		LineThrough: options.BoolDefault("line_through", false),
		CharSpacing: options.FloatDefault("char_spacing", 0),
		WordSpacing: options.FloatDefault("word_spacing", 0),
	}
	addedText, err := richTextFromTextPiece(piece, fonts)
	if err != nil {
		return richText, err
	}
	return append(richText, addedText...), nil
}

func (richText RichText) Len() int {
	result := 0
	for i := 0; i < len(richText); i++ {
		result += len(richText[i].Text)
	}
	return result
}

func (richText RichText) Merge() RichText {
	mergedText := make(RichText, 0, len(richText))
	var last *TextPiece
	for _, piece := range richText {
		if last != nil && piece.MatchesAttributes(last) {
			last.Text += piece.Text
		} else {
			newPiece := *piece
			mergedText = append(mergedText, &newPiece)
			last = &newPiece
		}
	}
	return mergedText
}

func richTextFromTextPiece(piece *TextPiece, fonts []*Font) (richText RichText, err error) {
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
					richText = append(richText, &newPiece)
				} else {
					var newPieces RichText
					newPieces, err = richTextFromTextPiece(&newPiece, fonts[1:])
					richText = append(richText, newPieces...)
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
			richText = append(richText, &newPiece)
		} else {
			var newPieces RichText
			newPieces, err = richTextFromTextPiece(&newPiece, fonts[1:])
			richText = append(richText, newPieces...)
		}
	}
	return
}
