package ltml

import (
	"math"
	"testing"
)

func testHBoxLabel(t *testing.T, text string) *StdLabel {
	t.Helper()
	label := &StdLabel{}
	label.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 11}
	label.SetAttrs(map[string]string{"padding": "6pt"})
	label.AddText(text)
	return label
}

type aspectRatioTestWidget struct {
	positionedTestWidget
	aspectRatio float64
}

func (w *aspectRatioTestWidget) IntrinsicAspectRatio(Writer) (float64, bool) {
	if w.aspectRatio <= 0 {
		return 0, false
	}
	return w.aspectRatio, true
}

func TestLayoutVBox_AlignSelfCenterCentersTopChildHorizontally(t *testing.T) {
	c := positionedContainer(0, 0, 300, 200)
	style := &LayoutStyle{}

	w := &positionedTestWidget{preferredWidth: 100, preferredHeight: 30}
	w.SetWidth(100)
	w.SetAttrs(map[string]string{"align": "top", "align-self": "center"})
	w.SetContainer(c)
	c.AddChild(w)

	LayoutVBox(c, style, nil)

	if got := w.Left(); got != 100 {
		t.Errorf("vbox top child left = %v, want 100", got)
	}
	if got := w.Top(); got != 0 {
		t.Errorf("vbox top child top = %v, want 0", got)
	}
}

func TestLayoutVBox_UsesAspectInferredWidth(t *testing.T) {
	c := positionedContainer(0, 0, 50, 120)
	widget := &aspectRatioTestWidget{
		positionedTestWidget: positionedTestWidget{preferredWidth: 100, preferredHeight: 20},
		aspectRatio:          4,
	}
	widget.SetHeight(20)
	if err := widget.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(widget)

	c.prepareForLayout(&labelTestWriter{t: t})
	LayoutVBox(c, &LayoutStyle{}, &labelTestWriter{t: t})

	if got := widget.Width(); got != 80 {
		t.Fatalf("widget width = %v, want aspect-inferred width 80", got)
	}
}

func TestLayoutVBox_UsesAspectInferredHeight(t *testing.T) {
	c := positionedContainer(0, 0, 120, 120)
	widget := &aspectRatioTestWidget{
		positionedTestWidget: positionedTestWidget{preferredWidth: 100, preferredHeight: 100},
		aspectRatio:          4,
	}
	widget.SetWidth(80)
	if err := widget.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(widget)

	c.prepareForLayout(&labelTestWriter{t: t})
	LayoutVBox(c, &LayoutStyle{}, &labelTestWriter{t: t})

	if got := widget.Height(); got != 20 {
		t.Fatalf("widget height = %v, want aspect-inferred height 20", got)
	}
}

func TestPrepareAspectRatioDimensions_LeavesAutoAndUnspecifiedDimensionsLayoutManaged(t *testing.T) {
	c := positionedContainer(0, 0, 120, 120)
	auto := &aspectRatioTestWidget{
		positionedTestWidget: positionedTestWidget{preferredWidth: 100, preferredHeight: 100},
		aspectRatio:          4,
	}
	auto.SetHeight(20)
	auto.SetWidthAuto()
	if err := auto.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(auto)

	unspecified := &aspectRatioTestWidget{
		positionedTestWidget: positionedTestWidget{preferredWidth: 100, preferredHeight: 100},
		aspectRatio:          4,
	}
	if err := unspecified.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(unspecified)

	c.prepareForLayout(&labelTestWriter{t: t})

	if auto.WidthAspectInferred() || auto.WidthIsSet() {
		t.Fatalf("width=auto should remain layout-managed")
	}
	if unspecified.WidthAspectInferred() || unspecified.HeightAspectInferred() {
		t.Fatalf("fully unspecified widget should not infer dimensions during preparation")
	}
}

