// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"testing"
)

func TestDimensions_SetAttrs(t *testing.T) {
	tests := []struct {
		name          string
		attrs         map[string]string
		wantWidth     float64
		wantWidthPct  float64
		wantWidthRel  float64
		wantHeight    float64
		wantHeightPct float64
		wantHeightRel float64
		wantWidthSet  bool
		wantHeightSet bool
	}{
		{name: "Width", attrs: map[string]string{"width": "30"}, wantWidth: 30, wantWidthSet: true},
		{name: "WidthPct", attrs: map[string]string{"width": "40%"}, wantWidthPct: 40, wantWidthSet: true},
		{name: "WidthRelPlus", attrs: map[string]string{"width": "+50"}, wantWidthRel: 50, wantWidthSet: true},
		{name: "WidthRelMinus", attrs: map[string]string{"width": "-60"}, wantWidthRel: -60, wantWidthSet: true},
		{name: "Height", attrs: map[string]string{"height": "30"}, wantHeight: 30, wantHeightSet: true},
		{name: "HeightPct", attrs: map[string]string{"height": "40%"}, wantHeightPct: 40, wantHeightSet: true},
		{name: "HeightRelPlus", attrs: map[string]string{"height": "+50"}, wantHeightRel: 50, wantHeightSet: true},
		{name: "HeightRelMinus", attrs: map[string]string{"height": "-60"}, wantHeightRel: -60, wantHeightSet: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var d Dimensions
			d.SetAttrs(tc.attrs, "")

			if got := float64(d.width); got != tc.wantWidth {
				t.Errorf("width: expected %v, got %v", tc.wantWidth, got)
			}
			if got := float64(d.widthPct); got != tc.wantWidthPct {
				t.Errorf("widthPct: expected %v, got %v", tc.wantWidthPct, got)
			}
			if got := float64(d.widthRel); got != tc.wantWidthRel {
				t.Errorf("widthRel: expected %v, got %v", tc.wantWidthRel, got)
			}
			if got := float64(d.height); got != tc.wantHeight {
				t.Errorf("height: expected %v, got %v", tc.wantHeight, got)
			}
			if got := float64(d.heightPct); got != tc.wantHeightPct {
				t.Errorf("heightPct: expected %v, got %v", tc.wantHeightPct, got)
			}
			if got := float64(d.heightRel); got != tc.wantHeightRel {
				t.Errorf("heightRel: expected %v, got %v", tc.wantHeightRel, got)
			}
			if d.widthSet != tc.wantWidthSet {
				t.Errorf("widthSet: expected %v, got %v", tc.wantWidthSet, d.widthSet)
			}
			if d.heightSet != tc.wantHeightSet {
				t.Errorf("heightSet: expected %v, got %v", tc.wantHeightSet, d.heightSet)
			}
		})
	}
}
