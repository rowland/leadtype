package ltml

import (
	"testing"

	"github.com/rowland/leadtype/pdf"
)

type shapeCall struct {
	name     string
	x        float64
	y        float64
	a        float64
	b        float64
	c        float64
	d        float64
	i        int
	border   bool
	fill     bool
	reverse  bool
	rotation float64
	move     bool
}

type shapeTestWriter struct {
	labelTestWriter
	calls         []shapeCall
	appendPaths   []pdf.ClosedShape
	events        []string
	inPath        bool
	strokes       int
	fills         int
	fillAndStrokes int
	pathRuns      int
	curves        int
}

func (w *shapeTestWriter) AppendClosedShapePath(shape pdf.ClosedShape) error {
	w.appendPaths = append(w.appendPaths, shape)
	w.events = append(w.events, "append")
	return nil
}

func (w *shapeTestWriter) AppendPiePath(x, y, r, startAngle, endAngle float64, reverse bool) error {
	w.events = append(w.events, "appendPie")
	return nil
}

func (w *shapeTestWriter) AppendArchPath(x, y, r1, r2, startAngle, endAngle float64, reverse bool) error {
	w.events = append(w.events, "appendArch")
	return nil
}

func (w *shapeTestWriter) Fill() error {
	if !w.inPath {
		t := w.t
		if t != nil {
			t.Fatalf("Fill() called outside Path()")
		}
	}
	w.fills++
	w.events = append(w.events, "fill")
	return nil
}

func (w *shapeTestWriter) FillAndStroke() error {
	if !w.inPath {
		t := w.t
		if t != nil {
			t.Fatalf("FillAndStroke() called outside Path()")
		}
	}
	w.fillAndStrokes++
	w.events = append(w.events, "fillAndStroke")
	return nil
}

func (w *shapeTestWriter) Print(text string) error {
	w.events = append(w.events, "print:"+text)
	return w.labelTestWriter.Print(text)
}

func (w *shapeTestWriter) Circle(x, y, r float64, border, fill, reverse bool) error {
	w.calls = append(w.calls, shapeCall{name: "circle", x: x, y: y, a: r, border: border, fill: fill, reverse: reverse})
	return nil
}

func (w *shapeTestWriter) ClosedShapeBounds(shape pdf.ClosedShape) (pdf.Bounds, error) {
	return shape.Bounds()
}

func (w *shapeTestWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	w.calls = append(w.calls, shapeCall{name: "ellipse", x: x, y: y, a: rx, b: ry, border: border, fill: fill, reverse: reverse})
	return nil
}

func (w *shapeTestWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	w.calls = append(w.calls, shapeCall{name: "polygon", x: x, y: y, a: r, i: sides, border: border, fill: fill, reverse: reverse, rotation: rotation})
	return nil
}

func (w *shapeTestWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	w.calls = append(w.calls, shapeCall{name: "star", x: x, y: y, a: r1, b: r2, i: points, border: border, fill: fill, reverse: reverse, rotation: rotation})
	return nil
}

func (w *shapeTestWriter) ClipClosedShape(shape pdf.ClosedShape, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}

