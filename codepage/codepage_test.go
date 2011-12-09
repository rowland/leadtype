package codepage

import "testing"

func TestRanges_CharForCodepoint(t *testing.T) {
	// Full ASCII range
	for cp := 0; cp < 128; cp++ {
		if ch, found := ASCII.CharForCodepoint(cp); found {
			if ch != cp {
				t.Errorf("ASCII: expected %d, got %d", cp, ch)
			}
		} else {
			t.Errorf("ASCII: codepoint %d not found", cp)
		}
	}
	// Outside of ASCII range
	if _, found := ASCII.CharForCodepoint(128); found {
		t.Errorf("ASCII: codepoint %d should not be found", 128)
	}
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

// 7.8 ns
func BenchmarkRanges_CharForCodepoint_ASCII(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ASCII.CharForCodepoint(i % 0x2122)
	}
}

// 29.9 ns
func BenchmarkRanges_CharForCodepoint_CP1250(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1250.CharForCodepoint(i % 0x2122)
	}
}

// 25.9 ns
func BenchmarkRanges_CharForCodepoint_CP1251(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1251.CharForCodepoint(i % 0x2122)
	}
}

// 22.0 ns
func BenchmarkRanges_CharForCodepoint_CP1252(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1252.CharForCodepoint(i % 0x2122)
	}
}

// 17.6 ns
func BenchmarkRanges_CharForCodepoint_CP1253(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1253.CharForCodepoint(i % 0x2122)
	}
}

// 21.7 ns
func BenchmarkRanges_CharForCodepoint_CP1254(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1254.CharForCodepoint(i % 0x2122)
	}
}

// 26.4 ns
func BenchmarkRanges_CharForCodepoint_CP1256(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1256.CharForCodepoint(i % 0x2122)
	}
}

// 29.9 ns
func BenchmarkRanges_CharForCodepoint_CP1257(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1257.CharForCodepoint(i % 0x2122)
	}
}

// 25.4 ns
func BenchmarkRanges_CharForCodepoint_CP1258(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP1258.CharForCodepoint(i % 0x2122)
	}
}

// 16.3 ns
func BenchmarkRanges_CharForCodepoint_CP874(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CP874.CharForCodepoint(i % 0x2122)
	}
}

// 8.2 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_1.CharForCodepoint(i % 0x2122)
	}
}

// 30.2 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_2.CharForCodepoint(i % 0x2122)
	}
}

// 26.8 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_3.CharForCodepoint(i % 0x2122)
	}
}

// 30.3 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_4.CharForCodepoint(i % 0x2122)
	}
}

// 17.8 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_5.CharForCodepoint(i % 0x2122)
	}
}

// 17.9 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_6(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_6.CharForCodepoint(i % 0x2122)
	}
}

// 18.1 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_7(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_7.CharForCodepoint(i % 0x2122)
	}
}

// 14.3 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_8.CharForCodepoint(i % 0x2122)
	}
}

// 18.5 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_9(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_9.CharForCodepoint(i % 0x2122)
	}
}

// 29.9 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_10.CharForCodepoint(i % 0x2122)
	}
}

// 10.7 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_11(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_11.CharForCodepoint(i % 0x2122)
	}
}

// 26.0 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_13(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_13.CharForCodepoint(i % 0x2122)
	}
}

// 21.6 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_14(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_14.CharForCodepoint(i % 0x2122)
	}
}

// 18.1 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_15(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_15.CharForCodepoint(i % 0x2122)
	}
}

// 25.7 ns
func BenchmarkRanges_CharForCodepoint_ISO_8859_16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ISO_8859_16.CharForCodepoint(i % 0x2122)
	}
}

func TestCodepageRanges_CodepageForCodepoint(t *testing.T) {
	// Full ASCII range
	for cp := 0; cp < 128; cp++ {
		if page, found := CodepointCodepages.CodepageForCodepoint(cp); found {
			if page != idx_ASCII {
				t.Errorf("CodepointCodepages: expected %d, got %d", idx_ASCII, page)
			}
		} else {
			t.Errorf("CodepointCodepages: codepoint %d not found", cp)
		}
	}
	// Outside of ASCII range
	for cp := 128; cp < 256; cp++ {
		if page, found := CodepointCodepages.CodepageForCodepoint(cp); found {
			if page != idx_ISO_8859_1 {
				t.Errorf("CodepointCodepages: expected %d, got %d", idx_ISO_8859_1, page)
			}
		} else {
			t.Errorf("CodepointCodepages: codepoint %d not found", cp)
		}
	}
	// 2-codepoint range within CP1252
	if page, found := CodepointCodepages.CodepageForCodepoint(0x2013); found {
		if page != idx_CP1252 {
			t.Errorf("CodepointCodepages: expected %d, got %d", idx_CP1252, page)
		}
	} else {
		t.Errorf("CodepointCodepages: codepoint '%d' not found", 0x2013)
	}
	if page, found := CodepointCodepages.CodepageForCodepoint(0x2014); found {
		if page != idx_CP1252 {
			t.Errorf("CodepointCodepages: expected %d, got %d", idx_CP1252, page)
		}
	} else {
		t.Errorf("CodepointCodepages: codepoint %d not found", 0x2013)
	}
	// Codepoints before and after previous range
	// 2-codepoint range within CP1252
	if _, found := CodepointCodepages.CodepageForCodepoint(0x2012); found {
		t.Errorf("CodepointCodepages: codepoint %d should not be found", 0x2012)
	}
	if page, found := CodepointCodepages.CodepageForCodepoint(0x2015); found {
		if page != idx_ISO_8859_7 {
			t.Errorf("CodepointCodepages: expected %d, got %d", idx_ISO_8859_7, page)
		}
	} else {
		t.Errorf("CodepointCodepages: codepoint %x not found", 0x2015)
	}
}

// 26.1 ns
func BenchmarkCodepageRanges_CodepageForCodepoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CodepointCodepages.CodepageForCodepoint(i % 0x2122)
	}
}
