// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/rowland/leadtype/colors"
)

type markerUnits uint8

const (
	markerUnitsStrokeWidth markerUnits = iota
	markerUnitsUserSpace
)

type markerCoordinate struct {
	value   float64
	percent bool
}

type StdMarker struct {
	StdContainer
	markerWidth  float64
	markerHeight float64
	refX         markerCoordinate
	refY         markerCoordinate
	stemCutback  float64
	markerUnits  markerUnits
	builtin      string
	parseErr     error
}

func (m *StdMarker) SetAttrs(attrs map[string]string) {
	m.StdContainer.SetAttrs(attrs)
	m.parseErr = nil
	m.markerUnits = markerUnitsStrokeWidth
	switch units := strings.TrimSpace(attrs["marker-units"]); units {
	case "", "stroke-width":
	case "user-space":
		m.markerUnits = markerUnitsUserSpace
	default:
		m.parseErr = fmt.Errorf("invalid marker-units %q", units)
	}
	m.markerWidth = m.parseSize(attrs["width"])
	m.markerHeight = m.parseSize(attrs["height"])
	m.refX = m.parseCoordinate(attrs["ref-x"], 50)
	m.refY = m.parseCoordinate(attrs["ref-y"], 50)
	m.stemCutback = m.parseSize(attrs["stem-cutback"])
}

func (m *StdMarker) parseSize(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if m.markerUnits == markerUnitsUserSpace {
		return ParseMeasurement(value, m.Units())
	}
	result, err := strconv.ParseFloat(value, 64)
	if err != nil && m.parseErr == nil {
		m.parseErr = fmt.Errorf("invalid stroke-width marker measurement %q", value)
	}
	return result
}

func (m *StdMarker) parseCoordinate(value string, fallbackPercent float64) markerCoordinate {
	value = strings.TrimSpace(value)
	if value == "" {
		return markerCoordinate{value: fallbackPercent, percent: true}
	}
	if strings.HasSuffix(value, "%") {
		number, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(value, "%")), 64)
		if err != nil && m.parseErr == nil {
			m.parseErr = fmt.Errorf("invalid marker percentage %q", value)
		}
		return markerCoordinate{value: number, percent: true}
	}
	number := 0.0
	if m.markerUnits == markerUnitsUserSpace {
		number = ParseMeasurement(value, m.Units())
	} else {
		var err error
		number, err = strconv.ParseFloat(value, 64)
		if err != nil && m.parseErr == nil {
			m.parseErr = fmt.Errorf("invalid marker coordinate %q", value)
		}
	}
	return markerCoordinate{value: number}
}

func (m *StdMarker) validateDefinition() error {
	if m.parseErr != nil {
		return fmt.Errorf("<marker id=%q>: %w", m.ID, m.parseErr)
	}
	if strings.TrimSpace(m.ID) == "" {
		return fmt.Errorf("<marker> requires an id")
	}
	if m.markerWidth <= 0 || m.markerHeight <= 0 {
		return fmt.Errorf("<marker id=%q> requires positive width and height", m.ID)
	}
	if m.stemCutback < 0 {
		return fmt.Errorf("<marker id=%q> stem-cutback cannot be negative", m.ID)
	}
	children := m.Widgets()
	if len(children) != 1 {
		return fmt.Errorf("<marker id=%q> requires exactly one component", m.ID)
	}
	child := children[0]
	switch child.(type) {
	case *StdCircle, *StdEllipse, *StdPolygon, *StdStar, *StdImage, *StdSVG:
	default:
		return fmt.Errorf("<marker id=%q> does not support child %T", m.ID, child)
	}
	if container, ok := child.(Container); ok && len(container.Widgets()) != 0 {
		return fmt.Errorf("<marker id=%q> component cannot contain child widgets", m.ID)
	}
	if raw, ok := child.(interface{ RawAttrs() map[string]string }); ok {
		if err := validateMarkerChildAttrs(raw.RawAttrs()); err != nil {
			return fmt.Errorf("<marker id=%q>: %w", m.ID, err)
		}
	}
	return nil
}

func validateMarkerChildAttrs(attrs map[string]string) error {
	for _, name := range []string{
		"left", "right", "top", "bottom", "start", "end",
		"center-x", "center-y", "width", "height", "position",
		"r", "rx", "ry",
	} {
		if _, exists := attrs[name]; exists {
			return fmt.Errorf("marker component cannot set %q", name)
		}
	}
	return nil
}

func (m *StdMarker) resolved(strokeWidth float64) resolvedMarker {
	scale := 1.0
	if m.markerUnits == markerUnitsStrokeWidth {
		scale = strokeWidth
	}
	width := m.markerWidth * scale
	height := m.markerHeight * scale
	refX := m.refX.value * scale
	refY := m.refY.value * scale
	if m.refX.percent {
		refX = width * m.refX.value / 100
	}
	if m.refY.percent {
		refY = height * m.refY.value / 100
	}
	return resolvedMarker{
		definition: m,
		width:      width,
		height:     height,
		refX:       refX,
		refY:       refY,
		cutback:    m.stemCutback * scale,
	}
}

