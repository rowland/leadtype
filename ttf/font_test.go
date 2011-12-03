package ttf

import "testing"

func TestLoadFont(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	if f == nil {
		t.Fatal("Font not loaded")
	}

	expectUI32(t, "scalar", 0x00010000, f.scalar)
	expectUI16(t, "nTables", 0x0018, f.nTables)
	expectUI16(t, "searchRange", 0x0100, f.searchRange)
	expectUI16(t, "entrySelector", 0x0004, f.entrySelector)
	expectUI16(t, "rangeShift", 0x0080, f.rangeShift)

	for i, entry := range f.tableDir.entries {
		expectS(t, "arialTableNames", arialTableNames[i], entry.tag)
	}
	for _, tag := range arialTableNames {
		if f.tableDir.table(tag) == nil {
			t.Fatalf("Table for tag %s not found", tag)
		}
	}
	expectI(t, "UnitsPerEm", 2048, f.UnitsPerEm())
	expectI(t, "NumGlyphs", 3381, f.NumGlyphs())
	expectI(t, "BoundingBox.XMin", -1361, f.BoundingBox().XMin)
	expectI(t, "BoundingBox.YMin", -665, f.BoundingBox().YMin)
	expectI(t, "BoundingBox.XMax", 4096, f.BoundingBox().XMax)
	expectI(t, "BoundingBox.YMax", 2060, f.BoundingBox().YMax)

	expectI(t, "registered", 1509, f.AdvanceWidth(0xAE))
	expectI(t, "copyright", 1509, f.AdvanceWidth(0xA9))
	expectI(t, "epsilon", 913, f.AdvanceWidth(0x03B5))
	expectI(t, "l-cedilla", 1139, f.AdvanceWidth(0x013B))
	expectI(t, "afii57414", 1307, f.AdvanceWidth(0x0626))
	expectI(t, "trademark", 2048, f.AdvanceWidth(0x2122))
	expectI(t, "reversed-e", 1366, f.AdvanceWidth(0x018E))
	expectI(t, "t-with-comma", 1251, f.AdvanceWidth(0x021A))

	expectI(t, "Ascent", 1854, f.Ascent())
	expectI(t, "Descent", -434, f.Descent())
	expectF(t, "ItalicAngle", 0, f.ItalicAngle())
	expectI(t, "StemV", 87, f.StemV())
	expectI(t, "XHeight", 1062, f.XHeight())
	// Could not find MissingWidth in TTF.
	expectI(t, "Leading", 1854 - -434 + 67, f.Leading())
	// TODO: Verify this works in practice. Some examples indicate equivalence to lineGap instead.
	expectI(t, "MaxWidth", 4096, f.MaxWidth())
	expectI(t, "AvgWidth", 904, f.AvgWidth())
	expectI(t, "CapHeight", 1467, f.CapHeight())
	expect(t, "Embeddable", f.Embeddable())

	expectS(t, "PostScriptName", "ArialMT", f.PostScriptName())
	expectS(t, "fFllName", "Arial", f.FullName())
	expectS(t, "Family", "Arial", f.Family())
	expectS(t, "Style", "Regular", f.Style())
	expectS(t, "Version", "Version 5.01.2x", f.Version())
	expectS(t, "UniqueName", "Monotype:Arial Regular:Version 5.01 (Microsoft)", f.UniqueName())
	expectS(t, "Manufacturer", "The Monotype Corporation", f.Manufacturer())
	expectS(t, "Designer", "Monotype Type Drawing Office - Robin Nicholas, Patricia Saunders 1982", f.Designer())
	expectS(t, "Copyright", "© 2006 The Monotype Corporation. All Rights Reserved.", f.Copyright())
	expectS(t, "Trademark", "Arial is a trademark of The Monotype Corporation in the United States and/or other countries.", f.Trademark())
	expectS(t, "License", "You may use this font to display and print content as permitted by the license terms for the product in which this font is included. You may only (i) embed this font in content as permitted by the embedding restrictions included in this font; and (ii) temporarily download this font to a printer or other output device to help print content.", f.License())
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

// 50.4 ns
func BenchmarkAdvanceWidth(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		panic("Error loading font")
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.AdvanceWidth(int(f.os2Table.fsFirstCharIndex) + i%int(f.os2Table.fsLastCharIndex-f.os2Table.fsFirstCharIndex+1))
	}
}

// 9,151,820 ns
func BenchmarkLoadFont(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadFont("/Library/Fonts/Arial.ttf")
	}
}
