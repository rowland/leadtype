// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package font

type FontMetrics interface {
	AdvanceWidth(codepoint rune) (width int, err bool)
	// AdvanceWidthForGlyph returns the advance width for the given glyph ID in
	// glyph-space units. Returns 0 for font types that do not support glyph IDs.
	AdvanceWidthForGlyph(glyphID uint16) int
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
	// GlyphIndex returns the glyph ID for the given Unicode codepoint, or 0 if
	// the codepoint is not present. Returns 0 for font types that do not use
	// glyph IDs (e.g. AFM/Type1).
	GlyphIndex(r rune) uint16
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
