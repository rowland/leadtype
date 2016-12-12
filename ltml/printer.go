// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Printer interface {
	Print(w Writer) error
	Top() float64
	Right() float64
	Bottom() float64
	Left() float64
	PreferredWidth() float64
	SetLeft(float64)
	SetWidth(float64)
}
