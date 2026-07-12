// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/shaping"
	"github.com/rowland/leadtype/ttf_fonts"
)

type fullCFFFallbackSource struct {
	metrics *fullCFFFallbackMetrics
}

func (s *fullCFFFallbackSource) Select(family, weight, style string, ranges []string) (font.FontMetrics, error) {
	if family != "Full CFF" {
		return nil, errors.New("not found")
	}
	return s.metrics, nil
}

func (s *fullCFFFallbackSource) SubType() string {
	return "TrueType"
}

type fullCFFFallbackMetrics struct {
	bytes []byte
}

func (m *fullCFFFallbackMetrics) AdvanceWidth(r rune) (int, bool) {
	if m.GlyphIndex(r) == 0 {
		return 0, true
	}
	return 512, false
}

func (m *fullCFFFallbackMetrics) AdvanceWidthForGlyph(glyphID uint16) int {
	if glyphID == 0 {
		return 0
	}
	return 512
}

func (m *fullCFFFallbackMetrics) Ascent() int             { return 800 }
func (m *fullCFFFallbackMetrics) BoundingBox() [4]int     { return [4]int{0, -200, 512, 800} }
func (m *fullCFFFallbackMetrics) Bytes() []byte           { return m.bytes }
func (m *fullCFFFallbackMetrics) CapHeight() int          { return 700 }
func (m *fullCFFFallbackMetrics) Copyright() string       { return "" }
func (m *fullCFFFallbackMetrics) Descent() int            { return -200 }
func (m *fullCFFFallbackMetrics) Family() string          { return "Full CFF" }
func (m *fullCFFFallbackMetrics) Filename() string        { return "full-cff.otf" }
func (m *fullCFFFallbackMetrics) Flags() uint32           { return 4 }
func (m *fullCFFFallbackMetrics) FontKey() string         { return "full-cff" }
func (m *fullCFFFallbackMetrics) FullName() string        { return "Full CFF" }
func (m *fullCFFFallbackMetrics) ItalicAngle() float64    { return 0 }
func (m *fullCFFFallbackMetrics) Leading() int            { return 0 }
func (m *fullCFFFallbackMetrics) LineGap() int            { return 0 }
func (m *fullCFFFallbackMetrics) NumGlyphs() int          { return 64 }
func (m *fullCFFFallbackMetrics) OutlineKind() string     { return "CFF" }
func (m *fullCFFFallbackMetrics) PostScriptName() string  { return "FullCFF" }
func (m *fullCFFFallbackMetrics) StemV() int              { return 80 }
func (m *fullCFFFallbackMetrics) StrikeoutPosition() int  { return 300 }
func (m *fullCFFFallbackMetrics) StrikeoutThickness() int { return 50 }
func (m *fullCFFFallbackMetrics) Style() string           { return "Regular" }
func (m *fullCFFFallbackMetrics) Subset([]uint16) ([]byte, error) {
	return nil, errors.New("subset unsupported")
}
func (m *fullCFFFallbackMetrics) SupportsArabic() bool    { return false }
func (m *fullCFFFallbackMetrics) UnderlinePosition() int  { return -100 }
func (m *fullCFFFallbackMetrics) UnderlineThickness() int { return 50 }
func (m *fullCFFFallbackMetrics) UnitsPerEm() int         { return 1000 }
func (m *fullCFFFallbackMetrics) Version() string         { return "" }
func (m *fullCFFFallbackMetrics) XHeight() int            { return 500 }

func (m *fullCFFFallbackMetrics) GlyphIndex(r rune) uint16 {
	switch r {
	case '中':
		return 10
	case '文':
		return 11
	default:
		return 0
	}
}

// ── buildCIDWidthArray ────────────────────────────────────────────────────────

func TestCIDSystemInfo_DefaultSerializationRemainsStable(t *testing.T) {
	var buf bytes.Buffer
	(&cidSystemInfo{registry: "Adobe", ordering: "Identity", supplement: 0}).write(&buf)
	const want = "<< /Registry (Adobe) /Ordering (Identity) /Supplement 0 >> "
	if got := buf.String(); got != want {
		t.Fatalf("CIDSystemInfo = %q, want %q", got, want)
	}
}

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

