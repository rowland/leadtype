// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type BoolGrid struct {
	cols, rows int
	cells      []bool
}

func NewBoolGrid(cols, rows int) *BoolGrid {
	return &BoolGrid{cols: cols, rows: rows, cells: make([]bool, cols*rows)}
}

func (bg *BoolGrid) Cell(col, row int) bool {
	if col < bg.cols && row < bg.rows {
		return bg.cells[row*bg.cols+col]
	}
	return false
}

func (bg *BoolGrid) SetCell(col, row int, value bool) {
	if col >= bg.cols {
		bg.SetCols(col + 1)
	}
	if row >= bg.rows {
		bg.SetRows(row + 1)
	}
	bg.cells[row*bg.cols+col] = value
}

func (bg *BoolGrid) Cols() int {
	return bg.cols
}

func (bg *BoolGrid) Col(col int) []bool {
	values := make([]bool, bg.rows)
	if col < bg.cols {
		for row := 0; row < bg.rows; row++ {
			values[row] = bg.Cell(col, row)
		}
	}
	return values
}

func (bg *BoolGrid) SetCols(cols int) {
	newCells := make([]bool, cols*bg.rows)
	for row := 0; row < bg.rows; row++ {
		for col := 0; col < cols && col < bg.cols; col++ {
			newCells[row*cols+col] = bg.cells[row*bg.cols+col]
		}
	}
	bg.cells = newCells
	bg.cols = cols
}

func (bg *BoolGrid) Row(row int) []bool {
	values := make([]bool, bg.cols)
	if row < bg.rows {
		copy(values, bg.cells[row*bg.cols:])
	}
	return values
}

func (bg *BoolGrid) Rows() int {
	return bg.rows
}

func (bg *BoolGrid) SetRows(rows int) {
	newCells := make([]bool, bg.cols*rows)
	for row := 0; row < rows && row < bg.rows; row++ {
		for col := 0; col < bg.cols; col++ {
			newCells[row*bg.cols+col] = bg.cells[row*bg.cols+col]
		}
	}
	bg.cells = newCells
	bg.rows = rows
}
