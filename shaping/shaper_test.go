// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package shaping_test

import (
	"os"
	"testing"
	"unicode/utf8"

	"github.com/rowland/leadtype/shaping"
)

// arabicText is "مرحبا" (hello in Arabic).
const arabicText = "مرحبا"

func TestNewShaper_stub(t *testing.T) {
	s := shaping.NewShaper()
	if s == nil {
		t.Fatal("NewShaper returned nil")
	}
}

// TestShape_stub verifies the no-op shaper returns nil without error.
// This test runs for all build configurations.
func TestShape_noFont(t *testing.T) {
	s := shaping.NewShaper()
	runes := []rune(arabicText)
	glyphs, err := s.Shape(runes, []byte{}, 12)
	// Stub returns nil, nil. Active shapers should error on empty font data.
	if err == nil && glyphs != nil {
		t.Errorf("expected nil glyphs for empty font data, got %d glyphs", len(glyphs))
	}
}

// TestShape_withFont exercises the shaper when a real Arabic font is available.
// Set LEADTYPE_ARABIC_FONT to the path of a TTF/OTF Arabic font to run this test.
func TestShape_withFont(t *testing.T) {
	fontPath := os.Getenv("LEADTYPE_ARABIC_FONT")
	if fontPath == "" {
		t.Skip("LEADTYPE_ARABIC_FONT not set; skipping live shaping test")
	}
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		t.Fatalf("read font: %v", err)
	}

	s := shaping.NewShaper()
	runes := []rune(arabicText)
	glyphs, err := s.Shape(runes, fontBytes, 12)
	if err != nil {
		t.Fatalf("Shape: %v", err)
	}

	runeCount := utf8.RuneCountInString(arabicText)
	t.Logf("input: %d runes, output: %d glyphs", runeCount, len(glyphs))

	if len(glyphs) == 0 {
		// Stub build: nil is acceptable.
		return
	}

	for i, g := range glyphs {
		if g.GlyphID == 0 {
			t.Errorf("glyph[%d]: GlyphID is 0 (missing glyph)", i)
		}
		if g.XAdvance == 0 && g.YAdvance == 0 {
			t.Logf("glyph[%d]: zero advance (may be a combining mark)", i)
		}
		t.Logf("glyph[%d]: id=%d xAdv=%d yAdv=%d xOff=%d yOff=%d",
			i, g.GlyphID, g.XAdvance>>6, g.YAdvance>>6, g.XOffset>>6, g.YOffset>>6)
	}
}
