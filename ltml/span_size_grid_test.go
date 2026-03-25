// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"reflect"
	"testing"
)

var (
	ss0 = SpanSize{}
	ss1 = SpanSize{1, 1}
)

func altSpanSizeValue(value, value1, value2 SpanSize) SpanSize {
	if value == value1 {
		return value2
	} else {
		return value1
	}
}

func TestSpanSizeGrid(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		bg := NewSpanSizeGrid(3, 2)
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
		bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
		tests := []struct {
			name     string
			col      int
			expected []SpanSize
		}{
			{name: "First", col: 0, expected: []SpanSize{ss1, ss0}},
			{name: "Second", col: 1, expected: []SpanSize{ss0, ss1}},
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
		bg := NewSpanSizeGrid(3, 2)
		if bg.Cols() != 3 {
			t.Errorf("Expected %d, got %d", 3, bg.Cols())
		}
	})

	t.Run("SetCols", func(t *testing.T) {
		tests := []struct {
			name     string
			size     int
			expected []SpanSize
		}{
			{
				name: "Larger",
				size: 4,
				expected: []SpanSize{
					ss1, ss0, ss1, ss0,
					ss0, ss1, ss0, ss0,
				},
			},
			{
				name: "Smaller",
				size: 2,
				expected: []SpanSize{
					ss1, ss0,
					ss0, ss1,
				},
			},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
				bg.SetCols(tc.size)
				if !reflect.DeepEqual(tc.expected, bg.cells) {
					t.Errorf("Expected %#v, got %#v", tc.expected, bg.cells)
				}
			})
		}
	})

	t.Run("Row", func(t *testing.T) {
		bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
		tests := []struct {
			name     string
			row      int
			expected []SpanSize
		}{
			{name: "First", row: 0, expected: []SpanSize{ss1, ss0, ss1}},
			{name: "Second", row: 1, expected: []SpanSize{ss0, ss1, ss0}},
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
		bg := NewSpanSizeGrid(3, 2)
		if bg.Rows() != 2 {
			t.Errorf("Expected %d, got %d", 2, bg.Rows())
		}
	})

	t.Run("SetRows", func(t *testing.T) {
		tests := []struct {
			name     string
			size     int
			expected []SpanSize
		}{
			{
				name: "Larger",
				size: 3,
				expected: []SpanSize{
					ss1, ss0, ss1,
					ss0, ss1, ss0,
					ss0, ss0, ss0,
				},
			},
			{
				name: "Smaller",
				size: 1,
				expected: []SpanSize{
					ss1, ss0, ss1,
				},
			},
		}
		for _, tc := range tests {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
				bg.SetRows(tc.size)
				if !reflect.DeepEqual(tc.expected, bg.cells) {
					t.Errorf("Expected %#v, got %#v", tc.expected, bg.cells)
				}
			})
		}
	})

	t.Run("Cell", func(t *testing.T) {
		bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
		expected := ss0
		for row := 0; row < 2; row++ {
			for col := 0; col < 3; col++ {
				expected = altSpanSizeValue(expected, ss0, ss1)
				if actual := bg.Cell(col, row); actual != expected {
					t.Errorf("Expected %v, got %v", expected, actual)
				}
			}
		}
	})

	t.Run("SetCell", func(t *testing.T) {
		bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
		expected := ss0
		for row := 0; row < 2; row++ {
			for col := 0; col < 3; col++ {
				expected = altSpanSizeValue(expected, ss0, ss1)
				bg.SetCell(col, row, expected)
				if actual := bg.Cell(col, row); actual != expected {
					t.Errorf("Expected %v, got %v", expected, actual)
				}
			}
		}
	})

	t.Run("SetCellLarger", func(t *testing.T) {
		bg := NewSpanSizeGrid(1, 1)
		bg.SetCell(1, 1, ss1)
		if bg.Cols() != 2 {
			t.Errorf("Expected %d, got %d", 2, bg.Cols())
		}
		if bg.Rows() != 2 {
			t.Errorf("Expected %d, got %d", 2, bg.Rows())
		}
		if actual := bg.Cell(1, 1); actual != ss1 {
			t.Errorf("Expected %v, got %v", ss1, bg.Cell(1, 1))
		}
	})
}
