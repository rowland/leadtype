package afm

import "testing"

func TestLoadFont(t *testing.T) {
	f, err := LoadFont("data/fonts/Helvetica.afm")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	if f == nil {
		t.Fatal("Font not loaded")
	}

	expectI(t, "UnitsPerEm", 1000, f.UnitsPerEm())
	expectI(t, "NumGlyphs", 315, f.NumGlyphs())
	expect(t, "Serif", !f.Serif())
	// expectI(t, "registered", 1509, f.AdvanceWidth(0xAE))
	// expectI(t, "copyright", 1509, f.AdvanceWidth(0xA9))
	// expectI(t, "epsilon", 913, f.AdvanceWidth(0x03B5))
	// expectI(t, "l-cedilla", 1139, f.AdvanceWidth(0x013B))
	// expectI(t, "afii57414", 1307, f.AdvanceWidth(0x0626))
	// expectI(t, "trademark", 2048, f.AdvanceWidth(0x2122))
	// expectI(t, "reversed-e", 1366, f.AdvanceWidth(0x018E))
	// expectI(t, "t-with-comma", 1251, f.AdvanceWidth(0x021A))
	// 
	// // Could not find MissingWidth in TTF.
	// expectI(t, "MaxWidth", 4096, f.MaxWidth())

}

// func BenchmarkAdvanceWidth(b *testing.B) {
// 	b.StopTimer()
// 	f, err := LoadFont("data/fonts/Helvetica.afm")
// 	if err != nil {
// 		panic("Error loading font")
// 	}
// 	b.StartTimer()
// 	for i := 0; i < b.N; i++ {
// 		f.AdvanceWidth(int(f.os2Table.fsFirstCharIndex) + i%int(f.os2Table.fsLastCharIndex-f.os2Table.fsFirstCharIndex+1))
// 	}
// }
