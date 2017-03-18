// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"reflect"
	"testing"
)

func TestNewWidgetGrid(t *testing.T) {
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
}

func TestWidgetGrid_Col(t *testing.T) {
	w := new(StdWidget)
	bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
	expected1 := []Widget{w, nil}
	if actual1 := bg.Col(0); !reflect.DeepEqual(expected1, actual1) {
		t.Errorf("Expected %#v, got %#v", expected1, actual1)
	}
	expected2 := []Widget{nil, w}
	if actual2 := bg.Col(1); !reflect.DeepEqual(expected2, actual2) {
		t.Errorf("Expected %#v, got %#v", expected2, actual2)
	}
}

func TestWidgetGrid_Cols(t *testing.T) {
	bg := NewWidgetGrid(3, 2)
	if bg.Cols() != 3 {
		t.Errorf("Expected %d, got %d", 3, bg.Cols())
	}
}

func TestWidgetGrid_SetCols_larger(t *testing.T) {
	w := new(StdWidget)
	bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
	expected := []Widget{
		w, nil, w, nil,
		nil, w, nil, nil,
	}
	bg.SetCols(4)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func TestWidgetGrid_SetCols_smaller(t *testing.T) {
	w := new(StdWidget)
	bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
	expected := []Widget{
		w, nil,
		nil, w,
	}
	bg.SetCols(2)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func TestWidgetGrid_Row(t *testing.T) {
	w := new(StdWidget)
	bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
	expected1 := []Widget{w, nil, w}
	if actual1 := bg.Row(0); !reflect.DeepEqual(expected1, actual1) {
		t.Errorf("Expected %#v, got %#v", expected1, actual1)
	}
	expected2 := []Widget{nil, w, nil}
	if actual2 := bg.Row(1); !reflect.DeepEqual(expected2, actual2) {
		t.Errorf("Expected %#v, got %#v", expected2, actual2)
	}
}

func TestWidgetGrid_Rows(t *testing.T) {
	bg := NewWidgetGrid(3, 2)
	if bg.Rows() != 2 {
		t.Errorf("Expected %d, got %d", 2, bg.Rows())
	}
}

func TestWidgetGrid_SetRows_larger(t *testing.T) {
	w := new(StdWidget)
	bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
	expected := []Widget{
		w, nil, w,
		nil, w, nil,
		nil, nil, nil,
	}
	bg.SetRows(3)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func TestWidgetGrid_SetRows_smaller(t *testing.T) {
	w := new(StdWidget)
	bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}
	expected := []Widget{
		w, nil, w,
	}
	bg.SetRows(1)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func altWidgetValue(value, value1, value2 Widget) Widget {
	if value == value1 {
		return value2
	} else {
		return value1
	}
}

func TestWidgetGrid_Cell(t *testing.T) {
	w := new(StdWidget)
	bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}

	var expected Widget = nil
	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			expected = altWidgetValue(expected, nil, w)
			if actual := bg.Cell(col, row); actual != expected {
				t.Errorf("Expected %t, got %t", expected, actual)
			}
		}
	}
}

func TestWidgetGrid_SetCell(t *testing.T) {
	w := new(StdWidget)
	bg := WidgetGrid{cols: 3, rows: 2, cells: []Widget{w, nil, w, nil, w, nil}}

	var expected Widget = nil
	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			expected = altWidgetValue(expected, nil, w)
			bg.SetCell(col, row, expected)
			if actual := bg.Cell(col, row); actual != expected {
				t.Errorf("Expected %t, got %t", expected, actual)
			}
		}
	}
}

func TestWidgetGrid_SetCell_larger(t *testing.T) {
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
		t.Errorf("Expected %t, got %t", w, bg.Cell(1, 1))
	}
}
