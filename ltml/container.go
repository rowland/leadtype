// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type TableOrder int

const (
	TableOrderRows = TableOrder(iota)
	TableOrderCols
)

type Container interface {
	Widget
	HasFont

	AddChild(value Widget)
	Cols() int
	Container() Container
	LayoutStyle() *LayoutStyle
	Order() TableOrder
	ParagraphStyle() *ParagraphStyle
	Query(f func(value Widget) bool) []Widget
	Rows() int
	Units() Units
	Widgets() []Widget
}

func LayoutContainer(c Container, w Writer) {
	c.LayoutStyle().Layout(c, w)
}

func MaxContentHeight(c Container) float64 {
	return MaxHeightAvail(c) - NonContentHeight(c)
}

func MaxHeightAvail(c Container) float64 {
	if h := c.Height(); h != 0 {
		return h
	}
	var top float64
	if c.TopIsSet() {
		top = c.Top()
	} else {
		top = ContentTop(c.Container())
	}
	return MaxContentHeight(c.Container()) - top - ContentTop(c.Container())
}
