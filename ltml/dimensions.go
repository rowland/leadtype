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
	width     float64
	widthPct  float64
	widthRel  float64
	widthSet  bool
	height    float64
	heightPct float64
	heightRel float64
	heightSet bool
}

var (
	rePct = regexp.MustCompile(`^(\d+(\.\d+)?)%$`)
	reRel = regexp.MustCompile(`^[+-](\d+(\.\d+)?)`)
)

func (d *Dimensions) MarginTop() float64 {
	return d.margin[topSide].Value
}

func (d *Dimensions) MarginRight() float64 {
	return d.margin[rightSide].Value
}

func (d *Dimensions) MarginBottom() float64 {
	return d.margin[bottomSide].Value
}

func (d *Dimensions) MarginLeft() float64 {
	return d.margin[leftSide].Value
}

func (d *Dimensions) PaddingTop() float64 {
	return d.padding[topSide].Value
}

func (d *Dimensions) PaddingRight() float64 {
	return d.padding[rightSide].Value
}

func (d *Dimensions) PaddingBottom() float64 {
	return d.padding[bottomSide].Value
}

func (d *Dimensions) PaddingLeft() float64 {
	return d.padding[leftSide].Value
}

func (d *Dimensions) SetAttrs(attrs map[string]string, units Units) {
	d.sides.SetAttrs("", attrs, units)

	if margin, ok := attrs["margin"]; ok {
		d.margin.SetAll(margin, units)
	}
	d.margin.SetAttrs("margin-", attrs, units)

	if padding, ok := attrs["padding"]; ok {
		d.padding.SetAll(padding, units)
	}
	d.padding.SetAttrs("padding-", attrs, units)

	if corners, ok := attrs["corners"]; ok {
		d.corners.SetAll(corners, units)
	}

	if width, ok := attrs["width"]; ok {
		if rePct.MatchString(width) {
			d.widthPct, _ = strconv.ParseFloat(width, 64)
		} else if reRel.MatchString(width) {
			d.widthRel, _ = strconv.ParseFloat(width, 64)
		} else {
			d.width = ParseMeasurement(width, units)
		}
	}
	if height, ok := attrs["height"]; ok {
		if rePct.MatchString(height) {
			heightPct, _ := strconv.ParseFloat(height, 64)
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
	d.height, d.heightPct, d.heightRel, d.heightSet = value, 0, 0, true
}

func (d *Dimensions) SetHeightPct(value float64) {
	d.heightPct, d.height, d.heightRel, d.heightSet = value, 0, 0, true
}

func (d *Dimensions) SetHeightRel(value float64) {
	d.heightRel, d.height, d.heightPct, d.heightSet = value, 0, 0, true
}

func (d *Dimensions) HeightSet() bool {
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
	d.width, d.widthPct, d.widthRel, d.widthSet = value, 0, 0, true
}

func (d *Dimensions) SetWidthPct(value float64) {
	d.widthPct, d.widthRel, d.widthPct, d.widthSet = value, 0, 0, true
}

func (d *Dimensions) SetWidthRel(value float64) {
	d.widthRel, d.widthPct, d.width, d.widthSet = value, 0, 0, true
}

func (d *Dimensions) String() string {
	return fmt.Sprintf("Dimensions width=%f height=%f margin=%s padding=%s corners=%s",
		d.width, d.height, &d.margin, &d.padding, &d.corners)
}

func (d *Dimensions) WidthSet() bool {
	return d.widthSet
}
