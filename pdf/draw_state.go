// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "github.com/rowland/leadtype/colors"

type drawState struct {
	charSpacing     float64
	fillColor       colors.Color
	fontColor       colors.Color
	fontKey         string
	fontSize        float64
	lineCapStyle    LineCapStyle
	lineColor       colors.Color
	lineDashPattern string
	lineSpacing     float64
	lineWidth       float64
	loc             Location
	strikeout       bool
	underline       bool
	wordSpacing     float64
}
