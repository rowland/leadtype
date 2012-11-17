// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"testing"
)

func TestLinePatterns(t *testing.T) {
	expectNS(t, "solid", "[] 0", LinePatterns["solid"].String())
	expectNS(t, "dotted", "[1 2] 0", LinePatterns["dotted"].String())
	expectNS(t, "dashed", "[4 2] 0", LinePatterns["dashed"].String())
	check(t, LinePatterns["bogus"] == nil, "No pattern should be returned for bogus")
}
