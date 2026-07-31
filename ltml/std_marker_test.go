package ltml

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/rowland/leadtype/pdf"
)

type markerTestWriter struct {
	lineTestWriter
	polygons int
	arcs     []markerArcCall
}

type markerArcCall struct {
	x, y, radius, start, end float64
}

func (w *markerTestWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	w.polygons++
	return nil
}

func (w *markerTestWriter) Arc(x, y, r, start, end float64, moveToStart bool) error {
	w.arcs = append(w.arcs, markerArcCall{x: x, y: y, radius: r, start: start, end: end})
	return nil
}

func TestParseMarkerRegistersScopedSingleComponent(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <marker id="triangle" width="4" height="3" ref-x="75%" ref-y="50%">
    <polygon sides="3" rotation="30" />
  </marker>
  <page />
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	marker, ok := doc.Root().MarkerFor("triangle")
	if !ok {
		t.Fatal("marker triangle was not registered")
	}
	resolved := marker.resolved(2)
	if resolved.width != 8 || resolved.height != 6 || resolved.refX != 6 || resolved.refY != 3 {
		t.Fatalf("resolved marker = %#v, want 8x6 ref=(6,3)", resolved)
	}
}

func TestMarkerUserSpaceSizingAndScopedBuiltinShadowing(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <marker id="arrow" width="6pt" height="4pt"
          marker-units="user-space" ref-x="100%" ref-y="50%">
    <circle />
  </marker>
  <page />
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	marker, ok := doc.Root().MarkerFor("arrow")
	if !ok || marker.builtin != "" {
		t.Fatalf("MarkerFor(arrow) = %#v, want scoped custom marker", marker)
	}
	resolved := marker.resolved(20)
	if resolved.width != 6 || resolved.height != 4 || resolved.refX != 6 || resolved.refY != 2 {
		t.Fatalf("user-space marker = %#v, want 6x4 ref=(6,2)", resolved)
	}
}

func TestParseMarkerRejectsInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{"missing child", `<marker id="m" width="3" height="3" />`, "exactly one"},
		{"multiple children", `<marker id="m" width="3" height="3"><circle/><star/></marker>`, "exactly one"},
		{"unsupported child", `<marker id="m" width="3" height="3"><label>x</label></marker>`, "does not support"},
		{"positioned child", `<marker id="m" width="3" height="3"><circle width="2"/></marker>`, `cannot set "width"`},
		{"invalid units", `<marker id="m" width="3" height="3" marker-units="em"><circle/></marker>`, "invalid marker-units"},
		{"physical stroke unit", `<marker id="m" width="3pt" height="3"><circle/></marker>`, "invalid stroke-width marker measurement"},
		{"duplicate", `<marker id="m" width="3" height="3"><circle/></marker><marker id="m" width="3" height="3"><circle/></marker>`, "duplicate marker"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse([]byte(`<ltml>` + tc.body + `</ltml>`))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Parse() error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestStdLineMarkersUsePenDefaultsAndLineOverride(t *testing.T) {
	line := &StdLine{}
	line.SetLeft(10)
	line.SetTop(20)
	line.SetWidth(100)
	line.SetHeight(20)
	line.SetScope(&defaultScope)
	line.style = &PenStyle{
		id:          "axis",
		color:       NamedColor("gray"),
		width:       2,
		pattern:     "solid",
		markerStart: "arrow",
		markerEnd:   "arrow",
	}
	line.markerStart = "none"
	w := &markerTestWriter{}

	if err := line.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.lines) != 1 {
		t.Fatalf("shaft count = %d, want 1", len(w.lines))
	}
	if got := w.lines[0]; math.Abs(got.length-97) > 0.001 {
		t.Fatalf("shaft = %#v, want length 97 after end marker inset", got)
	}
	if len(w.rotations) != 1 {
		t.Fatalf("marker rotations = %v, want one end marker", w.rotations)
	}
	if got := w.rotations[0]; got.angle != 0 || math.Abs(got.x-107) > 0.001 || got.y != 30 {
		t.Fatalf("end marker rotation = %#v, want angle=0 at (107,30)", got)
	}
}

