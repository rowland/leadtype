// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Container interface {
	AddChild(value Printer)
	HasFont
	Printer
	Query(f func(value Printer) bool) []Printer
	Units() Units
}