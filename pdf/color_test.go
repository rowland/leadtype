// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"testing"
)

func TestColor_RGB(t *testing.T) {
	c := Color(0x030507)
	r, g, b := c.RGB()
	expectI(t, 3, int(r))
	expectI(t, 5, int(g))
	expectI(t, 7, int(b))
}

func TestColor_RGB64(t *testing.T) {
	c := Color(0xFF3300)
	r, g, b := c.RGB64()
	expectF(t, 1.00, r)
	expectF(t, 0.20, g)
	expectF(t, 0.00, b)
}