func TestStdLineEmptyMarkerInheritsAndNoneRemovesPenMarker(t *testing.T) {
	style := &PenStyle{markerStart: "arrow", markerEnd: "arrow"}
	line := &StdLine{}

	line.SetAttrs(map[string]string{"marker-start": "  "})
	start, end := line.markerIDs(style)
	if start != "arrow" || end != "arrow" {
		t.Fatalf("empty override resolved markers = %q/%q, want inherited arrow/arrow", start, end)
	}

	line.SetAttrs(map[string]string{"marker-start": "none", "marker-end": "none"})
	start, end = line.markerIDs(style)
	if start != "none" || end != "none" {
		t.Fatalf("none override resolved markers = %q/%q, want none/none", start, end)
	}

	line.SetAttrs(map[string]string{"marker-start": "", "marker-end": ""})
	start, end = line.markerIDs(style)
	if start != "arrow" || end != "arrow" {
		t.Fatalf("cleared overrides resolved markers = %q/%q, want inherited arrow/arrow", start, end)
	}
}

func TestStdLineCustomMarkerCutbackAndStrokeWidthScaling(t *testing.T) {
	scope := &Scope{parent: &defaultScope}
	marker := &StdMarker{
		markerWidth:  4,
		markerHeight: 2,
		refX:         markerCoordinate{value: 100, percent: true},
		refY:         markerCoordinate{value: 50, percent: true},
		stemCutback:  0.5,
	}
	polygon := &StdPolygon{sides: 3}
	marker.AddChild(polygon)
	marker.ID = "custom"
	if err := scope.AddMarker(marker); err != nil {
		t.Fatal(err)
	}
	line := &StdLine{markerEnd: "custom"}
	line.SetLeft(0)
	line.SetTop(0)
	line.SetWidth(100)
	line.SetHeight(10)
	line.SetScope(scope)
	line.style = &PenStyle{color: NamedColor("red"), width: 3, pattern: "solid"}
	w := &markerTestWriter{}

	if err := line.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.lines) != 1 || math.Abs(w.lines[0].length-98.5) > 0.001 {
		t.Fatalf("shaft calls = %#v, want length 98.5", w.lines)
	}
	if len(w.rotations) != 1 || w.polygons != 1 {
		t.Fatalf("rotations/polygons = %d/%d, want 1/1", len(w.rotations), w.polygons)
	}
}

func TestStdLineMissingMarkerReturnsError(t *testing.T) {
	line := &StdLine{markerEnd: "not-defined"}
	line.SetWidth(100)
	line.SetScope(&defaultScope)

	err := line.DrawContent(&markerTestWriter{})
	if err == nil || !strings.Contains(err.Error(), `missing marker "not-defined"`) {
		t.Fatalf("DrawContent() error = %v, want missing marker", err)
	}
}

func TestFitResolvedMarkersKeepsViewportsInsideShortLine(t *testing.T) {
	definition := &StdMarker{}
	start := &resolvedMarker{definition: definition, width: 6, height: 6, refX: 3, refY: 3, cutback: 1}
	end := &resolvedMarker{definition: definition, width: 6, height: 6, refX: 3, refY: 3, cutback: 1}

	fittedStart, fittedEnd := fitResolvedMarkers(start, end, 6)
	if fittedStart.width != 3 || fittedEnd.width != 3 {
		t.Fatalf("fitted widths = %v/%v, want 3/3", fittedStart.width, fittedEnd.width)
	}
	if fittedStart.height != 3 || fittedStart.refX != 1.5 || fittedStart.cutback != 0.5 {
		t.Fatalf("fitted start = %#v, want uniformly scaled geometry", fittedStart)
	}
	if start.width != 6 || end.width != 6 {
		t.Fatal("fitResolvedMarkers mutated source markers")
	}
}

