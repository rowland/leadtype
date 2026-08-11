// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"regexp"
	"strconv"
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
	widthSet       bool
	pattern        string
	cap            string
	markerStart    string
	markerEnd      string
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
		ps.widthSet = true
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
	if marker, ok := attrs["marker-start"]; ok {
		ps.markerStart = strings.TrimSpace(marker)
	}
	if marker, ok := attrs["marker-end"]; ok {
		ps.markerEnd = strings.TrimSpace(marker)
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

var rePenWidth = regexp.MustCompile(`^\+?(?:\d+(?:\.\d*)?|\.\d+)([a-z]+)?$`)

func (ps *PenStyle) Cap() string {
	if ps.cap == "" {
		return defaultPenCap
	}
	return ps.cap
}

// Width returns the pen stroke width in points.
func (ps *PenStyle) Width() float64 {
	if ps == nil {
		return 0
	}
	return ps.width
}

func (ps *PenStyle) hasWidth() bool {
	return ps != nil && (ps.widthSet || ps.width != 0)
}

func (ps *PenStyle) Color() colors.Color {
	if ps == nil {
		return 0
	}
	return ps.color
}

func (ps *PenStyle) MarkerStart() string {
	if ps == nil {
		return ""
	}
	return ps.markerStart
}

func (ps *PenStyle) MarkerEnd() string {
	if ps == nil {
		return ""
	}
	return ps.markerEnd
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

// penStyleForValue resolves a named pen or color, or parses an ad hoc
// [width] [pattern] [color] description. Ad hoc pens are not added to scope.
func penStyleForValue(value string, scope HasScope, units Units) *PenStyle {
	pen, colorOnly, ok := resolvePenStyleValue(value, scope, units)
	if ok {
		if colorOnly {
			return PenStyleFor(strings.TrimSpace(value), scope)
		}
		return pen
	}
	return PenStyleFor(strings.TrimSpace(value), scope)
}

func resolvePenStyleValue(value string, scope HasScope, units Units) (*PenStyle, bool, bool) {
	value = strings.TrimSpace(value)
	if style, ok := namedPenStyleFor(value, scope); ok {
		return style, false, true
	}

	if pen, colorOnly, ok := parsePenShorthand(value, units); ok {
		return pen, colorOnly, true
	}
	return nil, false, false
}

func namedPenStyleFor(id string, scope HasScope) (*PenStyle, bool) {
	if scope == nil {
		return nil, false
	}
	if style, ok := scope.StyleFor(id); ok {
		pen, isPen := style.(*PenStyle)
		return pen, isPen
	}
	if style, ok := scope.StyleFor("pen_" + id); ok {
		pen, isPen := style.(*PenStyle)
		return pen, isPen
	}
	return nil, false
}

func parsePenShorthand(value string, units Units) (*PenStyle, bool, bool) {
	tokens := strings.Fields(value)
	if len(tokens) == 0 {
		return nil, false, false
	}

	pen := &PenStyle{
		color:   NamedColor("black"),
		pattern: defaultPenPattern,
		cap:     defaultPenCap,
	}
	hasWidth, hasPattern, hasColor := false, false, false
	for _, token := range tokens {
		if width, ok := parsePenWidth(token, units); ok {
			if hasWidth {
				return nil, false, false
			}
			pen.width = width
			pen.widthSet = true
			hasWidth = true
			continue
		}
		switch token {
		case "solid", "dashed", "dotted":
			if hasPattern {
				return nil, false, false
			}
			pen.pattern = token
			hasPattern = true
			continue
		}
		if color, ok := parseLTMLColor(token); ok {
			if hasColor {
				return nil, false, false
			}
			pen.color = color
			hasColor = true
			continue
		}
		return nil, false, false
	}

	return pen, hasColor && !hasWidth && !hasPattern && len(tokens) == 1, true
}

func parsePenWidth(value string, units Units) (float64, bool) {
	matches := rePenWidth.FindStringSubmatch(value)
	if len(matches) != 2 {
		return 0, false
	}
	measurementUnits := units
	if matches[1] != "" {
		measurementUnits = Units(matches[1])
		if _, ok := UnitConversions[measurementUnits]; !ok {
			return 0, false
		}
	}
	number := strings.TrimSuffix(value, matches[1])
	width, err := strconv.ParseFloat(number, 64)
	if err != nil || width < 0 {
		return 0, false
	}
	return FromUnits(width, measurementUnits), true
}

// SetPenStyle sets a pen style field from attrName and any prefixed overrides
// in attrs. Overrides are applied to a clone so a style resolved from scope is
// not mutated.
//
// A third-party widget can use this from SetAttrs with its own pen field:
//
//	SetPenStyle(&w.outline, "outline", attrs, w.Scope(), w.Units())
func SetPenStyle(field **PenStyle, attrName string, attrs map[string]string, scope HasScope, units Units) {
	if id, ok := attrs[attrName]; ok {
		*field = penStyleForValue(id, scope, units)
	}
	prefix := attrName + "."
	if !MapHasKeyPrefix(attrs, prefix) {
		return
	}
	if *field == nil {
		*field = &PenStyle{pattern: defaultPenPattern, cap: defaultPenCap}
	} else {
		*field = (*field).Clone()
	}
	(*field).SetAttrs(addUnits(filterMapAttrs(prefix, attrs), units))
}

// setOptionalPenStyle applies a border-style pen reference while preserving
// the distinction between an absent value and an explicit lowercase "none".
// Once disabled, prefixed pen attributes remain dormant until a later layer
// supplies an explicit non-none pen reference.
func setOptionalPenStyle(field **PenStyle, explicitlySet *bool, attrName string, attrs map[string]string, scope HasScope, units Units, fallback *PenStyle) {
	if id, ok := attrs[attrName]; ok {
		*explicitlySet = true
		if strings.TrimSpace(id) == "none" {
			*field = nil
		} else {
			*field = penStyleForValue(id, scope, units)
		}
	}
	prefix := attrName + "."
	if !MapHasKeyPrefix(attrs, prefix) || (*explicitlySet && *field == nil) {
		return
	}
	*explicitlySet = true
	base := *field
	if base == nil {
		base = fallback
	}
	if base == nil {
		*field = &PenStyle{pattern: defaultPenPattern, cap: defaultPenCap}
	} else {
		*field = base.Clone()
	}
	(*field).SetAttrs(addUnits(filterMapAttrs(prefix, attrs), units))
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