func TestLayoutVBox_ParagraphDefaultsToFullWidth(t *testing.T) {
	c := positionedContainer(0, 0, 300, 200)
	p := &StdParagraph{}
	if err := p.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	p.paragraphStyle = &ParagraphStyle{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.AddText("Short heading")
	c.AddChild(p)

	LayoutVBox(c, &LayoutStyle{}, &labelTestWriter{t: t})

	if got := p.Width(); got != 300 {
		t.Fatalf("paragraph width = %v, want 300", got)
	}
	if got := p.Left(); got != 0 {
		t.Fatalf("paragraph left = %v, want 0", got)
	}
}

func TestStdContainer_PreferredWidthForVBoxUsesChildren(t *testing.T) {
	c := &StdContainer{}
	c.layout = &LayoutStyle{manager: "vbox"}

	label := testHBoxLabel(t, "Natural Heading")
	label.SetWidth(180)
	if err := label.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(label)

	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.AddText("This paragraph is deliberately long enough that it wants more than the heading.")
	if err := p.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(p)

	if got := mustPreferredWidth(t, c, &labelTestWriter{t: t}); got <= 180 {
		t.Fatalf("vbox PreferredWidth() = %v, want greater than heading width 180", got)
	}
}

func TestLayoutHBox_UsesAspectRatioProviderForThirdPartyWidgets(t *testing.T) {
	hbox := positionedContainer(0, 0, 300, 120)
	hbox.layout = &LayoutStyle{manager: "hbox"}

	aspect := &aspectRatioTestWidget{
		positionedTestWidget: positionedTestWidget{preferredWidth: 100, preferredHeight: 40},
		aspectRatio:          2,
	}
	aspect.SetHeight(40)
	if err := aspect.SetContainer(hbox); err != nil {
		t.Fatal(err)
	}
	hbox.AddChild(aspect)

	panel := &positionedTestWidget{preferredWidth: 260, preferredHeight: 40}
	panel.SetWidthAuto()
	if err := panel.SetContainer(hbox); err != nil {
		t.Fatal(err)
	}
	hbox.AddChild(panel)

	hbox.prepareForLayout(&labelTestWriter{t: t})
	LayoutHBox(hbox, hbox.layout, &labelTestWriter{t: t})

	if got := aspect.Width(); got != 80 {
		t.Fatalf("aspect widget width = %v, want 80", got)
	}
	if got := panel.Left(); got != 80 {
		t.Fatalf("auto sibling left = %v, want after aspect-inferred width 80", got)
	}
}

func TestLayoutHBox_ConstrainedAutoVBoxPropagatesWidthToNestedLabel(t *testing.T) {
	hbox := positionedContainer(0, 0, 500, 100)
	hbox.layout = &LayoutStyle{manager: "hbox", hpadding: 20}

	outer := &StdContainer{}
	outer.layout = &LayoutStyle{manager: "vbox"}
	outer.SetWidthAuto()
	if err := outer.SetContainer(hbox); err != nil {
		t.Fatal(err)
	}
	hbox.AddChild(outer)

	inner := &StdContainer{}
	inner.layout = &LayoutStyle{manager: "vbox"}
	if err := inner.SetContainer(outer); err != nil {
		t.Fatal(err)
	}
	outer.AddChild(inner)

	title := testHBoxLabel(t, "A deliberately long heading that is much wider than the available header track and must shrink")
	title.shrinkToFit = true
	if err := title.SetContainer(inner); err != nil {
		t.Fatal(err)
	}
	inner.AddChild(title)

	logo := &positionedTestWidget{preferredWidth: 40, preferredHeight: 40}
	logo.align = AlignRight
	if err := logo.SetContainer(hbox); err != nil {
		t.Fatal(err)
	}
	hbox.AddChild(logo)

	writer := &labelTestWriter{t: t}
	if err := LayoutHBox(hbox, hbox.layout, writer); err != nil {
		t.Fatal(err)
	}

	if got := outer.Width(); got != 440 {
		t.Fatalf("outer.Width() = %v, want 440", got)
	}
	if got := inner.Width(); got != 440 {
		t.Fatalf("inner.Width() = %v, want 440", got)
	}
	if got := title.Width(); got != 440 {
		t.Fatalf("title.Width() = %v, want 440", got)
	}
}

func TestLayoutHBox_SpecifiedWidthsFitWhenContainerMatchesPreferredSum(t *testing.T) {
	const childW = 72.0
	const hpad = 25.2

	hbox := &StdContainer{}
	hbox.layout = &LayoutStyle{manager: "hbox", hpadding: hpad}

	left := &positionedTestWidget{preferredWidth: childW, preferredHeight: childW}
	left.SetWidth(childW)
	if err := left.SetContainer(hbox); err != nil {
		t.Fatal(err)
	}
	hbox.AddChild(left)

	right := &positionedTestWidget{preferredWidth: childW, preferredHeight: childW}
	right.SetWidth(childW)
	if err := right.SetContainer(hbox); err != nil {
		t.Fatal(err)
	}
	hbox.AddChild(right)

	writer := &labelTestWriter{t: t}
	wantWidth := childW + hpad + childW
	if got := mustPreferredWidth(t, hbox, writer); math.Abs(got-wantWidth) > layoutFitEpsilon {
		t.Fatalf("PreferredWidth() = %v, want %v", got, wantWidth)
	}
	hbox.SetWidth(wantWidth)
	hbox.SetHeight(100)

	LayoutHBox(hbox, hbox.layout, writer)

	if left.Disabled() || right.Disabled() {
		t.Fatalf("specified children disabled with exact-fit width: left=%v right=%v", left.Disabled(), right.Disabled())
	}
}

func TestLayoutHBox_HeightOnlyImageReservesAspectWidthBeforeAutoSiblingShrinks(t *testing.T) {
	hbox := positionedContainer(0, 0, 300, 120)
	hbox.layout = &LayoutStyle{manager: "hbox"}

	img := &StdImage{src: "fixture.png"}
	img.SetDoc(newDocWithOptions(WithAssetFS(testingMapFS("fixture.png", "image-data"))))
	img.SetHeight(80)
	if err := img.SetContainer(hbox); err != nil {
		t.Fatal(err)
	}
	hbox.AddChild(img)

	panel := &positionedTestWidget{preferredWidth: 260, preferredHeight: 40}
	panel.SetWidthAuto()
	if err := panel.SetContainer(hbox); err != nil {
		t.Fatal(err)
	}
	hbox.AddChild(panel)

	writer := &imageTestWriter{dimensions: map[string][2]int{"fixture.png": {100, 100}}}
	hbox.prepareForLayout(writer)
	LayoutHBox(hbox, hbox.layout, writer)

	if got := img.Width(); got != 80 {
		t.Fatalf("image width = %v, want aspect width 80", got)
	}
	if got := panel.Left(); got != 80 {
		t.Fatalf("auto sibling left = %v, want immediately after image width 80", got)
	}
	if got := panel.Width(); got != 220 {
		t.Fatalf("auto sibling width = %v, want remaining width 220", got)
	}
}

func TestLayoutHBox_AutoWidthCanCollapseWhenOmittedPreferredWidthsConsumeSpace(t *testing.T) {
	c := positionedContainer(0, 0, 300, 100)
	style := &LayoutStyle{}

	vbox := &StdContainer{}
	vbox.layout = &LayoutStyle{manager: "vbox"}
	if err := vbox.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(vbox)

	label := testHBoxLabel(t, "Natural Heading")
	label.SetWidth(180)
	if err := label.SetContainer(vbox); err != nil {
		t.Fatal(err)
	}
	vbox.AddChild(label)

	p := &StdParagraph{}
	p.font = &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}, size: 12}
	p.AddText("This paragraph is deliberately long enough that its max-content width would be too large for a natural heading column.")
	if err := p.SetContainer(vbox); err != nil {
		t.Fatal(err)
	}
	vbox.AddChild(p)

	auto := &positionedTestWidget{preferredWidth: 0, preferredHeight: 20}
	auto.SetWidthAuto()
	if err := auto.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(auto)

	LayoutHBox(c, style, &labelTestWriter{t: t})

	if got := auto.Width(); got != 0 {
		t.Fatalf("auto Width() = %v, want 0", got)
	}
	if got := vbox.Width(); got <= 0 {
		t.Fatalf("vbox Width() = %v, want positive width", got)
	}
}