func TestSectorAutoLinesShareRemainingCurvedArc(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <pen id="axis" width="2pt" color="#c8c8c8" />
  <page units="pt" layout="absolute">
    <div left="0" top="0" width="400" height="400"
         layout="radial-out" rows="1" cols="2" r0="140">
      <sector layout.hpadding="6">
        <line width="auto" style="axis" marker-end="arrow" />
        <label>TRANSLATED LABEL</label>
        <line width="auto" style="axis" marker-start="arrow" />
      </sector>
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	w := &markerTestWriter{}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	page := doc.Root().Page(0)
	radial := page.Widgets()[0].(*StdContainer)
	sector := radial.Widgets()[0].(*StdSector)
	children := sector.Widgets()
	left := children[0].(*StdLine)
	label := children[1].(*StdLabel)
	right := children[2].(*StdLine)
	leftSlot := sector.flowSlots[left]
	labelSlot := sector.flowSlots[label]
	rightSlot := sector.flowSlots[right]
	leftWidth := leftSlot.MaxX - leftSlot.MinX
	rightWidth := rightSlot.MaxX - rightSlot.MinX
	if leftWidth <= 0 || math.Abs(leftWidth-rightWidth) > 0.001 {
		t.Fatalf("auto line slots = %#v and %#v, want equal positive widths", leftSlot, rightSlot)
	}
	centerY := func(slot radialBounds) float64 { return (slot.MinY + slot.MaxY) / 2 }
	if math.Abs(centerY(leftSlot)-centerY(labelSlot)) > 0.001 ||
		math.Abs(centerY(rightSlot)-centerY(labelSlot)) > 0.001 {
		t.Fatalf("curved row centers = %g, %g, %g; want a shared radius",
			centerY(leftSlot), centerY(labelSlot), centerY(rightSlot))
	}
	if len(w.arcs) != 2 {
		t.Fatalf("arc calls = %d, want 2", len(w.arcs))
	}
	if len(w.rotations) != 2 {
		t.Fatalf("marker rotations = %d, want 2", len(w.rotations))
	}
}

func TestRightAxisCurvedLabelKeepsBaselineTowardCenter(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <pen id="axis" width="2pt" color="#c8c8c8" />
  <page units="pt" layout="absolute">
    <div center-x="200" center-y="200" layout="radial"
         rows="1" cols="2" row-angle-offsets="90" r0="140" r1="150">
      <sector>
        <line width="auto" style="axis" />
        <label>LEFT</label>
        <line width="auto" style="axis" />
      </sector>
      <sector>
        <line width="auto" style="axis" />
        <label>RIGHT</label>
        <line width="auto" style="axis" />
      </sector>
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	w := &markerTestWriter{}
	w.t = t
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if len(w.curvedOpts) != 2 {
		t.Fatalf("curved label draws = %d, want 2", len(w.curvedOpts))
	}
	right := w.curvedOpts[1]
	if right.Direction != pdf.CurvedTextClockwise || right.Facing != pdf.CurvedTextFacingUpright {
		t.Fatalf("right label orientation = %v/%v, want clockwise/upright",
			right.Direction, right.Facing)
	}
}

func TestHBoxAutoLinesShareRemainingWidthAroundLabel(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <pen id="axis" width="2pt" color="#c8c8c8" />
  <page units="pt" layout="absolute">
    <div left="0" top="0" width="300" layout="hbox" layout.hpadding="6">
      <line width="auto" style="axis" marker-start="arrow" />
      <label>HELLO</label>
      <line width="auto" style="axis" marker-end="arrow" />
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	w := &markerTestWriter{}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	box := doc.Root().Page(0).Widgets()[0].(*StdContainer)
	left := box.Widgets()[0].(*StdLine)
	right := box.Widgets()[2].(*StdLine)
	if left.Width() <= 0 || math.Abs(left.Width()-right.Width()) > 0.001 {
		t.Fatalf("auto line widths = %v and %v, want equal positive widths", left.Width(), right.Width())
	}
	if len(w.lines) != 2 || len(w.rotations) != 2 {
		t.Fatalf("shafts/markers = %d/%d, want 2/2", len(w.lines), len(w.rotations))
	}
}

