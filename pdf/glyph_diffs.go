// Copyright 2015 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"leadtype/afm"
	"leadtype/codepage"
)

func glyphDiffs(cp1, cp2 codepage.CodepageIndex, first, last int) (diffs array) {
	map1, map2 := cp1.Map(), cp2.Map()
	same := true
	for i := first; i <= last; i++ {
		if same {
			if map1[i] != map2[i] {
				same = false
				r := afm.RuneGlyphs[map2[i]]
				if r == "" {
					r = "question"
				}
				diffs = append(diffs, integer(i), name(r))
			}
		} else {
			if map1[i] == map2[i] {
				same = true
			} else {
				r := afm.RuneGlyphs[map2[i]]
				if r == "" {
					r = "question"
				}
				diffs = append(diffs, name(r))
			}
		}
	}
	return
}
