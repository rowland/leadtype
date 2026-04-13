// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package font

import (
	"testing"

	"github.com/rowland/leadtype/afm"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/shaping"
	"github.com/rowland/leadtype/ttf"
)

func TestFont_HasRune(t *testing.T) {
	arial, err := ttf.LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Skip(err)
	}
	f := &Font{metrics: arial}
	check(t, f.HasRune('A'), "Arial should have 'A'.")
	check(t, !f.HasRune(0x9999), "Arial should not have 0x9999.")
}

func TestFont_Matches(t *testing.T) {
	arial1, err := ttf.LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Skip(err)
	}
	f1 := &Font{metrics: arial1}

	arial2, err := ttf.LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Skip(err)
	}
	f2 := &Font{metrics: arial2}

	check(t, f1.Matches(f2), "Fonts should match.")
}

func TestFont_Matches_DistinguishesDecorationCapOverrides(t *testing.T) {
	metrics, err := afm.LoadFont("data/fonts/Helvetica.afm")
	if err != nil {
		t.Skip(err)
	}
	f1 := &Font{metrics: metrics}
	f2 := &Font{metrics: metrics, underlineCapStyle: "round_cap", underlineCapStyleSet: true}
	f3 := &Font{metrics: metrics, underlineCapStyle: "round_cap", underlineCapStyleSet: true}

	check(t, !f1.Matches(f2), "Fonts with different underline cap overrides should not match.")
	check(t, f2.Matches(f3), "Fonts with matching underline cap overrides should match.")
	check(t, f2.UnderlineCapStyle() == "round_cap", "UnderlineCapStyle should return explicit override.")
}

func TestFont_Matches_DistinguishesDecorationColorOverrides(t *testing.T) {
	metrics, err := afm.LoadFont("data/fonts/Helvetica.afm")
	if err != nil {
		t.Skip(err)
	}
	f1 := &Font{metrics: metrics}
	f2 := &Font{metrics: metrics, underlineColor: colors.Green, underlineColorSet: true}
	f3 := &Font{metrics: metrics, underlineColor: colors.Green, underlineColorSet: true}

	check(t, !f1.Matches(f2), "Fonts with different underline color overrides should not match.")
	check(t, f2.Matches(f3), "Fonts with matching underline color overrides should match.")
	color, ok := f2.UnderlineColor()
	check(t, ok, "UnderlineColor should report an explicit override.")
	check(t, color == colors.Green, "UnderlineColor should return explicit override.")
}

func TestDecorationColorOption_IgnoresInvalidStrings(t *testing.T) {
	_, ok := decorationColorOption(options.Options{"underline_color": "not-a-color"}, "underline_color")
	check(t, !ok, "Invalid decoration colors should be ignored.")

	color, ok := decorationColorOption(options.Options{"underline_color": "black"}, "underline_color")
	check(t, ok, "Valid named colors should be accepted.")
	check(t, color == colors.Black, "Valid named colors should round-trip.")
}

func TestDecorationLineCapStyleOption_IgnoresInvalidValues(t *testing.T) {
	_, ok := decorationLineCapStyleOption(options.Options{"underline_cap": "not-a-cap"}, "underline_cap")
	check(t, !ok, "Invalid decoration cap styles should be ignored.")

	_, ok = decorationLineCapStyleOption(options.Options{"underline_cap": 1}, "underline_cap")
	check(t, !ok, "Non-string decoration cap styles should be ignored.")
}

func TestDecorationLineCapStyleOption_AcceptsValidValues(t *testing.T) {
	style, ok := decorationLineCapStyleOption(options.Options{"underline_cap": "round_cap"}, "underline_cap")
	check(t, ok, "Valid decoration cap styles should be accepted.")
	check(t, style == "round_cap", "Valid decoration cap styles should round-trip.")
}

func TestDecorationInt16Option_IgnoresInvalidStrings(t *testing.T) {
	_, ok := decorationInt16Option(options.Options{"underline_position": "not-a-number"}, "underline_position")
	check(t, !ok, "Invalid decoration numbers should be ignored.")

	_, ok = decorationInt16Option(options.Options{"underline_position": 40000}, "underline_position")
	check(t, !ok, "Out-of-range decoration numbers should be ignored.")
}

func TestDecorationInt16Option_AcceptsValidValues(t *testing.T) {
	value, ok := decorationInt16Option(options.Options{"underline_position": "-123"}, "underline_position")
	check(t, ok, "Valid string numbers should be accepted.")
	check(t, value == -123, "Valid string numbers should round-trip.")

	value, ok = decorationInt16Option(options.Options{"underline_thickness": 50.9}, "underline_thickness")
	check(t, ok, "Valid float values should be accepted.")
	check(t, value == 50, "Float decoration numbers should match existing truncation semantics.")
}

func TestFont_SupportsArabic_TTFArabic(t *testing.T) {
	amiri, err := ttf.LoadFont("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil {
		t.Skip(err)
	}
	f := &Font{metrics: amiri, Shaper: shaping.NewShaper()}
	check(t, f.SupportsArabic(), "Amiri should report Arabic support.")
}

func TestFont_SupportsArabic_TTFLatinOnly(t *testing.T) {
	minimal, err := ttf.LoadFont("../ttf/testdata/minimal.ttf")
	if err != nil {
		t.Skip(err)
	}
	f := &Font{metrics: minimal}
	check(t, !f.SupportsArabic(), "minimal.ttf should not report Arabic support.")
}

func TestFont_SupportsArabic_AFMFalse(t *testing.T) {
	metrics, err := afm.LoadFont("data/fonts/Helvetica.afm")
	if err != nil {
		t.Skip(err)
	}
	f := &Font{metrics: metrics}
	check(t, !f.SupportsArabic(), "AFM font should not report Arabic support.")
}

// 55.2 ns
// 46.1 ns go1.1.1
// 46.0 ns go1.1.2
// 46.1 ns go1.2.1
// 49.2 ns go1.4.2
func BenchmarkFont_HasRune(b *testing.B) {
	b.StopTimer()
	arial, err := ttf.LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		b.Fatal(err)
	}
	f := &Font{metrics: arial}
	b.StartTimer()
	for range b.N {
		f.HasRune('A')
	}
}

func TestStringSlicesEqual(t *testing.T) {
	a := []string{"abc", "def"}
	b := []string{"abc", "def"}
	c := []string{"abc", "def", "ghi"}
	d := []string{"abc", "ghi"}
	check(t, stringSlicesEqual(a, b), "Slices a and b should be equal.")
	check(t, !stringSlicesEqual(a, c), "Slices a and c should not be equal.")
	check(t, !stringSlicesEqual(a, d), "Slices a and d should not be equal.")
}

func check(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Error(msg)
	}
}
