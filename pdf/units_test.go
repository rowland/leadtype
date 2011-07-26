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
}
