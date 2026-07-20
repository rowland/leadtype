package ltml

import (
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/profile"
	"github.com/rowland/leadtype/rich_text"
)

type layoutProbeWriter struct {
	base Writer
}

func newLayoutProbeWriter(base Writer) Writer {
	if factory, ok := base.(interface{ LayoutProbeWriter() any }); ok {
		if probe, ok := factory.LayoutProbeWriter().(Writer); ok {
			base = probe
		}
	}
	return &layoutProbeWriter{base: base}
}

func (w *layoutProbeWriter) Profiler() *profile.Profiler {
	return profilerForWriter(w.base)
}

func (w *layoutProbeWriter) SetProfiler(profiler *profile.Profiler) {
	setWriterProfiler(w.base, profiler)
}

func (w *layoutProbeWriter) AddFont(family string, opts options.Options) ([]*font.Font, error) {
	return w.base.AddFont(family, opts)
}
func (w *layoutProbeWriter) Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error {
	return nil
}
func (w *layoutProbeWriter) Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error {
	return nil
}
func (w *layoutProbeWriter) Clip(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *layoutProbeWriter) ClipClosedShape(shape pdf.ClosedShape, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *layoutProbeWriter) ClipRichText(text *rich_text.RichText, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *layoutProbeWriter) ClipText(text string, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *layoutProbeWriter) CompressEmbeddedFonts(value bool) *pdf.DocWriter {
	return w.base.CompressEmbeddedFonts(value)
}
func (w *layoutProbeWriter) CompressPages(value bool) *pdf.DocWriter {
	return w.base.CompressPages(value)
}
func (w *layoutProbeWriter) CompressToUnicode(value bool) *pdf.DocWriter {
	return w.base.CompressToUnicode(value)
}
func (w *layoutProbeWriter) DrawRichTextOnCircle(text *rich_text.RichText, x, y, r, startAngle float64, opts pdf.CurvedTextOptions) error {
	return nil
}
func (w *layoutProbeWriter) DrawTextOnCircle(text string, x, y, r, startAngle float64, opts pdf.CurvedTextOptions) error {
	return nil
}
func (w *layoutProbeWriter) DrawClosedShape(shape pdf.ClosedShape, border, fill bool) error {
	return nil
}
func (w *layoutProbeWriter) EnableTaggedPDF(value bool)                               { w.base.EnableTaggedPDF(value) }
func (w *layoutProbeWriter) Circle(x, y, r float64, border, fill, reverse bool) error { return nil }
func (w *layoutProbeWriter) ClosedShapeBounds(shape pdf.ClosedShape) (pdf.Bounds, error) {
	return shape.Bounds()
}
func (w *layoutProbeWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	return nil
}
func (w *layoutProbeWriter) FontColor() colors.Color { return w.base.FontColor() }
func (w *layoutProbeWriter) Fonts() []*font.Font     { return w.base.Fonts() }
func (w *layoutProbeWriter) FontSize() float64       { return w.base.FontSize() }
func (w *layoutProbeWriter) ImageDimensions(data []byte) (width, height int, err error) {
	return w.base.ImageDimensions(data)
}
func (w *layoutProbeWriter) SVGDimensions(data []byte) (width, height int, err error) {
	return w.base.SVGDimensions(data)
}
func (w *layoutProbeWriter) SVGDimensionsFromFile(filename string) (width, height int, err error) {
	return w.base.SVGDimensionsFromFile(filename)
}
func (w *layoutProbeWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
	return w.base.ImageDimensionsFromFile(filename)
}
func (w *layoutProbeWriter) Line(x, y, angle, length float64) {}
func (w *layoutProbeWriter) LineSpacing() float64             { return w.base.LineSpacing() }
func (w *layoutProbeWriter) LineTo(x, y float64)              {}
func (w *layoutProbeWriter) Loc() (x, y float64)              { return 0, 0 }
func (w *layoutProbeWriter) MoveTo(x, y float64)              {}
func (w *layoutProbeWriter) NewPage()                         {}
func (w *layoutProbeWriter) Path(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *layoutProbeWriter) Pie(x, y, r, startAngle, endAngle float64, border, fill, reverse bool) error {
	return nil
}
func (w *layoutProbeWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	return nil
}
func (w *layoutProbeWriter) Print(text string) error { return nil }
func (w *layoutProbeWriter) PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *layoutProbeWriter) PrintSVG(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *layoutProbeWriter) PrintSVGFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *layoutProbeWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *layoutProbeWriter) PaintImageFile(filename string, x, y, width, height, opacity float64) error {
	return nil
}
func (w *layoutProbeWriter) PaintLinearGradient(lg *pdf.LinearGradient) error                { return nil }
func (w *layoutProbeWriter) PaintRadialGradient(rg *pdf.RadialGradient) error                { return nil }
func (w *layoutProbeWriter) PaintSweepBand(sb *pdf.SweepBand) error                          { return nil }
func (w *layoutProbeWriter) PrintParagraph(para []*rich_text.RichText, opts options.Options) {}
func (w *layoutProbeWriter) PrintRichText(text *rich_text.RichText)                          {}
func (w *layoutProbeWriter) Rectangle(x, y, width, height float64, border bool, fill bool)   {}
func (w *layoutProbeWriter) Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool) {
}
func (w *layoutProbeWriter) Rotate(angle, x, y float64, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *layoutProbeWriter) SetFillColor(value any) (prev colors.Color) {
	return w.base.FontColor()
}
func (w *layoutProbeWriter) SetFillLinearGradient(lg *pdf.LinearGradient) error { return nil }
func (w *layoutProbeWriter) SetFillRadialGradient(rg *pdf.RadialGradient) error { return nil }
func (w *layoutProbeWriter) ClearFillGradient()                                 {}
func (w *layoutProbeWriter) SetLineLinearGradient(lg *pdf.LinearGradient) error { return nil }
func (w *layoutProbeWriter) SetLineRadialGradient(rg *pdf.RadialGradient) error { return nil }
func (w *layoutProbeWriter) ClearLineGradient()                                 {}
func (w *layoutProbeWriter) SetFont(name string, size float64, opts options.Options) ([]*font.Font, error) {
	return w.base.SetFont(name, size, opts)
}
func (w *layoutProbeWriter) SetLineCapStyle(style string) (prev string) { return "" }
func (w *layoutProbeWriter) SetLineColor(value colors.Color) (prev colors.Color) {
	return 0
}
func (w *layoutProbeWriter) SetLineDashPattern(pattern string) (prev string) { return "" }
func (w *layoutProbeWriter) SetLineSpacing(lineSpacing float64) (prev float64) {
	return w.base.SetLineSpacing(lineSpacing)
}
func (w *layoutProbeWriter) SetLineWidth(width float64)  {}
func (w *layoutProbeWriter) SetLanguage(language string) { w.base.SetLanguage(language) }
func (w *layoutProbeWriter) SetSVGBlendMode(mode pdf.SVGBlendMode) pdf.SVGBlendMode {
	return w.base.SetSVGBlendMode(mode)
}
func (w *layoutProbeWriter) SetSVGGradientStopOpacityMode(mode pdf.SVGGradientStopOpacityMode) pdf.SVGGradientStopOpacityMode {
	return w.base.SetSVGGradientStopOpacityMode(mode)
}
func (w *layoutProbeWriter) SetStrikeout(strikeout bool) (prev bool) {
	return w.base.SetStrikeout(strikeout)
}
func (w *layoutProbeWriter) SetUnderline(underline bool) (prev bool) {
	return w.base.SetUnderline(underline)
}
func (w *layoutProbeWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	return nil
}
func (w *layoutProbeWriter) Strikeout() bool { return w.base.Strikeout() }
func (w *layoutProbeWriter) Stroke() error   { return nil }
func (w *layoutProbeWriter) Underline() bool { return w.base.Underline() }
func (w *layoutProbeWriter) TaggedPDFEnabled() bool {
	return w.base.TaggedPDFEnabled()
}
func (w *layoutProbeWriter) WithAccessibilityArtifact(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *layoutProbeWriter) WithAccessibilityTag(tag string, opts pdf.AccessibilityOptions, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
