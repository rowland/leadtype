// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdContainer struct {
	Widget
	Children
}

func (c *StdContainer) Print(w Writer) error {
	fmt.Printf("Printing %s\n", c)
	for _, c := range c.children {
		if c, ok := c.(Printer); ok {
			if err := c.Print(w); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *StdContainer) SetAttrs(attrs map[string]string) {
	c.Widget.SetAttrs(attrs)
}

func (c *StdContainer) String() string {
	return fmt.Sprintf("StdContainer %s", &c.Widget)
}

func init() {
	registerTag(DefaultSpace, "div", func() interface{} { return &StdContainer{} })
}

var _ Container = (*StdContainer)(nil)
var _ HasAttrs = (*StdContainer)(nil)
var _ HasParent = (*StdContainer)(nil)
var _ Identifier = (*StdContainer)(nil)
var _ Printer = (*StdContainer)(nil)
