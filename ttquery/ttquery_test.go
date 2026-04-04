// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rowland/leadtype/ttf"
)

func TestSupportsArabic(t *testing.T) {
	var ranges ttf.CharRanges
	if supportsArabic(&ranges) {
		t.Fatal("expected empty ranges not to report Arabic support")
	}

	ranges[13>>5] = 1 << uint(13&31)
	if !supportsArabic(&ranges) {
		t.Fatal("expected Arabic Unicode range bit to report Arabic support")
	}

	ranges = ttf.CharRanges{}
	ranges[63>>5] = 1 << uint(63&31)
	if !supportsArabic(&ranges) {
		t.Fatal("expected Arabic presentation forms A bit to report Arabic support")
	}

	ranges = ttf.CharRanges{}
	ranges[67>>5] = 1 << uint(67&31)
	if !supportsArabic(&ranges) {
		t.Fatal("expected Arabic presentation forms B bit to report Arabic support")
	}
}

func TestPrintCapabilitiesSorted(t *testing.T) {
	var buf bytes.Buffer
	printCapabilities(&buf)
	got := strings.TrimSpace(buf.String())
	want := strings.Join([]string{
		"arabic\tfonts whose OS/2 Unicode ranges indicate Arabic support",
		"arabic-shaping\tfonts with Arabic Unicode coverage and OpenType shaping tables",
		"emoji\tfonts with outline glyph coverage for representative emoji-style symbols",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected capabilities output: %q", got)
	}
}
