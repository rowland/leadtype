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
	// Could not find MissingWidth in TTF.
	expectI(t, "Leading", 1854 - -434 + 67, f.Leading())
	// TODO: Verify this works in practice. Some examples indicate equivalence to lineGap instead.
	expectI(t, "MaxWidth", 4096, f.MaxWidth())

}

// 9,151,820 ns
func BenchmarkLoadFont(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadFont("/Library/Fonts/Arial.ttf")
	}
}
