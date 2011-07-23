package pdf

import "testing"

func TestUnits(t *testing.T) {
	expectF(t, 1, UnitConversions["pt"].factor)
	expectF(t, 72, UnitConversions["in"].factor)
	expectF(t, 28.35, UnitConversions["cm"].factor)

	UnitConversions.Add("dp", 0.072)
	expectF(t, 0.072, UnitConversions["dp"].factor)

	expectF(t, 100, UnitConversions["pt"].toPts(100))
	expectF(t, 7200, UnitConversions["in"].toPts(100))
	expectF(t, 2835, UnitConversions["cm"].toPts(100))
	expectF(t, 7.2, UnitConversions["dp"].toPts(100))
}
