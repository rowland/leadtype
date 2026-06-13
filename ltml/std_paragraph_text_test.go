package ltml

import (
	"math"
	"strings"
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

func (w *bulletTestWriter) ClipClosedShape(shape pdf.ClosedShape, fn func()) error {
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

func (w *bulletTestWriter) ClosedShapeBounds(shape pdf.ClosedShape) (pdf.Bounds, error) {
	return shape.Bounds()
}

func (w *bulletTestWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "ellipse", x: x, y: y, a: rx, b: ry, border: border, fill: fill, reverse: reverse})
	return nil
}

func (w *bulletTestWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "polygon", x: x, y: y, a: r, i: sides, border: border, fill: fill, reverse: reverse, rotation: rotation})
	return nil
}

func (w *bulletTestWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	w.shapeCalls = append(w.shapeCalls, shapeCall{name: "star", x: x, y: y, a: r1, b: r2, i: points, border: border, fill: fill, reverse: reverse, rotation: rotation})
	return nil
}

func (w *bulletTestWriter) DrawClosedShape(shape pdf.ClosedShape, border, fill bool) error {
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

func TestStdParagraph_LeaderWrapsWithLeaderOnFinalLine(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.SetWidth(90)
	p.AddText("A very long introduction heading ")
	p.AddInlineWithFont(&StdLeader{text: "."}, p.font)
	p.AddText("2")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) < 2 {
		t.Fatalf("wrapped line count = %d, want at least 2", len(lines))
	}
	if strings.Contains(lines[0].String(), "2") || strings.Contains(lines[0].String(), ".") {
		t.Fatalf("first line = %q, want no leader or page number", lines[0].String())
	}
	last := lines[len(lines)-1].String()
	if !strings.Contains(last, "2") || !strings.Contains(last, ".") {
		t.Fatalf("last line = %q, want final-line leader and page number", last)
	}
}

func TestStdParagraph_LeaderPathStillDrawsBullet(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{text: "*", width: 18, font: p.font}}
	p.SetWidth(120)
	p.AddText("Alpha")
	p.AddInlineWithFont(&StdLeader{text: "-"}, p.font)
	p.AddText("Omega")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.plainPrinted) != 1 || w.plainPrinted[0] != "*" {
		t.Fatalf("bullet output = %#v, want [*]", w.plainPrinted)
	}
	if len(w.printed) != 1 || !strings.Contains(w.printed[0].String(), "Omega") || !strings.Contains(w.printed[0].String(), "-") {
		t.Fatalf("printed lines = %#v, want leader paragraph output", w.printed)
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
	p.bullets = []*BulletStyle{{text: "*", width: 18, font: p.font}}
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

func TestStdParagraph_DrawContent_PlacesTextBulletInRTLSlot(t *testing.T) {
	container := positionedContainer(0, 0, 200, 100)
	container.dirExplicit = true
	container.dir = DirRTL

	p := &StdParagraph{}
	if err := p.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{text: "*", width: 24, font: p.font}}
	p.SetWidth(120)
	p.SetLeft(10)
	p.SetTop(20)
	p.AddText("Hello world")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.moves) < 3 {
		t.Fatalf("move count = %d, want at least 3", len(w.moves))
	}
	wantBulletX := ContentRight(p) - p.bulletTextWidth(w, p.Bullet())
	if got := w.moves[1][0]; math.Abs(got-wantBulletX) > 0.001 {
		t.Fatalf("bullet move x = %v, want %v", got, wantBulletX)
	}
	if got := w.moves[0][0]; math.Abs(got-ContentLeft(p)) > 0.001 {
		t.Fatalf("initial text move x = %v, want %v", got, ContentLeft(p))
	}
	if got := w.moves[2][0]; math.Abs(got-ContentLeft(p)) > 0.001 {
		t.Fatalf("post-bullet text move x = %v, want %v", got, ContentLeft(p))
	}
	if len(w.paragraphOpts) != 1 {
		t.Fatalf("paragraph opts count = %d, want 1", len(w.paragraphOpts))
	}
	if got := w.paragraphOpts[0]["text-align"]; got != "right" {
		t.Fatalf("paragraph text-align = %v, want right", got)
	}
}

