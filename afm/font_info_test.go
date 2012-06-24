// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package afm

import "testing"

func TestLoadFontInfo(t *testing.T) {
	fi, err := LoadFontInfo("data/fonts/Helvetica.afm")
	if err != nil {
		t.Fatalf("Error loading font info: %v", err)
	}

	expectS(t, "Family", "Helvetica", fi.Family())
	expectS(t, "Style", "Medium", fi.Style())

	expectI(t, "Ascent", 718, fi.Ascent())
	// expectI(t, "AvgWidth", 904, fi.AvgWidth())
	expectI(t, "CapHeight", 718, fi.CapHeight())
	expectS(t, "Copyright", "(c) 1985, 1987, 1989, 1990, 1997 Adobe Systems Incorporated.  All Rights Reserved.Helvetica is a trademark of Linotype-Hell AG and/or its subsidiaries.", fi.Copyright())
	expectI(t, "Descent", -207, fi.Descent())
	// expectS(t, "Designer", "Monotype Type Drawing Office - Robin Nicholas, Patricia Saunders 1982", fi.Designer())
	// expect(t, "Embeddable", fi.Embeddable())
	expectS(t, "FullName", "Helvetica", fi.FullName())
	expectF(t, "ItalicAngle", 0, fi.ItalicAngle())
	// expectS(t, "License", "You may use this font to display and print content as permitted by the license terms for the product in which this font is included. You may only (i) embed this font in content as permitted by the embedding restrictions included in this font; and (ii) temporarily download this font to a printer or other output device to help print content.", fi.License())
	expectI(t, "Leading", (718 - -207)*115/100, fi.Leading())
	// expectS(t, "Manufacturer", "The Monotype Corporation", fi.Manufacturer())
	expectS(t, "PostScriptName", "Helvetica", fi.PostScriptName())
	expectI(t, "StemV", 88, fi.StemV())
	// expectS(t, "Trademark", "Arial is a trademark of The Monotype Corporation in the United States and/or other countries.", fi.Trademark())
	// expectS(t, "UniqueName", "Monotype:Arial Regular:Version 5.01 (Microsoft)", fi.UniqueName())
	expectS(t, "Version", "002.000", fi.Version())
	expectI(t, "XHeight", 523, fi.XHeight())

	expectI(t, "BoundingBox.XMin", -166, fi.BoundingBox().XMin)
	expectI(t, "BoundingBox.YMin", -225, fi.BoundingBox().YMin)
	expectI(t, "BoundingBox.XMax", 1000, fi.BoundingBox().XMax)
	expectI(t, "BoundingBox.YMax", 931, fi.BoundingBox().YMax)
}

// 323,331 ns
// 169,941 ns
// 169,467 ns
// 161,887 ns weekly.2012-02-22
// 165,638 ns go1
func BenchmarkLoadFontInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadFontInfo("data/fonts/Helvetica.afm")
	}
}
