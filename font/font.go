// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package font

import "github.com/rowland/leadtype/options"

type Font struct {
	family       string
	Weight       string
	style        string
	subType      string
	Ranges       []string
	RuneSet      RuneSet
	RelativeSize float64
	metrics      FontMetrics
}

func New(family string, options options.Options, fontSources FontSources) (*Font, error) {
	font := &Font{
		family:       family,
		Weight:       options.StringDefault("weight", ""),
		style:        options.StringDefault("style", ""),
		RelativeSize: options.FloatDefault("relative_size", 100) / 100.0,
	}
	if Ranges, ok := options["ranges"]; ok {
		switch Ranges := Ranges.(type) {
		case []string:
			font.Ranges = Ranges
		case RuneSet:
			font.RuneSet = Ranges
		}
	}
	var err error
	for _, fontSource := range fontSources {
		if font.metrics, err = fontSource.Select(font.family, font.Weight, font.style, font.Ranges); err == nil {
			font.subType = fontSource.SubType()
			return font, nil
		}
	}
	return nil, err
}

func (font *Font) AdvanceWidth(codepoint rune) (width int, err bool) {
	return font.metrics.AdvanceWidth(codepoint)
}

func (font *Font) Ascent() int {
	return font.metrics.Ascent()
}

func (font *Font) BoundingBox() [4]int {
	return font.metrics.BoundingBox()
}

func (font *Font) CapHeight() int {
	return font.metrics.CapHeight()
}

func (font *Font) Copyright() string {
	return font.metrics.Copyright()
}

func (font *Font) Descent() int {
	return font.metrics.Descent()
}

func (font *Font) Family() string {
	return font.metrics.Family()
}

func (font *Font) Flags() (flags uint32) {
	return font.metrics.Flags()
}

func (font *Font) FullName() string {
	return font.metrics.FullName()
}

func (font *Font) HasRune(rune rune) bool {
	if font.RuneSet == nil {
		_, err := font.metrics.AdvanceWidth(rune)
		return !err
	}
	if font.RuneSet.HasRune(rune) {
		_, err := font.metrics.AdvanceWidth(rune)
		return !err
	}
	return false
}

func (font *Font) HasMetrics() bool {
	return font.metrics != nil
}

func (font *Font) Height() int {
	return font.metrics.Ascent() + -font.metrics.Descent()
}

func (font *Font) ItalicAngle() float64 {
	return font.metrics.ItalicAngle()
}

func (font *Font) Leading() int {
	return font.metrics.Leading()
}

func (font *Font) LineGap() int {
	return font.metrics.LineGap()
}

func (font *Font) Matches(other *Font) bool {
	return font.family == other.family &&
		font.Weight == other.Weight &&
		font.style == other.style &&
		font.subType == other.subType &&
		font.RuneSet == other.RuneSet &&
		font.RelativeSize == other.RelativeSize &&
		stringSlicesEqual(font.Ranges, other.Ranges)
}

func (font *Font) NumGlyphs() int {
	return font.metrics.NumGlyphs()
}

func (font *Font) PostScriptName() string {
	return font.metrics.PostScriptName()
}

func (font *Font) StemV() int {
	return font.metrics.StemV()
}

func (font *Font) Style() string {
	return font.style
}

func (font *Font) SubType() string {
	return font.subType
}

func (font *Font) UnderlinePosition() int {
	return font.metrics.UnderlinePosition()
}

func (font *Font) UnderlineThickness() int {
	return font.metrics.UnderlineThickness()
}

func (font *Font) UnitsPerEm() int {
	return font.metrics.UnitsPerEm()
}

func (font *Font) Version() string {
	return font.metrics.Version()
}

func (font *Font) XHeight() int {
	return font.metrics.XHeight()
}

func stringSlicesEqual(sl1, sl2 []string) bool {
	if len(sl1) != len(sl2) {
		return false
	}
	for i, s := range sl1 {
		if s != sl2[i] {
			return false
		}
	}
	return true
}
