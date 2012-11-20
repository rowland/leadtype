// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "fmt"

type RichText []*TextPiece

func NewRichText(s string, fonts []*Font) (richText RichText, err error) {
	piece := &TextPiece{Text: s}
	richText, err = RichTextFromTextPiece(piece, fonts)
	return
}

func RichTextFromTextPiece(piece *TextPiece, fonts []*Font) (richText RichText, err error) {
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
					newPiece.Font = font;
					richText = append(richText, &newPiece)
				} else {
					var newPieces RichText
					newPieces, err = RichTextFromTextPiece(&newPiece, fonts[1:])
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
			newPiece.Font = font;
			richText = append(richText, &newPiece)
		} else {
			var newPieces RichText
			newPieces, err = RichTextFromTextPiece(&newPiece, fonts[1:])
			richText = append(richText, newPieces...)
		}
	}
	return
}
