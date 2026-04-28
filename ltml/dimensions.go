// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"regexp"
	"strconv"
)

type Dimensions struct {
	sides     Sides
	margin    Sides
	padding   Sides
	corners   Corners
	width     float32
	widthPct  float32
	widthRel  float32
	height    float32
	heightPct float32
	heightRel float32
	widthSet  bool
	heightSet bool
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
		if rePct.MatchString(width) {
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
		if rePct.MatchString(height) {
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
	d.height, d.heightPct, d.heightRel, d.heightSet = float32(value), 0, 0, true
}

func (d *Dimensions) SetHeightPct(value float64) {
	d.heightPct, d.height, d.heightRel, d.heightSet = float32(value), 0, 0, true
}

func (d *Dimensions) SetHeightRel(value float64) {
	d.heightRel, d.height, d.heightPct, d.heightSet = float32(value), 0, 0, true
}

func (d *Dimensions) HeightIsSet() bool {
	return d.heightSet
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
	d.width, d.widthPct, d.widthRel, d.widthSet = float32(value), 0, 0, true
}

func (d *Dimensions) SetWidthPct(value float64) {
	d.widthPct, d.widthRel, d.width, d.widthSet = float32(value), 0, 0, true
}

func (d *Dimensions) SetWidthRel(value float64) {
	d.widthRel, d.widthPct, d.width, d.widthSet = float32(value), 0, 0, true
}

func (d *Dimensions) String() string {
	return fmt.Sprintf("Dimensions width=%f height=%f margin=%s padding=%s corners=%s",
		d.width, d.height, &d.margin, &d.padding, &d.corners)
}

func (d *Dimensions) WidthPctIsSet() bool {
	return d.widthPct > 0
}

func (d *Dimensions) WidthRelIsSet() bool {
	return d.widthRel != 0
}

func (d *Dimensions) WidthIsSet() bool {
	return d.widthSet
}
