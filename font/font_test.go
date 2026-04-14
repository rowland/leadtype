// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package font

import (
	"testing"

	"github.com/rowland/leadtype/afm"
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
