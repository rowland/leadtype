// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func format12TestData(length, language, nGroups uint32, groups ...format12Group) []byte {
	data := make([]byte, 14+12*len(groups))
	binary.BigEndian.PutUint16(data[0:], 0)
	binary.BigEndian.PutUint32(data[2:], length)
	binary.BigEndian.PutUint32(data[6:], language)
	binary.BigEndian.PutUint32(data[10:], nGroups)
	for i, group := range groups {
		pos := 14 + i*12
		binary.BigEndian.PutUint32(data[pos:], group.startCharCode)
		binary.BigEndian.PutUint32(data[pos+4:], group.endCharCode)
		binary.BigEndian.PutUint32(data[pos+8:], group.startGlyphCode)
	}
	return data
}

func TestFormat12EncodingRecordInitBulkDecodesGroups(t *testing.T) {
	groups := []format12Group{
		{startCharCode: 0x20, endCharCode: 0x7e, startGlyphCode: 1},
		{startCharCode: 0x10000, endCharCode: 0x10002, startGlyphCode: 500},
	}
	data := format12TestData(16+12*uint32(len(groups)), 7, uint32(len(groups)), groups...)
	var enc format12EncodingRecord
	if err := enc.init(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}

	if enc.language != 7 || len(enc.groups) != len(groups) {
		t.Fatalf("decoded language/groups = %d/%d, want 7/%d", enc.language, len(enc.groups), len(groups))
	}
	for i := range groups {
		if enc.groups[i] != groups[i] {
			t.Fatalf("group %d = %#v, want %#v", i, enc.groups[i], groups[i])
		}
	}
	tests := map[int]int{
		0x1f:    -1,
		0x20:    1,
		0x7e:    95,
		0x7f:    -1,
		0x10000: 500,
		0x10002: 502,
	}
	for codepoint, want := range tests {
		if got := enc.glyphIndex(codepoint); got != want {
			t.Errorf("glyphIndex(%#x) = %d, want %d", codepoint, got, want)
		}
	}
}

func TestFormat12EncodingRecordInitRejectsMalformedLengthsAndCounts(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "header too short", data: format12TestData(15, 0, 0)},
		{name: "groups exceed declared length", data: format12TestData(16, 0, 1)},
		{name: "declared length exceeds groups", data: format12TestData(28, 0, 0)},
		{
			name: "group data truncated",
			data: format12TestData(28, 0, 1, format12Group{startCharCode: 1, endCharCode: 2, startGlyphCode: 3})[:25],
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var enc format12EncodingRecord
			if err := enc.init(bytes.NewReader(tc.data)); err == nil {
				t.Fatal("init succeeded, want malformed subtable error")
			}
		})
	}
}

func TestFormat12EncodingRecordGlyphIndexEmpty(t *testing.T) {
	var enc format12EncodingRecord
	if err := enc.init(bytes.NewReader(format12TestData(16, 0, 0))); err != nil {
		t.Fatal(err)
	}
	if got := enc.glyphIndex('A'); got != -1 {
		t.Fatalf("glyphIndex(A) = %d, want -1", got)
	}
}

func TestCmapTableInitDoesNotReadFormat12IntoAdjacentTable(t *testing.T) {
	group := format12Group{startCharCode: 1, endCharCode: 2, startGlyphCode: 3}
	subtable := append([]byte{0, 12}, format12TestData(28, 0, 1, group)...)
	data := make([]byte, 12, 12+len(subtable))
	binary.BigEndian.PutUint16(data[2:], 1)
	binary.BigEndian.PutUint16(data[4:], MicrosoftPlatformID)
	binary.BigEndian.PutUint16(data[6:], UCS4PlatformSpecificID)
	binary.BigEndian.PutUint32(data[8:], 12)
	data = append(data, subtable...)

	// The final byte belongs to the next table. Without a table-bounded reader,
	// it would incorrectly complete the format-12 group.
	entry := &tableDirEntry{length: uint32(len(data) - 1)}
	var table cmapTable
	if err := table.init(bytes.NewReader(data), entry); err == nil {
		t.Fatal("init succeeded by reading format-12 data beyond the cmap table")
	}
}

// import "fmt"

func TestCmapTable_glyphIndex_Arial(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Skipf("Error loading font: %s", err)
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
		t.Skipf("Error loading font: %s", err)
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
// 44.1 ns go1.2.1
// 42.1 ns go1.4.2
// 28.3 ns go1.6.2 mbp
// 30.1 ns go1.7.3
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
// 40.8 ns go1.2.1
// 38.0 ns go1.4.2
// 26.5 ns go1.6.2 mbp
// 25.6 ns go1.7.3
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
// 60.1 ns go1.2.1
// 61.8 ns go1.4.2
// missing from mbp
// func BenchmarkGlyphIndex_format12(b *testing.B) {
// 	b.StopTimer()
// 	f, err := LoadFont("/Library/Fonts/华文楷体.ttf")
// 	if err != nil {
// 		panic("Error loading font: " + err.Error())
// 	}
// 	b.StartTimer()
// 	for i := 0; i < b.N; i++ {
// 		f.cmapTable.format12Indexer.glyphIndex(int(f.os2Table.fsFirstCharIndex) + i%int(f.os2Table.fsLastCharIndex-f.os2Table.fsFirstCharIndex+1))
// 	}
// }
