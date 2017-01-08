// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"testing"
)

func TestDimensions_SetAttrs_width(t *testing.T) {
	var d Dimensions
	d.SetAttrs(map[string]string{"width": "30"}, "")
	if d.width != 30 {
		t.Errorf("Expected 30, got %v", d.width)
	}
	if !d.widthSet {
		t.Errorf("Expected true, got %v", d.widthSet)
	}
}

func TestDimensions_SetAttrs_widthPct(t *testing.T) {
	var d Dimensions
	d.SetAttrs(map[string]string{"width": "40%"}, "")
	if d.widthPct != 40 {
		t.Errorf("Expected 40, got %v", d.widthPct)
	}
	if !d.widthSet {
		t.Errorf("Expected true, got %v", d.widthSet)
	}
}

func TestDimensions_SetAttrs_widthRelPlus(t *testing.T) {
	var d Dimensions
	d.SetAttrs(map[string]string{"width": "+50"}, "")
	if d.widthRel != 50 {
		t.Errorf("Expected +50, got %v", d.widthRel)
	}
	if !d.widthSet {
		t.Errorf("Expected true, got %v", d.widthSet)
	}
}

func TestDimensions_SetAttrs_widthRelMinus(t *testing.T) {
	var d Dimensions
	d.SetAttrs(map[string]string{"width": "-60"}, "")
	if d.widthRel != -60 {
		t.Errorf("Expected -60, got %v", d.widthRel)
	}
	if !d.widthSet {
		t.Errorf("Expected true, got %v", d.widthSet)
	}
}

func TestDimensions_SetAttrs_height(t *testing.T) {
	var d Dimensions
	d.SetAttrs(map[string]string{"height": "30"}, "")
	if d.height != 30 {
		t.Errorf("Expected 30, got %v", d.height)
	}
	if !d.heightSet {
		t.Errorf("Expected true, got %v", d.heightSet)
	}
}

func TestDimensions_SetAttrs_heightPct(t *testing.T) {
	var d Dimensions
	d.SetAttrs(map[string]string{"height": "40%"}, "")
	if d.heightPct != 40 {
		t.Errorf("Expected 40, got %v", d.heightPct)
	}
	if !d.heightSet {
		t.Errorf("Expected true, got %v", d.heightSet)
	}
}

func TestDimensions_SetAttrs_heightRelPlus(t *testing.T) {
	var d Dimensions
	d.SetAttrs(map[string]string{"height": "+50"}, "")
	if d.heightRel != 50 {
		t.Errorf("Expected +50, got %v", d.heightRel)
	}
	if !d.heightSet {
		t.Errorf("Expected true, got %v", d.heightSet)
	}
}

func TestDimensions_SetAttrs_heightRelMinus(t *testing.T) {
	var d Dimensions
	d.SetAttrs(map[string]string{"height": "-60"}, "")
	if d.heightRel != -60 {
		t.Errorf("Expected -60, got %v", d.heightRel)
	}
	if !d.heightSet {
		t.Errorf("Expected true, got %v", d.heightSet)
	}
}
