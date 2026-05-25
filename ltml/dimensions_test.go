// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"math"
	"testing"
)

func TestDimensions_SetAttrs(t *testing.T) {
	tests := []struct {
		name               string
		attrs              map[string]string
		wantWidth          float64
		wantWidthValue     float64
		wantWidthMode      DimensionMode
		wantHeight         float64
		wantHeightValue    float64
		wantHeightMode     DimensionMode
		wantMaxWidthValue  float64
		wantMaxWidthMode   DimensionMode
		wantMaxHeightValue float64
		wantMaxHeightMode  DimensionMode
		wantWidthSet       bool
		wantHeightSet      bool
		wantMaxWidthSet    bool
		wantMaxHeightSet   bool
	}{
		{name: "Width", attrs: map[string]string{"width": "30"}, wantWidth: 30, wantWidthValue: 30, wantWidthMode: DimLiteral, wantWidthSet: true},
		{name: "WidthPct", attrs: map[string]string{"width": "40%"}, wantWidthValue: 40, wantWidthMode: DimPct, wantWidthSet: true},
		{name: "WidthRelPlus", attrs: map[string]string{"width": "+50"}, wantWidthValue: 50, wantWidthMode: DimRel, wantWidthSet: true},
		{name: "WidthRelWithUnits", attrs: map[string]string{"width": "+180pt"}, wantWidthValue: 180, wantWidthMode: DimRel, wantWidthSet: true},
		{name: "WidthRelMinus", attrs: map[string]string{"width": "-60"}, wantWidthValue: -60, wantWidthMode: DimRel, wantWidthSet: true},
		{name: "WidthAuto", attrs: map[string]string{"width": "auto"}, wantWidthMode: DimAuto},
		{name: "Height", attrs: map[string]string{"height": "30"}, wantHeight: 30, wantHeightValue: 30, wantHeightMode: DimLiteral, wantHeightSet: true},
		{name: "HeightPct", attrs: map[string]string{"height": "40%"}, wantHeightValue: 40, wantHeightMode: DimPct, wantHeightSet: true},
		{name: "HeightRelPlus", attrs: map[string]string{"height": "+50"}, wantHeightValue: 50, wantHeightMode: DimRel, wantHeightSet: true},
		{name: "HeightRelMinus", attrs: map[string]string{"height": "-60"}, wantHeightValue: -60, wantHeightMode: DimRel, wantHeightSet: true},
		{name: "HeightRelWithUnits", attrs: map[string]string{"height": "-0.25in"}, wantHeightValue: -18, wantHeightMode: DimRel, wantHeightSet: true},
		{name: "HeightAuto", attrs: map[string]string{"height": "auto"}, wantHeightMode: DimAuto},
		{name: "MaxWidth", attrs: map[string]string{"max-width": "30"}, wantMaxWidthValue: 30, wantMaxWidthMode: DimLiteral, wantMaxWidthSet: true},
		{name: "MaxWidthRelWithUnits", attrs: map[string]string{"max-width": "+10mm"}, wantMaxWidthValue: 10 * ptsPerMM, wantMaxWidthMode: DimRel, wantMaxWidthSet: true},
		{name: "MaxWidthAutoClears", attrs: map[string]string{"max-width": "auto"}},
		{name: "MaxHeight", attrs: map[string]string{"max-height": "40%"}, wantMaxHeightValue: 40, wantMaxHeightMode: DimPct, wantMaxHeightSet: true},
		{name: "MaxHeightRelBare", attrs: map[string]string{"max-height": "-12"}, wantMaxHeightValue: -12, wantMaxHeightMode: DimRel, wantMaxHeightSet: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var d Dimensions
			d.SetAttrs(tc.attrs, "")

			if got := float64(d.width); !closeEnough(got, tc.wantWidth) {
				t.Errorf("width: expected %v, got %v", tc.wantWidth, got)
			}
			if got := float64(d.widthValue); !closeEnough(got, tc.wantWidthValue) {
				t.Errorf("widthValue: expected %v, got %v", tc.wantWidthValue, got)
			}
			if got := d.WidthMode(); got != tc.wantWidthMode {
				t.Errorf("WidthMode: expected %v, got %v", tc.wantWidthMode, got)
			}
			if got := float64(d.height); !closeEnough(got, tc.wantHeight) {
				t.Errorf("height: expected %v, got %v", tc.wantHeight, got)
			}
			if got := float64(d.heightValue); !closeEnough(got, tc.wantHeightValue) {
				t.Errorf("heightValue: expected %v, got %v", tc.wantHeightValue, got)
			}
			if got := d.HeightMode(); got != tc.wantHeightMode {
				t.Errorf("HeightMode: expected %v, got %v", tc.wantHeightMode, got)
			}
			if got := d.WidthIsSet(); got != tc.wantWidthSet {
				t.Errorf("WidthIsSet: expected %v, got %v", tc.wantWidthSet, got)
			}
			if got := d.HeightIsSet(); got != tc.wantHeightSet {
				t.Errorf("HeightIsSet: expected %v, got %v", tc.wantHeightSet, got)
			}
			if got := d.MaxWidthIsSet(); got != tc.wantMaxWidthSet {
				t.Errorf("MaxWidthIsSet: expected %v, got %v", tc.wantMaxWidthSet, got)
			}
			if got := d.MaxHeightIsSet(); got != tc.wantMaxHeightSet {
				t.Errorf("MaxHeightIsSet: expected %v, got %v", tc.wantMaxHeightSet, got)
			}
			if got := float64(d.max.widthValue); !closeEnough(got, tc.wantMaxWidthValue) {
				t.Errorf("maxWidthValue: expected %v, got %v", tc.wantMaxWidthValue, got)
			}
			if got := d.max.widthMode; got != tc.wantMaxWidthMode {
				t.Errorf("maxWidthMode: expected %v, got %v", tc.wantMaxWidthMode, got)
			}
			if got := float64(d.max.heightValue); !closeEnough(got, tc.wantMaxHeightValue) {
				t.Errorf("maxHeightValue: expected %v, got %v", tc.wantMaxHeightValue, got)
			}
			if got := d.max.heightMode; got != tc.wantMaxHeightMode {
				t.Errorf("maxHeightMode: expected %v, got %v", tc.wantMaxHeightMode, got)
			}
		})
	}
}