func TestStdParagraph_DrawContent_AlignsTextBulletWithinLTRSlot(t *testing.T) {
	tests := []struct {
		name   string
		alignX string
		wantX  func(*StdParagraph, *labelTestWriter, *BulletStyle) float64
	}{
		{
			name:   "center",
			alignX: "center",
			wantX: func(p *StdParagraph, w *labelTestWriter, b *BulletStyle) float64 {
				return ContentLeft(p) + max((b.Width()-p.bulletTextWidth(w, b))/2, 0)
			},
		},
		{
			name:   "end",
			alignX: "end",
			wantX: func(p *StdParagraph, w *labelTestWriter, b *BulletStyle) float64 {
				return ContentLeft(p) + max(b.Width()-p.bulletTextWidth(w, b), 0)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &StdParagraph{}
			p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
			p.paragraphStyle = &ParagraphStyle{}
			p.bullets = []*BulletStyle{{text: "1.", width: 40, font: p.font, alignX: tc.alignX}}
			p.SetLeft(10)
			p.SetTop(20)
			p.SetWidth(140)
			p.AddText("Hello world")

			w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
			if err := p.DrawContent(w); err != nil {
				t.Fatal(err)
			}
			if len(w.moves) < 3 {
				t.Fatalf("move count = %d, want at least 3", len(w.moves))
			}
			if got, want := w.moves[1][0], tc.wantX(p, w, p.Bullet()); math.Abs(got-want) > 0.001 {
				t.Fatalf("bullet x = %v, want %v", got, want)
			}
		})
	}
}

func TestStdParagraph_SetAttrs_ParsesMultipleBulletReferences(t *testing.T) {
	scope := &Scope{}
	first := &BulletStyle{id: "first", text: "*", width: 12}
	second := &BulletStyle{id: "second", text: "-", width: 18}
	if err := scope.AddStyle(first); err != nil {
		t.Fatal(err)
	}
	if err := scope.AddStyle(second); err != nil {
		t.Fatal(err)
	}

	p := &StdParagraph{StdContainer: StdContainer{StdWidget: StdWidget{scope: scope}}}
	p.SetAttrs(map[string]string{"bullet": " first missing second "})

	if got := p.Bullets(); len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("bullets = %#v, want first and second", got)
	}
	if got := p.Bullet(); got != first {
		t.Fatalf("Bullet() = %#v, want first bullet", got)
	}

	p.SetAttrs(map[string]string{"bullet": "missing"})
	if got := p.Bullets(); len(got) != 0 {
		t.Fatalf("unknown bullets = %#v, want none", got)
	}
	if got := p.Bullet(); got != nil {
		t.Fatalf("Bullet() after unknown = %#v, want nil", got)
	}
}

func TestParagraphStyle_SetAttrs_ParsesMultipleBulletReferences(t *testing.T) {
	scope := &Scope{}
	first := &BulletStyle{id: "first", text: "*", width: 12}
	second := &BulletStyle{id: "second", text: "-", width: 18}
	if err := scope.AddStyle(first); err != nil {
		t.Fatal(err)
	}
	if err := scope.AddStyle(second); err != nil {
		t.Fatal(err)
	}

	style := &ParagraphStyle{scope: scope}
	style.SetAttrs(map[string]string{"bullet": "first second"})

	if got := style.Bullets(); len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("style bullets = %#v, want first and second", got)
	}
	if got := style.Bullet(); got != first {
		t.Fatalf("style Bullet() = %#v, want first bullet", got)
	}
}

