package ltml

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/rowland/leadtype/pdf"
)

func testSectorFont() *FontStyle {
	return &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
}

func addTestSectorLabel(t *testing.T, sector *StdSector, text string, attrs map[string]string) *StdLabel {
	t.Helper()
	label := &StdLabel{}
	label.font = testSectorFont()
	label.SetAttrs(attrs)
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.AddText(text)
	sector.AddChild(label)
	return label
}

func TestStdSectorResolvedBordersFollowSweepDirection(t *testing.T) {
	outer := &PenStyle{id: "outer"}
	inner := &PenStyle{id: "inner"}
	left := &PenStyle{id: "left"}
	right := &PenStyle{id: "right"}
	sector := &StdSector{}
	sector.borders[topSide] = outer
	sector.borders[bottomSide] = inner
	sector.borders[leftSide] = left
	sector.borders[rightSide] = right
	parent := &StdContainer{}
	sector.container = parent

	gotOuter, gotInner, gotStart, gotEnd := sector.resolvedSectorBorders()
	if gotOuter != outer || gotInner != inner || gotStart != left || gotEnd != right {
		t.Fatalf("ccw borders = %#v/%#v/%#v/%#v, want outer/inner/left/right", gotOuter, gotInner, gotStart, gotEnd)
	}

	parent.radialSweep = radialSweepCW
	gotOuter, gotInner, gotStart, gotEnd = sector.resolvedSectorBorders()
	if gotOuter != outer || gotInner != inner || gotStart != right || gotEnd != left {
		t.Fatalf("cw borders = %#v/%#v/%#v/%#v, want outer/inner/right/left", gotOuter, gotInner, gotStart, gotEnd)
	}
}

func TestStdSectorBorderAliasesOverrideLogicalSides(t *testing.T) {
	sector := &StdSector{}
	sector.SetAttrs(map[string]string{
		"border-top":   "Red",
		"border-left":  "Blue",
		"border-outer": "Gold",
		"border-start": "Green",
	})

	outer, _, start, _ := sector.resolvedSectorBorders()
	if outer == nil || outer.color != NamedColor("Gold") {
		t.Fatalf("outer border = %#v, want Gold alias", outer)
	}
	if start == nil || start.color != NamedColor("Green") {
		t.Fatalf("start border = %#v, want Green alias", start)
	}
}

func TestStdSectorDrawBorderStrokesIndividualArcsWithoutRadialSeam(t *testing.T) {
	sector := &StdSector{}
	sector.geometry = radialSectorGeometry{
		CenterX: 20, CenterY: 30,
		InnerRadius: 5, OuterRadius: 10,
		StartAngle: 22.5, EndAngle: 382.5,
	}
	sector.borders[topSide] = &PenStyle{color: NamedColor("Red"), width: 1}
	sector.borders[bottomSide] = &PenStyle{color: NamedColor("Blue"), width: 2}
	writer := &shapeTestWriter{labelTestWriter: labelTestWriter{t: t}}

	if err := sector.DrawBorder(writer); err != nil {
		t.Fatal(err)
	}
	if writer.pathRuns != 2 || writer.strokes != 2 {
		t.Fatalf("paths/strokes = %d/%d, want 2/2", writer.pathRuns, writer.strokes)
	}
	if len(writer.calls) != 2 || writer.calls[0].name != "arc" || writer.calls[1].name != "arc" {
		t.Fatalf("calls = %#v, want two arcs", writer.calls)
	}
	if writer.calls[0].a != 10 || writer.calls[1].a != 5 {
		t.Fatalf("arc radii = %v/%v, want 10/5", writer.calls[0].a, writer.calls[1].a)
	}
}

func TestStdSectorDrawBorderStrokesFourIndividuallyStyledEdges(t *testing.T) {
	sector := &StdSector{}
	sector.geometry = radialSectorGeometry{
		CenterX: 20, CenterY: 30,
		InnerRadius: 5, OuterRadius: 10,
		StartAngle: 0, EndAngle: 90,
	}
	for i := range sector.borders {
		sector.borders[i] = &PenStyle{color: NamedColor(sideNames[i]), width: float64(i + 1)}
	}
	writer := &shapeTestWriter{labelTestWriter: labelTestWriter{t: t}}

	if err := sector.DrawBorder(writer); err != nil {
		t.Fatal(err)
	}
	if writer.pathRuns != 4 || writer.strokes != 4 {
		t.Fatalf("paths/strokes = %d/%d, want 4/4", writer.pathRuns, writer.strokes)
	}
	if len(writer.calls) != 2 {
		t.Fatalf("arc calls = %d, want 2", len(writer.calls))
	}
	if len(writer.moves) != 2 {
		t.Fatalf("radial move count = %d, want 2", len(writer.moves))
	}
}

func TestStdSectorDrawBorderUsesClosedShapeForAggregateBorder(t *testing.T) {
	sector := &StdSector{}
	sector.geometry = radialSectorGeometry{
		CenterX: 20, CenterY: 30,
		InnerRadius: 5, OuterRadius: 10,
		StartAngle: 0, EndAngle: 90,
	}
	sector.border = &PenStyle{color: NamedColor("Red"), width: 1}
	writer := &shapeTestWriter{labelTestWriter: labelTestWriter{t: t}}

	if err := sector.DrawBorder(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.calls) != 1 || writer.calls[0].name != "arch" {
		t.Fatalf("calls = %#v, want one aggregate arch", writer.calls)
	}
	if writer.pathRuns != 0 {
		t.Fatalf("individual path runs = %d, want 0", writer.pathRuns)
	}
}

func TestParse_RadialWrapsDirectChildInSector(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <div layout="radial" cols="2">
      <label id="wrapped">Hello</label>
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	radial := page.children[0].(*StdContainer)
	if radial.LayoutStyle().manager != "radial" {
		t.Fatalf("layout manager = %q, want radial", radial.LayoutStyle().manager)
	}
	if len(radial.children) != 1 {
		t.Fatalf("radial child count = %d, want 1", len(radial.children))
	}
	sector, ok := radial.children[0].(*StdSector)
	if !ok {
		t.Fatalf("wrapped child type = %T, want *StdSector", radial.children[0])
	}
	if len(sector.children) != 1 {
		t.Fatalf("sector child count = %d, want 1", len(sector.children))
	}
	label, ok := sector.children[0].(*StdLabel)
	if !ok {
		t.Fatalf("inner child type = %T, want *StdLabel", sector.children[0])
	}
	if label.Container() != sector {
		t.Fatalf("label container = %T, want *StdSector", label.Container())
	}
	if path := label.Path(); strings.Contains(path, "/sector/") {
		t.Fatalf("wrapped label path = %q, should preserve source path", path)
	}
}

func TestParse_RadialOutWrapsDirectChildInSector(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <div layout="radial-out" cols="2">
      <label id="wrapped">Hello</label>
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	radial := page.children[0].(*StdContainer)
	if radial.LayoutStyle().manager != "radial-out" {
		t.Fatalf("layout manager = %q, want radial-out", radial.LayoutStyle().manager)
	}
	if len(radial.children) != 1 {
		t.Fatalf("radial-out child count = %d, want 1", len(radial.children))
	}
	sector, ok := radial.children[0].(*StdSector)
	if !ok {
		t.Fatalf("wrapped child type = %T, want *StdSector", radial.children[0])
	}
	if len(sector.children) != 1 {
		t.Fatalf("sector child count = %d, want 1", len(sector.children))
	}
	label, ok := sector.children[0].(*StdLabel)
	if !ok {
		t.Fatalf("inner child type = %T, want *StdLabel", sector.children[0])
	}
	if label.Container() != sector {
		t.Fatalf("label container = %T, want *StdSector", label.Container())
	}
	if path := label.Path(); strings.Contains(path, "/sector/") {
		t.Fatalf("wrapped label path = %q, should preserve source path", path)
	}
}