func TestLayoutHBox_RightAlignedOmittedWidthReservesSpaceBeforeCenterRun(t *testing.T) {
	c := positionedContainer(0, 0, 300, 100)
	style := &LayoutStyle{}

	center := &positionedTestWidget{preferredWidth: 300, preferredHeight: 20}
	if err := center.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(center)

	spacer := &positionedTestWidget{preferredWidth: 0, preferredHeight: 20}
	spacer.SetWidthAuto()
	if err := spacer.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(spacer)

	right := &positionedTestWidget{preferredWidth: 40, preferredHeight: 20}
	right.align = AlignRight
	if err := right.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(right)

	LayoutHBox(c, style, nil)

	if got := right.Width(); got != 40 {
		t.Fatalf("right.Width() = %v, want 40", got)
	}
	if got := right.Left(); got != 260 {
		t.Fatalf("right.Left() = %v, want 260", got)
	}
	if got := center.Width(); got > 260 {
		t.Fatalf("center.Width() = %v, want no more than 260", got)
	}
	if center.Left()+center.Width() > right.Left()+layoutFitEpsilon {
		t.Fatalf("center overlaps right panel: center right %v, right left %v", center.Left()+center.Width(), right.Left())
	}
}

func TestLayoutHBox_AutoWidthAbsorbsOnlySurplusSpace(t *testing.T) {
	c := positionedContainer(0, 0, 400, 100)
	style := &LayoutStyle{hpadding: 10}

	fixed := &positionedTestWidget{preferredWidth: 40, preferredHeight: 20}
	fixed.SetWidth(40)
	if err := fixed.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(fixed)

	pct := &positionedTestWidget{preferredWidth: 30, preferredHeight: 20}
	pct.SetWidthPct(25)
	if err := pct.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(pct)

	omitted := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	if err := omitted.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(omitted)

	auto1 := &positionedTestWidget{preferredWidth: 60, preferredHeight: 20}
	auto1.SetWidthAuto()
	if err := auto1.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(auto1)

	auto2 := &positionedTestWidget{preferredWidth: 40, preferredHeight: 20}
	auto2.SetWidthAuto()
	if err := auto2.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(auto2)

	LayoutHBox(c, style, nil)

	if got := fixed.Width(); got != 40 {
		t.Fatalf("fixed.Width() = %v, want 40", got)
	}
	if got := pct.Width(); got != 100 {
		t.Fatalf("pct.Width() = %v, want 100", got)
	}
	if got := omitted.Width(); got != 80 {
		t.Fatalf("omitted.Width() = %v, want 80", got)
	}
	if got := auto1.Width(); got != 80 {
		t.Fatalf("auto1.Width() = %v, want 80", got)
	}
	if got := auto2.Width(); got != 60 {
		t.Fatalf("auto2.Width() = %v, want 60", got)
	}
}

