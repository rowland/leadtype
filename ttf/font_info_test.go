// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import "testing"

func TestLoadFontInfo(t *testing.T) {
	fi, err := LoadFontInfo("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatalf("Error loading font info: %v", err)
	}
	expectS(t, "Filename", "/Library/Fonts/Arial.ttf", fi.Filename())
	expectUI32(t, "scalar", 0x00010000, fi.scalar)
	expectUI16(t, "nTables", 0x0018, fi.nTables)
	expectUI16(t, "searchRange", 0x0100, fi.searchRange)
	expectUI16(t, "entrySelector", 0x0004, fi.entrySelector)
	expectUI16(t, "rangeShift", 0x0080, fi.rangeShift)

	expectS(t, "Family", "Arial", fi.Family())
	expectS(t, "Style", "Regular", fi.Style())

	for i, entry := range fi.tableDir.entries {
		expectS(t, "arialTableNames", arialTableNames[i], entry.tag)
	}
	for _, tag := range arialTableNames {
		if fi.tableDir.table(tag) == nil {
			t.Fatalf("Table for tag %s not found", tag)
		}
	}

	expectI(t, "AvgWidth", 904, fi.AvgWidth())
	expectI(t, "CapHeight", 1467, fi.CapHeight())
	expectS(t, "Copyright", "© 2006 The Monotype Corporation. All Rights Reserved.", fi.Copyright())
	expectS(t, "Designer", "Monotype Type Drawing Office - Robin Nicholas, Patricia Saunders 1982", fi.Designer())
	expect(t, "Embeddable", fi.Embeddable())
	expectS(t, "FullName", "Arial", fi.FullName())
	expectS(t, "License", "You may use this font to display and print content as permitted by the license terms for the product in which this font is included. You may only (i) embed this font in content as permitted by the embedding restrictions included in this font; and (ii) temporarily download this font to a printer or other output device to help print content.", fi.License())
	expectS(t, "Manufacturer", "The Monotype Corporation", fi.Manufacturer())
	expectS(t, "PostScriptName", "ArialMT", fi.PostScriptName())
	expectI(t, "StemV", 87, fi.StemV())
	expectS(t, "Trademark", "Arial is a trademark of The Monotype Corporation in the United States and/or other countries.", fi.Trademark())
	expectS(t, "UniqueName", "Monotype:Arial Regular:Version 5.01 (Microsoft)", fi.UniqueName())
	expectS(t, "Version", "Version 5.01.2x", fi.Version())
	expectI(t, "XHeight", 1062, fi.XHeight())

	cr := fi.CharRanges()
	expectUI32(t, "CharRanges[0]", 3758107391, cr[0])
	expectUI32(t, "CharRanges[1]", 3221256259, cr[1])
	expectUI32(t, "CharRanges[2]", 9, cr[2])
	expectUI32(t, "CharRanges[3]", 0, cr[3])
}

var arialTableNames = []string{
	"DSIG",
	"GDEF",
	"GPOS",
	"GSUB",
	"JSTF",
	"LTSH",
	"OS/2",
	"PCLT",
	"VDMX",
	"cmap",
	"cvt ",
	"fpgm",
	"gasp",
	"glyf",
	"hdmx",
	"head",
	"hhea",
	"hmtx",
	"kern",
	"loca",
	"maxp",
	"name",
	"post",
	"prep",
}

// 1,077,216 ns
//   472,090 ns
//   351,630 ns go1.1.1
//   297,153 ns go1.2.1
//   434,997 ns go1.4.2
//   184,676 ns go1.6.2 mbp
//   169,796 ns go1.7.3
func BenchmarkLoadFontInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadFontInfo("/Library/Fonts/Arial.ttf")
	}
}
