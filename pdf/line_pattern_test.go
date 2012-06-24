// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"testing"
)

func TestLinePatterns(t *testing.T) {
	expectS(t, "[] 0", LinePatterns["solid"].String())
	expectS(t, "[1 2] 0", LinePatterns["dotted"].String())
	expectS(t, "[4 2] 0", LinePatterns["dashed"].String())
	check(t, LinePatterns["bogus"] == nil, "No pattern should be returned for bogus")
}
