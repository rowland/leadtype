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
	w := buildCIDWidthArray(nil, 0)
	var buf bytes.Buffer
	w.write(&buf)
	got := strings.TrimSpace(buf.String())
	if got != "[]" {
		t.Errorf("expected empty array [], got %q", got)
	}
}

func TestBuildCIDWidthArray_SingleGlyph(t *testing.T) {
	w := buildCIDWidthArray(map[uint16]int{5: 600}, 0)
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
	w := buildCIDWidthArray(map[uint16]int{3: 500, 4: 600, 5: 700}, 0)
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
	w := buildCIDWidthArray(map[uint16]int{1: 500, 2: 500, 10: 700, 11: 700}, 0)
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

func TestBuildCIDWidthArray_OmitsDefaultWidth(t *testing.T) {
	// Glyphs 1,2,3 all have width 500 (the default); glyph 4 has width 700.
	// With defaultWidth=500, only glyph 4 should appear in /W.
	w := buildCIDWidthArray(map[uint16]int{1: 500, 2: 500, 3: 500, 4: 700}, 500)
	var buf bytes.Buffer
	w.write(&buf)
	got := buf.String()
	if strings.Contains(got, "500") {
		t.Errorf("expected default width 500 to be omitted from /W, got %q", got)
	}
	if !strings.Contains(got, "700") {
		t.Errorf("expected non-default width 700 in /W, got %q", got)
	}
}

func TestBuildCIDWidthArray_AllDefault_Empty(t *testing.T) {
	// When every glyph has the default width, /W should be empty.
	w := buildCIDWidthArray(map[uint16]int{1: 500, 2: 500}, 500)
	var buf bytes.Buffer
	w.write(&buf)
	got := strings.TrimSpace(buf.String())
	if got != "[]" {
		t.Errorf("expected empty /W when all widths match default, got %q", got)
	}
}

func TestMostCommonWidth(t *testing.T) {
	w := mostCommonWidth(map[uint16]int{1: 500, 2: 500, 3: 700, 4: 500})
	if w != 500 {
		t.Errorf("expected most common width 500, got %d", w)
	}
}

func TestMostCommonWidth_Empty(t *testing.T) {
	w := mostCommonWidth(nil)
	if w != 0 {
		t.Errorf("expected 0 for empty map, got %d", w)
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

// TestUnicodeMode_FontFile2 verifies that a /FontFile2 entry is present in the
// FontDescriptor and that the embedded data is a parseable TTF font.
func TestUnicodeMode_FontFile2(t *testing.T) {
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

	if !strings.Contains(pdf, "/FontFile2") {
		t.Errorf("expected /FontFile2 in FontDescriptor, got pdf containing:\n%s",
			extractSection(pdf, "/FontDescriptor", 400))
	}
	if !strings.Contains(pdf, "/Length1") {
		t.Errorf("expected /Length1 in FontFile2 stream dictionary")
	}
}

// TestUnicodeMode_SubsetTag verifies that the embedded font uses the
// "XXXXXX+FontName" subset tag format in all three name locations:
// FontDescriptor/FontName, CIDFont/BaseFont, and Type0/BaseFont.
func TestUnicodeMode_SubsetTag(t *testing.T) {
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

	// The tag must match /[A-Z]{6}\+Minimal/.
	// We check by counting occurrences of "+Minimal" — should appear 3 times
	// (FontDescriptor FontName, CIDFont BaseFont, Type0 BaseFont).
	count := strings.Count(pdf, "+Minimal")
	if count != 3 {
		t.Errorf("expected 3 occurrences of '+Minimal' subset tag, got %d\npdf excerpt:\n%s",
			count, extractSection(pdf, "/FontDescriptor", 600))
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

// ── Phase 3: no codepage splitting in unicode mode ────────────────────────────

// TestUnicodeMode_MultiScriptSingleTj verifies that a string containing both
// Latin and Greek characters (all present in the fixture font) is written as a
// single Tj operator in unicode mode, not split by codepage.
func TestUnicodeMode_MultiScriptSingleTj(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriterUnicode()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	pw.SetFont("Minimal", 12, options.Options{})
	pw.MoveTo(72, 720)
	// "Hello ΑΒΓΔ": ASCII (CP1252) + Greek capitals (ISO-8859-7).
	// Both ranges are in minimal.ttf, so splitByFont assigns a single leaf.
	// EachCodepage would split this into two callbacks; the Phase 3 path must not.
	pw.Print("Hello ΑΒΓΔ")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	tjCount := strings.Count(pdf, " Tj\n")
	if tjCount != 1 {
		t.Errorf("expected 1 Tj operator for single-font multi-script string, got %d\npdf excerpt:\n%s",
			tjCount, extractSection(pdf, "BT", 300))
	}
}

// TestUnicodeMode_MultiScriptTwoFonts verifies that when a string requires two
// fonts (e.g. one font for Latin, a fallback for extra glyphs) each font still
// produces its own Tj — the per-font boundary is preserved.
func TestUnicodeMode_MultiScriptTwoFonts(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriterUnicode()
	dw.AddFontSource(fc)

	// Use two fonts: same fixture for both (acts as primary and fallback).
	// splitByFont will keep it as one leaf, so still 1 Tj — this test
	// confirms that the per-leaf boundary is the unit of segmentation.
	pw := dw.NewPage()
	pw.SetFont("Minimal", 12, options.Options{})
	pw.MoveTo(72, 720)
	pw.Print("AB")
	pw.MoveTo(72, 700)
	pw.Print("ΑΒ")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	// Two separate Print calls each flush one leaf → two Tj operators.
	tjCount := strings.Count(pdf, " Tj\n")
	if tjCount != 2 {
		t.Errorf("expected 2 Tj operators for two Print calls, got %d", tjCount)
	}
}

// TestNonUnicodeMode_MultiScriptSplitsByCodepage verifies that in non-unicode
// mode a Latin+Greek string is still split into two Tj operators by EachCodepage,
// confirming the legacy path is untouched.
func TestNonUnicodeMode_MultiScriptSplitsByCodepage(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter() // non-unicode mode
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	pw.SetFont("Minimal", 12, options.Options{})
	pw.MoveTo(72, 720)
	pw.Print("Hello ΑΒΓΔ")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	// EachCodepage splits "Hello " (CP1252) and "ΑΒΓΔ" (ISO-8859-7) → 2 Tj.
	tjCount := strings.Count(pdf, " Tj\n")
	if tjCount != 2 {
		t.Errorf("expected 2 Tj operators in non-unicode mode for Latin+Greek, got %d", tjCount)
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
