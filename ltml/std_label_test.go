package ltml

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/rich_text"
)

func defaultTestFonts(t testing.TB) []*font.Font {
	t.Helper()

	fontSource, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	face, err := font.New("Helvetica", options.Options{"size": 12.0}, font.FontSources{fontSource})
	if err != nil {
		t.Fatal(err)
	}
	return []*font.Font{face}
}

type labelTestWriter struct {
	fonts          []*font.Font
	fontColor      colors.Color
	fillColor      colors.Color
	fontSize       float64
	lineSpacing    float64
	strikeout      bool
	underline      bool
	moves          [][2]float64
	printed        []*rich_text.RichText
	clipped        []*rich_text.RichText
	clippedText    []string
	paragraphOpts  []options.Options
	printedPages   []int
	plainPrinted   []string
	plainPages     []int
	rotations      []rotationCall
	curvedCount    int
	curvedStarts   []float64
	curvedOpts     []pdf.CurvedTextOptions
	pageCount      int
	rectPages      []int
	fillRectPages  []int
	linearPaints   []*pdf.LinearGradient
	radialPaints   []*pdf.RadialGradient
	lineColors     []colors.Color
	lineWidths     []float64
	lineLinear     []*pdf.LinearGradient
	lineRadial     []*pdf.RadialGradient
	lineClears     int
	imagePaints    []paintedImageCall
	fileDimensions map[string][2]int
	t              testing.TB
}

type rotationCall struct {
	angle float64
	x     float64
	y     float64
}

