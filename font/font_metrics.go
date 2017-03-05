// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package font

type FontMetrics interface {
	AdvanceWidth(codepoint rune) (width int, err bool)
	Ascent() int
	// AvgWidth() int
	BoundingBox() [4]int
	CapHeight() int
	Copyright() string
	Descent() int
	// Designer() string
	// Embeddable() bool
	Family() string
	Filename() string
	Flags() (flags uint32)
	FullName() string
	ItalicAngle() float64
	Leading() int
	// License() string
	LineGap() int
	// Manufacturer() string
	// MaxWidth() int
	NumGlyphs() int
	PostScriptName() string
	// Serif() bool TODO
	StemV() int
	StrikeoutPosition() int
	StrikeoutThickness() int
	Style() string
	// Trademark() string
	UnderlinePosition() int
	UnderlineThickness() int
	// UniqueName() string
	UnitsPerEm() int
	Version() string
	XHeight() int
}
