// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type DimensionMode int8

const (
	DimUnspecified DimensionMode = iota
	DimLiteral
	DimPct
	DimRel
	DimAuto
)

type Dimensions struct {
	sides                Sides
	margin               Sides
	padding              Sides
	corners              Corners
	width                float32
	height               float32
	widthValue           float32
	heightValue          float32
	widthMode            DimensionMode
	heightMode           DimensionMode
	widthValid           bool
	heightValid          bool
	widthAspectInferred  bool
	heightAspectInferred bool
	max                  maxDimensions
}

type maxDimensions struct {
	widthValue  float32
	heightValue float32
	widthMode   DimensionMode
	heightMode  DimensionMode
}

type dimensionState struct {
	resolved       float32
	value          float32
	mode           DimensionMode
	valid          bool
	aspectInferred bool
}

type dimensionsState struct {
	width  dimensionState
	height dimensionState
	max    maxDimensions
}

var (
	rePct = regexp.MustCompile(`^(\d+(\.\d+)?)%$`)
	reRel = regexp.MustCompile(`^[+-](\d+(\.\d+)?)([a-z]+)?$`)
)

func parseDimensionAttr(value string, units Units) (DimensionMode, float64) {
	value = strings.TrimSpace(value)
	if value == "auto" {
		return DimAuto, 0
	}
	if rePct.MatchString(value) {
		pct, _ := strconv.ParseFloat(value[:len(value)-1], 64)
		return DimPct, pct
	}
	if reRel.MatchString(value) {
		return DimRel, ParseMeasurement(value, units)
	}
	return DimLiteral, ParseMeasurement(value, units)
}

func (d *Dimensions) MarginTop() float64 {
	return d.margin[topSide].Float64()
}

func (d *Dimensions) MarginRight() float64 {
	return d.margin[rightSide].Float64()
}

func (d *Dimensions) MarginBottom() float64 {
	return d.margin[bottomSide].Float64()
}

func (d *Dimensions) MarginLeft() float64 {
	return d.margin[leftSide].Float64()
}

func (d *Dimensions) PaddingTop() float64 {
	return d.padding[topSide].Float64()
}

func (d *Dimensions) PaddingRight() float64 {
	return d.padding[rightSide].Float64()
}

func (d *Dimensions) PaddingBottom() float64 {
	return d.padding[bottomSide].Float64()
}

func (d *Dimensions) PaddingLeft() float64 {
	return d.padding[leftSide].Float64()
}

func (d *Dimensions) SetAttrs(attrs map[string]string, units Units) {
	d.sides.SetAttrs(attrs, units)

	if margin, ok := attrs["margin"]; ok {
		d.margin.SetAll(margin, units)
	}
	if MapHasKeyPrefix(attrs, "margin-") {
		d.margin.SetAttrs(filterMapAttrs("margin-", attrs), units)
	}

	if padding, ok := attrs["padding"]; ok {
		d.padding.SetAll(padding, units)
	}
	if MapHasKeyPrefix(attrs, "padding-") {
		d.padding.SetAttrs(filterMapAttrs("padding-", attrs), units)
	}

	if corners, ok := attrs["corners"]; ok {
		d.corners.SetAll(corners, units)
	}

	if width, ok := attrs["width"]; ok {
		switch mode, value := parseDimensionAttr(width, units); mode {
		case DimAuto:
			d.SetWidthAuto()
		case DimPct:
			d.SetWidthPct(value)
		case DimRel:
			d.SetWidthRel(value)
		default:
			d.SetWidth(value)
		}
	}
	if height, ok := attrs["height"]; ok {
		switch mode, value := parseDimensionAttr(height, units); mode {
		case DimAuto:
			d.SetHeightAuto()
		case DimPct:
			d.SetHeightPct(value)
		case DimRel:
			d.SetHeightRel(value)
		default:
			d.SetHeight(value)
		}
	}
	if width, ok := attrs["max-width"]; ok {
		d.setMaxWidthAttr(strings.TrimSpace(width), units)
	}
	if height, ok := attrs["max-height"]; ok {
		d.setMaxHeightAttr(strings.TrimSpace(height), units)
	}
}

