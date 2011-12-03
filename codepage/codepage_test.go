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
