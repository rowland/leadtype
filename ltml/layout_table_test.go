package ltml

import "testing"

func testTableContainer(width, height float64, cols int) *StdContainer {
	page := &StdPage{pageStyle: &PageStyle{width: 400, height: 400}}
	c := &StdContainer{}
	_ = c.SetContainer(page)
	c.SetWidth(width)
	if height > 0 {
		c.SetHeight(height)
	}
	c.cols = cols
	c.order = TableOrderRows
	c.layout = defaultLayouts["table"].Clone()
	return c
}

func addTableTestWidget(t *testing.T, c *StdContainer, preferredWidth, preferredHeight float64) *positionedTestWidget {
	t.Helper()
	w := &positionedTestWidget{preferredWidth: preferredWidth, preferredHeight: preferredHeight}
	if err := w.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(w)
	return w
}

func TestLayoutTable_AutoWidthUsesSurplusWithoutChangingOmittedColumns(t *testing.T) {
	c := testTableContainer(300, 80, 4)

	fixed := addTableTestWidget(t, c, 40, 20)
	fixed.SetWidth(60)

	omitted := addTableTestWidget(t, c, 35, 20)
	autoA := addTableTestWidget(t, c, 40, 20)
	autoA.SetWidthAuto()
	autoB := addTableTestWidget(t, c, 30, 20)
	autoB.SetWidthAuto()

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := fixed.Width(); got != 60 {
		t.Fatalf("fixed.Width() = %v, want 60", got)
	}
	if got := omitted.Width(); got != 35 {
		t.Fatalf("omitted.Width() = %v, want preferred width 35", got)
	}
	if got := autoA.Width(); got != 102.5 {
		t.Fatalf("autoA.Width() = %v, want 102.5", got)
	}
	if got := autoB.Width(); got != 102.5 {
		t.Fatalf("autoB.Width() = %v, want 102.5", got)
	}
}

func TestLayoutTable_AutoWidthPreservesPercentColumns(t *testing.T) {
	c := testTableContainer(400, 100, 4)

	percent := addTableTestWidget(t, c, 30, 20)
	percent.SetWidthPct(25)
	fixed := addTableTestWidget(t, c, 30, 20)
	fixed.SetWidth(50)
	omitted := addTableTestWidget(t, c, 40, 20)
	auto := addTableTestWidget(t, c, 50, 20)
	auto.SetWidthAuto()

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := percent.Width(); got != 100 {
		t.Fatalf("percent.Width() = %v, want 100", got)
	}
	if got := fixed.Width(); got != 50 {
		t.Fatalf("fixed.Width() = %v, want 50", got)
	}
	if got := omitted.Width(); got != 40 {
		t.Fatalf("omitted.Width() = %v, want preferred width 40", got)
	}
	if got := auto.Width(); got != 210 {
		t.Fatalf("auto.Width() = %v, want remaining width 210", got)
	}
}

func TestLayoutTable_LaterSingleColumnAutoCellMarksColumnAuto(t *testing.T) {
	c := testTableContainer(220, 100, 2)

	omittedA := addTableTestWidget(t, c, 40, 20)
	omittedB := addTableTestWidget(t, c, 30, 20)
	laterAutoA := addTableTestWidget(t, c, 45, 20)
	laterAutoA.SetWidthAuto()
	laterAutoB := addTableTestWidget(t, c, 35, 20)

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := omittedB.Width(); got != 35 {
		t.Fatalf("omittedB.Width() = %v, want preferred width 35", got)
	}
	if got := omittedA.Width(); got != 185 {
		t.Fatalf("omittedA.Width() = %v, want auto column surplus width 185", got)
	}
	if got := laterAutoA.Width(); got != 185 {
		t.Fatalf("laterAutoA.Width() = %v, want auto column surplus width 185", got)
	}
	if got := laterAutoB.Width(); got != 35 {
		t.Fatalf("laterAutoB.Width() = %v, want omitted column preferred width 35", got)
	}
}

func TestLayoutTable_OmittedWidthsKeepLegacyEqualShareWithoutAutoColumns(t *testing.T) {
	c := testTableContainer(300, 80, 3)

	a := addTableTestWidget(t, c, 20, 20)
	b := addTableTestWidget(t, c, 60, 20)
	d := addTableTestWidget(t, c, 100, 20)

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	for i, w := range []*positionedTestWidget{a, b, d} {
		if got := w.Width(); got != 100 {
			t.Fatalf("widget %d Width() = %v, want legacy equal share 100", i, got)
		}
	}
}

