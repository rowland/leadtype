// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Children struct {
	children []interface{}
}

func (c *Children) AddChild(child interface{}) {
	c.children = append(c.children, child)
}

func (c *Children) Query(f func(child interface{}) bool) []interface{} {
	var results []interface{}
	for _, child := range c.children {
		if f(child) {
			results = append(results, child)
		}
	}
	return results
}
