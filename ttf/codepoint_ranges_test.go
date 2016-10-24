// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import "testing"

func TestCodepointRanges_RangeForRune(t *testing.T) {
	expectI8(t, "Bit", 0, CodepointRanges.RangeForRune(0x50).Bit)
	expectI8(t, "Bit", 1, CodepointRanges.RangeForRune(0x85).Bit)
	expectI8(t, "Bit", 4, CodepointRanges.RangeForRune(0x1D80).Bit)
}

func TestCodepointRanges_RangeByName(t *testing.T) {
	expect(t, "Bogus range name", CodepointRangesByName["Bogus"] == nil)
	expectI8(t, "Missing range name", 30, CodepointRangesByName["Greek Extended"].Bit)
}

func TestNewCodepointRangeSet(t *testing.T) {
	empty, err := NewCodepointRangeSet()
	if err != nil {
		t.Error(err)
	}
	if len(empty) != 0 {
		t.Error("Expecting empty list")
	}

	_, err = NewCodepointRangeSet("bogus")
	if err == nil {
		t.Error("Expecting error from bogus range.")
	}

	single, err := NewCodepointRangeSet("Basic Latin")
	if err != nil {
		t.Error(err)
	}
	if len(single) != 1 {
		t.Error("Expecting single CodepointRange.")
	}
	if single[0] != &NestedCodepointRanges[0][0] {
		t.Errorf("Expecting %v, got %v", &NestedCodepointRanges[0][0], single[0])
	}

	double, err := NewCodepointRangeSet("Basic Latin", "Greek Extended")
	if err != nil {
		t.Error(err)
	}
	if len(double) != 2 {
		t.Error("Expecting 2 CodepointRanges.")
	}
	if double[0] != &NestedCodepointRanges[0][0] {
		t.Errorf("Expecting %v, got %v", &NestedCodepointRanges[0][0], double[0])
	}
	if double[1] != &NestedCodepointRanges[30][0] {
		t.Errorf("Expecting %v, got %v", &NestedCodepointRanges[30][0], double[1])
	}
}

// 23.2 ns
// 21.2 ns go1.1.1
// 22.1 ns go1.2.1
// 21.3 ns go1.4.2
// 13.2 ns go1.6.2 mbp
// 11.8 ns go1.7.3
func BenchmarkRangeForRune(b *testing.B) {
	count := 0
	for i := 0; i < b.N; i++ {
		rune := rune(i & 0xFFFF)
		cpr := CodepointRanges.RangeForRune(rune)
		if cpr != nil {
			count++
			if rune < cpr.Low || rune > cpr.High {
				panic("Oops")
			}
		}
	}
}

// 98 ns
// 43.8 ns go1.1.1
// 48.2 ns go1.2.1
// 45.5 ns go1.4.2
// 27.7 ns go1.6.2 mbp
func BenchmarkRangeByName(b *testing.B) {
	count := 0
	for i := 0; i < b.N; i++ {
		name := CodepointRanges[count].Name
		cpr := CodepointRangesByName[name]
		if cpr == nil {
			panic("Oops")
		}
		count = (count + 1) % len(CodepointRanges)
	}
}