func TestLayoutTable_AutoWidthPreservesOmittedPreferredWidthWhenAutosCanShrink(t *testing.T) {
	c := testTableContainer(120, 80, 3)

	omitted := addTableTestWidget(t, c, 100, 20)
	autoA := addTableTestWidget(t, c, 80, 20)
	autoA.SetWidthAuto()
	autoB := addTableTestWidget(t, c, 80, 20)
	autoB.SetWidthAuto()

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := omitted.Width(); got != 100 {
		t.Fatalf("omitted.Width() = %v, want preferred width 100", got)
	}
	if got := autoA.Width(); got != 10 {
		t.Fatalf("autoA.Width() = %v, want remaining auto share 10", got)
	}
	if got := autoB.Width(); got != 10 {
		t.Fatalf("autoB.Width() = %v, want remaining auto share 10", got)
	}
}

func TestLayoutTable_AutoWidthFallsBackToEqualShareWhenOmittedPreferredCannotFit(t *testing.T) {
	c := testTableContainer(90, 80, 3)

	omitted := addTableTestWidget(t, c, 100, 20)
	autoA := addTableTestWidget(t, c, 80, 20)
	autoA.SetWidthAuto()
	autoB := addTableTestWidget(t, c, 80, 20)
	autoB.SetWidthAuto()

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	for i, w := range []*positionedTestWidget{omitted, autoA, autoB} {
		if got := w.Width(); got != 30 {
			t.Fatalf("widget %d Width() = %v, want impossible-case equal share 30", i, got)
		}
	}
}

func TestLayoutTable_AspectInferredWidthDefinesColumnWidth(t *testing.T) {
	c := testTableContainer(300, 100, 2)

	aspect := &aspectRatioTestWidget{
		positionedTestWidget: positionedTestWidget{preferredWidth: 150, preferredHeight: 20},
		aspectRatio:          4,
	}
	aspect.SetHeight(20)
	if err := aspect.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(aspect)

	omitted := addTableTestWidget(t, c, 30, 20)

	c.prepareForLayout(&labelTestWriter{t: t})
	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := aspect.Width(); got != 80 {
		t.Fatalf("aspect widget width = %v, want aspect-inferred column width 80", got)
	}
	if got := omitted.Width(); got != 220 {
		t.Fatalf("omitted widget width = %v, want remaining column width 220", got)
	}
}

func TestLayoutTable_ColspanDoesNotCreateAutoColumnConstraint(t *testing.T) {
	c := testTableContainer(240, 100, 2)

	spanning := addTableTestWidget(t, c, 200, 20)
	spanning.SetWidthAuto()
	spanning.SetAttrs(map[string]string{"colspan": "2"})

	a := addTableTestWidget(t, c, 30, 20)
	b := addTableTestWidget(t, c, 50, 20)

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := a.Width(); got != 120 {
		t.Fatalf("a.Width() = %v, want legacy equal share 120", got)
	}
	if got := b.Width(); got != 120 {
		t.Fatalf("b.Width() = %v, want legacy equal share 120", got)
	}
	if got := spanning.Width(); got != 240 {
		t.Fatalf("spanning.Width() = %v, want full two-column span 240", got)
	}
}

func TestLayoutTable_AutoHeightRowsShareSurplus(t *testing.T) {
	c := testTableContainer(200, 150, 2)

	fixedA := addTableTestWidget(t, c, 30, 20)
	fixedA.SetHeight(20)
	fixedB := addTableTestWidget(t, c, 30, 20)
	fixedB.SetHeight(20)

	omittedA := addTableTestWidget(t, c, 30, 30)
	omittedB := addTableTestWidget(t, c, 30, 30)

	autoA := addTableTestWidget(t, c, 30, 10)
	autoA.SetHeightAuto()
	autoB := addTableTestWidget(t, c, 30, 10)
	autoB.SetHeightAuto()

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := omittedA.Height(); got != 30 {
		t.Fatalf("omittedA.Height() = %v, want preferred height 30", got)
	}
	if got := omittedB.Height(); got != 30 {
		t.Fatalf("omittedB.Height() = %v, want preferred height 30", got)
	}
	if got := autoA.Height(); got != 100 {
		t.Fatalf("autoA.Height() = %v, want 100", got)
	}
	if got := autoB.Height(); got != 100 {
		t.Fatalf("autoB.Height() = %v, want 100", got)
	}
}

