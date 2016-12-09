// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type PenStyle struct {
	id      string
	color   Color
	width   float64
	pattern string
}

func (ps *PenStyle) Apply(w Writer) {
	fmt.Printf("Applying %s\n", ps)
}

func (ps *PenStyle) ID() string {
	return ps.id
}

func (ps *PenStyle) SetAttrs(attrs map[string]string) {
	if id, ok := attrs["id"]; ok {
		ps.id = id
	}
	ps.color.SetAttrs(attrs)
	if width, ok := attrs["width"]; ok {
		ps.width = ParseMeasurement(width, "pt")
	}
	if pattern, ok := attrs["pattern"]; ok {
		ps.pattern = pattern
	}
}

func (ps *PenStyle) String() string {
	return fmt.Sprintf("PenStyle color=%s width=%f pattern=%s", ps.color, ps.width, ps.pattern)
}

func PenStyleFor(id string, scope HasScope) *PenStyle {
	style, ok := scope.Style(id)
	if !ok {
		style, ok = scope.Style("pen_" + id)
	}
	if ok {
		ps, _ := style.(*PenStyle)
		return ps
	}
	if validColor(id) {
		ps := &PenStyle{id: "pen_" + id, color: Color(id), pattern: defaultPenPattern}
		scope.AddStyle(ps)
		return ps
	}
	return nil
}

func init() {
	registerTag(DefaultSpace, "pen", func() interface{} { return &PenStyle{} })
}