func TestStdParagraph_MultipleBulletsSumIndentAndDrawInLTRSlots(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{
		{text: "*", width: 12, font: p.font},
		{text: "-", width: 18, font: p.font},
	}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.AddText("Hello world")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if got, want := p.textIndent(), 30.0; got != want {
		t.Fatalf("text indent = %v, want %v", got, want)
	}
	if got, want := p.lineWidth(), ContentWidth(p)-30; got != want {
		t.Fatalf("line width = %v, want %v", got, want)
	}
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(w.plainPrinted, ""); got != "*-" {
		t.Fatalf("plain bullets = %q, want *-", got)
	}
	if len(w.moves) < 4 {
		t.Fatalf("move count = %d, want at least 4", len(w.moves))
	}
	if got, want := w.moves[0][0], ContentLeft(p)+30; math.Abs(got-want) > 0.001 {
		t.Fatalf("initial text x = %v, want %v", got, want)
	}
	if got, want := w.moves[1][0], ContentLeft(p); math.Abs(got-want) > 0.001 {
		t.Fatalf("first bullet x = %v, want %v", got, want)
	}
	if got, want := w.moves[2][0], ContentLeft(p)+12; math.Abs(got-want) > 0.001 {
		t.Fatalf("second bullet x = %v, want %v", got, want)
	}
}

func TestStdParagraph_MultipleBulletsDrawInRTLSlots(t *testing.T) {
	container := positionedContainer(0, 0, 200, 100)
	container.dirExplicit = true
	container.dir = DirRTL

	p := &StdParagraph{}
	if err := p.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{
		{text: "*", width: 24, font: p.font},
		{text: "-", width: 18, font: p.font},
	}
	p.SetWidth(120)
	p.SetLeft(10)
	p.SetTop(20)
	p.AddText("Hello world")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(w.plainPrinted, ""); got != "*-" {
		t.Fatalf("plain bullets = %q, want *-", got)
	}
	if len(w.moves) < 4 {
		t.Fatalf("move count = %d, want at least 4", len(w.moves))
	}
	wantFirst := ContentRight(p) - p.bulletTextWidth(w, p.bullets[0])
	if got := w.moves[1][0]; math.Abs(got-wantFirst) > 0.001 {
		t.Fatalf("first RTL bullet x = %v, want %v", got, wantFirst)
	}
	secondSlotX := ContentRight(p) - p.bullets[0].Width() - p.bullets[1].Width()
	wantSecond := secondSlotX + max(p.bullets[1].Width()-p.bulletTextWidth(w, p.bullets[1]), 0)
	if got := w.moves[2][0]; math.Abs(got-wantSecond) > 0.001 {
		t.Fatalf("second RTL bullet x = %v, want %v", got, wantSecond)
	}
	if got := w.moves[0][0]; math.Abs(got-ContentLeft(p)) > 0.001 {
		t.Fatalf("initial text x = %v, want %v", got, ContentLeft(p))
	}
}

func TestParse_MultipleBulletParagraphPrintsAllBullets(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <bullet id="first" text="*" width="12pt" />
  <bullet id="second" text="-" width="18pt" />
  <page width="200pt" height="100pt">
    <p bullet="first second" width="120pt">Hello world</p>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(w.plainPrinted, ""); got != "*-" {
		t.Fatalf("plain bullets = %q, want *-", got)
	}
}

func TestStdParagraph_SplitForHeight_RespectsDefaultsAndSuppressesBullet(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = defaultLayouts["vbox"].Clone()

	p := &StdParagraph{}
	_ = p.SetContainer(page)
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.bullets = []*BulletStyle{{text: "*", width: 18, font: p.font}}
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

func TestStdParagraph_SplitForHeight_PreservesMultipleBulletIndent(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 200}}
	page.layout = defaultLayouts["vbox"].Clone()

	p := &StdParagraph{}
	_ = p.SetContainer(page)
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.bullets = []*BulletStyle{
		{text: "*", width: 18, font: p.font},
		{text: "-", width: 12, font: p.font},
	}
	p.orphans = 2
	p.widows = 2
	p.SetWidth(90)
	p.AddText("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.")

	w := &labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) < 5 {
		t.Fatalf("wrapped line count = %d, want at least 5", len(lines))
	}

	result, err := p.SplitForHeight(p.heightForLines(lines[:2], w), w)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected paragraph to split")
	}

	head := result.Head.(*StdParagraph)
	tail := result.Tail.(*StdParagraph)
	if head.suppressBullet {
		t.Fatal("head should keep bullets")
	}
	if !tail.suppressBullet {
		t.Fatal("tail should suppress continuation bullets")
	}
	if got, want := tail.continuationIndent, 30.0; got != want {
		t.Fatalf("tail continuation indent = %v, want %v", got, want)
	}
	if got, want := tail.textIndent(), 30.0; got != want {
		t.Fatalf("tail text indent = %v, want %v", got, want)
	}
}

