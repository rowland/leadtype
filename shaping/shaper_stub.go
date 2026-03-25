// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

//go:build !arabic && !harfbuzz

package shaping

// NewShaper returns the no-op shaper for builds without Arabic shaping support.
// Shape returns nil, nil; callers should treat this as "shaping unavailable"
// and fall back to unshaped font metrics.
func NewShaper() Shaper { return noopShaper{} }

type noopShaper struct{}

func (noopShaper) Shape(_ []rune, _ FontReader, _ float32) ([]GlyphPosition, error) {
	return nil, nil
}