func (w *shapeTestWriter) DrawClosedShape(shape pdf.ClosedShape, border, fill bool) error {
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

func (w *shapeTestWriter) Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error {
	w.calls = append(w.calls, shapeCall{name: "arc", x: x, y: y, a: r, b: startAngle, c: endAngle, move: moveToStart})
	return nil
}

func (w *shapeTestWriter) Pie(x, y, r, startAngle, endAngle float64, border, fill, reverse bool) error {
	w.calls = append(w.calls, shapeCall{name: "pie", x: x, y: y, a: r, b: startAngle, c: endAngle, border: border, fill: fill, reverse: reverse})
	return nil
}

func (w *shapeTestWriter) Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error {
	w.calls = append(w.calls, shapeCall{name: "arch", x: x, y: y, a: r1, b: r2, c: startAngle, d: endAngle, border: border, fill: fill, reverse: reverse})
	return nil
}

func (w *shapeTestWriter) Path(fn func()) error {
	w.inPath = true
	w.pathRuns++
	if fn != nil {
		fn()
	}
	w.inPath = false
	return nil
}

func (w *shapeTestWriter) CurvePoints(points []pdf.Location) error {
	w.curves++
	return nil
}

func (w *shapeTestWriter) Stroke() error {
	if !w.inPath {
		t := w.t
		if t != nil {
			t.Fatalf("Stroke() called outside Path()")
		}
	}
	w.strokes++
	return nil
}

func TestStdCircle_DrawContent_CompositeTireUsesSinglePathFill(t *testing.T) {
	outer := &StdCircle{}
	outer.SetLeft(0)
	outer.SetTop(0)
	outer.SetWidth(100)
	outer.SetHeight(100)
	outer.fill = &BrushStyle{id: "fill", color: NamedColor("gold")}

	inner := &StdCircle{}
	inner.SetAttrs(map[string]string{"r": "20pt", "reverse": "true"})
	outer.AddChild(inner)
	if err := inner.SetContainer(outer); err != nil {
		t.Fatal(err)
	}

	w := &shapeTestWriter{}
	w.t = t
	if err := outer.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if w.pathRuns != 1 {
		t.Fatalf("pathRuns = %d, want 1", w.pathRuns)
	}
	if w.fills != 1 {
		t.Fatalf("fills = %d, want 1", w.fills)
	}
	if len(w.appendPaths) != 2 {
		t.Fatalf("appendPaths = %d, want 2", len(w.appendPaths))
	}
	if w.appendPaths[0].Reverse {
		t.Fatal("outer append path should not be reversed")
	}
	if !w.appendPaths[1].Reverse {
		t.Fatal("inner append path should be reversed")
	}
	if len(w.calls) != 0 {
		t.Fatalf("immediate shape calls = %#v, want none", w.calls)
	}
}

func TestStdCircle_DrawContent_CompositePaintsLabelAfterFill(t *testing.T) {
	outer := &StdCircle{}
	outer.SetLeft(0)
	outer.SetTop(0)
	outer.SetWidth(100)
	outer.SetHeight(100)
	outer.fill = &BrushStyle{id: "fill", color: NamedColor("gold")}

	inner := &StdCircle{}
	inner.SetAttrs(map[string]string{"r": "20pt", "reverse": "true"})

	w := &shapeTestWriter{}
	w.t = t
	child := &paintOrderChild{w: w}
	inner.AddChild(child)
	outer.AddChild(inner)
	if err := inner.SetContainer(outer); err != nil {
		t.Fatal(err)
	}
	if err := child.SetContainer(inner); err != nil {
		t.Fatal(err)
	}

	if err := outer.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	fillIdx := -1
	childIdx := -1
	for i, event := range w.events {
		if event == "fill" {
			fillIdx = i
		}
		if event == "childDraw" {
			childIdx = i
		}
	}
	if fillIdx < 0 {
		t.Fatalf("events = %#v, want fill", w.events)
	}
	if childIdx < 0 {
		t.Fatalf("events = %#v, want child draw", w.events)
	}
	if childIdx < fillIdx {
		t.Fatalf("child drew before fill: events = %#v", w.events)
	}
}

type paintOrderChild struct {
	StdWidget
	w *shapeTestWriter
}

func (c *paintOrderChild) DrawContent(w Writer) error {
	c.w.events = append(c.w.events, "childDraw")
	return nil
}

func TestStdCircle_DrawContent_ReversedChildWithFillRedrawsAfterComposite(t *testing.T) {
	outer := &StdCircle{}
	outer.SetLeft(0)
	outer.SetTop(0)
	outer.SetWidth(100)
	outer.SetHeight(100)
	outer.fill = &BrushStyle{id: "fill", color: NamedColor("gold")}

	inner := &StdCircle{}
	inner.SetAttrs(map[string]string{"r": "20pt", "reverse": "true", "fill": "SkyBlue"})

	outer.AddChild(inner)
	if err := inner.SetContainer(outer); err != nil {
		t.Fatal(err)
	}

	w := &shapeTestWriter{}
	w.t = t
	if err := outer.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.appendPaths) != 2 || !w.appendPaths[1].Reverse {
		t.Fatalf("appendPaths = %#v, want outer + reversed inner", w.appendPaths)
	}
	if len(w.calls) != 1 {
		t.Fatalf("immediate shape calls = %#v, want one inner redraw", w.calls)
	}
	call := w.calls[0]
	if !call.fill || call.border {
		t.Fatalf("inner redraw = %#v, want fill without border", call)
	}
}