func TestStdParagraph_DrawContent_PlacesImageBulletInRTLSlot(t *testing.T) {
	container := positionedContainer(0, 0, 200, 100)
	container.dirExplicit = true
	container.dir = DirRTL

	p := &StdParagraph{}
	if err := p.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{src: "fixture.svg", width: 24, height: 12}}
	p.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.svg", "<svg/>"))))
	p.SetWidth(120)
	p.SetLeft(10)
	p.SetTop(20)
	p.AddText("Hello world")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{
		t:           t,
		fonts:       defaultTestFonts(t),
		lineSpacing: 1.0,
		fileDimensions: map[string][2]int{
			"fixture.svg": {10, 10},
		},
	}}

	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.fileCalls) != 1 {
		t.Fatalf("image bullet calls = %d, want 1", len(w.fileCalls))
	}
	if w.fileCalls[0].width == nil {
		t.Fatal("image bullet width pointer is nil")
	}
	wantX := ContentRight(p) - *w.fileCalls[0].width
	if got := w.fileCalls[0].x; math.Abs(got-wantX) > 0.001 {
		t.Fatalf("image bullet x = %v, want %v", got, wantX)
	}
}

func TestStdParagraph_DrawContent_PlacesShapeBulletFlushRightInRTLSlot(t *testing.T) {
	container := positionedContainer(0, 0, 200, 100)
	container.dirExplicit = true
	container.dir = DirRTL

	p := &StdParagraph{}
	if err := p.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{shape: "circle", width: 24, height: 18}}
	p.SetWidth(120)
	p.SetLeft(10)
	p.SetTop(20)
	p.AddText("Hello world")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{
		t:           t,
		fonts:       defaultTestFonts(t),
		lineSpacing: 1.0,
	}}

	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.shapeCalls) != 1 {
		t.Fatalf("shape bullet calls = %d, want 1", len(w.shapeCalls))
	}
	call := w.shapeCalls[0]
	if call.name != "circle" {
		t.Fatalf("shape bullet name = %q, want circle", call.name)
	}
	wantX := ContentRight(p) - call.a
	if got := call.x; math.Abs(got-wantX) > 0.001 {
		t.Fatalf("shape bullet center x = %v, want %v", got, wantX)
	}
}

func TestStdParagraph_DrawContent_PreservesEllipseBulletWidth(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{shape: "ellipse", width: 24, rx: 9, ry: 12}}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.SetHeight(40)
	p.AddText("Hello")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{
		t:           t,
		fonts:       defaultTestFonts(t),
		lineSpacing: 1.0,
	}}
	if err := p.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.shapeCalls) != 1 {
		t.Fatalf("shape bullet calls = %d, want 1", len(w.shapeCalls))
	}
	call := w.shapeCalls[0]
	if call.name != "ellipse" {
		t.Fatalf("shape bullet name = %q, want ellipse", call.name)
	}
	if got := call.a; math.Abs(got-9) > 0.001 {
		t.Fatalf("ellipse bullet rx = %v, want 9", got)
	}
	if got := call.b; math.Abs(got-12) > 0.001 {
		t.Fatalf("ellipse bullet ry = %v, want 12", got)
	}
}

func TestStdParagraph_BulletLayout_ShapeRadiusDoesNotConsumeFullSlotWidth(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{shape: "circle", width: 24, r: 9}}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.SetHeight(40)
	p.AddText("Hello")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}}
	lines := p.Lines(w, p.lineWidth())
	layout := p.bulletLayout(w, p.Bullet(), lines[0], 10, 20+lines[0].Ascent(), p.textContentHeightForLines(lines, w))
	if got := layout.renderWidth; math.Abs(got-18) > 0.0001 {
		t.Fatalf("renderWidth = %v, want 18", got)
	}
	if got := layout.slotWidth; math.Abs(got-24) > 0.0001 {
		t.Fatalf("slotWidth = %v, want 24", got)
	}
}

