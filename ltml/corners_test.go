// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"testing"
)

func TestCorners_SetAll_default(t *testing.T) {
	var corners Corners
	if len(corners) != 0 {
		t.Errorf("Expected 0, got %d", len(corners))
	}
}

func TestCorners_SetAll_1(t *testing.T) {
	var corners Corners
	var expected = []float64{3.0}
	corners.SetAll("3", "")
	if len(corners) != len(expected) {
		t.Errorf("Expected %d, got %d", len(expected), len(corners))
	}
	for i := 0; i < len(corners); i++ {
		if corners[i] != expected[i] {
			t.Errorf("Expected %f, got %f", expected[i], corners[i])
		}
	}
}

func TestCorners_SetAll_2(t *testing.T) {
	var corners Corners
	var expected = []float64{4.0, 5.0}
	corners.SetAll("4 5", "")
	if len(corners) != len(expected) {
		t.Errorf("Expected %d, got %d", len(expected), len(corners))
	}
	for i := 0; i < len(corners); i++ {
		if corners[i] != expected[i] {
			t.Errorf("Expected %f, got %f", expected[i], corners[i])
		}
	}
}

func TestCorners_SetAll_4(t *testing.T) {
	var corners Corners
	var expected = []float64{4.0, 5.0, 6.0, 7.0}
	corners.SetAll("4 5 6 7", "")
	if len(corners) != len(expected) {
		t.Errorf("Expected %d, got %d", len(expected), len(corners))
	}
	for i := 0; i < len(corners); i++ {
		if corners[i] != expected[i] {
			t.Errorf("Expected %f, got %f", expected[i], corners[i])
		}
	}
}

func TestCorners_SetAll_8(t *testing.T) {
	var corners Corners
	var expected = []float64{6.0, 7.0, 8.0, 9.0, 10.0, 11.0, 12.0, 13.0}
	corners.SetAll("6 7 8 9 10 11 12 13", "")
	if len(corners) != len(expected) {
		t.Errorf("Expected %d, got %d", len(expected), len(corners))
	}
	for i := 0; i < len(corners); i++ {
		if corners[i] != expected[i] {
			t.Errorf("Expected %f, got %f", expected[i], corners[i])
		}
	}
}
