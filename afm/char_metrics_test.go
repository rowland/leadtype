// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package afm

import "testing"
import "sort"

func TestCharMetrics(t *testing.T) {
	cms := NewCharMetrics(3)
	cms[0].Code = 15
	cms[0].Name = "fifteen"
	cms[0].Width = 150
	cms[1].Code = 10
	cms[1].Name = "ten"
	cms[1].Width = 100
	cms[2].Code = 5
	cms[2].Name = "five"
	cms[2].Width = 50

	sort.Sort(cms)
	expectI(t, "len", 3, len(cms))
	expectI(t, "five", 5, int(cms[0].Code))
	expectI(t, "ten", 10, int(cms[1].Code))
	expectI(t, "fifteen", 15, int(cms[2].Code))

	expectI32(t, "five", 50, cms.ForRune(5).Width)
	expectI32(t, "ten", 100, cms.ForRune(10).Width)
	expectI32(t, "fifteen", 150, cms.ForRune(15).Width)
}
