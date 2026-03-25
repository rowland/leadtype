// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"reflect"
	"testing"
)

func TestBoolGrid(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		bg := NewBoolGrid(3, 2)
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
		bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
		tests := []struct {
			name     string
			col      int
			expected []bool
		}{
			{name: "First", col: 0, expected: []bool{true, false}},
			{name: "Second", col: 1, expected: []bool{false, true}},
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
		bg := NewBoolGrid(3, 2)
		if bg.Cols() != 3 {
			t.Errorf("Expected %d, got %d", 3, bg.Cols())
		}
	})

	t.Run("SetCols", func(t *testing.T) {
		tests := []struct {
			name     string
			size     int
			expected []bool
		}{
			{
				name: "Larger",
				size: 4,
				expected: []bool{
					true, false, true, false,
					false, true, false, false,
				},
			},
			{
				name: "Smaller",
				size: 2,
				expected: []bool{
					true, false,
					false, true,
				},
			},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
				bg.SetCols(tc.size)
				if !reflect.DeepEqual(tc.expected, bg.cells) {
					t.Errorf("Expected %#v, got %#v", tc.expected, bg.cells)
				}
			})
		}
	})

	t.Run("Row", func(t *testing.T) {
		bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
		tests := []struct {
			name     string
			row      int
			expected []bool
		}{
			{name: "First", row: 0, expected: []bool{true, false, true}},
			{name: "Second", row: 1, expected: []bool{false, true, false}},
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
		bg := NewBoolGrid(3, 2)
		if bg.Rows() != 2 {
			t.Errorf("Expected %d, got %d", 2, bg.Rows())
		}
	})

	t.Run("SetRows", func(t *testing.T) {
		tests := []struct {
			name     string
			size     int
			expected []bool
		}{
			{
				name: "Larger",
				size: 3,
				expected: []bool{
					true, false, true,
					false, true, false,
					false, false, false,
				},
			},
			{
				name: "Smaller",
				size: 1,
				expected: []bool{
					true, false, true,
				},
			},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
				bg.SetRows(tc.size)
				if !reflect.DeepEqual(tc.expected, bg.cells) {
					t.Errorf("Expected %#v, got %#v", tc.expected, bg.cells)
				}
			})
		}
	})

	t.Run("Cell", func(t *testing.T) {
		bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
		expected := false
		for row := 0; row < 2; row++ {
			for col := 0; col < 3; col++ {
				expected = !expected
				if actual := bg.Cell(col, row); actual != expected {
					t.Errorf("Expected %t, got %t", expected, actual)
				}
			}
		}
	})

	t.Run("SetCell", func(t *testing.T) {
		bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
		expected := false
		for row := 0; row < 2; row++ {
			for col := 0; col < 3; col++ {
				expected = !expected
				bg.SetCell(col, row, expected)
				if actual := bg.Cell(col, row); actual != expected {
					t.Errorf("Expected %t, got %t", expected, actual)
				}
			}
		}
	})

	t.Run("SetCellLarger", func(t *testing.T) {
		bg := NewBoolGrid(1, 1)
		bg.SetCell(1, 1, true)
		if bg.Cols() != 2 {
			t.Errorf("Expected %d, got %d", 2, bg.Cols())
		}
		if bg.Rows() != 2 {
			t.Errorf("Expected %d, got %d", 2, bg.Rows())
		}
		if actual := bg.Cell(1, 1); actual != true {
			t.Errorf("Expected %t, got %t", true, bg.Cell(1, 1))
		}
	})
}
