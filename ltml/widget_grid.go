// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type WidgetGrid struct {
	cols, rows int
	cells      []Widget
}

func NewWidgetGrid(cols, rows int) *WidgetGrid {
	return &WidgetGrid{cols: cols, rows: rows, cells: make([]Widget, cols*rows)}
}

func (bg *WidgetGrid) Cell(col, row int) Widget {
	if col < bg.cols && row < bg.rows {
		return bg.cells[row*bg.cols+col]
	}
	return nil
}

func (bg *WidgetGrid) SetCell(col, row int, value Widget) {
	if col >= bg.cols {
		bg.SetCols(col + 1)
	}
	if row >= bg.rows {
		bg.SetRows(row + 1)
	}
	bg.cells[row*bg.cols+col] = value
}

func (bg *WidgetGrid) Cols() int {
	return bg.cols
}

func (bg *WidgetGrid) Col(col int) []Widget {
	values := make([]Widget, bg.rows)
	if col < bg.cols {
		for row := 0; row < bg.rows; row++ {
			values[row] = bg.Cell(col, row)
		}
	}
	return values
}

func (bg *WidgetGrid) SetCols(cols int) {
	newCells := make([]Widget, cols*bg.rows)
	for row := 0; row < bg.rows; row++ {
		for col := 0; col < cols && col < bg.cols; col++ {
			newCells[row*cols+col] = bg.cells[row*bg.cols+col]
		}
	}
	bg.cells = newCells
	bg.cols = cols
}

func (bg *WidgetGrid) Row(row int) []Widget {
	values := make([]Widget, bg.cols)
	if row < bg.rows {
		copy(values, bg.cells[row*bg.cols:])
	}
	return values
}

func (bg *WidgetGrid) Rows() int {
	return bg.rows
}

func (bg *WidgetGrid) SetRows(rows int) {
	newCells := make([]Widget, bg.cols*rows)
	for row := 0; row < rows && row < bg.rows; row++ {
		for col := 0; col < bg.cols; col++ {
			newCells[row*bg.cols+col] = bg.cells[row*bg.cols+col]
		}
	}
	bg.cells = newCells
	bg.rows = rows
}