type resolvedMarker struct {
	definition *StdMarker
	width      float64
	height     float64
	refX       float64
	refY       float64
	cutback    float64
}

func fitResolvedMarkers(start, end *resolvedMarker, available float64) (*resolvedMarker, *resolvedMarker) {
	required := 0.0
	if start != nil {
		required += start.width
	}
	if end != nil {
		required += end.width
	}
	if required <= 0 || available >= required {
		return start, end
	}
	factor := max(available, 0) / required
	scale := func(source *resolvedMarker) *resolvedMarker {
		if source == nil {
			return nil
		}
		clone := *source
		clone.width *= factor
		clone.height *= factor
		clone.refX *= factor
		clone.refY *= factor
		clone.cutback *= factor
		return &clone
	}
	return scale(start), scale(end)
}

func (m *resolvedMarker) referenceInset() float64 {
	if m == nil {
		return 0
	}
	return max(m.width-m.refX, 0)
}

func (m *resolvedMarker) draw(w Writer, x, y, angle float64, color colors.Color) error {
	if m == nil || m.definition == nil || m.width <= 0 || m.height <= 0 {
		return nil
	}
	var drawErr error
	err := w.WithAccessibilityArtifact(func() {
		drawErr = w.Rotate(angle, x, y, func() {
			left := x - m.refX
			top := y - m.refY
			if m.definition.builtin == "arrow" {
				w.SetFillColor(color)
				drawErr = w.Path(func() {
					w.MoveTo(left, top)
					w.LineTo(left+m.width, top+m.height/2)
					w.LineTo(left, top+m.height)
					w.LineTo(left, top)
					if err := w.Fill(); err != nil {
						drawErr = err
					}
				})
				return
			}
			child, err := cloneWidgetShallow(m.definition.Widgets()[0])
			if err != nil {
				drawErr = err
				return
			}
			if wc, ok := child.(WantsContainer); ok {
				if err := wc.SetContainer(m.definition); err != nil {
					drawErr = err
					return
				}
			}
			child.SetPosition(Absolute)
			child.SetLeft(left)
			child.SetTop(top)
			child.ResolveWidth(m.width)
			child.ResolveHeight(m.height)
			switch shape := child.(type) {
			case *StdCircle:
				if shape.fill == nil && markerUsesContextFill(&shape.StdWidget) {
					shape.fill = &BrushStyle{kind: BrushKindSolid, color: color}
				}
			case *StdEllipse:
				if shape.fill == nil && markerUsesContextFill(&shape.StdWidget) {
					shape.fill = &BrushStyle{kind: BrushKindSolid, color: color}
				}
			case *StdPolygon:
				if shape.fill == nil && markerUsesContextFill(&shape.StdWidget) {
					shape.fill = &BrushStyle{kind: BrushKindSolid, color: color}
				}
			case *StdStar:
				if shape.fill == nil && markerUsesContextFill(&shape.StdWidget) {
					shape.fill = &BrushStyle{kind: BrushKindSolid, color: color}
				}
			}
			drawErr = Print(child, w)
		})
	})
	if err != nil {
		return err
	}
	return drawErr
}

func markerUsesContextFill(widget *StdWidget) bool {
	if widget == nil {
		return false
	}
	for name := range widget.RawAttrs() {
		if name == "fill" || strings.HasPrefix(name, "fill.") {
			return false
		}
	}
	return true
}

func builtinMarker(id string) (*StdMarker, bool) {
	if id != "arrow" {
		return nil, false
	}
	return &StdMarker{
		markerWidth:  3,
		markerHeight: 3,
		refX:         markerCoordinate{value: 50, percent: true},
		refY:         markerCoordinate{value: 50, percent: true},
		markerUnits:  markerUnitsStrokeWidth,
		builtin:      "arrow",
	}, true
}

func resolveMarker(scope HasScope, id string, strokeWidth float64) (*resolvedMarker, error) {
	id = strings.TrimSpace(id)
	if id == "" || id == "none" {
		return nil, nil
	}
	if scope == nil {
		return nil, fmt.Errorf("missing marker %q", id)
	}
	definition, ok := scope.MarkerFor(id)
	if !ok {
		return nil, fmt.Errorf("missing marker %q", id)
	}
	if strokeWidth < 0 || math.IsNaN(strokeWidth) || math.IsInf(strokeWidth, 0) {
		strokeWidth = 0
	}
	resolved := definition.resolved(strokeWidth)
	return &resolved, nil
}

func init() {
	registerTag(DefaultSpace, "marker", func() any { return &StdMarker{} })
}

var _ Container = (*StdMarker)(nil)
var _ HasAttrs = (*StdMarker)(nil)
var _ WantsContainer = (*StdMarker)(nil)
var _ WantsDoc = (*StdMarker)(nil)
var _ WantsScope = (*StdMarker)(nil)
