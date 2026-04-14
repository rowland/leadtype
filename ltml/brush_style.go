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

func (bs *BrushStyle) SetAttrs(prefix string, attrs map[string]string) {
	if id, ok := attrs[prefix+"id"]; ok {
		bs.id = id
	}
	if kind, ok := attrs[prefix+"kind"]; ok {
		bs.kind = BrushKind(strings.TrimSpace(kind))
	}
	if color, ok := attrs[prefix+"color"]; ok {
		bs.color = NamedColor(color)
	}
	if stops, ok := attrs[prefix+"stops"]; ok {
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
		prefix+"x0",
		prefix+"y0",
		prefix+"x1",
		prefix+"y1",
	) {
		switch bs.Kind() {
		case BrushKindRadialGradient:
			gradient := bs.ensureRadialGradient()
			gradient.X0 = parseFloatAttr(attrs, prefix+"x0", gradient.X0)
			gradient.Y0 = parseFloatAttr(attrs, prefix+"y0", gradient.Y0)
			gradient.X1 = parseFloatAttr(attrs, prefix+"x1", gradient.X1)
			gradient.Y1 = parseFloatAttr(attrs, prefix+"y1", gradient.Y1)
		case BrushKindImage:
			// Ignore gradient geometry when explicitly configured as an image brush.
		default:
			gradient := bs.ensureLinearGradient()
			gradient.X0 = parseFloatAttr(attrs, prefix+"x0", gradient.X0)
			gradient.Y0 = parseFloatAttr(attrs, prefix+"y0", gradient.Y0)
			gradient.X1 = parseFloatAttr(attrs, prefix+"x1", gradient.X1)
			gradient.Y1 = parseFloatAttr(attrs, prefix+"y1", gradient.Y1)
			if bs.kind == "" {
				bs.kind = BrushKindLinearGradient
			}
		}
	}
	if hasAnyAttr(attrs, prefix+"r0", prefix+"r1") {
		gradient := bs.ensureRadialGradient()
		gradient.R0 = parseFloatAttr(attrs, prefix+"r0", gradient.R0)
		gradient.R1 = parseFloatAttr(attrs, prefix+"r1", gradient.R1)
		if bs.kind == "" {
			bs.kind = BrushKindRadialGradient
		}
	}
	if hasAnyAttr(attrs,
		prefix+"src",
		prefix+"fit",
		prefix+"anchor",
		prefix+"repeat",
		prefix+"opacity",
	) {
		image := bs.ensureImage()
		if value, ok := attrs[prefix+"src"]; ok {
			image.Src = strings.TrimSpace(value)
		}
		if value, ok := attrs[prefix+"fit"]; ok {
			image.Fit = strings.TrimSpace(value)
		}
		if value, ok := attrs[prefix+"anchor"]; ok {
			image.Anchor = strings.TrimSpace(value)
		}
		if value, ok := attrs[prefix+"repeat"]; ok {
			image.Repeat = strings.TrimSpace(value)
		}
		image.Opacity = parseFloatAttr(attrs, prefix+"opacity", image.Opacity)
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

var _ HasAttrsPrefix = (*BrushStyle)(nil)
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

func hasAnyAttr(attrs map[string]string, keys ...string) bool {
	for _, key := range keys {
		if _, ok := attrs[key]; ok {
			return true
		}
	}
	return false
}