func TestLayoutTable_AutoHeightRowsPreservePercentRows(t *testing.T) {
	c := testTableContainer(200, 200, 2)

	percentA := addTableTestWidget(t, c, 30, 20)
	percentA.SetHeightPct(25)
	percentB := addTableTestWidget(t, c, 30, 20)
	percentB.SetHeightPct(25)

	omittedA := addTableTestWidget(t, c, 30, 30)
	omittedB := addTableTestWidget(t, c, 30, 30)

	autoA := addTableTestWidget(t, c, 30, 10)
	autoA.SetHeightAuto()
	autoB := addTableTestWidget(t, c, 30, 10)
	autoB.SetHeightAuto()

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := percentA.Height(); got != 50 {
		t.Fatalf("percentA.Height() = %v, want 50", got)
	}
	if got := percentB.Height(); got != 50 {
		t.Fatalf("percentB.Height() = %v, want 50", got)
	}
	if got := omittedA.Height(); got != 30 {
		t.Fatalf("omittedA.Height() = %v, want preferred height 30", got)
	}
	if got := omittedB.Height(); got != 30 {
		t.Fatalf("omittedB.Height() = %v, want preferred height 30", got)
	}
	if got := autoA.Height(); got != 120 {
		t.Fatalf("autoA.Height() = %v, want remaining auto height 120", got)
	}
	if got := autoB.Height(); got != 120 {
		t.Fatalf("autoB.Height() = %v, want remaining auto height 120", got)
	}
}

func TestLayoutTable_KeepsRowsThatOnlyExceedBottomByFitEpsilon(t *testing.T) {
	c := testTableContainer(100, 30-layoutFitEpsilon/2, 1)

	first := addTableTestWidget(t, c, 30, 15)
	second := addTableTestWidget(t, c, 30, 15)

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if !first.Visible() {
		t.Fatal("first row hidden")
	}
	if !second.Visible() {
		t.Fatal("second row hidden despite fitting within layout tolerance")
	}
}

func TestLayoutTable_NaturalHeightLeavesAutoRowsAtPreferredHeight(t *testing.T) {
	c := testTableContainer(200, 0, 1)

	auto := addTableTestWidget(t, c, 30, 25)
	auto.SetHeightAuto()
	omitted := addTableTestWidget(t, c, 30, 35)

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := auto.Height(); got != 25 {
		t.Fatalf("auto.Height() = %v, want preferred height 25", got)
	}
	if got := omitted.Height(); got != 35 {
		t.Fatalf("omitted.Height() = %v, want preferred height 35", got)
	}
	if got := c.Height(); got != 60 {
		t.Fatalf("container.Height() = %v, want natural height 60", got)
	}
}

func TestLayoutTable_AlignSelfPositionsCellsWithoutStretching(t *testing.T) {
	c := testTableContainer(400, 0, 4)

	stretched := addTableTestWidget(t, c, 30, 60)

	start := addTableTestWidget(t, c, 30, 20)
	start.SetAttrs(map[string]string{"align-self": "start"})

	center := addTableTestWidget(t, c, 30, 10)
	center.SetHeight(30)
	center.SetAttrs(map[string]string{"align-self": "center"})

	end := addTableTestWidget(t, c, 30, 20)
	end.SetAttrs(map[string]string{"align-self": "end"})

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := stretched.Height(); got != 60 {
		t.Fatalf("default cell height = %v, want stretched row height 60", got)
	}
	if got := start.Top(); got != 0 {
		t.Fatalf("start cell top = %v, want 0", got)
	}
	if got := start.Height(); got != 20 {
		t.Fatalf("start cell height = %v, want preferred height 20", got)
	}
	if got := center.Top(); got != 15 {
		t.Fatalf("center cell top = %v, want 15", got)
	}
	if got := center.Height(); got != 30 {
		t.Fatalf("center cell height = %v, want declared height 30", got)
	}
	if got := end.Top(); got != 40 {
		t.Fatalf("end cell top = %v, want 40", got)
	}
	if got := end.Height(); got != 20 {
		t.Fatalf("end cell height = %v, want preferred height 20", got)
	}
}

func TestLayoutTable_AlignSelfUsesFullRowspanHeightIncludingPadding(t *testing.T) {
	c := testTableContainer(200, 0, 2)
	c.layout.vpadding = 10

	spanning := addTableTestWidget(t, c, 30, 20)
	spanning.SetAttrs(map[string]string{"rowspan": "2", "align-self": "end"})
	addTableTestWidget(t, c, 30, 30)
	addTableTestWidget(t, c, 30, 40)

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := spanning.Top(); got != 60 {
		t.Fatalf("spanning cell top = %v, want 60 within 80-point span", got)
	}
	if got := spanning.Height(); got != 20 {
		t.Fatalf("spanning cell height = %v, want preferred height 20", got)
	}
}

func TestLayoutTable_AlignSelfStartPreservesIntrinsicShapeHeight(t *testing.T) {
	c := testTableContainer(200, 0, 2)

	addTableTestWidget(t, c, 30, 60)
	circle := &StdCircle{radiusValue: 10}
	circle.SetAttrs(map[string]string{"align-self": "start"})
	if err := circle.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(circle)

	LayoutTable(c, c.LayoutStyle(), &labelTestWriter{t: t})

	if got := circle.Top(); got != 0 {
		t.Fatalf("circle top = %v, want row top 0", got)
	}
	if got := circle.Height(); got != 20 {
		t.Fatalf("circle height = %v, want intrinsic diameter 20", got)
	}
}
