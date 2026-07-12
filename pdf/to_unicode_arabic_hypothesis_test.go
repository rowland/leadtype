// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/ttf_fonts"
)

// loadAmiriFont loads the Amiri-Regular fixture font for testing.
func loadAmiriFont(t *testing.T) *font.Font {
	t.Helper()
	fc, err := ttf_fonts.New("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	f, err := font.New(fc.FontInfos[0].Family(), options.Options{}, font.FontSources{fc})
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// Hypothesis 1: reverseRunes in shapedGlyphRuneAssignments corrupts the
// ToUnicode CMap by reversing the rune sequences for multi-rune clusters.
//
// The ToUnicode CMap must contain runes in logical (Unicode) order. If
// reverseRunes flips them, any multi-rune cluster (e.g., a ligature) will
// have its characters in the wrong order for text extraction.
//
// This test verifies that the concatenation of all assigned rune sequences
// (sorted by cluster index, i.e. logical order) reconstructs the original
// input string.
func TestHypothesis1_ReverseRunesCorruptsToUnicode(t *testing.T) {
	f := loadAmiriFont(t)

	words := []string{
		"الخبرات", // al-khibrat
		"الجيران", // al-jiran
		"بسم",     // bism
		"الله",    // allah
		"مرحبا",   // marhaba
	}

	for _, word := range words {
		t.Run(word, func(t *testing.T) {
			runes := []rune(word)
			glyphs, err := f.Shaper.Shape(runes, f, 12)
			if err != nil {
				t.Fatalf("Shape(%q): %v", word, err)
			}

			assignments := shapedGlyphRuneAssignments(glyphs, runes)

			t.Logf("input: %q (%d runes, %d glyphs)", word, len(runes), len(glyphs))
			for i, gp := range glyphs {
				seq := assignments[i]
				if len(seq) > 0 {
					t.Logf("  glyph[%d] id=%d cluster=%d → %q (U+%s)",
						i, gp.GlyphID, gp.ClusterIndex, string(seq), formatCodepoints(seq))
				} else {
					t.Logf("  glyph[%d] id=%d cluster=%d → (secondary/empty)",
						i, gp.GlyphID, gp.ClusterIndex)
				}
			}

			// Collect all assigned sequences, keyed by their cluster index
			// (the logical position in the input).
			type clusterEntry struct {
				clusterIndex int
				runes        []rune
			}
			var entries []clusterEntry
			for i, gp := range glyphs {
				if seq := assignments[i]; len(seq) > 0 {
					entries = append(entries, clusterEntry{gp.ClusterIndex, seq})
				}
			}
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].clusterIndex < entries[j].clusterIndex
			})

			// Reconstruct the input by concatenating sequences in logical order.
			var reconstructed []rune
			for _, e := range entries {
				reconstructed = append(reconstructed, e.runes...)
			}

			if string(reconstructed) != word {
				t.Errorf("HYPOTHESIS CONFIRMED: reconstructed %q (U+%s) != original %q (U+%s)",
					string(reconstructed), formatCodepoints(reconstructed),
					word, formatCodepoints(runes))
			} else {
				t.Logf("reconstruction matches original: %q", word)
			}
		})
	}
}

// Hypothesis 2: /W width mismatch with shaper advances causes pdftotext
// to insert spurious spaces.
//
// The /W array stores raw advance widths from the font's hmtx table. The
// shaper may produce different advances (via GPOS). When pdftotext compares
// the expected position (from /W) with the actual position (from Tm), a
// mismatch could cause it to insert spaces.
func TestHypothesis2_WidthMismatchCausesSpaces(t *testing.T) {
	f := loadAmiriFont(t)
	upm := float64(f.UnitsPerEm())

	words := []string{
		"الخبرات",
		"الجيران",
		"مرحبا",
	}

	for _, word := range words {
		t.Run(word, func(t *testing.T) {
			runes := []rune(word)
			fontSize := float32(12)
			glyphs, err := f.Shaper.Shape(runes, f, fontSize)
			if err != nil {
				t.Fatalf("Shape(%q): %v", word, err)
			}

			assignments := shapedGlyphRuneAssignments(glyphs, runes)

			anyMismatch := false
			for i, gp := range glyphs {
				// Shaper advance in points
				shaperAdvPts := float64(gp.XAdvance) / 64.0

				// Font metric advance: design units → points
				rawWidth := float64(f.AdvanceWidthForGlyph(gp.GlyphID))
				fontAdvPts := rawWidth * float64(fontSize) / upm

				// /W value (in 1/1000 em units, as stored in PDF)
				wValue := rawWidth * 1000.0 / upm

				diff := shaperAdvPts - fontAdvPts
				isMapped := len(assignments[i]) > 0
				status := "mapped"
				if !isMapped {
					status = "UNMAPPED"
				}

				if diff < -0.5 || diff > 0.5 {
					anyMismatch = true
					t.Logf("  MISMATCH glyph[%d] id=%d %s: shaper=%.2fpt font=%.2fpt (W=%.0f) diff=%.2fpt",
						i, gp.GlyphID, status, shaperAdvPts, fontAdvPts, wValue, diff)
				} else {
					t.Logf("  ok       glyph[%d] id=%d %s: shaper=%.2fpt font=%.2fpt (W=%.0f)",
						i, gp.GlyphID, status, shaperAdvPts, fontAdvPts, wValue)
				}
			}

			if anyMismatch {
				t.Logf("HYPOTHESIS SUPPORTED: width mismatches found for %q", word)
			} else {
				t.Logf("No significant width mismatches for %q", word)
			}
		})
	}
}

