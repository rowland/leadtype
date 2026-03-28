package ltml

import (
	"math"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
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
	fonts         []*font.Font
	fontColor     colors.Color
	fillColor     colors.Color
	fontSize      float64
	lineSpacing   float64
	strikeout     bool
	underline     bool
	moves         [][2]float64
	printed       []*rich_text.RichText
	printedPages  []int
	plainPrinted  []string
	plainPages    []int
	rotations     []rotationCall
	pageCount     int
	rectPages     []int
	fillRectPages []int
	t             testing.TB
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
func (w *labelTestWriter) Circle(x, y, r float64, border, fill, reverse bool) error { return nil }
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
func (w *labelTestWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
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
func (w *labelTestWriter) Loc() (x, y float64)                        { return 0, 0 }
func (w *labelTestWriter) MoveTo(x, y float64)                        { w.moves = append(w.moves, [2]float64{x, y}) }
func (w *labelTestWriter) NewPage()                                   { w.pageCount++ }
func (w *labelTestWriter) Print(text string) error {
	w.plainPrinted = append(w.plainPrinted, text)
	w.plainPages = append(w.plainPages, w.pageCount)
	return nil
}
func (w *labelTestWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return 0, 0, nil
}
func (w *labelTestWriter) PrintParagraph(para []*rich_text.RichText, opts options.Options) {
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
func (w *labelTestWriter) SetFillColor(value interface{}) (prev colors.Color) {
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
func (w *labelTestWriter) SetLineColor(value colors.Color) (prev colors.Color) { return 0 }
func (w *labelTestWriter) SetLineDashPattern(pattern string) (prev string)     { return "" }
func (w *labelTestWriter) SetLineSpacing(lineSpacing float64) (prev float64)   { return w.lineSpacing }
func (w *labelTestWriter) SetLineWidth(width float64)                          {}
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
func (w *labelTestWriter) Stroke() error   { return nil }
func (w *labelTestWriter) Strikeout() bool { return w.strikeout }
func (w *labelTestWriter) Underline() bool { return w.underline }

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

	got := l.PreferredHeight(w)
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

func TestStdLabel_DrawContent_TextAlignAffectsAnchor(t *testing.T) {
	cases := []struct {
		name      string
		align     HAlign
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

	got := l.PreferredHeight(w)
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
	doc.ltmls[0].documentPageNo = 7

	label := doc.ltmls[0].Page(0).children[0].(*StdLabel)
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
	doc.ltmls[0].documentPageNo = 9

	page := doc.ltmls[0].Page(0)
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

	page := doc.ltmls[0].Page(0)
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
