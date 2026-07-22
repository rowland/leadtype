// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"testing"

	"github.com/rowland/leadtype/pdf"
)

func TestDisplayMode_ParseAndString(t *testing.T) {
	tests := []struct {
		input string
		want  DisplayMode
		text  string
	}{
		{input: "once", want: DisplayOnce, text: "once"},
		{input: "always", want: DisplayAlways, text: "always"},
		{input: "first", want: DisplayFirst, text: "first"},
		{input: "succeeding", want: DisplaySucceeding, text: "succeeding"},
		{input: "even", want: DisplayEven, text: "even"},
		{input: "odd", want: DisplayOdd, text: "odd"},
		{input: "last", want: DisplayLast, text: "last"},
		{input: "none", want: DisplayNone, text: "none"},
		{input: "  LAST  ", want: DisplayLast, text: "last"},
		{input: "unknown", want: DisplayOnce, text: "once"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := ParseDisplayMode(tc.input); got != tc.want {
				t.Fatalf("ParseDisplayMode(%q) = %v, want %v", tc.input, got, tc.want)
			}
			if got := tc.want.String(); got != tc.text {
				t.Fatalf("%v.String() = %q, want %q", tc.want, got, tc.text)
			}
		})
	}
}

func TestStdWidget_SetAttrs_ParsesSideSpecificBorders(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)
	widget := &StdWidget{}
	widget.SetScope(scope)

	widget.SetAttrs(map[string]string{"border-right": "dashed"})

	if widget.borders[rightSide] == nil {
		t.Fatal("right border is nil, want parsed pen style")
	}
	if got := widget.borders[rightSide].pattern; got != "dashed" {
		t.Fatalf("right border pattern = %q, want dashed", got)
	}
}

func TestStdWidget_SetAttrs_ParsesExactLowercaseNoneForBorders(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)
	widget := &StdWidget{}
	widget.SetScope(scope)

	widget.SetAttrs(map[string]string{"border": "solid", "border-top": "dashed"})
	widget.SetAttrs(map[string]string{"border": "  none  ", "border-top": "none"})
	if widget.border != nil || !widget.borderSet {
		t.Fatalf("aggregate border = %#v set=%v, want explicit none", widget.border, widget.borderSet)
	}
	if widget.borders[topSide] != nil || !widget.borderSideSet[topSide] {
		t.Fatalf("top border = %#v set=%v, want explicit none", widget.borders[topSide], widget.borderSideSet[topSide])
	}
	if _, ok := scope.StyleFor("pen_none"); ok {
		t.Fatal("lowercase none unexpectedly registered a pen style")
	}

	widget.SetAttrs(map[string]string{"border.color": "Red", "border-top.width": "3pt"})
	if widget.border != nil || widget.borders[topSide] != nil {
		t.Fatalf("subattributes revived disabled borders: aggregate=%#v top=%#v", widget.border, widget.borders[topSide])
	}

	widget.SetAttrs(map[string]string{"border": "dotted", "border-top": "solid"})
	if widget.border == nil || widget.borders[topSide] == nil {
		t.Fatalf("explicit pens did not revive borders: aggregate=%#v top=%#v", widget.border, widget.borders[topSide])
	}
}

func TestStdWidget_SetAttrs_NoneIsCaseSensitive(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)
	widget := &StdWidget{}
	widget.SetScope(scope)

	widget.SetAttrs(map[string]string{"border": "None", "border-left": "NONE"})
	if widget.border == nil || widget.borders[leftSide] == nil {
		t.Fatalf("differently cased pen names were treated as none: aggregate=%#v left=%#v", widget.border, widget.borders[leftSide])
	}
}

