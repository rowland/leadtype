package ttf

import "testing"

type fontSelection struct {
	family         string
	weight         string
	style          string
	postscriptName string
	fontOptions    FontOptions
}

var testSelectData = []fontSelection{
	{"Arial", "", "", "ArialMT", FontOptions{}},
	{"Arial", "", "Italic", "Arial-ItalicMT", FontOptions{}},
	{"Arial", "Bold", "", "Arial-BoldMT", FontOptions{}},
	{"Arial", "Bold", "Italic", "Arial-BoldItalicMT", FontOptions{}},

	{"Courier New", "", "", "CourierNewPSMT", FontOptions{}},
	{"Courier New", "", "Italic", "CourierNewPS-ItalicMT", FontOptions{}},
	{"Courier New", "Bold", "", "CourierNewPS-BoldMT", FontOptions{}},
	{"Courier New", "Bold", "Italic", "CourierNewPS-BoldItalicMT", FontOptions{}},

	{"Times New Roman", "", "", "TimesNewRomanPSMT", FontOptions{}},
	{"Times New Roman", "", "Italic", "TimesNewRomanPS-ItalicMT", FontOptions{}},
	{"Times New Roman", "Bold", "", "TimesNewRomanPS-BoldMT", FontOptions{}},
	{"Times New Roman", "Bold", "Italic", "TimesNewRomanPS-BoldItalicMT", FontOptions{}},

	{"Arial Unicode MS", "", "", "ArialUnicodeMS", FontOptions{CharRanges: []int{59}}},
	{"Myanmar Sangam MN", "", "", "MyanmarSangamMN", FontOptions{CharRanges: []int{74}}},
}

func TestFontCollection(t *testing.T) {
	var fc FontCollection

	if err := fc.Add("/Library/Fonts/*.ttf"); err != nil {
		t.Error(err)
	}

	expectI(t, "Len", 114, fc.Len())
	for _, fs := range testSelectData {
		f, err := fc.Select(fs.family, fs.weight, fs.style, fs.fontOptions)
		if err == nil {
			expectS(t, fs.postscriptName, fs.postscriptName, f.PostScriptName())
		} else {
			t.Error(err)
		}
	}
	bogusFont, err2 := fc.Select("Bogus", "Regular", "", FontOptions{})
	expect(t, "Bogus Select Font", bogusFont == nil)
	expectS(t, "Bogus Select Error", "Font Bogus Regular not found", err2.Error())
}

// 81,980,000 ns
func BenchmarkFontCollection_Add(b *testing.B) {
	var fc FontCollection

	for i := 0; i < b.N; i++ {
		fc.Add("/Library/Fonts/*.ttf")
	}
}
