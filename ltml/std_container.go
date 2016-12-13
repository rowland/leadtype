// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdContainer struct {
	StdWidget
	Children
	layout *LayoutStyle
}

func (c *StdContainer) LayoutStyle() *LayoutStyle {
	if c.layout == nil {
		return LayoutStyleFor("vbox", c.scope)
	}
	return c.layout
}

func (c *StdContainer) Print(w Writer) error {
	fmt.Printf("Printing %s\n", c)
	for _, c := range c.children {
		if err := c.Print(w); err != nil {
			return err
		}
	}
	return nil
}

func (c *StdContainer) SetAttrs(attrs map[string]string) {
	c.StdWidget.SetAttrs(attrs)
	if layout, ok := attrs["layout"]; ok {
		c.layout = LayoutStyleFor(layout, c.scope)
	}
}

func (c *StdContainer) String() string {
	return fmt.Sprintf("StdContainer %s", &c.StdWidget)
}

func init() {
	registerTag(DefaultSpace, "div", func() interface{} { return &StdContainer{} })
}

var _ Container = (*StdContainer)(nil)
var _ HasAttrs = (*StdContainer)(nil)
var _ Identifier = (*StdContainer)(nil)
var _ Printer = (*StdContainer)(nil)
var _ WantsContainer = (*StdContainer)(nil)
