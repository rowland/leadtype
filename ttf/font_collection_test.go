package ttf

import "testing"

type fontSelection struct {
	family         string
	weight         string
	style          string
	postscriptName string
}

var testSelectData = []fontSelection{
	{ "Arial", "", "", "ArialMT"},
	{ "Arial", "", "Italic", "Arial-ItalicMT"},
	{ "Arial", "Bold", "", "Arial-BoldMT"},
	{ "Arial", "Bold", "Italic", "Arial-BoldItalicMT"},

	{ "Courier New", "", "", "CourierNewPSMT"},
	{ "Courier New", "", "Italic", "CourierNewPS-ItalicMT"},
	{ "Courier New", "Bold", "", "CourierNewPS-BoldMT"},
	{ "Courier New", "Bold", "Italic", "CourierNewPS-BoldItalicMT"},

	{ "Times New Roman", "", "", "TimesNewRomanPSMT"},
	{ "Times New Roman", "", "Italic", "TimesNewRomanPS-ItalicMT"},
	{ "Times New Roman", "Bold", "", "TimesNewRomanPS-BoldMT"},
	{ "Times New Roman", "Bold", "Italic", "TimesNewRomanPS-BoldItalicMT"},
}

func TestFontCollection(t *testing.T) {
	var fc FontCollection

	if err := fc.Add("/Library/Fonts/*.ttf"); err != nil {
		t.Error(err)
	}

	expectI(t, "Len", 114, fc.Len())
	for _, fs := range testSelectData {
		f, err := fc.Select(fs.family, fs.weight, fs.style, FontOptions{})
		if err == nil {
			expectS(t, fs.postscriptName, fs.postscriptName, f.PostScriptName())
		} else {
			t.Error(err)
		}
	}
	bogusFont, err2 := fc.Select("Bogus", "Regular", "", FontOptions{})
	expect(t, "Bogus Select Font", bogusFont == nil)
	expectS(t, "Bogus Select Error", "Font Bogus Regular not found", err2.String())
}

// 124,346,300 ns
func BenchmarkFontCollection_Add(b *testing.B) {
	var fc FontCollection

	for i := 0; i < b.N; i++ {
		fc.Add("/Library/Fonts/*.ttf")
	}
}