func TestParse_ImplicitSectorOwnsCellAttrsAcrossCascadeLayers(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <style>
      label { border: Red; padding: 2pt; font.size: 9pt; }
      label:first-child { fill: Gold; z-index: 7; }
    </style>
    <div layout="radial" cols="2">
      <label id="source" class="number" units="pt" colspan="2" border="Blue" angle="0" width="30">26</label>
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	sector := doc.Root().Page(0).children[0].(*StdContainer).children[0].(*StdSector)
	label := sector.children[0].(*StdLabel)
	if sector.ColSpan() != 2 || label.ColSpan() != 1 {
		t.Fatalf("colspans sector/label = %d/%d, want 2/1", sector.ColSpan(), label.ColSpan())
	}
	if sector.border == nil || sector.border.color != NamedColor("Blue") || label.border != nil {
		t.Fatalf("borders sector/label = %#v/%#v, want Blue/nil", sector.border, label.border)
	}
	if sector.fill == nil || label.fill != nil || sector.ZIndex() != 7 || label.ZIndex() != 0 {
		t.Fatalf("fill/z sector=%#v/%d label=%#v/%d", sector.fill, sector.ZIndex(), label.fill, label.ZIndex())
	}
	if sector.PaddingTop() != 2 || label.PaddingTop() != 0 {
		t.Fatalf("padding sector/label = %v/%v, want 2/0", sector.PaddingTop(), label.PaddingTop())
	}
	if sector.Units() != "pt" || label.Units() != "pt" {
		t.Fatalf("units sector/label = %q/%q, want pt/pt", sector.Units(), label.Units())
	}
	if label.GetID() != "source" || !slices.Equal(label.Classes, []string{"number"}) || sector.GetID() != "" || len(sector.Classes) != 0 {
		t.Fatalf("identity sector=%q/%v label=%q/%v", sector.GetID(), sector.Classes, label.GetID(), label.Classes)
	}
	if !label.angleSet || label.angle != 0 || label.Width() != 30 || label.Font().size != 9 {
		t.Fatalf("child attrs angle/width/font = %v/%v/%v/%v", label.angleSet, label.angle, label.Width(), label.Font().size)
	}
}

func TestParse_ExplicitSectorAndChildRetainSeparateBorders(t *testing.T) {
	doc, err := Parse([]byte(`<ltml><page><div layout="radial" cols="1">
    <sector border="Red"><label angle="0" border="Blue">Both</label></sector>
  </div></page></ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	sector := doc.Root().Page(0).children[0].(*StdContainer).children[0].(*StdSector)
	label := sector.children[0].(*StdLabel)
	if sector.border == nil || sector.border.color != NamedColor("Red") || label.border == nil || label.border.color != NamedColor("Blue") {
		t.Fatalf("borders sector/label = %#v/%#v", sector.border, label.border)
	}
}

func TestParse_SectorRejectsTextAndInlineContentButAllowsWhitespace(t *testing.T) {
	for _, body := range []string{"Alpha", "<span>Alpha</span>"} {
		_, err := Parse([]byte(`<ltml><page><div layout="radial" cols="1"><sector>` + body + `</sector></div></page></ltml>`))
		if err == nil || !strings.Contains(err.Error(), "ltml/page/div/sector") || !strings.Contains(err.Error(), "<label>") {
			t.Fatalf("body %q error = %v, want path-qualified label guidance", body, err)
		}
	}
	doc, err := Parse([]byte(`<ltml><page><div layout="radial" cols="1"><sector>
      <label>Alpha</label>
    </sector></div></page></ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	sector := doc.Root().Page(0).children[0].(*StdContainer).children[0].(*StdSector)
	if len(sector.children) != 1 {
		t.Fatalf("sector children = %d, want one explicit label", len(sector.children))
	}
}

func TestParse_SectorDoesNotProvideLabelDefaults(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <style>
	      sector:first-child {
	        angle: 23;
        facing: upside-down;
        text-align: right;
        text-valign: bottom;
        origin-x: start;
      }
	      sector > label { text-align: center; }
    </style>
    <div layout="radial" cols="1">
      <sector><label>Alpha</label></sector>
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	sector := doc.Root().Page(0).children[0].(*StdContainer).children[0].(*StdSector)
	label := sector.children[0].(*StdLabel)
	if angle, straight := label.sectorTextAngle(); straight || angle != 0 {
		t.Fatalf("effective label angle = %v/%v, want curved label", straight, angle)
	}
	if got := label.sectorTextFacing(); got != sectorFacingAuto {
		t.Fatalf("effective label facing = %v, want automatic", got)
	}
	if got := label.sectorTextAlign(); got != HAlignCenter {
		t.Fatalf("effective label alignment = %v, want label selector center", got)
	}
	if got := label.sectorTextVAlign(); got != VAlignMiddle {
		t.Fatalf("effective label vertical alignment = %v, want label default middle", got)
	}
	if got := label.OriginX(); got != OriginXCenter {
		t.Fatalf("effective label angular origin = %v, want default center", got)
	}
}

func TestParse_DiscAliasBehavesLikeRadialContainer(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <disc cols="2">
      <label id="wrapped">Hello</label>
    </disc>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.Root().Page(0)
	disc := page.children[0].(*StdContainer)
	if disc.LayoutStyle().manager != "radial" {
		t.Fatalf("layout manager = %q, want radial", disc.LayoutStyle().manager)
	}
	if disc.Path() != "ltml/page/disc" {
		t.Fatalf("disc path = %q, want ltml/page/disc", disc.Path())
	}
	if len(disc.children) != 1 {
		t.Fatalf("disc child count = %d, want 1", len(disc.children))
	}
	if _, ok := disc.children[0].(*StdSector); !ok {
		t.Fatalf("wrapped child type = %T, want *StdSector", disc.children[0])
	}
}

func TestLayoutRadialTable_DerivesInnerTrackFromExtraCells(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{"layout": "radial", "cols": "2"})

	s1 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	s2 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	s3 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	for _, sector := range []*StdSector{s1, s2, s3} {
		sector.font = testSectorFont()
		if err := sector.SetContainer(container); err != nil {
			t.Fatal(err)
		}
		container.AddChild(sector)
	}

	LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t})

	if s1.geometry.StartAngle != 0 || s1.geometry.EndAngle != 180 {
		t.Fatalf("outer sector angles = %v..%v, want 0..180", s1.geometry.StartAngle, s1.geometry.EndAngle)
	}
	if s2.geometry.StartAngle != 180 || s2.geometry.EndAngle != 360 {
		t.Fatalf("second sector angles = %v..%v, want 180..360", s2.geometry.StartAngle, s2.geometry.EndAngle)
	}
	if !(s3.geometry.OuterRadius < s1.geometry.OuterRadius) {
		t.Fatalf("inner sector outer radius = %v, want less than %v", s3.geometry.OuterRadius, s1.geometry.OuterRadius)
	}
}

func TestLayoutRadialTable_UsesExplicitAnglesAndBaseAngle(t *testing.T) {
	container := positionedContainer(0, 0, 240, 240)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout":     "radial",
		"rows":       "1",
		"base-angle": "10",
		"angles":     "0,120,360",
	})

	s1 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	s2 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	for _, sector := range []*StdSector{s1, s2} {
		sector.font = testSectorFont()
		if err := sector.SetContainer(container); err != nil {
			t.Fatal(err)
		}
		container.AddChild(sector)
	}

	LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t})

	if s1.geometry.StartAngle != 10 || s1.geometry.EndAngle != 130 {
		t.Fatalf("sector 1 angles = %v..%v, want 10..130", s1.geometry.StartAngle, s1.geometry.EndAngle)
	}
	if s2.geometry.StartAngle != 130 || s2.geometry.EndAngle != 370 {
		t.Fatalf("sector 2 angles = %v..%v, want 130..370", s2.geometry.StartAngle, s2.geometry.EndAngle)
	}
}

func TestLayoutRadialTable_RowAngleOffsetsStaggerConcentricRows(t *testing.T) {
	container := positionedContainer(0, 0, 240, 240)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout":            "radial",
		"rows":              "3",
		"cols":              "4",
		"base-angle":        "10",
		"row-angle-offsets": "45,0,45",
	})

	outer := []*StdSector{
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
	}
	middle := []*StdSector{
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
	}
	inner := []*StdSector{
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
	}
	for _, sector := range append(append(outer, middle...), inner...) {
		sector.font = testSectorFont()
		if sector == outer[0] || sector == outer[1] || sector == inner[0] || sector == inner[1] {
			sector.SetAttrs(map[string]string{"colspan": "2"})
		}
		if err := sector.SetContainer(container); err != nil {
			t.Fatal(err)
		}
		container.AddChild(sector)
	}

	if err := LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
	if got, want := outer[0].geometry.StartAngle, 55.0; !floatEquals(got, want) {
		t.Fatalf("outer start angle = %v, want %v", got, want)
	}
	if got, want := outer[0].geometry.EndAngle, 235.0; !floatEquals(got, want) {
		t.Fatalf("outer end angle = %v, want %v", got, want)
	}
	if got, want := middle[0].geometry.StartAngle, 10.0; !floatEquals(got, want) {
		t.Fatalf("middle start angle = %v, want %v", got, want)
	}
	if got, want := middle[0].geometry.EndAngle, 100.0; !floatEquals(got, want) {
		t.Fatalf("middle end angle = %v, want %v", got, want)
	}
	if got, want := inner[0].geometry.StartAngle, 55.0; !floatEquals(got, want) {
		t.Fatalf("inner start angle = %v, want %v", got, want)
	}
	if got, want := inner[0].geometry.EndAngle, 235.0; !floatEquals(got, want) {
		t.Fatalf("inner end angle = %v, want %v", got, want)
	}
}

