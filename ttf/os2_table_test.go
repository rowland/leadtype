// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import "testing"

func TestCharRanges(t *testing.T) {
	f, err := LoadFont("/Library/Fonts/Arial.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %s", err)
	}
	expect(t, "Basic Latin", f.CharRanges().IsSet(0))
	expect(t, "Latin-1 Supplement", f.CharRanges().IsSet(1))
	expect(t, "Latin Extended-A", f.CharRanges().IsSet(2))
	expect(t, "Latin Extended-B", f.CharRanges().IsSet(3))
	expect(t, "IPA Extensions", f.CharRanges().IsSet(4))
	expect(t, "Spacing Modifier Letters", f.CharRanges().IsSet(5))
	expect(t, "Combining Diacritical Marks", f.CharRanges().IsSet(6))
	expect(t, "Greek and Coptic", f.CharRanges().IsSet(7))

	expect(t, "Coptic", !f.CharRanges().IsSet(8))

	expect(t, "Cyrillic", f.CharRanges().IsSet(9))

	expect(t, "Armenian", !f.CharRanges().IsSet(10))
}

// 7.3 ns
// 3.6 ns
// 3.9 ns go1.2.1
func BenchmarkCharRanges(b *testing.B) {
	cr := CharRanges{3758107391, 3221256259, 9, 0}
	for i := 0; i < b.N; i++ {
		cr.IsSet(i % 128)
	}
}
