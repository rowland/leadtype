// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"strconv"
)

type Units struct {
	units string
}

func (u *Units) SetAttrs(attrs map[string]string) {
	if units, ok := attrs["units"]; ok {
		u.units = units
	}
}

// UnitConversions map custom units to points.
var UnitConversions = map[string]float64{
	"pt": 1,
	"in": 72,
	"cm": 28.35,
}

func FromUnits(measurement float64, units string) float64 {
	if points, ok := UnitConversions[units]; ok {
		return measurement * points
	}
	return measurement
}

func ParseMeasurement(measurement string, units string) float64 {
	// TODO: Parse units out of value, if present, overriding units parameter.
	// /([+-]?\d+(\.\d+)?)([a-z]+)/
	if v, err := strconv.ParseFloat(measurement, 64); err == nil {
		return FromUnits(v, units)
	}
	return 0
}
