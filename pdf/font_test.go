// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"leadtype/ttf"
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
