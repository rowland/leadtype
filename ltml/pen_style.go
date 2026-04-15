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
	cap     string
}

func (ps *PenStyle) Apply(w Writer) {
	// fmt.Printf("Applying %s\n", ps)
	w.SetLineColor(colors.Color(ps.color))
	w.SetLineWidth(ps.width)
	w.SetLineDashPattern(ps.pattern)
	w.SetLineCapStyle(ps.Cap())
}

func (ps *PenStyle) Clone() *PenStyle {
	clone := *ps
	return &clone
}

func (ps *PenStyle) ID() string {
	return ps.id
}

func (ps *PenStyle) SetAttrs(attrs map[string]string) {
	if id, ok := attrs["id"]; ok {
		ps.id = id
	}
	if color, ok := attrs["color"]; ok {
		ps.color = NamedColor(color)
	}
	var units Units = "pt"
	units.SetAttrs(attrs)
	if width, ok := attrs["width"]; ok {
		ps.width = ParseMeasurement(width, units)
	}
	if pattern, ok := attrs["pattern"]; ok {
		ps.pattern = pattern
	}
	if cap, ok := attrs["cap"]; ok {
		switch cap {
		case "round_cap", "projecting_square_cap", "butt_cap":
			ps.cap = cap
		}
	}
}

func (ps *PenStyle) String() string {
	return fmt.Sprintf("PenStyle id=%s color=%v width=%f pattern=%s cap=%s", ps.id, ps.color, ps.width, ps.pattern, ps.cap)
}

const defaultPenPattern = "solid"
const defaultPenCap = "butt_cap"

func (ps *PenStyle) Cap() string {
	if ps.cap == "" {
		return defaultPenCap
	}
	return ps.cap
}

func PenStyleFor(id string, scope HasScope) *PenStyle {
	style, ok := scope.StyleFor(id)
	if !ok {
		style, ok = scope.StyleFor("pen_" + id)
	}
	if ok {
		ps, _ := style.(*PenStyle)
		return ps
	}
	ps := &PenStyle{id: "pen_" + id, color: NamedColor(id), pattern: defaultPenPattern, cap: defaultPenCap}
	scope.AddStyle(ps)
	return ps
}

var _ HasAttrs = (*PenStyle)(nil)
var _ Styler = (*PenStyle)(nil)

func init() {
	registerTag(DefaultSpace, "pen", func() any { return &PenStyle{} })
}
