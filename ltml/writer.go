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
	Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error
	Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error
	Circle(x, y, r float64, border, fill, reverse bool) error
	Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error
	FontColor() colors.Color
	Fonts() []*font.Font
	FontSize() float64
	ImageDimensionsFromFile(filename string) (width, height int, err error)
	LineSpacing() float64
	SetLineCapStyle(style string) (prev string)
	Line(x, y, angle, length float64)
	LineTo(x, y float64)
	Loc() (x, y float64)
	MoveTo(x, y float64)
	NewPage()
	Print(text string) (err error)
	PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error)
	PrintParagraph(para []*rich_text.RichText, options options.Options)
	PrintRichText(text *rich_text.RichText)
	Path(fn func()) error
	Pie(x, y, r, startAngle, endAngle float64, border, fill, reverse bool) error
	Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error
	Rectangle(x, y, width, height float64, border bool, fill bool)
	Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool)
	Rotate(angle, x, y float64, fn func()) error
	AddFont(family string, options options.Options) ([]*font.Font, error)
	SetFont(name string, size float64, options options.Options) ([]*font.Font, error)
	SetFillColor(value any) (prev colors.Color)
	SetLineColor(value colors.Color) (prev colors.Color)
	SetLineDashPattern(pattern string) (prev string)
	SetLineSpacing(lineSpacing float64) (prev float64)
	SetLineWidth(width float64)
	SetStrikeout(strikeout bool) (prev bool)
	SetUnderline(underline bool) (prev bool)
	Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error
	Stroke() error
	Strikeout() bool
	Underline() bool
}
