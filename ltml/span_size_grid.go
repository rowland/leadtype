// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type SpanSize struct {
	Span int
	Size float64
}

type SpanSizeGrid struct {
	cols, rows int
	cells      []SpanSize
}

func NewSpanSizeGrid(cols, rows int) *SpanSizeGrid {
	return &SpanSizeGrid{cols: cols, rows: rows, cells: make([]SpanSize, cols*rows)}
}

func (bg *SpanSizeGrid) Cell(col, row int) SpanSize {
	if col < bg.cols && row < bg.rows {
		return bg.cells[row*bg.cols+col]
	}
	return SpanSize{}
}

func (bg *SpanSizeGrid) SetCell(col, row int, value SpanSize) {
	if col >= bg.cols {
		bg.SetCols(col + 1)
	}
	if row >= bg.rows {
		bg.SetRows(row + 1)
	}
	bg.cells[row*bg.cols+col] = value
}

func (bg *SpanSizeGrid) Cols() int {
	return bg.cols
}

func (bg *SpanSizeGrid) Col(col int) []SpanSize {
	values := make([]SpanSize, bg.rows)
	if col < bg.cols {
		for row := 0; row < bg.rows; row++ {
			values[row] = bg.Cell(col, row)
		}
	}
	return values
}

func (bg *SpanSizeGrid) SetCols(cols int) {
	newCells := make([]SpanSize, cols*bg.rows)
	for row := 0; row < bg.rows; row++ {
		for col := 0; col < cols && col < bg.cols; col++ {
			newCells[row*cols+col] = bg.cells[row*bg.cols+col]
		}
	}
	bg.cells = newCells
	bg.cols = cols
}

func (bg *SpanSizeGrid) Row(row int) []SpanSize {
	values := make([]SpanSize, bg.cols)
	if row < bg.rows {
		copy(values, bg.cells[row*bg.cols:])
	}
	return values
}

func (bg *SpanSizeGrid) Rows() int {
	return bg.rows
}

func (bg *SpanSizeGrid) SetRows(rows int) {
	newCells := make([]SpanSize, bg.cols*rows)
	for row := 0; row < rows && row < bg.rows; row++ {
		for col := 0; col < bg.cols; col++ {
			newCells[row*bg.cols+col] = bg.cells[row*bg.cols+col]
		}
	}
	bg.cells = newCells
	bg.rows = rows
}
