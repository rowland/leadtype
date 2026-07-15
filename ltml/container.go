// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import "fmt"

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
	Dir() Dir
	LayoutStyle() *LayoutStyle
	Order() TableOrder
	ParagraphStyle() *ParagraphStyle
	Query(f func(value Widget) bool) []Widget
	Rows() int
	SplitEnabled() bool
	Units() Units
	Widgets() []Widget
}

func LayoutContainer(c Container, w Writer) error {
	if c == nil {
		return wrapLayoutError("", "", fmt.Errorf("container is nil"))
	}
	if preparer, ok := c.(interface{ prepareForLayout(Writer) }); ok {
		preparer.prepareForLayout(w)
	}
	style := c.LayoutStyle()
	if style == nil {
		return wrapLayoutError("", c.Path(), fmt.Errorf("layout style is nil"))
	}
	return style.Layout(c, w)
}

func containerPath(c Container) string {
	if c == nil {
		return ""
	}
	return c.Path()
}

func MaxContentHeight(c Container) float64 {
	return MaxHeightAvail(c) - NonContentHeight(c)
}

func MaxHeightAvail(c Container) float64 {
	if h := c.Height(); h != 0 {
		return h
	}
	var top float64
	containerTop := 0.0
	if c.TopIsSet() {
		top = c.Top()
	} else {
		top = ContentTop(c.Container())
	}
	if c.Container() != nil {
		containerTop = ContentTop(c.Container())
	}
	return MaxContentHeight(c.Container()) + containerTop - top
}

func rootPageForContainer(c Container) *StdPage {
	for c != nil {
		if page, ok := c.(*StdPage); ok {
			return page
		}
		c = c.Container()
	}
	return nil
}

// containerHasEffectiveContinuation reports whether container can continue its
// own non-fitting content through every ancestor to an overflow-enabled page.
// A table can continue its own rows, but only a vbox can propagate a child
// container's continuation recursively.
func containerHasEffectiveContinuation(container Container) bool {
	if !containerContinuationEnabled(container) {
		return false
	}
	if _, ok := container.(*StdPage); ok {
		return true
	}

	for parent := container.Container(); parent != nil; parent = parent.Container() {
		if page, ok := parent.(*StdPage); ok {
			return page.effectiveOverflow() && page.supportsOverflowRetry()
		}
		if !containerPropagatesContinuation(parent) {
			return false
		}
	}
	return false
}

func containerContinuationEnabled(container Container) bool {
	if page, ok := container.(*StdPage); ok {
		return page.effectiveOverflow() && page.supportsOverflowRetry()
	}
	return container.SplitEnabled()
}

func containerPropagatesContinuation(container Container) bool {
	if container.LayoutStyle() == nil || container.LayoutStyle().manager != "vbox" {
		return false
	}
	return container.SplitEnabled()
}
