// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type StdLine struct {
	StdWidget
	angle       float64
	angleSet    bool
	length      float64
	style       *PenStyle
	markerStart string
	markerEnd   string
}

func (l *StdLine) DrawContent(w Writer) error {
	return withGraphicAccessibility(w, &l.StdWidget, "Figure", func() error {
		if sector, ok := l.Container().(*StdSector); ok && l.Position() == Static && !l.angleSet {
			return sector.drawSectorLine(l, w)
		}
		style := l.Style()
		if style != nil {
			x := l.Left() + l.MarginLeft()
			y := l.Top() + l.MarginTop()
			width := l.Width() - l.MarginLeft() - l.MarginRight()
			height := l.Height() - l.MarginTop() - l.MarginBottom()
			if err := style.ApplyInRect(w, x, y, width, height); err != nil {
				return err
			}
		}
		return l.drawStraightLine(w, style)
	})
}

func (l *StdLine) drawStraightLine(w Writer, style *PenStyle) error {
	startMarker, endMarker, err := l.resolvedMarkers(style)
	if err != nil {
		return err
	}
	x, y := l.originForQuadrant()
	length := l.Length()
	startMarker, endMarker = fitResolvedMarkers(startMarker, endMarker, length)
	theta := degreesToRadians(l.Angle())
	dx, dy := math.Cos(theta), -math.Sin(theta)
	startRefX, startRefY := x, y
	endRefX, endRefY := x+dx*length, y+dy*length
	if startMarker != nil {
		inset := startMarker.referenceInset()
		startRefX += dx * inset
		startRefY += dy * inset
	}
	if endMarker != nil {
		inset := endMarker.referenceInset()
		endRefX -= dx * inset
		endRefY -= dy * inset
	}
	shaftStartX, shaftStartY := startRefX, startRefY
	shaftEndX, shaftEndY := endRefX, endRefY
	if startMarker != nil {
		shaftStartX += dx * startMarker.cutback
		shaftStartY += dy * startMarker.cutback
	}
	if endMarker != nil {
		shaftEndX -= dx * endMarker.cutback
		shaftEndY -= dy * endMarker.cutback
	}
	shaftLength := (shaftEndX-shaftStartX)*dx + (shaftEndY-shaftStartY)*dy
	if shaftLength > 0 {
		w.Line(shaftStartX, shaftStartY, l.Angle(), shaftLength)
	}
	if startMarker != nil {
		if err := startMarker.draw(w, startRefX, startRefY, l.Angle()+180, style.Color()); err != nil {
			return err
		}
	}
	if endMarker != nil {
		if err := endMarker.draw(w, endRefX, endRefY, l.Angle(), style.Color()); err != nil {
			return err
		}
	}
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

func (l *StdLine) PreferredHeight(Writer) (float64, error) {
	if l.height != 0 {
		return float64(l.height), nil
	}
	cross, err := l.markerContentCrossSize()
	if err != nil {
		return 0, err
	}
	cos, sin := lineAxisFactors(l.Angle())
	return sin*l.Length() + cos*cross + NonContentHeight(l), nil
}

func (l *StdLine) PreferredWidth(Writer) (float64, error) {
	if l.width != 0 {
		return float64(l.width), nil
	}
	cross, err := l.markerContentCrossSize()
	if err != nil {
		return 0, err
	}
	cos, sin := lineAxisFactors(l.Angle())
	return cos*l.Length() + sin*cross + NonContentWidth(l), nil
}

func (l *StdLine) SetAttrs(attrs map[string]string) {
	l.StdWidget.SetAttrs(attrs)
	if angle, ok := attrs["angle"]; ok {
		l.angle, _ = strconv.ParseFloat(angle, 64)
		l.angleSet = true
	}
	if marker, ok := attrs["marker-start"]; ok {
		l.markerStart = strings.TrimSpace(marker)
	}
	if marker, ok := attrs["marker-end"]; ok {
		l.markerEnd = strings.TrimSpace(marker)
	}
	if length, ok := attrs["length"]; ok {
		l.length = ParseMeasurement(length, l.Units())
	}
	if style, ok := attrs["style"]; ok {
		l.style = PenStyleFor(style, l.scope)
	}
	if MapHasKeyPrefix(attrs, "style.") {
		switch {
		case l.style != nil:
			l.style = l.style.Clone()
		case l.scope != nil:
			l.style = l.Style().Clone()
		default:
			l.style = &PenStyle{pattern: defaultPenPattern, cap: defaultPenCap}
		}
		l.style.SetAttrs(addUnits(filterMapAttrs("style.", attrs), l.Units()))
	}
}

func (l *StdLine) markerIDs(style *PenStyle) (start, end string) {
	if style != nil {
		start, end = style.MarkerStart(), style.MarkerEnd()
	}
	if l.markerStart != "" {
		start = l.markerStart
	}
	if l.markerEnd != "" {
		end = l.markerEnd
	}
	return
}

func (l *StdLine) resolvedMarkers(style *PenStyle) (start, end *resolvedMarker, err error) {
	startID, endID := l.markerIDs(style)
	start, err = resolveMarker(l.scope, startID, style.Width())
	if err != nil {
		return nil, nil, err
	}
	end, err = resolveMarker(l.scope, endID, style.Width())
	if err != nil {
		return nil, nil, err
	}
	return
}

func (l *StdLine) markerCrossSize() (float64, error) {
	height, err := l.markerContentCrossSize()
	if err != nil {
		return 0, err
	}
	height = max(height, l.Style().Width())
	return height + NonContentHeight(l), nil
}

func (l *StdLine) markerContentCrossSize() (float64, error) {
	start, end, err := l.resolvedMarkers(l.Style())
	if err != nil {
		return 0, err
	}
	height := 0.0
	if start != nil {
		height = max(height, start.height)
	}
	if end != nil {
		height = max(height, end.height)
	}
	return height, nil
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
	cos, sin := lineAxisFactors(l.Angle())

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
	cos, sin := lineAxisFactors(l.Angle())
	prefWidth := cos * l.Length()
	prefHeight := sin * l.Length()
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

func lineAxisFactors(angle float64) (cos, sin float64) {
	theta := degreesToRadians(angle)
	cos = math.Abs(math.Cos(theta))
	sin = math.Abs(math.Sin(theta))
	if cos < 1e-12 {
		cos = 0
	}
	if sin < 1e-12 {
		sin = 0
	}
	return cos, sin
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
	registerTag(DefaultSpace, "line", func() any { return &StdLine{} })
}

var _ HasAttrs = (*StdLine)(nil)
var _ Identifier = (*StdLine)(nil)
var _ Printer = (*StdLine)(nil)
var _ WantsContainer = (*StdLine)(nil)