func (w *labelTestWriter) FontColor() colors.Color { return w.fontColor }
func (w *labelTestWriter) Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error {
	return nil
}
func (w *labelTestWriter) Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error {
	return nil
}
func (w *labelTestWriter) Clip(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *labelTestWriter) ClipClosedShape(shape pdf.ClosedShape, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *labelTestWriter) ClipRichText(text *rich_text.RichText, fn func()) error {
	w.clipped = append(w.clipped, text)
	if fn != nil {
		fn()
	}
	return nil
}
func (w *labelTestWriter) ClipText(text string, fn func()) error {
	w.clippedText = append(w.clippedText, text)
	if fn != nil {
		fn()
	}
	return nil
}
func (w *labelTestWriter) Circle(x, y, r float64, border, fill, reverse bool) error { return nil }
func (w *labelTestWriter) ClosedShapeBounds(shape pdf.ClosedShape) (pdf.Bounds, error) {
	return shape.Bounds()
}
func (w *labelTestWriter) CompressEmbeddedFonts(bool) *pdf.DocWriter { return nil }
func (w *labelTestWriter) CompressPages(bool) *pdf.DocWriter         { return nil }
func (w *labelTestWriter) CompressToUnicode(bool) *pdf.DocWriter     { return nil }
func (w *labelTestWriter) DrawRichTextOnCircle(text *rich_text.RichText, x, y, r, startAngle float64, opts pdf.CurvedTextOptions) error {
	w.curvedCount++
	w.curvedStarts = append(w.curvedStarts, startAngle)
	w.curvedOpts = append(w.curvedOpts, opts)
	w.printed = append(w.printed, text)
	w.printedPages = append(w.printedPages, w.pageCount)
	return nil
}
func (w *labelTestWriter) DrawTextOnCircle(text string, x, y, r, startAngle float64, opts pdf.CurvedTextOptions) error {
	w.curvedCount++
	w.curvedStarts = append(w.curvedStarts, startAngle)
	w.curvedOpts = append(w.curvedOpts, opts)
	w.plainPrinted = append(w.plainPrinted, text)
	w.plainPages = append(w.plainPages, w.pageCount)
	return nil
}
func (w *labelTestWriter) DrawClosedShape(shape pdf.ClosedShape, border, fill bool) error {
	switch shape.Kind {
	case pdf.ClosedShapeCircle:
		r := shape.Radius
		if r == 0 {
			r = shape.RadiusX
		}
		return w.Circle(shape.Center.X, shape.Center.Y, r, border, fill, shape.Reverse)
	case pdf.ClosedShapeEllipse:
		return w.Ellipse(shape.Center.X, shape.Center.Y, shape.RadiusX, shape.RadiusY, border, fill, shape.Reverse)
	case pdf.ClosedShapePolygon:
		return w.Polygon(shape.Center.X, shape.Center.Y, shape.Radius, shape.Sides, border, fill, shape.Reverse, shape.Rotation)
	case pdf.ClosedShapeStar:
		return w.Star(shape.Center.X, shape.Center.Y, shape.Radius, shape.InnerRadius, shape.Points, border, fill, shape.Reverse, shape.Rotation)
	default:
		return nil
	}
}
func (w *labelTestWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	return nil
}
func (w *labelTestWriter) Fonts() []*font.Font {
	if len(w.fonts) == 0 && w.t != nil {
		w.fonts = defaultTestFonts(w.t)
	}
	return w.fonts
}
func (w *labelTestWriter) FontSize() float64 { return w.fontSize }
func (w *labelTestWriter) ImageDimensions(data []byte) (width, height int, err error) {
	return 0, 0, nil
}
func (w *labelTestWriter) SVGDimensions(data []byte) (width, height int, err error) {
	return 0, 0, nil
}
func (w *labelTestWriter) SVGDimensionsFromFile(filename string) (width, height int, err error) {
	return 0, 0, nil
}
func (w *labelTestWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
	if dims, ok := w.fileDimensions[filename]; ok {
		return dims[0], dims[1], nil
	}
	return 0, 0, nil
}
func (w *labelTestWriter) LineSpacing() float64 {
	if w.lineSpacing == 0 {
		return 1.0
	}
	return w.lineSpacing
}
func (w *labelTestWriter) SetLineCapStyle(style string) (prev string) { return "" }
func (w *labelTestWriter) Line(x, y, angle, length float64)           {}
func (w *labelTestWriter) LineTo(x, y float64)                        {}
func (w *labelTestWriter) Loc() (x, y float64) {
	if len(w.moves) == 0 {
		return 0, 0
	}
	last := w.moves[len(w.moves)-1]
	return last[0], last[1]
}
func (w *labelTestWriter) MoveTo(x, y float64) { w.moves = append(w.moves, [2]float64{x, y}) }
func (w *labelTestWriter) NewPage()            { w.pageCount++ }
func (w *labelTestWriter) Print(text string) error {
	w.plainPrinted = append(w.plainPrinted, text)
	w.plainPages = append(w.plainPages, w.pageCount)
	return nil
}
func (w *labelTestWriter) PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *labelTestWriter) PrintSVG(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *labelTestWriter) PrintSVGFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *labelTestWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *labelTestWriter) PaintImageFile(filename string, x, y, width, height, opacity float64) error {
	w.imagePaints = append(w.imagePaints, paintedImageCall{
		filename: filename,
		x:        x,
		y:        y,
		width:    width,
		height:   height,
		opacity:  opacity,
	})
	return nil
}
func (w *labelTestWriter) PaintLinearGradient(lg *pdf.LinearGradient) error {
	w.linearPaints = append(w.linearPaints, lg)
	return nil
}
func (w *labelTestWriter) PaintRadialGradient(rg *pdf.RadialGradient) error {
	w.radialPaints = append(w.radialPaints, rg)
	return nil
}
func (w *labelTestWriter) PrintParagraph(para []*rich_text.RichText, opts options.Options) {
	w.paragraphOpts = append(w.paragraphOpts, opts)
	for _, line := range para {
		w.printed = append(w.printed, line)
		w.printedPages = append(w.printedPages, w.pageCount)
	}
}
func (w *labelTestWriter) PrintRichText(text *rich_text.RichText) {
	w.printed = append(w.printed, text)
	w.printedPages = append(w.printedPages, w.pageCount)
}
func (w *labelTestWriter) Pie(x, y, r, startAngle, endAngle float64, border, fill, reverse bool) error {
	return nil
}
func (w *labelTestWriter) Path(fn func()) error {
	fn()
	return nil
}
func (w *labelTestWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	return nil
}
func (w *labelTestWriter) Rectangle(x, y, width, height float64, border bool, fill bool) {}
func (w *labelTestWriter) Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool) {
	w.rectPages = append(w.rectPages, w.pageCount)
	if fill {
		w.fillRectPages = append(w.fillRectPages, w.pageCount)
	}
}
func (w *labelTestWriter) Rotate(angle, x, y float64, fn func()) error {
	w.rotations = append(w.rotations, rotationCall{angle: angle, x: x, y: y})
	if fn != nil {
		fn()
	}
	return nil
}
func (w *labelTestWriter) AddFont(family string, opts options.Options) ([]*font.Font, error) {
	return w.Fonts(), nil
}
func (w *labelTestWriter) SetFont(name string, size float64, opts options.Options) ([]*font.Font, error) {
	w.fontSize = size
	if color, ok := opts["color"].(colors.Color); ok {
		w.fontColor = color
	}
	return w.Fonts(), nil
}
func (w *labelTestWriter) SetFillColor(value any) (prev colors.Color) {
	prev = w.fillColor
	switch value := value.(type) {
	case colors.Color:
		w.fillColor = value
	case int:
		w.fillColor = colors.Color(value)
	case int32:
		w.fillColor = colors.Color(value)
	}
	return prev
}
func (w *labelTestWriter) SetFillLinearGradient(lg *pdf.LinearGradient) error { return nil }
func (w *labelTestWriter) SetFillRadialGradient(rg *pdf.RadialGradient) error { return nil }
func (w *labelTestWriter) ClearFillGradient()                                 {}
func (w *labelTestWriter) SetLineLinearGradient(lg *pdf.LinearGradient) error {
	w.lineLinear = append(w.lineLinear, lg)
	return nil
}
func (w *labelTestWriter) SetLineRadialGradient(rg *pdf.RadialGradient) error {
	w.lineRadial = append(w.lineRadial, rg)
	return nil
}
func (w *labelTestWriter) ClearLineGradient() { w.lineClears++ }
func (w *labelTestWriter) SetLineColor(value colors.Color) (prev colors.Color) {
	w.lineColors = append(w.lineColors, value)
	return 0
}
func (w *labelTestWriter) SetLineDashPattern(pattern string) (prev string)   { return "" }
func (w *labelTestWriter) SetLineSpacing(lineSpacing float64) (prev float64) { return w.lineSpacing }
func (w *labelTestWriter) SetLineWidth(width float64)                        { w.lineWidths = append(w.lineWidths, width) }
func (w *labelTestWriter) SetLanguage(language string)                       {}
func (w *labelTestWriter) SetSVGBlendMode(pdf.SVGBlendMode) pdf.SVGBlendMode {
	return pdf.SVGBlendModeRespect
}
func (w *labelTestWriter) SetSVGGradientStopOpacityMode(pdf.SVGGradientStopOpacityMode) pdf.SVGGradientStopOpacityMode {
	return pdf.SVGGradientStopOpacityModeSoftMask
}
func (w *labelTestWriter) SetStrikeout(strikeout bool) (prev bool) {
	prev = w.strikeout
	w.strikeout = strikeout
	return prev
}
func (w *labelTestWriter) SetUnderline(underline bool) (prev bool) {
	prev = w.underline
	w.underline = underline
	return prev
}
func (w *labelTestWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	return nil
}
func (w *labelTestWriter) Stroke() error              { return nil }
func (w *labelTestWriter) Strikeout() bool            { return w.strikeout }
func (w *labelTestWriter) Underline() bool            { return w.underline }
func (w *labelTestWriter) EnableTaggedPDF(value bool) {}
func (w *labelTestWriter) TaggedPDFEnabled() bool     { return false }
func (w *labelTestWriter) WithAccessibilityArtifact(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}
func (w *labelTestWriter) WithAccessibilityTag(tag string, opts pdf.AccessibilityOptions, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}

