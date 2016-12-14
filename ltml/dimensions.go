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
	height    float64
	heightPct float64
	heightRel float64
}

var (
	rePct = regexp.MustCompile(`^(\d+(\.\d+)?)%$`)
	reRel = regexp.MustCompile(`^[+-](\d+(\.\d+)?)`)
)

func (d *Dimensions) MarginTop() float64 {
	return d.margin[topSide]
}

func (d *Dimensions) MarginRight() float64 {
	return d.margin[rightSide]
}

func (d *Dimensions) MarginBottom() float64 {
	return d.margin[bottomSide]
}

func (d *Dimensions) MarginLeft() float64 {
	return d.margin[leftSide]
}

func (d *Dimensions) PaddingTop() float64 {
	return d.padding[topSide]
}

func (d *Dimensions) PaddingRight() float64 {
	return d.padding[rightSide]
}

func (d *Dimensions) PaddingBottom() float64 {
	return d.padding[bottomSide]
}

func (d *Dimensions) PaddingLeft() float64 {
	return d.padding[leftSide]
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
			d.heightPct, _ = strconv.ParseFloat(height, 64)
		} else if reRel.MatchString(height) {
			d.heightRel, _ = strconv.ParseFloat(height, 64)
		} else {
			d.height = ParseMeasurement(height, units)
		}
	}
}

func (d *Dimensions) SetHeight(value float64) {
	d.height = value
	d.heightPct = 0
}

func (d *Dimensions) SetTop(value float64) {
	d.sides[topSide] = value
}

func (d *Dimensions) SetRight(value float64) {
	d.sides[rightSide] = value
}

func (d *Dimensions) SetBottom(value float64) {
	d.sides[bottomSide] = value
}

func (d *Dimensions) SetLeft(value float64) {
	d.sides[leftSide] = value
}

func (d *Dimensions) SetWidth(value float64) {
	d.width = value
	d.widthPct = 0
}

func (d *Dimensions) String() string {
	return fmt.Sprintf("Dimensions width=%f height=%f margin=%s padding=%s corners=%s",
		d.width, d.height, &d.margin, &d.padding, &d.corners)
}