func TestStdCircle_DrawContent_LabelOnlyUsesImmediateDraw(t *testing.T) {
	circle := &StdCircle{}
	circle.SetLeft(0)
	circle.SetTop(0)
	circle.SetWidth(100)
	circle.SetHeight(100)
	circle.fill = &BrushStyle{id: "fill", color: NamedColor("gold")}

	w := &shapeTestWriter{}
	w.t = t
	child := &paintOrderChild{w: w}
	circle.AddChild(child)
	if err := child.SetContainer(circle); err != nil {
		t.Fatal(err)
	}

	if err := circle.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if w.pathRuns != 0 {
		t.Fatalf("pathRuns = %d, want 0", w.pathRuns)
	}
	if len(w.calls) != 1 || !w.calls[0].fill {
		t.Fatalf("shape calls = %#v, want one filled circle", w.calls)
	}
	if len(w.appendPaths) != 0 {
		t.Fatalf("appendPaths = %d, want 0", len(w.appendPaths))
	}
}

func TestStdCircle_DrawContent_UsesContentBoxCenterAndRadius(t *testing.T) {
	circle := &StdCircle{}
	circle.SetLeft(10)
	circle.SetTop(20)
	circle.SetWidth(100)
	circle.SetHeight(80)
	circle.padding.SetAll("10", "pt")
	circle.border = &PenStyle{id: "solid", width: 1}
	circle.fill = &BrushStyle{id: "fill", color: NamedColor("gold")}
	w := &shapeTestWriter{}

	if err := circle.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(w.calls))
	}
	call := w.calls[0]
	if call.name != "circle" {
		t.Fatalf("shape = %q, want circle", call.name)
	}
	if call.x != 60 || call.y != 60 {
		t.Fatalf("center = (%v,%v), want (60,60)", call.x, call.y)
	}
	if call.a != 40 {
		t.Fatalf("radius = %v, want 40", call.a)
	}
	if !call.border || !call.fill {
		t.Fatalf("border/fill = %v/%v, want true/true", call.border, call.fill)
	}
}

func TestStdCircle_DrawContent_RejectsSweepGradientFill(t *testing.T) {
	circle := &StdCircle{}
	circle.SetWidth(40)
	circle.SetHeight(40)
	circle.fill = &BrushStyle{
		kind:          BrushKindSweepGradient,
		sweepGradient: &sweepGradientStyle{},
	}

	if err := circle.DrawContent(&shapeTestWriter{}); err == nil {
		t.Fatal("expected sector-only sweep-gradient error")
	}
}

func TestStdCircle_DrawContent_BorderNoneDisablesShapeOutline(t *testing.T) {
	circle := &StdCircle{}
	circle.SetWidth(40)
	circle.SetHeight(40)
	circle.SetAttrs(map[string]string{"border": "none", "fill": "Gold"})
	w := &shapeTestWriter{}

	if err := circle.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.calls) != 1 || w.calls[0].border || !w.calls[0].fill {
		t.Fatalf("shape calls = %#v, want fill without border", w.calls)
	}
}

