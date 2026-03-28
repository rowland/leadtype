// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import "github.com/rowland/leadtype/colors"

type VerticalTextAlign uint8

const (
	VTextAlignBase VerticalTextAlign = iota
	VTextAlignAbove
	VTextAlignTop
	VTextAlignMiddle
	VTextAlignBelow
)

func parseVerticalTextAlign(value string) VerticalTextAlign {
	switch value {
	case "above":
		return VTextAlignAbove
	case "top":
		return VTextAlignTop
	case "middle":
		return VTextAlignMiddle
	case "below":
		return VTextAlignBelow
	default:
		return VTextAlignBase
	}
}

func (v VerticalTextAlign) String() string {
	switch v {
	case VTextAlignAbove:
		return "above"
	case VTextAlignTop:
		return "top"
	case VTextAlignMiddle:
		return "middle"
	case VTextAlignBelow:
		return "below"
	default:
		return "base"
	}
}

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
	vTextAlign      VerticalTextAlign
	wordSpacing     float64
}
