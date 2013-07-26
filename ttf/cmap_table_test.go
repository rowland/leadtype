// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import "testing"

// import "fmt"

func TestCmapTable_glyphIndex_Arial(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	expectI(t, "registered", 138, f.cmapTable.glyphIndex(0xAE))
	expectI(t, "copyright", 139, f.cmapTable.glyphIndex(0xA9))
	expectI(t, "epsilon", 304, f.cmapTable.glyphIndex(0x03B5))
	expectI(t, "l-cedilla", 441, f.cmapTable.glyphIndex(0x013B))
	expectI(t, "afii57414", 905, f.cmapTable.glyphIndex(0x0626))
	expectI(t, "trademark", 140, f.cmapTable.glyphIndex(0x2122))

	expectI(t, "reversed-e", 1688, f.cmapTable.glyphIndex(0x018E))
	expectI(t, "t-with-comma", 1801, f.cmapTable.glyphIndex(0x021A))
	// for i := uint16(0); i < 65535; i++ {
	// 	fmt.Printf("[%d] %d\n", i, f.cmapTable.glyphIndex(i))
	// }
}

func TestCmapTable_glyphIndex_Courier(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Courier New.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	expectI(t, "registered", 138, f.cmapTable.glyphIndex(0xAE))
	expectI(t, "copyright", 139, f.cmapTable.glyphIndex(0xA9))
	expectI(t, "epsilon", 304, f.cmapTable.glyphIndex(0x03B5))
	expectI(t, "l-cedilla", 441, f.cmapTable.glyphIndex(0x013B))
	expectI(t, "afii57414", 905, f.cmapTable.glyphIndex(0x0626))
	expectI(t, "trademark", 140, f.cmapTable.glyphIndex(0x2122))

	expectI(t, "reversed-e", 1693, f.cmapTable.glyphIndex(0x018E))
	expectI(t, "t-with-comma", 1806, f.cmapTable.glyphIndex(0x021A))
	// for i := uint16(0); i < 256; i++ {
	// 	fmt.Printf("[%d] %d\n", i, f.cmapTable.glyphIndex(i))
	// }
}

// 41.9 ns
// 42.1 ns go1.1.1
func BenchmarkGlyphIndex(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		panic("Error loading font: " + err.Error())
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.cmapTable.glyphIndex(int(f.os2Table.fsFirstCharIndex) + i%int(f.os2Table.fsLastCharIndex-f.os2Table.fsFirstCharIndex+1))
	}
}

// 35.9 ns
// 39.1 ns go1.1.1
func BenchmarkGlyphIndex_format4(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		panic("Error loading font: " + err.Error())
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.cmapTable.format4Indexer.glyphIndex(int(f.os2Table.fsFirstCharIndex) + i%int(f.os2Table.fsLastCharIndex-f.os2Table.fsFirstCharIndex+1))
	}
}

// 51.7 ns
// 56.7 ns go1.1.1
func BenchmarkGlyphIndex_format12(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("/Library/Fonts/华文楷体.ttf")
	if err != nil {
		panic("Error loading font: " + err.Error())
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.cmapTable.format12Indexer.glyphIndex(int(f.os2Table.fsFirstCharIndex) + i%int(f.os2Table.fsLastCharIndex-f.os2Table.fsFirstCharIndex+1))
	}
}