func TestLayoutHBox_AutoWidthsFitAfterSpecifiedWidthsWithoutPercentGroup(t *testing.T) {
	c := positionedContainer(0, 0, 400, 100)
	style := &LayoutStyle{hpadding: 30}

	for i := 0; i < 2; i++ {
		fixed := &positionedTestWidget{preferredWidth: 150, preferredHeight: 20}
		fixed.SetWidth(150)
		if err := fixed.SetContainer(c); err != nil {
			t.Fatal(err)
		}
		c.AddChild(fixed)
	}

	auto := make([]*positionedTestWidget, 2)
	for i := range auto {
		auto[i] = &positionedTestWidget{preferredWidth: 10, preferredHeight: 20}
		auto[i].SetWidthAuto()
		if err := auto[i].SetContainer(c); err != nil {
			t.Fatal(err)
		}
		c.AddChild(auto[i])
	}

	if err := LayoutHBox(c, style, nil); err != nil {
		t.Fatal(err)
	}

	for i, widget := range auto {
		if widget.Disabled() {
			t.Fatalf("auto[%d] disabled, want it fitted into the remaining width", i)
		}
		if got := widget.Width(); got != 5 {
			t.Fatalf("auto[%d].Width() = %v, want 5", i, got)
		}
	}
}

func TestLayoutHBox_AutoWidthScalesPreferredWidthsWhenConstrained(t *testing.T) {
	c := positionedContainer(0, 0, 340, 100)
	style := &LayoutStyle{hpadding: 10}

	fixed := &positionedTestWidget{preferredWidth: 40, preferredHeight: 20}
	fixed.SetWidth(40)
	if err := fixed.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(fixed)

	pct := &positionedTestWidget{preferredWidth: 30, preferredHeight: 20}
	pct.SetWidthPct(25)
	if err := pct.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(pct)

	omitted := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	if err := omitted.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(omitted)

	auto1 := &positionedTestWidget{preferredWidth: 60, preferredHeight: 20}
	auto1.SetWidthAuto()
	if err := auto1.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(auto1)

	auto2 := &positionedTestWidget{preferredWidth: 40, preferredHeight: 20}
	auto2.SetWidthAuto()
	if err := auto2.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(auto2)

	LayoutHBox(c, style, nil)

	if got := omitted.Width(); math.Abs(got-(80*175.0/180.0)) > 0.001 {
		t.Fatalf("omitted.Width() = %v, want scaled preferred width", got)
	}
	if got := auto1.Width(); math.Abs(got-(60*175.0/180.0)) > 0.001 {
		t.Fatalf("auto1.Width() = %v, want scaled preferred width", got)
	}
	if got := auto2.Width(); math.Abs(got-(40*175.0/180.0)) > 0.001 {
		t.Fatalf("auto2.Width() = %v, want scaled preferred width", got)
	}
}

