// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package font

import (
	"github.com/rowland/leadtype/ttf"
	"testing"
)

func TestFont_HasRune(t *testing.T) {
	arial, err := ttf.LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatal(err)
	}
	f := &Font{metrics: arial}
	check(t, f.HasRune('A'), "Arial should have 'A'.")
	check(t, !f.HasRune(0x9999), "Arial should not have 0x9999.")
}

func TestFont_Matches(t *testing.T) {
	arial1, err := ttf.LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatal(err)
	}
	f1 := &Font{metrics: arial1}

	arial2, err := ttf.LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatal(err)
	}
	f2 := &Font{metrics: arial2}

	check(t, f1.Matches(f2), "Fonts should match.")
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
	for i := 0; i < b.N; i++ {
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
	if !condition {
		t.Error(msg)
	}
}