func TestStdCircle_RadiusMatchesExplicitDiameterInHBox(t *testing.T) {
	hbox := &StdContainer{}
	hbox.SetWidth(200)
	hbox.SetHeight(60)
	hbox.SetAttrs(map[string]string{"layout": "hbox"})

	fromRadius := &StdCircle{}
	fromRadius.SetAttrs(map[string]string{"r": "20pt"})
	hbox.AddChild(fromRadius)
	if err := fromRadius.SetContainer(hbox); err != nil {
		t.Fatal(err)
	}

	fromDiameter := &StdCircle{}
	fromDiameter.SetAttrs(map[string]string{"width": "40pt", "height": "40pt"})
	hbox.AddChild(fromDiameter)
	if err := fromDiameter.SetContainer(hbox); err != nil {
		t.Fatal(err)
	}

	hbox.LayoutWidget(&shapeTestWriter{})

	if fromRadius.Width() != fromDiameter.Width() || fromRadius.Height() != fromDiameter.Height() {
		t.Fatalf("radius circle size = (%v, %v), diameter circle size = (%v, %v)",
			fromRadius.Width(), fromRadius.Height(), fromDiameter.Width(), fromDiameter.Height())
	}
	if fromRadius.Width() != 40 || fromRadius.Height() != 40 {
		t.Fatalf("radius circle size = (%v, %v), want (40, 40)", fromRadius.Width(), fromRadius.Height())
	}
	if !fromRadius.WidthAspectInferred() || !fromRadius.HeightAspectInferred() {
		t.Fatal("radius circle dimensions were not marked as intrinsically inferred")
	}
}

func TestStdCircle_DrawContent_AppliesGradientBorderInShapeBox(t *testing.T) {
	circle := &StdCircle{}
	circle.SetLeft(10)
	circle.SetTop(20)
	circle.SetWidth(100)
	circle.SetHeight(80)
	circle.border = &PenStyle{
		kind: PenKindLinearGradient,
		linearGradient: &pdf.LinearGradient{
			Stops: []pdf.GradientStop{
				{Position: 0, Color: NamedColor("Tomato")},
				{Position: 1, Color: NamedColor("SteelBlue")},
			},
		},
		linearPct: &linearGradientPct{X0: float64Ptr(0), Y0: float64Ptr(50), X1: float64Ptr(100), Y1: float64Ptr(50)},
	}
	w := &shapeTestWriter{}

	if err := circle.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.lineLinear) != 1 {
		t.Fatalf("line linear gradient count = %d, want 1", len(w.lineLinear))
	}
	got := w.lineLinear[0]
	if got.X0 != 10 || got.Y0 != 60 || got.X1 != 110 || got.Y1 != 60 {
		t.Fatalf("gradient coords = %#v, want shape-box coords", got)
	}
}

func TestStdPolygon_DrawContent_UsesAttrs(t *testing.T) {
	polygon := &StdPolygon{}
	polygon.SetLeft(0)
	polygon.SetTop(0)
	polygon.SetWidth(120)
	polygon.SetHeight(120)
	polygon.SetAttrs(map[string]string{
		"sides":    "6",
		"rotation": "30",
		"reverse":  "true",
	})
	polygon.border = &PenStyle{id: "solid", width: 1}
	w := &shapeTestWriter{}

	if err := polygon.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	call := w.calls[0]
	if call.i != 6 {
		t.Fatalf("sides = %d, want 6", call.i)
	}
	if call.rotation != 30 {
		t.Fatalf("rotation = %v, want 30", call.rotation)
	}
	if !call.reverse {
		t.Fatal("reverse = false, want true")
	}
	if call.a != 60 {
		t.Fatalf("radius = %v, want 60", call.a)
	}
}

