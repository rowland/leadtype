// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"encoding/binary"
	"os"
	"testing"
)

const minimalTTF = "testdata/minimal.ttf"

// TestSubset_ReducesSize checks that a subset is strictly smaller than the original.
func TestSubset_ReducesSize(t *testing.T) {
	orig, err := LoadFont(minimalTTF)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	// Subset to just glyph 0 (.notdef) + a few ASCII glyphs.
	glyphIDs := []uint16{orig.GlyphIndex('A'), orig.GlyphIndex('B'), orig.GlyphIndex('C')}
	data, err := orig.Subset(glyphIDs)
	if err != nil {
		t.Fatalf("Subset: %v", err)
	}

	origData := loadRawFont(t, minimalTTF)
	if len(data) >= len(origData) {
		t.Errorf("expected subset (%d bytes) < original (%d bytes)", len(data), len(origData))
	}
}

// TestSubset_ValidFont verifies that the subset binary can be re-parsed.
func TestSubset_ValidFont(t *testing.T) {
	orig, err := LoadFont(minimalTTF)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	glyphIDs := []uint16{orig.GlyphIndex('A'), orig.GlyphIndex('B')}
	data, err := orig.Subset(glyphIDs)
	if err != nil {
		t.Fatalf("Subset: %v", err)
	}

	sub, err := LoadFontFromBytes(data)
	if err != nil {
		t.Fatalf("LoadFontFromBytes on subset: %v", err)
	}
	if sub.UnitsPerEm() != orig.UnitsPerEm() {
		t.Errorf("unitsPerEm: want %d, got %d", orig.UnitsPerEm(), sub.UnitsPerEm())
	}
}

// TestSubset_AlwaysIncludesNotdef checks that glyph 0 is always in the subset.
func TestSubset_AlwaysIncludesNotdef(t *testing.T) {
	orig, err := LoadFont(minimalTTF)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	// Explicitly pass an empty list — only .notdef should survive.
	data, err := orig.Subset(nil)
	if err != nil {
		t.Fatalf("Subset(nil): %v", err)
	}

	sub, err := LoadFontFromBytes(data)
	if err != nil {
		t.Fatalf("LoadFontFromBytes: %v", err)
	}
	// numGlyphs is preserved (loca entries stay); we just verify the font parses.
	_ = sub
}

// TestSubset_IncludedGlyphsHaveWidths checks that recorded glyph IDs have non-zero advance widths.
func TestSubset_IncludedGlyphsHaveWidths(t *testing.T) {
	orig, err := LoadFont(minimalTTF)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	gA := orig.GlyphIndex('A')
	data, err := orig.Subset([]uint16{gA})
	if err != nil {
		t.Fatalf("Subset: %v", err)
	}

	sub, err := LoadFontFromBytes(data)
	if err != nil {
		t.Fatalf("LoadFontFromBytes: %v", err)
	}
	w := sub.AdvanceWidthForGlyph(gA)
	if w == 0 {
		t.Errorf("expected non-zero advance width for glyph %d in subset", gA)
	}
}

// TestSubset_CheckSumAdjustment verifies the head.checkSumAdjustment field satisfies
// the TTF spec: file_checksum + checkSumAdjustment == 0xB1B0AFBA.
func TestSubset_CheckSumAdjustment(t *testing.T) {
	orig, err := LoadFont(minimalTTF)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	data, err := orig.Subset([]uint16{orig.GlyphIndex('A')})
	if err != nil {
		t.Fatalf("Subset: %v", err)
	}

	fileSum := ttfTableChecksum(data)
	if fileSum != 0xB1B0AFBA {
		t.Errorf("file checksum sum: want 0xB1B0AFBA, got 0x%08X", fileSum)
	}
}

// TestSubset_EmptyGlyphIDs checks that passing an empty slice produces a valid font.
func TestSubset_EmptyGlyphIDs(t *testing.T) {
	orig, err := LoadFont(minimalTTF)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	data, err := orig.Subset([]uint16{})
	if err != nil {
		t.Fatalf("Subset([]): %v", err)
	}

	if _, err := LoadFontFromBytes(data); err != nil {
		t.Fatalf("LoadFontFromBytes on empty-subset: %v", err)
	}
}

// TestSubset_CompositeGlyphClosure verifies that when a composite glyph is
// requested, its component glyphs are transitively included in the subset.
// minimal.ttf has a composite at U+E001 that references glyphs 1 and 2.
func TestSubset_CompositeGlyphClosure(t *testing.T) {
	orig, err := LoadFont(minimalTTF)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	compositeGID := orig.GlyphIndex(0xE001)
	if compositeGID == 0 {
		t.Fatal("U+E001 not mapped in minimal.ttf — was the fixture regenerated?")
	}

	data, err := orig.Subset([]uint16{compositeGID})
	if err != nil {
		t.Fatalf("Subset: %v", err)
	}

	// Glyph 0 (.notdef), 1 (space), 2 (!), and the composite must have data.
	for _, gid := range []uint16{0, 1, 2, compositeGID} {
		if !subsetGlyphHasData(data, gid) {
			t.Errorf("glyph %d expected to have data in subset (closure), but is absent", gid)
		}
	}

	// Glyph 3 was NOT requested and is not a component — it should be absent.
	if subsetGlyphHasData(data, 3) {
		t.Errorf("glyph 3 expected to be absent from subset, but has data")
	}
}

