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
	DimSpecified
	DimPct
	DimRel
	DimAuto
)

type Dimensions struct {
	sides       Sides
	margin      Sides
	padding     Sides
	corners     Corners
	width       float32
	height      float32
	widthValue  float32
	heightValue float32
	widthMode   DimensionMode
	heightMode  DimensionMode
}

type dimensionState struct {
	resolved float32
	value    float32
	mode     DimensionMode
}

type dimensionsState struct {
	width  dimensionState
	height dimensionState
}

var (
	rePct = regexp.MustCompile(`^(\d+(\.\d+)?)%$`)
	reRel = regexp.MustCompile(`^[+-](\d+(\.\d+)?)`)
)

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
		width = strings.TrimSpace(width)
		if width == "auto" {
			d.SetWidthAuto()
		} else if rePct.MatchString(width) {
			widthPct, _ := strconv.ParseFloat(width[:len(width)-1], 64)
			d.SetWidthPct(widthPct)
		} else if reRel.MatchString(width) {
			widthRel, _ := strconv.ParseFloat(width, 64)
			d.SetWidthRel(widthRel)
		} else {
			width := ParseMeasurement(width, units)
			d.SetWidth(width)
		}
	}
	if height, ok := attrs["height"]; ok {
		height = strings.TrimSpace(height)
		if height == "auto" {
			d.SetHeightAuto()
		} else if rePct.MatchString(height) {
			heightPct, _ := strconv.ParseFloat(height[:len(height)-1], 64)
			d.SetHeightPct(heightPct)
		} else if reRel.MatchString(height) {
			heightRel, _ := strconv.ParseFloat(height, 64)
			d.SetHeightRel(heightRel)
		} else {
			height := ParseMeasurement(height, units)
			d.SetHeight(height)
		}
	}
}

func (d *Dimensions) SetHeight(value float64) {
	d.height = float32(value)
	d.heightValue = float32(value)
	d.heightMode = DimSpecified
}

func (d *Dimensions) SetHeightAuto() {
	d.height = 0
	d.heightValue = 0
	d.heightMode = DimAuto
}

func (d *Dimensions) ClearHeight() {
	d.height = 0
	d.heightValue = 0
	d.heightMode = DimUnspecified
}

func (d *Dimensions) SetHeightPct(value float64) {
	d.height = 0
	d.heightValue = float32(value)
	d.heightMode = DimPct
}

func (d *Dimensions) SetHeightRel(value float64) {
	d.height = 0
	d.heightValue = float32(value)
	d.heightMode = DimRel
}

func (d *Dimensions) HeightIsSet() bool {
	switch d.heightMode {
	case DimSpecified, DimPct, DimRel:
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
	d.widthMode = DimSpecified
}

func (d *Dimensions) SetWidthAuto() {
	d.width = 0
	d.widthValue = 0
	d.widthMode = DimAuto
}

func (d *Dimensions) ClearWidth() {
	d.width = 0
	d.widthValue = 0
	d.widthMode = DimUnspecified
}

func (d *Dimensions) SetWidthPct(value float64) {
	d.width = 0
	d.widthValue = float32(value)
	d.widthMode = DimPct
}

func (d *Dimensions) SetWidthRel(value float64) {
	d.width = 0
	d.widthValue = float32(value)
	d.widthMode = DimRel
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
	switch d.widthMode {
	case DimSpecified, DimPct, DimRel:
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

func (d *Dimensions) SaveState() dimensionsState {
	return dimensionsState{
		width: dimensionState{
			resolved: d.width,
			value:    d.widthValue,
			mode:     d.widthMode,
		},
		height: dimensionState{
			resolved: d.height,
			value:    d.heightValue,
			mode:     d.heightMode,
		},
	}
}

func (d *Dimensions) RestoreState(state dimensionsState) {
	d.width = state.width.resolved
	d.widthValue = state.width.value
	d.widthMode = state.width.mode
	d.height = state.height.resolved
	d.heightValue = state.height.value
	d.heightMode = state.height.mode
}
