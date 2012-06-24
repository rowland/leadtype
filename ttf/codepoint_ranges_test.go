// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import "testing"

func TestCodepointRanges(t *testing.T) {
	expectI8(t, "Bit", 0, CodepointRanges.RangeForRune(0x50).Bit)
	expectI8(t, "Bit", 1, CodepointRanges.RangeForRune(0x85).Bit)
	expectI8(t, "Bit", 4, CodepointRanges.RangeForRune(0x1D80).Bit)
}

// 23.2 ns
func BenchmarkRangeForRune(b *testing.B) {
	count := 0
	for i := 0; i < b.N; i++ {
		rune := uint32(i & 0xFFFF)
		cpr := CodepointRanges.RangeForRune(rune)
		if cpr != nil {
			count++
			if rune < cpr.Low || rune > cpr.High {
				panic("Oops")
			}
		}
	}
}
