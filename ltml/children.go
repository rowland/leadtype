// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Children struct {
	children []Widget
}

func (c *Children) AddChild(child Widget) {
	c.children = append(c.children, child)
}

func (c *Children) Query(f func(child Widget) bool) []Widget {
	var results []Widget
	for _, child := range c.children {
		if f(child) {
			results = append(results, child)
		}
	}
	return results
}

func (c *Children) Widgets() []Widget {
	return c.children
}
