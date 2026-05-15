// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/pdf"
)

type PenKind string

const (
	PenKindSolid          PenKind = "solid"
	PenKindLinearGradient PenKind = "linear-gradient"
	PenKindRadialGradient PenKind = "radial-gradient"
)

type PenStyle struct {
	id             string
	kind           PenKind
	color          colors.Color
	width          float64
	pattern        string
	cap            string
	linearGradient *pdf.LinearGradient
	linearPct      *linearGradientPct
	radialGradient *pdf.RadialGradient
	radialPct      *radialGradientPct
}

func (ps *PenStyle) Apply(w Writer) {
	// fmt.Printf("Applying %s\n", ps)
	w.ClearLineGradient()
	w.SetLineColor(colors.Color(ps.color))
	w.SetLineWidth(ps.width)
	w.SetLineDashPattern(ps.pattern)
	w.SetLineCapStyle(ps.Cap())
}

func (ps *PenStyle) ApplyInRect(w Writer, x, y, width, height float64) error {
	if ps == nil {
		return nil
	}
	w.SetLineWidth(ps.width)
	w.SetLineDashPattern(ps.pattern)
	w.SetLineCapStyle(ps.Cap())
	switch ps.Kind() {
	case PenKindLinearGradient:
		if ps.linearGradient == nil {
			ps.Apply(w)
			return nil
		}
		return w.SetLineLinearGradient(resolveLinearGradient(ps.linearGradient, ps.linearPct, x, y, width, height))
	case PenKindRadialGradient:
		if ps.radialGradient == nil {
			ps.Apply(w)
			return nil
		}
		return w.SetLineRadialGradient(resolveRadialGradient(ps.radialGradient, ps.radialPct, x, y, width, height))
	default:
		ps.Apply(w)
		return nil
	}
}

func (ps *PenStyle) Clone() *PenStyle {
	clone := *ps
	if ps.linearGradient != nil {
		gradientClone := *ps.linearGradient
		gradientClone.Stops = append([]pdf.GradientStop(nil), ps.linearGradient.Stops...)
		clone.linearGradient = &gradientClone
	}
	if ps.linearPct != nil {
		pctClone := *ps.linearPct
		pctClone.X0 = cloneFloat64Ptr(ps.linearPct.X0)
		pctClone.Y0 = cloneFloat64Ptr(ps.linearPct.Y0)
		pctClone.X1 = cloneFloat64Ptr(ps.linearPct.X1)
		pctClone.Y1 = cloneFloat64Ptr(ps.linearPct.Y1)
		clone.linearPct = &pctClone
	}
	if ps.radialGradient != nil {
		gradientClone := *ps.radialGradient
		gradientClone.Stops = append([]pdf.GradientStop(nil), ps.radialGradient.Stops...)
		clone.radialGradient = &gradientClone
	}
	if ps.radialPct != nil {
		pctClone := *ps.radialPct
		pctClone.X0 = cloneFloat64Ptr(ps.radialPct.X0)
		pctClone.Y0 = cloneFloat64Ptr(ps.radialPct.Y0)
		pctClone.R0 = cloneFloat64Ptr(ps.radialPct.R0)
		pctClone.X1 = cloneFloat64Ptr(ps.radialPct.X1)
		pctClone.Y1 = cloneFloat64Ptr(ps.radialPct.Y1)
		pctClone.R1 = cloneFloat64Ptr(ps.radialPct.R1)
		clone.radialPct = &pctClone
	}
	return &clone
}

func (ps *PenStyle) ID() string {
	return ps.id
}

func (ps *PenStyle) Kind() PenKind {
	if ps == nil {
		return PenKindSolid
	}
	if ps.kind != "" {
		return ps.kind
	}
	if ps.radialGradient != nil {
		return PenKindRadialGradient
	}
	if ps.linearGradient != nil {
		return PenKindLinearGradient
	}
	return PenKindSolid
}

