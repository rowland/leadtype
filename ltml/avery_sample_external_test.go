package ltml_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	_ "github.com/rowland/leadtype/avery"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/ltml"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/rich_text"
)

func TestSample_PaperCatalogsRender(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	samples := []string{
		"test_051_paper_sizes.ltml",
		"test_052_large_paper_sizes.ltml",
		"test_053_avery_labels.ltml",
	}
	for _, name := range samples {
		name := name
		t.Run(name, func(t *testing.T) {
			sample := filepath.Join(filepath.Dir(file), "samples", name)
			doc, err := ltml.ParseFile(sample)
			if err != nil {
				t.Fatal(err)
			}
			w := &sampleTestWriter{t: t}
			if err := doc.Print(w); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type sampleTestWriter struct {
	fonts    []*font.Font
	fontSize float64
	t        testing.TB
}

func (w *sampleTestWriter) ensureFonts() []*font.Font {
	if len(w.fonts) != 0 {
		return w.fonts
	}
	fontSource, err := afm_fonts.Default()
	if err != nil {
		w.t.Fatal(err)
	}
	face, err := font.New("Helvetica", options.Options{"size": 12.0}, font.FontSources{fontSource})
	if err != nil {
		w.t.Fatal(err)
	}
	w.fonts = []*font.Font{face}
	return w.fonts
}

func (w *sampleTestWriter) Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error {
	return nil
}
func (w *sampleTestWriter) Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error {
	return nil
}
func (w *sampleTestWriter) Clip(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *sampleTestWriter) ClipClosedShape(shape pdf.ClosedShape, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *sampleTestWriter) ClipRichText(text *rich_text.RichText, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *sampleTestWriter) ClipText(text string, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *sampleTestWriter) ClosedShapeBounds(shape pdf.ClosedShape) (pdf.Bounds, error) {
	return shape.Bounds()
}
func (w *sampleTestWriter) DrawRichTextOnCircle(text *rich_text.RichText, x, y, r, startAngle float64, opts pdf.CurvedTextOptions) error {
	return nil
}
func (w *sampleTestWriter) DrawTextOnCircle(text string, x, y, r, startAngle float64, opts pdf.CurvedTextOptions) error {
	return nil
}
func (w *sampleTestWriter) DrawClosedShape(shape pdf.ClosedShape, border, fill bool) error {
	return nil
}
func (w *sampleTestWriter) EnableTaggedPDF(bool)                                     {}
func (w *sampleTestWriter) Circle(x, y, r float64, border, fill, reverse bool) error { return nil }
func (w *sampleTestWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	return nil
}
func (w *sampleTestWriter) FontColor() colors.Color { return colors.Black }
func (w *sampleTestWriter) Fonts() []*font.Font     { return w.ensureFonts() }
func (w *sampleTestWriter) FontSize() float64 {
	if w.fontSize == 0 {
		return 12
	}
	return w.fontSize
}
func (w *sampleTestWriter) ImageDimensions(data []byte) (width, height int, err error) {
	return 0, 0, nil
}
func (w *sampleTestWriter) SVGDimensions(data []byte) (width, height int, err error) {
	return 0, 0, nil
}
func (w *sampleTestWriter) SVGDimensionsFromFile(filename string) (width, height int, err error) {
	return 0, 0, nil
}
func (w *sampleTestWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
	return 0, 0, nil
}
func (w *sampleTestWriter) LineSpacing() float64                       { return 1.0 }
func (w *sampleTestWriter) SetLineCapStyle(style string) (prev string) { return "" }
func (w *sampleTestWriter) Line(x, y, angle, length float64)           {}
func (w *sampleTestWriter) LineTo(x, y float64)                        {}
func (w *sampleTestWriter) Loc() (x, y float64)                        { return 0, 0 }
func (w *sampleTestWriter) MoveTo(x, y float64)                        {}
func (w *sampleTestWriter) NewPage()                                   {}
func (w *sampleTestWriter) Print(text string) error                    { return nil }
func (w *sampleTestWriter) PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *sampleTestWriter) PrintSVG(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *sampleTestWriter) PrintSVGFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *sampleTestWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *sampleTestWriter) PaintImageFile(filename string, x, y, width, height, opacity float64) error {
	return nil
}
func (w *sampleTestWriter) PaintLinearGradient(lg *pdf.LinearGradient) error { return nil }
func (w *sampleTestWriter) PaintRadialGradient(rg *pdf.RadialGradient) error { return nil }
func (w *sampleTestWriter) PrintParagraph(para []*rich_text.RichText, options options.Options) {
}
func (w *sampleTestWriter) PrintRichText(text *rich_text.RichText) {}
func (w *sampleTestWriter) Path(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *sampleTestWriter) Pie(x, y, r, startAngle, endAngle float64, border, fill, reverse bool) error {
	return nil
}
func (w *sampleTestWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	return nil
}
func (w *sampleTestWriter) Rectangle(x, y, width, height float64, border bool, fill bool) {}
func (w *sampleTestWriter) Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool) {
}
func (w *sampleTestWriter) Rotate(angle, x, y float64, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *sampleTestWriter) AddFont(family string, options options.Options) ([]*font.Font, error) {
	return w.ensureFonts(), nil
}
func (w *sampleTestWriter) SetFont(name string, size float64, options options.Options) ([]*font.Font, error) {
	w.fontSize = size
	return w.ensureFonts(), nil
}
func (w *sampleTestWriter) SetFillColor(value any) (prev colors.Color)          { return colors.Black }
func (w *sampleTestWriter) SetFillLinearGradient(lg *pdf.LinearGradient) error  { return nil }
func (w *sampleTestWriter) SetFillRadialGradient(rg *pdf.RadialGradient) error  { return nil }
func (w *sampleTestWriter) ClearFillGradient()                                  {}
func (w *sampleTestWriter) SetLineLinearGradient(lg *pdf.LinearGradient) error  { return nil }
func (w *sampleTestWriter) SetLineRadialGradient(rg *pdf.RadialGradient) error  { return nil }
func (w *sampleTestWriter) ClearLineGradient()                                  {}
func (w *sampleTestWriter) SetLineColor(value colors.Color) (prev colors.Color) { return colors.Black }
func (w *sampleTestWriter) SetLineDashPattern(pattern string) (prev string)     { return "" }
func (w *sampleTestWriter) SetLineSpacing(lineSpacing float64) (prev float64)   { return 1.0 }
func (w *sampleTestWriter) SetLineWidth(width float64)                          {}
func (w *sampleTestWriter) SetStrikeout(strikeout bool) (prev bool)             { return false }
func (w *sampleTestWriter) SetUnderline(underline bool) (prev bool)             { return false }
func (w *sampleTestWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	return nil
}
func (w *sampleTestWriter) Stroke() error          { return nil }
func (w *sampleTestWriter) Strikeout() bool        { return false }
func (w *sampleTestWriter) TaggedPDFEnabled() bool { return false }
func (w *sampleTestWriter) Underline() bool        { return false }
func (w *sampleTestWriter) WithAccessibilityArtifact(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *sampleTestWriter) WithAccessibilityTag(tag string, opts pdf.AccessibilityOptions, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
