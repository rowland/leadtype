// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import "testing"

func aw(width int, err bool) int {
	return width
}

func TestLoadFont(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	if f == nil {
		t.Fatal("Font not loaded")
	}

	expectI(t, "UnitsPerEm", 2048, f.UnitsPerEm())
	expectI(t, "NumGlyphs", 3381, f.NumGlyphs())
	expectI(t, "BoundingBox[0]", -1361, f.BoundingBox()[0])
	expectI(t, "BoundingBox[1]", -665, f.BoundingBox()[1])
	expectI(t, "BoundingBox[2]", 4096, f.BoundingBox()[2])
	expectI(t, "BoundingBox[3]", 2060, f.BoundingBox()[3])

	expectI(t, "registered", 1509, aw(f.AdvanceWidth(0xAE)))
	expectI(t, "copyright", 1509, aw(f.AdvanceWidth(0xA9)))
	expectI(t, "epsilon", 913, aw(f.AdvanceWidth(0x03B5)))
	expectI(t, "l-cedilla", 1139, aw(f.AdvanceWidth(0x013B)))
	expectI(t, "afii57414", 1307, aw(f.AdvanceWidth(0x0626)))
	expectI(t, "trademark", 2048, aw(f.AdvanceWidth(0x2122)))
	expectI(t, "reversed-e", 1366, aw(f.AdvanceWidth(0x018E)))
	expectI(t, "t-with-comma", 1251, aw(f.AdvanceWidth(0x021A)))

	expectI(t, "Ascent", 1854, f.Ascent())
	expectI(t, "Descent", -434, f.Descent())
	expectF(t, "ItalicAngle", 0, f.ItalicAngle())
	// Could not find MissingWidth in TTF.
	expectI(t, "Leading", 1854 - -434 + 67, f.Leading())
	// TODO: Verify this works in practice. Some examples indicate equivalence to lineGap instead.
	expectI(t, "MaxWidth", 4096, f.MaxWidth())
	expectI(t, "UnderlinePosition", -217, f.UnderlinePosition())
	expectI(t, "UnderlineThickness", 150, f.UnderlineThickness())
}

// 9,151,820 ns
// 5,327,058 ns
// 3,346,759 ns go1.1.1
func BenchmarkLoadFont(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadFont("/Library/Fonts/Arial.ttf")
	}
}

// 50.4 ns
// 74.6 ns after adding error result
// 47.8 ns go1 (returning bool err)
// 47.8 ns go1.1.1
func BenchmarkAdvanceWidth(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		panic("Error loading font")
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.AdvanceWidth(rune(int(f.os2Table.fsFirstCharIndex) + i%int(f.os2Table.fsLastCharIndex-f.os2Table.fsFirstCharIndex+1)))
	}
}