func TestStdLabel_AddTextWithFont_NormalizesXMLWhitespace(t *testing.T) {
	l := &StdLabel{}
	font := &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}

	l.AddTextWithFont("\n        Four score and seven years ago\n        our fathers brought forth.\n", font)

	if len(l.textPieces) != 1 {
		t.Fatalf("expected 1 text piece, got %d", len(l.textPieces))
	}
	got := l.textPieces[0].ResolvedText(nil)
	want := "Four score and seven years ago our fathers brought forth."
	if got != want {
		t.Fatalf("normalized text = %q, want %q", got, want)
	}
}

func TestStdLabel_PreferredHeight_EmptyLabelUsesFontLineHeight(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	w := &labelTestWriter{t: t, lineSpacing: 1.25}

	got := mustPreferredHeight(t, l, w)
	want := 15.0
	if got != want {
		t.Fatalf("PreferredHeight() = %v, want %v", got, want)
	}
}

func TestStdLabel_DrawContent_PrintsRichText(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetLeft(10)
	l.SetTop(20)
	l.AddText("Hello")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}

	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.moves) != 1 {
		t.Fatalf("MoveTo count = %d, want 1", len(w.moves))
	}
	if len(w.printed) != 1 {
		t.Fatalf("PrintRichText count = %d, want 1", len(w.printed))
	}
	if got := w.printed[0].String(); got != "Hello" {
		t.Fatalf("printed text = %q, want %q", got, "Hello")
	}
	if len(w.rotations) != 0 {
		t.Fatalf("rotation count = %d, want 0", len(w.rotations))
	}
}

func TestStdLabel_DrawContent_TextFillClipsAndPaintsSolidBrush(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetLeft(10)
	l.SetTop(20)
	l.SetWidth(120)
	l.SetHeight(40)
	l.SetAttrs(map[string]string{"text-fill": "Gold"})
	l.AddText("Hello")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.printed) != 0 {
		t.Fatalf("PrintRichText count = %d, want 0", len(w.printed))
	}
	if len(w.clipped) != 1 {
		t.Fatalf("ClipRichText count = %d, want 1", len(w.clipped))
	}
	if len(w.fillRectPages) != 1 {
		t.Fatalf("fill rect count = %d, want 1", len(w.fillRectPages))
	}
}

