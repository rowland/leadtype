// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

type Font struct {
	family  string
	size    float64
	weight  string
	style   string
	color   Color
	subType string
	metrics FontMetrics
}