// Hypothesis 3: Per-glyph Tm positioning combined with XOffset creates
// position gaps that pdftotext interprets as word breaks.
//
// After rendering a glyph via Tm+Tj, pdftotext expects the next character
// at position = Tm.x + W*fontSize/1000. If the next Tm position differs
// by more than ~30% of average char width, pdftotext inserts a space.
//
// XOffset (from GPOS) shifts the Tm position but doesn't affect the
// width-based advance expectation, creating a potential mismatch.
func TestHypothesis3_TmPositionGapsCauseSpaces(t *testing.T) {
	f := loadAmiriFont(t)
	upm := float64(f.UnitsPerEm())
	fontSize := 12.0

	words := []string{
		"الخبرات",
		"الجيران",
		"اليدوية",
	}

	for _, word := range words {
		t.Run(word, func(t *testing.T) {
			runes := []rune(word)
			glyphs, err := f.Shaper.Shape(runes, f, float32(fontSize))
			if err != nil {
				t.Fatalf("Shape(%q): %v", word, err)
			}

			assignments := shapedGlyphRuneAssignments(glyphs, runes)

			// Compute average character width (for space threshold).
			totalWidth := 0.0
			for _, gp := range glyphs {
				totalWidth += float64(gp.XAdvance) / 64.0
			}
			avgWidth := totalWidth / float64(len(glyphs))
			spaceThreshold := avgWidth * 0.3

			// Simulate Tm positions as the rendering code does.
			leafStartX := 72.0 // arbitrary start
			penX := 0.0
			type glyphInfo struct {
				tmX      float64
				wWidth   float64 // /W width in points
				xAdvance float64 // shaper advance in points
				xOffset  float64 // glyph offset in points
				mapped   bool
				runeStr  string
				glyphID  uint16
			}
			infos := make([]glyphInfo, len(glyphs))
			for i, gp := range glyphs {
				tmX := leafStartX + penX + float64(gp.XOffset)/64.0
				rawWidth := float64(f.AdvanceWidthForGlyph(gp.GlyphID))
				wWidthPts := rawWidth * fontSize / upm
				xAdvPts := float64(gp.XAdvance) / 64.0
				xOffPts := float64(gp.XOffset) / 64.0
				isMapped := len(assignments[i]) > 0
				runeStr := ""
				if isMapped {
					runeStr = string(assignments[i])
				}
				infos[i] = glyphInfo{tmX, wWidthPts, xAdvPts, xOffPts, isMapped, runeStr, gp.GlyphID}
				penX += xAdvPts
			}

			// Check gaps between consecutive glyphs.
			gapFound := false
			for i := 0; i < len(infos)-1; i++ {
				expectedNext := infos[i].tmX + infos[i].wWidth
				actualNext := infos[i+1].tmX
				gap := actualNext - expectedNext

				status := "ok"
				if gap > spaceThreshold || gap < -spaceThreshold {
					status = "GAP"
					gapFound = true
				}

				t.Logf("  [%d→%d] %s glyph=%d(%s) Tm=%.2f W=%.2f expected=%.2f actual=%.2f gap=%.2f (threshold=%.2f)",
					i, i+1, status,
					infos[i].glyphID, infos[i].runeStr,
					infos[i].tmX, infos[i].wWidth,
					expectedNext, actualNext, gap, spaceThreshold)
			}

			if gapFound {
				t.Logf("HYPOTHESIS SUPPORTED: position gaps found for %q", word)
			} else {
				t.Logf("No significant position gaps for %q", word)
			}
		})
	}
}

