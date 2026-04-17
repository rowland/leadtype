package ltml

import (
	"math"
	"testing"

	"github.com/rowland/leadtype/pdf"
)

type bulletTestWriter struct {
	labelTestWriter
	fileCalls  []imageFilePrintCall
	shapeCalls []shapeCall
	pathCount  int
	clipCount  int
}

func (w *bulletTestWriter) Path(fn func()) error {
	w.pathCount++
	if fn != nil {
		fn()
	}
	return nil
}

func (w *bulletTestWriter) Clip(fn func()) error {
	w.clipCount++
	if fn != nil {
		fn()
	}
	return nil
}

func (w *bulletTestWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (float64, float64, error) {
	w.fileCalls = append(w.fileCalls, imageFilePrintCall{
		filename: filename,
		x:        x,
		y:        y,
		width:    width,
		height:   height,
	})
	return 0, 0, nil
}

func (w *bulletTestWriter) Circle(x, y, r float64, border, fill, reverse bool) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "circle", x: x, y: y, a: r, border: border, fill: fill, reverse: reverse})
	return nil
}

func (w *bulletTestWriter) CirclePath(x, y, r float64, reverse bool) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "circle-path", x: x, y: y, a: r, reverse: reverse})
	return nil
}

func (w *bulletTestWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "ellipse", x: x, y: y, a: rx, b: ry, border: border, fill: fill, reverse: reverse})
	return nil
}

func (w *bulletTestWriter) EllipsePath(x, y, rx, ry float64, reverse bool) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "ellipse-path", x: x, y: y, a: rx, b: ry, reverse: reverse})
	return nil
}

func (w *bulletTestWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "polygon", x: x, y: y, a: r, i: sides, border: border, fill: fill, reverse: reverse, rotation: rotation})
	return nil
}

func (w *bulletTestWriter) PolygonPath(x, y, r float64, sides int, reverse bool, rotation float64) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "polygon-path", x: x, y: y, a: r, i: sides, reverse: reverse, rotation: rotation})
	return nil
}

func (w *bulletTestWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "star", x: x, y: y, a: r1, b: r2, i: points, border: border, fill: fill, reverse: reverse, rotation: rotation})
	return nil
}

func (w *bulletTestWriter) StarPath(x, y, r1, r2 float64, points int, reverse bool, rotation float64) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "star-path", x: x, y: y, a: r1, b: r2, i: points, reverse: reverse, rotation: rotation})
	return nil
}

func TestStdParagraph_AddTextWithFont_NormalizesXMLWhitespace(t *testing.T) {
	p := &StdParagraph{}
	font := &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}

	p.AddTextWithFont("\n        Four score and seven years ago\n        our fathers brought forth.\n", font)

	if len(p.textPieces) != 1 {
		t.Fatalf("expected 1 text piece, got %d", len(p.textPieces))
	}
	got := p.textPieces[0].ResolvedText(nil)
	want := "Four score and seven years ago our fathers brought forth. "
	if got != want {
		t.Fatalf("normalized text = %q, want %q", got, want)
	}
}

func TestStdParagraph_AddTextWithFont_PreservesSpanBoundarySpaces(t *testing.T) {
	p := &StdParagraph{}
	body := &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	emph := &FontStyle{id: "emph", entries: []fontEntry{{name: "Helvetica"}}, size: 12, weight: "Bold"}

	p.AddTextWithFont("Hello ", body)
	p.AddTextWithFont("big", emph)
	p.AddTextWithFont(" world", body)

	if len(p.textPieces) != 3 {
		t.Fatalf("expected 3 text pieces, got %d", len(p.textPieces))
	}
	if got := p.textPieces[0].ResolvedText(nil); got != "Hello " {
		t.Fatalf("first piece = %q, want %q", got, "Hello ")
	}
	if got := p.textPieces[1].ResolvedText(nil); got != "big" {
		t.Fatalf("second piece = %q, want %q", got, "big")
	}
	if got := p.textPieces[2].ResolvedText(nil); got != " world" {
		t.Fatalf("third piece = %q, want %q", got, " world")
	}
}

func TestStdParagraph_RichText_ReappliesFontsWhenUsingCachedRichText(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.AddText("Hello")

	probe := &mockWriter{t: t}
	if got := p.RichText(probe); got == nil {
		t.Fatal("expected cached rich text to be built")
	}

	render := &mockWriter{t: t}
	if got := p.RichText(render); got == nil {
		t.Fatal("expected cached rich text to be returned")
	}
	if len(render.setFontCalls) == 0 {
		t.Fatal("expected cached rich text path to apply fonts to the render writer")
	}
}

