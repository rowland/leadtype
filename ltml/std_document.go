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

func (d *StdDocument) Page(i int) *StdPage {
	if i < len(d.children) {
		return d.children[i].(*StdPage)
	}
	return nil
}

func (d *StdDocument) PageStyle() *PageStyle {
	if d.pageStyle == nil {
		style := PageStyleFor("letter", d.scope)
		if style == nil {
			panic("default page style missing")
		}
		return style
	}
	return d.pageStyle
}

func (d *StdDocument) Print(w Writer) error {
	return d.DrawContent(w)
}

func (d *StdDocument) String() string {
	return fmt.Sprintf("StdDocument %s units=%s margin=%s", &d.Identity, d.units, &d.margin)
}

func init() {
	registerTag(DefaultSpace, "ltml", func() interface{} { return &StdDocument{} })
}

var _ HasAttrs = (*StdDocument)(nil)
var _ Identifier = (*StdDocument)(nil)
