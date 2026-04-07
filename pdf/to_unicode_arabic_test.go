// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/ttf_fonts"
)

func TestGlyphRecorder_AssignsDistinctCIDsForGlyphReuse(t *testing.T) {
	fc, err := ttf_fonts.New("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	f, err := font.New(fc.FontInfos[0].Family(), options.Options{}, font.FontSources{fc})
	if err != nil {
		t.Fatal(err)
	}

	words := []string{
		"الحِرَف",
		"صُممت",
		"الخبرات",
		"الجيران",
	}

	seen := map[uint16]map[string]struct{}{}
	recorder := newGlyphRecorder()
	var collisionGlyph uint16
	var collisionSeqs []string
	for _, word := range words {
		glyphs, err := f.Shaper.Shape([]rune(word), f, 12)
		if err != nil {
			t.Fatalf("shape %q: %v", word, err)
		}
		assignments := shapedGlyphRuneAssignments(glyphs, []rune(word))
		for i, gp := range glyphs {
			if seq := assignments[i]; len(seq) > 0 {
				bucket := seen[gp.GlyphID]
				if bucket == nil {
					bucket = map[string]struct{}{}
					seen[gp.GlyphID] = bucket
				}
				bucket[string(seq)] = struct{}{}
				recorder.recordRunes(gp.GlyphID, seq)
			}
		}
	}

	for gid, seqs := range seen {
		if len(seqs) > 1 {
			collisionGlyph = gid
			for seq := range seqs {
				collisionSeqs = append(collisionSeqs, seq)
			}
			break
		}
	}
	if collisionGlyph == 0 {
		t.Fatal("expected at least one reused glyph in sample words")
	}

	mapping := recorder.mapping()
	found := 0
	for _, seq := range collisionSeqs {
		for _, got := range mapping {
			if string(got) == seq {
				found++
				break
			}
		}
	}
	if found != len(collisionSeqs) {
		t.Fatalf("expected recorder to preserve all reused-glyph sequences for glyph %d, found %d of %d", collisionGlyph, found, len(collisionSeqs))
	}
}

func TestGlyphRecorder_MappingIncludesEmptyDestinations(t *testing.T) {
	recorder := newGlyphRecorder()
	textCID := recorder.recordRunes(10, []rune("خ"))
	emptyCID := recorder.recordEmpty(11)

	mapping := recorder.mapping()
	if got := string(mapping[textCID]); got != "خ" {
		t.Fatalf("mapped CID text = %q, want %q", got, "خ")
	}
	empty, ok := mapping[emptyCID]
	if !ok {
		t.Fatal("expected empty CID to be present in mapping")
	}
	if len(empty) != 0 {
		t.Fatalf("empty CID mapping = %q, want empty destination", string(empty))
	}
}

func TestShapedGlyphEmissionOrder_RoundTripsToOriginalText(t *testing.T) {
	f := loadAmiriFont(t)

	words := []string{
		"الحِرَف",
		"صُممت",
		"الخبرات",
		"الجيران",
		"اليدوية",
	}

	for _, word := range words {
		t.Run(word, func(t *testing.T) {
			runes := []rune(word)
			glyphs, err := f.Shaper.Shape(runes, f, 12)
			if err != nil {
				t.Fatalf("shape %q: %v", word, err)
			}

			assignments := shapedGlyphRuneAssignments(glyphs, runes)
			order := shapedGlyphEmissionOrder(glyphs, runes)
			if len(order) != len(glyphs) {
				t.Fatalf("emission order len = %d, want %d", len(order), len(glyphs))
			}

			var reconstructed []rune
			for _, glyphIndex := range order {
				reconstructed = append(reconstructed, assignments[glyphIndex]...)
			}
			if got := string(reconstructed); got != word {
				t.Fatalf("reconstructed text = %q, want %q", got, word)
			}
		})
	}
}

func TestUnicodeMode_ArabicExtractionRoundTripsExactly(t *testing.T) {
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
		"صُممت",
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
			path := filepath.Join(t.TempDir(), "arabic.pdf")
			if err := os.WriteFile(path, pdf.Bytes(), 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}

			out, err := exec.Command("pdftotext", path, "-").CombinedOutput()
			if err != nil {
				t.Fatalf("pdftotext: %v\n%s", err, out)
			}
			got := normalizeArabicExtractorOutput(string(out))
			if got != word {
				t.Fatalf("expected exact Arabic round-trip, got raw %q normalized %q want %q", strings.TrimSpace(string(out)), got, word)
			}
		})
	}
}

