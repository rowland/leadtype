// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"

	"github.com/rowland/leadtype/colors"
)

type BrushStyle struct {
	id    string
	color colors.Color
}

func (bs *BrushStyle) Apply(w Writer) {
	fmt.Printf("Applying %s\n", bs)
	w.SetFillColor(bs.color)
}

func (bs *BrushStyle) Clone() *BrushStyle {
	clone := *bs
	return &clone
}

func (bs *BrushStyle) ID() string {
	return bs.id
}

func (bs *BrushStyle) SetAttrs(prefix string, attrs map[string]string) {
	if id, ok := attrs[prefix+"id"]; ok {
		bs.id = id
	}
	if color, ok := attrs[prefix+"color"]; ok {
		bs.color = NamedColor(color)
	}
	fmt.Println("color:", bs.color)
}

func (bs *BrushStyle) String() string {
	return fmt.Sprintf("BrushStyle id=%s color=%v", bs.id, bs.color)
}

func BrushStyleFor(id string, scope HasScope) *BrushStyle {
	style, ok := scope.StyleFor(id)
	if !ok {
		style, ok = scope.StyleFor("brush_" + id)
	}
	if ok {
		bs, _ := style.(*BrushStyle)
		return bs
	}
	bs := &BrushStyle{id: "brush_" + id, color: NamedColor(id)}
	scope.AddStyle(bs)
	return bs
}

var _ HasAttrsPrefix = (*BrushStyle)(nil)
var _ Styler = (*BrushStyle)(nil)

func init() {
	registerTag(DefaultSpace, "brush", func() interface{} { return &BrushStyle{} })
}
