// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

type Font struct {
	family       string
	weight       string
	style        string
	subType      string
	ranges       []string
	runeSet      RuneSet
	relativeSize float64
	metrics      FontMetrics
}

func NewFont(family string, options Options, fontSources FontSources) (*Font, error) {
	font := &Font{
		family:       family,
		weight:       options.StringDefault("weight", ""),
		style:        options.StringDefault("style", ""),
		relativeSize: options.FloatDefault("relative_size", 100) / 100.0,
	}
	if ranges, ok := options["ranges"]; ok {
		switch ranges := ranges.(type) {
		case []string:
			font.ranges = ranges
		case RuneSet:
			font.runeSet = ranges
		}
	}
	var err error
	for _, fontSource := range fontSources {
		if font.metrics, err = fontSource.Select(font.family, font.weight, font.style, font.ranges); err == nil {
			font.subType = fontSource.SubType()
			return font, nil
		}
	}
	return nil, err
}

func (font *Font) Ascent() int {
	return font.metrics.Ascent()
}

func (font *Font) Descent() int {
	return font.metrics.Descent()
}

func (font *Font) HasRune(rune rune) bool {
	if font.runeSet == nil {
		_, err := font.metrics.AdvanceWidth(rune)
		return !err
	}
	if font.runeSet.HasRune(rune) {
		_, err := font.metrics.AdvanceWidth(rune)
		return !err
	}
	return false
}

func (font *Font) Height() int {
	return font.metrics.Ascent() + -font.metrics.Descent()
}

func (font *Font) Matches(other *Font) bool {
	return font.family == other.family &&
		font.weight == other.weight &&
		font.style == other.style &&
		font.subType == other.subType &&
		font.runeSet == other.runeSet &&
		font.relativeSize == other.relativeSize &&
		stringSlicesEqual(font.ranges, other.ranges)
}
