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
		{name: "Width", attrs: map[string]string{"width": "30"}, wantWidth: 30, wantWidthValue: 30, wantWidthMode: DimSpecified, wantWidthSet: true},
		{name: "WidthPct", attrs: map[string]string{"width": "40%"}, wantWidthValue: 40, wantWidthMode: DimPct, wantWidthSet: true},
		{name: "WidthRelPlus", attrs: map[string]string{"width": "+50"}, wantWidthValue: 50, wantWidthMode: DimRel, wantWidthSet: true},
		{name: "WidthRelMinus", attrs: map[string]string{"width": "-60"}, wantWidthValue: -60, wantWidthMode: DimRel, wantWidthSet: true},
		{name: "Height", attrs: map[string]string{"height": "30"}, wantHeight: 30, wantHeightValue: 30, wantHeightMode: DimSpecified, wantHeightSet: true},
		{name: "HeightPct", attrs: map[string]string{"height": "40%"}, wantHeightValue: 40, wantHeightMode: DimPct, wantHeightSet: true},
		{name: "HeightRelPlus", attrs: map[string]string{"height": "+50"}, wantHeightValue: 50, wantHeightMode: DimRel, wantHeightSet: true},
		{name: "HeightRelMinus", attrs: map[string]string{"height": "-60"}, wantHeightValue: -60, wantHeightMode: DimRel, wantHeightSet: true},
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
			if d.widthMode != tc.wantWidthMode {
				t.Errorf("widthMode: expected %v, got %v", tc.wantWidthMode, d.widthMode)
			}
			if got := float64(d.height); got != tc.wantHeight {
				t.Errorf("height: expected %v, got %v", tc.wantHeight, got)
			}
			if got := float64(d.heightValue); got != tc.wantHeightValue {
				t.Errorf("heightValue: expected %v, got %v", tc.wantHeightValue, got)
			}
			if d.heightMode != tc.wantHeightMode {
				t.Errorf("heightMode: expected %v, got %v", tc.wantHeightMode, d.heightMode)
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

func TestStdIndex_ClearMeasuredGeometry_ClearsOnlyImplicitDimensions(t *testing.T) {
	index := &StdIndex{}
	index.width = 140
	index.widthValue = 25
	index.widthMode = DimPct
	index.height = 90
	index.heightValue = 12
	index.heightMode = DimRel
	index.clearMeasuredGeometry()

	if index.width != 0 || index.widthValue != 0 || index.widthMode != DimUnspecified {
		t.Fatalf("implicit width not cleared: width=%v value=%v mode=%v", index.width, index.widthValue, index.widthMode)
	}
	if index.height != 0 || index.heightValue != 0 || index.heightMode != DimUnspecified {
		t.Fatalf("implicit height not cleared: height=%v value=%v mode=%v", index.height, index.heightValue, index.heightMode)
	}

	index.SetAttrs(map[string]string{"width": "40%", "height": "30"})
	index.width = 160
	index.height = 30
	index.clearMeasuredGeometry()

	if index.width != 160 || index.widthValue != 40 || index.widthMode != DimPct {
		t.Fatalf("explicit width was not preserved: width=%v value=%v mode=%v", index.width, index.widthValue, index.widthMode)
	}
	if index.height != 30 || index.heightValue != 30 || index.heightMode != DimSpecified {
		t.Fatalf("explicit height was not preserved: height=%v value=%v mode=%v", index.height, index.heightValue, index.heightMode)
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
	if d.heightMode != DimSpecified || d.heightValue != 24 || d.height != 24 {
		t.Fatalf("height state restore failed: mode=%v value=%v height=%v", d.heightMode, d.heightValue, d.height)
	}
}
