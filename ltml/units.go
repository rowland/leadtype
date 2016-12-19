// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"regexp"
	"strconv"
)

type Units string

func (u *Units) SetAttrs(attrs map[string]string) {
	if units, ok := attrs["units"]; ok {
		*u = Units(units)
	}
}

// UnitConversions map custom units to points.
var UnitConversions = map[Units]float64{
	"pt": 1,
	"in": 72,
	"cm": 28.35,
}

func FromUnits(measurement float64, units Units) float64 {
	if points, ok := UnitConversions[units]; ok {
		return measurement * points
	}
	return measurement
}

var reMeasurement = regexp.MustCompile(`([+-]?\d+(\.\d+)?)([a-z]+)`)

// ParseMeasurement parses units out of a measurement, if present, and multiplies by unit conversion.
func ParseMeasurement(measurement string, units Units) float64 {
	if matches := reMeasurement.FindStringSubmatch(measurement); len(matches) >= 4 {
		if v, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return FromUnits(v, Units(matches[3]))
		}
		return 0
	}
	if v, err := strconv.ParseFloat(measurement, 64); err == nil {
		return FromUnits(v, units)
	}
	return 0
}

var _ HasAttrs = (*Units)(nil)