func closeEnough(got, want float64) bool {
	return math.Abs(got-want) < 0.0001
}

func TestDimensions_SetAttrs_AutoIsCaseSensitive(t *testing.T) {
	var d Dimensions
	d.SetAttrs(map[string]string{"width": "AUTO", "height": "Auto"}, "")

	if got := d.WidthMode(); got == DimAuto {
		t.Fatalf("WidthMode() = %v, want non-auto for uppercase AUTO", got)
	}
	if got := d.HeightMode(); got == DimAuto {
		t.Fatalf("HeightMode() = %v, want non-auto for mixed-case Auto", got)
	}
}

func TestStdWidget_DimensionResolution(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 120}}

	pct := &StdWidget{}
	_ = pct.SetContainer(page)
	pct.SetWidthPct(25)
	pct.SetHeightPct(50)
	if got := pct.Width(); got != 50 {
		t.Fatalf("pct.Width() = %v, want 50", got)
	}
	if got := pct.Height(); got != 60 {
		t.Fatalf("pct.Height() = %v, want 60", got)
	}

	rel := &StdWidget{}
	_ = rel.SetContainer(page)
	rel.SetWidthRel(-20)
	rel.SetHeightRel(15)
	if got := rel.Width(); got != 180 {
		t.Fatalf("rel.Width() = %v, want 180", got)
	}
	if got := rel.Height(); got != 135 {
		t.Fatalf("rel.Height() = %v, want 135", got)
	}

	relUnits := &StdWidget{}
	_ = relUnits.SetContainer(page)
	relUnits.SetAttrs(map[string]string{"width": "+0.25in", "height": "-10pt"})
	if got := relUnits.Width(); got != 218 {
		t.Fatalf("relUnits.Width() = %v, want 218", got)
	}
	if got := relUnits.Height(); got != 110 {
		t.Fatalf("relUnits.Height() = %v, want 110", got)
	}

	auto := &StdWidget{}
	_ = auto.SetContainer(page)
	auto.SetWidthAuto()
	auto.SetHeightAuto()
	auto.SetLeft(10)
	auto.SetRight(-10)
	auto.SetTop(5)
	auto.SetBottom(-5)
	if got := auto.Width(); got != 180 {
		t.Fatalf("auto.Width() = %v, want 180", got)
	}
	if got := auto.Height(); got != 110 {
		t.Fatalf("auto.Height() = %v, want 110", got)
	}
}

func TestDetectTableColumnTracks_PreservesPercentClassification(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 120}}
	grid := NewWidgetGrid(2, 1)

	pct := &StdWidget{}
	_ = pct.SetContainer(page)
	pct.SetWidthPct(40)
	grid.SetCell(0, 0, pct)

	specified := &StdWidget{}
	_ = specified.SetContainer(page)
	specified.SetWidth(80)
	grid.SetCell(1, 0, specified)

	tracks := detectTableColumnTracks(grid, nil)
	if got := tracks[0].kind; got != tableTrackPercent {
		t.Fatalf("tracks[0].kind = %v, want %v", got, tableTrackPercent)
	}
	if got := tracks[0].size; got != 80 {
		t.Fatalf("tracks[0].size = %v, want 80", got)
	}
	if got := tracks[1].kind; got != tableTrackSpecified {
		t.Fatalf("tracks[1].kind = %v, want %v", got, tableTrackSpecified)
	}
	if got := tracks[1].size; got != 80 {
		t.Fatalf("tracks[1].size = %v, want 80", got)
	}
}