func (d *Dimensions) setMaxWidthAttr(width string, units Units) {
	if width = strings.TrimSpace(width); width == "" {
		d.ClearMaxWidth()
		return
	}
	switch mode, value := parseDimensionAttr(width, units); mode {
	case DimAuto:
		d.ClearMaxWidth()
	case DimPct:
		d.SetMaxWidthPct(value)
	case DimRel:
		d.SetMaxWidthRel(value)
	default:
		d.SetMaxWidth(value)
	}
}

func (d *Dimensions) setMaxHeightAttr(height string, units Units) {
	if height = strings.TrimSpace(height); height == "" {
		d.ClearMaxHeight()
		return
	}
	switch mode, value := parseDimensionAttr(height, units); mode {
	case DimAuto:
		d.ClearMaxHeight()
	case DimPct:
		d.SetMaxHeightPct(value)
	case DimRel:
		d.SetMaxHeightRel(value)
	default:
		d.SetMaxHeight(value)
	}
}

func (d *Dimensions) SetMaxWidth(value float64) {
	d.max.widthValue = float32(value)
	d.max.widthMode = DimLiteral
}

func (d *Dimensions) SetMaxWidthPct(value float64) {
	d.max.widthValue = float32(value)
	d.max.widthMode = DimPct
}

func (d *Dimensions) SetMaxWidthRel(value float64) {
	d.max.widthValue = float32(value)
	d.max.widthMode = DimRel
}

func (d *Dimensions) ClearMaxWidth() {
	d.max.widthValue = 0
	d.max.widthMode = DimUnspecified
}

func (d *Dimensions) MaxWidthIsSet() bool {
	return maxDimensionIsSet(d.max.widthMode)
}

func (d *Dimensions) SetMaxHeight(value float64) {
	d.max.heightValue = float32(value)
	d.max.heightMode = DimLiteral
}

func (d *Dimensions) SetMaxHeightPct(value float64) {
	d.max.heightValue = float32(value)
	d.max.heightMode = DimPct
}

func (d *Dimensions) SetMaxHeightRel(value float64) {
	d.max.heightValue = float32(value)
	d.max.heightMode = DimRel
}

func (d *Dimensions) ClearMaxHeight() {
	d.max.heightValue = 0
	d.max.heightMode = DimUnspecified
}

func (d *Dimensions) MaxHeightIsSet() bool {
	return maxDimensionIsSet(d.max.heightMode)
}

func maxDimensionIsSet(mode DimensionMode) bool {
	switch mode {
	case DimLiteral, DimPct, DimRel:
		return true
	default:
		return false
	}
}

func (d *Dimensions) SetHeight(value float64) {
	d.height = float32(value)
	d.heightValue = float32(value)
	d.heightMode = DimLiteral
	d.heightValid = true
	d.heightAspectInferred = false
}

func (d *Dimensions) SetHeightAuto() {
	d.height = 0
	d.heightValue = 0
	d.heightMode = DimAuto
	d.heightValid = false
	d.heightAspectInferred = false
}

func (d *Dimensions) ClearHeight() {
	d.height = 0
	d.heightValue = 0
	d.heightMode = DimUnspecified
	d.heightValid = false
	d.heightAspectInferred = false
}

func (d *Dimensions) SetHeightPct(value float64) {
	d.height = 0
	d.heightValue = float32(value)
	d.heightMode = DimPct
	d.heightValid = false
	d.heightAspectInferred = false
}

func (d *Dimensions) SetHeightRel(value float64) {
	d.height = 0
	d.heightValue = float32(value)
	d.heightMode = DimRel
	d.heightValid = false
	d.heightAspectInferred = false
}

func (d *Dimensions) ResolveHeight(value float64) {
	d.height = float32(value)
	d.heightValid = true
	d.heightAspectInferred = false
}

func (d *Dimensions) ResolveAspectHeight(value float64) {
	d.height = float32(value)
	d.heightValid = true
	d.heightAspectInferred = true
}

func (d *Dimensions) ClearResolvedHeight() {
	if d.heightMode == DimLiteral {
		d.height = d.heightValue
	} else {
		d.height = 0
	}
	d.heightValid = false
	d.heightAspectInferred = false
}

