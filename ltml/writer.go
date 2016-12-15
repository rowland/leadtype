// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Writer interface {
	LineTo(x, y float64)
	MoveTo(x, y float64)
	NewPage()
	Print(text string) (err error)
	Rectangle(x, y, width, height float64, border bool, fill bool)
	Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool)
	SetFont(name string, size float64) error
	SetLineColor(color string)
	SetLineDashPattern(pattern string)
	SetLineWidth(width float64)
}
