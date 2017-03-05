// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"

	"github.com/rowland/leadtype/colors"
)

type PenStyle struct {
	id      string
	color   colors.Color
	width   float64
	pattern string
}

func (ps *PenStyle) Apply(w Writer) {
	// fmt.Printf("Applying %s\n", ps)
	w.SetLineColor(colors.Color(ps.color))
	w.SetLineWidth(ps.width)
	w.SetLineDashPattern(ps.pattern)
}

func (ps *PenStyle) ID() string {
	return ps.id
}

func (ps *PenStyle) SetAttrs(prefix string, attrs map[string]string) {
	if id, ok := attrs[prefix+"id"]; ok {
		ps.id = id
	}
	if color, ok := attrs[prefix+"color"]; ok {
		ps.color = NamedColor(color)
	}
	if width, ok := attrs[prefix+"width"]; ok {
		ps.width = ParseMeasurement(width, "pt")
	}
	if pattern, ok := attrs[prefix+"pattern"]; ok {
		ps.pattern = pattern
	}
}

func (ps *PenStyle) String() string {
	return fmt.Sprintf("PenStyle id=%s color=%v width=%f pattern=%s", ps.id, ps.color, ps.width, ps.pattern)
}

const defaultPenPattern = "solid"

func PenStyleFor(id string, scope HasScope) *PenStyle {
	style, ok := scope.StyleFor(id)
	if !ok {
		style, ok = scope.StyleFor("pen_" + id)
	}
	if ok {
		ps, _ := style.(*PenStyle)
		return ps
	}
	ps := &PenStyle{id: "pen_" + id, color: NamedColor(id), pattern: defaultPenPattern}
	scope.AddStyle(ps)
	return ps
}

var _ HasAttrsPrefix = (*PenStyle)(nil)
var _ Styler = (*PenStyle)(nil)

func init() {
	registerTag(DefaultSpace, "pen", func() interface{} { return &PenStyle{} })
}