func TestDetectTableColumnTracks_ClassifiesAuto(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 120}}
	grid := NewWidgetGrid(1, 1)

	auto := &positionedTestWidget{preferredWidth: 35}
	_ = auto.SetContainer(page)
	auto.SetWidthAuto()
	grid.SetCell(0, 0, auto)

	tracks := detectTableColumnTracks(grid, nil)
	if got := tracks[0].kind; got != tableTrackAuto {
		t.Fatalf("tracks[0].kind = %v, want %v", got, tableTrackAuto)
	}
	if got := tracks[0].preferred; got != 35 {
		t.Fatalf("tracks[0].preferred = %v, want 35", got)
	}
}

func TestStdIndex_ClearMeasuredGeometry_ClearsOnlyImplicitDimensions(t *testing.T) {
	index := &StdIndex{}
	index.ResolveWidth(140)
	index.ResolveHeight(90)
	index.clearMeasuredGeometry()

	if index.width != 0 || index.widthValue != 0 || index.widthMode != DimUnspecified || index.widthValid {
		t.Fatalf("implicit width not cleared: width=%v value=%v mode=%v valid=%v", index.width, index.widthValue, index.widthMode, index.widthValid)
	}
	if index.height != 0 || index.heightValue != 0 || index.heightMode != DimUnspecified || index.heightValid {
		t.Fatalf("implicit height not cleared: height=%v value=%v mode=%v valid=%v", index.height, index.heightValue, index.heightMode, index.heightValid)
	}

	index.SetAttrs(map[string]string{"width": "40%", "height": "30", "units": "pt"})
	index.ResolveWidth(160)
	index.clearMeasuredGeometry()

	if index.width != 0 || index.widthValue != 40 || index.widthMode != DimPct || index.widthValid {
		t.Fatalf("explicit width was not preserved: width=%v value=%v mode=%v valid=%v", index.width, index.widthValue, index.widthMode, index.widthValid)
	}
	if index.height != 30 || index.heightValue != 30 || index.heightMode != DimLiteral || index.heightValid {
		t.Fatalf("explicit height was not preserved: height=%v value=%v mode=%v valid=%v", index.height, index.heightValue, index.heightMode, index.heightValid)
	}
}

func TestDimensions_SaveStateAndClearHelpers(t *testing.T) {
	var d Dimensions
	d.SetWidthPct(40)
	d.SetHeight(24)

	saved := d.SaveState()
	d.ClearWidth()
	d.ClearHeight()

	if d.WidthIsSet() || d.HeightIsSet() {
		t.Fatalf("dimensions should be cleared, got widthMode=%v heightMode=%v", d.widthMode, d.heightMode)
	}

	d.RestoreState(saved)

	if d.widthMode != DimPct || d.widthValue != 40 || d.width != 0 {
		t.Fatalf("width state restore failed: mode=%v value=%v width=%v", d.widthMode, d.widthValue, d.width)
	}
	if d.heightMode != DimLiteral || d.heightValue != 24 || d.height != 24 {
		t.Fatalf("height state restore failed: mode=%v value=%v height=%v", d.heightMode, d.heightValue, d.height)
	}
}

func TestDimensions_AspectInferredStateSavesRestoresAndClears(t *testing.T) {
	var d Dimensions
	d.ResolveAspectWidth(80)
	d.ResolveAspectHeight(20)

	if !d.WidthIsSet() || !d.HeightIsSet() {
		t.Fatalf("aspect-inferred dimensions should be set")
	}
	if !d.WidthAspectInferred() || !d.HeightAspectInferred() {
		t.Fatalf("aspect-inferred flags not set")
	}

	saved := d.SaveState()
	d.ResolveWidth(100)
	d.ResolveHeight(25)
	if d.WidthAspectInferred() || d.HeightAspectInferred() {
		t.Fatalf("ordinary resolved dimensions should clear aspect-inferred flags")
	}

	d.RestoreState(saved)
	if !d.WidthAspectInferred() || !d.HeightAspectInferred() {
		t.Fatalf("aspect-inferred flags not restored")
	}

	d.ClearResolvedWidth()
	d.ClearResolvedHeight()
	if d.WidthAspectInferred() || d.HeightAspectInferred() || d.WidthIsSet() || d.HeightIsSet() {
		t.Fatalf("clear resolved should clear aspect-inferred dimensions")
	}
}

