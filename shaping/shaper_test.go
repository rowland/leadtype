// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package shaping_test

import (
	"testing"

	"github.com/rowland/leadtype/shaping"
)

func TestNewShaper_notNil(t *testing.T) {
	if shaping.NewShaper() == nil {
		t.Fatal("NewShaper returned nil")
	}
}

// emptyFontReader is a FontReader that returns no data, used to test error paths.
type emptyFontReader struct{}

func (emptyFontReader) FontKey() string { return "empty" }
func (emptyFontReader) Bytes() []byte   { return nil }

// TestShape_emptyFont verifies behaviour on empty font data across all builds.
// The stub returns nil, nil; active shapers should return an error.
func TestShape_emptyFont(t *testing.T) {
	s := shaping.NewShaper()
	glyphs, err := s.Shape([]rune("مرحبا"), emptyFontReader{}, 12)
	if err == nil && glyphs != nil {
		t.Errorf("expected nil glyphs for empty font data, got %d glyphs", len(glyphs))
	}
}
