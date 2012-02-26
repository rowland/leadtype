package pdf

import "testing"

func TestUnits(t *testing.T) {
	expectF(t, 1, UnitConversions["pt"].ratio)
	expectF(t, 72, UnitConversions["in"].ratio)
	expectF(t, 28.35, UnitConversions["cm"].ratio)

	UnitConversions.Add("dp", 0.072)
	expectF(t, 0.072, UnitConversions["dp"].ratio)

	expectF(t, 100, UnitConversions["pt"].toPts(100))
	expectF(t, 7200, UnitConversions["in"].toPts(100))
	expectF(t, 2835, UnitConversions["cm"].toPts(100))
	expectFdelta(t, 7.2, UnitConversions["dp"].toPts(100), 0.0001)

	expectF(t, 100, UnitConversions["pt"].fromPts(100))
	expectF(t, 100, UnitConversions["in"].fromPts(7200))
	expectF(t, 100, UnitConversions["cm"].fromPts(2835))
	expectFdelta(t, 100, UnitConversions["dp"].fromPts(7.2), 0.0001)
}

func TestUnitsFromPts(t *testing.T) {
	expectF(t, 100, unitsFromPts("pt", 100))
	expectF(t, 100, unitsFromPts("in", 7200))
	expectF(t, 100, unitsFromPts("cm", 2835))
	expectFdelta(t, 100, unitsFromPts("dp", 7.2), 0.0001)

	defer func() {
		if p := recover(); p == nil {
			t.Error("Expecting panic from invalid units")
		}
	}()
	unitsFromPts("bogus", 100)
	t.Error("function above should panic")
}

func TestUnitsToPts(t *testing.T) {
	expectF(t, 100, unitsToPts("pt", 100))
	expectF(t, 7200, unitsToPts("in", 100))
	expectF(t, 2835, unitsToPts("cm", 100))
	expectFdelta(t, 7.2, unitsToPts("dp", 100), 0.0001)

	defer func() {
		if p := recover(); p == nil {
			t.Error("Expecting panic from invalid units")
		}
	}()
	unitsToPts("bogus", 100)
	t.Error("function above should panic")
}
