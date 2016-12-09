// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdPage struct {
	Scope
	StdContainer
}

func (p *StdPage) Print(w Writer) error {
	fmt.Printf("Printing %s\n", p)
	fmt.Print(&p.Scope)
	return p.StdContainer.Print(w)
}

func (p *StdPage) SetAttrs(attrs map[string]string) {
	p.StdContainer.SetAttrs(attrs)
}

func (p *StdPage) SetParent(value interface{}) error {
	if d, ok := value.(*StdDocument); ok {
		p.margin = d.margin
		p.units = d.units
		return p.StdContainer.SetParent(value)
	} else {
		return fmt.Errorf("page must be child of ltml.")
	}
}

func (p *StdPage) String() string {
	return fmt.Sprintf("StdPage %s", &p.StdContainer)
}

func init() {
	registerTag(DefaultSpace, "page", func() interface{} { return &StdPage{} })
}

var _ Container = (*StdPage)(nil)
var _ HasAttrs = (*StdPage)(nil)
var _ HasParent = (*StdPage)(nil)
var _ Identifier = (*StdPage)(nil)
var _ Printer = (*StdPage)(nil)