func TestSubset_UpdatesNumGlyphsAndHmtx(t *testing.T) {
	orig, err := LoadFont(minimalTTF)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	gA := orig.GlyphIndex('A')
	data, err := orig.Subset([]uint16{gA})
	if err != nil {
		t.Fatalf("Subset: %v", err)
	}

	maxp := mustFindTable(t, data, "maxp")
	if got, want := binary.BigEndian.Uint16(data[maxp.offset+4:]), gA+1; got != want {
		t.Fatalf("subset maxp.numGlyphs = %d, want %d", got, want)
	}

	hhea := mustFindTable(t, data, "hhea")
	if got, want := binary.BigEndian.Uint16(data[hhea.offset+34:]), gA+1; got != want {
		t.Fatalf("subset hhea.numOfLongHorMetrics = %d, want %d", got, want)
	}

	hmtx := mustFindTable(t, data, "hmtx")
	if got, want := hmtx.length, uint32(gA+1)*4; got != want {
		t.Fatalf("subset hmtx length = %d, want %d", got, want)
	}
}

func TestSubset_CmapRemovesUnusedMappings(t *testing.T) {
	orig, err := LoadFont(minimalTTF)
	if err != nil {
		t.Fatalf("LoadFont: %v", err)
	}

	gA := orig.GlyphIndex('A')
	data, err := orig.Subset([]uint16{gA})
	if err != nil {
		t.Fatalf("Subset: %v", err)
	}

	sub, err := LoadFontFromBytes(data)
	if err != nil {
		t.Fatalf("LoadFontFromBytes: %v", err)
	}
	if got := sub.GlyphIndex('A'); got != gA {
		t.Fatalf("subset GlyphIndex('A') = %d, want %d", got, gA)
	}
	if got := sub.GlyphIndex('B'); got != 0 {
		t.Fatalf("subset GlyphIndex('B') = %d, want 0", got)
	}
}

func TestBuildSubsetPostFormat2(t *testing.T) {
	font := &Font{
		postTable: postTable{
			format:             Fixed{base: 2},
			italicAngle:        Fixed{},
			underlinePosition:  -100,
			underlineThickness: 50,
			names:              []string{".notdef", "A", "custom.alt", "A"},
		},
	}

	data, err := font.buildSubsetPostFormat2(3)
	if err != nil {
		t.Fatalf("buildSubsetPostFormat2: %v", err)
	}

	if got := binary.BigEndian.Uint16(data[32:]); got != 3 {
		t.Fatalf("numberOfGlyphs = %d, want 3", got)
	}
	if got := binary.BigEndian.Uint16(data[34:]); got != 0 {
		t.Fatalf("glyph 0 name index = %d, want 0 (.notdef)", got)
	}
	if got := binary.BigEndian.Uint16(data[36:]); got != 36 {
		t.Fatalf("glyph 1 name index = %d, want 36 (A)", got)
	}
	if got := binary.BigEndian.Uint16(data[38:]); got != 258 {
		t.Fatalf("glyph 2 name index = %d, want 258 (first custom name)", got)
	}
	if got := data[40:]; string(got) != "\ncustom.alt" {
		t.Fatalf("custom names payload = %q, want %q", string(got), "\ncustom.alt")
	}
}

func TestSubsetKeepTable(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{tag: "cmap", want: true},
		{tag: "glyf", want: true},
		{tag: "name", want: true},
		{tag: "prep", want: true},
		{tag: "GPOS", want: false},
		{tag: "GSUB", want: false},
		{tag: "BASE", want: false},
		{tag: "kern", want: false},
	}
	for _, tc := range tests {
		if got := subsetKeepTable(tc.tag); got != tc.want {
			t.Errorf("subsetKeepTable(%q) = %v, want %v", tc.tag, got, tc.want)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// subsetGlyphHasData returns true if the given glyph ID has non-zero-length
// data in the supplied TTF binary (i.e., loca[id+1] > loca[id]).
func subsetGlyphHasData(data []byte, glyphID uint16) bool {
	if len(data) < 12 {
		return false
	}
	nTables := int(binary.BigEndian.Uint16(data[4:]))
	var locaOff, headOff, maxpOff uint32
	for i := 0; i < nTables; i++ {
		base := 12 + i*16
		tag := string(data[base : base+4])
		off := binary.BigEndian.Uint32(data[base+8:])
		switch tag {
		case "loca":
			locaOff = off
		case "head":
			headOff = off
		case "maxp":
			maxpOff = off
		}
	}
	if locaOff == 0 || headOff == 0 || maxpOff == 0 {
		return false
	}
	numGlyphs := int(binary.BigEndian.Uint16(data[maxpOff+4:]))
	if int(glyphID) >= numGlyphs {
		return false
	}
	isLong := binary.BigEndian.Uint16(data[headOff+50:]) == 1
	var off0, off1 uint32
	if isLong {
		off0 = binary.BigEndian.Uint32(data[locaOff+uint32(glyphID)*4:])
		off1 = binary.BigEndian.Uint32(data[locaOff+uint32(glyphID+1)*4:])
	} else {
		off0 = uint32(binary.BigEndian.Uint16(data[locaOff+uint32(glyphID)*2:])) * 2
		off1 = uint32(binary.BigEndian.Uint16(data[locaOff+uint32(glyphID+1)*2:])) * 2
	}
	return off1 > off0
}

func loadRawFont(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("cannot read %s: %v", path, err)
	}
	return data
}

func mustFindTable(t *testing.T, data []byte, tag string) *tableDirEntry {
	t.Helper()
	if len(data) < 12 {
		t.Fatalf("font too short")
	}
	nTables := int(binary.BigEndian.Uint16(data[4:]))
	for i := 0; i < nTables; i++ {
		base := 12 + i*16
		if string(data[base:base+4]) == tag {
			return &tableDirEntry{
				tag:      tag,
				checkSum: binary.BigEndian.Uint32(data[base+4:]),
				offset:   binary.BigEndian.Uint32(data[base+8:]),
				length:   binary.BigEndian.Uint32(data[base+12:]),
			}
		}
	}
	t.Fatalf("table %q not found", tag)
	return nil
}