func TestStdWidget_DrawBorder_SideNoneSuppressesAggregateEdge(t *testing.T) {
	widget := &StdWidget{}
	widget.SetLeft(10)
	widget.SetTop(20)
	widget.SetWidth(120)
	widget.SetHeight(40)
	widget.SetAttrs(map[string]string{"border": "solid", "border-top": "none"})
	w := &shapeTestWriter{labelTestWriter: labelTestWriter{t: t}}

	if err := widget.DrawBorder(w); err != nil {
		t.Fatal(err)
	}
	if w.pathRuns != 1 || w.strokes != 1 || len(w.rectPages) != 0 {
		t.Fatalf("edge drawing paths/strokes/rectangles = %d/%d/%d, want 1/1/0", w.pathRuns, w.strokes, len(w.rectPages))
	}
}

func TestStdWidget_DrawBorder_SideCanOverrideDisabledAggregate(t *testing.T) {
	widget := &StdWidget{}
	widget.SetWidth(100)
	widget.SetHeight(30)
	widget.SetAttrs(map[string]string{"border": "none", "border-bottom": "solid"})
	w := &shapeTestWriter{labelTestWriter: labelTestWriter{t: t}}

	if err := widget.DrawBorder(w); err != nil {
		t.Fatal(err)
	}
	if w.pathRuns != 1 || w.strokes != 1 || len(w.rectPages) != 0 {
		t.Fatalf("edge drawing paths/strokes/rectangles = %d/%d/%d, want 1/1/0", w.pathRuns, w.strokes, len(w.rectPages))
	}
}

func TestStdWidget_DrawBorder_ClosedAggregateRetainsRoundedRectanglePath(t *testing.T) {
	widget := &StdWidget{}
	widget.SetWidth(100)
	widget.SetHeight(30)
	widget.SetAttrs(map[string]string{"border": "solid", "corners": "5pt"})
	w := &shapeTestWriter{labelTestWriter: labelTestWriter{t: t}}

	if err := widget.DrawBorder(w); err != nil {
		t.Fatal(err)
	}
	if w.pathRuns != 0 || len(w.rectPages) != 1 {
		t.Fatalf("aggregate drawing paths/rectangles = %d/%d, want 0/1", w.pathRuns, len(w.rectPages))
	}
}

func TestStdWidget_DrawBorder_SideOverridesPreserveCornersAndGroupMatchingEdges(t *testing.T) {
	widget := &StdWidget{}
	widget.SetWidth(100)
	widget.SetHeight(30)
	widget.SetAttrs(map[string]string{"border": "solid", "border-top": "none", "corners": "5pt"})
	w := &shapeTestWriter{labelTestWriter: labelTestWriter{t: t}}

	if err := widget.DrawBorder(w); err != nil {
		t.Fatal(err)
	}
	if w.pathRuns != 1 || w.strokes != 1 {
		t.Fatalf("matching edge paths/strokes = %d/%d, want 1/1", w.pathRuns, w.strokes)
	}
	if w.curves != 4 {
		t.Fatalf("rounded corner curves = %d, want 4", w.curves)
	}
}

func TestStdWidget_DrawBorder_ExplicitMatchingSideKeepsClosedRoundedPath(t *testing.T) {
	widget := &StdWidget{}
	widget.SetWidth(100)
	widget.SetHeight(30)
	widget.SetAttrs(map[string]string{"border": "solid", "border-top": "solid", "corners": "5pt"})
	w := &shapeTestWriter{labelTestWriter: labelTestWriter{t: t}}

	if err := widget.DrawBorder(w); err != nil {
		t.Fatal(err)
	}
	if w.pathRuns != 0 || len(w.rectPages) != 1 {
		t.Fatalf("matching closed border paths/rectangles = %d/%d, want 0/1", w.pathRuns, len(w.rectPages))
	}
}