func TestLayoutHBox_AutoWidthReceivesSpaceWhenOmittedPreferredWidthIsZero(t *testing.T) {
	c := positionedContainer(0, 0, 300, 100)
	style := &LayoutStyle{}

	omitted := &positionedTestWidget{preferredWidth: 0, preferredHeight: 20}
	if err := omitted.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(omitted)

	auto := &positionedTestWidget{preferredWidth: 0, preferredHeight: 20}
	auto.SetWidthAuto()
	if err := auto.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(auto)

	LayoutHBox(c, style, nil)

	if got := omitted.Width(); got != 0 {
		t.Fatalf("omitted.Width() = %v, want 0", got)
	}
	if got := auto.Width(); math.Abs(got-300) > 0.001 {
		t.Fatalf("auto.Width() = %v, want remaining width 300", got)
	}
}

func TestLayoutHBox_AutoWidthMatchesDirectLayoutAfterVBoxProbe(t *testing.T) {
	writer := &labelTestWriter{t: t}
	hboxStyle := &LayoutStyle{manager: "hbox", hpadding: 10}

	buildBox := func() (*StdContainer, []*StdLabel) {
		box := &StdContainer{}
		box.layout = hboxStyle
		box.SetWidth(3.9 * 72)
		box.SetAttrs(map[string]string{"padding": "8pt"})

		fixed := testHBoxLabel(t, "fixed")
		fixed.SetWidth(0.75 * 72)
		if err := fixed.SetContainer(box); err != nil {
			t.Fatal(err)
		}
		box.AddChild(fixed)

		preferred := testHBoxLabel(t, "preferred width")
		if err := preferred.SetContainer(box); err != nil {
			t.Fatal(err)
		}
		box.AddChild(preferred)

		auto1 := testHBoxLabel(t, "auto A")
		auto1.SetWidthAuto()
		if err := auto1.SetContainer(box); err != nil {
			t.Fatal(err)
		}
		box.AddChild(auto1)

		auto2 := testHBoxLabel(t, "auto B")
		auto2.SetWidthAuto()
		if err := auto2.SetContainer(box); err != nil {
			t.Fatal(err)
		}
		box.AddChild(auto2)

		return box, []*StdLabel{fixed, preferred, auto1, auto2}
	}

	directBox, directLabels := buildBox()
	LayoutHBox(directBox, hboxStyle, writer)
	directWidths := make([]float64, len(directLabels))
	for i, label := range directLabels {
		directWidths[i] = label.Width()
	}

	probedBox, probedLabels := buildBox()
	outer := positionedContainer(0, 0, 400, 300)
	outer.layout = &LayoutStyle{manager: "vbox"}
	if err := probedBox.SetContainer(outer); err != nil {
		t.Fatal(err)
	}
	outer.AddChild(probedBox)

	LayoutVBox(outer, outer.layout, writer)

	for i, label := range probedLabels {
		if got := label.Width(); math.Abs(got-directWidths[i]) > 0.001 {
			t.Fatalf("label %d width after vbox probe = %v, want %v", i, got, directWidths[i])
		}
	}
}

