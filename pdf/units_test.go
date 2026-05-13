// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "testing"

func TestUnits(t *testing.T) {
	expectF(t, 1, UnitConversions["pt"].ratio)
	expectF(t, 72, UnitConversions["in"].ratio)
	expectFdelta(t, 10*ptsPerMM, UnitConversions["cm"].ratio, 1e-12)
	expectFdelta(t, ptsPerMM, UnitConversions["mm"].ratio, 1e-12)
	expectF(t, 0.072, UnitConversions["dp"].ratio)

	const cm100pts = 100 * 10 * ptsPerMM

	expectF(t, 100, UnitConversions["pt"].toPts(100))
	expectF(t, 7200, UnitConversions["in"].toPts(100))
	expectFdelta(t, cm100pts, UnitConversions["cm"].toPts(100), 1e-9)
	expectFdelta(t, 7.2, UnitConversions["dp"].toPts(100), 0.0001)
	expectFdelta(t, 100*ptsPerMM, UnitConversions["mm"].toPts(100), 1e-9)

	expectF(t, 100, UnitConversions["pt"].fromPts(100))
	expectF(t, 100, UnitConversions["in"].fromPts(7200))
	expectF(t, 100, UnitConversions["cm"].fromPts(cm100pts))
	expectFdelta(t, 100, UnitConversions["dp"].fromPts(7.2), 0.0001)
	expectFdelta(t, 100, UnitConversions["mm"].fromPts(100*ptsPerMM), 1e-9)
}

func TestUnitConversionsAddExtension(t *testing.T) {
	const name = "_leadtype_test_custom_unit"
	t.Cleanup(func() { delete(UnitConversions, name) })
	UnitConversions.Add(name, 2.5)
	expectF(t, 2.5, UnitConversions[name].ratio)
	expectF(t, 250, unitsToPts(name, 100))
	expectF(t, 100, unitsFromPts(name, 250))
}

func TestUnitsFromPts(t *testing.T) {
	cm100pts := 100 * 10 * ptsPerMM
	expectF(t, 100, unitsFromPts("pt", 100))
	expectF(t, 100, unitsFromPts("in", 7200))
	expectF(t, 100, unitsFromPts("cm", cm100pts))
	expectFdelta(t, 100, unitsFromPts("dp", 7.2), 0.0001)
	expectFdelta(t, 100, unitsFromPts("mm", 100*ptsPerMM), 1e-9)

	defer func() {
		if p := recover(); p == nil {
			t.Error("Expecting panic from invalid units")
		}
	}()
	unitsFromPts("bogus", 100)
	t.Error("function above should panic")
}

func TestUnitsToPts(t *testing.T) {
	cm100pts := 100 * 10 * ptsPerMM
	expectF(t, 100, unitsToPts("pt", 100))
	expectF(t, 7200, unitsToPts("in", 100))
	expectFdelta(t, cm100pts, unitsToPts("cm", 100), 1e-9)
	expectFdelta(t, 7.2, unitsToPts("dp", 100), 0.0001)
	expectFdelta(t, 100*ptsPerMM, unitsToPts("mm", 100), 1e-9)

	defer func() {
		if p := recover(); p == nil {
			t.Error("Expecting panic from invalid units")
		}
	}()
	unitsToPts("bogus", 100)
	t.Error("function above should panic")
}