func TestStdLabel_DrawContent_TextFillClipsAndPaintsGradientBrush(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetLeft(10)
	l.SetTop(20)
	l.SetWidth(120)
	l.SetHeight(40)
	l.SetAttrs(map[string]string{
		"text-fill.kind":  "linear-gradient",
		"text-fill.x0":    "0",
		"text-fill.y0":    "0",
		"text-fill.x1":    "120",
		"text-fill.y1":    "0",
		"text-fill.stops": "0:Blue,1:Gold",
	})
	l.AddText("Hello")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.clipped) != 1 {
		t.Fatalf("ClipRichText count = %d, want 1", len(w.clipped))
	}
	if len(w.linearPaints) != 1 {
		t.Fatalf("linear paint count = %d, want 1", len(w.linearPaints))
	}
	if got := w.linearPaints[0].X0; got != 10 {
		t.Fatalf("gradient x0 = %v, want 10", got)
	}
}

func TestSample_TextFillClipping_ExercisesLabelAndParagraphTextFill(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_046_text_fill_clipping.ltml"))
	if err != nil {
		t.Fatal(err)
	}
	writer := &labelTestWriter{
		t: t,
		fileDimensions: map[string][2]int{
			filepath.Join(filepath.Dir(sampleFile("test_046_text_fill_clipping.ltml")), "../../pdf/testdata/testimg.jpg"): {640, 480},
		},
	}

	if err := doc.Print(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.clipped) == 0 {
		t.Fatal("expected sample to clip rich text")
	}
	if len(writer.linearPaints) == 0 {
		t.Fatal("expected sample to paint a gradient through clipped text")
	}
	if len(writer.radialPaints) == 0 {
		t.Fatal("expected sample to paint a radial gradient through clipped text")
	}
	if len(writer.imagePaints) == 0 {
		t.Fatal("expected sample to paint an image through clipped text")
	}
}

func TestSample_TextFillClipping_FileExists(t *testing.T) {
	if _, err := os.Stat(sampleFile("test_046_text_fill_clipping.ltml")); err != nil {
		t.Fatal(err)
	}
}

func TestStdLabel_DrawContent_AngleRotatesAroundLeftAnchor(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.angle = 45
	l.SetLeft(10)
	l.SetTop(20)
	l.AddText("Hello")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	rt := l.fittedRichText(w)

	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.rotations) != 1 {
		t.Fatalf("rotation count = %d, want 1", len(w.rotations))
	}
	call := w.rotations[0]
	if call.angle != 45 {
		t.Fatalf("rotation angle = %v, want 45", call.angle)
	}
	if math.Abs(call.x-ContentLeft(l)) > 0.001 {
		t.Fatalf("rotation x = %v, want %v", call.x, ContentLeft(l))
	}
	if math.Abs(call.y-(ContentTop(l)+rt.Ascent())) > 0.001 {
		t.Fatalf("rotation y = %v, want %v", call.y, ContentTop(l)+rt.Ascent())
	}
}

func TestStdLabel_RichText_ReappliesFontsWhenUsingCachedRichText(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.AddText("Hello")

	probe := &mockWriter{t: t}
	if got := l.RichText(probe); got == nil {
		t.Fatal("expected cached rich text to be built")
	}

	render := &mockWriter{t: t}
	if got := l.RichText(render); got == nil {
		t.Fatal("expected cached rich text to be returned")
	}
	if len(render.setFontCalls) == 0 {
		t.Fatal("expected cached rich text path to apply fonts to the render writer")
	}
}

func TestStdLabel_LeaderRespectsTextAlignAndAngle(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.angle = 30
	l.textAlign = HAlignRight
	l.textAlignSet = true
	l.SetLeft(10)
	l.SetTop(20)
	l.SetWidth(120)
	l.AddText("Beta")
	l.AddInlineWithFont(&StdLeader{text: "*"}, l.font)
	l.AddText("Gamma")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	rt := l.layoutRichText(w)
	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.moves) != 1 {
		t.Fatalf("move count = %d, want 1", len(w.moves))
	}
	wantX := ContentRight(l) - rt.Width()
	wantY := ContentTop(l) + rt.Ascent()
	if math.Abs(w.moves[0][0]-wantX) > 0.001 {
		t.Fatalf("move x = %v, want %v", w.moves[0][0], wantX)
	}
	if math.Abs(w.moves[0][1]-wantY) > 0.001 {
		t.Fatalf("move y = %v, want %v", w.moves[0][1], wantY)
	}
	if len(w.rotations) != 1 {
		t.Fatalf("rotation count = %d, want 1", len(w.rotations))
	}
	if math.Abs(w.rotations[0].x-ContentRight(l)) > 0.001 {
		t.Fatalf("rotation x = %v, want %v", w.rotations[0].x, ContentRight(l))
	}
	if math.Abs(w.rotations[0].y-wantY) > 0.001 {
		t.Fatalf("rotation y = %v, want %v", w.rotations[0].y, wantY)
	}
}