func (ps *PenStyle) SetAttrs(attrs map[string]string) {
	if id, ok := attrs["id"]; ok {
		ps.id = id
	}
	if kind, ok := attrs["kind"]; ok {
		ps.kind = PenKind(strings.TrimSpace(kind))
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
	if stops, ok := attrs["stops"]; ok {
		parsedStops := parseGradientStops(stops)
		if ps.Kind() == PenKindRadialGradient {
			gradient := ps.ensureRadialGradient()
			gradient.Stops = parsedStops
		} else {
			gradient := ps.ensureLinearGradient()
			gradient.Stops = parsedStops
			if ps.kind == "" {
				ps.kind = PenKindLinearGradient
			}
		}
	}
	if hasAnyAttr(attrs, "x0", "y0", "x1", "y1") {
		switch ps.Kind() {
		case PenKindRadialGradient:
			gradient := ps.ensureRadialGradient()
			pct := ps.ensureRadialPct()
			parseGradientMeasurementOrPctAttr(attrs, "x0", units, &gradient.X0, &pct.X0)
			parseGradientMeasurementOrPctAttr(attrs, "y0", units, &gradient.Y0, &pct.Y0)
			parseGradientMeasurementOrPctAttr(attrs, "x1", units, &gradient.X1, &pct.X1)
			parseGradientMeasurementOrPctAttr(attrs, "y1", units, &gradient.Y1, &pct.Y1)
		default:
			gradient := ps.ensureLinearGradient()
			pct := ps.ensureLinearPct()
			parseGradientMeasurementOrPctAttr(attrs, "x0", units, &gradient.X0, &pct.X0)
			parseGradientMeasurementOrPctAttr(attrs, "y0", units, &gradient.Y0, &pct.Y0)
			parseGradientMeasurementOrPctAttr(attrs, "x1", units, &gradient.X1, &pct.X1)
			parseGradientMeasurementOrPctAttr(attrs, "y1", units, &gradient.Y1, &pct.Y1)
			if ps.kind == "" {
				ps.kind = PenKindLinearGradient
			}
		}
	}
	if hasAnyAttr(attrs, "r0", "r1") {
		gradient := ps.ensureRadialGradient()
		pct := ps.ensureRadialPct()
		parseGradientMeasurementOrPctAttr(attrs, "r0", units, &gradient.R0, &pct.R0)
		parseGradientMeasurementOrPctAttr(attrs, "r1", units, &gradient.R1, &pct.R1)
		if ps.kind == "" {
			ps.kind = PenKindRadialGradient
		}
	}
}

func (ps *PenStyle) String() string {
	return fmt.Sprintf("PenStyle id=%s kind=%s color=%v width=%f pattern=%s cap=%s", ps.id, ps.Kind(), ps.color, ps.width, ps.pattern, ps.cap)
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
	if scope == nil {
		return &PenStyle{id: "pen_" + id, color: NamedColor(id), pattern: defaultPenPattern, cap: defaultPenCap}
	}
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

func (ps *PenStyle) ensureLinearGradient() *pdf.LinearGradient {
	if ps.linearGradient == nil {
		ps.linearGradient = &pdf.LinearGradient{}
	}
	return ps.linearGradient
}

func (ps *PenStyle) ensureRadialGradient() *pdf.RadialGradient {
	if ps.radialGradient == nil {
		ps.radialGradient = &pdf.RadialGradient{}
	}
	return ps.radialGradient
}

func (ps *PenStyle) ensureLinearPct() *linearGradientPct {
	if ps.linearPct == nil {
		ps.linearPct = &linearGradientPct{}
	}
	return ps.linearPct
}

func (ps *PenStyle) ensureRadialPct() *radialGradientPct {
	if ps.radialPct == nil {
		ps.radialPct = &radialGradientPct{}
	}
	return ps.radialPct
}

var _ HasAttrs = (*PenStyle)(nil)
var _ Styler = (*PenStyle)(nil)

func init() {
	registerTag(DefaultSpace, "pen", func() any { return &PenStyle{} })
}
