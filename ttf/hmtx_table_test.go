// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import "testing"

func TestHmtxTable_lookup_Arial(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	// 1st glyph
	hmtx0 := f.hmtxTable.lookup(0)
	expectUI16(t, "0 advanceWidth", 1536, hmtx0.advanceWidth)
	expectI16(t, "0 leftSideBearing", 256, hmtx0.leftSideBearing)

	expectUI16(t, "0 lookupAdvanceWidth", 1536, f.hmtxTable.lookupAdvanceWidth(0))
	expectI16(t, "0 lookupLeftSideBearing", 256, f.hmtxTable.lookupLeftSideBearing(0))
	// 1st glyph with advanceWidth of 0
	hmtx1 := f.hmtxTable.lookup(1)
	expectUI16(t, "1 advanceWidth", 0, hmtx1.advanceWidth)
	expectI16(t, "1 leftSideBearing", 0, hmtx1.leftSideBearing)

	expectUI16(t, "1 lookupAdvanceWidth", 0, f.hmtxTable.lookupAdvanceWidth(1))
	expectI16(t, "1 lookupLeftSideBearing", 0, f.hmtxTable.lookupLeftSideBearing(1))
	// last glyph
	hmtx3380 := f.hmtxTable.lookup(3380)
	expectUI16(t, "3380 advanceWidth", 455, hmtx3380.advanceWidth)
	expectI16(t, "3380 leftSideBearing", 136, hmtx3380.leftSideBearing)

	expectUI16(t, "3380 lookupAdvanceWidth", 455, f.hmtxTable.lookupAdvanceWidth(3380))
	expectI16(t, "3380 lookupLeftSideBearing", 136, f.hmtxTable.lookupLeftSideBearing(3380))
	// index beyond last glyph
	hmtx3381 := f.hmtxTable.lookup(3381)
	expectUI16(t, "3381 advanceWidth", 0, hmtx3381.advanceWidth)
	expectI16(t, "3381 leftSideBearing", 0, hmtx3381.leftSideBearing)

	expectUI16(t, "3381 lookupAdvanceWidth", 455, f.hmtxTable.lookupAdvanceWidth(3381)) // value from last entry
	expectI16(t, "3381 lookupLeftSideBearing", 0, f.hmtxTable.lookupLeftSideBearing(3381))
}

func TestHmtxTable_lookup_Courier(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Courier New.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	// 1st glyph
	hmtx0 := f.hmtxTable.lookup(0)
	expectUI16(t, "0 advanceWidth", 1229, hmtx0.advanceWidth)
	expectI16(t, "0 leftSideBearing", 103, hmtx0.leftSideBearing)
	// 1st glyph with advanceWidth of 0
	hmtx1 := f.hmtxTable.lookup(1)
	expectUI16(t, "1 advanceWidth", 0, hmtx1.advanceWidth)
	expectI16(t, "1 leftSideBearing", 0, hmtx1.leftSideBearing)
	// last glyph with advanceWidth defined
	hmtx2 := f.hmtxTable.lookup(2)
	expectUI16(t, "2 advanceWidth", 1229, hmtx2.advanceWidth)
	expectI16(t, "2 leftSideBearing", 0, hmtx2.leftSideBearing)
	// index beyond last glyph with advanceWidth defined
	hmtx3 := f.hmtxTable.lookup(3)
	expectUI16(t, "3 advanceWidth", 1229, hmtx3.advanceWidth)
	expectI16(t, "3 leftSideBearing", 0, hmtx3.leftSideBearing)
	// last glyph with leftSideBearing defined
	hmtx3150 := f.hmtxTable.lookup(3150)
	expectUI16(t, "3150 advanceWidth", 1229, hmtx3150.advanceWidth)
	expectI16(t, "3150 leftSideBearing", 189, hmtx3150.leftSideBearing)
	// index byond last glyph
	hmtx3151 := f.hmtxTable.lookup(3151)
	expectUI16(t, "3151 advanceWidth", 0, hmtx3151.advanceWidth)
	expectI16(t, "3151 leftSideBearing", 0, hmtx3151.leftSideBearing)
}

// 9.4 ns
func BenchmarkLookup_LongHorMetric(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		panic("Error loading font")
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.hmtxTable.lookup(i % int(f.maxpTable.numGlyphs))
	}
}

// 7.7 ns
// 5.5 ns
func BenchmarkLookupAdvanceWidth(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		panic("Error loading font")
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.hmtxTable.lookupAdvanceWidth(i % int(f.maxpTable.numGlyphs))
	}
}

// 7.4 ns
// 8.8 ns
func BenchmarkLookupLeftSideBearing(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		panic("Error loading font")
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.hmtxTable.lookupLeftSideBearing(i % int(f.maxpTable.numGlyphs))
	}
}
