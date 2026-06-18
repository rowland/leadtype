// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/pdf"
)

type BrushKind string

const (
	BrushKindSolid          BrushKind = "solid"
	BrushKindLinearGradient BrushKind = "linear-gradient"
	BrushKindRadialGradient BrushKind = "radial-gradient"
	BrushKindImage          BrushKind = "image"
)

type BrushImageStyle struct {
	Src           string
	Fit           string
	Anchor        string
	Repeat        string
	Opacity       float64
	TileWidth     float64
	TileWidthPct  float64
	TileHeight    float64
	TileHeightPct float64
}

type linearGradientPct struct {
	X0 *float64
	Y0 *float64
	X1 *float64
	Y1 *float64
}

type radialGradientPct struct {
	X0 *float64
	Y0 *float64
	R0 *float64
	X1 *float64
	Y1 *float64
	R1 *float64
}

type BrushStyle struct {
	id             string
	kind           BrushKind
	color          colors.Color
	linearGradient *pdf.LinearGradient
	linearPct      *linearGradientPct
	radialGradient *pdf.RadialGradient
	radialPct      *radialGradientPct
	image          *BrushImageStyle
	opacity        *float64
}

func (bs *BrushStyle) Apply(w Writer) {
	if bs == nil || bs.Kind() != BrushKindSolid {
		return
	}
	w.SetFillColor(bs.color)
}

func (bs *BrushStyle) Clone() *BrushStyle {
	clone := *bs
	if bs.linearGradient != nil {
		gradientClone := *bs.linearGradient
		gradientClone.Stops = append([]pdf.GradientStop(nil), bs.linearGradient.Stops...)
		clone.linearGradient = &gradientClone
	}
	if bs.linearPct != nil {
		pctClone := *bs.linearPct
		pctClone.X0 = cloneFloat64Ptr(bs.linearPct.X0)
		pctClone.Y0 = cloneFloat64Ptr(bs.linearPct.Y0)
		pctClone.X1 = cloneFloat64Ptr(bs.linearPct.X1)
		pctClone.Y1 = cloneFloat64Ptr(bs.linearPct.Y1)
		clone.linearPct = &pctClone
	}
	if bs.radialGradient != nil {
		gradientClone := *bs.radialGradient
		gradientClone.Stops = append([]pdf.GradientStop(nil), bs.radialGradient.Stops...)
		clone.radialGradient = &gradientClone
	}
	if bs.radialPct != nil {
		pctClone := *bs.radialPct
		pctClone.X0 = cloneFloat64Ptr(bs.radialPct.X0)
		pctClone.Y0 = cloneFloat64Ptr(bs.radialPct.Y0)
		pctClone.R0 = cloneFloat64Ptr(bs.radialPct.R0)
		pctClone.X1 = cloneFloat64Ptr(bs.radialPct.X1)
		pctClone.Y1 = cloneFloat64Ptr(bs.radialPct.Y1)
		pctClone.R1 = cloneFloat64Ptr(bs.radialPct.R1)
		clone.radialPct = &pctClone
	}
	if bs.image != nil {
		imageClone := *bs.image
		clone.image = &imageClone
	}
	clone.opacity = cloneFloat64Ptr(bs.opacity)
	return &clone
}

func (bs *BrushStyle) ID() string {
	return bs.id
}

func (bs *BrushStyle) Kind() BrushKind {
	if bs == nil {
		return BrushKindSolid
	}
	if bs.kind != "" {
		return bs.kind
	}
	if bs.radialGradient != nil {
		return BrushKindRadialGradient
	}
	if bs.linearGradient != nil {
		return BrushKindLinearGradient
	}
	if bs.image != nil {
		return BrushKindImage
	}
	return BrushKindSolid
}

