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

// NoopWriter is a zero-value Writer implementation intended to be embedded in
// test doubles. Embed it and override only the operations a test needs to
// observe or customize.
//
// Callback-based operations still invoke their callback so that wrapping
// drawing operations in a clip, path, rotation, accessibility tag, or text
// direction does not suppress the operations under test.
type NoopWriter struct{}

var (
	_ Writer = NoopWriter{}
	_ Writer = (*NoopWriter)(nil)
)

func (NoopWriter) AddFont(string, options.Options) ([]*font.Font, error) { return nil, nil }
func (NoopWriter) AppendArchPath(float64, float64, float64, float64, float64, float64, bool) error {
	return nil
}
func (NoopWriter) AppendClosedShapePath(pdf.ClosedShape) error { return nil }
func (NoopWriter) AppendPiePath(float64, float64, float64, float64, float64, bool) error {
	return nil
}
func (NoopWriter) Arch(float64, float64, float64, float64, float64, float64, bool, bool, bool) error {
	return nil
}
func (NoopWriter) Arc(float64, float64, float64, float64, float64, bool) error { return nil }
func (NoopWriter) Circle(float64, float64, float64, bool, bool, bool) error    { return nil }
func (NoopWriter) ClearFillGradient()                                          {}
func (NoopWriter) ClearLineGradient()                                          {}
func (NoopWriter) Clip(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (NoopWriter) ClipClosedShape(_ pdf.ClosedShape, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (NoopWriter) ClipRichText(_ *rich_text.RichText, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (NoopWriter) ClipText(_ string, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (NoopWriter) ClosedShapeBounds(shape pdf.ClosedShape) (pdf.Bounds, error) {
	return shape.Bounds()
}
func (NoopWriter) CompressEmbeddedFonts(bool) *pdf.DocWriter { return nil }
func (NoopWriter) CompressPages(bool) *pdf.DocWriter         { return nil }
func (NoopWriter) CompressToUnicode(bool) *pdf.DocWriter     { return nil }
func (NoopWriter) CurvePoints([]pdf.Location) error          { return nil }
func (NoopWriter) DrawClosedShape(pdf.ClosedShape, bool, bool) error {
	return nil
}
func (NoopWriter) DrawRichTextOnCircle(*rich_text.RichText, float64, float64, float64, float64, pdf.CurvedTextOptions) error {
	return nil
}
func (NoopWriter) DrawTextOnCircle(string, float64, float64, float64, float64, pdf.CurvedTextOptions) error {
	return nil
}
func (NoopWriter) Ellipse(float64, float64, float64, float64, bool, bool, bool) error {
	return nil
}
func (NoopWriter) EnableTaggedPDF(bool) {}
func (NoopWriter) Fill() error          { return nil }
func (NoopWriter) FillAndStroke() error { return nil }
func (NoopWriter) FontColor() colors.Color {
	return colors.Black
}
func (NoopWriter) Fonts() []*font.Font { return nil }
func (NoopWriter) FontSize() float64   { return 12 }
func (NoopWriter) ImageDimensions([]byte) (int, int, error) {
	return 0, 0, nil
}
func (NoopWriter) ImageDimensionsFromFile(string) (int, int, error) {
	return 0, 0, nil
}
func (NoopWriter) Line(float64, float64, float64, float64) {}
func (NoopWriter) LineSpacing() float64                    { return 1 }
func (NoopWriter) LineTo(float64, float64)                 {}
func (NoopWriter) Loc() (float64, float64)                 { return 0, 0 }
func (NoopWriter) MoveTo(float64, float64)                 {}
func (NoopWriter) NewPage()                                {}
func (NoopWriter) PaintImageFile(string, float64, float64, float64, float64, float64) error {
	return nil
}
func (NoopWriter) PaintLinearGradient(*pdf.LinearGradient) error { return nil }
func (NoopWriter) PaintRadialGradient(*pdf.RadialGradient) error { return nil }
func (NoopWriter) PaintSweepBand(*pdf.SweepBand) error           { return nil }
func (NoopWriter) Path(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (NoopWriter) Pie(float64, float64, float64, float64, float64, bool, bool, bool) error {
	return nil
}
func (NoopWriter) Polygon(float64, float64, float64, int, bool, bool, bool, float64) error {
	return nil
}
func (NoopWriter) Print(string) error { return nil }
func (NoopWriter) PrintImage([]byte, float64, float64, *float64, *float64) (float64, float64, error) {
	return 0, 0, nil
}
func (NoopWriter) PrintImageFile(string, float64, float64, *float64, *float64) (float64, float64, error) {
	return 0, 0, nil
}
func (NoopWriter) PrintParagraph([]*rich_text.RichText, options.Options) {}
func (NoopWriter) PrintRichText(*rich_text.RichText)                     {}
func (NoopWriter) PrintSVG([]byte, float64, float64, *float64, *float64) (float64, float64, error) {
	return 0, 0, nil
}
func (NoopWriter) PrintSVGFile(string, float64, float64, *float64, *float64) (float64, float64, error) {
	return 0, 0, nil
}
func (NoopWriter) Rectangle(float64, float64, float64, float64, bool, bool) {}
func (NoopWriter) Rectangle2(float64, float64, float64, float64, bool, bool, []float64, bool, bool) {
}
func (NoopWriter) Rotate(_ float64, _ float64, _ float64, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (NoopWriter) SetFillColor(any) colors.Color { return colors.Black }
func (NoopWriter) SetFillLinearGradient(*pdf.LinearGradient) error {
	return nil
}
func (NoopWriter) SetFillRadialGradient(*pdf.RadialGradient) error {
	return nil
}
func (NoopWriter) SetFont(string, float64, options.Options) ([]*font.Font, error) {
	return nil, nil
}
func (NoopWriter) SetLanguage(string)            {}
func (NoopWriter) SetLineCapStyle(string) string { return "" }
func (NoopWriter) SetLineColor(colors.Color) colors.Color {
	return colors.Black
}
func (NoopWriter) SetLineDashPattern(string) string { return "" }
func (NoopWriter) SetLineLinearGradient(*pdf.LinearGradient) error {
	return nil
}
func (NoopWriter) SetLineRadialGradient(*pdf.RadialGradient) error {
	return nil
}
func (NoopWriter) SetLineSpacing(float64) float64 { return 1 }
func (NoopWriter) SetLineWidth(float64)           {}
func (NoopWriter) SetStrikeout(bool) bool         { return false }
func (NoopWriter) SetSVGBlendMode(pdf.SVGBlendMode) pdf.SVGBlendMode {
	return pdf.SVGBlendModeRespect
}
func (NoopWriter) SetSVGGradientStopOpacityMode(pdf.SVGGradientStopOpacityMode) pdf.SVGGradientStopOpacityMode {
	return pdf.SVGGradientStopOpacityModeSoftMask
}
func (NoopWriter) SetUnderline(bool) bool { return false }
func (NoopWriter) Star(float64, float64, float64, float64, int, bool, bool, bool, float64) error {
	return nil
}
func (NoopWriter) Stroke() error   { return nil }
func (NoopWriter) Strikeout() bool { return false }
func (NoopWriter) SVGDimensions([]byte) (int, int, error) {
	return 0, 0, nil
}
func (NoopWriter) SVGDimensionsFromFile(string) (int, int, error) {
	return 0, 0, nil
}
func (NoopWriter) TaggedPDFEnabled() bool { return false }
func (NoopWriter) Underline() bool        { return false }
func (NoopWriter) WithAccessibilityArtifact(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (NoopWriter) WithAccessibilityTag(_ string, _ pdf.AccessibilityOptions, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (NoopWriter) WithTextDirection(_ pdf.TextDirection, fn func() error) error {
	if fn == nil {
		return nil
	}
	return fn()
}
