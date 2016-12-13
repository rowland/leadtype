// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Container interface {
	Widget
	HasFont

	AddChild(value Widget)
	LayoutStyle() *LayoutStyle
	Query(f func(value Widget) bool) []Widget
	Units() Units
	Widgets() []Widget
}

func LayoutContainer(c Container, w Writer) {
	c.LayoutStyle().Layout(c, w)
}
