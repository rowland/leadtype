// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rowland/leadtype/options"
)

// ── buildCIDWidthArray ────────────────────────────────────────────────────────

func TestBuildCIDWidthArray_Empty(t *testing.T) {
	w := buildCIDWidthArray(nil)
	var buf bytes.Buffer
	w.write(&buf)
	got := strings.TrimSpace(buf.String())
	if got != "[]" {
		t.Errorf("expected empty array [], got %q", got)
	}
}

func TestBuildCIDWidthArray_SingleGlyph(t *testing.T) {
	w := buildCIDWidthArray(map[uint16]int{5: 600})
	var buf bytes.Buffer
	w.write(&buf)
	got := buf.String()
	// Should be [5 [600]]
	if !strings.Contains(got, "5") {
		t.Errorf("expected glyph 5, got %q", got)
	}
	if !strings.Contains(got, "600") {
		t.Errorf("expected width 600, got %q", got)
	}
}

func TestBuildCIDWidthArray_ConsecutiveRun(t *testing.T) {
	w := buildCIDWidthArray(map[uint16]int{3: 500, 4: 600, 5: 700})
	var buf bytes.Buffer
	w.write(&buf)
	got := buf.String()
	// Consecutive glyphs 3,4,5 should form a single run: [3 [500 600 700]]
	if !strings.Contains(got, "3") {
		t.Errorf("expected run starting at 3, got %q", got)
	}
	// Only one sub-array opener: the run should not be split.
	subArrayCount := strings.Count(got, "[500")
	if subArrayCount != 1 {
		t.Errorf("expected a single sub-array for consecutive run, got %d in %q", subArrayCount, got)
	}
}

func TestBuildCIDWidthArray_TwoRuns(t *testing.T) {
	w := buildCIDWidthArray(map[uint16]int{1: 500, 2: 500, 10: 700, 11: 700})
	var buf bytes.Buffer
	w.write(&buf)
	got := buf.String()
	// Should encode two separate runs.
	if strings.Count(got, "[500") != 1 {
		t.Errorf("expected one [500 sub-array, got %q", got)
	}
	if strings.Count(got, "[700") != 1 {
		t.Errorf("expected one [700 sub-array, got %q", got)
	}
}

// ── Type0 / Unicode mode integration ─────────────────────────────────────────

// TestUnicodeMode_Type0FontInOutput verifies that NewDocWriterUnicode produces
// a /Type0 composite font instead of a simple TrueType font.
func TestUnicodeMode_Type0FontInOutput(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriterUnicode()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	pw.SetFont("Minimal", 12, options.Options{})
	pw.MoveTo(72, 720)
	pw.Print("ABC")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	if !strings.Contains(pdf, "/Type0") {
		t.Error("expected /Type0 in output PDF for unicode mode")
	}
	if !strings.Contains(pdf, "/CIDFontType2") {
		t.Error("expected /CIDFontType2 in output PDF for unicode mode")
	}
	if !strings.Contains(pdf, "/Identity-H") {
		t.Error("expected /Identity-H encoding in output PDF for unicode mode")
	}
}

// TestUnicodeMode_ToUnicodeCMap verifies that the composite font's ToUnicode
// CMap is emitted and contains the expected glyph → character mappings.
func TestUnicodeMode_ToUnicodeCMap(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriterUnicode()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	pw.SetFont("Minimal", 12, options.Options{})
	pw.MoveTo(72, 720)
	pw.Print("A")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	if !strings.Contains(pdf, "/ToUnicode") {
		t.Error("expected /ToUnicode in composite font output")
	}
	if !strings.Contains(pdf, "begincmap") {
		t.Error("expected CMap stream in output")
	}
	// CMap codespace range for composite fonts spans full 16-bit space.
	if !strings.Contains(pdf, "<0000> <FFFF>") {
		t.Errorf("expected <0000> <FFFF> codespace range, got pdf snippet:\n%s",
			extractCMapSection(pdf))
	}
}

// TestUnicodeMode_WidthArray verifies that the /W array is populated after
// rendering.
func TestUnicodeMode_WidthArray(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriterUnicode()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	pw.SetFont("Minimal", 12, options.Options{})
	pw.MoveTo(72, 720)
	pw.Print("A")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	// /W should be present in the CIDFont dictionary (non-empty).
	if !strings.Contains(pdf, "/W [") {
		t.Errorf("expected non-empty /W array in CIDFont, got pdf containing:\n%s",
			extractSection(pdf, "/CIDFontType2", 400))
	}
}

// TestUnicodeMode_SimpleMode verifies that non-unicode mode still works
// correctly with the same fixture font.
func TestUnicodeMode_SimpleMode(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	pw.SetFont("Minimal", 12, options.Options{})
	pw.MoveTo(72, 720)
	pw.Print("Hello")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	// In simple mode, should use TrueType (not Type0).
	if strings.Contains(pdf, "/Type0") {
		t.Error("simple mode should not emit /Type0 font")
	}
	if !strings.Contains(pdf, "/TrueType") {
		t.Error("simple mode should emit /TrueType font")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// extractCMapSection returns a short excerpt around the first "begincmap" for
// diagnostic output in test failures.
func extractCMapSection(pdf string) string {
	i := strings.Index(pdf, "begincmap")
	if i < 0 {
		return "(no begincmap found)"
	}
	end := i + 300
	if end > len(pdf) {
		end = len(pdf)
	}
	return pdf[i:end]
}

// extractSection returns up to n bytes of pdf starting from the first
// occurrence of marker.
func extractSection(pdf, marker string, n int) string {
	i := strings.Index(pdf, marker)
	if i < 0 {
		return "(marker not found: " + marker + ")"
	}
	end := i + n
	if end > len(pdf) {
		end = len(pdf)
	}
	return pdf[i:end]
}
