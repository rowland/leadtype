package ttf

import "testing"
// import "fmt"

func TestCmapTable_glyphIndex_Arial(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	expectUI16(t, "registered", 138, f.cmapTable.glyphIndex(0xAE))
	expectUI16(t, "copyright", 139, f.cmapTable.glyphIndex(0xA9))
	expectUI16(t, "epsilon", 304, f.cmapTable.glyphIndex(0x03B5))
	expectUI16(t, "l-cedilla", 441, f.cmapTable.glyphIndex(0x013B))
	expectUI16(t, "afii57414", 905, f.cmapTable.glyphIndex(0x0626))
	expectUI16(t, "trademark", 140, f.cmapTable.glyphIndex(0x2122))

	expectUI16(t, "reversed-e", 1688, f.cmapTable.glyphIndex(0x018E))
	expectUI16(t, "t-with-comma", 1801, f.cmapTable.glyphIndex(0x021A))
	// for i := uint16(0); i < 65535; i++ {
	// 	fmt.Printf("[%d] %d\n", i, f.cmapTable.glyphIndex(i))
	// }
}

func TestCmapTable_glyphIndex_Courier(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Courier New.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	expectUI16(t, "registered", 138, f.cmapTable.glyphIndex(0xAE))
	expectUI16(t, "copyright", 139, f.cmapTable.glyphIndex(0xA9))
	expectUI16(t, "epsilon", 304, f.cmapTable.glyphIndex(0x03B5))
	expectUI16(t, "l-cedilla", 441, f.cmapTable.glyphIndex(0x013B))
	expectUI16(t, "afii57414", 905, f.cmapTable.glyphIndex(0x0626))
	expectUI16(t, "trademark", 140, f.cmapTable.glyphIndex(0x2122))

	expectUI16(t, "reversed-e", 1693, f.cmapTable.glyphIndex(0x018E))
	expectUI16(t, "t-with-comma", 1806, f.cmapTable.glyphIndex(0x021A))
	// for i := uint16(0); i < 256; i++ {
	// 	fmt.Printf("[%d] %d\n", i, f.cmapTable.glyphIndex(i))
	// }
}

func BenchmarkGlyphIndex(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		panic("Error loading font")
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.cmapTable.glyphIndex(uint16(i))
	}
}