func TestVBoxAutoHeightVerticalMarkedLineSharesRemainingHeight(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <pen id="axis" width="2pt" color="#c8c8c8" />
  <page units="pt" layout="absolute">
    <vbox left="0" top="0" width="100" height="72">
      <label>TOP</label>
      <line angle="90" width="100%" height="auto" style="axis"
            marker-start="arrow" marker-end="arrow" />
      <label>BOTTOM</label>
    </vbox>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	w := &markerTestWriter{}
	w.t = t
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	page := doc.Root().Page(0)
	box := page.Widgets()[0].(*StdContainer)
	children := box.Widgets()
	top := children[0].(*StdLabel)
	line := children[1].(*StdLine)
	bottom := children[2].(*StdLabel)
	if !top.Visible() || !line.Visible() || !bottom.Visible() {
		t.Fatalf("vbox visibility = %v/%v/%v, want all children visible",
			top.Visible(), line.Visible(), bottom.Visible())
	}
	if line.Height() <= 0 || line.Top() < top.Bottom()-layoutFitEpsilon || bottom.Top() < line.Bottom()-layoutFitEpsilon {
		t.Fatalf("vbox placements: top=%g..%g line=%g..%g bottom=%g..%g",
			top.Top(), top.Bottom(), line.Top(), line.Bottom(), bottom.Top(), bottom.Bottom())
	}
	if len(w.lines) != 1 || len(w.rotations) != 2 {
		t.Fatalf("shafts/markers = %d/%d, want 1/2", len(w.lines), len(w.rotations))
	}
}

func TestSectorMarkedLinesRespectSweepAndRTL(t *testing.T) {
	tests := []struct {
		name  string
		sweep string
		dir   string
	}{
		{name: "clockwise", sweep: `sweep="cw"`},
		{name: "rtl", dir: `dir="rtl"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := fmt.Sprintf(`
<ltml>
  <pen id="axis" width="2pt" color="#c8c8c8" />
  <page units="pt" layout="absolute">
    <div left="0" top="0" width="400" height="400"
         layout="radial-out" rows="1" cols="4" %s r0="140">
      <sector colspan="2" %s layout.hpadding="6">
        <line width="auto" style="axis" marker-end="arrow" />
        <label>ORIENTATION</label>
        <line width="auto" style="axis" marker-start="arrow" />
      </sector>
    </div>
  </page>
</ltml>`, tc.sweep, tc.dir)
			doc, err := Parse([]byte(input))
			if err != nil {
				t.Fatal(err)
			}
			w := &markerTestWriter{}
			if err := doc.Print(w); err != nil {
				t.Fatal(err)
			}
			if len(w.arcs) != 2 || len(w.rotations) != 2 {
				t.Fatalf("arcs/markers = %d/%d, want 2/2", len(w.arcs), len(w.rotations))
			}
			for _, rotation := range w.rotations {
				if math.IsNaN(rotation.angle) || math.IsInf(rotation.angle, 0) {
					t.Fatalf("invalid tangent rotation %#v", rotation)
				}
			}
		})
	}
}

func TestSectorMarkedLinesPreserveSweepAcrossAngleSeam(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <pen id="axis" width="2pt" color="#c8c8c8" />
  <page units="pt" layout="absolute">
    <div left="0" top="0" width="400" height="400"
         layout="radial-out" rows="1" angles="0,180"
         row-angle-offsets="90" r0="140">
      <sector layout.hpadding="6">
        <line width="auto" style="axis" marker-start="arrow" />
        <label>LEFT</label>
        <line width="auto" style="axis" marker-end="arrow" />
      </sector>
      <sector layout.hpadding="6">
        <line width="auto" style="axis" marker-start="arrow" />
        <label>RIGHT</label>
        <line width="auto" style="axis" marker-end="arrow" />
      </sector>
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	w := &markerTestWriter{}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	if len(w.arcs) != 4 {
		t.Fatalf("arc calls = %d, want 4", len(w.arcs))
	}
	for i, arc := range w.arcs {
		span := math.Abs(arc.end - arc.start)
		if span <= 0 || span >= 90 {
			t.Fatalf("arc %d span = %g (%g to %g), want the allocated sub-quadrant rather than the complementary sweep",
				i, span, arc.start, arc.end)
		}
	}
}
