// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"testing"
)

func TestDimensions_SetAttrs(t *testing.T) {
	tests := []struct {
		name            string
		attrs           map[string]string
		wantWidth       float64
		wantWidthValue  float64
		wantWidthMode   DimensionMode
		wantHeight      float64
		wantHeightValue float64
		wantHeightMode  DimensionMode
		wantWidthSet    bool
		wantHeightSet   bool
	}{
		{name: "Width", attrs: map[string]string{"width": "30"}, wantWidth: 30, wantWidthValue: 30, wantWidthMode: DimLiteral, wantWidthSet: true},
		{name: "WidthPct", attrs: map[string]string{"width": "40%"}, wantWidthValue: 40, wantWidthMode: DimPct, wantWidthSet: true},
		{name: "WidthRelPlus", attrs: map[string]string{"width": "+50"}, wantWidthValue: 50, wantWidthMode: DimRel, wantWidthSet: true},
		{name: "WidthRelMinus", attrs: map[string]string{"width": "-60"}, wantWidthValue: -60, wantWidthMode: DimRel, wantWidthSet: true},
		{name: "WidthAuto", attrs: map[string]string{"width": "auto"}, wantWidthMode: DimAuto},
		{name: "Height", attrs: map[string]string{"height": "30"}, wantHeight: 30, wantHeightValue: 30, wantHeightMode: DimLiteral, wantHeightSet: true},
		{name: "HeightPct", attrs: map[string]string{"height": "40%"}, wantHeightValue: 40, wantHeightMode: DimPct, wantHeightSet: true},
		{name: "HeightRelPlus", attrs: map[string]string{"height": "+50"}, wantHeightValue: 50, wantHeightMode: DimRel, wantHeightSet: true},
		{name: "HeightRelMinus", attrs: map[string]string{"height": "-60"}, wantHeightValue: -60, wantHeightMode: DimRel, wantHeightSet: true},
		{name: "HeightAuto", attrs: map[string]string{"height": "auto"}, wantHeightMode: DimAuto},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var d Dimensions
			d.SetAttrs(tc.attrs, "")

			if got := float64(d.width); got != tc.wantWidth {
				t.Errorf("width: expected %v, got %v", tc.wantWidth, got)
			}
			if got := float64(d.widthValue); got != tc.wantWidthValue {
				t.Errorf("widthValue: expected %v, got %v", tc.wantWidthValue, got)
			}
			if got := d.WidthMode(); got != tc.wantWidthMode {
				t.Errorf("WidthMode: expected %v, got %v", tc.wantWidthMode, got)
			}
			if got := float64(d.height); got != tc.wantHeight {
				t.Errorf("height: expected %v, got %v", tc.wantHeight, got)
			}
			if got := float64(d.heightValue); got != tc.wantHeightValue {
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
		})
	}
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

func TestDetectWidths_PreservesPercentClassification(t *testing.T) {
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

	widths := detectWidths(grid, nil)
	if got := widths[0].How; got != Percent {
		t.Fatalf("widths[0].How = %v, want %v", got, Percent)
	}
	if got := widths[0].Size; got != 80 {
		t.Fatalf("widths[0].Size = %v, want 80", got)
	}
	if got := widths[1].How; got != Specified {
		t.Fatalf("widths[1].How = %v, want %v", got, Specified)
	}
	if got := widths[1].Size; got != 80 {
		t.Fatalf("widths[1].Size = %v, want 80", got)
	}
}

func TestDetectWidths_TreatsAutoAsUnspecified(t *testing.T) {
	page := &StdPage{pageStyle: &PageStyle{width: 200, height: 120}}
	grid := NewWidgetGrid(1, 1)

	auto := &StdWidget{}
	_ = auto.SetContainer(page)
	auto.SetWidthAuto()
	grid.SetCell(0, 0, auto)

	widths := detectWidths(grid, nil)
	if got := widths[0].How; got != Unspecified {
		t.Fatalf("widths[0].How = %v, want %v", got, Unspecified)
	}
	if got := widths[0].Size; got != 0 {
		t.Fatalf("widths[0].Size = %v, want 0", got)
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