func TestLayoutRadialTable_CWColspanIncludesRowAngleOffset(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout":            "radial",
		"rows":              "1",
		"cols":              "4",
		"sweep":             "cw",
		"row-angle-offsets": "22.5",
	})

	merged := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	merged.font = testSectorFont()
	merged.SetAttrs(map[string]string{"colspan": "2"})
	if err := merged.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(merged)
	for i := 0; i < 2; i++ {
		sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
		sector.font = testSectorFont()
		if err := sector.SetContainer(container); err != nil {
			t.Fatal(err)
		}
		container.AddChild(sector)
	}

	if err := LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
	if got, want := merged.geometry.StartAngle, 22.5; !floatEquals(got, want) {
		t.Fatalf("merged start angle = %v, want %v", got, want)
	}
	if got, want := merged.geometry.EndAngle, -157.5; !floatEquals(got, want) {
		t.Fatalf("merged end angle = %v, want %v", got, want)
	}
}

func TestLayoutRadialTable_RowspanRejectsDifferentRowAngleOffsets(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout":            "radial",
		"rows":              "2",
		"cols":              "1",
		"row-angle-offsets": "0,22.5",
	})
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	sector.SetAttrs(map[string]string{"rowspan": "2"})
	if err := sector.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(sector)

	err := LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t})
	if err == nil || !strings.Contains(err.Error(), "different angle offsets") {
		t.Fatalf("error = %v, want mismatched row-angle-offset error", err)
	}
}

func TestLayoutRadialTable_RowspanAllowsEquivalentRowAngleOffsets(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout":            "radial",
		"rows":              "2",
		"cols":              "1",
		"row-angle-offsets": "0,360",
	})
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	sector.SetAttrs(map[string]string{"rowspan": "2"})
	if err := sector.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(sector)

	if err := LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
}

func TestLayoutRadialTable_RowAngleOffsetsIgnoreUnusedTrailingValues(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout":            "radial",
		"rows":              "1",
		"cols":              "1",
		"row-angle-offsets": "22.5,90",
	})
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	if err := sector.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(sector)

	if err := LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
	if got, want := sector.geometry.StartAngle, 22.5; !floatEquals(got, want) {
		t.Fatalf("start angle = %v, want %v", got, want)
	}
}

func TestLayoutRadialTable_CWSweepUsesClockwiseQuarterSectors(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout": "radial",
		"rows":   "1",
		"cols":   "4",
		"sweep":  "cw",
	})

	sectors := []*StdSector{
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
	}
	for _, sector := range sectors {
		sector.font = testSectorFont()
		if err := sector.SetContainer(container); err != nil {
			t.Fatal(err)
		}
		container.AddChild(sector)
	}

	LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t})

	expectations := []struct {
		start float64
		end   float64
	}{
		{0, -90},
		{-90, -180},
		{-180, -270},
		{-270, -360},
	}
	for i, want := range expectations {
		if got := sectors[i].geometry.StartAngle; !floatEquals(got, want.start) {
			t.Fatalf("sector %d start angle = %v, want %v", i+1, got, want.start)
		}
		if got := sectors[i].geometry.EndAngle; !floatEquals(got, want.end) {
			t.Fatalf("sector %d end angle = %v, want %v", i+1, got, want.end)
		}
	}
}

func TestLayoutRadialTable_CWColspanMergesClockwiseSectors(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout":     "radial",
		"rows":       "1",
		"cols":       "4",
		"base-angle": "180",
		"sweep":      "cw",
	})

	merged := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	merged.font = testSectorFont()
	merged.SetAttrs(map[string]string{"colspan": "2"})
	remaining := []*StdSector{
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
	}
	sectors := append([]*StdSector{merged}, remaining...)
	for _, sector := range sectors {
		sector.font = testSectorFont()
		if err := sector.SetContainer(container); err != nil {
			t.Fatal(err)
		}
		container.AddChild(sector)
	}

	if err := LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
	if got, want := merged.geometry.StartAngle, 180.0; !floatEquals(got, want) {
		t.Fatalf("merged start angle = %v, want %v", got, want)
	}
	if got, want := merged.geometry.EndAngle, 0.0; !floatEquals(got, want) {
		t.Fatalf("merged end angle = %v, want %v", got, want)
	}
}

func TestLayoutRadialTable_CWSweepNormalizesSortsAndDedupesExplicitAngles(t *testing.T) {
	container := positionedContainer(0, 0, 240, 240)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout": "radial",
		"rows":   "1",
		"cols":   "4",
		"sweep":  "cw",
		"angles": "270,0,180,90,360",
	})

	sectors := []*StdSector{
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
	}
	for _, sector := range sectors {
		sector.font = testSectorFont()
		if err := sector.SetContainer(container); err != nil {
			t.Fatal(err)
		}
		container.AddChild(sector)
	}

	LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t})

	expectations := []struct {
		start float64
		end   float64
	}{
		{0, -90},
		{-90, -180},
		{-180, -270},
		{-270, -360},
	}
	for i, want := range expectations {
		if got := sectors[i].geometry.StartAngle; !floatEquals(got, want.start) {
			t.Fatalf("sector %d start angle = %v, want %v", i+1, got, want.start)
		}
		if got := sectors[i].geometry.EndAngle; !floatEquals(got, want.end) {
			t.Fatalf("sector %d end angle = %v, want %v", i+1, got, want.end)
		}
	}
}

func TestLayoutRadialTable_SingleExplicitAngleProducesFullCircleSector(t *testing.T) {
	containerCCW := positionedContainer(0, 0, 200, 200)
	containerCCW.SetScope(&defaultScope)
	containerCCW.SetAttrs(map[string]string{
		"layout": "radial",
		"rows":   "1",
		"angles": "0",
	})

	sectorCCW := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sectorCCW.font = testSectorFont()
	if err := sectorCCW.SetContainer(containerCCW); err != nil {
		t.Fatal(err)
	}
	containerCCW.AddChild(sectorCCW)

	LayoutRadialTable(containerCCW, containerCCW.LayoutStyle(), &labelTestWriter{t: t})

	if got, want := sectorCCW.geometry.StartAngle, 0.0; !floatEquals(got, want) {
		t.Fatalf("ccw full-circle start angle = %v, want %v", got, want)
	}
	if got, want := sectorCCW.geometry.EndAngle, 360.0; !floatEquals(got, want) {
		t.Fatalf("ccw full-circle end angle = %v, want %v", got, want)
	}

	containerCW := positionedContainer(0, 0, 200, 200)
	containerCW.SetScope(&defaultScope)
	containerCW.SetAttrs(map[string]string{
		"layout": "radial",
		"rows":   "1",
		"angles": "0",
		"sweep":  "cw",
	})

	sectorCW := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sectorCW.font = testSectorFont()
	if err := sectorCW.SetContainer(containerCW); err != nil {
		t.Fatal(err)
	}
	containerCW.AddChild(sectorCW)

	LayoutRadialTable(containerCW, containerCW.LayoutStyle(), &labelTestWriter{t: t})

	if got, want := sectorCW.geometry.StartAngle, 0.0; !floatEquals(got, want) {
		t.Fatalf("cw full-circle start angle = %v, want %v", got, want)
	}
	if got, want := sectorCW.geometry.EndAngle, -360.0; !floatEquals(got, want) {
		t.Fatalf("cw full-circle end angle = %v, want %v", got, want)
	}
}

