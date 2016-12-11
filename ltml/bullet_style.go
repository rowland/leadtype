// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type BulletStyle struct {
	scope HasScope
	id    string
	font  *FontStyle
	text  string
	width float64
	units Units
}

func (bs *BulletStyle) Apply(w Writer) {
	fmt.Printf("Applying %s\n", bs)
}

func (bs *BulletStyle) ID() string {
	return bs.id
}

func (bs *BulletStyle) SetAttrs(attrs map[string]string) {
	if id, ok := attrs["id"]; ok {
		bs.id = id
	}
	if units, ok := attrs["units"]; ok {
		bs.units = Units(units)
	}
	if width, ok := attrs["width"]; ok {
		bs.width = ParseMeasurement(width, bs.units)
	}
	if font, ok := attrs["font"]; ok {
		bs.font = FontStyleFor(font, bs.scope)
	}
}

func (bs *BulletStyle) SetScope(scope HasScope) {
	bs.scope = scope
}

func (bs *BulletStyle) String() string {
	return fmt.Sprintf("BulletStyle id=%s text=%s width=%f units=%s font=%s",
		bs.id, bs.text, bs.width, bs.units, bs.font)
}

func BulletStyleFor(id string, scope HasScope) *BulletStyle {
	if style, ok := scope.Style(id); ok {
		bs, _ := style.(*BulletStyle)
		return bs
	}
	return nil
}

var _ HasAttrs = (*BulletStyle)(nil)
var _ Styler = (*BulletStyle)(nil)
var _ WantsScope = (*BulletStyle)(nil)

func init() {
	registerTag(DefaultSpace, "bullet", func() interface{} { return &BulletStyle{} })
}
