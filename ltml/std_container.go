// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdContainer struct {
	StdWidget
	Children
	layout          *LayoutStyle
	paragraphStyle  *ParagraphStyle
	preferredHeight float64
	preferredWidth  float64
}

func (c *StdContainer) Container() Container {
	return c.container
}

func (c *StdContainer) DrawContent(w Writer) error {
	// fmt.Printf("DrawContent %s\n", c)
	for _, child := range c.children {
		if err := Print(child, w); err != nil {
			return err
		}
	}
	return nil
}

func (c *StdContainer) LayoutStyle() *LayoutStyle {
	if c.layout == nil {
		return LayoutStyleFor("vbox", c.scope)
	}
	return c.layout
}

func (c *StdContainer) LayoutWidget(w Writer) {
	LayoutContainer(c, w)
}

func (c *StdContainer) ParagraphStyle() *ParagraphStyle {
	if c.paragraphStyle == nil {
		return c.container.ParagraphStyle()
	}
	return c.paragraphStyle
}

func (c *StdContainer) SetAttrs(attrs map[string]string) {
	c.StdWidget.SetAttrs(attrs)
	if layout, ok := attrs["layout"]; ok {
		c.layout = LayoutStyleFor(layout, c.scope)
	}
	if ps, ok := attrs["paragraph-style"]; ok {
		c.paragraphStyle = ParagraphStyleFor(ps, c.scope)
	}
	if MapHasKeyPrefix(attrs, "paragraph-style.") {
		c.paragraphStyle = c.ParagraphStyle().Clone()
		c.paragraphStyle.SetAttrs("paragraph-style.", attrs)
	}
}

func (c *StdContainer) String() string {
	return fmt.Sprintf("StdContainer layout=%v paragraphStyle=%v %s", c.layout, c.paragraphStyle, &c.StdWidget)
}

func init() {
	registerTag(DefaultSpace, "div", func() interface{} { return &StdContainer{} })
}

var _ Container = (*StdContainer)(nil)
var _ HasAttrs = (*StdContainer)(nil)
var _ Identifier = (*StdContainer)(nil)
var _ Printer = (*StdContainer)(nil)
var _ WantsContainer = (*StdContainer)(nil)