func TestLayoutRadialOut_RowZeroIsInnermostAndRowspanExtendsOutward(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout": "radial-out",
		"rows":   "3",
		"cols":   "1",
		"r0":     "10",
	})

	s1 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	s1.font = testSectorFont()
	s1.SetAttrs(map[string]string{"rowspan": "2"})
	if err := s1.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(s1)

	s2 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	s2.font = testSectorFont()
	if err := s2.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(s2)

	LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t})

	if got, want := s1.geometry.InnerRadius, 10.0; !floatEquals(got, want) {
		t.Fatalf("row 0 inner radius = %v, want %v", got, want)
	}
	if got, want := s1.geometry.OuterRadius, 70.0; !floatEquals(got, want) {
		t.Fatalf("row 0 rowspan=2 outer radius = %v, want %v", got, want)
	}
	if got, want := s2.geometry.InnerRadius, 70.0; !floatEquals(got, want) {
		t.Fatalf("row 2 inner radius = %v, want %v", got, want)
	}
	if got, want := s2.geometry.OuterRadius, 100.0; !floatEquals(got, want) {
		t.Fatalf("row 2 outer radius = %v, want %v", got, want)
	}
}

func TestLayoutRadialTable_R0AliasSetsInnerRadius(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout": "radial",
		"rows":   "1",
		"cols":   "1",
		"r":      "60",
		"r0":     "10",
	})

	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	if err := sector.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	container.AddChild(sector)

	LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t})

	if got, want := sector.geometry.InnerRadius, 10.0; !floatEquals(got, want) {
		t.Fatalf("r0 inner radius = %v, want %v", got, want)
	}
	if got, want := sector.geometry.OuterRadius, 60.0; !floatEquals(got, want) {
		t.Fatalf("r outer radius = %v, want %v", got, want)
	}
}

func TestLayoutRadialOut_OrderColsFillsOutwardBeforeNextAngularSlot(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout": "radial-out",
		"rows":   "2",
		"order":  "cols",
	})

	s1 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	s2 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	s3 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	for _, sector := range []*StdSector{s1, s2, s3} {
		sector.font = testSectorFont()
		if err := sector.SetContainer(container); err != nil {
			t.Fatal(err)
		}
		container.AddChild(sector)
	}

	LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t})

	if got, want := s1.geometry.InnerRadius, 0.0; !floatEquals(got, want) {
		t.Fatalf("first sector inner radius = %v, want %v", got, want)
	}
	if got, want := s1.geometry.OuterRadius, 50.0; !floatEquals(got, want) {
		t.Fatalf("first sector outer radius = %v, want %v", got, want)
	}
	if got, want := s1.geometry.StartAngle, 0.0; !floatEquals(got, want) {
		t.Fatalf("first sector start angle = %v, want %v", got, want)
	}
	if got, want := s1.geometry.EndAngle, 180.0; !floatEquals(got, want) {
		t.Fatalf("first sector end angle = %v, want %v", got, want)
	}
	if got, want := s2.geometry.InnerRadius, 50.0; !floatEquals(got, want) {
		t.Fatalf("second sector inner radius = %v, want %v", got, want)
	}
	if got, want := s2.geometry.OuterRadius, 100.0; !floatEquals(got, want) {
		t.Fatalf("second sector outer radius = %v, want %v", got, want)
	}
	if got, want := s2.geometry.StartAngle, 0.0; !floatEquals(got, want) {
		t.Fatalf("second sector start angle = %v, want %v", got, want)
	}
	if got, want := s2.geometry.EndAngle, 180.0; !floatEquals(got, want) {
		t.Fatalf("second sector end angle = %v, want %v", got, want)
	}
	if got, want := s3.geometry.InnerRadius, 0.0; !floatEquals(got, want) {
		t.Fatalf("third sector inner radius = %v, want %v", got, want)
	}
	if got, want := s3.geometry.OuterRadius, 50.0; !floatEquals(got, want) {
		t.Fatalf("third sector outer radius = %v, want %v", got, want)
	}
	if got, want := s3.geometry.StartAngle, 180.0; !floatEquals(got, want) {
		t.Fatalf("third sector start angle = %v, want %v", got, want)
	}
	if got, want := s3.geometry.EndAngle, 360.0; !floatEquals(got, want) {
		t.Fatalf("third sector end angle = %v, want %v", got, want)
	}
}

func TestLayoutRadialOut_UsesExplicitAnglesAndBaseAngle(t *testing.T) {
	container := positionedContainer(0, 0, 240, 240)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout":     "radial-out",
		"rows":       "1",
		"base-angle": "10",
		"angles":     "0,120,360",
	})

	s1 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	s2 := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	for _, sector := range []*StdSector{s1, s2} {
		sector.font = testSectorFont()
		if err := sector.SetContainer(container); err != nil {
			t.Fatal(err)
		}
		container.AddChild(sector)
	}

	LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t})

	if s1.geometry.StartAngle != 10 || s1.geometry.EndAngle != 130 {
		t.Fatalf("sector 1 angles = %v..%v, want 10..130", s1.geometry.StartAngle, s1.geometry.EndAngle)
	}
	if s2.geometry.StartAngle != 130 || s2.geometry.EndAngle != 370 {
		t.Fatalf("sector 2 angles = %v..%v, want 130..370", s2.geometry.StartAngle, s2.geometry.EndAngle)
	}
}

func TestLayoutRadialOut_CWSweepPreservesInsideOutRowSemantics(t *testing.T) {
	container := positionedContainer(0, 0, 200, 200)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout": "radial-out",
		"rows":   "1",
		"cols":   "4",
		"r0":     "20",
		"sweep":  "cw",
	})

	sectors := []*StdSector{
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
		{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}},
	}
	for _, sector := range sectors {
		sector.font = testSectorFont()
		if err := sector.SetContainer(container); err != nil {
			t.Fatal(err)
		}
		container.AddChild(sector)
	}

	LayoutRadialTable(container, container.LayoutStyle(), &labelTestWriter{t: t})

	if got, want := sectors[0].geometry.InnerRadius, 20.0; !floatEquals(got, want) {
		t.Fatalf("innermost row inner radius = %v, want %v", got, want)
	}
	if got, want := sectors[0].geometry.OuterRadius, 100.0; !floatEquals(got, want) {
		t.Fatalf("innermost row outer radius = %v, want %v", got, want)
	}
	if got, want := sectors[0].geometry.StartAngle, 0.0; !floatEquals(got, want) {
		t.Fatalf("first clockwise sector start angle = %v, want %v", got, want)
	}
	if got, want := sectors[0].geometry.EndAngle, -90.0; !floatEquals(got, want) {
		t.Fatalf("first clockwise sector end angle = %v, want %v", got, want)
	}
}

func TestLayoutVBox_RadialChildWithWidthOnlyAndInnerRadiusDoesNotPanic(t *testing.T) {
	page := positionedContainer(0, 0, 360, 720)
	page.SetScope(&defaultScope)
	page.SetAttrs(map[string]string{"layout": "vbox"})

	radial := &StdContainer{}
	radial.SetScope(&defaultScope)
	if err := radial.SetContainer(page); err != nil {
		t.Fatal(err)
	}
	radial.SetAttrs(map[string]string{
		"layout": "radial-out",
		"rows":   "1",
		"cols":   "1",
		"r0":     "43.2",
		"width":  "100%",
	})
	page.AddChild(radial)

	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	if err := sector.SetContainer(radial); err != nil {
		t.Fatal(err)
	}
	addTestSectorLabel(t, sector, "Luke", nil)
	radial.AddChild(sector)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LayoutVBox panicked for width-only radial child: %v", r)
		}
	}()

	LayoutVBox(page, page.LayoutStyle(), &labelTestWriter{t: t})

	if radial.Height() <= 0 {
		t.Fatalf("radial height = %v, want > 0", radial.Height())
	}
	if sector.geometry.OuterRadius <= sector.geometry.InnerRadius {
		t.Fatalf("sector radii = %v..%v, want outer > inner", sector.geometry.InnerRadius, sector.geometry.OuterRadius)
	}
}

