// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package codepage

import "testing"

func TestRanges_CharForCodepoint(t *testing.T) {
	// 2-codepoint range within CP1252
	if ch, found := CP1252.CharForCodepoint(0x2013); found {
		if ch != 150 {
			t.Errorf("CP1252: expected %d, got %d", 150, ch)
		}
	} else {
		t.Errorf("CP1252: codepoint %d not found", 0x2013)
	}
	if ch, found := CP1252.CharForCodepoint(0x2014); found {
		if ch != 151 {
			t.Errorf("CP1252: expected %d, got %d", 151, ch)
		}
	} else {
		t.Errorf("CP1252: codepoint %d not found", 0x2013)
	}
	// Codepoints before and after previous range
	if _, found := CP1252.CharForCodepoint(0x2012); found {
		t.Errorf("CP1252: codepoint %d should not be found", 0x2012)
	}
	if _, found := CP1252.CharForCodepoint(0x2015); found {
		t.Errorf("CP1252: codepoint %d should not be found", 0x2015)
	}
}

// 29.9 ns
// 27.6 ns go1.1.1
// 29.2 ns go1.2.1
// 29.4 ns go1.4.2
// 15.5 ns go1.6.2 mbp
// 15.9 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_CP1250(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1250.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 25.9 ns
// 24.2 ns go1.1.1
// 25.9 ns go1.2.1
// 25.4 ns go1.4.2
// 13.8 ns go1.6.2 mbp
// 13.8 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_CP1251(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1251.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 22.0 ns
// 20.5 ns go1.1.1
// 21.6 ns go1.2.1
// 21.9 ns go1.4.2
// 12.4 ns go1.6.2 mbp
// 11.4 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_CP1252(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1252.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 17.6 ns
// 16.3 ns go1.1.1
// 17.2 ns go1.2.1
// 17.1 ns go1.4.2
// 10.1 ns go1.6.2 mbp
//  9.4 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_CP1253(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1253.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 21.7 ns
// 20.2 ns go1.1.1
// 21.5 ns go1.2.1
// 21.3 ns go1.4.2
// 12.0 ns go1.6.2 mbp
// 11.6 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_CP1254(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1254.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 26.4 ns
// 24.2 ns go1.1.1
// 25.4 ns go1.2.1
// 25.4 ns go1.4.2
// 13.9 ns go1.6.2 mbp
// 13.5 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_CP1256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1256.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 29.9 ns
// 27.4 ns go1.1.1
// 28.9 ns go1.2.1
// 28.7 ns go1.4.2
// 15.4 ns go1.6.2 mbp
// 15.8 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_CP1257(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1257.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 25.4 ns
// 23.5 ns go1.1.1
// 25.0 ns go1.2.1
// 24.8 ns go1.4.2
// 13.6 ns go1.6.2 mbp
// 13.6 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_CP1258(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1258.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 16.3 ns
// 15.1 ns go1.1.1
// 16.0 ns go1.2.1
// 16.1 ns go1.4.2
//  9.3 ns go1.6.2 mbp
//  8.7 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_CP874(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP874.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 8.2 ns
// 7.3 ns go1.1.1
// 8.0 ns go1.2.1
// 8.6 ns go1.4.2
// 5.2 ns go1.6.2 mbp
// 5.2 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_1.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 30.2 ns
// 28.2 ns go1.1.1
// 29.9 ns go1.2.1
// 29.4 ns go1.4.2
// 16.1 ns go1.6.2 mbp
// 15.9 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_2.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 26.8 ns
// 24.7 ns go1.1.1
// 26.1 ns go1.2.1
// 26.2 ns go1.4.2
// 14.5 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_3.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 30.3 ns
// 28.8 ns go1.1.1
// 29.6 ns go1.2.1
// 29.4 ns go1.4.2
// 16.0 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_4.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 17.8 ns
// 16.4 ns go1.1.1
// 17.3 ns go1.2.1
// 17.6 ns go1.4.2
// 14.5 ns go1.6.2 mbp
//  9.4 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_5.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 17.9 ns
// 18.3 ns go1.1.1
// 17.5 ns go1.2.1
// 17.8 ns go1.4.2
// 10.5 ns go1.6.2 mbp
// 10.0 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_6.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 18.1 ns
// 24.8 ns go1.1.1
// 17.8 ns go1.2.1
// 18.1 ns go1.4.2
// 10.6 ns go1.6.2 mbp
//  9.8 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_7(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_7.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 14.3 ns
// 14.4 ns go1.1.1
// 14.2 ns go1.2.1
// 14.3 ns go1.4.2
//  8.4 ns go1.6.2 mbp
//  8.2 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_8.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 18.5 ns
// 18.4 ns go1.1.1
// 18.1 ns go1.2.1
// 18.2 ns go1.4.2
// 10.8 ns go1.6.2 mbp
// 10.1 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_9(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_9.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 29.9 ns
// 29.0 ns go1.1.1
// 29.5 ns go1.2.1
// 29.2 ns go1.4.2
// 15.9 ns go1.6.2 mbp
// 16.1 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_10.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 10.7 ns
// 11.5 ns go1.1.1
// 11.0 ns go1.2.1
// 11.1 ns go1.4.2
//  7.0 ns go1.6.2 mbp
//  6.5 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_11(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_11.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 26.0 ns
// 26.3 ns go1.1.1
// 25.6 ns go1.2.1
// 25.1 ns go1.4.2
// 13.7 ns go1.6.2 mbp
// 13.6 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_13(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_13.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 21.6 ns
// 21.8 ns go1.1.1
// 21.3 ns go1.2.1
// 21.2 ns go1.4.2
// 11.8 ns go1.6.2 mbp
// 11.4 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_14(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_14.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 18.1 ns
// 17.7 ns go1.1.1
// 17.9 ns go1.2.1
// 17.9 ns go1.4.2
// 10.3 ns go1.6.2 mbp
// 10.1 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_15(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_15.CharForCodepoint(rune(i) % 0x2122)
	}
}

