// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

type FontMetrics interface {
	AdvanceWidth(codepoint rune) (width int, err bool)
	Ascent() int
	// AvgWidth() int
	// BoundingBox() BoundingBox xxx
	CapHeight() int
	Copyright() string
	Descent() int
	// Designer() string
	// Embeddable() bool
	Family() string
	// Flags() (flags uint32)
	FullName() string
	ItalicAngle() float64
	Leading() int
	// License() string
	// Manufacturer() string
	// MaxWidth() int
	NumGlyphs() int
	PostScriptName() string
	// Serif() bool TODO
	StemV() int
	Style() string
	// Trademark() string
	// UniqueName() string
	UnitsPerEm() int
	Version() string
	XHeight() int
}