func TestLayoutVBox_RadialChildWithHeightOnlyAndInnerRadiusDoesNotPanic(t *testing.T) {
	page := positionedContainer(0, 0, 360, 720)
	page.SetScope(&defaultScope)
	page.SetAttrs(map[string]string{"layout": "vbox"})

	radial := &StdContainer{}
	radial.SetScope(&defaultScope)
	if err := radial.SetContainer(page); err != nil {
		t.Fatal(err)
	}
	radial.SetAttrs(map[string]string{
		"layout": "radial-out",
		"rows":   "1",
		"cols":   "1",
		"r0":     "43.2",
		"height": "120",
	})
	page.AddChild(radial)

	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	if err := sector.SetContainer(radial); err != nil {
		t.Fatal(err)
	}
	addTestSectorLabel(t, sector, "Leia", nil)
	radial.AddChild(sector)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LayoutVBox panicked for height-only radial child: %v", r)
		}
	}()

	LayoutVBox(page, page.LayoutStyle(), &labelTestWriter{t: t})

	if radial.Width() <= 0 {
		t.Fatalf("radial width = %v, want > 0", radial.Width())
	}
	if sector.geometry.OuterRadius <= sector.geometry.InnerRadius {
		t.Fatalf("sector radii = %v..%v, want outer > inner", sector.geometry.InnerRadius, sector.geometry.OuterRadius)
	}
}

func TestStdSector_OriginAliasesResolveToSectorReferencePoints(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 35, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX:     100,
		CenterY:     100,
		InnerRadius: 20,
		OuterRadius: 50,
		StartAngle:  0,
		EndAngle:    90,
		AnchorAngle: 45,
		AnchorX:     ax,
		AnchorY:     ay,
	})

	child := &StdWidget{}
	if err := child.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	child.SetAttrs(map[string]string{
		"position": "relative",
		"origin-x": "start",
		"origin-y": "outer",
		"left":     "0",
		"top":      "0",
	})

	if got, want := child.Left(), sector.ResolveSectorReferenceX(child); got != want {
		t.Fatalf("Left() = %v, want %v", got, want)
	}
	if got, want := child.Top(), sector.ResolveSectorReferenceY(child); got != want {
		t.Fatalf("Top() = %v, want %v", got, want)
	}
	if got, want := child.OriginXValue(), sector.ResolveSectorReferenceX(child); got != want {
		t.Fatalf("OriginXValue() = %v, want %v", got, want)
	}
	if got, want := child.OriginYValue(), sector.ResolveSectorReferenceY(child); got != want {
		t.Fatalf("OriginYValue() = %v, want %v", got, want)
	}
}

func TestStdSector_UnspecifiedOriginsResolveToSectorMidpoint(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 35, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX:     100,
		CenterY:     100,
		InnerRadius: 20,
		OuterRadius: 50,
		StartAngle:  0,
		EndAngle:    90,
		AnchorAngle: 45,
		AnchorX:     ax,
		AnchorY:     ay,
	})

	child := &StdWidget{}
	if err := child.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	child.SetAttrs(map[string]string{
		"position": "relative",
		"left":     "0",
		"top":      "0",
	})

	wantX, wantY := ax, ay
	if got := child.OriginX(); got != OriginXUnspecified {
		t.Fatalf("OriginX() = %v, want %v", got, OriginXUnspecified)
	}
	if got := child.OriginY(); got != OriginYUnspecified {
		t.Fatalf("OriginY() = %v, want %v", got, OriginYUnspecified)
	}
	if got := sector.ResolveSectorReferenceX(child); got != wantX {
		t.Fatalf("ResolveSectorReferenceX() = %v, want %v", got, wantX)
	}
	if got := sector.ResolveSectorReferenceY(child); got != wantY {
		t.Fatalf("ResolveSectorReferenceY() = %v, want %v", got, wantY)
	}
	if got := child.Left(); got != wantX {
		t.Fatalf("Left() = %v, want %v", got, wantX)
	}
	if got := child.Top(); got != wantY {
		t.Fatalf("Top() = %v, want %v", got, wantY)
	}
}

func TestStdSector_MixedCustomAndRadialOriginsResolvePerAxis(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 35, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX:     100,
		CenterY:     100,
		InnerRadius: 20,
		OuterRadius: 50,
		StartAngle:  0,
		EndAngle:    90,
		AnchorAngle: 45,
		AnchorX:     ax,
		AnchorY:     ay,
	})

	tests := []struct {
		name  string
		attrs map[string]string
		wantX float64
		wantY float64
	}{
		{
			name: "custom x unspecified y",
			attrs: map[string]string{
				"position": "relative",
				"units":    "pt",
				"origin-x": "12",
				"left":     "0",
				"top":      "0",
			},
			wantX: sector.geometry.AnchorX + sector.localBounds.MinX + 12,
			wantY: func() float64 {
				_, y := sector.toLocal(ax, ay)
				return sector.geometry.AnchorY + y
			}(),
		},
		{
			name: "unspecified x custom y",
			attrs: map[string]string{
				"position": "relative",
				"units":    "pt",
				"origin-y": "8",
				"left":     "0",
				"top":      "0",
			},
			wantX: ax,
			wantY: sector.geometry.AnchorY + sector.localBounds.MinY + 8,
		},
		{
			name: "start x custom y",
			attrs: map[string]string{
				"position": "relative",
				"units":    "pt",
				"origin-x": "start",
				"origin-y": "8",
				"left":     "0",
				"top":      "0",
			},
			wantX: func() float64 {
				x, y := radialPointAt(sector.geometry.CenterX, sector.geometry.CenterY, 35, sector.geometry.StartAngle)
				localX, _ := sector.toLocal(x, y)
				return sector.geometry.AnchorX + localX
			}(),
			wantY: sector.geometry.AnchorY + sector.localBounds.MinY + 8,
		},
		{
			name: "custom x outer y",
			attrs: map[string]string{
				"position": "relative",
				"units":    "pt",
				"origin-x": "12",
				"origin-y": "outer",
				"left":     "0",
				"top":      "0",
			},
			wantX: sector.geometry.AnchorX + sector.localBounds.MinX + 12,
			wantY: func() float64 {
				x, y := radialPointAt(sector.geometry.CenterX, sector.geometry.CenterY, sector.geometry.OuterRadius, sector.geometry.AnchorAngle)
				_, localY := sector.toLocal(x, y)
				return sector.geometry.AnchorY + localY
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := &StdWidget{}
			if err := child.SetContainer(sector); err != nil {
				t.Fatal(err)
			}
			child.SetAttrs(tt.attrs)

			if got := sector.ResolveSectorReferenceX(child); got != tt.wantX {
				t.Fatalf("ResolveSectorReferenceX() = %v, want %v", got, tt.wantX)
			}
			if got := sector.ResolveSectorReferenceY(child); got != tt.wantY {
				t.Fatalf("ResolveSectorReferenceY() = %v, want %v", got, tt.wantY)
			}
			if got := child.Left(); got != tt.wantX {
				t.Fatalf("Left() = %v, want %v", got, tt.wantX)
			}
			if got := child.Top(); got != tt.wantY {
				t.Fatalf("Top() = %v, want %v", got, tt.wantY)
			}
		})
	}
}

func TestStdSector_DrawContent_UsesCurvedTextUnlessAngleOverrides(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	label := addTestSectorLabel(t, sector, "Radial", nil)
	ax, ay := radialPointAt(100, 100, 40, 90)
	sector.setGeometry(radialSectorGeometry{
		CenterX:     100,
		CenterY:     100,
		InnerRadius: 20,
		OuterRadius: 60,
		StartAngle:  45,
		EndAngle:    135,
		AnchorAngle: 90,
		AnchorX:     ax,
		AnchorY:     ay,
	})

	w := &labelTestWriter{t: t}
	if err := sector.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if w.curvedCount != 1 {
		t.Fatalf("curved draw count = %d, want 1", w.curvedCount)
	}
	if len(w.curvedOpts) != 1 || w.curvedOpts[0].VAlign != pdf.VTextAlignMiddle {
		t.Fatalf("curved valign = %v, want %v", w.curvedOpts, pdf.VTextAlignMiddle)
	}
	if len(w.rotations) != 0 {
		t.Fatalf("rotation count = %d, want 0 for curved text", len(w.rotations))
	}

	label.SetAttrs(map[string]string{"angle": "90"})
	w2 := &labelTestWriter{t: t}
	if err := sector.DrawContent(w2); err != nil {
		t.Fatal(err)
	}
	if w2.curvedCount != 0 {
		t.Fatalf("curved draw count with explicit angle = %d, want 0", w2.curvedCount)
	}
	if len(w2.rotations) != 1 {
		t.Fatalf("rotation count with explicit angle = %d, want 1", len(w2.rotations))
	}
}

