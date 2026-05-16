// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/rich_text"
)

type Writer interface {
	Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error
	Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error
	Clip(fn func()) error
	ClipClosedShape(shape pdf.ClosedShape, fn func()) error
	ClipRichText(text *rich_text.RichText, fn func()) error
	ClipText(text string, fn func()) error
	ClosedShapeBounds(shape pdf.ClosedShape) (pdf.Bounds, error)
	CompressEmbeddedFonts(bool) *pdf.DocWriter
	CompressPages(bool) *pdf.DocWriter
	CompressToUnicode(bool) *pdf.DocWriter
	DrawRichTextOnCircle(text *rich_text.RichText, x, y, r, startAngle float64, opts pdf.CurvedTextOptions) error
	DrawTextOnCircle(text string, x, y, r, startAngle float64, opts pdf.CurvedTextOptions) error
	DrawClosedShape(shape pdf.ClosedShape, border, fill bool) error
	EnableTaggedPDF(bool)
	Circle(x, y, r float64, border, fill, reverse bool) error
	Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error
	FontColor() colors.Color
	Fonts() []*font.Font
	FontSize() float64
	ImageDimensions(data []byte) (width, height int, err error)
	SVGDimensions(data []byte) (width, height int, err error)
	SVGDimensionsFromFile(filename string) (width, height int, err error)
	ImageDimensionsFromFile(filename string) (width, height int, err error)
	LineSpacing() float64
	SetLineCapStyle(style string) (prev string)
	Line(x, y, angle, length float64)
	LineTo(x, y float64)
	Loc() (x, y float64)
	MoveTo(x, y float64)
	NewPage()
	Print(text string) (err error)
	PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error)
	PrintSVG(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error)
	PrintSVGFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error)
	PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error)
	// PaintImageFile is the clip-oriented image-as-fill operation used by LTML
	// background/text fill code. Unlike generic image placement, callers are
	// expected to establish any clipping region before invoking it.
	PaintImageFile(filename string, x, y, width, height, opacity float64) error
	PaintLinearGradient(lg *pdf.LinearGradient) error
	PaintRadialGradient(rg *pdf.RadialGradient) error
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
	SetFillLinearGradient(lg *pdf.LinearGradient) error
	SetFillRadialGradient(rg *pdf.RadialGradient) error
	ClearFillGradient()
	SetLineLinearGradient(lg *pdf.LinearGradient) error
	SetLineRadialGradient(rg *pdf.RadialGradient) error
	ClearLineGradient()
	SetLineColor(value colors.Color) (prev colors.Color)
	SetLineDashPattern(pattern string) (prev string)
	SetLineSpacing(lineSpacing float64) (prev float64)
	SetLineWidth(width float64)
	SetSVGBlendMode(pdf.SVGBlendMode) pdf.SVGBlendMode
	SetSVGGradientStopOpacityMode(pdf.SVGGradientStopOpacityMode) pdf.SVGGradientStopOpacityMode
	SetStrikeout(strikeout bool) (prev bool)
	SetUnderline(underline bool) (prev bool)
	Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error
	Stroke() error
	Strikeout() bool
	TaggedPDFEnabled() bool
	Underline() bool
	WithAccessibilityArtifact(fn func()) error
	WithAccessibilityTag(tag string, opts pdf.AccessibilityOptions, fn func()) error
}

type PageOptionWriter interface {
	NewPageWithOptions(options.Options)
}
