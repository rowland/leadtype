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
	"unicode"

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

func TestUnicodeMode_ArabicExtractionPreservesDistinctSequences(t *testing.T) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		t.Skipf("pdftotext unavailable: %v", err)
	}

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
	pw.Print("الخبرات الجيران الحِرَف صُممت")

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
	got := strings.TrimSpace(string(out))
	normalized := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '\u202b' || r == '\u202c' {
			return -1
		}
		return r
	}, got)
	if !strings.Contains(normalized, "الخبرات") || !strings.Contains(normalized, "الجيران") || !strings.Contains(normalized, "الحِرَف") || !strings.Contains(normalized, "صُممت") {
		t.Fatalf("expected extracted text to preserve Arabic letters after normalization, got raw %q normalized %q", got, normalized)
	}
	if strings.Contains(got, "الححِررَف") || strings.Contains(got, "صُ᎗ممت") || strings.Contains(got, "الషخبرات") || strings.Contains(got, "ال఼خيران") {
		t.Fatalf("expected corrupted extraction artifacts to be gone, got %q", got)
	}
}
