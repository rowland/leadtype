// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"reflect"
	"testing"
)

func TestNewSpanSizeGrid(t *testing.T) {
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
}

var (
	ss0 = SpanSize{}
	ss1 = SpanSize{1, 1}
)

func TestSpanSizeGrid_Col(t *testing.T) {
	bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
	expected1 := []SpanSize{ss1, ss0}
	if actual1 := bg.Col(0); !reflect.DeepEqual(expected1, actual1) {
		t.Errorf("Expected %#v, got %#v", expected1, actual1)
	}
	expected2 := []SpanSize{ss0, ss1}
	if actual2 := bg.Col(1); !reflect.DeepEqual(expected2, actual2) {
		t.Errorf("Expected %#v, got %#v", expected2, actual2)
	}
}

func TestSpanSizeGrid_Cols(t *testing.T) {
	bg := NewSpanSizeGrid(3, 2)
	if bg.Cols() != 3 {
		t.Errorf("Expected %d, got %d", 3, bg.Cols())
	}
}

func TestSpanSizeGrid_SetCols_larger(t *testing.T) {
	bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
	expected := []SpanSize{
		ss1, ss0, ss1, ss0,
		ss0, ss1, ss0, ss0,
	}
	bg.SetCols(4)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func TestSpanSizeGrid_SetCols_smaller(t *testing.T) {
	bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
	expected := []SpanSize{
		ss1, ss0,
		ss0, ss1,
	}
	bg.SetCols(2)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func TestSpanSizeGrid_Row(t *testing.T) {
	bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
	expected1 := []SpanSize{ss1, ss0, ss1}
	if actual1 := bg.Row(0); !reflect.DeepEqual(expected1, actual1) {
		t.Errorf("Expected %#v, got %#v", expected1, actual1)
	}
	expected2 := []SpanSize{ss0, ss1, ss0}
	if actual2 := bg.Row(1); !reflect.DeepEqual(expected2, actual2) {
		t.Errorf("Expected %#v, got %#v", expected2, actual2)
	}
}

func TestSpanSizeGrid_Rows(t *testing.T) {
	bg := NewSpanSizeGrid(3, 2)
	if bg.Rows() != 2 {
		t.Errorf("Expected %d, got %d", 2, bg.Rows())
	}
}

func TestSpanSizeGrid_SetRows_larger(t *testing.T) {
	bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
	expected := []SpanSize{
		ss1, ss0, ss1,
		ss0, ss1, ss0,
		ss0, ss0, ss0,
	}
	bg.SetRows(3)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func TestSpanSizeGrid_SetRows_smaller(t *testing.T) {
	bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}
	expected := []SpanSize{
		ss1, ss0, ss1,
	}
	bg.SetRows(1)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func altSpanSizeValue(value, value1, value2 SpanSize) SpanSize {
	if value == value1 {
		return value2
	} else {
		return value1
	}
}

func TestSpanSizeGrid_Cell(t *testing.T) {
	bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}

	expected := ss0
	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			expected = altSpanSizeValue(expected, ss0, ss1)
			if actual := bg.Cell(col, row); actual != expected {
				t.Errorf("Expected %t, got %t", expected, actual)
			}
		}
	}
}

func TestSpanSizeGrid_SetCell(t *testing.T) {
	bg := SpanSizeGrid{cols: 3, rows: 2, cells: []SpanSize{ss1, ss0, ss1, ss0, ss1, ss0}}

	expected := ss0
	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			expected = altSpanSizeValue(expected, ss0, ss1)
			bg.SetCell(col, row, expected)
			if actual := bg.Cell(col, row); actual != expected {
				t.Errorf("Expected %t, got %t", expected, actual)
			}
		}
	}
}

func TestSpanSizeGrid_SetCell_larger(t *testing.T) {
	bg := NewSpanSizeGrid(1, 1)

	bg.SetCell(1, 1, ss1)
	if bg.Cols() != 2 {
		t.Errorf("Expected %d, got %d", 2, bg.Cols())
	}
	if bg.Rows() != 2 {
		t.Errorf("Expected %d, got %d", 2, bg.Rows())
	}
	if actual := bg.Cell(1, 1); actual != ss1 {
		t.Errorf("Expected %t, got %t", ss1, bg.Cell(1, 1))
	}
}
