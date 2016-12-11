// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Writer interface {
	MoveTo(x, y float64)
	NewPage()
	Print(text string) (err error)
	SetFont(name string, size float64) error
}