func (bs *BrushStyle) SetAttrs(attrs map[string]string) {
	if id, ok := attrs["id"]; ok {
		bs.id = id
	}
	var units Units = "pt"
	units.SetAttrs(attrs)
	if kind, ok := attrs["kind"]; ok {
		bs.kind = BrushKind(strings.TrimSpace(kind))
	}
	if color, ok := attrs["color"]; ok {
		bs.color = NamedColor(color)
	}
	if stops, ok := attrs["stops"]; ok {
		parsedStops := parseGradientStops(stops)
		if bs.Kind() == BrushKindRadialGradient {
			gradient := bs.ensureRadialGradient()
			gradient.Stops = parsedStops
		} else {
			gradient := bs.ensureLinearGradient()
			gradient.Stops = parsedStops
			if bs.kind == "" {
				bs.kind = BrushKindLinearGradient
			}
		}
	}
	if hasAnyAttr(attrs,
		"x0",
		"y0",
		"x1",
		"y1",
	) {
		switch bs.Kind() {
		case BrushKindRadialGradient:
			gradient := bs.ensureRadialGradient()
			pct := bs.ensureRadialPct()
			parseGradientMeasurementOrPctAttr(attrs, "x0", units, &gradient.X0, &pct.X0)
			parseGradientMeasurementOrPctAttr(attrs, "y0", units, &gradient.Y0, &pct.Y0)
			parseGradientMeasurementOrPctAttr(attrs, "x1", units, &gradient.X1, &pct.X1)
			parseGradientMeasurementOrPctAttr(attrs, "y1", units, &gradient.Y1, &pct.Y1)
		case BrushKindImage:
			// Ignore gradient geometry when explicitly configured as an image brush.
		default:
			gradient := bs.ensureLinearGradient()
			pct := bs.ensureLinearPct()
			parseGradientMeasurementOrPctAttr(attrs, "x0", units, &gradient.X0, &pct.X0)
			parseGradientMeasurementOrPctAttr(attrs, "y0", units, &gradient.Y0, &pct.Y0)
			parseGradientMeasurementOrPctAttr(attrs, "x1", units, &gradient.X1, &pct.X1)
			parseGradientMeasurementOrPctAttr(attrs, "y1", units, &gradient.Y1, &pct.Y1)
			if bs.kind == "" {
				bs.kind = BrushKindLinearGradient
			}
		}
	}
	if hasAnyAttr(attrs, "r0", "r1") {
		gradient := bs.ensureRadialGradient()
		pct := bs.ensureRadialPct()
		parseGradientMeasurementOrPctAttr(attrs, "r0", units, &gradient.R0, &pct.R0)
		parseGradientMeasurementOrPctAttr(attrs, "r1", units, &gradient.R1, &pct.R1)
		if bs.kind == "" {
			bs.kind = BrushKindRadialGradient
		}
	}
	if value, ok := attrs["opacity"]; ok && (bs.Kind() == BrushKindLinearGradient || bs.Kind() == BrushKindRadialGradient) {
		opacity := parseOpacityValue(value, 1)
		bs.opacity = &opacity
	}
	if hasAnyAttr(attrs,
		"src",
		"fit",
		"anchor",
		"repeat",
		"tile-width",
		"tile-height",
	) || hasImageOpacityAttr(attrs, bs) {
		image := bs.ensureImage()
		if value, ok := attrs["src"]; ok {
			image.Src = strings.TrimSpace(value)
		}
		if value, ok := attrs["fit"]; ok {
			image.Fit = strings.TrimSpace(value)
		}
		if value, ok := attrs["anchor"]; ok {
			image.Anchor = strings.TrimSpace(value)
		}
		if value, ok := attrs["repeat"]; ok {
			image.Repeat = strings.TrimSpace(value)
		}
		image.Opacity = parseOpacityAttr(attrs, "opacity", image.Opacity)
		if value, ok := attrs["tile-width"]; ok {
			image.TileWidth, image.TileWidthPct = parseMeasurementOrPct(strings.TrimSpace(value), units)
		}
		if value, ok := attrs["tile-height"]; ok {
			image.TileHeight, image.TileHeightPct = parseMeasurementOrPct(strings.TrimSpace(value), units)
		}
		if bs.kind == "" {
			bs.kind = BrushKindImage
		}
	}
}

func hasImageOpacityAttr(attrs map[string]string, bs *BrushStyle) bool {
	_, ok := attrs["opacity"]
	return ok && bs.Kind() == BrushKindImage
}

func (bs *BrushStyle) String() string {
	return fmt.Sprintf("BrushStyle id=%s kind=%s color=%v", bs.id, bs.Kind(), bs.color)
}

