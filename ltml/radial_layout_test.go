package ltml

import (
	"math"
	"strings"
	"testing"

	"github.com/rowland/leadtype/pdf"
)

func testSectorFont() *FontStyle {
	return &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
}

func TestStdSector_RichText_ReappliesFontsWhenUsingCachedRichText(t *testing.T) {
	sector := &StdSector{StdContainer: StdContainer{paragraphStyle: &ParagraphStyle{}}}
	sector.font = testSectorFont()
	sector.AddText("Hello")

	probe := &mockWriter{t: t}
	if got := sector.RichText(probe); got == nil {
		t.Fatal("expected cached rich text to be built")
	}

	render := &mockWriter{t: t}
	if got := sector.RichText(render); got == nil {
		t.Fatal("expected cached rich text to be returned")
	}
	if len(render.setFontCalls) == 0 {
		t.Fatal("expected cached rich text path to apply fonts to the render writer")
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
		{90, 0},
		{180, 90},
		{270, 180},
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
		{90, 0},
		{180, 90},
		{270, 180},
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
	sector.AddText("Luke")
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
	sector.AddText("Leia")
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
	sector.AddText("Radial")
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

	sector.angle = 90
	sector.angleSet = true
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
	sector.SetAttrs(map[string]string{"text-align": "right"})
	sector.AddText("Radial")
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

	centerX, centerY := sector.contentLocalCenter()
	if got, want := (label.Left()+label.Right())/2, sector.geometry.AnchorX+centerX; math.Abs(got-want) > 0.5 {
		t.Fatalf("label center x = %v, want near %v", got, want)
	}
	if got, want := (label.Top()+label.Bottom())/2, sector.geometry.AnchorY+centerY; math.Abs(got-want) > 0.5 {
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

func TestRadialSample_ImplicitParagraphsKeepLegacyPlacement(t *testing.T) {
	doc, err := ParseFile(sampleFile("test_038_radial_layout.ltml"))
	if err != nil {
		t.Fatal(err)
	}

	w := &labelTestWriter{t: t}
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}

	var direct *StdParagraph
	var other *StdParagraph
	walkWidgets(doc.Root(), func(widget Widget) bool {
		paragraph, ok := widget.(*StdParagraph)
		if !ok {
			return true
		}
		text := paragraph.AccessibilityText()
		if strings.HasPrefix(text, "This paragraph is a direct child") {
			direct = paragraph
		}
		if strings.HasPrefix(text, "Another implicit sector paragraph") {
			other = paragraph
		}
		return true
	})

	if direct == nil || other == nil {
		t.Fatalf("missing implicit paragraphs: direct=%v other=%v", direct != nil, other != nil)
	}
	if got, want := direct.Top(), 175.03; math.Abs(got-want) > 0.5 {
		t.Fatalf("direct implicit paragraph top = %v, want near %v", got, want)
	}
	if got, want := other.Top(), 386.59; math.Abs(got-want) > 0.5 {
		t.Fatalf("secondary implicit paragraph top = %v, want near %v", got, want)
	}
}
