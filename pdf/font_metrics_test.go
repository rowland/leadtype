// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"leadtype/afm"
	"leadtype/ttf"
	"testing"
)

func TestAfmSatisfiesFontMetrics(t *testing.T) {
	var _ FontMetrics = new(afm.Font)
}

func TestTtfSatisfiesFontMetrics(t *testing.T) {
	var _ FontMetrics = new(ttf.Font)
}