func TestLayoutHBox_WithoutAutoKeepsExistingEqualShareBehavior(t *testing.T) {
	c := positionedContainer(0, 0, 300, 100)
	style := &LayoutStyle{hpadding: 10}

	fixed := &positionedTestWidget{preferredWidth: 40, preferredHeight: 20}
	fixed.SetWidth(40)
	if err := fixed.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(fixed)

	pct := &positionedTestWidget{preferredWidth: 30, preferredHeight: 20}
	pct.SetWidthPct(100.0 / 3.0)
	if err := pct.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(pct)

	omitted1 := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	if err := omitted1.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(omitted1)

	omitted2 := &positionedTestWidget{preferredWidth: 50, preferredHeight: 20}
	if err := omitted2.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(omitted2)

	LayoutHBox(c, style, nil)

	if got := omitted1.Width(); math.Abs(got-65) > 0.001 {
		t.Fatalf("omitted1.Width() = %v, want 65", got)
	}
	if got := omitted2.Width(); math.Abs(got-65) > 0.001 {
		t.Fatalf("omitted2.Width() = %v, want 65", got)
	}
}

func TestLayoutHBox_AlignSelfCenterCentersLeftChildVertically(t *testing.T) {
	c := positionedContainer(0, 0, 300, 100)
	style := &LayoutStyle{}

	w := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	w.SetWidth(80)
	w.SetAttrs(map[string]string{"align": "left", "align-self": "center"})
	w.SetContainer(c)
	c.AddChild(w)

	LayoutHBox(c, style, nil)

	if got := w.Top(); got != 40 {
		t.Errorf("hbox left child top = %v, want 40", got)
	}
	if got := w.Left(); got != 0 {
		t.Errorf("hbox left child left = %v, want 0", got)
	}
}

func TestLayoutVBox_AutoWidthMatchesOmittedWidth(t *testing.T) {
	c := positionedContainer(0, 0, 200, 100)
	style := &LayoutStyle{}

	omitted := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	if err := omitted.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(omitted)

	auto := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	auto.SetWidthAuto()
	if err := auto.SetContainer(c); err != nil {
		t.Fatal(err)
	}
	c.AddChild(auto)

	LayoutVBox(c, style, nil)

	if got := omitted.Width(); got != 80 {
		t.Fatalf("omitted.Width() = %v, want 80", got)
	}
	if got := auto.Width(); got != 80 {
		t.Fatalf("auto.Width() = %v, want 80", got)
	}
}

func TestLayoutVBox_AutoHeightAbsorbsOnlySurplusSpace(t *testing.T) {
	c := positionedContainer(0, 0, 200, 200)
	style := &LayoutStyle{vpadding: 10}

	fixed := &positionedTestWidget{preferredWidth: 80, preferredHeight: 40}
	fixed.SetHeight(40)
	_ = fixed.SetContainer(c)
	c.AddChild(fixed)

	pct := &positionedTestWidget{preferredWidth: 80, preferredHeight: 15}
	pct.SetHeightPct(25)
	_ = pct.SetContainer(c)
	c.AddChild(pct)

	omitted := &positionedTestWidget{preferredWidth: 80, preferredHeight: 30}
	_ = omitted.SetContainer(c)
	c.AddChild(omitted)

	auto1 := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	auto1.SetHeightAuto()
	_ = auto1.SetContainer(c)
	c.AddChild(auto1)

	auto2 := &positionedTestWidget{preferredWidth: 80, preferredHeight: 10}
	auto2.SetHeightAuto()
	_ = auto2.SetContainer(c)
	c.AddChild(auto2)

	LayoutVBox(c, style, nil)

	if got := fixed.Height(); got != 40 {
		t.Fatalf("fixed.Height() = %v, want 40", got)
	}
	if got := pct.Height(); got != 50 {
		t.Fatalf("pct.Height() = %v, want 50", got)
	}
	if got := omitted.Height(); got != 30 {
		t.Fatalf("omitted.Height() = %v, want 30", got)
	}
	if got := auto1.Height(); got != 25 {
		t.Fatalf("auto1.Height() = %v, want 25", got)
	}
	if got := auto2.Height(); got != 15 {
		t.Fatalf("auto2.Height() = %v, want 15", got)
	}
}