func TestStdWidget_DrawBorder_DifferentPensSplitRoundedCornersBetweenRuns(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)
	widget := &StdWidget{}
	widget.SetScope(scope)
	widget.SetWidth(100)
	widget.SetHeight(30)
	widget.SetAttrs(map[string]string{"border": "solid", "border-bottom": "dashed", "corners": "5pt"})
	w := &shapeTestWriter{labelTestWriter: labelTestWriter{t: t}}

	if err := widget.DrawBorder(w); err != nil {
		t.Fatal(err)
	}
	if w.pathRuns != 2 || w.strokes != 2 || len(w.rectPages) != 0 {
		t.Fatalf("different-pen paths/strokes/rectangles = %d/%d/%d, want 2/2/0", w.pathRuns, w.strokes, len(w.rectPages))
	}
	if w.curves != 6 {
		t.Fatalf("whole and split corner curves = %d, want 6", w.curves)
	}
}

func TestParse_BorderNoneAppliesToPageContainerAndShape(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page border="none">
    <vbox border="none">
      <circle border="none" width="20" height="20" />
    </vbox>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}
	page := doc.Root().Page(0)
	container := page.children[0].(*StdContainer)
	shape := container.children[0].(*StdCircle)
	for name, widget := range map[string]*StdWidget{
		"page": &page.StdWidget, "container": &container.StdWidget, "shape": &shape.StdWidget,
	} {
		if !widget.borderSet || widget.border != nil {
			t.Fatalf("%s border = %#v set=%v, want explicit none", name, widget.border, widget.borderSet)
		}
	}
}

func TestStdWidget_SetAttrs_ClonesBorderForBorderPrefixOverrides(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)

	widget := &StdWidget{}
	widget.SetScope(scope)
	widget.SetAttrs(map[string]string{
		"border":       "dashed",
		"border.color": "red",
	})

	if widget.border == nil {
		t.Fatal("border is nil, want cloned pen style")
	}
	base := PenStyleFor("dashed", scope)
	if widget.border == base {
		t.Fatal("border reused shared pen style, want clone")
	}
	if got := widget.border.color; got != NamedColor("red") {
		t.Fatalf("border color = %v, want red", got)
	}
	if got := base.color; got == NamedColor("red") {
		t.Fatalf("shared dashed pen color = %v, want unchanged", got)
	}
}

func TestStdWidget_SetAttrs_ClonesSideBorderForSidePrefixOverrides(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)

	widget := &StdWidget{}
	widget.SetScope(scope)
	widget.SetAttrs(map[string]string{
		"border-right":       "dashed",
		"border-right.color": "red",
	})

	if widget.borders[rightSide] == nil {
		t.Fatal("right border is nil, want cloned pen style")
	}
	base := PenStyleFor("dashed", scope)
	if widget.borders[rightSide] == base {
		t.Fatal("right border reused shared dashed pen, want clone")
	}
	if got := widget.borders[rightSide].color; got != NamedColor("red") {
		t.Fatalf("right border color = %v, want red", got)
	}
	if got := base.color; got == NamedColor("red") {
		t.Fatalf("shared dashed pen color = %v, want unchanged", got)
	}
}

func TestStdWidget_SetAttrs_SideBorderPrefixOverridesCloneMainBorderWhenNeeded(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)

	widget := &StdWidget{}
	widget.SetScope(scope)
	widget.SetAttrs(map[string]string{
		"border":           "dashed",
		"border-top.color": "red",
	})

	if widget.border == nil {
		t.Fatal("main border is nil, want parsed pen style")
	}
	if widget.borders[topSide] == nil {
		t.Fatal("top border is nil, want derived clone")
	}
	if widget.borders[topSide] == widget.border {
		t.Fatal("top border reused main border, want clone")
	}
	if got := widget.borders[topSide].color; got != NamedColor("red") {
		t.Fatalf("top border color = %v, want red", got)
	}
	if got := widget.border.color; got == NamedColor("red") {
		t.Fatalf("main border color = %v, want unchanged", got)
	}
	if got := widget.borders[topSide].pattern; got != widget.border.pattern {
		t.Fatalf("top border pattern = %q, want %q", got, widget.border.pattern)
	}
}

