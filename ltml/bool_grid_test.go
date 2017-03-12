// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"reflect"
	"testing"
)

func TestNewBoolGrid(t *testing.T) {
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
}

func TestBoolGrid_Col(t *testing.T) {
	bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
	expected1 := []bool{true, false}
	if actual1 := bg.Col(0); !reflect.DeepEqual(expected1, actual1) {
		t.Errorf("Expected %#v, got %#v", expected1, actual1)
	}
	expected2 := []bool{false, true}
	if actual2 := bg.Col(1); !reflect.DeepEqual(expected2, actual2) {
		t.Errorf("Expected %#v, got %#v", expected2, actual2)
	}
}

func TestBoolGrid_Cols(t *testing.T) {
	bg := NewBoolGrid(3, 2)
	if bg.Cols() != 3 {
		t.Errorf("Expected %d, got %d", 3, bg.Cols())
	}
}

func TestBoolGrid_SetCols_larger(t *testing.T) {
	bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
	expected := []bool{
		true, false, true, false,
		false, true, false, false,
	}
	bg.SetCols(4)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func TestBoolGrid_SetCols_smaller(t *testing.T) {
	bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
	expected := []bool{
		true, false,
		false, true,
	}
	bg.SetCols(2)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func TestBoolGrid_Row(t *testing.T) {
	bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
	expected1 := []bool{true, false, true}
	if actual1 := bg.Row(0); !reflect.DeepEqual(expected1, actual1) {
		t.Errorf("Expected %#v, got %#v", expected1, actual1)
	}
	expected2 := []bool{false, true, false}
	if actual2 := bg.Row(1); !reflect.DeepEqual(expected2, actual2) {
		t.Errorf("Expected %#v, got %#v", expected2, actual2)
	}
}

func TestBoolGrid_Rows(t *testing.T) {
	bg := NewBoolGrid(3, 2)
	if bg.Rows() != 2 {
		t.Errorf("Expected %d, got %d", 2, bg.Rows())
	}
}

func TestBoolGrid_SetRows_larger(t *testing.T) {
	bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
	expected := []bool{
		true, false, true,
		false, true, false,
		false, false, false,
	}
	bg.SetRows(3)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func TestBoolGrid_SetRows_smaller(t *testing.T) {
	bg := BoolGrid{cols: 3, rows: 2, cells: []bool{true, false, true, false, true, false}}
	expected := []bool{
		true, false, true,
	}
	bg.SetRows(1)
	if !reflect.DeepEqual(expected, bg.cells) {
		t.Errorf("Expected %#v, got %#v", expected, bg.cells)
	}
}

func TestBoolGrid_Cell(t *testing.T) {
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
}

func TestBoolGrid_SetCell(t *testing.T) {
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
}

func TestBoolGrid_SetCell_larger(t *testing.T) {
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
}