func TestStdStar_DefaultsInnerRadiusAndPoints(t *testing.T) {
	star := &StdStar{}
	star.SetLeft(0)
	star.SetTop(0)
	star.SetWidth(100)
	star.SetHeight(100)
	star.border = &PenStyle{id: "solid", width: 1}
	w := &shapeTestWriter{}

	if err := star.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	call := w.calls[0]
	if call.i != 5 {
		t.Fatalf("points = %d, want 5", call.i)
	}
	if call.a != 50 || call.b != 25 {
		t.Fatalf("r1/r2 = %v/%v, want 50/25", call.a, call.b)
	}
}

func TestStdStar_AllowsFourPoints(t *testing.T) {
	star := &StdStar{}
	star.SetLeft(0)
	star.SetTop(0)
	star.SetWidth(100)
	star.SetHeight(100)
	star.points = 4
	star.border = &PenStyle{id: "solid", width: 1}
	w := &shapeTestWriter{}

	if err := star.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	call := w.calls[0]
	if call.i != 4 {
		t.Fatalf("points = %d, want 4", call.i)
	}
}

func TestStdStar_AllowsTwoPoints(t *testing.T) {
	star := &StdStar{}
	star.SetLeft(0)
	star.SetTop(0)
	star.SetWidth(100)
	star.SetHeight(100)
	star.points = 2
	star.border = &PenStyle{id: "solid", width: 1}
	w := &shapeTestWriter{}

	if err := star.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	call := w.calls[0]
	if call.i != 2 {
		t.Fatalf("points = %d, want 2", call.i)
	}
}

func TestStdStar_AllowsThreePoints(t *testing.T) {
	star := &StdStar{}
	star.SetLeft(0)
	star.SetTop(0)
	star.SetWidth(100)
	star.SetHeight(100)
	star.points = 3
	star.border = &PenStyle{id: "solid", width: 1}
	w := &shapeTestWriter{}

	if err := star.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	call := w.calls[0]
	if call.i != 3 {
		t.Fatalf("points = %d, want 3", call.i)
	}
}

func TestStdArcAndArch_DrawContent(t *testing.T) {
	arc := &StdArc{}
	arc.SetLeft(0)
	arc.SetTop(0)
	arc.SetWidth(80)
	arc.SetHeight(80)
	arc.SetAttrs(map[string]string{"start-angle": "45", "end-angle": "180"})

	arch := &StdArch{}
	arch.SetLeft(0)
	arch.SetTop(0)
	arch.SetWidth(120)
	arch.SetHeight(120)
	arch.SetAttrs(map[string]string{"r2": "20", "start-angle": "0", "end-angle": "270"})
	arch.border = &PenStyle{id: "solid", width: 1}
	w := &shapeTestWriter{}

	if err := arc.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if err := arch.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(w.calls))
	}
	if w.calls[0].name != "arc" || w.calls[0].a != 40 || w.calls[0].b != 45 || w.calls[0].c != 180 || !w.calls[0].move {
		t.Fatalf("arc call = %#v", w.calls[0])
	}
	if w.pathRuns != 1 || w.strokes != 1 {
		t.Fatalf("pathRuns/strokes = %d/%d, want 1/1", w.pathRuns, w.strokes)
	}
	if w.calls[1].name != "arch" || w.calls[1].a != 60 || w.calls[1].b != 20 || w.calls[1].c != 0 || w.calls[1].d != 270 {
		t.Fatalf("arch call = %#v", w.calls[1])
	}
}

