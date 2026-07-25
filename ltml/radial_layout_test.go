package ltml

import (
	"maps"
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
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.SetAttrs(attrs)
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

func TestStdSectorBorderNoneUsesPhysicalLogicalAggregatePrecedence(t *testing.T) {
	sector := &StdSector{}
	sector.SetAttrs(map[string]string{
		"border":       "solid",
		"border-top":   "Red",
		"border-outer": "none",
		"border-left":  "none",
	})

	outer, inner, start, end := sector.resolvedSectorBorders()
	if outer != nil {
		t.Fatalf("outer border = %#v, want physical none", outer)
	}
	if inner == nil || end == nil {
		t.Fatalf("aggregate fallback inner/end = %#v/%#v, want solid", inner, end)
	}
	if start != nil {
		t.Fatalf("start border = %#v, want mapped logical none", start)
	}
}

func TestStdSectorBorderNoneRequiresExplicitPenToReviveAlias(t *testing.T) {
	sector := &StdSector{}
	sector.SetAttrs(map[string]string{"border-outer": "none"})
	sector.SetAttrs(map[string]string{"border-outer.color": "Red"})
	if outer, _, _, _ := sector.resolvedSectorBorders(); outer != nil {
		t.Fatalf("outer border = %#v, subattribute revived none", outer)
	}
	sector.SetAttrs(map[string]string{"border-outer": "solid"})
	if outer, _, _, _ := sector.resolvedSectorBorders(); outer == nil {
		t.Fatal("explicit pen did not revive outer border")
	}
}

func TestStdSectorContentClipping(t *testing.T) {
	sector := &StdSector{
		contentPolygon: []radialPoint{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 0, Y: 10}},
	}

	for _, test := range []struct {
		name      string
		attrs     []map[string]string
		wantClips int
	}{
		{name: "default", wantClips: 1},
		{name: "explicit true", attrs: []map[string]string{{"clip": "true"}}, wantClips: 1},
		{name: "disabled", attrs: []map[string]string{{"clip": "false"}}, wantClips: 0},
		{name: "later disabled", attrs: []map[string]string{{"clip": "true"}, {"clip": "false"}}, wantClips: 0},
		{name: "later enabled", attrs: []map[string]string{{"clip": "false"}, {"clip": "true"}}, wantClips: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			sector.clipDisabled = false
			for _, attrs := range test.attrs {
				sector.SetAttrs(attrs)
			}
			writer := &labelTestWriter{t: t}
			painted := 0
			if err := sector.withSectorClip(writer, func() error {
				painted++
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			if writer.clipCalls != test.wantClips || painted != 1 {
				t.Fatalf("clip calls/paint calls = %d/%d, want %d/1", writer.clipCalls, painted, test.wantClips)
			}
		})
	}
}

func TestStdSectorClipDisabledStillSuppressesCollapsedGeometry(t *testing.T) {
	sector := &StdSector{clipDisabled: true}
	writer := &labelTestWriter{t: t}
	painted := false
	if err := sector.withSectorClip(writer, func() error {
		painted = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if painted || writer.clipCalls != 0 {
		t.Fatalf("collapsed sector painted/clipped = %v/%d, want false/0", painted, writer.clipCalls)
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
      label { border: Red; padding: 2pt; font.size: 9pt; clip: true; }
      label:first-child { fill: Gold; z-index: 7; clip: false; }
    </style>
    <div layout="radial" cols="2">
      <label id="source" class="number" units="pt" colspan="2" border="Blue"
             angle="0" width="30" start="4" outer="5">26</label>
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
	if !sector.clipDisabled {
		t.Fatal("implicit sector did not receive clip=false")
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
	placement := sector.positionedChildren[&label.StdWidget]
	if placement.angularEdge != sectorAngularStart || placement.angularInset != 4 ||
		placement.radialEdge != sectorRadialOuter || placement.radialInset != 5 {
		t.Fatalf("child radial placement = %#v, want start 4/outer 5", placement)
	}
	if _, ok := sector.positionedChildren[&sector.StdWidget]; ok {
		t.Fatal("implicit wrapper unexpectedly owns child radial placement")
	}
	if label.Position() != Relative {
		t.Fatalf("radial attributes position = %v, want relative", label.Position())
	}
}

func TestParse_RadialPlacementCascadeAcrossDefaultsSelectorsPseudoAndDirectAttrs(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <define id="radial-label" tag="label" units="pt" start="1" outer="1" />
    <style>
      sector > label { start: 2pt; outer: 3pt; }
      sector > label:first-child { end: 4pt; inner: 5pt; }
    </style>
    <div layout="radial" cols="4">
      <sector>
        <label id="pseudo">Pseudo</label>
        <radial-label id="defaults">Defaults</radial-label>
        <label id="selector">Selector</label>
        <label id="direct" end="8pt" inner="9pt">Direct</label>
      </sector>
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	sector := doc.Root().Page(0).children[0].(*StdContainer).children[0].(*StdSector)
	if len(sector.children) != 4 {
		t.Fatalf("sector children = %d, want 4", len(sector.children))
	}
	assertPlacement := func(index int, angular sectorAngularEdge, angularInset float32, radial sectorRadialEdge, radialInset float32) {
		t.Helper()
		label := sector.children[index].(*StdLabel)
		got := sector.positionedChildren[&label.StdWidget]
		if got.angularEdge != angular || got.angularInset != angularInset ||
			got.radialEdge != radial || got.radialInset != radialInset {
			t.Fatalf("child %d placement = %#v, want angular %v/%v radial %v/%v",
				index, got, angular, angularInset, radial, radialInset)
		}
	}
	assertPlacement(0, sectorAngularEnd, 4, sectorRadialInner, 5)
	assertPlacement(1, sectorAngularStart, 1, sectorRadialOuter, 1)
	assertPlacement(2, sectorAngularStart, 2, sectorRadialOuter, 3)
	assertPlacement(3, sectorAngularEnd, 8, sectorRadialInner, 9)
}

func TestParse_ImplicitSectorRoutesBorderNoneAcrossCascadeLayers(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <style>
      label { border: solid; border-top: none; }
      label:first-child { border-right: none; }
    </style>
    <div layout="radial" cols="1">
      <label border-left="none">Only the bottom edge remains</label>
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	sector := doc.Root().Page(0).children[0].(*StdContainer).children[0].(*StdSector)
	label := sector.children[0].(*StdLabel)
	if sector.border == nil {
		t.Fatal("implicit sector aggregate border is nil, want solid")
	}
	for _, side := range []int{topSide, rightSide, leftSide} {
		if !sector.borderSideSet[side] || sector.borders[side] != nil {
			t.Fatalf("sector side %s = %#v set=%v, want explicit none", sideNames[side], sector.borders[side], sector.borderSideSet[side])
		}
	}
	if label.border != nil || label.borderSet {
		t.Fatalf("child border = %#v set=%v, want border attrs owned only by wrapper", label.border, label.borderSet)
	}
}

func TestParse_ScopeResourceReplayPreservesBorderNoneAndResolvesLatePen(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page border="late" border-top="none">
    <pen id="late" color="Gold" width="2pt" />
    <label>Late pen</label>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.Root().Page(0)
	if page.border == nil || page.border.color != NamedColor("Gold") || page.border.width != 2 {
		t.Fatalf("late aggregate border = %#v, want resolved Gold 2pt pen", page.border)
	}
	if !page.borderSideSet[topSide] || page.borders[topSide] != nil {
		t.Fatalf("top border = %#v set=%v, want replayed none", page.borders[topSide], page.borderSideSet[topSide])
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
	if got := label.OriginX(); got != OriginXUnspecified {
		t.Fatalf("effective label box origin = %v, want ordinary unspecified/start default", got)
	}
}

func TestParse_SectorDoesNotProvideParagraphDefaults(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <style>
      sector { angle: 0; facing: upside-down; }
      p.curved { angle: 35; facing: upright; }
      p.horizontal { angle: 35; }
    </style>
    <div layout="radial" cols="3">
      <sector><p>Unset</p></sector>
      <sector><p class="curved">Nonzero</p></sector>
      <p class="horizontal" angle="0">Zero wins</p>
    </div>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	radial := doc.Root().Page(0).children[0].(*StdContainer)
	unset := radial.children[0].(*StdSector).children[0].(*StdParagraph)
	nonzero := radial.children[1].(*StdSector).children[0].(*StdParagraph)
	horizontal := radial.children[2].(*StdSector).children[0].(*StdParagraph)
	if unset.angleSet || unset.sectorTextFacing() != sectorFacingAuto || !unset.curvedInSector() {
		t.Fatalf("unset paragraph inherited sector defaults: angle=%v facing=%v curved=%v",
			unset.angleSet, unset.sectorTextFacing(), unset.curvedInSector())
	}
	if !nonzero.angleSet || nonzero.angle != 35 || !nonzero.curvedInSector() || nonzero.sectorTextFacing() != sectorFacingUpright {
		t.Fatalf("styled paragraph attrs = angle %v/%v curved=%v facing=%v",
			nonzero.angleSet, nonzero.angle, nonzero.curvedInSector(), nonzero.sectorTextFacing())
	}
	if !horizontal.angleSet || horizontal.angle != 0 || horizontal.curvedInSector() {
		t.Fatalf("direct paragraph angle = %v/%v curved=%v, want local zero",
			horizontal.angleSet, horizontal.angle, horizontal.curvedInSector())
	}

	invalid := &StdParagraph{}
	invalid.SetAttrs(map[string]string{"angle": "not-an-angle"})
	if invalid.angleSet {
		t.Fatal("invalid paragraph angle unexpectedly selected a mode")
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

func TestLayoutRadialTable_R1SetsOuterRadiusAndInfersDimensions(t *testing.T) {
	for _, manager := range []string{"radial", "radial-out"} {
		t.Run(manager, func(t *testing.T) {
			container := &StdContainer{}
			container.SetScope(&defaultScope)
			container.SetAttrs(map[string]string{
				"layout": manager,
				"rows":   "1",
				"cols":   "1",
				"r0":     "20",
				"r1":     "60",
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
			if got, want := container.Width(), 120.0; !floatEquals(got, want) {
				t.Fatalf("inferred width = %v, want %v", got, want)
			}
			if got, want := container.Height(), 120.0; !floatEquals(got, want) {
				t.Fatalf("inferred height = %v, want %v", got, want)
			}
			if got, want := sector.geometry.CenterX, 60.0; !floatEquals(got, want) {
				t.Fatalf("center x = %v, want %v", got, want)
			}
			if got, want := sector.geometry.CenterY, 60.0; !floatEquals(got, want) {
				t.Fatalf("center y = %v, want %v", got, want)
			}
			if got, want := sector.geometry.InnerRadius, 20.0; !floatEquals(got, want) {
				t.Fatalf("inner radius = %v, want %v", got, want)
			}
			if got, want := sector.geometry.OuterRadius, 60.0; !floatEquals(got, want) {
				t.Fatalf("outer radius = %v, want %v", got, want)
			}
		})
	}
}

func TestLayoutPositionedRadialCenter_AbsoluteUsesPhysicalPage(t *testing.T) {
	for _, manager := range []string{"radial", "radial-out"} {
		t.Run(manager, func(t *testing.T) {
			page := &StdPage{pageStyle: &PageStyle{width: 400, height: 300}}
			page.SetScope(&defaultScope)
			page.SetAttrs(map[string]string{
				"layout":  "absolute",
				"margin":  "30pt",
				"padding": "10pt",
			})

			radial := &StdContainer{}
			radial.SetScope(&defaultScope)
			if err := radial.SetContainer(page); err != nil {
				t.Fatal(err)
			}
			radial.SetAttrs(map[string]string{
				"layout":        manager,
				"position":      "absolute",
				"rows":          "1",
				"cols":          "1",
				"center-x":      "50%",
				"center-y":      "25%",
				"r0":            "10pt",
				"r1":            "40pt",
				"margin-left":   "3pt",
				"margin-top":    "5pt",
				"padding-left":  "7pt",
				"padding-right": "11pt",
				"padding-top":   "13pt",
			})
			page.AddChild(radial)

			sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
			sector.font = testSectorFont()
			if err := sector.SetContainer(radial); err != nil {
				t.Fatal(err)
			}
			radial.AddChild(sector)

			if err := LayoutAbsolute(page, page.LayoutStyle(), &labelTestWriter{t: t}); err != nil {
				t.Fatal(err)
			}
			if got, want := radial.Width(), 101.0; !floatEquals(got, want) {
				t.Fatalf("inferred width = %v, want %v", got, want)
			}
			if got, want := sector.geometry.CenterX, 200.0; !floatEquals(got, want) {
				t.Fatalf("absolute center x = %v, want %v", got, want)
			}
			if got, want := sector.geometry.CenterY, 75.0; !floatEquals(got, want) {
				t.Fatalf("absolute center y = %v, want %v", got, want)
			}
		})
	}
}

func TestLayoutPositionedRadialCenter_RelativeUsesParentBorderBox(t *testing.T) {
	parent := positionedContainer(50, 40, 300, 200)
	parent.SetScope(&defaultScope)
	parent.SetAttrs(map[string]string{
		"layout":        "relative",
		"margin":        "17pt",
		"padding-left":  "23pt",
		"padding-right": "29pt",
		"padding-top":   "31pt",
	})

	radial := &StdContainer{}
	radial.SetScope(&defaultScope)
	if err := radial.SetContainer(parent); err != nil {
		t.Fatal(err)
	}
	radial.SetAttrs(map[string]string{
		"layout":   "radial",
		"rows":     "1",
		"cols":     "1",
		"center-x": "25%",
		"center-y": "75%",
		"r1":       "30pt",
	})
	parent.AddChild(radial)

	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	if err := sector.SetContainer(radial); err != nil {
		t.Fatal(err)
	}
	radial.AddChild(sector)

	if radial.Position() != Relative {
		t.Fatalf("center attrs implied position = %v, want relative", radial.Position())
	}
	if err := LayoutRelative(parent, parent.LayoutStyle(), &labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
	if got, want := sector.geometry.CenterX, 125.0; !floatEquals(got, want) {
		t.Fatalf("relative center x = %v, want %v", got, want)
	}
	if got, want := sector.geometry.CenterY, 190.0; !floatEquals(got, want) {
		t.Fatalf("relative center y = %v, want %v", got, want)
	}
}

func TestLayoutPositionedRadialCenter_OmittedAxisDefaultsToMidpointAndShiftIsFinal(t *testing.T) {
	parent := positionedContainer(20, 30, 200, 100)
	parent.SetScope(&defaultScope)
	parent.SetAttrs(map[string]string{"layout": "relative"})

	radial := &StdContainer{}
	radial.SetScope(&defaultScope)
	if err := radial.SetContainer(parent); err != nil {
		t.Fatal(err)
	}
	radial.SetAttrs(map[string]string{
		"layout":   "radial",
		"rows":     "1",
		"cols":     "1",
		"center-x": "125%",
		"r1":       "20pt",
		"left":     "1pt",
		"right":    "2pt",
		"top":      "3pt",
		"bottom":   "4pt",
		"shift-x":  "6pt",
		"shift-y":  "-7pt",
	})
	parent.AddChild(radial)

	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	if err := sector.SetContainer(radial); err != nil {
		t.Fatal(err)
	}
	radial.AddChild(sector)

	if err := LayoutRelative(parent, parent.LayoutStyle(), &labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
	if got, want := sector.geometry.CenterX, 276.0; !floatEquals(got, want) {
		t.Fatalf("shifted center x = %v, want %v", got, want)
	}
	if got, want := sector.geometry.CenterY, 73.0; !floatEquals(got, want) {
		t.Fatalf("midpoint/shift center y = %v, want %v", got, want)
	}
}

func TestStdContainer_RadialCenterPositionCascade(t *testing.T) {
	container := &StdContainer{}
	container.SetAttrs(map[string]string{
		"layout":   "radial",
		"center-x": "50%",
		"position": "static",
	})
	if container.Position() != Static {
		t.Fatalf("same-layer explicit position = %v, want static", container.Position())
	}

	container.SetAttrs(map[string]string{"center-y": "25%"})
	if container.Position() != Relative {
		t.Fatalf("later center position = %v, want relative", container.Position())
	}

	container.SetAttrs(map[string]string{"position": "absolute"})
	if container.Position() != Absolute {
		t.Fatalf("later explicit position = %v, want absolute", container.Position())
	}
}

func TestStdContainer_RadialCenterModesFollowCascade(t *testing.T) {
	container := &StdContainer{}
	container.SetAttrs(map[string]string{"units": "pt", "center-x": "12.5%"})
	if container.centerXMode != DimPct || !floatEquals(container.centerX, 12.5) {
		t.Fatalf("percentage center x = mode %v value %v, want pct 12.5", container.centerXMode, container.centerX)
	}
	container.SetAttrs(map[string]string{"center-x": "72pt"})
	if container.centerXMode != DimLiteral || !floatEquals(container.centerX, 72) {
		t.Fatalf("literal center x = mode %v value %v, want literal 72", container.centerXMode, container.centerX)
	}
	container.SetAttrs(map[string]string{"center-x": "125%"})
	if container.centerXMode != DimPct || !floatEquals(container.centerX, 125) {
		t.Fatalf("later percentage center x = mode %v value %v, want pct 125", container.centerXMode, container.centerX)
	}

	for _, tc := range []struct {
		value float64
		mode  DimensionMode
		want  float64
	}{
		{value: 0, mode: DimPct, want: 10},
		{value: 50, mode: DimPct, want: 110},
		{value: 100, mode: DimPct, want: 210},
		{value: 12.5, mode: DimPct, want: 35},
		{value: 40, mode: DimLiteral, want: 50},
	} {
		if got := resolveRadialCenterAxis(tc.value, tc.mode, 10, 200); !floatEquals(got, tc.want) {
			t.Errorf("resolved center (%v,%v) = %v, want %v", tc.value, tc.mode, got, tc.want)
		}
	}
}

func TestLayoutRadialTable_StaticCenterAttrsAreDormant(t *testing.T) {
	container := positionedContainer(10, 20, 120, 80)
	container.SetScope(&defaultScope)
	container.SetAttrs(map[string]string{
		"layout":   "radial",
		"position": "static",
		"rows":     "1",
		"cols":     "1",
		"center-x": "0%",
		"center-y": "100%",
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
	if got, want := sector.geometry.CenterX, 70.0; !floatEquals(got, want) {
		t.Fatalf("static center x = %v, want box midpoint %v", got, want)
	}
	if got, want := sector.geometry.CenterY, 60.0; !floatEquals(got, want) {
		t.Fatalf("static center y = %v, want box midpoint %v", got, want)
	}
}

func TestStdContainer_R1AndRAliasFollowCascadePrecedence(t *testing.T) {
	container := &StdContainer{}
	container.SetAttrs(map[string]string{"units": "pt", "r": "40", "r1": "60"})
	if got, want := container.OuterRadius(), 60.0; !floatEquals(got, want) {
		t.Fatalf("same-layer outer radius = %v, want r1 value %v", got, want)
	}

	container.SetAttrs(map[string]string{"r": "70"})
	if got, want := container.OuterRadius(), 70.0; !floatEquals(got, want) {
		t.Fatalf("later r outer radius = %v, want %v", got, want)
	}

	container.SetAttrs(map[string]string{"r1": "80"})
	if got, want := container.OuterRadius(), 80.0; !floatEquals(got, want) {
		t.Fatalf("later r1 outer radius = %v, want %v", got, want)
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

func positionedSectorTestFixture(t *testing.T, start, end float64) *StdSector {
	t.Helper()
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	anchorAngle := start + (end-start)/2
	ax, ay := radialPointAt(100, 100, 35, anchorAngle)
	sector.setGeometry(radialSectorGeometry{
		CenterX:     100,
		CenterY:     100,
		InnerRadius: 20,
		OuterRadius: 50,
		StartAngle:  start,
		EndAngle:    end,
		AnchorAngle: anchorAngle,
		AnchorX:     ax,
		AnchorY:     ay,
	})
	return sector
}

func TestStdSector_RadialAttrsImplyRelativeAndRectangularSidesStayDormant(t *testing.T) {
	sector := positionedSectorTestFixture(t, 0, 90)

	radial := &StdWidget{}
	if err := radial.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	radial.SetAttrs(map[string]string{"units": "pt", "start": "4"})
	if got := radial.Position(); got != Relative {
		t.Fatalf("radial position = %v, want relative", got)
	}

	rectangular := &StdWidget{}
	if err := rectangular.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	rectangular.SetAttrs(map[string]string{"units": "pt", "left": "4", "top": "5"})
	if got := rectangular.Position(); got != Static {
		t.Fatalf("rectangular-side position = %v, want static in sector", got)
	}

}

func TestStdSector_ExplicitPositionOverridesRadialInference(t *testing.T) {
	sector := positionedSectorTestFixture(t, 0, 90)
	for _, tt := range []struct {
		value string
		want  Position
	}{
		{value: "static", want: Static},
		{value: "absolute", want: Absolute},
		{value: "relative", want: Relative},
	} {
		t.Run(tt.value, func(t *testing.T) {
			child := &StdWidget{}
			if err := child.SetContainer(sector); err != nil {
				t.Fatal(err)
			}
			child.SetAttrs(map[string]string{
				"position": tt.value,
				"start":    "4pt",
				"outer":    "5pt",
			})
			if got := child.Position(); got != tt.want {
				t.Fatalf("position = %v, want %v", got, tt.want)
			}
		})
	}

}

func TestStdSector_PositionAndRadialInferenceFollowCascadeLayers(t *testing.T) {
	sector := positionedSectorTestFixture(t, 0, 90)
	child := &StdWidget{}
	if err := child.SetContainer(sector); err != nil {
		t.Fatal(err)
	}

	child.SetAttrs(map[string]string{"position": "static"})
	child.SetAttrs(map[string]string{"start": "4pt"})
	if got := child.Position(); got != Relative {
		t.Fatalf("later radial position = %v, want relative", got)
	}

	child.SetAttrs(map[string]string{"position": "absolute"})
	child.SetAttrs(map[string]string{"font.size": "9pt"})
	if got := child.Position(); got != Absolute {
		t.Fatalf("unrelated later layer position = %v, want absolute", got)
	}

	child.SetAttrs(map[string]string{"outer": "5pt"})
	if got := child.Position(); got != Relative {
		t.Fatalf("later radial position after absolute = %v, want relative", got)
	}

	child.SetAttrs(map[string]string{"position": "static", "end": "6pt"})
	if got := child.Position(); got != Static {
		t.Fatalf("same-layer explicit position = %v, want static", got)
	}
}

func TestStdSector_RadialEdgeCascadeAndSameLayerPrecedence(t *testing.T) {
	sector := positionedSectorTestFixture(t, 0, 90)
	child := &StdWidget{}
	if err := child.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	child.SetAttrs(map[string]string{"units": "pt", "start": "4", "outer": "6"})
	child.SetAttrs(map[string]string{"units": "pt", "end": "9", "inner": "8"})
	got := sector.resolvePositionedReference(child)
	wantRadius := 28.0
	wantAngle := 90 - math.Asin(9.0/wantRadius)*180/math.Pi
	if math.Abs(got.radius-wantRadius) > 0.001 || math.Abs(got.angle-wantAngle) > 0.001 {
		t.Fatalf("later-layer reference = %v/%v, want %v/%v", got.radius, got.angle, wantRadius, wantAngle)
	}

	child.SetAttrs(map[string]string{
		"units": "pt",
		"start": "3",
		"end":   "7",
		"outer": "5",
		"inner": "9",
	})
	got = sector.resolvePositionedReference(child)
	wantRadius = 45
	wantAngle = math.Asin(3.0/wantRadius) * 180 / math.Pi
	if math.Abs(got.radius-wantRadius) > 0.001 || math.Abs(got.angle-wantAngle) > 0.001 {
		t.Fatalf("same-layer reference = %v/%v, want start/outer %v/%v", got.radius, got.angle, wantRadius, wantAngle)
	}
}

func TestStdSector_PositionedRadialEdgesSelectReference(t *testing.T) {
	tests := []struct {
		name       string
		start, end float64
		attrs      map[string]string
		wantRadius float64
		wantAngle  float64
	}{
		{
			name: "ccw start outer", start: 0, end: 90,
			attrs:      map[string]string{"start": "0", "outer": "0"},
			wantRadius: 50, wantAngle: 0,
		},
		{
			name: "ccw end inner", start: 0, end: 90,
			attrs:      map[string]string{"end": "0", "inner": "0"},
			wantRadius: 20, wantAngle: 90,
		},
		{
			name: "cw start outer", start: 90, end: 0,
			attrs:      map[string]string{"start": "0", "outer": "0"},
			wantRadius: 50, wantAngle: 90,
		},
		{
			name: "cw end inner", start: 90, end: 0,
			attrs:      map[string]string{"end": "0", "inner": "0"},
			wantRadius: 20, wantAngle: 0,
		},
		{
			name: "lower quadrant", start: 180, end: 270,
			attrs:      map[string]string{"start": "0", "inner": "0"},
			wantRadius: 20, wantAngle: 180,
		},
		{
			name: "crosses lower right quadrant", start: -90, end: 0,
			attrs:      map[string]string{"end": "0", "outer": "0"},
			wantRadius: 50, wantAngle: 0,
		},
		{
			name: "full circle start inset", start: 0, end: 360,
			attrs:      map[string]string{"start": "4", "outer": "0"},
			wantRadius: 50, wantAngle: math.Asin(4.0/50.0) * 180 / math.Pi,
		},
		{
			name: "full circle end inset", start: 0, end: 360,
			attrs:      map[string]string{"end": "4", "inner": "0"},
			wantRadius: 20, wantAngle: 360 - math.Asin(4.0/20.0)*180/math.Pi,
		},
		{
			name: "omitted axes use midpoint", start: 0, end: 90,
			attrs:      map[string]string{},
			wantRadius: 35, wantAngle: 45,
		},
		{
			name: "start and outer win in one layer", start: 0, end: 90,
			attrs:      map[string]string{"start": "4", "end": "9", "outer": "6", "inner": "8"},
			wantRadius: 44, wantAngle: math.Asin(4.0/44.0) * 180 / math.Pi,
		},
		{
			name: "negative offsets cross edges", start: 0, end: 90,
			attrs:      map[string]string{"start": "-4", "outer": "-6"},
			wantRadius: 56, wantAngle: math.Asin(-4.0/56.0) * 180 / math.Pi,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sector := positionedSectorTestFixture(t, tt.start, tt.end)
			child := &StdWidget{}
			if err := child.SetContainer(sector); err != nil {
				t.Fatal(err)
			}
			attrs := maps.Clone(tt.attrs)
			attrs["position"] = "relative"
			attrs["units"] = "pt"
			child.SetAttrs(attrs)

			got := sector.resolvePositionedReference(child)
			if math.Abs(got.radius-tt.wantRadius) > 0.001 || math.Abs(got.angle-tt.wantAngle) > 0.001 {
				t.Fatalf("reference radius/angle = %v/%v, want %v/%v", got.radius, got.angle, tt.wantRadius, tt.wantAngle)
			}
		})
	}
}

func TestStdSector_RectangularSidesAreDormantForRelativeAndPageAxisForAbsolute(t *testing.T) {
	sector := positionedSectorTestFixture(t, 10, 100)

	relative := &StdWidget{}
	if err := relative.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	relative.SetAttrs(map[string]string{
		"position": "relative",
		"units":    "pt",
		"left":     "7",
		"right":    "8",
		"top":      "9",
		"bottom":   "10",
	})
	got := sector.resolvePositionedReference(relative)
	if math.Abs(got.radius-35) > 0.001 || math.Abs(got.angle-55) > 0.001 {
		t.Fatalf("relative reference = %v/%v, want midpoint 35/55", got.radius, got.angle)
	}
	if relative.Width() != 0 || relative.Height() != 0 {
		t.Fatalf("relative size from dormant sides = %v/%v, want 0/0", relative.Width(), relative.Height())
	}

	absolute := &StdWidget{}
	if err := absolute.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	absolute.SetAttrs(map[string]string{
		"position": "absolute",
		"units":    "pt",
		"left":     "12",
		"top":      "14",
		"width":    "20",
		"height":   "10",
		"start":    "3",
		"outer":    "4",
	})
	if absolute.Left() != 12 || absolute.Top() != 14 {
		t.Fatalf("absolute page point = (%v,%v), want (12,14)", absolute.Left(), absolute.Top())
	}
}

func TestStdSector_PositionedReferenceIncludesSectorPaddingAndChildInset(t *testing.T) {
	sector := positionedSectorTestFixture(t, 0, 90)
	sector.SetAttrs(map[string]string{
		"units":         "pt",
		"padding-top":   "3",
		"padding-left":  "2",
		"padding-right": "7",
	})
	sector.rebuildContentGeometry()
	child := &StdWidget{}
	if err := child.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	child.SetAttrs(map[string]string{"position": "relative", "units": "pt", "start": "5", "outer": "4"})

	got := sector.resolvePositionedReference(child)
	wantRadius := 43.0
	wantAngle := math.Asin((2.0+5.0)/wantRadius) * 180 / math.Pi
	if math.Abs(got.radius-wantRadius) > 0.001 || math.Abs(got.angle-wantAngle) > 0.001 {
		t.Fatalf("padded reference radius/angle = %v/%v, want %v/%v", got.radius, got.angle, wantRadius, wantAngle)
	}
}

func TestStdSector_PositionedOriginsAttachChildBox(t *testing.T) {
	for _, x := range []struct {
		name   string
		factor float64
	}{
		{name: "start", factor: 0},
		{name: "center", factor: 0.5},
		{name: "end", factor: 1},
	} {
		for _, y := range []struct {
			name   string
			factor float64
		}{
			{name: "top", factor: 0},
			{name: "middle", factor: 0.5},
			{name: "bottom", factor: 1},
		} {
			t.Run(x.name+"/"+y.name, func(t *testing.T) {
				sector := positionedSectorTestFixture(t, 0, 90)
				child := &StdWidget{}
				if err := child.SetContainer(sector); err != nil {
					t.Fatal(err)
				}
				child.SetAttrs(map[string]string{
					"position": "relative",
					"units":    "pt",
					"start":    "5",
					"outer":    "4",
					"origin-x": x.name,
					"origin-y": y.name,
					"width":    "20",
					"height":   "10",
				})
				sector.preparePositionedChildren()

				placement := sector.ResolveSectorPlacement(child)
				if math.Abs(placement.boxLeft-(placement.anchorX-20*x.factor)) > 0.001 ||
					math.Abs(placement.boxTop-(placement.anchorY-10*y.factor)) > 0.001 {
					t.Fatalf("box/anchor = (%v,%v)/(%v,%v), want factors %v/%v",
						placement.boxLeft, placement.boxTop, placement.anchorX, placement.anchorY,
						x.factor, y.factor)
				}
				if math.Abs(child.OriginXValue()-placement.anchorX) > 0.001 ||
					math.Abs(child.OriginYValue()-placement.anchorY) > 0.001 {
					t.Fatalf("widget origin = (%v,%v), want placement anchor (%v,%v)",
						child.OriginXValue(), child.OriginYValue(), placement.anchorX, placement.anchorY)
				}
			})
		}
	}
}

func TestStdSector_PositionedCustomOriginsRetainTransformCoordinates(t *testing.T) {
	sector := positionedSectorTestFixture(t, 0, 90)
	child := &StdWidget{}
	if err := child.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	child.SetAttrs(map[string]string{
		"position": "relative",
		"units":    "pt",
		"origin-x": "12",
		"origin-y": "8",
		"width":    "20",
		"height":   "10",
	})
	if got := child.OriginXValue(); got != 12 {
		t.Fatalf("custom origin x = %v, want 12", got)
	}
	if got := child.OriginYValue(); got != 8 {
		t.Fatalf("custom origin y = %v, want 8", got)
	}
}

func TestStdSector_PositionedOpposingRadialEdgesDoNotStretchChild(t *testing.T) {
	sector := positionedSectorTestFixture(t, 0, 90)
	child := &StdWidget{}
	if err := child.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	child.SetAttrs(map[string]string{
		"position": "relative",
		"units":    "pt",
		"start":    "3",
		"end":      "4",
		"outer":    "5",
		"inner":    "6",
	})
	if got := child.Width(); got != 0 {
		t.Fatalf("width from opposing sector sides = %v, want no stretch", got)
	}
	if got := child.Height(); got != 0 {
		t.Fatalf("height from opposing sector sides = %v, want no stretch", got)
	}
}

func TestStdSector_PositionedChildrenShareRadialReferenceAndPageAxisShift(t *testing.T) {
	sector := positionedSectorTestFixture(t, 20, 110)
	image := &StdImage{}
	label := &StdLabel{}
	paragraph := &StdParagraph{}
	container := &StdContainer{}
	children := []*StdWidget{
		{},
		&image.StdWidget,
		&label.StdWidget,
		&paragraph.StdWidget,
		&container.StdWidget,
	}
	var want sectorPositionedReference
	for i, child := range children {
		if err := child.SetContainer(sector); err != nil {
			t.Fatal(err)
		}
		child.SetAttrs(map[string]string{
			"position": "relative",
			"units":    "pt",
			"end":      "7",
			"inner":    "3",
			"shift-x":  "4",
			"shift-y":  "-6",
		})
		got := sector.resolvePositionedReference(child)
		if i == 0 {
			want = got
			continue
		}
		if math.Abs(got.pageX-want.pageX) > 0.001 || math.Abs(got.pageY-want.pageY) > 0.001 ||
			math.Abs(got.radius-want.radius) > 0.001 || math.Abs(got.angle-want.angle) > 0.001 {
			t.Fatalf("child %d reference = %+v, want %+v", i, got, want)
		}
	}

	unshiftedX, unshiftedY := radialPointAt(sector.geometry.CenterX, sector.geometry.CenterY, want.radius, want.angle)
	if math.Abs(want.pageX-(unshiftedX+4)) > 0.001 || math.Abs(want.pageY-(unshiftedY-6)) > 0.001 {
		t.Fatalf("shifted page point = (%v,%v), want (%v,%v)", want.pageX, want.pageY, unshiftedX+4, unshiftedY-6)
	}
	rotatedX, rotatedY := rotatePagePoint(want.localX, want.localY,
		sector.geometry.AnchorX, sector.geometry.AnchorY, sector.contentRotation)
	if math.Abs(rotatedX-want.pageX) > 0.001 || math.Abs(rotatedY-want.pageY) > 0.001 {
		t.Fatalf("local reference rotates to (%v,%v), want page point (%v,%v)",
			rotatedX, rotatedY, want.pageX, want.pageY)
	}
}

func TestStdSector_PositionedFacingDoesNotMoveReference(t *testing.T) {
	sector := positionedSectorTestFixture(t, 190, 280)
	var want sectorPositionedReference
	for i, facing := range []string{"upright", "upside-down"} {
		label := &StdLabel{}
		if err := label.SetContainer(sector); err != nil {
			t.Fatal(err)
		}
		label.SetAttrs(map[string]string{
			"position": "relative",
			"units":    "pt",
			"start":    "6",
			"outer":    "4",
			"facing":   facing,
		})
		got := sector.resolvePositionedReference(&label.StdWidget)
		if i == 0 {
			want = got
			continue
		}
		if math.Abs(got.pageX-want.pageX) > 0.001 || math.Abs(got.pageY-want.pageY) > 0.001 {
			t.Fatalf("%s reference = (%v,%v), want (%v,%v)", facing, got.pageX, got.pageY, want.pageX, want.pageY)
		}
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
	addTestSectorLabel(t, sector, "Radial", map[string]string{"position": "relative", "end": "0", "text-align": "right"})
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
		{"position": "relative", "start": "0", "inner": "0", "text-align": "left"},
		{"position": "relative", "text-align": "center"},
		{"position": "relative", "end": "0", "outer": "0", "text-align": "right"},
	}
	for i, attr := range attrs {
		label := &StdLabel{}
		label.font = testSectorFont()
		if err := label.SetContainer(sector); err != nil {
			t.Fatal(err)
		}
		label.SetAttrs(attr)
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

func TestStdSector_LabelOffsetsUseSectorEdges(t *testing.T) {
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
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.SetAttrs(map[string]string{
		"units": "pt",
		"start": "5",
		"outer": "-3",
	})
	label.AddText("offset")
	sector.AddChild(label)
	if err := sector.LayoutWidget(&labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
	wantRadius := 63.0
	wantAngle := math.Asin(5/wantRadius) * 180 / math.Pi
	wantX, wantY := radialPointAt(100, 100, wantRadius, wantAngle)
	layout := sector.cachedLabelLayout(label)
	if !floatEquals(layout.anchorX, wantX) || !floatEquals(layout.anchorY, wantY) {
		t.Fatalf("offset anchor = (%v,%v), want (%v,%v)", layout.anchorX, layout.anchorY, wantX, wantY)
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
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.SetAttrs(map[string]string{
		"position":    "relative",
		"start":       "0",
		"text-align":  "center",
		"facing":      "upside-down",
		"text-valign": "bottom",
	})
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
	if got := w.curvedOpts[0].Direction; got != pdf.CurvedTextCounterClockwise {
		t.Fatalf("curved direction = %v, want counter-clockwise for upside-down facing", got)
	}
	if got := w.curvedOpts[0].VAlign; got != pdf.VTextAlignBelow {
		t.Fatalf("curved valign = %v, want below", got)
	}
}

func TestStdSector_ExplicitLabelFacingPreservesReadingDirectionForLTRAndRTL(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		text string
	}{
		{name: "left to right", dir: "ltr", text: "Facing override"},
		{name: "right to left", dir: "rtl", text: "مرحبا"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
			sector.font = testSectorFont()
			ax, ay := radialPointAt(100, 100, 40, 270)
			sector.setGeometry(radialSectorGeometry{
				CenterX: 100, CenterY: 100,
				InnerRadius: 20, OuterRadius: 60,
				StartAngle: 225, EndAngle: 315,
				AnchorAngle: 270, AnchorX: ax, AnchorY: ay,
			})
			label := addTestSectorLabel(t, sector, tt.text, map[string]string{
				"position": "relative",
				"facing":   "upright",
				"dir":      tt.dir,
			})
			if got := label.Dir(); got != ParseDir(tt.dir) {
				t.Fatalf("label direction = %v, want %v", got, ParseDir(tt.dir))
			}
			if got := label.AccessibilityText(); got != tt.text {
				t.Fatalf("logical label text = %q, want %q", got, tt.text)
			}

			w := &labelTestWriter{t: t}
			if err := sector.LayoutWidget(w); err != nil {
				t.Fatal(err)
			}
			if err := sector.DrawContent(w); err != nil {
				t.Fatal(err)
			}
			if len(w.curvedOpts) != 1 {
				t.Fatalf("curved draws = %d, want 1", len(w.curvedOpts))
			}
			if got := w.curvedOpts[0]; got.Direction != pdf.CurvedTextClockwise || got.Facing != pdf.CurvedTextFacingUpright {
				t.Fatalf("orientation = direction %v facing %v, want clockwise/upright", got.Direction, got.Facing)
			}
		})
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

	for _, attrs := range []map[string]string{
		{"position": "relative", "start": "0"},
		{"position": "relative", "end": "0"},
	} {
		label := &StdLabel{}
		label.font = testSectorFont()
		if err := label.SetContainer(sector); err != nil {
			t.Fatal(err)
		}
		label.SetAttrs(attrs)
		label.AddText("anchor")
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
	originalX := sector.cachedLabelLayout(label).anchorX

	label.SetAttrs(map[string]string{"units": "pt", "start": "12"})
	if sector.cachedLabelLayout(label) != nil {
		t.Fatal("SetAttrs did not invalidate sector placement")
	}
	if err := label.paintWithTransform(w, func() error {
		if sector.cachedLabelLayout(label) == nil {
			t.Fatal("placement was not restored before paint phases")
		}
		return label.PaintBackground(w)
	}); err != nil {
		t.Fatal(err)
	}
	if got := sector.cachedLabelLayout(label).anchorX; floatEquals(got, originalX) {
		t.Fatalf("recomputed anchor x = %v, want offset from %v", got, originalX)
	}
	if len(w.fillRectPages) != 1 {
		t.Fatalf("background fills = %d, want one after lazy placement", len(w.fillRectPages))
	}
}

func TestStdSector_TextAttributeChangesInvalidateOnlyChangedChild(t *testing.T) {
	sector := &StdSector{}
	firstLabel := &StdLabel{}
	secondLabel := &StdLabel{}
	paragraph := &StdParagraph{}
	for _, child := range []WantsContainer{firstLabel, secondLabel, paragraph} {
		if err := child.SetContainer(sector); err != nil {
			t.Fatal(err)
		}
	}

	firstLayout := &sectorLabelLayout{}
	secondLayout := &sectorLabelLayout{}
	paragraphLayout := &sectorParagraphLayout{}
	sector.labelLayouts = map[*StdLabel]*sectorLabelLayout{
		firstLabel:  firstLayout,
		secondLabel: secondLayout,
	}
	sector.paragraphLayouts = map[*StdParagraph]*sectorParagraphLayout{
		paragraph: paragraphLayout,
	}

	firstLabel.SetAttrs(map[string]string{"angle": "0"})
	if sector.cachedLabelLayout(firstLabel) != nil {
		t.Fatal("changed label layout was not invalidated")
	}
	if sector.cachedLabelLayout(secondLabel) != secondLayout {
		t.Fatal("sibling label layout was invalidated")
	}
	if sector.paragraphLayouts[paragraph] != paragraphLayout {
		t.Fatal("paragraph layout was invalidated by a label change")
	}

	sector.setLabelLayout(firstLabel, firstLayout)
	paragraph.SetAttrs(map[string]string{"angle": "0"})
	if sector.paragraphLayouts[paragraph] != nil {
		t.Fatal("changed paragraph layout was not invalidated")
	}
	if sector.cachedLabelLayout(firstLabel) != firstLayout ||
		sector.cachedLabelLayout(secondLabel) != secondLayout {
		t.Fatal("label layout was invalidated by a paragraph change")
	}
}

func TestStdSector_GeometryChangeInvalidatesAllTextLayouts(t *testing.T) {
	sector := &StdSector{
		labelLayouts: map[*StdLabel]*sectorLabelLayout{
			{}: {},
		},
		paragraphLayouts: map[*StdParagraph]*sectorParagraphLayout{
			{}: {},
		},
	}
	sector.setGeometry(radialSectorGeometry{
		CenterX: 100, CenterY: 100,
		InnerRadius: 20, OuterRadius: 60,
		StartAngle: 0, EndAngle: 90,
	})
	if sector.labelLayouts != nil || sector.paragraphLayouts != nil {
		t.Fatal("geometry change did not invalidate all text layouts")
	}
}

func TestStdLabelAndParagraph_OutsideSectorDoNotUseSectorLayouts(t *testing.T) {
	container := &StdContainer{}
	label := &StdLabel{}
	paragraph := &StdParagraph{}
	if err := label.SetContainer(container); err != nil {
		t.Fatal(err)
	}
	if err := paragraph.SetContainer(container); err != nil {
		t.Fatal(err)
	}

	label.SetAttrs(map[string]string{"angle": "0"})
	paragraph.SetAttrs(map[string]string{"angle": "0"})
	if label.cachedSectorLayout() != nil {
		t.Fatal("ordinary label unexpectedly has a sector layout")
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
	if err := label.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	label.SetAttrs(map[string]string{"position": "relative", "inner": "0"})
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
	p.SetAttrs(map[string]string{"angle": "0"})
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
	p.SetAttrs(map[string]string{"angle": "0"})
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
		"position": "relative", "end": "0", "outer": "0",
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
	if sector.cachedLabelLayout(overlay) == nil {
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
	firstLayout := sector.cachedLabelLayout(first)
	secondLayout := sector.cachedLabelLayout(second)
	if w.curvedCount != 2 || floatEquals(firstLayout.angle, secondLayout.angle) {
		t.Fatalf("curved flow draws/angles = %d, %v/%v", w.curvedCount, firstLayout.angle, secondLayout.angle)
	}
}

func TestStdSector_WithParagraphChild_DoesNotDefaultToTangentRotation(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	p := &StdParagraph{}
	p.paragraphStyle = &ParagraphStyle{}
	p.SetAttrs(map[string]string{"angle": "0"})
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
	curved, horizontal := 0, 0
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
		if paragraph.curvedInSector() {
			curved++
		} else {
			horizontal++
		}
	}
	if curved != 1 || horizontal != 1 {
		t.Fatalf("implicit paragraph modes = %d curved, %d horizontal; want one of each", curved, horizontal)
	}
}

func addTestSectorParagraph(t *testing.T, sector *StdSector, text string, attrs map[string]string) *StdParagraph {
	t.Helper()
	paragraph := &StdParagraph{}
	paragraph.paragraphStyle = &ParagraphStyle{}
	if err := paragraph.SetContainer(sector); err != nil {
		t.Fatal(err)
	}
	paragraph.SetAttrs(attrs)
	paragraph.AddText(text)
	sector.AddChild(paragraph)
	return paragraph
}

func curvedParagraphTestSector(t *testing.T, anchorAngle float64) *StdSector {
	t.Helper()
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	ax, ay := radialPointAt(120, 120, 75, anchorAngle)
	sector.setGeometry(radialSectorGeometry{
		CenterX: 120, CenterY: 120, InnerRadius: 20, OuterRadius: 130,
		StartAngle: anchorAngle - 50, EndAngle: anchorAngle + 50,
		AnchorAngle: anchorAngle, AnchorX: ax, AnchorY: ay,
	})
	return sector
}

func TestStdSector_ParagraphAngleSelectsCurvedOrHorizontalMode(t *testing.T) {
	tests := []struct {
		name   string
		angle  string
		curved bool
	}{
		{name: "unset", curved: true},
		{name: "zero", angle: "0", curved: false},
		{name: "decimal zero", angle: "0.0", curved: false},
		{name: "negative zero", angle: "-0", curved: false},
		{name: "nonzero dormant", angle: "45", curved: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sector := curvedParagraphTestSector(t, 90)
			attrs := map[string]string{}
			if tt.angle != "" {
				attrs["angle"] = tt.angle
			}
			paragraph := addTestSectorParagraph(t, sector, "Curved paragraph mode", attrs)
			if got := paragraph.curvedInSector(); got != tt.curved {
				t.Fatalf("curved mode = %v, want %v", got, tt.curved)
			}
			w := &labelTestWriter{t: t}
			if err := sector.LayoutWidget(w); err != nil {
				t.Fatal(err)
			}
			if err := sector.DrawContent(w); err != nil {
				t.Fatal(err)
			}
			if got := w.curvedCount > 0; got != tt.curved {
				t.Fatalf("curved draws = %d, curved mode %v", w.curvedCount, tt.curved)
			}
		})
	}
}

func TestStdSector_CurvedParagraphLineOrderFollowsFacing(t *testing.T) {
	tests := []struct {
		name          string
		anchorAngle   float64
		facing        string
		increasing    bool
		wantDirection pdf.CurvedTextDirection
		wantFacing    pdf.CurvedTextFacing
	}{
		{name: "top automatic", anchorAngle: 90, wantDirection: pdf.CurvedTextClockwise, wantFacing: pdf.CurvedTextFacingUpright},
		{name: "bottom automatic", anchorAngle: 270, increasing: true, wantDirection: pdf.CurvedTextCounterClockwise, wantFacing: pdf.CurvedTextFacingUpsideDown},
		{name: "top upside down", anchorAngle: 90, facing: "upside-down", increasing: true, wantDirection: pdf.CurvedTextCounterClockwise, wantFacing: pdf.CurvedTextFacingUpsideDown},
		{name: "bottom upright", anchorAngle: 270, facing: "upright", wantDirection: pdf.CurvedTextClockwise, wantFacing: pdf.CurvedTextFacingUpright},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sector := curvedParagraphTestSector(t, tt.anchorAngle)
			attrs := map[string]string{}
			if tt.facing != "" {
				attrs["facing"] = tt.facing
			}
			addTestSectorParagraph(t, sector,
				"One two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen sixteen", attrs)
			w := &labelTestWriter{t: t}
			if err := sector.LayoutWidget(w); err != nil {
				t.Fatal(err)
			}
			if err := sector.DrawContent(w); err != nil {
				t.Fatal(err)
			}
			if len(w.curvedRadii) < 2 {
				t.Fatalf("curved line count = %d, want at least 2", len(w.curvedRadii))
			}
			increasing := w.curvedRadii[1] > w.curvedRadii[0]
			if increasing != tt.increasing {
				t.Fatalf("first radii = %v, %v; increasing = %v, want %v", w.curvedRadii[0], w.curvedRadii[1], increasing, tt.increasing)
			}
			for i, opts := range w.curvedOpts {
				if opts.Direction != tt.wantDirection || opts.Facing != tt.wantFacing {
					t.Fatalf("line %d orientation = direction %v facing %v, want direction %v facing %v",
						i, opts.Direction, opts.Facing, tt.wantDirection, tt.wantFacing)
				}
			}
		})
	}
}

func TestStdSector_CurvedParagraphAlignmentUsesReadableArcEndpoints(t *testing.T) {
	tests := []struct {
		name       string
		startAngle float64
		endAngle   float64
		anchor     float64
		align      string
		rtl        bool
		wantEnd    bool
		wantCenter bool
	}{
		{name: "counterclockwise start", startAngle: 40, endAngle: 140, anchor: 90, align: "start"},
		{name: "counterclockwise end", startAngle: 40, endAngle: 140, anchor: 90, align: "end", wantEnd: true},
		{name: "counterclockwise rtl start", startAngle: 40, endAngle: 140, anchor: 90, align: "start", rtl: true, wantEnd: true},
		{name: "counterclockwise center", startAngle: 40, endAngle: 140, anchor: 90, align: "center", wantCenter: true},
		{name: "clockwise start", startAngle: 140, endAngle: 40, anchor: 90, align: "start"},
		{name: "clockwise end", startAngle: 140, endAngle: 40, anchor: 90, align: "end", wantEnd: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sector := curvedParagraphTestSector(t, tt.anchor)
			sector.geometry.StartAngle = tt.startAngle
			sector.geometry.EndAngle = tt.endAngle
			sector.rebuildContentGeometry()
			attrs := map[string]string{"style.text-align": tt.align}
			if tt.rtl {
				attrs["dir"] = "rtl"
			}
			paragraph := addTestSectorParagraph(t, sector, "Aligned curved paragraph", attrs)
			w := &labelTestWriter{t: t}
			if err := sector.LayoutWidget(w); err != nil {
				t.Fatal(err)
			}
			layout := sector.sectorParagraphLayoutFor(paragraph, w)
			if len(layout.curvedLines) == 0 {
				t.Fatal("curved paragraph has no line placements")
			}
			line := layout.curvedLines[0]
			pathStart, pathEnd := sector.curvedParagraphArcEndpoints(line.radius, layout.direction)
			want := pathStart
			if tt.wantEnd {
				want = pathEnd
			}
			if tt.wantCenter {
				want = tt.anchor
			}
			if !floatEquals(line.angle, want) {
				t.Fatalf("line angle = %v, want %v (path %v..%v)", line.angle, want, pathStart, pathEnd)
			}
		})
	}
}

func TestStdSector_CurvedParagraphJustifiesAllButFinalLine(t *testing.T) {
	sector := curvedParagraphTestSector(t, 90)
	paragraph := addTestSectorParagraph(t, sector,
		"one two three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen sixteen",
		map[string]string{"style.text-align": "justify"})
	w := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	layout := sector.sectorParagraphLayoutFor(paragraph, w)
	if len(layout.lines) < 2 {
		t.Fatalf("line count = %d, want at least two", len(layout.lines))
	}
	if err := sector.DrawContent(w); err != nil {
		t.Fatal(err)
	}
	if len(w.printed) != len(layout.lines) {
		t.Fatalf("printed lines = %d, want %d", len(w.printed), len(layout.lines))
	}
	if math.Abs(w.printed[0].Width()-layout.curvedLines[0].arcWidth) > 0.05 {
		t.Fatalf("first justified width = %v, want arc width %v", w.printed[0].Width(), layout.curvedLines[0].arcWidth)
	}
	last := len(layout.lines) - 1
	if !floatEquals(w.printed[last].Width(), layout.lines[last].Width()) {
		t.Fatalf("final width = %v, original %v; final line should not justify", w.printed[last].Width(), layout.lines[last].Width())
	}
}

func TestStdSector_CurvedParagraphUsesCompleteFullCircleArc(t *testing.T) {
	for _, span := range []float64{360, -360} {
		t.Run(strconv.FormatFloat(span, 'f', 0, 64), func(t *testing.T) {
			sector := curvedParagraphTestSector(t, 90)
			sector.geometry.StartAngle = 0
			sector.geometry.EndAngle = span
			sector.rebuildContentGeometry()
			paragraph := addTestSectorParagraph(t, sector, "Full circle curved paragraph", nil)
			w := &labelTestWriter{t: t}
			if err := sector.LayoutWidget(w); err != nil {
				t.Fatal(err)
			}
			layout := sector.sectorParagraphLayoutFor(paragraph, w)
			if len(layout.curvedLines) == 0 {
				t.Fatal("full-circle paragraph has no lines")
			}
			for i, line := range layout.curvedLines {
				want := 2 * math.Pi * line.radius
				if math.Abs(line.arcWidth-want) > 0.01 {
					t.Fatalf("line %d arc width = %v, want circumference %v", i, line.arcWidth, want)
				}
			}
		})
	}
}

func TestStdSector_CurvedParagraphBoxAttrsAreDormantUntilAngleZero(t *testing.T) {
	attrs := map[string]string{
		"units": "pt", "width": "80", "height": "40", "margin": "7", "padding": "6",
		"fill": "Gold", "border": "Blue", "text-fill": "Red", "rotate": "18",
	}
	curvedSector := curvedParagraphTestSector(t, 45)
	curved := addTestSectorParagraph(t, curvedSector, "Dormant curved paragraph box", attrs)
	curvedWriter := &labelTestWriter{t: t}
	if err := curvedSector.LayoutWidget(curvedWriter); err != nil {
		t.Fatal(err)
	}
	curvedLayout := curvedSector.sectorParagraphLayoutFor(curved, curvedWriter)
	if curved.Width() != 0 || !floatEquals(curved.Height(), curvedLayout.total) {
		t.Fatalf("curved box = %vx%v, want zero width and natural height %v", curved.Width(), curved.Height(), curvedLayout.total)
	}
	if err := curvedSector.DrawContent(curvedWriter); err != nil {
		t.Fatal(err)
	}
	if len(curvedWriter.rotations) != 0 || len(curvedWriter.fillRectPages) != 0 || len(curvedWriter.clipped) != 0 {
		t.Fatalf("curved dormant paints rotations/fills/text clips = %d/%d/%d",
			len(curvedWriter.rotations), len(curvedWriter.fillRectPages), len(curvedWriter.clipped))
	}

	horizontalSector := curvedParagraphTestSector(t, 90)
	horizontalAttrs := maps.Clone(attrs)
	horizontalAttrs["angle"] = "0"
	horizontal := addTestSectorParagraph(t, horizontalSector, "Active horizontal paragraph box", horizontalAttrs)
	horizontalWriter := &labelTestWriter{t: t}
	if err := horizontalSector.LayoutWidget(horizontalWriter); err != nil {
		t.Fatal(err)
	}
	if horizontal.Width() != 80 || horizontal.Height() != 40 {
		t.Fatalf("horizontal box = %vx%v, want 80x40", horizontal.Width(), horizontal.Height())
	}
	if err := horizontalSector.DrawContent(horizontalWriter); err != nil {
		t.Fatal(err)
	}
	if len(horizontalWriter.rotations) == 0 || len(horizontalWriter.fillRectPages) == 0 || len(horizontalWriter.clipped) == 0 {
		t.Fatalf("horizontal active paints rotations/fills/text clips = %d/%d/%d",
			len(horizontalWriter.rotations), len(horizontalWriter.fillRectPages), len(horizontalWriter.clipped))
	}
}

func TestStdSector_StaticParagraphModesCannotMix(t *testing.T) {
	sector := curvedParagraphTestSector(t, 90)
	sector.path = "ltml/page/div/sector"
	addTestSectorParagraph(t, sector, "Curved", nil)
	addTestSectorParagraph(t, sector, "Horizontal", map[string]string{"angle": "0"})
	err := sector.LayoutWidget(&labelTestWriter{t: t})
	if err == nil || !strings.Contains(err.Error(), "ltml/page/div/sector") || !strings.Contains(err.Error(), "static curved and angle=\"0\" paragraphs") {
		t.Fatalf("mixed-mode error = %v", err)
	}
}

func TestStdSector_PositionedParagraphModesMayMix(t *testing.T) {
	sector := curvedParagraphTestSector(t, 90)
	addTestSectorParagraph(t, sector, "Curved", nil)
	addTestSectorParagraph(t, sector, "Horizontal overlay", map[string]string{
		"angle": "0", "position": "relative", "width": "80pt",
	})
	if err := sector.LayoutWidget(&labelTestWriter{t: t}); err != nil {
		t.Fatal(err)
	}
}

func TestStdSector_PositionedHorizontalParagraphUsesPageAxisSectorAnchor(t *testing.T) {
	sector := curvedParagraphTestSector(t, 45)
	addTestSectorParagraph(t, sector, "Curved flow", nil)
	horizontal := addTestSectorParagraph(t, sector, "Horizontal overlay", map[string]string{
		"angle": "0", "position": "relative", "origin-x": "center", "origin-y": "middle", "width": "80pt",
	})
	w := &labelTestWriter{t: t}
	if err := sector.LayoutWidget(w); err != nil {
		t.Fatal(err)
	}
	inner, outer := sector.contentRadii()
	wantX, wantY := radialPointAt(sector.geometry.CenterX, sector.geometry.CenterY,
		(inner+outer)/2, sector.geometry.AnchorAngle)
	if math.Abs(horizontal.OriginXValue()-wantX) > 0.01 || math.Abs(horizontal.OriginYValue()-wantY) > 0.01 {
		t.Fatalf("horizontal overlay origin = (%v,%v), want page point (%v,%v)",
			horizontal.OriginXValue(), horizontal.OriginYValue(), wantX, wantY)
	}
	if math.Abs(horizontal.Left()-(wantX-horizontal.Width()/2)) > 0.01 ||
		math.Abs(horizontal.Top()-(wantY-horizontal.Height()/2)) > 0.01 {
		t.Fatalf("horizontal overlay box = (%v,%v), want centered on (%v,%v)",
			horizontal.Left(), horizontal.Top(), wantX, wantY)
	}
}

func TestStdSector_CurvedParagraphRejectsInvalidLineRadius(t *testing.T) {
	sector := curvedParagraphTestSector(t, 90)
	sector.geometry.InnerRadius = 0
	sector.rebuildContentGeometry()
	addTestSectorParagraph(t, sector, "At the center", map[string]string{
		"position": "relative", "origin-x": "center", "origin-y": "middle", "units": "pt", "shift-y": "65",
	})
	err := sector.LayoutWidget(&labelTestWriter{t: t})
	if err == nil || !strings.Contains(err.Error(), "positive finite line radii") {
		t.Fatalf("invalid-radius error = %v", err)
	}
}
