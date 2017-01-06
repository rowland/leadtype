// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"testing"
)

func TestSides_SetAll_default(t *testing.T) {
	var sides Sides
	for i := 0; i < len(sides); i++ {
		if sides[i].Value != 0.0 {
			t.Errorf("Expected %f, got %f", 0.0, sides[i].Value)
		}
	}
}

func TestSides_SetAll_1(t *testing.T) {
	var sides Sides
	sides.SetAll("3", "")
	for i := 0; i < len(sides); i++ {
		if sides[i].Value != 3.0 {
			t.Errorf("Side %s: expected %f, got %f", sideNames[i], 3.0, sides[i].Value)
		}
	}
}

func TestSides_SetAll_2(t *testing.T) {
	var sides Sides
	var expected = []float64{4.0, 5.0, 4.0, 5.0}
	sides.SetAll("4 5", "")
	for i := 0; i < len(sides); i++ {
		if sides[i].Value != expected[i] {
			t.Errorf("Side %s: expected %f, got %f", sideNames[i], expected[i], sides[i].Value)
		}
	}
}

func TestSides_SetAll_3(t *testing.T) {
	var sides Sides
	var expected = []float64{4.0, 5.0, 6.0, 0.0}
	sides.SetAll("4 5 6", "")
	for i := 0; i < len(sides); i++ {
		if sides[i].Value != expected[i] {
			t.Errorf("Side %s: expected %f, got %f", sideNames[i], expected[i], sides[i].Value)
		}
	}
}

func TestSides_SetAll_4(t *testing.T) {
	var sides Sides
	var expected = []float64{6.0, 7.0, 8.0, 9.0}
	sides.SetAll("6 7 8 9", "")
	for i := 0; i < len(sides); i++ {
		if sides[i].Value != expected[i] {
			t.Errorf("Side %s: expected %f, got %f", sideNames[i], expected[i], sides[i].Value)
		}
	}
}

func TestSides_SetAttrs(t *testing.T) {
	var sides Sides
	var expected = []float64{6.0, 7.0, 8.0, 9.0}
	var attrs = map[string]string{"top": "6", "right": "7", "bottom": "8", "left": "9"}
	sides.SetAttrs("", attrs, "")
	for i := 0; i < len(sides); i++ {
		if sides[i].Value != expected[i] {
			t.Errorf("Side %s: expected %f, got %f", sideNames[i], expected[i], sides[i].Value)
		}
	}
}

func TestSides_SetAttrs_prefix(t *testing.T) {
	var sides Sides
	var expected = []float64{6.5, 7.5, 8.0, 9.0}
	var attrs = map[string]string{"margin-top": "6.5", "margin-right": "7.5", "margin-bottom": "8", "margin-left": "9"}
	sides.SetAttrs("margin-", attrs, "")
	for i := 0; i < len(sides); i++ {
		if sides[i].Value != expected[i] {
			t.Errorf("Side %s: expected %f, got %f", sideNames[i], expected[i], sides[i].Value)
		}
	}
}