func TestStdLabel_DrawContent_TextAlignAffectsAnchor(t *testing.T) {
	cases := []struct {
		name       string
		align      HAlign
		wantAnchor func(*StdLabel, *rich_text.RichText) (float64, float64)
		wantStart  func(*StdLabel, *rich_text.RichText) (float64, float64)
	}{
		{
			name:  "center",
			align: HAlignCenter,
			wantAnchor: func(l *StdLabel, rt *rich_text.RichText) (float64, float64) {
				return (ContentLeft(l) + ContentRight(l)) / 2, ContentTop(l) + rt.Ascent()
			},
			wantStart: func(l *StdLabel, rt *rich_text.RichText) (float64, float64) {
				return ((ContentLeft(l)+ContentRight(l))/2 - (rt.Width() / 2)), ContentTop(l) + rt.Ascent()
			},
		},
		{
			name:  "right",
			align: HAlignRight,
			wantAnchor: func(l *StdLabel, rt *rich_text.RichText) (float64, float64) {
				return ContentRight(l), ContentTop(l) + rt.Ascent()
			},
			wantStart: func(l *StdLabel, rt *rich_text.RichText) (float64, float64) {
				return ContentRight(l) - rt.Width(), ContentTop(l) + rt.Ascent()
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := &StdLabel{}
			l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
			l.angle = 30
			l.textAlign = tc.align
			l.textAlignSet = true
			l.SetLeft(10)
			l.SetTop(20)
			l.SetWidth(100)
			l.AddText("Hello")

			w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
			rt := l.fittedRichText(w)

			if err := l.DrawContent(w); err != nil {
				t.Fatal(err)
			}
			if len(w.moves) != 1 {
				t.Fatalf("move count = %d, want 1", len(w.moves))
			}
			wantX, wantY := tc.wantStart(l, rt)
			if math.Abs(w.moves[0][0]-wantX) > 0.001 {
				t.Fatalf("move x = %v, want %v", w.moves[0][0], wantX)
			}
			if math.Abs(w.moves[0][1]-wantY) > 0.001 {
				t.Fatalf("move y = %v, want %v", w.moves[0][1], wantY)
			}
			if len(w.rotations) != 1 {
				t.Fatalf("rotation count = %d, want 1", len(w.rotations))
			}
			wantAnchorX, wantAnchorY := tc.wantAnchor(l, rt)
			if math.Abs(w.rotations[0].x-wantAnchorX) > 0.001 {
				t.Fatalf("rotation x = %v, want %v", w.rotations[0].x, wantAnchorX)
			}
			if math.Abs(w.rotations[0].y-wantAnchorY) > 0.001 {
				t.Fatalf("rotation y = %v, want %v", w.rotations[0].y, wantAnchorY)
			}
		})
	}
}

func TestStdLabel_DrawContent_TextVAlignMiddleAffectsAnchor(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.textAlign = HAlignCenter
	l.textAlignSet = true
	l.textVAlign = VAlignMiddle
	l.SetLeft(10)
	l.SetTop(20)
	l.SetWidth(120)
	l.SetHeight(60)
	l.AddText("Hello")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	rt := l.fittedRichText(w)

	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.moves) != 1 {
		t.Fatalf("move count = %d, want 1", len(w.moves))
	}
	textHeight := rt.Ascent() - rt.Descent()
	wantY := ContentTop(l) + (ContentHeight(l)-textHeight)/2 + rt.Ascent()
	if math.Abs(w.moves[0][1]-wantY) > 0.001 {
		t.Fatalf("move y = %v, want %v", w.moves[0][1], wantY)
	}
}

func TestStdLabel_DrawContent_DefaultsToRightAlignInRTLContainer(t *testing.T) {
	container := positionedContainer(0, 0, 200, 100)
	container.dirExplicit = true
	container.dir = DirRTL

	l := &StdLabel{}
	if err := l.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetWidth(100)
	l.SetLeft(0)
	l.SetTop(20)
	l.AddText("Hello")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	rt := l.fittedRichText(w)

	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.moves) != 1 {
		t.Fatalf("MoveTo count = %d, want 1", len(w.moves))
	}
	wantX := ContentRight(l) - rt.Width()
	if got := w.moves[0][0]; math.Abs(got-wantX) > 0.001 {
		t.Fatalf("move x = %v, want %v", got, wantX)
	}
}