func TestStdParagraph_DrawContent_TextFillClipsParagraphLinesAcrossWidgetRect(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(90)
	p.SetHeight(80)
	p.SetAttrs(map[string]string{"text-fill": "Gold"})
	p.AddText("Lorem ipsum dolor sit amet, consectetur adipiscing elit.")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) < 2 {
		t.Fatalf("wrapped line count = %d, want at least 2", len(lines))
	}
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.printed) != 0 {
		t.Fatalf("PrintRichText count = %d, want 0", len(w.printed))
	}
	if len(w.clipped) != len(lines) {
		t.Fatalf("ClipRichText count = %d, want %d", len(w.clipped), len(lines))
	}
	if len(w.fillRectPages) != len(lines) {
		t.Fatalf("fill rect count = %d, want %d", len(w.fillRectPages), len(lines))
	}
	if len(w.moves) < len(lines) {
		t.Fatalf("move count = %d, want at least %d", len(w.moves), len(lines))
	}
	lineMoves := w.moves[len(w.moves)-len(lines):]
	for i := 1; i < len(lines); i++ {
		if lineMoves[i][1] <= lineMoves[i-1][1] {
			t.Fatalf("line %d baseline y = %v, want greater than previous line y %v", i, lineMoves[i][1], lineMoves[i-1][1])
		}
	}
}

func TestStdParagraph_DrawContent_TextFillUsesParagraphAlignmentWithPlainBullet(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{TextStyle: TextStyle{textAlign: HAlignCenter, textAlignSet: true}}
	p.bullet = &BulletStyle{text: "*", width: 18, font: p.font}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.SetHeight(40)
	p.SetAttrs(map[string]string{
		"text-fill.kind":  "linear-gradient",
		"text-fill.x0":    "0",
		"text-fill.y0":    "0",
		"text-fill.x1":    "120",
		"text-fill.y1":    "0",
		"text-fill.stops": "0:Blue,1:Gold",
	})
	p.AddText("Hello")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	lines := p.Lines(w, p.lineWidth())
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.plainPrinted) != 1 || w.plainPrinted[0] != "*" {
		t.Fatalf("plain bullet output = %#v, want [*]", w.plainPrinted)
	}
	if len(w.clippedText) != 0 {
		t.Fatalf("bullet should not be clipped, got %#v", w.clippedText)
	}
	if len(w.clipped) != len(lines) {
		t.Fatalf("ClipRichText count = %d, want %d", len(w.clipped), len(lines))
	}
	if len(w.linearPaints) == 0 {
		t.Fatal("expected gradient paints")
	}
	opts := paragraphTextFillOptions(p)
	if got := opts.StringDefault("text-align", ""); got != "center" {
		t.Fatalf("text-align = %q, want center", got)
	}
	if got := opts.FloatDefault("width", 0); got != ContentWidth(p)-p.textIndent() {
		t.Fatalf("width = %v, want %v", got, ContentWidth(p)-p.textIndent())
	}
}

func TestStdParagraph_SplitForHeight_RespectsDefaultsAndSuppressesBullet(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = defaultLayouts["vbox"].Clone()

	p := &StdParagraph{}
	_ = p.SetContainer(page)
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.bullet = &BulletStyle{text: "*", width: 18, font: p.font}
	p.splitEnabled = true
	p.orphans = 2
	p.widows = 2
	p.SetWidth(90)
	p.AddText("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) < 5 {
		t.Fatalf("wrapped line count = %d, want at least 5", len(lines))
	}

	avail := p.heightForLines(lines[:2], w)
	result, err := p.SplitForHeight(avail, w)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected paragraph to split")
	}

	head := result.Head.(*StdParagraph)
	tail := result.Tail.(*StdParagraph)
	if len(head.splitLines) != 2 {
		t.Fatalf("head line count = %d, want 2", len(head.splitLines))
	}
	if len(tail.splitLines) != len(lines)-2 {
		t.Fatalf("tail line count = %d, want %d", len(tail.splitLines), len(lines)-2)
	}
	if head.suppressBullet {
		t.Fatal("head should keep bullet")
	}
	if !tail.suppressBullet {
		t.Fatal("tail should suppress bullet")
	}
	if tail.continuationIndent != p.bulletWidth() {
		t.Fatalf("tail continuation indent = %v, want %v", tail.continuationIndent, p.bulletWidth())
	}
}

