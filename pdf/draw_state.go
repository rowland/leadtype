// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

type drawState struct {
	fontColor       Color
	fontSize        float64
	lineCapStyle    LineCapStyle
	lineColor       Color
	lineDashPattern string
	lineWidth       float64
	loc             location
}
