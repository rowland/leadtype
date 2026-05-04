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
	if got := auto1.Width(); got != 70 {
		t.Fatalf("auto1.Width() = %v, want 70", got)
	}
	if got := auto2.Width(); got != 70 {
		t.Fatalf("auto2.Width() = %v, want 70", got)
	}
}

func TestLayoutHBox_AutoWidthHasNoEffectWhenConstrained(t *testing.T) {
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

	for _, widget := range []*positionedTestWidget{omitted, auto1, auto2} {
		if got := widget.Width(); math.Abs(got-(175.0/3.0)) > 0.001 {
			t.Fatalf("widget.Width() = %v, want %v", got, 175.0/3.0)
		}
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
