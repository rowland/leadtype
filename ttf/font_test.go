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