func TestStdWidget_SetAttrs_FontDecorationOverridesCloneAndApplyFontOverrides(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)
	if err := scope.AddStyle(&PenStyle{id: "accent", color: NamedColor("red"), width: 2, pattern: "dotted", cap: "round_cap"}); err != nil {
		t.Fatal(err)
	}

	widget := &StdWidget{}
	widget.SetScope(scope)
	widget.font = defaultFont.Clone()
	widget.font.SetScope(scope)

	widget.SetAttrs(map[string]string{
		"font.units":         "cm",
		"font.underline-pen": "accent",
		"font.underline-pos": "1",
	})

	if widget.font == nil {
		t.Fatal("font is nil, want cloned font style")
	}
	if widget.font.underlinePenID != "accent" {
		t.Fatalf("underlinePenID = %q, want accent", widget.font.underlinePenID)
	}
	if widget.font.underlinePos == nil || *widget.font.underlinePos != FromUnits(1, "cm") {
		t.Fatalf("underlinePos = %v, want 1cm in points", widget.font.underlinePos)
	}
}

func TestStdWidget_DrawBorder_AppliesGradientPenInWidgetBox(t *testing.T) {
	widget := &StdWidget{}
	widget.SetLeft(10)
	widget.SetTop(20)
	widget.SetWidth(120)
	widget.SetHeight(40)
	widget.border = &PenStyle{
		kind: PenKindLinearGradient,
		linearGradient: &pdf.LinearGradient{
			Stops: []pdf.GradientStop{
				{Position: 0, Color: NamedColor("Tomato")},
				{Position: 1, Color: NamedColor("SteelBlue")},
			},
		},
		linearPct: &linearGradientPct{X0: float64Ptr(0), Y0: float64Ptr(50), X1: float64Ptr(100), Y1: float64Ptr(50)},
	}
	writer := &labelTestWriter{}

	if err := widget.DrawBorder(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.lineLinear) != 1 {
		t.Fatalf("line linear gradient count = %d, want 1", len(writer.lineLinear))
	}
	got := writer.lineLinear[0]
	if got.X0 != 10 || got.Y0 != 40 || got.X1 != 130 || got.Y1 != 40 {
		t.Fatalf("gradient coords = %#v, want widget-local coords", got)
	}
}

func TestStdWidget_DrawBorder_AppliesGradientSidePenInWidgetBox(t *testing.T) {
	widget := &StdWidget{}
	widget.SetLeft(10)
	widget.SetTop(20)
	widget.SetWidth(120)
	widget.SetHeight(40)
	widget.border = &PenStyle{pattern: "solid"}
	widget.borders[rightSide] = &PenStyle{
		kind: PenKindLinearGradient,
		linearGradient: &pdf.LinearGradient{
			Stops: []pdf.GradientStop{
				{Position: 0, Color: NamedColor("Tomato")},
				{Position: 1, Color: NamedColor("SteelBlue")},
			},
		},
		linearPct: &linearGradientPct{X0: float64Ptr(0), Y0: float64Ptr(0), X1: float64Ptr(0), Y1: float64Ptr(100)},
	}
	widget.borderSideSet[rightSide] = true
	writer := &shapeTestWriter{labelTestWriter: labelTestWriter{t: t}}

	if err := widget.DrawBorder(writer); err != nil {
		t.Fatal(err)
	}
	if len(writer.lineLinear) != 1 {
		t.Fatalf("line linear gradient count = %d, want 1", len(writer.lineLinear))
	}
	if writer.pathRuns != 2 || writer.strokes != 2 || len(writer.rectPages) != 0 {
		t.Fatalf("edge drawing paths/strokes/rectangles = %d/%d/%d, want 2/2/0", writer.pathRuns, writer.strokes, len(writer.rectPages))
	}
	got := writer.lineLinear[0]
	if got.X0 != 10 || got.Y0 != 20 || got.X1 != 10 || got.Y1 != 60 {
		t.Fatalf("gradient coords = %#v, want side border coords resolved against widget box", got)
	}
}