func TestStdSector_DrawContent_RightAlignedCurvedTextUsesSectorEndAngle(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	addTestSectorLabel(t, sector, "Radial", map[string]string{"position": "relative", "origin-x": "end", "text-align": "right"})
	ax, ay := radialPointAt(100, 100, 40, -45)
	sector.setGeometry(radialSectorGeometry{
		CenterX:     100,
		CenterY:     100,
		InnerRadius: 20,
		OuterRadius: 60,
		StartAngle:  0,
		EndAngle:    -90,
		AnchorAngle: -45,
		AnchorX:     ax,
		AnchorY:     ay,
	})

	w := &labelTestWriter{t: t}
	if err := sector.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.curvedStarts) != 1 {
		t.Fatalf("curved start count = %d, want 1", len(w.curvedStarts))
	}
	if got, want := w.curvedStarts[0], -90.0; !floatEquals(got, want) {
		t.Fatalf("curved text start angle = %v, want %v", got, want)
	}
}

func TestStdSector_LabelsUseIndependentRadialAnchors(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 40, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 20, OuterRadius: 60,
		StartAngle: 0, EndAngle: 90,
		AnchorAngle: 45, AnchorX: ax, AnchorY: ay,
	})

	attrs := []map[string]string{
		{"position": "relative", "origin-x": "start", "origin-y": "inner"},
		{"position": "relative", "origin-x": "center", "origin-y": "middle"},
		{"position": "relative", "origin-x": "end", "origin-y": "outer"},
	}
	for i, attr := range attrs {
		label := &StdLabel{}
		label.font = testSectorFont()
		label.SetAttrs(attr)
		if err := label.SetContainer(sector); err != nil {
			t.Fatal(err)
		}
		label.AddText(strconv.Itoa(i + 1))
		sector.AddChild(label)
	}

	w := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	if err := sector.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(w.curvedStarts, []float64{0, 45, 90}) {
		t.Fatalf("curved anchors = %v, want 0/45/90", w.curvedStarts)
	}
	for i, want := range []float64{20, 40, 60} {
		if !floatEquals(w.curvedRadii[i], want) {
			t.Fatalf("curved radii = %v, want 20/40/60", w.curvedRadii)
		}
	}
	wantAlign := []pdf.CurvedTextHAlign{pdf.CurvedTextAlignLeft, pdf.CurvedTextAlignCenter, pdf.CurvedTextAlignRight}
	for i, opts := range w.curvedOpts {
		if opts.Align != wantAlign[i] {
			t.Fatalf("label %d curved align = %v, want %v", i, opts.Align, wantAlign[i])
		}
	}
}

func TestStdSector_LabelOffsetsUseSectorLocalFrame(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 40, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 20, OuterRadius: 60,
		StartAngle: 0, EndAngle: 90,
		AnchorAngle: 45, AnchorX: ax, AnchorY: ay,
	})
	label := &StdLabel{}
	label.font = testSectorFont()
	label.SetAttrs(map[string]string{
		"units": "pt",
		"left":  "5",
		"top":   "-3",
	})
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.AddText("offset")
	sector.AddChild(label)
	if err := sector.LayoutWidget(&labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
	wantX, wantY := rotatePagePoint(ax+5, ay-3, ax, ay, sector.contentRotation)
	if !floatEquals(label.sectorPlacement.anchorX, wantX) || !floatEquals(label.sectorPlacement.anchorY, wantY) {
		t.Fatalf("offset anchor = (%v,%v), want (%v,%v)", label.sectorPlacement.anchorX, label.sectorPlacement.anchorY, wantX, wantY)
	}
}

func TestStdSector_LabelAlignmentAndFacingOverrideAnchorDefaults(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 40, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 20, OuterRadius: 60,
		StartAngle: 0, EndAngle: 90,
		AnchorAngle: 45, AnchorX: ax, AnchorY: ay,
	})
	label := &StdLabel{}
	label.font = testSectorFont()
	label.SetAttrs(map[string]string{
		"position":    "relative",
		"origin-x":    "start",
		"text-align":  "center",
		"facing":      "upside-down",
		"text-valign": "bottom",
	})
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.AddText("override")
	sector.AddChild(label)

	w := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	if err := sector.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if got := w.curvedStarts[0]; got != 0 {
		t.Fatalf("curved anchor = %v, want sector start", got)
	}
	if got := w.curvedOpts[0].Align; got != pdf.CurvedTextAlignCenter {
		t.Fatalf("curved align = %v, want explicit center", got)
	}
	if got := w.curvedOpts[0].Facing; got != pdf.CurvedTextFacingUpsideDown {
		t.Fatalf("curved facing = %v, want explicit upside-down", got)
	}
	if got := w.curvedOpts[0].VAlign; got != pdf.VTextAlignBelow {
		t.Fatalf("curved valign = %v, want below", got)
	}
}

func TestStdSector_LabelAutomaticOrientationUsesItsOwnAnchor(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 40, 0)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 20, OuterRadius: 60,
		StartAngle: -90, EndAngle: 90,
		AnchorAngle: 0, AnchorX: ax, AnchorY: ay,
	})

	for _, origin := range []string{"start", "end"} {
		label := &StdLabel{}
		label.font = testSectorFont()
		label.SetAttrs(map[string]string{"position": "relative", "origin-x": origin})
		if err := label.SetContainer(sector); err != nil {
			t.Fatal(err)
		}
		label.AddText(origin)
		sector.AddChild(label)
	}

	w := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	if err := sector.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if got := w.curvedOpts[0]; got.Direction != pdf.CurvedTextCounterClockwise || got.Facing != pdf.CurvedTextFacingUpsideDown {
		t.Fatalf("lower-anchor orientation = direction %v facing %v, want counter-clockwise/upside-down", got.Direction, got.Facing)
	}
	if got := w.curvedOpts[1]; got.Direction != pdf.CurvedTextClockwise || got.Facing != pdf.CurvedTextFacingUpright {
		t.Fatalf("upper-anchor orientation = direction %v facing %v, want clockwise/upright", got.Direction, got.Facing)
	}
}

func TestStdSector_ExplicitLabelDoesNotInheritSectorTextSettings(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	sector.SetAttrs(map[string]string{
		"text-align":  "right",
		"text-valign": "bottom",
		"facing":      "upside-down",
	})
	label := addTestSectorLabel(t, sector, "independent", nil)
	if _, straight := label.sectorTextAngle(); straight {
		t.Fatal("sector angle unexpectedly made the label straight")
	}
	if got := label.sectorTextAlign(); got != HAlignCenter {
		t.Fatalf("label align = %v, want center", got)
	}
	if got := label.sectorTextVAlign(); got != VAlignMiddle {
		t.Fatalf("label valign = %v, want middle", got)
	}
	if got := label.sectorTextFacing(); got != sectorFacingAuto {
		t.Fatalf("label facing = %v, want auto", got)
	}
}

func TestStdSector_ExplicitZeroAngleIsAbsoluteAndRetainsStraightLabelBox(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 40, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 20, OuterRadius: 60,
		StartAngle: 0, EndAngle: 90,
		AnchorAngle: 45, AnchorX: ax, AnchorY: ay,
	})
	label := &StdLabel{}
	label.font = testSectorFont()
	label.SetAttrs(map[string]string{
		"angle":   "0",
		"width":   "40pt",
		"height":  "20pt",
		"padding": "2pt",
		"fill":    "Gold",
		"border":  "Blue",
	})
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.AddText("26")
	sector.AddChild(label)

	w := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	if err := sector.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if w.curvedCount != 0 {
		t.Fatalf("curved draws = %d, want straight label", w.curvedCount)
	}
	if len(w.rotations) != 0 {
		t.Fatalf("rotations = %v, want absolute horizontal text without sector rotation", w.rotations)
	}
	if len(w.fillRectPages) != 1 || len(w.rectPages) == 0 {
		t.Fatalf("straight box fills/rectangles = %d/%d, want retained fill and border", len(w.fillRectPages), len(w.rectPages))
	}
}

