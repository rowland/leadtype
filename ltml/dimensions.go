// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"regexp"
	"strconv"
)

type Dimensions struct {
	Units
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

func (d *Dimensions) SetAttrs(attrs map[string]string) {
	d.Units.SetAttrs(attrs)

	if margin, ok := attrs["margin"]; ok {
		d.margin.SetAll(margin, d.units)
	}
	d.margin.SetAttrs("margin-", attrs, d.units)

	if padding, ok := attrs["padding"]; ok {
		d.padding.SetAll(padding, d.units)
	}
	d.padding.SetAttrs("padding-", attrs, d.units)

	if corners, ok := attrs["corners"]; ok {
		d.corners.SetAll(corners, d.units)
	}

	if width, ok := attrs["width"]; ok {
		if rePct.MatchString(width) {
			d.widthPct, _ = strconv.ParseFloat(width, 64)
		} else if reRel.MatchString(width) {
			d.widthRel, _ = strconv.ParseFloat(width, 64)
		} else {
			d.width = ParseMeasurement(width, d.units)
		}
	}
	if height, ok := attrs["height"]; ok {
		if rePct.MatchString(height) {
			d.heightPct, _ = strconv.ParseFloat(height, 64)
		} else if reRel.MatchString(height) {
			d.heightRel, _ = strconv.ParseFloat(height, 64)
		} else {
			d.height = ParseMeasurement(height, d.units)
		}
	}
}

func (d *Dimensions) String() string {
	return fmt.Sprintf("Dimensions units=%s width=%f height=%f margin=%s padding=%s corners=%s",
		d.units, d.width, d.height, &d.margin, &d.padding, &d.corners)
}
