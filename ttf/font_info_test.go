package ttf

import "testing"

func TestLoadFontInfo(t *testing.T) {
	fi, err := LoadFontInfo("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatalf("Error loading font info: %v", err)
	}
	expectS(t, "Family", "Arial", fi.Family())
	expectS(t, "Style", "Regular", fi.Style())
}

// 1,077,216 ns
func BenchmarkLoadFontInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadFontInfo("/Library/Fonts/Arial.ttf")
	}
}
