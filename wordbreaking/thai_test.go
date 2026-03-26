// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package wordbreaking

import (
	"strings"
	"testing"
)

func TestSegmentThai_noThai(t *testing.T) {
	input := "Hello, world!"
	if got := SegmentThai(input); got != input {
		t.Errorf("SegmentThai(%q) = %q, want unchanged", input, got)
	}
}

func TestSegmentThai_singleWord(t *testing.T) {
	// "ภาษา" (language) is a single Thai word; no ZWS should be inserted.
	input := "ภาษา"
	got := SegmentThai(input)
	if strings.ContainsRune(got, ZeroWidthSpace) {
		t.Errorf("SegmentThai(%q) inserted unexpected ZWS: %q", input, got)
	}
	if got != input {
		t.Errorf("SegmentThai(%q) = %q, want unchanged", input, got)
	}
}

func TestSegmentThai_twoWords(t *testing.T) {
	// "เขาไป" = "เขา" (he/she) + "ไป" (go); both in the dictionary, compound is not.
	input := "เขาไป"
	want := "เขา\u200Bไป"
	got := SegmentThai(input)
	if got != want {
		t.Errorf("SegmentThai(%q) = %q, want %q", input, got, want)
	}
}

func TestSegmentThai_mixed(t *testing.T) {
	// Non-Thai content surrounding a single Thai word should be unchanged.
	input := "Hello ภาษา world"
	got := SegmentThai(input)
	if got != input {
		t.Errorf("SegmentThai(%q) = %q, want unchanged", input, got)
	}
}

func TestSegmentThai_markRuneAttributes(t *testing.T) {
	// Verify that ZWS inserted by SegmentThai produces SoftBreak flags
	// when passed through MarkRuneAttributes.
	segmented := SegmentThai("เขาไป")
	flags := make([]Flags, len(segmented))
	MarkRuneAttributes(segmented, flags)

	// Find the ZWS byte offset in the segmented string.
	zwsOffset := -1
	for i, r := range segmented {
		if r == ZeroWidthSpace {
			zwsOffset = i
			break
		}
	}
	if zwsOffset < 0 {
		t.Fatal("expected ZWS in segmented Thai text, found none")
	}
	if flags[zwsOffset]&SoftBreak == 0 {
		t.Errorf("ZWS at offset %d missing SoftBreak flag, got %08b", zwsOffset, flags[zwsOffset])
	}
}

func TestMarkRuneAttributes_thaiWithoutInsertedZWS(t *testing.T) {
	text := "เขาไป"
	flags := make([]Flags, len(text))
	MarkRuneAttributes(text, flags)

	breakOffset := len("เขา")
	if flags[breakOffset]&SoftBreak == 0 {
		t.Fatalf("expected soft break at offset %d, got %08b", breakOffset, flags[breakOffset])
	}
	if flags[breakOffset]&WordStop == 0 {
		t.Fatalf("expected word stop at offset %d, got %08b", breakOffset, flags[breakOffset])
	}
}

func BenchmarkSegmentThai(b *testing.B) {
	const text = "ภาษาไทยเป็นภาษาที่ใช้ในประเทศไทย"
	for i := 0; i < b.N; i++ {
		SegmentThai(text)
	}
}
