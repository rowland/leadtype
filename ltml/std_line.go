// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"math"
	"strconv"
)

type StdLine struct {
	StdWidget
	angle  float64
	length float64
	style  *PenStyle
}

func (l *StdLine) DrawContent(w Writer) error {
	style := l.Style()
	if style != nil {
		style.Apply(w)
	}
	x, y := l.originForQuadrant()
	w.Line(x, y, l.Angle(), l.Length())
	return nil
}

func (l *StdLine) Angle() float64 {
	return l.angle
}

func (l *StdLine) Length() float64 {
	if l.length > 0 {
		return l.length
	}
	return l.calcLength()
}

func (l *StdLine) PreferredHeight(Writer) float64 {
	if l.height != 0 {
		return l.height
	}
	return math.Abs(math.Sin(degreesToRadians(l.Angle())))*l.Length() + NonContentHeight(l)
}

func (l *StdLine) PreferredWidth(Writer) float64 {
	if l.width != 0 {
		return l.width
	}
	return math.Abs(math.Cos(degreesToRadians(l.Angle())))*l.Length() + NonContentWidth(l)
}

func (l *StdLine) SetAttrs(attrs map[string]string) {
	l.StdWidget.SetAttrs(attrs)
	if angle, ok := attrs["angle"]; ok {
		l.angle, _ = strconv.ParseFloat(angle, 64)
	}
	if length, ok := attrs["length"]; ok {
		l.length = ParseMeasurement(length, l.Units())
	}
	if style, ok := attrs["style"]; ok {
		l.style = PenStyleFor(style, l.scope)
	}
}

func (l *StdLine) String() string {
	return fmt.Sprintf("StdLine angle=%v length=%v %s", l.angle, l.length, &l.StdWidget)
}

func (l *StdLine) Style() *PenStyle {
	if l.style != nil {
		return l.style
	}
	return PenStyleFor("solid", l.scope)
}

func (l *StdLine) calcLength() float64 {
	contentWidth := ContentWidth(l)
	contentHeight := ContentHeight(l)
	theta := degreesToRadians(l.Angle())
	cos := math.Abs(math.Cos(theta))
	sin := math.Abs(math.Sin(theta))

	switch {
	case contentWidth > 0 && contentHeight > 0:
		length := math.Hypot(contentWidth, contentHeight)
		width := cos * length
		height := sin * length
		if width > contentWidth && width > 0 {
			length *= contentWidth / width
		}
		if height > contentHeight && height > 0 {
			length *= contentHeight / height
		}
		return length
	case contentWidth > 0 && cos > 0:
		return contentWidth / cos
	case contentHeight > 0 && sin > 0:
		return contentHeight / sin
	default:
		return 0
	}
}

func (l *StdLine) originForQuadrant() (float64, float64) {
	contentWidth := ContentWidth(l)
	contentHeight := ContentHeight(l)
	prefWidth := math.Abs(math.Cos(degreesToRadians(l.Angle())) * l.Length())
	prefHeight := math.Abs(math.Sin(degreesToRadians(l.Angle())) * l.Length())
	xOffset := (contentWidth - prefWidth) / 2
	yOffset := (contentHeight - prefHeight) / 2
	left := l.Left() + l.MarginLeft() + l.PaddingLeft()
	top := l.Top() + l.MarginTop() + l.PaddingTop()
	right := left + contentWidth
	bottom := top + contentHeight

	switch quadrant(l.Angle()) {
	case 1:
		return left + xOffset, bottom - yOffset
	case 2:
		return right - xOffset, bottom - yOffset
	case 3:
		return right - xOffset, top + yOffset
	default:
		return left + xOffset, top + yOffset
	}
}

func degreesToRadians(angle float64) float64 {
	return angle * math.Pi / 180.0
}

func quadrant(angle float64) int {
	a := math.Mod(angle, 360)
	if a < 0 {
		a += 360
	}
	switch {
	case a <= 90:
		return 1
	case a <= 180:
		return 2
	case a <= 270:
		return 3
	default:
		return 4
	}
}

func init() {
	registerTag(DefaultSpace, "line", func() interface{} { return &StdLine{} })
}

var _ HasAttrs = (*StdLine)(nil)
var _ Identifier = (*StdLine)(nil)
var _ Printer = (*StdLine)(nil)
var _ WantsContainer = (*StdLine)(nil)
