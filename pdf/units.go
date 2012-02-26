package pdf

import "fmt"

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

func (ucm UnitConversionMap) Add(name string, factor float64) {
	ucm[name] = &units{name, factor}
}

var UnitConversions = UnitConversionMap{
	"pt": &units{"pt", 1},
	"in": &units{"in", 72},
	"cm": &units{"cm", 28.35},
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
