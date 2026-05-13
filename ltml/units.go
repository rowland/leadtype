// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"regexp"
	"strconv"
)

// ptsPerMM is PDF points per millimeter (72 pt/in ÷ 25.4 mm/in).
const ptsPerMM = 72.0 / 25.4

type Units string

func (u *Units) SetAttrs(attrs map[string]string) {
	if units, ok := attrs["units"]; ok {
		*u = Units(units)
	}
}

// UnitConversions maps unit names (suffixes and units= defaults) to points
// per unit (multiply the measurement by this value to get points).
var UnitConversions = map[Units]float64{
	"pt": 1,
	"in": 72,
	"cm": 10 * ptsPerMM,
	"mm": ptsPerMM,
	// Thousandths of an inch (1000 dp = 1 in).
	"dp": 72.0 / 1000,
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

// ParseOptionalMeasurement parses units out of a measurement, if present, and multiplies by unit conversion.
// If the measurement is empty, returns nil.
func ParseOptionalMeasurement(measurement string, units Units) *float64 {
	if measurement == "" {
		return nil
	}
	if matches := reMeasurement.FindStringSubmatch(measurement); len(matches) >= 4 {
		if v, err := strconv.ParseFloat(matches[1], 64); err == nil {
			value := FromUnits(v, Units(matches[3]))
			return &value
		}
		return nil
	}
	if v, err := strconv.ParseFloat(measurement, 64); err == nil {
		value := FromUnits(v, units)
		return &value
	}
	return nil
}

var _ HasAttrs = (*Units)(nil)