func TestMostCommonWidth_TieBreaksDeterministically(t *testing.T) {
	w := mostCommonWidth(map[uint16]int{1: 556, 2: 277, 3: 556, 4: 277})
	if w != 277 {
		t.Errorf("expected deterministic tied width 277, got %d", w)
	}
}

func TestMostCommonWidth_Empty(t *testing.T) {
	w := mostCommonWidth(nil)
	if w != 0 {
		t.Errorf("expected 0 for empty map, got %d", w)
	}
}

// ── Type0 / Unicode mode integration ─────────────────────────────────────────

func systemFontSourceSupportingRune(t *testing.T, r rune, preferredFamilies ...string) (font.FontSource, string) {
	t.Helper()

	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Skipf("system fonts unavailable: %v", err)
	}

	tryFamily := func(family string) bool {
		f, err := font.New(family, options.Options{}, font.FontSources{fc})
		if err != nil {
			return false
		}
		return f.GlyphIndex(r) != 0
	}

	for _, family := range preferredFamilies {
		if tryFamily(family) {
			return fc, family
		}
	}

	seen := make(map[string]bool, len(fc.FontInfos))
	families := make([]string, 0, len(fc.FontInfos))
	for _, fi := range fc.FontInfos {
		family := fi.Family()
		if family == "" || seen[family] {
			continue
		}
		seen[family] = true
		families = append(families, family)
	}
	sort.Strings(families)
	for _, family := range families {
		if tryFamily(family) {
			return fc, family
		}
	}

	t.Skipf("no system font found with glyph for %U", r)
	return nil, ""
}

