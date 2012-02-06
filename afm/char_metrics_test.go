package afm

import "testing"
import "sort"

func TestCharMetrics(t *testing.T) {
	cms := NewCharMetrics(3)
	cms[0].code = 15
	cms[0].name = "fifteen"
	cms[0].width = 150
	cms[1].code = 10
	cms[1].name = "ten"
	cms[1].width = 100
	cms[2].code = 5
	cms[2].name = "five"
	cms[2].width = 50

	sort.Sort(cms)
	expectI(t, "len", 3, len(cms))
	expectI(t, "five", 5, int(cms[0].code))
	expectI(t, "ten", 10, int(cms[1].code))
	expectI(t, "fifteen", 15, int(cms[2].code))

	expectI32(t, "five", 50, cms.ForRune(5).width)
	expectI32(t, "ten", 100, cms.ForRune(10).width)
	expectI32(t, "fifteen", 150, cms.ForRune(15).width)
}
