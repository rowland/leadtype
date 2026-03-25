// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"testing"
)

// TestMinimalTTF_Metrics verifies that the generated minimal.ttf fixture has
// the expected metric values and cmap coverage.
func TestMinimalTTF_Metrics(t *testing.T) {
	font, err := LoadFont("testdata/minimal.ttf")
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	expectS(t, "Family", "Minimal", font.Family())
	expectI(t, "UnitsPerEm", 1000, font.UnitsPerEm())
	expectI(t, "NumGlyphs", 188, font.NumGlyphs())
	expectI(t, "Ascent", 800, font.Ascent())
	expectI(t, "Descent", -200, font.Descent())
}

// TestMinimalTTF_CmapCoverage verifies that the fixture cmap maps the expected
// Unicode codepoints and returns the correct advance width for each.
func TestMinimalTTF_CmapCoverage(t *testing.T) {
	font, err := LoadFont("testdata/minimal.ttf")
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	cases := []struct {
		r    rune
		desc string
	}{
		{0x0020, "space (ASCII)"},
		{0x0041, "A (ASCII)"},
		{0x007E, "~ (ASCII last)"},
		{0x00C0, "À (Latin supplement first)"},
		{0x00FF, "ÿ (Latin supplement last)"},
		{0x0391, "Α (Greek capital first)"},
		{0x03A9, "Ω (Greek capital last)"},
		{0x4E2D, "中 (CJK)"},
		{0x6587, "文 (CJK)"},
	}
	for _, tc := range cases {
		w, notFound := font.AdvanceWidth(tc.r)
		if notFound {
			t.Errorf("codepoint %U (%s): not found in cmap", tc.r, tc.desc)
			continue
		}
		if w != 512 {
			t.Errorf("codepoint %U (%s): advance width = %d, want 512", tc.r, tc.desc, w)
		}
	}
}

// TestMinimalTTC_SubFonts verifies the TTC fixture contains two sub-fonts with
// the expected family/style/offset values.
func TestMinimalTTC_SubFonts(t *testing.T) {
	infos, err := LoadFontInfosFromTTC("testdata/minimal.ttc")
	if err != nil {
		t.Fatalf("LoadFontInfosFromTTC: %v", err)
	}
	if len(infos) != 2 {
		t.Fatalf("expected 2 sub-fonts, got %d", len(infos))
	}

	expectS(t, "infos[0].Family", "Minimal", infos[0].Family())
	expectS(t, "infos[0].Style", "Regular", infos[0].Style())
	expectS(t, "infos[1].Family", "Minimal", infos[1].Family())
	expectS(t, "infos[1].Style", "Bold", infos[1].Style())

	// Offsets must be nonzero and different.
	if infos[0].TTCOffset() == 0 {
		t.Error("infos[0].TTCOffset should be nonzero")
	}
	if infos[0].TTCOffset() == infos[1].TTCOffset() {
		t.Errorf("sub-fonts share the same TTC offset %d", infos[0].TTCOffset())
	}
}

// TestCJKSample_Coverage verifies that cjk-sample.ttf contains the expected
// Latin and CJK codepoints with nonzero advance widths.
func TestCJKSample_Coverage(t *testing.T) {
	font, err := LoadFont("testdata/cjk-sample.ttf")
	if err != nil {
		t.Fatalf("LoadFont cjk-sample.ttf: %v", err)
	}

	if font.NumGlyphs() < 100 {
		t.Errorf("expected at least 100 glyphs, got %d", font.NumGlyphs())
	}

	// Latin fallback
	for _, r := range []rune{' ', 'A', 'z', '~'} {
		w, notFound := font.AdvanceWidth(r)
		if notFound {
			t.Errorf("Latin codepoint %U not found in cmap", r)
		} else if w <= 0 {
			t.Errorf("Latin codepoint %U: advance width = %d, want > 0", r, w)
		}
	}

	// CJK codepoints
	for _, r := range []rune{0x4E00, 0x4E2D, 0x4E63} {
		w, notFound := font.AdvanceWidth(r)
		if notFound {
			t.Errorf("CJK codepoint %U not found in cmap", r)
		} else if w <= 0 {
			t.Errorf("CJK codepoint %U: advance width = %d, want > 0", r, w)
		}
	}
}

// TestMinimalTTC_LoadFontAtOffset verifies that each TTC sub-font can be
// loaded as a full Font via its recorded offset.
func TestMinimalTTC_LoadFontAtOffset(t *testing.T) {
	infos, err := LoadFontInfosFromTTC("testdata/minimal.ttc")
	if err != nil {
		t.Fatalf("LoadFontInfosFromTTC: %v", err)
	}

	for i, fi := range infos {
		font, err := LoadFontAtOffset(fi.Filename(), fi.TTCOffset())
		if err != nil {
			t.Errorf("sub-font %d: LoadFontAtOffset: %v", i, err)
			continue
		}
		expectI(t, "UnitsPerEm", 1000, font.UnitsPerEm())
		expectI(t, "NumGlyphs", 188, font.NumGlyphs())

		// Verify CJK codepoints are accessible.
		for _, r := range []rune{0x4E2D, 0x6587} {
			if _, notFound := font.AdvanceWidth(r); notFound {
				t.Errorf("sub-font %d: codepoint %U not found", i, r)
			}
		}
	}
}
