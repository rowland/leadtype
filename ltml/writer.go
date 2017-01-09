// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

type Writer interface {
	FontColor() colors.Color
	Fonts() []*font.Font
	FontSize() float64
	LineSpacing() float64
	LineTo(x, y float64)
	Loc() (x, y float64)
	MoveTo(x, y float64)
	NewPage()
	Print(text string) (err error)
	PrintParagraph(para []*rich_text.RichText, options options.Options)
	PrintRichText(text *rich_text.RichText)
	Rectangle(x, y, width, height float64, border bool, fill bool)
	Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool)
	SetFont(name string, size float64, options options.Options) ([]*font.Font, error)
	SetFillColor(value interface{}) (prev colors.Color)
	SetLineColor(value colors.Color) (prev colors.Color)
	SetLineDashPattern(pattern string)
	SetLineSpacing(lineSpacing float64) (prev float64)
	SetLineWidth(width float64)
	SetStrikeout(strikeout bool) (prev bool)
	SetUnderline(underline bool) (prev bool)
	Strikeout() bool
	Underline() bool
}
