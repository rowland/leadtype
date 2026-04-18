// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"
	"strings"
)

type BulletStyle struct {
	scope       HasScope
	id          string
	font        *FontStyle
	text        string
	src         string
	shape       string
	width       float64
	height      float64
	units       Units
	pen         *PenStyle
	brush       *BrushStyle
	alignX      string
	alignY      string
	sides       int
	points      int
	rotation    float64
	rotationSet bool
	r           float64
	rx          float64
	ry          float64
	r0          float64
}

func (bs *BulletStyle) AddText(text string) {
	bs.text = strings.TrimSpace(text)
}

func (bs *BulletStyle) Apply(w Writer) {
	// fmt.Printf("Applying %s\n", bs)
	if bs.font != nil {
		bs.font.Apply(w)
	}
}

func (bs *BulletStyle) ID() string {
	return bs.id
}

func (bs *BulletStyle) SetAttrs(attrs map[string]string) {
	if id, ok := attrs["id"]; ok {
		bs.id = id
	}
	if units, ok := attrs["units"]; ok {
		bs.units = Units(units)
	}
	if width, ok := attrs["width"]; ok {
		bs.width = ParseMeasurement(width, bs.units)
	}
	if height, ok := attrs["height"]; ok {
		bs.height = ParseMeasurement(height, bs.units)
	}
	if r, ok := attrs["r"]; ok {
		bs.r = ParseMeasurement(r, bs.units)
	}
	if rx, ok := attrs["rx"]; ok {
		bs.rx = ParseMeasurement(rx, bs.units)
	}
	if ry, ok := attrs["ry"]; ok {
		bs.ry = ParseMeasurement(ry, bs.units)
	}
	if font, ok := attrs["font"]; ok {
		bs.font = FontStyleFor(font, bs.scope)
	}
	if text, ok := attrs["text"]; ok {
		bs.text = strings.TrimSpace(text)
	}
	if src, ok := attrs["src"]; ok {
		bs.src = strings.TrimSpace(src)
	}
	if shape, ok := attrs["shape"]; ok {
		bs.shape = strings.TrimSpace(shape)
	}
	if pen, ok := attrs["pen"]; ok {
		bs.pen = PenStyleFor(pen, bs.scope)
	}
	if brush, ok := attrs["brush"]; ok {
		bs.brush = BrushStyleFor(brush, bs.scope)
	}
	if alignX, ok := attrs["align-x"]; ok {
		bs.alignX = strings.TrimSpace(alignX)
	}
	if alignY, ok := attrs["align-y"]; ok {
		bs.alignY = strings.TrimSpace(alignY)
	}
	if sides, ok := attrs["sides"]; ok {
		bs.sides, _ = strconv.Atoi(sides)
	}
	if points, ok := attrs["points"]; ok {
		bs.points, _ = strconv.Atoi(points)
	}
	if rotation, ok := attrs["rotation"]; ok {
		bs.rotation, _ = strconv.ParseFloat(rotation, 64)
		bs.rotationSet = true
	}
	if r0, ok := attrs["r0"]; ok {
		bs.r0 = ParseMeasurement(r0, bs.units)
	}
}

func (bs *BulletStyle) SetScope(scope HasScope) {
	bs.scope = scope
}

func (bs *BulletStyle) String() string {
	return fmt.Sprintf("BulletStyle id=%s text=%s src=%s shape=%s width=%f height=%f units=%s font=%s pen=%v brush=%v",
		bs.id, bs.text, bs.src, bs.shape, bs.width, bs.height, bs.units, bs.font, bs.pen, bs.brush)
}

func (bs *BulletStyle) Text() string {
	return bs.text
}

func (bs *BulletStyle) Width() float64 {
	return bs.width
}

func (bs *BulletStyle) Height() float64 {
	return bs.height
}

func (bs *BulletStyle) Radius() float64 {
	return bs.r
}

func (bs *BulletStyle) RadiusX() float64 {
	return bs.rx
}

func (bs *BulletStyle) RadiusY() float64 {
	return bs.ry
}

func (bs *BulletStyle) AlignX() string {
	if bs.alignX == "" {
		return "start"
	}
	return bs.alignX
}

func (bs *BulletStyle) AlignY() string {
	if bs.alignY == "" {
		return "top"
	}
	return bs.alignY
}

func (bs *BulletStyle) Source() string {
	return bs.src
}

func (bs *BulletStyle) Shape() string {
	return bs.shape
}

func (bs *BulletStyle) IsImage() bool {
	return strings.TrimSpace(bs.src) != ""
}

func (bs *BulletStyle) IsShape() bool {
	return !bs.IsImage() && strings.TrimSpace(bs.shape) != ""
}

func BulletStyleFor(id string, scope HasScope) *BulletStyle {
	if style, ok := scope.StyleFor(id); ok {
		bs, _ := style.(*BulletStyle)
		return bs
	}
	return nil
}

var _ HasAttrs = (*BulletStyle)(nil)
var _ Styler = (*BulletStyle)(nil)
var _ WantsScope = (*BulletStyle)(nil)

func init() {
	registerTag(DefaultSpace, "bullet", func() any { return &BulletStyle{width: 36} })
}
