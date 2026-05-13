// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "fmt"

// ptsPerMM is PDF points per millimeter (72 pt/in ÷ 25.4 mm/in).
const ptsPerMM = 72.0 / 25.4

type units struct {
	name  string
	ratio float64
}

func (units *units) fromPts(measurement float64) float64 {
	return measurement / units.ratio
}

func (units *units) toPts(measurement float64) float64 {
	return units.ratio * measurement
}

type UnitConversionMap map[string]*units

// Add registers a custom unit with the given points-per-unit ratio (multiply
// by this value to convert from the named unit to points). Built-in names are
// pt, in, cm, mm, and dp; prefer those for normal measurements—Add remains for
// rare specialized conversions.
func (ucm UnitConversionMap) Add(name string, factor float64) {
	ucm[name] = &units{name, factor}
}

var UnitConversions = UnitConversionMap{
	"pt": &units{"pt", 1},
	"in": &units{"in", 72},
	"cm": &units{"cm", 10 * ptsPerMM},
	"mm": &units{"mm", ptsPerMM},
	// Thousandths of an inch (1000 dp = 1 in).
	"dp": &units{"dp", 72.0 / 1000},
}

func unitsFromPts(units string, measurement float64) float64 {
	u := UnitConversions[units]
	if u == nil {
		panic(fmt.Sprintf("Invalid units %s", units))
	}
	return u.fromPts(measurement)
}

func unitsToPts(units string, measurement float64) float64 {
	u := UnitConversions[units]
	if u == nil {
		panic(fmt.Sprintf("Invalid units %s", units))
	}
	return u.toPts(measurement)
}