func TestStdLabel_LogicalAndPhysicalTextAlignment(t *testing.T) {
	tests := []struct {
		name  string
		dir   string
		align string
		want  HAlign
	}{
		{name: "LTR start", dir: "ltr", align: "start", want: HAlignLeft},
		{name: "LTR end", dir: "ltr", align: "end", want: HAlignRight},
		{name: "RTL start", dir: "rtl", align: "start", want: HAlignRight},
		{name: "RTL end", dir: "rtl", align: "end", want: HAlignLeft},
		{name: "RTL physical left", dir: "rtl", align: "left", want: HAlignLeft},
		{name: "RTL physical right", dir: "rtl", align: "right", want: HAlignRight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := positionedContainer(0, 0, 200, 100)
			parent.SetAttrs(map[string]string{"dir": tt.dir})
			label := &StdLabel{}
			if err := label.SetContainer(parent); err != nil {
				t.Fatal(err)
			}
			label.SetAttrs(map[string]string{"text-align": tt.align})
			label.SetLeft(10)
			label.SetWidth(100)

			if got := label.resolvedTextAlign(); got != tt.want {
				t.Fatalf("resolved alignment = %s, want %s", got, tt.want)
			}
			wantX := ContentLeft(label)
			if tt.want == HAlignRight {
				wantX = ContentRight(label)
			}
			if got := label.textAnchorX(); got != wantX {
				t.Fatalf("text anchor x = %v, want %v", got, wantX)
			}
		})
	}
}

func TestStdLabel_LogicalAlignmentUsesExplicitLabelDirection(t *testing.T) {
	parent := positionedContainer(0, 0, 200, 100)
	parent.SetAttrs(map[string]string{"dir": "rtl"})
	label := &StdLabel{}
	if err := label.SetContainer(parent); err != nil {
		t.Fatal(err)
	}
	label.SetAttrs(map[string]string{"dir": "ltr", "text-align": "start"})

	if got := label.resolvedTextAlign(); got != HAlignLeft {
		t.Fatalf("explicit LTR label alignment = %s, want left", got)
	}
}

func TestStdParagraph_DrawContent_DefaultsToRightAlignInRTLContainer(t *testing.T) {
	container := positionedContainer(0, 0, 200, 100)
	container.dirExplicit = true
	container.dir = DirRTL

	p := &StdParagraph{}
	if err := p.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.SetWidth(100)
	p.SetLeft(0)
	p.SetTop(20)
	p.AddText("Hello world")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.paragraphOpts) != 1 {
		t.Fatalf("paragraph opts count = %d, want 1", len(w.paragraphOpts))
	}
	if got := w.paragraphOpts[0]["text-align"]; got != "right" {
		t.Fatalf("paragraph text-align = %v, want right", got)
	}
}

func TestStdLabel_DrawContent_ShrinksToFitWidth(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetWidth(35)
	l.shrinkToFit = true
	l.AddText("Hello world")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}

	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.printed) != 1 {
		t.Fatalf("PrintRichText count = %d, want 1", len(w.printed))
	}
	got := w.printed[0]
	if got.Width() > ContentWidth(l)+0.001 {
		t.Fatalf("printed width = %v, want <= %v", got.Width(), ContentWidth(l))
	}
	if gotWidth := got.Width(); gotWidth >= l.RichText(w).Width() {
		t.Fatalf("printed width = %v, want shrink from %v", gotWidth, l.RichText(w).Width())
	}
	assertAllLeafFontSizesBelow(t, got, 12)
}

func TestStdLabel_LeaderShrinksToFitWidth(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetWidth(45)
	l.shrinkToFit = true
	l.AddText("Beta")
	l.AddInlineWithFont(&StdLeader{text: "."}, l.font)
	l.AddText("Gamma")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	rt := l.layoutRichText(w)
	if rt.Width() > ContentWidth(l)+0.001 {
		t.Fatalf("leader layout width = %v, want <= %v", rt.Width(), ContentWidth(l))
	}
	if rt.Width() >= l.RichText(w).Width() {
		t.Fatalf("leader layout width = %v, want shrink from %v", rt.Width(), l.RichText(w).Width())
	}
}

func TestStdLabel_DrawContent_DoesNotShrinkWithoutFit(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetWidth(35)
	l.AddText("Hello world")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}

	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	got := w.printed[0]
	if math.Abs(got.Width()-l.RichText(w).Width()) > 0.001 {
		t.Fatalf("printed width = %v, want %v", got.Width(), l.RichText(w).Width())
	}
	assertAllLeafFontSizesEqual(t, got, 12)
}

func TestStdLabel_FittedRichText_DoesNotShrinkWhenTextFits(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetWidth(200)
	l.shrinkToFit = true
	l.AddText("Hello")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}

	got := l.fittedRichText(w)
	if got != l.RichText(w) {
		t.Fatalf("fittedRichText should return the original rich text when already fitting")
	}
}

func TestStdLabel_FittedRichText_ScalesInlineSpansProportionally(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetWidth(35)
	l.shrinkToFit = true
	l.AddText("Hello ")
	l.AddTextWithFont("big", &FontStyle{id: "big", entries: []fontEntry{{name: "Helvetica"}}, size: 18})

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}

	got := l.fittedRichText(w)
	sizes := leafFontSizes(got)
	if len(sizes) != 2 {
		t.Fatalf("leaf font size count = %d, want 2", len(sizes))
	}
	ratio := sizes[1] / sizes[0]
	if math.Abs(ratio-1.5) > 0.001 {
		t.Fatalf("scaled ratio = %v, want 1.5", ratio)
	}
}

