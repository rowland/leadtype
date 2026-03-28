package ltml

import (
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
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

func (w *layoutProbeWriter) AddFont(family string, opts options.Options) ([]*font.Font, error) {
	return w.base.AddFont(family, opts)
}
func (w *layoutProbeWriter) Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error {
	return nil
}
func (w *layoutProbeWriter) Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error {
	return nil
}
func (w *layoutProbeWriter) Circle(x, y, r float64, border, fill, reverse bool) error { return nil }
func (w *layoutProbeWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	return nil
}
func (w *layoutProbeWriter) FontColor() colors.Color { return w.base.FontColor() }
func (w *layoutProbeWriter) Fonts() []*font.Font     { return w.base.Fonts() }
func (w *layoutProbeWriter) FontSize() float64       { return w.base.FontSize() }
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
func (w *layoutProbeWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
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
func (w *layoutProbeWriter) SetLineWidth(width float64) {}
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