func TestStdParagraph_PreferredHeight_AccountsForTallBullet(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{shape: "ellipse", width: 24, height: 24, rx: 9, ry: 12}}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.AddText("Hello")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) != 1 {
		t.Fatalf("line count = %d, want 1", len(lines))
	}
	got := p.contentHeightForLines(lines, w)
	if got < 24 {
		t.Fatalf("contentHeightForLines = %v, want >= 24", got)
	}
}

func TestStdParagraph_DrawContent_ImageBulletUsesPrintImageFile(t *testing.T) {
	p := &StdParagraph{}
	p.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.svg", "<svg/>"))))
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{src: "fixture.svg", width: 20, height: 10}}
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
	p.bullets = []*BulletStyle{{
		shape:    "star",
		width:    18,
		r:        9,
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
	}}
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
	if w.clipCount == 0 {
		t.Fatal("expected shape bullet to use closed-shape clipping")
	}
	if got := w.shapeCalls[len(w.shapeCalls)-1].name; got != "star" {
		t.Fatalf("last shape call = %q, want star border draw", got)
	}
	lines := p.Lines(w, p.lineWidth())
	layout := p.bulletLayout(w, p.Bullet(), lines[0], 10, 20+lines[0].Ascent(), p.textContentHeightForLines(lines, w))
	if layout.shape == nil {
		t.Fatal("shape layout missing closed shape")
	}
	if math.Abs(layout.shapeBounds.MinX-layout.renderX) > 0.0001 {
		t.Fatalf("shape left edge = %v, want %v", layout.shapeBounds.MinX, layout.renderX)
	}
	if math.Abs(layout.renderX-10) > 0.0001 {
		t.Fatalf("renderX = %v, want 10", layout.renderX)
	}
	if len(w.clippedText) != 0 {
		t.Fatalf("shape bullet should not use text clipping, got %#v", w.clippedText)
	}
}

func TestStdParagraph_DrawContent_ShapeBulletWithGradientPenUsesShapeBox(t *testing.T) {
	p := &StdParagraph{}
	shape := pdf.ClosedShape{Kind: pdf.ClosedShapeCircle, Center: pdf.Location{X: 20, Y: 30}, Radius: 10}
	bullet := &BulletStyle{
		shape: "circle",
		pen: &PenStyle{
			kind: PenKindLinearGradient,
			linearGradient: &pdf.LinearGradient{
				Stops: []pdf.GradientStop{
					{Position: 0, Color: NamedColor("Tomato")},
					{Position: 1, Color: NamedColor("SteelBlue")},
				},
			},
			linearPct: &linearGradientPct{X0: float64Ptr(0), Y0: float64Ptr(50), X1: float64Ptr(100), Y1: float64Ptr(50)},
		},
	}
	layout := paragraphBulletLayout{
		shape:       &shape,
		shapeBounds: pdf.Bounds{MinX: 10, MinY: 20, MaxX: 30, MaxY: 40},
	}
	w := &bulletTestWriter{}

	if err := p.drawShapeBullet(w, bullet, layout); err != nil {
		t.Fatal(err)
	}
	if len(w.lineLinear) != 1 {
		t.Fatalf("line linear gradient count = %d, want 1", len(w.lineLinear))
	}
	got := w.lineLinear[0]
	if got.X0 != 10 || got.Y0 != 30 || got.X1 != 30 || got.Y1 != 30 {
		t.Fatalf("gradient coords = %#v, want bullet shape-box coords", got)
	}
}

func TestStdParagraph_DrawContent_FourPointStarBulletRenders(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{
		shape:  "star",
		width:  24,
		r:      9,
		points: 4,
		r0:     4,
		brush:  &BrushStyle{id: "sky", color: NamedColor("LightSkyBlue")},
	}}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.SetHeight(40)
	p.AddText("Hello")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}}
	if err := p.DrawContent(w); err != nil {
		t.Fatalf("DrawContent returned error: %v", err)
	}
	if len(w.shapeCalls) != 1 {
		t.Fatalf("shape call count = %d, want 1", len(w.shapeCalls))
	}
	call := w.shapeCalls[0]
	if call.name != "star" {
		t.Fatalf("shape name = %q, want star", call.name)
	}
	if call.i != 4 {
		t.Fatalf("star points = %d, want 4", call.i)
	}
}

