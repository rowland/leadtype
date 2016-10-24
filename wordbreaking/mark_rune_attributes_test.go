// Copyright 2013 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package wordbreaking

import "testing"

func TestMarkRuneAttributes(t *testing.T) {
	const quick = "The quick red fox jumps over the lazy brown dog."
	var quickFlags = []Flags{
		CharStop | WordStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // The
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // quick
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // red
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // fox
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // jumps
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // over
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // the
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // lazy
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // brown
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
		CharStop | WordStop, CharStop, CharStop, CharStop, CharStop, // Word-
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // breaking
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // test
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // with
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // regular
		CharStop | SoftBreak | WordStop, CharStop, CharStop, CharStop | WhiteSpace | SoftBreak, // and
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
// 977 ns go1.2.1
// 586 ns go1.6.2 mbp
func BenchmarkMarkRunAttributes(b *testing.B) {
	const quick = "The quick red fox jumps over the lazy brown dog."
	var flags [len(quick)]Flags

	for i := 0; i < b.N; i++ {
		MarkRuneAttributes(quick, flags[:])
	}
}