func TestStdParagraph_DrawContent_ImageBulletUsesPrintImageFile(t *testing.T) {
	p := &StdParagraph{}
	p.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.svg", "<svg/>"))))
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullet = &BulletStyle{src: "fixture.svg", width: 20, height: 10}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.SetHeight(40)
	p.AddText("Hello")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{
		t:              t,
		fonts:          defaultTestFonts(t),
		lineSpacing:    1.0,
		fileDimensions: map[string][2]int{"fixture.svg": {100, 50}},
	}}
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.fileCalls) != 1 {
		t.Fatalf("image call count = %d, want 1", len(w.fileCalls))
	}
	call := w.fileCalls[0]
	if call.filename != "fixture.svg" {
		t.Fatalf("filename = %q, want fixture.svg", call.filename)
	}
	if call.x != 10 {
		t.Fatalf("x = %v, want 10", call.x)
	}
	if call.width == nil || *call.width != 20 {
		t.Fatalf("width = %v, want 20", call.width)
	}
	if call.height == nil || *call.height != 10 {
		t.Fatalf("height = %v, want 10", call.height)
	}
	if len(w.plainPrinted) != 0 {
		t.Fatalf("plainPrinted = %#v, want no text bullet output", w.plainPrinted)
	}
}

func TestStdParagraph_DrawContent_ShapeBulletWithGradientUsesClipPath(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullet = &BulletStyle{
		shape:    "star",
		width:    18,
		height:   18,
		points:   6,
		r0:       4,
		rotation: 15,
		brush: &BrushStyle{
			kind: BrushKindLinearGradient,
			linearGradient: &pdf.LinearGradient{
				Stops: []pdf.GradientStop{
					{Position: 0, Color: NamedColor("Blue")},
					{Position: 1, Color: NamedColor("Gold")},
				},
			},
		},
		pen: &PenStyle{id: "solid", color: NamedColor("black"), width: 1, pattern: "solid"},
	}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.SetHeight(40)
	p.AddText("Hello")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}}
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.linearPaints) != 1 {
		t.Fatalf("linear paint count = %d, want 1", len(w.linearPaints))
	}
	if len(w.shapeCalls) < 2 {
		t.Fatalf("shape call count = %d, want at least 2", len(w.shapeCalls))
	}
	if got := w.shapeCalls[0].name; got != "star-path" {
		t.Fatalf("first shape call = %q, want star-path", got)
	}
	if got := w.shapeCalls[len(w.shapeCalls)-1].name; got != "star" {
		t.Fatalf("last shape call = %q, want star border draw", got)
	}
	layout := p.bulletLayout(w, p.bullet, p.Lines(w, p.lineWidth())[0], 10, 20+p.Lines(w, p.lineWidth())[0].Ascent())
	bounds := bulletShapeGeometryBounds(p.bullet, layout.renderWidth, layout.renderHeight)
	if got := layout.centerX + bounds.minX; math.Abs(got-layout.renderX) > 0.0001 {
		t.Fatalf("shape left edge = %v, want %v", got, layout.renderX)
	}
	if math.Abs(layout.renderX-10) > 0.0001 {
		t.Fatalf("renderX = %v, want 10", layout.renderX)
	}
	if len(w.clippedText) != 0 {
		t.Fatalf("shape bullet should not use text clipping, got %#v", w.clippedText)
	}
}

func TestStdParagraph_SplitForHeight_ImageBulletSuppressesContinuationRendering(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = defaultLayouts["vbox"].Clone()

	p := &StdParagraph{}
	_ = p.SetContainer(page)
	p.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.jpg", "image-data"))))
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullet = &BulletStyle{src: "fixture.jpg", width: 18, height: 12}
	p.splitEnabled = true
	p.orphans = 2
	p.widows = 2
	p.SetWidth(90)
	p.AddText("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{
		t:              t,
		fonts:          defaultTestFonts(t),
		lineSpacing:    1.0,
		fileDimensions: map[string][2]int{"fixture.jpg": {120, 80}},
	}}
	lines := p.Lines(w, p.lineWidth())
	result, err := p.SplitForHeight(p.heightForLines(lines[:2], w), w)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected paragraph to split")
	}

	if err := result.Head.(*StdParagraph).DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if err := result.Tail.(*StdParagraph).DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.fileCalls) != 1 {
		t.Fatalf("image bullet should render only on head, got %d calls", len(w.fileCalls))
	}
}