func TestStdLabel_FittedRichText_RespectsMinimumFontSize(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetWidth(5)
	l.shrinkToFit = true
	l.AddText("This heading is much too long to fit")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}

	got := l.fittedRichText(w)
	assertAllLeafFontSizesAtLeast(t, got, 6)
}

func TestStdLabel_PreferredHeight_UsesShrunkTextHeight(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetWidth(20)
	l.shrinkToFit = true
	l.AddText("Hello world")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}

	got := mustPreferredHeight(t, l, w)
	unshrunk := l.RichText(w).Leading()*w.LineSpacing() + NonContentHeight(l)
	if got >= unshrunk {
		t.Fatalf("PreferredHeight() = %v, want less than unshrunk %v", got, unshrunk)
	}
}

func TestStdLabel_DrawContent_AngleUsesFittedWidthForCenterAnchor(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.SetLeft(10)
	l.SetTop(20)
	l.SetWidth(35)
	l.shrinkToFit = true
	l.angle = 30
	l.textAlign = HAlignCenter
	l.textAlignSet = true
	l.AddText("Hello world")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	rt := l.fittedRichText(w)

	if err := l.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.moves) != 1 {
		t.Fatalf("move count = %d, want 1", len(w.moves))
	}
	centerX := (ContentLeft(l) + ContentRight(l)) / 2
	centerY := ContentTop(l) + rt.Ascent()
	wantX := centerX - (rt.Width() / 2)
	wantY := centerY
	if math.Abs(w.moves[0][0]-wantX) > 0.001 {
		t.Fatalf("move x = %v, want %v", w.moves[0][0], wantX)
	}
	if math.Abs(w.moves[0][1]-wantY) > 0.001 {
		t.Fatalf("move y = %v, want %v", w.moves[0][1], wantY)
	}
	if len(w.rotations) != 1 {
		t.Fatalf("rotation count = %d, want 1", len(w.rotations))
	}
	if math.Abs(w.rotations[0].x-centerX) > 0.001 {
		t.Fatalf("rotation x = %v, want %v", w.rotations[0].x, centerX)
	}
	if math.Abs(w.rotations[0].y-centerY) > 0.001 {
		t.Fatalf("rotation y = %v, want %v", w.rotations[0].y, centerY)
	}
}

func TestStdLabel_DrawContent_AngleSupportsDynamicContent(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <label angle="45" text-align="right" width="80">Page <pageno /></label>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	doc.Root().documentPageNo = 7

	label := doc.Root().Page(0).children[0].(*StdLabel)
	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}

	if err := label.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.rotations) != 1 {
		t.Fatalf("rotation count = %d, want 1", len(w.rotations))
	}
	if len(w.printed) != 1 || w.printed[0].String() != "Page 7" {
		t.Fatalf("printed = %q, want %q", printedText(w.printed), "Page 7")
	}
}

func TestParse_LabelRuleTextAlignmentSurvivesDirectAttrs(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <style>label.centered { text-align: center; text-valign: middle; }</style>
  <page>
    <label class="centered" width="100" height="40">Hello</label>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	label := doc.Root().Page(0).children[0].(*StdLabel)
	if !label.textAlignSet || label.textAlign != HAlignCenter {
		t.Fatalf("text align = %v set=%v, want center set=true", label.textAlign, label.textAlignSet)
	}
	if label.textVAlign != VAlignMiddle {
		t.Fatalf("text valign = %v, want middle", label.textVAlign)
	}
}

func TestParse_LabelDefineTextAlignmentSurvivesDirectAttrs(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <define id="right-label" tag="label" text-align="right" text-valign="bottom" />
  <page>
    <right-label width="100" height="40">Hello</right-label>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	label := doc.Root().Page(0).children[0].(*StdLabel)
	if !label.textAlignSet || label.textAlign != HAlignRight {
		t.Fatalf("text align = %v set=%v, want right set=true", label.textAlign, label.textAlignSet)
	}
	if label.textVAlign != VAlignBottom {
		t.Fatalf("text valign = %v, want bottom", label.textVAlign)
	}
}

func TestStdLabel_FittedRichText_DynamicContentUsesResolvedPageNumber(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <label width="10" fit="shrink">Page <pageno /></label>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	doc.Root().documentPageNo = 9

	page := doc.Root().Page(0)
	label, ok := page.children[0].(*StdLabel)
	if !ok {
		t.Fatalf("child type = %T, want *StdLabel", page.children[0])
	}

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	got := label.fittedRichText(w)
	if got.String() != "Page 9" {
		t.Fatalf("fitted text = %q, want %q", got.String(), "Page 9")
	}
}

