// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type BulletStyle struct {
	AParent
	id    string
	font  *FontStyle
	text  string
	width float64
	units string
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
		bs.units = units
	}
	if width, ok := attrs["width"]; ok {
		bs.width = ParseMeasurement(width, bs.units)
	}
	if font, ok := attrs["font"]; ok {
		scope := ScopeFor(bs.parent)
		bs.font = FontStyleFor(font, scope)
	}
}

func (bs *BulletStyle) String() string {
	return fmt.Sprintf("BulletStyle id=%s text=%s width=%f units=%s font=%s",
		bs.id, bs.text, bs.width, bs.units, bs.font)
}

func BulletStyleFor(id string, scope HasScope) *BulletStyle {
	style, ok := scope.Style(id)
	if !ok {
		style, ok = scope.Style("bullet_" + id)
	}
	if ok {
		bs, _ := style.(*BulletStyle)
		return bs
	}
	return nil
}

func init() {
	registerTag(DefaultSpace, "bullet", func() interface{} { return &BulletStyle{} })
}