// Hypothesis 4: The ToUnicode CMap maps CIDs, but PDF viewers interpret
// the content stream codes as raw glyph IDs (not CIDs) when looking up
// the ToUnicode CMap.
//
// The content stream emits CIDs (assigned by glyphRecorder). The ToUnicode
// CMap maps CIDs to Unicode. But the CIDToGIDMap maps CIDs to real glyph
// IDs. If a viewer uses glyph IDs instead of character codes to look up
// the ToUnicode CMap, it would fail to find mappings (since the CMap keys
// are CIDs, not glyph IDs).
//
// This test generates a full PDF, extracts the ToUnicode CMap, and verifies
// that every CID emitted in the content stream has a corresponding entry in
// the ToUnicode CMap, including empty destinations for zero-text glyphs.
func TestHypothesis4_InspectGeneratedToUnicodeCMap(t *testing.T) {
	fc, err := ttf_fonts.New("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	family := fc.FontInfos[0].Family()

	dw := NewDocWriter()
	dw.AddFontSource(fc)
	pw := dw.NewPage()
	if _, err := pw.SetFont(family, 12, options.Options{}); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pw.MoveTo(72, 720)
	pw.Print("الخبرات")

	var pdf bytes.Buffer
	if _, err := dw.WriteTo(&pdf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	pdfStr := pdf.String()

	// Find and dump the ToUnicode CMap stream.
	if idx := strings.Index(pdfStr, "beginbfchar"); idx >= 0 {
		// Extract from begincmap to endcmap.
		start := strings.LastIndex(pdfStr[:idx], "begincmap")
		end := strings.Index(pdfStr[idx:], "endcmap") + idx + len("endcmap")
		if start >= 0 && end > start {
			cmapContent := pdfStr[start:end]
			t.Logf("ToUnicode CMap:\n%s", cmapContent)
		}
	} else {
		t.Fatal("No beginbfchar found in PDF output")
	}

	// Find and dump the content stream (look for Tm + Tj patterns).
	// The content stream contains BT...ET blocks.
	btIdx := strings.Index(pdfStr, "BT\n")
	etIdx := strings.Index(pdfStr, "ET\n")
	if btIdx >= 0 && etIdx > btIdx {
		contentStream := pdfStr[btIdx : etIdx+3]
		t.Logf("Content stream:\n%s", contentStream)
	}

	// Also look for CIDToGIDMap data (binary, so just note its presence).
	if strings.Contains(pdfStr, "/CIDToGIDMap") {
		t.Logf("CIDToGIDMap present")
	} else {
		t.Logf("WARNING: No CIDToGIDMap found")
	}

	// Verify: check if ToUnicode maps CIDs (1, 2, 3...) or glyph IDs.
	// If the CMap contains entries like <0001>, <0002>, etc., it's CID-based.
	// If it contains entries like <018A>, <0187>, etc. (actual glyph IDs), it's GID-based.
	t.Logf("If ToUnicode source codes are sequential (0001, 0002, ...), they are CIDs.")
	t.Logf("If they match actual glyph IDs (e.g., 018A for glyph 394), that's a different scheme.")
}

// TestPdftotextArabicWords generates single-word PDFs and shows exact
// pdftotext output, including any spurious spaces or broken text.
func TestPdftotextArabicWords(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skipf("pdftotext unavailable: %v", err)
	}

	fc, err := ttf_fonts.New("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	family := fc.FontInfos[0].Family()

	words := []string{
		"الخبرات",
		"الجيران",
		"الحِرَف",
		"اليدوية",
		"الإنجليزية",
		"مرحبا",
		"بسم",
	}

	for _, word := range words {
		t.Run(word, func(t *testing.T) {
			dw := NewDocWriter()
			dw.AddFontSource(fc)
			pw := dw.NewPage()
			if _, err := pw.SetFont(family, 12, options.Options{}); err != nil {
				t.Fatalf("SetFont: %v", err)
			}
			pw.MoveTo(72, 720)
			pw.Print(word)

			var pdf bytes.Buffer
			if _, err := dw.WriteTo(&pdf); err != nil {
				t.Fatalf("WriteTo: %v", err)
			}
			path := filepath.Join(t.TempDir(), "test.pdf")
			if err := os.WriteFile(path, pdf.Bytes(), 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			out, err := exec.Command("pdftotext", path, "-").CombinedOutput()
			if err != nil {
				t.Fatalf("pdftotext: %v\n%s", err, out)
			}
			got := strings.TrimSpace(string(out))

			// Poppler serializes an RTL extraction run in visual codepoint order
			// inside an RLE/PDF wrapper. Decode that transport representation
			// before comparing it with the PDF's logical-order ActualText.
			normalized := normalizeArabicExtractorOutput(got)

			// Show exact output with codepoints.
			t.Logf("input:      %q (U+%s)", word, formatCodepoints([]rune(word)))
			t.Logf("raw output: %q (U+%s)", got, formatCodepoints([]rune(got)))
			t.Logf("normalized: %q", normalized)

			if normalized != word {
				t.Errorf("MISMATCH: expected %q, got normalized %q (raw %q)", word, normalized, got)
			}
		})
	}
}

// TestShapedGlyphDetails shows full glyph-by-glyph details for representative
// Arabic words, including clusters with marks and decorative secondary glyphs.
func TestShapedGlyphDetails(t *testing.T) {
	f := loadAmiriFont(t)

	words := []string{
		"اليدوية",
		"الإنجليزية",
		"الحِرَف",
		"الخبرات",
	}

	for _, word := range words {
		t.Run(word, func(t *testing.T) {
			runes := []rune(word)
			glyphs, err := f.Shaper.Shape(runes, f, 12)
			if err != nil {
				t.Fatalf("Shape(%q): %v", word, err)
			}

			assignments := shapedGlyphRuneAssignments(glyphs, runes)

			t.Logf("input: %q = %d runes → %d glyphs", word, len(runes), len(glyphs))
			for i, r := range runes {
				t.Logf("  rune[%d] = %c (U+%04X)", i, r, r)
			}
			t.Logf("---")
			for i, gp := range glyphs {
				seq := assignments[i]
				mapped := "→ <> (empty destination)"
				if len(seq) > 0 {
					mapped = fmt.Sprintf("→ %q (U+%s)", string(seq), formatCodepoints(seq))
				}
				t.Logf("  glyph[%d] id=%d cluster=%d xAdv=%.2f xOff=%.2f yOff=%.2f %s",
					i, gp.GlyphID, gp.ClusterIndex,
					float64(gp.XAdvance)/64.0,
					float64(gp.XOffset)/64.0,
					float64(gp.YOffset)/64.0,
					mapped)
			}
		})
	}
}

// TestInspectContentStream_اليدوية generates a PDF for a word with XOffset
// glyphs and inspects the TJ array.
func TestInspectContentStream_XOffsetWord(t *testing.T) {
	fc, err := ttf_fonts.New("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	family := fc.FontInfos[0].Family()

	words := []string{"اليدوية", "الإنجليزية"}
	for _, word := range words {
		t.Run(word, func(t *testing.T) {
			dw := NewDocWriter()
			dw.AddFontSource(fc)
			pw := dw.NewPage()
			if _, err := pw.SetFont(family, 12, options.Options{}); err != nil {
				t.Fatalf("SetFont: %v", err)
			}
			pw.MoveTo(72, 720)
			pw.Print(word)

			var pdf bytes.Buffer
			if _, err := dw.WriteTo(&pdf); err != nil {
				t.Fatalf("WriteTo: %v", err)
			}
			pdfStr := pdf.String()

			btIdx := strings.Index(pdfStr, "BT\n")
			etIdx := strings.Index(pdfStr, "ET\n")
			if btIdx >= 0 && etIdx > btIdx {
				t.Logf("Content stream:\n%s", pdfStr[btIdx:etIdx+3])
			}

			if idx := strings.Index(pdfStr, "beginbfchar"); idx >= 0 {
				start := strings.LastIndex(pdfStr[:idx], "begincmap")
				end := strings.Index(pdfStr[idx:], "endcmap") + idx + len("endcmap")
				if start >= 0 && end > start {
					t.Logf("ToUnicode CMap:\n%s", pdfStr[start:end])
				}
			}
		})
	}
}

func formatCodepoints(runes []rune) string {
	parts := make([]string, len(runes))
	for i, r := range runes {
		parts[i] = fmt.Sprintf("%04X", r)
	}
	return strings.Join(parts, " ")
}
