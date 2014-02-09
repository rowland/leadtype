// Copyright 2013 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package wordbreaking

import "testing"

func TestMarkRuneAttributes(t *testing.T) {
	const quick = "The quick red fox jumps over the lazy brown dog."
	var quickFlags = []Flags{
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace, // The
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // quick
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace, // red
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace, // fox
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // jumps
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // over
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace, // the
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // lazy
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // brown
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, // dog.
	}
	var flags [len(quick)]Flags

	MarkRuneAttributes(quick, flags[:])
	for i := 0; i < len(quick); i++ {
		if flags[i] != quickFlags[i] {
			t.Errorf("Expecting %d, got %d at index %d", quickFlags[i], flags[i], i)
		}
	}
}

func TestMarkRuneAttributes_with_hyphens(t *testing.T) {
	const hyphenTest = "Word-breaking test with regular and soft\u00ADhyphens."
	var hyphenFlags = []Flags{
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, // Word-
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // breaking
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // test
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // with
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace, // regular
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace, // and
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, 0, // soft-
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, // hyphens.
	}
	var flags [len(hyphenTest)]Flags

	MarkRuneAttributes(hyphenTest, flags[:])
	for i := 0; i < len(hyphenTest); i++ {
		if flags[i] != hyphenFlags[i] {
			t.Errorf("Expecting %d, got %d at index %d", hyphenFlags[i], flags[i], i)
		}
	}
}

// Baseline: 310 ns
// 833 ns
// 931 ns with hyphen tests
func BenchmarkMarkRunAttributes(b *testing.B) {
	const quick = "The quick red fox jumps over the lazy brown dog."
	var flags [len(quick)]Flags

	for i := 0; i < b.N; i++ {
		MarkRuneAttributes(quick, flags[:])
	}
}
