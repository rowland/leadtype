package pdf

type units struct {
	name  string
	ratio float32
}

func (units *units) toPts(measurement float32) float32 {
	return units.ratio * measurement
}

type UnitConversionMap map[string]*units

func (ucm UnitConversionMap) Add(name string, factor float32) {
	ucm[name] = &units{name, factor}
}

var UnitConversions = UnitConversionMap{
	"pt": &units{"pt", 1},
	"in": &units{"in", 72},
	"cm": &units{"cm", 28.35},
}