func (d *Dimensions) HeightIsSet() bool {
	if d.heightValid {
		return true
	}
	switch d.heightMode {
	case DimLiteral, DimPct, DimRel:
		return true
	default:
		return false
	}
}

func (d *Dimensions) SetTop(value float64) {
	d.sides[topSide].Set(value)
}

func (d *Dimensions) SetRight(value float64) {
	d.sides[rightSide].Set(value)
}

func (d *Dimensions) SetBottom(value float64) {
	d.sides[bottomSide].Set(value)
}

func (d *Dimensions) SetLeft(value float64) {
	d.sides[leftSide].Set(value)
}

func (d *Dimensions) SetWidth(value float64) {
	d.width = float32(value)
	d.widthValue = float32(value)
	d.widthMode = DimLiteral
	d.widthValid = true
	d.widthAspectInferred = false
}

func (d *Dimensions) SetWidthAuto() {
	d.width = 0
	d.widthValue = 0
	d.widthMode = DimAuto
	d.widthValid = false
	d.widthAspectInferred = false
}

func (d *Dimensions) ClearWidth() {
	d.width = 0
	d.widthValue = 0
	d.widthMode = DimUnspecified
	d.widthValid = false
	d.widthAspectInferred = false
}

func (d *Dimensions) SetWidthPct(value float64) {
	d.width = 0
	d.widthValue = float32(value)
	d.widthMode = DimPct
	d.widthValid = false
	d.widthAspectInferred = false
}

func (d *Dimensions) SetWidthRel(value float64) {
	d.width = 0
	d.widthValue = float32(value)
	d.widthMode = DimRel
	d.widthValid = false
	d.widthAspectInferred = false
}

func (d *Dimensions) ResolveWidth(value float64) {
	d.width = float32(value)
	d.widthValid = true
	d.widthAspectInferred = false
}

func (d *Dimensions) ResolveAspectWidth(value float64) {
	d.width = float32(value)
	d.widthValid = true
	d.widthAspectInferred = true
}

func (d *Dimensions) ClearResolvedWidth() {
	if d.widthMode == DimLiteral {
		d.width = d.widthValue
	} else {
		d.width = 0
	}
	d.widthValid = false
	d.widthAspectInferred = false
}

func (d *Dimensions) String() string {
	return fmt.Sprintf("Dimensions width=%f height=%f margin=%s padding=%s corners=%s",
		d.width, d.height, &d.margin, &d.padding, &d.corners)
}

func (d *Dimensions) WidthPctIsSet() bool {
	return d.widthMode == DimPct
}

func (d *Dimensions) WidthRelIsSet() bool {
	return d.widthMode == DimRel
}

func (d *Dimensions) WidthIsSet() bool {
	if d.widthValid {
		return true
	}
	switch d.widthMode {
	case DimLiteral, DimPct, DimRel:
		return true
	default:
		return false
	}
}

func (d *Dimensions) WidthMode() DimensionMode {
	return d.widthMode
}

func (d *Dimensions) HeightMode() DimensionMode {
	return d.heightMode
}

func (d *Dimensions) WidthAspectInferred() bool {
	return d.widthAspectInferred
}

func (d *Dimensions) HeightAspectInferred() bool {
	return d.heightAspectInferred
}

func (d *Dimensions) SaveState() dimensionsState {
	state := dimensionsState{
		width: dimensionState{
			resolved:       d.width,
			value:          d.widthValue,
			mode:           d.widthMode,
			valid:          d.widthValid,
			aspectInferred: d.widthAspectInferred,
		},
		height: dimensionState{
			resolved:       d.height,
			value:          d.heightValue,
			mode:           d.heightMode,
			valid:          d.heightValid,
			aspectInferred: d.heightAspectInferred,
		},
		max: d.max,
	}
	return state
}

func (d *Dimensions) RestoreState(state dimensionsState) {
	d.width = state.width.resolved
	d.widthValue = state.width.value
	d.widthMode = state.width.mode
	d.widthValid = state.width.valid
	d.widthAspectInferred = state.width.aspectInferred
	d.height = state.height.resolved
	d.heightValue = state.height.value
	d.heightMode = state.height.mode
	d.heightValid = state.height.valid
	d.heightAspectInferred = state.height.aspectInferred
	d.max = state.max
}
