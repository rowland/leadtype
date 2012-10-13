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

// 23.2 ns
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
