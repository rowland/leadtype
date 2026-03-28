// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type AParent struct {
	parent any
}

func (p *AParent) Parent() any {
	return p.parent
}

func (p *AParent) SetParent(value any) error {
	p.parent = value
	return nil
}
