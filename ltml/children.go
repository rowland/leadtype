// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Children struct {
	children []Printer
}

func (c *Children) AddChild(child Printer) {
	c.children = append(c.children, child)
}

func (c *Children) Query(f func(child Printer) bool) []Printer {
	var results []Printer
	for _, child := range c.children {
		if f(child) {
			results = append(results, child)
		}
	}
	return results
}

func (c *Children) Widgets() []Printer {
	return c.children
}