func TestUnicodeMode_ArabicExtractionPreservesDistinctSequences(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skipf("pdftotext unavailable: %v", err)
	}

	fc, err := ttf_fonts.New("../shaping/testdata/Amiri-Regular.ttf")
	if err != nil || len(fc.FontInfos) == 0 {
		t.Skipf("Arabic fixture font not found: %v", err)
	}
	family := fc.FontInfos[0].Family()

	const text = "الخبرات الجيران الحِرَف صُممت"

	dw := NewDocWriter()
	dw.AddFontSource(fc)
	pw := dw.NewPage()
	if _, err := pw.SetFont(family, 12, options.Options{}); err != nil {
		t.Fatalf("SetFont: %v", err)
	}
	pw.MoveTo(72, 720)
	pw.Print(text)

	var pdf bytes.Buffer
	if _, err := dw.WriteTo(&pdf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	path := filepath.Join(t.TempDir(), "arabic.pdf")
	if err := os.WriteFile(path, pdf.Bytes(), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	out, err := exec.Command("pdftotext", path, "-").CombinedOutput()
	if err != nil {
		t.Fatalf("pdftotext: %v\n%s", err, out)
	}
	got := normalizeArabicExtractorOutput(string(out))
	if got != text {
		t.Fatalf("expected exact phrase round-trip, got raw %q normalized %q want %q", strings.TrimSpace(string(out)), got, text)
	}
}

func TestUnicodeMode_ArabicPageTextHasNoUnmappedCIDs(t *testing.T) {
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
	pw.MoveTo(72, 720)
	pw.Print("الخبرات")

	var pdf bytes.Buffer
	if _, err := dw.WriteTo(&pdf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	gr := dw.glyphRecorders[fonts[0].PostScriptName()]
	mapping := gr.mapping()
	cids := gr.cids()
	if len(mapping) != len(cids) {
		t.Fatalf("expected every emitted CID to appear in ToUnicode mapping, got %d mappings for %d CIDs", len(mapping), len(cids))
	}

	emptyCount := 0
	for _, cid := range cids {
		runes, ok := mapping[cid]
		if !ok {
			t.Fatalf("missing ToUnicode entry for CID %d", cid)
		}
		if len(runes) == 0 {
			emptyCount++
		}
	}
	if emptyCount != 1 {
		t.Fatalf("expected exactly one empty ToUnicode destination, got %d", emptyCount)
	}
	if !strings.Contains(pdf.String(), "<>") {
		t.Fatalf("expected empty ToUnicode destination in PDF output, got excerpt:\n%s", extractCMapSection(pdf.String()))
	}
}

func TestUnicodeMode_ArabicCurvedTextHasNoUnmappedCIDs(t *testing.T) {
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
	if err := pw.DrawTextOnCircle("الخبرات", 3, 3, 1, 90, CurvedTextOptions{}); err != nil {
		t.Fatalf("DrawTextOnCircle: %v", err)
	}

	var pdf bytes.Buffer
	if _, err := dw.WriteTo(&pdf); err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	gr := dw.glyphRecorders[fonts[0].PostScriptName()]
	mapping := gr.mapping()
	cids := gr.cids()
	if len(mapping) != len(cids) {
		t.Fatalf("expected every emitted curved-text CID to appear in ToUnicode mapping, got %d mappings for %d CIDs", len(mapping), len(cids))
	}

	emptyCount := 0
	for _, cid := range cids {
		runes, ok := mapping[cid]
		if !ok {
			t.Fatalf("missing ToUnicode entry for curved-text CID %d", cid)
		}
		if len(runes) == 0 {
			emptyCount++
		}
	}
	if emptyCount != 1 {
		t.Fatalf("expected exactly one empty curved-text ToUnicode destination, got %d", emptyCount)
	}
	if !strings.Contains(pdf.String(), "<>") {
		t.Fatalf("expected empty curved-text ToUnicode destination in PDF output, got excerpt:\n%s", extractCMapSection(pdf.String()))
	}
}

func normalizeArabicExtractorOutput(s string) string {
	s = strings.TrimSpace(s)
	return strings.Map(func(r rune) rune {
		switch r {
		case '\u202B', '\u202C':
			return -1
		}
		return r
	}, s)
}