func TestStdWidget_SetAttrs_FontDecorationOverridesPreserveInheritedDecorationOverrides(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)

	base := defaultFont.Clone()
	base.SetScope(scope)
	base.SetAttrs(map[string]string{
		"underline-pos": "4pt",
		"strikeout-pos": "7pt",
		"strikeout-pen": "dashed",
	})

	widget := &StdWidget{}
	widget.SetScope(scope)
	widget.font = base

	widget.SetAttrs(map[string]string{
		"font.underline-pen": "dotted",
	})

	if widget.font == nil {
		t.Fatal("font is nil, want cloned font style")
	}
	if widget.font.underlinePenID != "dotted" {
		t.Fatalf("underlinePenID = %q, want dotted", widget.font.underlinePenID)
	}
	if widget.font.underlinePos == nil || *widget.font.underlinePos != 4 {
		t.Fatalf("underlinePos = %v, want inherited 4pt", widget.font.underlinePos)
	}
	if widget.font.strikeoutPos == nil || *widget.font.strikeoutPos != 7 {
		t.Fatalf("strikeoutPos = %v, want inherited 7pt", widget.font.strikeoutPos)
	}
	if widget.font.strikeoutPenID != "dashed" {
		t.Fatalf("strikeoutPenID = %q, want inherited dashed", widget.font.strikeoutPenID)
	}
}

func TestStdWidget_SetAttrs_ParsesLogicalOrigins(t *testing.T) {
	widget := &StdWidget{}
	widget.SetAttrs(map[string]string{
		"origin-x": "end",
		"origin-y": "middle",
	})

	if got := widget.originX; got != OriginXEnd {
		t.Fatalf("originX = %v, want %v", got, OriginXEnd)
	}
	if got := widget.originY; got != OriginYMiddle {
		t.Fatalf("originY = %v, want %v", got, OriginYMiddle)
	}
	if got := widget.OriginX(); got != OriginXEnd {
		t.Fatalf("OriginX() = %v, want %v", got, OriginXEnd)
	}
	if got := widget.OriginY(); got != OriginYMiddle {
		t.Fatalf("OriginY() = %v, want %v", got, OriginYMiddle)
	}
}

func TestStdWidget_SetAttrs_DefaultOriginsRemainUnspecified(t *testing.T) {
	widget := &StdWidget{}
	widget.SetAttrs(map[string]string{})

	if got := widget.originX; got != OriginXUnspecified {
		t.Fatalf("originX = %v, want %v", got, OriginXUnspecified)
	}
	if got := widget.originY; got != OriginYUnspecified {
		t.Fatalf("originY = %v, want %v", got, OriginYUnspecified)
	}
}

func TestStdWidget_SetAttrs_ParsesRadialVerticalOriginAliases(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  OriginY
	}{
		{name: "inner", token: "inner", want: OriginYTop},
		{name: "outer", token: "outer", want: OriginYBottom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			widget := &StdWidget{}
			widget.SetAttrs(map[string]string{
				"origin-y": tt.token,
			})

			if got := widget.originY; got != tt.want {
				t.Fatalf("originY = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStdWidget_SetAttrs_ParsesExplicitOriginMeasurements(t *testing.T) {
	widget := &StdWidget{}
	widget.SetAttrs(map[string]string{
		"units":    "in",
		"origin-x": "1.5",
		"origin-y": "0.25",
	})

	if got := widget.originX; got != OriginXCustom {
		t.Fatalf("originX = %v, want %v", got, OriginXCustom)
	}
	if got := widget.originY; got != OriginYCustom {
		t.Fatalf("originY = %v, want %v", got, OriginYCustom)
	}
	if got := widget.originXValue; got != 108 {
		t.Fatalf("originXValue = %v, want 108", got)
	}
	if got := widget.originYValue; got != 18 {
		t.Fatalf("originYValue = %v, want 18", got)
	}
	if got := widget.OriginY(); got != OriginYCustom {
		t.Fatalf("OriginY() = %v, want %v", got, OriginYCustom)
	}
}