func TestWidgetSpecifiedHelpersTreatOnlyAspectResolvedDimensionsAsSpecified(t *testing.T) {
	widget := &StdWidget{}

	widget.ResolveWidth(80)
	widget.ResolveHeight(20)
	if widgetWidthSpecified(widget) || widgetHeightSpecified(widget) {
		t.Fatalf("ordinary resolved dimensions should not be treated as specified")
	}

	widget.ResolveAspectWidth(80)
	widget.ResolveAspectHeight(20)
	if !widgetWidthSpecified(widget) || !widgetHeightSpecified(widget) {
		t.Fatalf("aspect-inferred dimensions should be treated as specified")
	}
}

func TestStdWidget_ResolveWidthPreservesSpecifiedModeAndOverridesUntilCleared(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 120}}
	widget := &StdWidget{}
	_ = widget.SetContainer(page)
	widget.SetWidthPct(25)

	if got := widget.Width(); got != 50 {
		t.Fatalf("Width() before resolve = %v, want 50", got)
	}

	widget.ResolveWidth(80)
	page.pageStyle.width = 400
	if got := widget.Width(); got != 80 {
		t.Fatalf("Width() after resolve = %v, want 80", got)
	}
	if got := widget.WidthMode(); got != DimPct {
		t.Fatalf("WidthMode() after resolve = %v, want %v", got, DimPct)
	}
	if got := widget.widthValue; got != 25 {
		t.Fatalf("widthValue after resolve = %v, want 25", got)
	}

	widget.ClearResolvedWidth()
	if got := widget.Width(); got != 100 {
		t.Fatalf("Width() after clear = %v, want 100", got)
	}
}

func TestStdWidget_ResolveAutoWidthOverridesSideResolutionUntilCleared(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 120}}
	widget := &StdWidget{}
	_ = widget.SetContainer(page)
	widget.SetWidthAuto()
	widget.SetLeft(10)
	widget.SetRight(-10)

	if got := widget.Width(); got != 180 {
		t.Fatalf("Width() before resolve = %v, want 180", got)
	}

	widget.ResolveWidth(90)
	widget.SetRight(-40)
	if got := widget.Width(); got != 90 {
		t.Fatalf("Width() after resolve = %v, want 90", got)
	}
	if got := widget.WidthIsSet(); !got {
		t.Fatalf("WidthIsSet() after resolve = %v, want true", got)
	}

	widget.ClearResolvedWidth()
	if got := widget.Width(); got != 150 {
		t.Fatalf("Width() after clear = %v, want 150", got)
	}
	if got := widget.WidthIsSet(); !got {
		t.Fatalf("WidthIsSet() after clear = %v, want true from left/right anchors", got)
	}
}

func TestStdWidget_MaxDimensionsCapResolvedAndContainerRelativeSizes(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 120}}
	widget := &StdWidget{}
	_ = widget.SetContainer(page)
	widget.SetWidthPct(75)
	widget.SetHeightRel(-10)
	widget.SetMaxWidth(80)
	widget.SetMaxHeightPct(50)

	if got := widget.MaxWidth(); got != 80 {
		t.Fatalf("MaxWidth() = %v, want 80", got)
	}
	if got := widget.MaxHeight(); got != 60 {
		t.Fatalf("MaxHeight() = %v, want 60", got)
	}
	if got := widget.Width(); got != 80 {
		t.Fatalf("Width() = %v, want capped 80", got)
	}
	if got := widget.Height(); got != 60 {
		t.Fatalf("Height() = %v, want capped 60", got)
	}

	widget.ClearMaxWidth()
	widget.ClearMaxHeight()
	if widget.MaxWidthIsSet() || widget.MaxHeightIsSet() {
		t.Fatalf("max dimensions should be clear")
	}
	if got := widget.Width(); got != 150 {
		t.Fatalf("Width() after clear = %v, want 150", got)
	}
	if got := widget.Height(); got != 110 {
		t.Fatalf("Height() after clear = %v, want 110", got)
	}
}

func TestDimensions_ClearResolvedHeightPreservesSpecifiedHeight(t *testing.T) {
	var d Dimensions
	d.SetHeight(24)
	d.ResolveHeight(36)
	d.ClearResolvedHeight()

	if got := d.HeightMode(); got != DimLiteral {
		t.Fatalf("HeightMode() = %v, want %v", got, DimLiteral)
	}
	if got := d.height; got != 24 {
		t.Fatalf("height = %v, want 24", got)
	}
	if got := d.HeightIsSet(); !got {
		t.Fatalf("HeightIsSet() = %v, want true", got)
	}
}
