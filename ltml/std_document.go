// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdDocument struct {
	StdPage
}

func (d *StdDocument) Font() *FontStyle {
	if d.font == nil {
		return defaultFont
	}
	return d.font
}

func (d *StdDocument) Print(w Writer) error {
	fmt.Printf("Printing %s\n", d)
	fmt.Print(&d.Scope)
	for _, c := range d.children {
		if err := c.Print(w); err != nil {
			return err
		}
	}
	return nil
}

func (d *StdDocument) SetAttrs(attrs map[string]string) {
	d.units.SetAttrs(attrs)
	if margin, ok := attrs["margin"]; ok {
		d.margin.SetAll(margin, d.units)
	}
	d.margin.SetAttrs("margin-", attrs, d.units)
}

func (d *StdDocument) String() string {
	return fmt.Sprintf("StdDocument %s units=%s margin=%s", &d.Identity, d.units, &d.margin)
}

func (d *StdDocument) Units() Units {
	return d.units
}

//----- to factor out?
func (d *StdDocument) PreferredWidth() float64 {
	return 0
}

func (d *StdDocument) ContentHeight() float64 {
	return 0
}

func (d *StdDocument) ContentWidth() float64 {
	return 0
}

//-----

func init() {
	registerTag(DefaultSpace, "ltml", func() interface{} { return &StdDocument{} })
}

var _ HasAttrs = (*StdDocument)(nil)
var _ Identifier = (*StdDocument)(nil)