func BrushStyleFor(id string, scope HasScope) *BrushStyle {
	if scope == nil {
		return &BrushStyle{id: "brush_" + id, color: NamedColor(id)}
	}
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

// SetBrushStyle sets a brush style field from attrName and any prefixed
// overrides in attrs. Overrides are applied to a clone so a style resolved
// from scope is not mutated.
//
// A third-party widget can use this from SetAttrs with its own brush field:
//
//	SetBrushStyle(&w.textFill, "text-fill", attrs, w.Scope(), w.Units())
func SetBrushStyle(field **BrushStyle, attrName string, attrs map[string]string, scope HasScope, units Units) {
	if id, ok := attrs[attrName]; ok {
		*field = BrushStyleFor(id, scope)
	}
	prefix := attrName + "."
	if !MapHasKeyPrefix(attrs, prefix) {
		return
	}
	if *field == nil {
		*field = &BrushStyle{}
	} else {
		*field = (*field).Clone()
	}
	(*field).SetAttrs(addUnits(filterMapAttrs(prefix, attrs), units))
}

var _ HasAttrs = (*BrushStyle)(nil)
var _ Styler = (*BrushStyle)(nil)

func init() {
	registerTag(DefaultSpace, "brush", func() any { return &BrushStyle{} })
}

func (bs *BrushStyle) ensureLinearGradient() *pdf.LinearGradient {
	if bs.linearGradient == nil {
		bs.linearGradient = &pdf.LinearGradient{}
	}
	return bs.linearGradient
}

func (bs *BrushStyle) ensureRadialGradient() *pdf.RadialGradient {
	if bs.radialGradient == nil {
		bs.radialGradient = &pdf.RadialGradient{}
	}
	return bs.radialGradient
}

func (bs *BrushStyle) ensureLinearPct() *linearGradientPct {
	if bs.linearPct == nil {
		bs.linearPct = &linearGradientPct{}
	}
	return bs.linearPct
}

func (bs *BrushStyle) ensureRadialPct() *radialGradientPct {
	if bs.radialPct == nil {
		bs.radialPct = &radialGradientPct{}
	}
	return bs.radialPct
}

func (bs *BrushStyle) ensureImage() *BrushImageStyle {
	if bs.image == nil {
		bs.image = &BrushImageStyle{Opacity: 1}
	}
	return bs.image
}

func parseMeasurementOrPct(value string, units Units) (measurement, pct float64) {
	if rePct.MatchString(value) {
		pct, _ = strconv.ParseFloat(value[:len(value)-1], 64)
		return 0, pct
	}
	return ParseMeasurement(value, units), 0
}

func parseGradientMeasurementOrPctAttr(attrs map[string]string, key string, units Units, measurement *float64, pct **float64) {
	value, ok := attrs[key]
	if !ok {
		return
	}
	parsedMeasurement, parsedPct := parseMeasurementOrPct(strings.TrimSpace(value), units)
	if rePct.MatchString(strings.TrimSpace(value)) {
		*measurement = 0
		*pct = float64Ptr(parsedPct)
		return
	}
	*measurement = parsedMeasurement
	*pct = nil
}

func cloneFloat64Ptr(value *float64) *float64 {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func float64Ptr(value float64) *float64 {
	return &value
}

func parseGradientStops(value string) []pdf.GradientStop {
	parts := strings.Split(value, ",")
	stops := make([]pdf.GradientStop, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		positionText, colorText, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		position, err := strconv.ParseFloat(strings.TrimSpace(positionText), 64)
		if err != nil {
			continue
		}
		stops = append(stops, pdf.GradientStop{
			Position: position,
			Color:    NamedColor(strings.TrimSpace(colorText)),
		})
	}
	return stops
}

func parseOpacityAttr(attrs map[string]string, key string, defaultValue float64) float64 {
	value, ok := attrs[key]
	if !ok {
		return defaultValue
	}
	return parseOpacityValue(value, defaultValue)
}

func parseOpacityValue(value string, defaultValue float64) float64 {
	value = strings.TrimSpace(value)
	if before, ok := strings.CutSuffix(value, "%"); ok {
		parsed, err := strconv.ParseFloat(strings.TrimSpace(before), 64)
		if err != nil {
			return defaultValue
		}
		return parsed / 100.0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func hasAnyAttr(attrs map[string]string, keys ...string) bool {
	for _, key := range keys {
		if _, ok := attrs[key]; ok {
			return true
		}
	}
	return false
}
