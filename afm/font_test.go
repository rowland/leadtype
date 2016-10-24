// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package afm

import "testing"

func aw(width int, err bool) int {
	return width
}

func TestLoadFont(t *testing.T) {
	f, err := LoadFont("data/fonts/Helvetica.afm")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	if f == nil {
		t.Fatal("Font not loaded")
	}

	expectI(t, "UnitsPerEm", 1000, f.UnitsPerEm())
	expectI(t, "NumGlyphs", 315, f.NumGlyphs())
	expect(t, "Serif", !f.Serif())
	expectUI32(t, "Flags", 0, f.Flags())
	expectI(t, "registered", 737, aw(f.AdvanceWidth(0xAE)))
	expectI(t, "copyright", 737, aw(f.AdvanceWidth(0xA9)))
	// expectI(t, "epsilon", 913, f.AdvanceWidth(0x03B5))
	// expectI(t, "l-cedilla", 1139, f.AdvanceWidth(0x013B))
	// expectI(t, "afii57414", 1307, f.AdvanceWidth(0x0626))
	expectI(t, "trademark", 1000, aw(f.AdvanceWidth(0x2122)))
	// expectI(t, "reversed-e", 1366, f.AdvanceWidth(0x018E))
	// expectI(t, "t-with-comma", 1251, f.AdvanceWidth(0x021A))
	//
	// // Could not find MissingWidth in TTF.
	// expectI(t, "MaxWidth", 4096, f.MaxWidth())

}

// 3,956,700 ns
// 2,446,984 ns
// 2,284,165 ns/2.284 ms
// 2,294,880 ns/2.294 ms
// 2,161,622 ns/2.161 ms go1.1.1
// 1,880,786 ns/1.880 ms go1.2.1
// 2,095,878 ns/2.095 ms go1.4.2
//   651,438 ns          go1.6.2 mbp
//   539,770 ns          go1.7.3
func BenchmarkLoadFont(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadFont("data/fonts/Helvetica.afm")
	}
}

// 58.4 ns
// 71.3 ns
// 41.5 ns
// 67.6 ns weekly.2012-02-22
// 42.5 ns go1 (returning bool err)
// 36.7 ns go1.1.1
// 39.1 ns go1.2.1
// 39.7 ns go1.4.2
// 23.7 ns go1.6.2 mbp
// 22.8 ns go1.7.3
func BenchmarkAdvanceWidth(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("data/fonts/Helvetica.afm")
	if err != nil {
		panic("Error loading font")
	}
	min, max := f.CharMetrics[0].Code, f.CharMetrics[len(f.CharMetrics)-1].Code
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.AdvanceWidth(min + rune(i)%(max-min+1))
	}
}
