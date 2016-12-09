// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type AParent struct {
	parent interface{}
}

func (p *AParent) Parent() interface{} {
	return p.parent
}

func (p *AParent) SetParent(value interface{}) error {
	p.parent = value
	return nil
}
