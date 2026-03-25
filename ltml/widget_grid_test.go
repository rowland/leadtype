// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"reflect"
	"testing"
)

func altWidgetValue(value, value1, value2 Widget) Widget {
	if value == value1 {
		return value2
	} else {
		return value1
	}
}

func TestWidgetGrid(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		bg := NewWidgetGrid(3, 2)
		if bg.cols != 3 {
			t.Errorf("Expected %d, got %d", 3, bg.cols)
		}
		if bg.rows != 2 {
			t.Errorf("Expected %d, got %d", 2, bg.rows)
		}
		if len(bg.cells) != (3 * 2) {
			t.Errorf("Expected %d, got %d", 3*2, len(bg.cells))
		}
	})

	t.Run("Col", func(t *testing.T) {
		w := new(StdWidget)
		bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
		tests := []struct {
			name     string
			col      int
			expected []Widget
		}{
			{name: "First", col: 0, expected: []Widget{w, nil}},
			{name: "Second", col: 1, expected: []Widget{nil, w}},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				if actual := bg.Col(tc.col); !reflect.DeepEqual(tc.expected, actual) {
					t.Errorf("Expected %#v, got %#v", tc.expected, actual)
				}
			})
		}
	})

	t.Run("Cols", func(t *testing.T) {
		bg := NewWidgetGrid(3, 2)
		if bg.Cols() != 3 {
			t.Errorf("Expected %d, got %d", 3, bg.Cols())
		}
	})

	t.Run("SetCols", func(t *testing.T) {
		w := new(StdWidget)
		tests := []struct {
			name     string
			size     int
			expected []Widget
		}{
			{
				name: "Larger",
				size: 4,
				expected: []Widget{
					w, nil, w, nil,
					nil, w, nil, nil,
				},
			},
			{
				name: "Smaller",
				size: 2,
				expected: []Widget{
					w, nil,
					nil, w,
				},
			},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
				bg.SetCols(tc.size)
				if !reflect.DeepEqual(tc.expected, bg.cells) {
					t.Errorf("Expected %#v, got %#v", tc.expected, bg.cells)
				}
			})
		}
	})

	t.Run("Row", func(t *testing.T) {
		w := new(StdWidget)
		bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
		tests := []struct {
			name     string
			row      int
			expected []Widget
		}{
			{name: "First", row: 0, expected: []Widget{w, nil, w}},
			{name: "Second", row: 1, expected: []Widget{nil, w, nil}},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				if actual := bg.Row(tc.row); !reflect.DeepEqual(tc.expected, actual) {
					t.Errorf("Expected %#v, got %#v", tc.expected, actual)
				}
			})
		}
	})

	t.Run("Rows", func(t *testing.T) {
		bg := NewWidgetGrid(3, 2)
		if bg.Rows() != 2 {
			t.Errorf("Expected %d, got %d", 2, bg.Rows())
		}
	})

	t.Run("SetRows", func(t *testing.T) {
		w := new(StdWidget)
		tests := []struct {
			name     string
			size     int
			expected []Widget
		}{
			{
				name: "Larger",
				size: 3,
				expected: []Widget{
					w, nil, w,
					nil, w, nil,
					nil, nil, nil,
				},
			},
			{
				name: "Smaller",
				size: 1,
				expected: []Widget{
					w, nil, w,
				},
			},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
				bg.SetRows(tc.size)
				if !reflect.DeepEqual(tc.expected, bg.cells) {
					t.Errorf("Expected %#v, got %#v", tc.expected, bg.cells)
				}
			})
		}
	})

	t.Run("Cell", func(t *testing.T) {
		w := new(StdWidget)
		bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
		var expected Widget
		for row := 0; row < 2; row++ {
			for col := 0; col < 3; col++ {
				expected = altWidgetValue(expected, nil, w)
				if actual := bg.Cell(col, row); actual != expected {
					t.Errorf("Expected %v, got %v", expected, actual)
				}
			}
		}
	})

	t.Run("SetCell", func(t *testing.T) {
		w := new(StdWidget)
		bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
		var expected Widget
		for row := 0; row < 2; row++ {
			for col := 0; col < 3; col++ {
				expected = altWidgetValue(expected, nil, w)
				bg.SetCell(col, row, expected)
				if actual := bg.Cell(col, row); actual != expected {
					t.Errorf("Expected %v, got %v", expected, actual)
				}
			}
		}
	})

	t.Run("SetCellLarger", func(t *testing.T) {
		bg := NewWidgetGrid(1, 1)
		w := new(StdWidget)
		bg.SetCell(1, 1, w)
		if bg.Cols() != 2 {
			t.Errorf("Expected %d, got %d", 2, bg.Cols())
		}
		if bg.Rows() != 2 {
			t.Errorf("Expected %d, got %d", 2, bg.Rows())
		}
		if actual := bg.Cell(1, 1); actual != w {
			t.Errorf("Expected %v, got %v", w, bg.Cell(1, 1))
		}
	})
}
