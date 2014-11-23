// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

type drawState struct {
	fillColor       Color
	fontColor       Color
	fontKey         string
	fontSize        float64
	lineCapStyle    LineCapStyle
	lineColor       Color
	lineDashPattern string
	lineSpacing     float64
	lineThrough     bool
	lineWidth       float64
	loc             location
	underline       bool
}
