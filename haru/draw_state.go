// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package haru

import (
	"github.com/rowland/hpdf"
	"github.com/rowland/leadtype/colors"
)

type drawState struct {
	charSpacing     float64
	fillColor       colors.Color
	fontColor       colors.Color
	fontHandle      *hpdf.Font
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
