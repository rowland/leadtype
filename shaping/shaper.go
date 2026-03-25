// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

// Package shaping provides Arabic (and complex-script) text shaping.
//
// Three implementations are available, selected at build time:
//
//   - Default (no tags): no-op stub; callers receive nil and must fall back
//     to unshaped glyph metrics from the font.
//   - -tags arabic: pure-Go shaping via github.com/go-text/typesetting.
//   - -tags harfbuzz: CGO shaping via the system libharfbuzz.
//
// The harfbuzz tag supersedes arabic; do not combine them.
package shaping

// GlyphPosition holds one shaped glyph's identifier and its positioning data.
//
// XAdvance, YAdvance, XOffset, and YOffset are expressed in 26.6 fixed-point
// scaled units (i.e. the raw int32 value divided by 64 gives the metric in the
// same unit as the ppem argument passed to Shape).
//
// ClusterIndex is the index of the first source rune (in the []rune slice
// passed to Shape) that was shaped into this glyph. For ligatures, multiple
// consecutive runes map to one glyph; ClusterIndex identifies the first.
type GlyphPosition struct {
	GlyphID      uint16
	XAdvance     int32
	YAdvance     int32
	XOffset      int32
	YOffset      int32
	ClusterIndex int
}

// Shaper shapes a run of Unicode text into positioned glyphs for a given font.
//
//   - text is the Arabic (or other complex-script) run to shape.
//   - fontBytes is the raw content of a TrueType or OpenType font file.
//   - ppem is the size of the font in the caller's coordinate system (e.g.
//     points for PDF output). It is used to scale the returned metrics.
//
// The returned slice is in visual order (already reordered for RTL text).
// A nil slice with a nil error indicates the no-op stub build; callers should
// fall back to standard font metrics in that case.
type Shaper interface {
	Shape(text []rune, fontBytes []byte, ppem float32) ([]GlyphPosition, error)
}
