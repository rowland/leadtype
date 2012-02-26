package afm

import "testing"

func aw(width int, err error) int {
	return width
}

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
	expectI(t, "registered", 737, aw(f.AdvanceWidth(0xAE)))
	expectI(t, "copyright", 737, aw(f.AdvanceWidth(0xA9)))
	// expectI(t, "epsilon", 913, f.AdvanceWidth(0x03B5))
	// expectI(t, "l-cedilla", 1139, f.AdvanceWidth(0x013B))
	// expectI(t, "afii57414", 1307, f.AdvanceWidth(0x0626))
	expectI(t, "trademark", 1000, aw(f.AdvanceWidth(0x2122)))
	// expectI(t, "reversed-e", 1366, f.AdvanceWidth(0x018E))
	// expectI(t, "t-with-comma", 1251, f.AdvanceWidth(0x021A))
	// 
	// // Could not find MissingWidth in TTF.
	// expectI(t, "MaxWidth", 4096, f.MaxWidth())

}

// 3,956,700 ns
// 2,446,984 ns
// 2,284,165 ns/2.284 ms
func BenchmarkLoadFont(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadFont("data/fonts/Helvetica.afm")
	}
}

// 58.4 ns
// 71.3 ns
// 41.5 ns
func BenchmarkAdvanceWidth(b *testing.B) {
	b.StopTimer()
	f, err := LoadFont("data/fonts/Helvetica.afm")
	if err != nil {
		panic("Error loading font")
	}
	min, max := f.charMetrics[0].code, f.charMetrics[len(f.charMetrics)-1].code
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		f.AdvanceWidth(min + rune(i)%(max-min+1))
	}
}