func TestStdSector_LabelPaintRecomputesInvalidatedPlacementBeforeBackground(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 40, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 20, OuterRadius: 60,
		StartAngle: 0, EndAngle: 90,
		AnchorAngle: 45, AnchorX: ax, AnchorY: ay,
	})
	label := &StdLabel{}
	label.font = testSectorFont()
	label.SetAttrs(map[string]string{"angle": "0", "fill": "Gold"})
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.AddText("moved")
	sector.AddChild(label)
	w := &labelTestWriter{t: t}
	if err := label.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	originalX := label.sectorPlacement.anchorX

	label.SetAttrs(map[string]string{"units": "pt", "left": "12"})
	if label.sectorPlacement != nil {
		t.Fatal("SetAttrs did not invalidate sector placement")
	}
	if err := label.paintWithTransform(w, func() error {
		if label.sectorPlacement == nil {
			t.Fatal("placement was not restored before paint phases")
		}
		return label.PaintBackground(w)
	}); err != nil {
		t.Fatal(err)
	}
	if got := label.sectorPlacement.anchorX; floatEquals(got, originalX) {
		t.Fatalf("recomputed anchor x = %v, want offset from %v", got, originalX)
	}
	if len(w.fillRectPages) != 1 {
		t.Fatalf("background fills = %d, want one after lazy placement", len(w.fillRectPages))
	}
}

func TestStdSector_CurvedLabelIgnoresBoxAttrsUntilMadeStraight(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 40, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 20, OuterRadius: 60,
		StartAngle: 0, EndAngle: 90,
		AnchorAngle: 45, AnchorX: ax, AnchorY: ay,
	})
	label := addTestSectorLabel(t, sector, "effective", map[string]string{
		"width": "20pt", "height": "14pt", "padding": "2pt", "fill": "Gold", "border": "Blue", "rotate": "15",
	})
	w := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatalf("curved box attrs rejected: %v", err)
	}
	if err := sector.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.fillRectPages) != 0 || len(w.rectPages) != 0 || len(w.rotations) != 0 {
		t.Fatalf("curved box paint = fills %d borders %d rotations %d, want all dormant", len(w.fillRectPages), len(w.rectPages), len(w.rotations))
	}
	label.SetAttrs(map[string]string{"angle": "0"})
	label.SetPrinted(false)
	w2 := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w2); err != nil {
		t.Fatal(err)
	}
	if err := sector.DrawContent(w2); err != nil {
		t.Fatal(err)
	}
	if len(w2.fillRectPages) == 0 || len(w2.rectPages) == 0 || len(w2.rotations) == 0 {
		t.Fatalf("straight box paint = fills %d borders %d rotations %d, want dormant attrs active", len(w2.fillRectPages), len(w2.rectPages), len(w2.rotations))
	}
}

func TestStdSector_CurvedLabelRejectsZeroRadius(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 30, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 0, OuterRadius: 60,
		StartAngle: 0, EndAngle: 90,
		AnchorAngle: 45, AnchorX: ax, AnchorY: ay,
	})
	label := &StdLabel{}
	label.font = testSectorFont()
	label.SetAttrs(map[string]string{"position": "relative", "origin-y": "inner"})
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.AddText("center")
	sector.AddChild(label)
	err := sector.LayoutWidget(&labelTestWriter{t: t})
	if err == nil || !strings.Contains(err.Error(), "positive finite radius") {
		t.Fatalf("LayoutWidget() error = %v, want zero-radius error", err)
	}
}

func TestStdSector_CurvedLabelShrinkFitsSectorArc(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 50, 5)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 40, OuterRadius: 60,
		StartAngle: 0, EndAngle: 10,
		AnchorAngle: 5, AnchorX: ax, AnchorY: ay,
	})
	label := &StdLabel{}
	label.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 24}
	label.SetAttrs(map[string]string{"fit": "shrink"})
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.AddText("A long curved label")
	sector.AddChild(label)
	w := &labelTestWriter{t: t}
	originalWidth := label.RichText(w).Width()
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	if err := sector.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.printed) != 1 || !(w.printed[0].Width() < originalWidth) {
		t.Fatalf("curved fitted width = %v, original %v", w.printed[0].Width(), originalWidth)
	}
}

func TestStdSector_LayoutWidget_CentersSingleStaticChildInSector(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 40, 90)
	sector.setGeometry(radialSectorGeometry{
		CenterX:     100,
		CenterY:     100,
		InnerRadius: 20,
		OuterRadius: 60,
		StartAngle:  45,
		EndAngle:    135,
		AnchorAngle: 90,
		AnchorX:     ax,
		AnchorY:     ay,
	})

	label := &StdLabel{}
	label.font = testSectorFont()
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.AddText("42")
	sector.AddChild(label)

	w := &labelTestWriter{t: t}
	sector.LayoutWidget(w)

	localX, localY := sector.contentLocalCenter()
	wantX, wantY := rotatePagePoint(sector.geometry.AnchorX+localX, sector.geometry.AnchorY+localY,
		sector.geometry.AnchorX, sector.geometry.AnchorY, sector.contentRotation)
	if got, want := (label.Left()+label.Right())/2, wantX; math.Abs(got-want) > 0.5 {
		t.Fatalf("label center x = %v, want near %v", got, want)
	}
	if got, want := (label.Top()+label.Bottom())/2, wantY; math.Abs(got-want) > 5 {
		t.Fatalf("label center y = %v, want near %v", got, want)
	}
}

func TestStdSector_ParagraphLayoutVariesLineWidthsAcrossSector(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(120, 120, 70, 90)
	sector.setGeometry(radialSectorGeometry{
		CenterX:     120,
		CenterY:     120,
		InnerRadius: 10,
		OuterRadius: 130,
		StartAngle:  75,
		EndAngle:    105,
		AnchorAngle: 90,
		AnchorX:     ax,
		AnchorY:     ay,
	})

	p := &StdParagraph{}
	p.paragraphStyle = &ParagraphStyle{}
	if err := p.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	p.AddText("This paragraph should wrap onto multiple sector-shaped lines instead of using one fixed rectangular width.")
	sector.AddChild(p)

	w := &labelTestWriter{t: t}
	sector.LayoutWidget(w)
	layout := sector.sectorParagraphLayoutFor(p, w)
	if len(layout.lines) < 2 {
		t.Fatalf("sector paragraph line count = %d, want at least 2", len(layout.lines))
	}
	varying := false
	for i := 1; i < len(layout.intervals); i++ {
		if math.Abs((layout.intervals[i].MaxX-layout.intervals[i].MinX)-(layout.intervals[i-1].MaxX-layout.intervals[i-1].MinX)) > 0.5 {
			varying = true
			break
		}
	}
	if !varying {
		t.Fatalf("expected varying line widths across the sector, got intervals %v", layout.intervals)
	}
}

func TestPolygonLineIntervalAtPrefersComponentNearestContentCentroid(t *testing.T) {
	// A horseshoe has two disconnected horizontal components through its
	// middle, just as a wide annular sector can. The chosen component must stay
	// on the side nearest the sector centroid instead of switching to whichever
	// component happens to be widest.
	polygon := []radialPoint{
		{X: -10, Y: -5}, {X: 10, Y: -5}, {X: 10, Y: 5}, {X: 6, Y: 5},
		{X: 6, Y: -1}, {X: -6, Y: -1}, {X: -6, Y: 5}, {X: -10, Y: 5},
	}
	bounds := boundsForPoints(polygon)
	if got := polygonLineIntervalAt(polygon, bounds, 0, 7); got.MinX != 6 || got.MaxX != 10 {
		t.Fatalf("right-side interval = %#v, want 6..10", got)
	}
	if got := polygonLineIntervalAt(polygon, bounds, 0, -7); got.MinX != -10 || got.MaxX != -6 {
		t.Fatalf("left-side interval = %#v, want -10..-6", got)
	}
}

func TestStdSector_ParagraphIntervalsContainCompleteLineBoxes(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	p := &StdParagraph{}
	p.paragraphStyle = &ParagraphStyle{}
	if err := p.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	p.AddText("Curved sector line widths must contain complete glyph boxes rather than only their centerlines.")
	sector.AddChild(p)

	ax, ay := radialPointAt(120, 120, 70, 55)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 120, CenterY: 120,
		InnerRadius: 25, OuterRadius: 115,
		StartAngle: 10, EndAngle: 100,
		AnchorAngle: 55, AnchorX: ax, AnchorY: ay,
	})
	w := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	layout := sector.sectorParagraphLayoutFor(p, w)
	y := ContentTop(p) - sector.geometry.AnchorY
	for i, line := range layout.lines {
		height := line.Leading() * w.LineSpacing()
		band := layout.intervals[i]
		for _, sampleY := range []float64{y, y + height/2, y + height} {
			chord := sector.contentLineIntervalAt(sampleY)
			if band.MinX < chord.MinX-0.01 || band.MaxX > chord.MaxX+0.01 {
				t.Fatalf("line %d band %#v escapes chord %#v at y=%v", i, band, chord, sampleY)
			}
		}
		y += height
	}
}