func TestStdLabel_RichText_UsesFontStyleColorInsteadOfWriterState(t *testing.T) {
	l := &StdLabel{}
	l.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	l.AddText("Color check")

	w := &labelTestWriter{
		t:         t,
		fonts:     defaultTestFonts(t),
		fontColor: NamedColor("LemonChiffon"),
		fillColor: NamedColor("LemonChiffon"),
	}

	rt := l.RichText(w)
	for _, color := range leafColors(rt) {
		if color != colors.Black {
			t.Fatalf("leaf color = %v, want %v", color, colors.Black)
		}
	}
}

func TestParse_LabelAndBrAlias(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <label>Hello <span font.weight="Bold">world</span></label>
    <br/>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	if page == nil {
		t.Fatal("page is nil")
	}
	if len(page.children) != 2 {
		t.Fatalf("child count = %d, want 2", len(page.children))
	}
	if _, ok := page.children[0].(*StdLabel); !ok {
		t.Fatalf("first child type = %T, want *StdLabel", page.children[0])
	}
	if _, ok := page.children[1].(*StdLabel); !ok {
		t.Fatalf("second child type = %T, want *StdLabel", page.children[1])
	}

	label := page.children[0].(*StdLabel)
	if len(label.textPieces) != 2 {
		t.Fatalf("text piece count = %d, want 2", len(label.textPieces))
	}
	if got := label.textPieces[0].ResolvedText(nil); got != "Hello " {
		t.Fatalf("piece 0 = %q, want %q", got, "Hello ")
	}
	if got := label.textPieces[1].ResolvedText(nil); got != "world" {
		t.Fatalf("piece 1 = %q, want %q", got, "world")
	}
}

func TestParse_PredefinedInlineSpanAliases(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <label><i>italic</i><u>underline</u><s>strike</s></label>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	if page == nil {
		t.Fatal("page is nil")
	}
	label, ok := page.children[0].(*StdLabel)
	if !ok {
		t.Fatalf("first child type = %T, want *StdLabel", page.children[0])
	}
	if len(label.textPieces) != 3 {
		t.Fatalf("text piece count = %d, want 3", len(label.textPieces))
	}

	cases := []struct {
		index     int
		text      string
		style     string
		underline bool
		strikeout bool
	}{
		{index: 0, text: "italic", style: "Italic"},
		{index: 1, text: "underline", underline: true},
		{index: 2, text: "strike", strikeout: true},
	}

	for _, tc := range cases {
		piece := label.textPieces[tc.index]
		if got := piece.ResolvedText(nil); got != tc.text {
			t.Fatalf("piece %d text = %q, want %q", tc.index, got, tc.text)
		}
		if piece.font == nil {
			t.Fatalf("piece %d font is nil", tc.index)
		}
		if piece.font.style != tc.style {
			t.Fatalf("piece %d style = %q, want %q", tc.index, piece.font.style, tc.style)
		}
		if piece.font.underline != tc.underline {
			t.Fatalf("piece %d underline = %v, want %v", tc.index, piece.font.underline, tc.underline)
		}
		if piece.font.strikeout != tc.strikeout {
			t.Fatalf("piece %d strikeout = %v, want %v", tc.index, piece.font.strikeout, tc.strikeout)
		}
	}
}

func leafFontSizes(rt *rich_text.RichText) []float64 {
	var sizes []float64
	rt.VisitAll(func(p *rich_text.RichText) {
		if p.IsLeaf() && p.Len() > 0 {
			sizes = append(sizes, p.FontSize)
		}
	})
	return sizes
}

func leafColors(rt *rich_text.RichText) []colors.Color {
	var result []colors.Color
	rt.VisitAll(func(p *rich_text.RichText) {
		if p.IsLeaf() && p.Len() > 0 {
			result = append(result, p.Color)
		}
	})
	return result
}

func printedText(pieces []*rich_text.RichText) string {
	if len(pieces) == 0 || pieces[0] == nil {
		return ""
	}
	return pieces[0].String()
}

func assertAllLeafFontSizesBelow(t testing.TB, rt *rich_text.RichText, max float64) {
	t.Helper()
	for _, size := range leafFontSizes(rt) {
		if size >= max {
			t.Fatalf("leaf font size = %v, want < %v", size, max)
		}
	}
}

func assertAllLeafFontSizesEqual(t testing.TB, rt *rich_text.RichText, want float64) {
	t.Helper()
	for _, size := range leafFontSizes(rt) {
		if math.Abs(size-want) > 0.001 {
			t.Fatalf("leaf font size = %v, want %v", size, want)
		}
	}
}

func assertAllLeafFontSizesAtLeast(t testing.TB, rt *rich_text.RichText, min float64) {
	t.Helper()
	for _, size := range leafFontSizes(rt) {
		if size < min {
			t.Fatalf("leaf font size = %v, want >= %v", size, min)
		}
	}
}
