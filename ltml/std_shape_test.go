package ltml

import "testing"

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
	calls    []shapeCall
	inPath   bool
	strokes  int
	pathRuns int
}

func (w *shapeTestWriter) Circle(x, y, r float64, border, fill, reverse bool) error {
	w.calls = append(w.calls, shapeCall{name: "circle", x: x, y: y, a: r, border: border, fill: fill, reverse: reverse})
	return nil
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

func TestStdArcAndArch_DrawContent(t *testing.T) {
	arc := &StdArc{}
	arc.SetLeft(0)
	arc.SetTop(0)
	arc.SetWidth(80)
	arc.SetHeight(80)
	arc.SetAttrs(map[string]string{"start_angle": "45", "end_angle": "180"})

	arch := &StdArch{}
	arch.SetLeft(0)
	arch.SetTop(0)
	arch.SetWidth(120)
	arch.SetHeight(120)
	arch.SetAttrs(map[string]string{"r2": "20", "start_angle": "0", "end_angle": "270"})
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
    <arc width="1in" height="1in" start_angle="0" end_angle="180" />
    <pie width="1in" height="1in" start_angle="45" end_angle="135" />
    <arch width="1in" height="1in" r2="0.2in" start_angle="90" end_angle="270" />
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.ltmls[0].Page(0)
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