func TestLayoutVBox_AutoHeightHasNoEffectWhenNotRoomy(t *testing.T) {
	c := positionedContainer(0, 0, 200, 130)
	style := &LayoutStyle{vpadding: 10}

	fixed := &positionedTestWidget{preferredWidth: 80, preferredHeight: 40}
	fixed.SetHeight(40)
	_ = fixed.SetContainer(c)
	c.AddChild(fixed)

	omitted := &positionedTestWidget{preferredWidth: 80, preferredHeight: 30}
	_ = omitted.SetContainer(c)
	c.AddChild(omitted)

	auto1 := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	auto1.SetHeightAuto()
	_ = auto1.SetContainer(c)
	c.AddChild(auto1)

	auto2 := &positionedTestWidget{preferredWidth: 80, preferredHeight: 10}
	auto2.SetHeightAuto()
	_ = auto2.SetContainer(c)
	c.AddChild(auto2)

	LayoutVBox(c, style, nil)

	if got := omitted.Height(); got != 30 {
		t.Fatalf("omitted.Height() = %v, want 30", got)
	}
	if got := auto1.Height(); got != 20 {
		t.Fatalf("auto1.Height() = %v, want 20", got)
	}
	if got := auto2.Height(); got != 10 {
		t.Fatalf("auto2.Height() = %v, want 10", got)
	}
}

func TestLayoutVBox_AutoHeightMatchesOmittedHeightWhenContainerIsNatural(t *testing.T) {
	c := positionedContainer(0, 0, 200, 0)
	c.ClearHeight()
	style := &LayoutStyle{vpadding: 10}

	omitted := &positionedTestWidget{preferredWidth: 80, preferredHeight: 30}
	_ = omitted.SetContainer(c)
	c.AddChild(omitted)

	auto := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	auto.SetHeightAuto()
	_ = auto.SetContainer(c)
	c.AddChild(auto)

	LayoutVBox(c, style, nil)

	if got := omitted.Height(); got != 30 {
		t.Fatalf("omitted.Height() = %v, want 30", got)
	}
	if got := auto.Height(); got != 20 {
		t.Fatalf("auto.Height() = %v, want 20", got)
	}
	if got := c.Height(); got != 60 {
		t.Fatalf("container.Height() = %v, want 60", got)
	}
}

func TestLayoutHBox_ContainerAlignBottomStillBottomAlignsChildrenByDefault(t *testing.T) {
	c := positionedContainer(0, 0, 300, 100)
	c.align = AlignBottom
	style := &LayoutStyle{}

	w := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	w.SetWidth(80)
	w.SetContainer(c)
	c.AddChild(w)

	LayoutHBox(c, style, nil)

	if got := w.Top(); got != 80 {
		t.Errorf("default hbox child top with container align bottom = %v, want 80", got)
	}
}

func TestLayoutHBox_AlignSelfEndBottomAlignsChild(t *testing.T) {
	c := positionedContainer(0, 0, 300, 100)
	style := &LayoutStyle{}

	w := &positionedTestWidget{preferredWidth: 80, preferredHeight: 20}
	w.SetWidth(80)
	w.SetAttrs(map[string]string{"align-self": "end"})
	w.SetContainer(c)
	c.AddChild(w)

	LayoutHBox(c, style, nil)

	if got := w.Top(); got != 80 {
		t.Errorf("hbox end-aligned child top = %v, want 80", got)
	}
}