func TestStdParagraph_DrawContent_TwoPointStarBulletRenders(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{
		shape:  "star",
		width:  24,
		r:      9,
		points: 2,
		r0:     4,
		brush:  &BrushStyle{id: "sky", color: NamedColor("LightSkyBlue")},
	}}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.SetHeight(40)
	p.AddText("Hello")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}}
	if err := p.DrawContent(w); err != nil {
		t.Fatalf("DrawContent returned error: %v", err)
	}
	if len(w.shapeCalls) != 1 {
		t.Fatalf("shape call count = %d, want 1", len(w.shapeCalls))
	}
	call := w.shapeCalls[0]
	if call.name != "star" {
		t.Fatalf("shape name = %q, want star", call.name)
	}
	if call.i != 2 {
		t.Fatalf("star points = %d, want 2", call.i)
	}
}

func TestStdParagraph_DrawContent_ThreePointStarBulletRenders(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{
		shape:  "star",
		width:  24,
		r:      9,
		points: 3,
		r0:     4,
		brush:  &BrushStyle{id: "sky", color: NamedColor("LightSkyBlue")},
	}}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.SetHeight(40)
	p.AddText("Hello")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}}
	if err := p.DrawContent(w); err != nil {
		t.Fatalf("DrawContent returned error: %v", err)
	}
	if len(w.shapeCalls) != 1 {
		t.Fatalf("shape call count = %d, want 1", len(w.shapeCalls))
	}
	call := w.shapeCalls[0]
	if call.name != "star" {
		t.Fatalf("shape name = %q, want star", call.name)
	}
	if call.i != 3 {
		t.Fatalf("star points = %d, want 3", call.i)
	}
}

func TestStdParagraph_BulletLayout_HonorsExplicitHeightAboveLineHeight(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{shape: "circle", width: 24, height: 30, r: 6}}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(120)
	p.SetHeight(40)
	p.AddText("Hello")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) != 1 {
		t.Fatalf("line count = %d, want 1", len(lines))
	}

	layout := p.bulletLayout(w, p.Bullet(), lines[0], 10, 20+lines[0].Ascent(), p.textContentHeightForLines(lines, w))
	if math.Abs(layout.renderHeight-12) > 0.0001 {
		t.Fatalf("renderHeight = %v, want 12", layout.renderHeight)
	}
	if math.Abs(layout.renderWidth-12) > 0.0001 {
		t.Fatalf("renderWidth = %v, want 12", layout.renderWidth)
	}
	if math.Abs(layout.renderY-20) > 0.0001 {
		t.Fatalf("renderY = %v, want 20", layout.renderY)
	}
	if got := p.contentHeightForLines(lines, w); got < 30 {
		t.Fatalf("contentHeightForLines = %v, want >= 30", got)
	}
}

func TestStdParagraph_BulletLayout_AlignYMiddleCentersOnTextBlock(t *testing.T) {
	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.paragraphStyle = &ParagraphStyle{}
	p.bullets = []*BulletStyle{{shape: "circle", width: 24, height: 30, r: 6, alignY: "middle"}}
	p.SetLeft(10)
	p.SetTop(20)
	p.SetWidth(70)
	p.AddText("Hello world this wraps onto multiple lines for vertical centering.")

	w := &bulletTestWriter{labelTestWriter: labelTestWriter{t: t, fonts: defaultTestFonts(t), lineSpacing: 1.0}}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) < 2 {
		t.Fatalf("line count = %d, want at least 2", len(lines))
	}

	textHeight := p.textContentHeightForLines(lines, w)
	layout := p.bulletLayout(w, p.Bullet(), lines[0], 10, 20+lines[0].Ascent(), textHeight)
	wantY := 20 + (textHeight-layout.renderHeight)/2
	if math.Abs(layout.renderY-wantY) > 0.0001 {
		t.Fatalf("renderY = %v, want %v", layout.renderY, wantY)
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
	p.bullets = []*BulletStyle{{src: "fixture.jpg", width: 18, height: 12}}
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
