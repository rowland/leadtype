// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

type Align int8

const (
	AlignDefault = Align(iota)
	AlignLeft
	AlignRight
	AlignTop
	AlignBottom
)

var alignStrings = []string{"default", "left", "right", "top", "bottom"}

func (a Align) String() string {
	if int(a) < len(alignStrings) {
		return alignStrings[a]
	}
	return "unknown"
}

type SelfAlign int8

const (
	SelfAlignDefault = SelfAlign(iota)
	SelfAlignStart
	SelfAlignCenter
	SelfAlignEnd
)

var selfAlignStrings = []string{"default", "start", "center", "end"}

func (sa SelfAlign) String() string {
	if int(sa) < len(selfAlignStrings) {
		return selfAlignStrings[sa]
	}
	return "unknown"
}

type OriginX int8

const (
	OriginXUnspecified OriginX = iota
	OriginXStart
	OriginXCenter
	OriginXEnd
	OriginXCustom
)

var originXStrings = []string{"unspecified", "start", "center", "end", "custom"}

func (o OriginX) String() string {
	if int(o) < len(originXStrings) {
		return originXStrings[o]
	}
	return "unknown"
}

type OriginY int8

const (
	OriginYUnspecified OriginY = iota
	OriginYTop
	OriginYMiddle
	OriginYBottom
	OriginYCustom
)

var originYStrings = []string{"unspecified", "top", "middle", "bottom", "custom"}

func (o OriginY) String() string {
	if int(o) < len(originYStrings) {
		return originYStrings[o]
	}
	return "unknown"
}

type HAlign int8

const (
	HAlignLeft = HAlign(iota)
	HAlignCenter
	HAlignRight
	HAlignJustify
	textAlignStart
	textAlignEnd
)

var hAlignStrings = []string{"left", "center", "right", "justify", "start", "end"}

func (ha HAlign) String() string {
	if int(ha) < len(hAlignStrings) {
		return hAlignStrings[ha]
	}
	return "unknown"
}

// parseTextAlign retains logical start/end values until the text is rendered.
// Invalid values intentionally fall back to physical left, matching the
// historical paragraph and label parsers.
func parseTextAlign(value string, allowJustify bool) HAlign {
	switch value {
	case "start":
		return textAlignStart
	case "end":
		return textAlignEnd
	case "center":
		return HAlignCenter
	case "right":
		return HAlignRight
	case "justify":
		if allowJustify {
			return HAlignJustify
		}
	}
	return HAlignLeft
}

// resolveTextAlign converts logical alignment to the physical alignment used
// by the PDF writer. Physical left/right values are never mirrored.
func resolveTextAlign(align HAlign, c Container) HAlign {
	rtl := c != nil && IsRTL(c)
	switch align {
	case textAlignStart:
		if rtl {
			return HAlignRight
		}
		return HAlignLeft
	case textAlignEnd:
		if rtl {
			return HAlignLeft
		}
		return HAlignRight
	default:
		return align
	}
}

type VAlign int8

const (
	VAlignTop = VAlign(iota)
	VAlignMiddle
	VAlignBottom
	VAlignBaseline
)

var vAlignStrings = []string{"top", "middle", "bottom", "baseline"}

func (va VAlign) String() string {
	if int(va) < len(vAlignStrings) {
		return vAlignStrings[va]
	}
	return "unknown"
}
