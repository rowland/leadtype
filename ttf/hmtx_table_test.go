// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bytes"
	"testing"
)

func TestHmtxTableInitBulkDecodesMetricsAndBearings(t *testing.T) {
	data := []byte{
		0x01, 0xf4, 0xff, 0xf6, // advance 500, bearing -10
		0x02, 0xbc, 0x00, 0x14, // advance 700, bearing 20
		0xff, 0xe2, // bearing -30
		0x00, 0x28, // bearing 40
	}
	source := append([]byte{0xaa, 0xbb}, data...)
	var table hmtxTable
	if err := table.init(bytes.NewReader(source), &tableDirEntry{offset: 2, length: uint32(len(data))}, 4, 2); err != nil {
		t.Fatal(err)
	}

	want := []longHorMetric{
		{advanceWidth: 500, leftSideBearing: -10},
		{advanceWidth: 700, leftSideBearing: 20},
	}
	for i := range want {
		if table.hMetrics[i] != want[i] {
			t.Fatalf("metric %d = %#v, want %#v", i, table.hMetrics[i], want[i])
		}
	}
	if got := table.lookup(2); got.advanceWidth != 700 || got.leftSideBearing != -30 {
		t.Fatalf("lookup(2) = %#v, want advance 700 and bearing -30", got)
	}
	if got := table.lookup(3); got.advanceWidth != 700 || got.leftSideBearing != 40 {
		t.Fatalf("lookup(3) = %#v, want advance 700 and bearing 40", got)
	}
}

func TestHmtxTableInitRejectsMalformedLengthsAndCounts(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		tableLen   uint32
		numGlyphs  uint16
		numMetrics uint16
	}{
		{name: "metrics exceed glyphs", tableLen: 8, numGlyphs: 1, numMetrics: 2},
		{name: "declared table too short", data: make([]byte, 10), tableLen: 9, numGlyphs: 3, numMetrics: 2},
		{name: "source data truncated", data: make([]byte, 9), tableLen: 10, numGlyphs: 3, numMetrics: 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var table hmtxTable
			if err := table.init(bytes.NewReader(tc.data), &tableDirEntry{length: tc.tableLen}, tc.numGlyphs, tc.numMetrics); err == nil {
				t.Fatal("init succeeded, want malformed table error")
			}
		})
	}
}

func BenchmarkHmtxTableInitLarge(b *testing.B) {
	const numGlyphs = uint16(65535)
	const numMetrics = uint16(65532)
	data := make([]byte, int(numMetrics)*4+int(numGlyphs-numMetrics)*2)
	entry := &tableDirEntry{length: uint32(len(data))}
	b.ReportAllocs()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		var table hmtxTable
		if err := table.init(bytes.NewReader(data), entry, numGlyphs, numMetrics); err != nil {
			b.Fatal(err)
		}
	}
}

func TestHmtxTable_lookup_Arial(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Skipf("Error loading font: %s", err)
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
		t.Skipf("Error loading font: %s", err)
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
// 16.7 ns go1.1.1
// 18.5 ns go1.2.1
// 18.0 ns go1.4.2
//	9.3 ns go1.6.2 mbp
//	9.2 ns go1.7.3

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
// 11.6 ns go1.1.1
// 18.6 ns go1.2.1
// 18.0 ns go1.4.2
//	9.3 ns go1.6.2 mbp
//	8.9 ns go1.7.3

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
// 16.6 ns go1.1.1
// 18.1 ns go1.2.1
// 17.8 ns go1.4.2
//	9.2 ns go1.6.2 mbp
//	9.0 ns go1.7.3

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