// TestUnicodeMode_Type0FontInOutput verifies that NewDocWriterUnicode produces
// a /Type0 composite font instead of a simple TrueType font.
func TestUnicodeMode_Type0FontInOutput(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
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

	dw := NewDocWriter()
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

func TestUnicodeMode_ToUnicodeCMap_Compressed(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
	dw.AddFontSource(fc)
	dw.CompressToUnicode(true)

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
	if !strings.Contains(pdf, "/Filter /FlateDecode") {
		t.Fatalf("expected compressed ToUnicode stream, got pdf excerpt:\n%s", extractSection(pdf, "/ToUnicode", 400))
	}
}

// TestUnicodeMode_WidthArray verifies that the /W array is populated after
// rendering.
func TestUnicodeMode_WidthArray(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
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

	dw := NewDocWriter()
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

func TestUnicodeMode_FontFile2_Compressed(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
	dw.AddFontSource(fc)
	dw.CompressEmbeddedFonts(true)

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
	if !strings.Contains(pdf, "/Filter /FlateDecode") {
		t.Fatalf("expected compressed FontFile2 stream, got pdf excerpt:\n%s", extractSection(pdf, "/FontFile2", 500))
	}
}

func TestUnicodeMode_CFFOpenTypeFontFile3(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal-cff.otf")

	render := func() string {
		dw := NewDocWriter()
		dw.AddFontSource(fc)

		pw := dw.NewPage()
		pw.SetFont("MinimalCFF", 12, options.Options{})
		pw.MoveTo(72, 720)
		pw.Print("MinimalCFF")

		var buf bytes.Buffer
		dw.WriteTo(&buf)
		return buf.String()
	}

	pdf := render()
	if !strings.Contains(pdf, "/Type0") {
		t.Fatal("expected /Type0 in CFF OTF output")
	}
	if !strings.Contains(pdf, "/CIDFontType0") {
		t.Fatalf("expected /CIDFontType0 in CFF OTF output, got excerpt:\n%s", extractSection(pdf, "/FontDescriptor", 800))
	}
	if strings.Contains(pdf, "/CIDFontType2") {
		t.Fatal("did not expect /CIDFontType2 for CFF OTF output")
	}
	if !strings.Contains(pdf, "/FontFile3") {
		t.Fatalf("expected /FontFile3 in CFF OTF output, got excerpt:\n%s", extractSection(pdf, "/FontDescriptor", 800))
	}
	if !strings.Contains(pdf, "/Subtype /OpenType") {
		t.Fatalf("expected FontFile3 /Subtype /OpenType, got excerpt:\n%s", extractSection(pdf, "/FontFile3", 800))
	}
	if strings.Contains(pdf, "/FontFile2") {
		t.Fatal("did not expect /FontFile2 for CFF OTF output")
	}
	if strings.Contains(pdf, "/CIDToGIDMap") {
		t.Fatal("did not expect /CIDToGIDMap for CFF OTF output")
	}
	if !strings.Contains(pdf, "/ToUnicode") {
		t.Fatal("expected /ToUnicode in CFF OTF output")
	}
	if !strings.Contains(pdf, "/W [") {
		t.Fatal("expected non-empty /W in CFF OTF output")
	}
	if second := render(); second != pdf {
		t.Fatal("expected deterministic CFF OTF PDF output")
	}
}

func TestUnicodeMode_CIDKeyedCFFUsesCustomEncodingAndNativeCFF(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal-cid-cff.otf")
	dw := NewDocWriter()
	dw.AddFontSource(fc)
	pw := dw.NewPage()
	pw.SetFont("MinimalCIDCFF", 12, options.Options{})
	pw.MoveTo(72, 720)
	pw.Print("A中")

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdf := buf.String()
	for _, want := range []string{
		"/Subtype /CIDFontType0",
		"/Subtype /CIDFontType0C",
		"/CMapName /LeadType-CID-Encoding",
		"begincidchar",
		"/CIDSet",
		"/ToUnicode",
	} {
		if !strings.Contains(pdf, want) {
			t.Fatalf("CID-keyed CFF PDF missing %q", want)
		}
	}
	if strings.Contains(pdf, "/Encoding /Identity-H") {
		t.Fatal("CID-keyed CFF Type0 font unexpectedly uses Identity-H")
	}
}

func TestUnicodeMode_CFFFullFontFallbackWhenSubsetFails(t *testing.T) {
	fullBytes, err := os.ReadFile("../ttf/testdata/minimal-cff.otf")
	if err != nil {
		t.Fatal(err)
	}
	dw := NewDocWriter()
	dw.AddFontSource(&fullCFFFallbackSource{metrics: &fullCFFFallbackMetrics{bytes: fullBytes}})

	pw := dw.NewPage()
	pw.SetFont("Full CFF", 12, options.Options{})
	pw.MoveTo(72, 720)
	pw.Print("中文")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	if !strings.Contains(pdf, "/CIDFontType0") {
		t.Fatalf("expected /CIDFontType0 for CFF full-font fallback, got excerpt:\n%s", extractSection(pdf, "/FontDescriptor", 800))
	}
	if !strings.Contains(pdf, "/FontFile3") {
		t.Fatalf("expected full CFF fallback to embed /FontFile3, got excerpt:\n%s", extractSection(pdf, "/FontDescriptor", 800))
	}
	if !strings.Contains(pdf, "/Subtype /OpenType") {
		t.Fatalf("expected FontFile3 /Subtype /OpenType, got excerpt:\n%s", extractSection(pdf, "/FontFile3", 800))
	}
	if strings.Contains(pdf, "/CIDToGIDMap") {
		t.Fatal("did not expect CIDToGIDMap for CFF full-font fallback")
	}
	if strings.Contains(pdf, "+FullCFF") {
		t.Fatal("did not expect subset tag when embedding the full fallback font")
	}
}

// TestUnicodeMode_SubsetTag verifies that the embedded font uses the
// "XXXXXX+FontName" subset tag format in all three name locations:
// FontDescriptor/FontName, CIDFont/BaseFont, and Type0/BaseFont.
func TestUnicodeMode_SubsetTag(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
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

func TestUnicodeMode_SubsetTag_IsDeterministic(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	render := func() string {
		dw := NewDocWriter()
		dw.AddFontSource(fc)

		pw := dw.NewPage()
		pw.SetFont("Minimal", 12, options.Options{})
		pw.MoveTo(72, 720)
		pw.Print("ABC")

		var buf bytes.Buffer
		dw.WriteTo(&buf)
		return buf.String()
	}

	first := render()
	second := render()

	if first != second {
		t.Fatalf("expected repeated renders to be byte-identical")
	}
}

// ── Phase 3: no codepage splitting in unicode mode ────────────────────────────

// TestUnicodeMode_MultiScriptSingleTj verifies that a string containing both
// Latin and Greek characters (all present in the fixture font) is written as a
// single Tj operator in unicode mode, not split by codepage.
func TestUnicodeMode_MultiScriptSingleTj(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
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

func TestUnicodeMode_RenderEmoji_SystemFont(t *testing.T) {
	fc, family := systemFontSourceSupportingRune(t, '😀', "Apple Color Emoji", "Apple Color Emoji UI")

	dw := NewDocWriter()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	fonts, err := pw.SetFont(family, 12, options.Options{})
	if err != nil {
		t.Fatalf("SetFont(%q): %v", family, err)
	}
	if len(fonts) == 0 || fonts[0].GlyphIndex('😀') == 0 {
		t.Fatalf("font %q does not expose a glyph for U+1F600 after selection", family)
	}
	pw.MoveTo(72, 720)
	pw.Print("😀")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	if !strings.Contains(pdf, "/Type0") {
		t.Fatalf("expected /Type0 in emoji PDF, got excerpt:\n%s", extractSection(pdf, "/Type0", 300))
	}
	if !strings.Contains(pdf, "/ToUnicode") {
		t.Fatalf("expected /ToUnicode in emoji PDF, got excerpt:\n%s", extractSection(pdf, "/ToUnicode", 300))
	}
	if !strings.Contains(pdf, "<D83DDE00>") {
		t.Fatalf("expected emoji ToUnicode mapping via UTF-16 surrogate pair, got CMap:\n%s", extractCMapSection(pdf))
	}
	if !strings.Contains(pdf, "/CIDToGIDMap") {
		t.Fatalf("expected composite font CIDToGIDMap in emoji PDF, got excerpt:\n%s", extractSection(pdf, "/CIDFontType2", 400))
	}
	if got := strings.Count(extractSection(pdf, "BT", 300), " Tj\n"); got != 1 {
		t.Fatalf("expected a single glyph show operation for emoji render, got %d", got)
	}
}

func TestUnicodeMode_UsesHexStringsForCompositeText(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	pw.SetFont("Minimal", 12, options.Options{})
	pw.MoveTo(72, 720)
	pw.Print("Hello ΑΒΓΔ")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()
	section := extractSection(pdf, "BT", 200)

	if strings.Contains(section, "(\x00") {
		t.Fatalf("expected composite text to avoid binary literal strings, got section:\n%s", section)
	}
	if !strings.Contains(section, "<") || !strings.Contains(section, " Tj\n") {
		t.Fatalf("expected composite text to be emitted as hex Tj, got section:\n%s", section)
	}
}

// TestUnicodeMode_MultiScriptTwoFonts verifies that when a string requires two
// fonts (e.g. one font for Latin, a fallback for extra glyphs) each font still
// produces its own Tj — the per-font boundary is preserved.
func TestUnicodeMode_MultiScriptTwoFonts(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
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

type offsetShaper struct {
	glyphs []shaping.GlyphPosition
}

func (s offsetShaper) Shape(_ []rune, _ shaping.FontReader, _ float32) ([]shaping.GlyphPosition, error) {
	return s.glyphs, nil
}

type errorShaper struct {
	err error
}

func (s errorShaper) Shape(_ []rune, _ shaping.FontReader, _ float32) ([]shaping.GlyphPosition, error) {
	return nil, s.err
}

type panicShaper struct{}

func (panicShaper) Shape(_ []rune, _ shaping.FontReader, _ float32) ([]shaping.GlyphPosition, error) {
	panic("shaper should not be called")
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	prev := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = prev
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestUnicodeMode_ShapedGlyphOffsetsWithVaryingYOffsetUsePerGlyphPositioning(t *testing.T) {
	fc, err := ttf_fonts.New("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	family := fc.FontInfos[0].Family()

	dw := NewDocWriter()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	fonts, err := pw.SetFont(family, 12, options.Options{})
	if err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	if len(fonts) == 0 {
		t.Fatal("SetFont returned no fonts")
	}
	fonts[0].Shaper = offsetShaper{glyphs: []shaping.GlyphPosition{
		{GlyphID: fonts[0].GlyphIndex('م'), XAdvance: 8 * 64, XOffset: 2 * 64, YOffset: 1 * 64, ClusterIndex: 0},
		{GlyphID: fonts[0].GlyphIndex('ر'), XAdvance: 8 * 64, XOffset: 0, YOffset: 0, ClusterIndex: 1},
	}}

	pw.MoveTo(72, 720)
	pw.Print("مر")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	if got := strings.Count(pdf, " Tm\n"); got < 3 {
		t.Fatalf("expected shaped text to emit per-glyph positioning matrices, got %d\npdf excerpt:\n%s", got, extractSection(pdf, "BT", 400))
	}
	if strings.Contains(pdf, "<0001><0002>") {
		t.Fatalf("expected shaped glyphs to be emitted individually, got pdf excerpt:\n%s", extractSection(pdf, "BT", 400))
	}
	if !strings.Contains(pdf, "<") || !strings.Contains(pdf, " Tj\n") {
		t.Fatalf("expected individual shaped glyph hex output, got pdf excerpt:\n%s", extractSection(pdf, "BT", 400))
	}
}

func TestUnicodeMode_ShapedGlyphOffsetsWithZeroYOffsetUseTJAndRawWidths(t *testing.T) {
	fc, err := ttf_fonts.New("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	family := fc.FontInfos[0].Family()

	dw := NewDocWriter()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	fonts, err := pw.SetFont(family, 12, options.Options{})
	if err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	if len(fonts) == 0 {
		t.Fatal("SetFont returned no fonts")
	}

	gidMeem := fonts[0].GlyphIndex('م')
	gidAlef := fonts[0].GlyphIndex('ا')
	fonts[0].Shaper = offsetShaper{glyphs: []shaping.GlyphPosition{
		{GlyphID: gidMeem, XAdvance: 9 * 64, XOffset: 0, YOffset: 0, ClusterIndex: 0},
		{GlyphID: gidAlef, XAdvance: 9 * 64, XOffset: 2 * 64, YOffset: 0, ClusterIndex: 1},
	}}

	pw.MoveTo(72, 720)
	pw.Print("ما")

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	pdf := buf.String()
	section := extractSection(pdf, "BT", 500)

	if strings.Count(section, " TJ\n") != 1 {
		t.Fatalf("expected a single TJ operator for zero-YOffset shaped text, got excerpt:\n%s", section)
	}
	if strings.Contains(section, " Tj\n") {
		t.Fatalf("expected zero-YOffset shaped text to avoid per-glyph Tj operators, got excerpt:\n%s", section)
	}
	if !regexp.MustCompile(`>[[:space:]]*-?[0-9]`).MatchString(section) {
		t.Fatalf("expected TJ array to contain numeric positioning adjustments, got excerpt:\n%s", section)
	}

	psName := fonts[0].PostScriptName()
	cidFont := dw.cidFonts[psName]
	defaultWidth, ok := cidFont.dict["DW"].(integer)
	if !ok {
		t.Fatalf("expected integer /DW entry, got %T", cidFont.dict["DW"])
	}
	var widthBuf bytes.Buffer
	cidFont.dict["W"].write(&widthBuf)
	widths := widthBuf.String()

	rawMeem := fonts[0].AdvanceWidthForGlyph(gidMeem)
	rawAlef := fonts[0].AdvanceWidthForGlyph(gidAlef)
	if upm := fonts[0].UnitsPerEm(); upm > 0 {
		rawMeem = rawMeem * 1000 / upm
		rawAlef = rawAlef * 1000 / upm
	}
	expectedDefault := rawAlef
	if rawMeem < rawAlef {
		expectedDefault = rawMeem
	}
	if int(defaultWidth) != expectedDefault {
		t.Fatalf("default width = %d, want raw metric width %d", defaultWidth, expectedDefault)
	}
	if rawMeem != expectedDefault && !strings.Contains(widths, strconv.Itoa(rawMeem)) {
		t.Fatalf("expected raw meem width %d in /W array, got %q", rawMeem, widths)
	}
	if rawAlef != expectedDefault && !strings.Contains(widths, strconv.Itoa(rawAlef)) {
		t.Fatalf("expected raw alef width %d in /W array, got %q", rawAlef, widths)
	}

	adjustedMeem := 916 // (9pt advance + 2pt next XOffset) scaled to 1000/fontSize.
	if adjustedMeem != rawMeem && adjustedMeem != rawAlef && strings.Contains(widths, strconv.Itoa(adjustedMeem)) {
		t.Fatalf("expected /W array to avoid synthetic positioned width %d, got %q", adjustedMeem, widths)
	}
}

func TestShapedGlyphRuneAssignments_DistributesRunesAcrossClusterGlyphs(t *testing.T) {
	assignments := shapedGlyphRuneAssignments([]shaping.GlyphPosition{
		{GlyphID: 10, ClusterIndex: 2},
		{GlyphID: 20, ClusterIndex: 0},
		{GlyphID: 21, ClusterIndex: 0},
	}, []rune("صُم"))

	if got := string(assignments[0]); got != "م" {
		t.Fatalf("glyph 0 assignment = %q, want %q", got, "م")
	}
	// With per-glyph distribution, cluster 0 ("صُ") splits across 2 glyphs
	// in reversed visual order: glyph[1] → damma, glyph[2] → saad.
	if got := string(assignments[1]); got != "ُ" {
		t.Fatalf("glyph 1 assignment = %q, want %q", got, "ُ")
	}
	if got := string(assignments[2]); got != "ص" {
		t.Fatalf("glyph 2 assignment = %q, want %q", got, "ص")
	}
}

func TestUnicodeMode_ToUnicodeCMap_ShapedClusterMapsOnceAndKeepsAllGlyphs(t *testing.T) {
	fc, err := ttf_fonts.New("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	family := fc.FontInfos[0].Family()

	dw := NewDocWriter()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	fonts, err := pw.SetFont(family, 12, options.Options{})
	if err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	if len(fonts) == 0 {
		t.Fatal("SetFont returned no fonts")
	}
	gidM := fonts[0].GlyphIndex('م')
	fonts[0].Shaper = offsetShaper{glyphs: []shaping.GlyphPosition{
		{GlyphID: gidM, XAdvance: 8 * 64, ClusterIndex: 2},
		{GlyphID: fonts[0].GlyphIndex('ُ'), XAdvance: 0, ClusterIndex: 0},
		{GlyphID: fonts[0].GlyphIndex('ص'), XAdvance: 8 * 64, ClusterIndex: 0},
	}}

	pw.MoveTo(72, 720)
	pw.Print("صُم")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	// With per-glyph rune distribution, the cluster "صُ" produces
	// individual ToUnicode entries for each glyph rather than one combined entry.
	if strings.Count(pdf, "<0635>") != 1 {
		t.Fatalf("expected saad mapping to appear once, got pdf excerpt:\n%s", extractCMapSection(pdf))
	}
	if strings.Count(pdf, "<064F>") != 1 {
		t.Fatalf("expected damma mapping to appear once, got pdf excerpt:\n%s", extractCMapSection(pdf))
	}
	if strings.Count(pdf, "<0645>") != 1 {
		t.Fatalf("expected standalone meem mapping to appear once, got pdf excerpt:\n%s", extractCMapSection(pdf))
	}
	if strings.Count(pdf, " beginbfchar\n") == 0 {
		t.Fatalf("expected ToUnicode CMap entries, got pdf excerpt:\n%s", extractCMapSection(pdf))
	}
	if !strings.Contains(pdf, "/CIDToGIDMap") {
		t.Fatalf("expected composite font to include CIDToGIDMap, got pdf excerpt:\n%s", extractSection(pdf, "/CIDFontType2", 400))
	}
}

func TestUnicodeMode_ShapingFailureWarnsDuringEmission(t *testing.T) {
	fc, err := ttf_fonts.New("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	family := fc.FontInfos[0].Family()

	dw := NewDocWriter()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	fonts, err := pw.SetFont(family, 12, options.Options{})
	if err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	if len(fonts) == 0 {
		t.Fatal("SetFont returned no fonts")
	}
	fonts[0].Shaper = errorShaper{err: errors.New("boom")}

	pw.MoveTo(72, 720)
	pw.Print("مرحبا")

	warnings := captureStderr(t, func() {
		var buf bytes.Buffer
		dw.WriteTo(&buf)
	})

	if !strings.Contains(warnings, "shaping failed during PDF emission") {
		t.Fatalf("expected PDF emission shaping warning, got %q", warnings)
	}
}

func TestFontsSupportArabicShaping_FalseForLatinOnlyChain(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")
	latin, err := font.New("Minimal", options.Options{}, font.FontSources{fc})
	if err != nil {
		t.Fatal(err)
	}
	if fontsSupportArabicShaping([]*font.Font{latin}) {
		t.Fatal("expected Latin-only chain to report no Arabic shaping support")
	}
}

func TestFontsSupportArabicShaping_TrueForMixedChain(t *testing.T) {
	latinSource := testFontSource(t, "../ttf/testdata/minimal.ttf")
	latin, err := font.New("Minimal", options.Options{}, font.FontSources{latinSource})
	if err != nil {
		t.Fatal(err)
	}

	arabicSource, err := ttf_fonts.New(filepath.Join("..", "shaping", "testdata", "Amiri-Regular.ttf"))
	if err != nil || len(arabicSource.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	arabic, err := font.New(arabicSource.FontInfos[0].Family(), options.Options{}, font.FontSources{arabicSource})
	if err != nil {
		t.Fatal(err)
	}

	if !fontsSupportArabicShaping([]*font.Font{latin, arabic}) {
		t.Fatal("expected mixed chain to report Arabic shaping support")
	}
}

func TestUnicodeMode_ArabicTextWithLatinOnlyFontSkipsShaper(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	fonts, err := pw.SetFont("Minimal", 12, options.Options{})
	if err != nil {
		t.Fatal(err)
	}
	fonts[0].Shaper = panicShaper{}
	pw.MoveTo(72, 720)
	pw.Print("مرحبا")

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
}

func TestUnicodeMode_WordSpacingUsesPositioningOperators(t *testing.T) {
	fc := testFontSource(t, "../ttf/testdata/minimal.ttf")

	dw := NewDocWriter()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	pw.SetFont("Minimal", 12, options.Options{})
	pw.MoveTo(72, 720)

	rt, err := pw.richTextForString("A A")
	if err != nil {
		t.Fatalf("richTextForString: %v", err)
	}
	rt.WordSpacing = 12
	pw.PrintRichText(rt)

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	if strings.Contains(pdf, "12 Tw\n") {
		t.Fatalf("expected composite-font word spacing to use explicit positioning, got pdf excerpt:\n%s", extractSection(pdf, "BT", 400))
	}
	if got := strings.Count(pdf, " Tm\n"); got < 4 {
		t.Fatalf("expected positioned glyph output for word-spaced composite text, got %d matrices\npdf excerpt:\n%s", got, extractSection(pdf, "BT", 400))
	}
	if got := strings.Count(pdf, " Tj\n"); got < 3 {
		t.Fatalf("expected each glyph to be emitted separately, got %d Tj operators\npdf excerpt:\n%s", got, extractSection(pdf, "BT", 400))
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
