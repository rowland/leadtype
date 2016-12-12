// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Container interface {
	AddChild(value Printer)
	ContentTop() float64
	ContentRight() float64
	ContentBottom() float64
	ContentLeft() float64
	ContentHeight() float64
	ContentWidth() float64
	Widgets() []Printer
	HasFont
	Printer
	Query(f func(value Printer) bool) []Printer
	Units() Units
}