// 25.7 ns
// 25.6 ns go1.1.1
// 25.6 ns go1.2.1
// 25.0 ns go1.4.2
// 13.7 ns go1.6.2 mbp
// 13.4 ns go1.7.3
func BenchmarkRanges_CharForCodepoint_ISO_8859_16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_16.CharForCodepoint(rune(i) % 0x2122)
	}
}

func TestCodepageRanges_CodepageIndexForCodepoint(t *testing.T) {
	// 2-codepoint range within CP1252
	if page, found := codepointCodepages.CodepageIndexForCodepoint(0x2013); found {
		if page != Idx_CP1252 {
			t.Errorf("codepointCodepages: expected %d, got %d", Idx_CP1252, page)
		}
	} else {
		t.Errorf("codepointCodepages: codepoint '%d' not found", 0x2013)
	}
	if page, found := codepointCodepages.CodepageIndexForCodepoint(0x2014); found {
		if page != Idx_CP1252 {
			t.Errorf("codepointCodepages: expected %d, got %d", Idx_CP1252, page)
		}
	} else {
		t.Errorf("codepointCodepages: codepoint %d not found", 0x2013)
	}
	// Codepoints before and after previous range
	// 2-codepoint range within CP1252
	if _, found := codepointCodepages.CodepageIndexForCodepoint(0x2012); found {
		t.Errorf("codepointCodepages: codepoint %d should not be found", 0x2012)
	}
	if page, found := codepointCodepages.CodepageIndexForCodepoint(0x2015); found {
		if page != Idx_ISO_8859_7 {
			t.Errorf("codepointCodepages: expected %d, got %d", Idx_ISO_8859_7, page)
		}
	} else {
		t.Errorf("codepointCodepages: codepoint %x not found", 0x2015)
	}
}

// 26.1 ns
// 25.2 ns go1.1.1
// 24.7 ns go1.2.1
// 24.4 ns go1.4.2
// 13.1 ns go1.6.2 mbp
// 13.1 ns go1.7.3
func BenchmarkCodepageRanges_CodepageIndexForCodepoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		codepointCodepages.CodepageIndexForCodepoint(rune(i) % 0x2122)
	}
}