func TestStdSector_StaticParagraphsParticipateInFlow(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(120, 120, 70, 90)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 120, CenterY: 120,
		InnerRadius: 20, OuterRadius: 120,
		StartAngle: 65, EndAngle: 115,
		AnchorAngle: 90, AnchorX: ax, AnchorY: ay,
	})
	paragraphs := make([]*StdParagraph, 2)
	for i := range paragraphs {
		paragraph := &StdParagraph{}
		paragraph.paragraphStyle = &ParagraphStyle{}
		if err := paragraph.SetContainer(sector); err != nil {
			t.Fatal(err)
		}
		paragraph.AddText("Independent sector paragraph")
		sector.AddChild(paragraph)
		paragraphs[i] = paragraph
	}

	if err := sector.LayoutWidget(&labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
	if paragraphs[1].Top() <= paragraphs[0].Top() || paragraphs[1].Top() < paragraphs[0].Bottom()-0.01 {
		t.Fatalf("paragraph placements = (%v,%v) and (%v,%v), want non-overlapping source-order flow",
			paragraphs[0].Left(), paragraphs[0].Top(), paragraphs[1].Left(), paragraphs[1].Top())
	}
}

func TestStdSector_ParagraphUsesShapedBandInsteadOfRectangularFlowWidth(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	p := &StdParagraph{}
	p.paragraphStyle = &ParagraphStyle{}
	if err := p.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	p.AddText("Paragraph flow footprint")
	sector.AddChild(p)
	ax, ay := radialPointAt(100, 100, 60, 90)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 20, OuterRadius: 100,
		StartAngle: 50, EndAngle: 130,
		AnchorAngle: 90, AnchorX: ax, AnchorY: ay,
	})
	items, err := sector.sectorFlowItems([]Widget{p}, &labelTestWriter{t: t})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || !items[0].fullBand || items[0].width != 0 {
		t.Fatalf("paragraph flow item = %#v, want a zero-width full band", items)
	}
}

func TestStdSector_RadialPaddingDefinesContentGeometry(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.SetAttrs(map[string]string{
		"units": "pt", "padding-top": "10", "padding-right": "6", "padding-bottom": "5", "padding-left": "4",
	})
	ax, ay := radialPointAt(100, 100, 40, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100, InnerRadius: 20, OuterRadius: 60,
		StartAngle: 0, EndAngle: 90, AnchorAngle: 45, AnchorX: ax, AnchorY: ay,
	})
	inner, outer := sector.contentRadii()
	if inner != 25 || outer != 50 {
		t.Fatalf("content radii = %v..%v, want 25..50", inner, outer)
	}
	startAngle := sector.contentBoundaryAngle(true, 40)
	endAngle := sector.contentBoundaryAngle(false, 40)
	if startAngle <= 0 || endAngle >= 90 {
		t.Fatalf("padded boundary angles = %v..%v, want inset from 0..90", startAngle, endAngle)
	}
	_, startY := radialPointAt(100, 100, 40, startAngle)
	if got := math.Abs(startY - 100); math.Abs(got-4) > 0.01 {
		t.Fatalf("start-edge physical inset = %v, want 4", got)
	}
	endX, _ := radialPointAt(100, 100, 40, endAngle)
	if got := math.Abs(endX - 100); math.Abs(got-6) > 0.01 {
		t.Fatalf("end-edge physical inset = %v, want 6", got)
	}
	if len(sector.contentPolygon) < 3 || sector.contentBounds.MaxX <= sector.contentBounds.MinX || sector.contentBounds.MaxY <= sector.contentBounds.MinY {
		t.Fatalf("invalid padded polygon: %v %#v", len(sector.contentPolygon), sector.contentBounds)
	}
}

func TestStdSector_StaticWidgetsPackAndPositionedOverlayLeavesFlow(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	sector.SetAttrs(map[string]string{"units": "pt", "layout": "hbox", "layout.hpadding": "3", "layout.vpadding": "4"})
	ax, ay := radialPointAt(120, 120, 65, 90)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 120, CenterY: 120, InnerRadius: 20, OuterRadius: 110,
		StartAngle: 55, EndAngle: 125, AnchorAngle: 90, AnchorX: ax, AnchorY: ay,
	})
	static := make([]*StdLabel, 3)
	for i := range static {
		static[i] = addTestSectorLabel(t, sector, strconv.Itoa(i+1), map[string]string{
			"angle": "0", "width": "22pt", "height": "14pt",
		})
	}
	overlay := addTestSectorLabel(t, sector, "overlay", map[string]string{
		"position": "relative", "origin-x": "end", "origin-y": "outer",
	})
	w := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	if len(sector.flowSlots) != len(static) {
		t.Fatalf("flow slots = %d, want %d static children", len(sector.flowSlots), len(static))
	}
	if _, ok := sector.flowSlots[overlay]; ok {
		t.Fatal("positioned overlay unexpectedly participated in flow")
	}
	for i := range static {
		for j := i + 1; j < len(static); j++ {
			a, b := sector.flowSlots[static[i]], sector.flowSlots[static[j]]
			overlaps := a.MinX < b.MaxX-0.01 && b.MinX < a.MaxX-0.01 && a.MinY < b.MaxY-0.01 && b.MinY < a.MaxY-0.01
			if overlaps {
				t.Fatalf("static flow slots overlap: %#v %#v", a, b)
			}
		}
	}
	if overlay.sectorPlacement == nil {
		t.Fatal("positioned overlay was not laid out")
	}
}

func TestStdSector_StaticCurvedLabelsReceiveDistinctFlowAnchors(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(100, 100, 45, 90)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100, InnerRadius: 20, OuterRadius: 70,
		StartAngle: 30, EndAngle: 150, AnchorAngle: 90, AnchorX: ax, AnchorY: ay,
	})
	first := addTestSectorLabel(t, sector, "Alpha", nil)
	second := addTestSectorLabel(t, sector, "Beta", nil)
	w := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	if err := sector.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if w.curvedCount != 2 || floatEquals(first.sectorPlacement.angle, second.sectorPlacement.angle) {
		t.Fatalf("curved flow draws/angles = %d, %v/%v", w.curvedCount, first.sectorPlacement.angle, second.sectorPlacement.angle)
	}
}

func TestStdSector_WithParagraphChild_DoesNotDefaultToTangentRotation(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	p := &StdParagraph{}
	p.paragraphStyle = &ParagraphStyle{}
	if err := p.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	p.AddText("Paragraph")
	sector.AddChild(p)

	ax, ay := radialPointAt(100, 100, 40, 45)
	sector.setGeometry(radialSectorGeometry{
		CenterX:     100,
		CenterY:     100,
		InnerRadius: 20,
		OuterRadius: 60,
		StartAngle:  0,
		EndAngle:    90,
		AnchorAngle: 45,
		AnchorX:     ax,
		AnchorY:     ay,
	})

	if got := sector.contentRotation; math.Abs(got) > 0.001 {
		t.Fatalf("content rotation = %v, want 0 for paragraph-bearing sector", got)
	}
}

func TestRadialSample_ImplicitParagraphsUseSectorFlow(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_038_radial_layout.ltml"))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	var implicit []*StdParagraph
	walkWidgets(doc.Root(), func(widget Widget) bool {
		paragraph, ok := widget.(*StdParagraph)
		if !ok {
			return true
		}
		if sector, ok := paragraph.Container().(*StdSector); ok && sector.Tag == "" {
			implicit = append(implicit, paragraph)
		}
		return true
	})

	if len(implicit) != 2 {
		t.Fatalf("implicit paragraph count = %d, want 2", len(implicit))
	}
	for _, paragraph := range implicit {
		sector, ok := paragraph.Container().(*StdSector)
		if !ok || sector.Tag != "" {
			t.Fatalf("paragraph container = %#v, want transparent sector", paragraph.Container())
		}
		slot, ok := sector.flowSlots[paragraph]
		if !ok {
			t.Fatalf("paragraph %q has no sector flow slot", paragraph.AccessibilityText())
		}
		if slot.MinY < sector.contentBounds.MinY-0.01 || slot.MaxY > sector.contentBounds.MaxY+0.01 {
			t.Fatalf("paragraph slot = %#v, usable bounds %#v", slot, sector.contentBounds)
		}
	}
}
