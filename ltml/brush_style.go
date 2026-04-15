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
	Src     string
	Fit     string
	Anchor  string
	Repeat  string
	Opacity float64
}

type BrushStyle struct {
	id             string
	kind           BrushKind
	color          colors.Color
	linearGradient *pdf.LinearGradient
	radialGradient *pdf.RadialGradient
	image          *BrushImageStyle
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
	if bs.radialGradient != nil {
		gradientClone := *bs.radialGradient
		gradientClone.Stops = append([]pdf.GradientStop(nil), bs.radialGradient.Stops...)
		clone.radialGradient = &gradientClone
	}
	if bs.image != nil {
		imageClone := *bs.image
		clone.image = &imageClone
	}
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
			gradient.X0 = parseMeasurementAttr(attrs, "x0", units, gradient.X0)
			gradient.Y0 = parseMeasurementAttr(attrs, "y0", units, gradient.Y0)
			gradient.X1 = parseMeasurementAttr(attrs, "x1", units, gradient.X1)
			gradient.Y1 = parseMeasurementAttr(attrs, "y1", units, gradient.Y1)
		case BrushKindImage:
			// Ignore gradient geometry when explicitly configured as an image brush.
		default:
			gradient := bs.ensureLinearGradient()
			gradient.X0 = parseMeasurementAttr(attrs, "x0", units, gradient.X0)
			gradient.Y0 = parseMeasurementAttr(attrs, "y0", units, gradient.Y0)
			gradient.X1 = parseMeasurementAttr(attrs, "x1", units, gradient.X1)
			gradient.Y1 = parseMeasurementAttr(attrs, "y1", units, gradient.Y1)
			if bs.kind == "" {
				bs.kind = BrushKindLinearGradient
			}
		}
	}
	if hasAnyAttr(attrs, "r0", "r1") {
		gradient := bs.ensureRadialGradient()
		gradient.R0 = parseMeasurementAttr(attrs, "r0", units, gradient.R0)
		gradient.R1 = parseMeasurementAttr(attrs, "r1", units, gradient.R1)
		if bs.kind == "" {
			bs.kind = BrushKindRadialGradient
		}
	}
	if hasAnyAttr(attrs,
		"src",
		"fit",
		"anchor",
		"repeat",
		"opacity",
	) {
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
		image.Opacity = parseFloatAttr(attrs, "opacity", image.Opacity)
		if bs.kind == "" {
			bs.kind = BrushKindImage
		}
	}
}

func (bs *BrushStyle) String() string {
	return fmt.Sprintf("BrushStyle id=%s kind=%s color=%v", bs.id, bs.Kind(), bs.color)
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

func (bs *BrushStyle) ensureImage() *BrushImageStyle {
	if bs.image == nil {
		bs.image = &BrushImageStyle{Opacity: 1}
	}
	return bs.image
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

func parseFloatAttr(attrs map[string]string, key string, defaultValue float64) float64 {
	value, ok := attrs[key]
	if !ok {
		return defaultValue
	}
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func parseMeasurementAttr(attrs map[string]string, key string, units Units, defaultValue float64) float64 {
	value, ok := attrs[key]
	if !ok {
		return defaultValue
	}
	return ParseMeasurement(strings.TrimSpace(value), units)
}

func hasAnyAttr(attrs map[string]string, keys ...string) bool {
	for _, key := range keys {
		if _, ok := attrs[key]; ok {
			return true
		}
	}
	return false
}
