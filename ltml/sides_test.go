// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"testing"
)

func TestSides_SetAll(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected []float64
	}{
		{name: "Default", expected: []float64{0, 0, 0, 0}},
		{name: "OneValue", value: "3", expected: []float64{3, 3, 3, 3}},
		{name: "TwoValues", value: "4 5", expected: []float64{4, 5, 4, 5}},
		{name: "ThreeValues", value: "4 5 6", expected: []float64{4, 5, 6, 0}},
		{name: "FourValues", value: "6 7 8 9", expected: []float64{6, 7, 8, 9}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var sides Sides
			if tc.value != "" {
				sides.SetAll(tc.value, "")
			}
			assertSides(t, sides, tc.expected)
		})
	}
}

func TestSides_SetAttrs(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		attrs    map[string]string
		expected []float64
	}{
		{
			name:     "NoPrefix",
			attrs:    map[string]string{"top": "6", "right": "7", "bottom": "8", "left": "9"},
			expected: []float64{6, 7, 8, 9},
		},
		{
			name:     "WithPrefix",
			prefix:   "margin-",
			attrs:    map[string]string{"margin-top": "6.5", "margin-right": "7.5", "margin-bottom": "8", "margin-left": "9"},
			expected: []float64{6.5, 7.5, 8, 9},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var sides Sides
			sides.SetAttrs(tc.prefix, tc.attrs, "")
			assertSides(t, sides, tc.expected)
		})
	}
}

func assertSides(t *testing.T, sides Sides, expected []float64) {
	t.Helper()
	for i := 0; i < len(sides); i++ {
		if sides[i].Value != expected[i] {
			t.Errorf("Side %s: expected %f, got %f", sideNames[i], expected[i], sides[i].Value)
		}
	}
}