func TestParse_ShapeTags(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <circle width="1in" height="1in" border="solid" />
    <ellipse width="2in" height="1in" rx="0.75in" />
    <polygon width="1in" height="1in" sides="5" rotation="36" />
    <star width="1in" height="1in" points="7" r2="0.2in" />
    <arc width="1in" height="1in" start-angle="0" end-angle="180" />
    <pie width="1in" height="1in" start-angle="45" end-angle="135" />
    <arch width="1in" height="1in" r2="0.2in" start-angle="90" end-angle="270" />
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	if len(page.children) != 7 {
		t.Fatalf("child count = %d, want 7", len(page.children))
	}
	if _, ok := page.children[0].(*StdCircle); !ok {
		t.Fatalf("child 0 type = %T, want *StdCircle", page.children[0])
	}
	if _, ok := page.children[1].(*StdEllipse); !ok {
		t.Fatalf("child 1 type = %T, want *StdEllipse", page.children[1])
	}
	if polygon, ok := page.children[2].(*StdPolygon); !ok || polygon.Sides() != 5 || polygon.rotation != 36 {
		t.Fatalf("polygon = %#v", page.children[2])
	}
	if star, ok := page.children[3].(*StdStar); !ok || star.Points() != 7 || star.r2 == 0 {
		t.Fatalf("star = %#v", page.children[3])
	}
	if arc, ok := page.children[4].(*StdArc); !ok || arc.startAngle != 0 || arc.endAngle != 180 {
		t.Fatalf("arc = %#v", page.children[4])
	}
	if pie, ok := page.children[5].(*StdPie); !ok || pie.startAngle != 45 || pie.endAngle != 135 {
		t.Fatalf("pie = %#v", page.children[5])
	}
	if arch, ok := page.children[6].(*StdArch); !ok || arch.r2 == 0 || arch.startAngle != 90 || arch.endAngle != 270 {
		t.Fatalf("arch = %#v", page.children[6])
	}
}

func TestStdCircle_LayoutAbsolute_CenterPlacement(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 400, height: 300}}
	page.SetScope(&defaultScope)
	page.SetAttrs(map[string]string{"layout": "absolute"})

	circle := &StdCircle{}
	circle.SetScope(&defaultScope)
	if err := circle.SetContainer(page); err != nil {
		t.Fatal(err)
	}
	circle.SetAttrs(map[string]string{
		"position":  "absolute",
		"center-x":  "50%",
		"center-y":  "25%",
		"r":         "40pt",
	})
	page.AddChild(circle)

	if err := LayoutAbsolute(page, page.LayoutStyle(), &labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}

	cx, cy := circle.center()
	if !floatEquals(cx, 200) || !floatEquals(cy, 75) {
		t.Fatalf("center = (%v,%v), want (200,75)", cx, cy)
	}
	if !floatEquals(circle.Left(), 160) || !floatEquals(circle.Top(), 35) {
		t.Fatalf("box origin = (%v,%v), want (160,35)", circle.Left(), circle.Top())
	}
}

func TestStdCircle_DrawContent_CompositeTripleRingUsesNestedPaths(t *testing.T) {
	outer := &StdCircle{}
	outer.SetLeft(0)
	outer.SetTop(0)
	outer.SetWidth(100)
	outer.SetHeight(100)
	outer.fill = &BrushStyle{id: "fill", color: NamedColor("gold")}

	middle := &StdCircle{}
	middle.SetAttrs(map[string]string{"r": "30pt", "reverse": "true", "fill": "SkyBlue"})

	inner := &StdCircle{}
	inner.SetAttrs(map[string]string{"r": "10pt", "reverse": "true"})

	outer.AddChild(middle)
	middle.AddChild(inner)
	if err := middle.SetContainer(outer); err != nil {
		t.Fatal(err)
	}
	if err := inner.SetContainer(middle); err != nil {
		t.Fatal(err)
	}

	w := &shapeTestWriter{}
	w.t = t
	if err := outer.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if w.pathRuns != 2 {
		t.Fatalf("pathRuns = %d, want 2", w.pathRuns)
	}
	if w.fills != 2 {
		t.Fatalf("fills = %d, want 2", w.fills)
	}
	if len(w.appendPaths) != 4 {
		t.Fatalf("appendPaths = %d, want 4", len(w.appendPaths))
	}
	if !w.appendPaths[1].Reverse {
		t.Fatal("middle append path in outer composite should be reversed")
	}
	if !w.appendPaths[3].Reverse {
		t.Fatal("inner append path in middle composite should be reversed")
	}
	if len(w.calls) != 0 {
		t.Fatalf("immediate shape calls = %#v, want none", w.calls)
	}
}
