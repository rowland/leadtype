// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

//go:build arabic || harfbuzz

package shaping_test

import (
	_ "embed"
	"os"
	"testing"
	"unicode/utf8"

	"github.com/rowland/leadtype/shaping"
)

//go:embed testdata/Amiri-Regular.ttf
var amiriFont []byte

// arabicText is "مرحبا" (hello in Arabic).
const arabicText = "مرحبا"

func TestShape_arabic(t *testing.T) {
	s := shaping.NewShaper()
	runes := []rune(arabicText)

	glyphs, err := s.Shape(runes, amiriFont, 12)
	if err != nil {
		t.Fatalf("Shape: %v", err)
	}
	if len(glyphs) == 0 {
		t.Fatal("expected shaped glyphs, got none")
	}

	runeCount := utf8.RuneCountInString(arabicText)
	t.Logf("input: %d runes → output: %d glyphs", runeCount, len(glyphs))

	for i, g := range glyphs {
		if g.GlyphID == 0 {
			t.Errorf("glyph[%d]: GlyphID is 0 (missing glyph)", i)
		}
		t.Logf("glyph[%d]: id=%d xAdv=%.2f xOff=%.2f yOff=%.2f",
			i, g.GlyphID,
			float64(g.XAdvance)/64,
			float64(g.XOffset)/64,
			float64(g.YOffset)/64,
		)
	}
}

func TestShape_corruptFont(t *testing.T) {
	s := shaping.NewShaper()
	_, err := s.Shape([]rune(arabicText), []byte("not a font"), 12)
	if err == nil {
		t.Fatal("expected error for corrupt font data, got nil")
	}
}

func TestShape_fontCacheHit(t *testing.T) {
	s := shaping.NewShaper()
	runes := []rune(arabicText)

	first, err := s.Shape(runes, amiriFont, 12)
	if err != nil {
		t.Fatalf("first Shape: %v", err)
	}
	second, err := s.Shape(runes, amiriFont, 12)
	if err != nil {
		t.Fatalf("second Shape: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("cache inconsistency: first=%d glyphs, second=%d glyphs", len(first), len(second))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("glyph[%d] differs between calls: %+v vs %+v", i, first[i], second[i])
		}
	}
}

// TestShape_envFont exercises the shaper against a user-supplied font.
// Set LEADTYPE_ARABIC_FONT to the path of any TTF/OTF Arabic font.
func TestShape_envFont(t *testing.T) {
	path := os.Getenv("LEADTYPE_ARABIC_FONT")
	if path == "" {
		t.Skip("LEADTYPE_ARABIC_FONT not set")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read font: %v", err)
	}
	s := shaping.NewShaper()
	glyphs, err := s.Shape([]rune(arabicText), data, 12)
	if err != nil {
		t.Fatalf("Shape: %v", err)
	}
	t.Logf("%d glyphs from %s", len(glyphs), path)
}
