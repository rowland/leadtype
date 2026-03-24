// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rowland/leadtype/codepage"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/ttf_fonts"
)

// TestToUnicodeCMapData_CP1252 verifies the CMap stream for CP1252 contains
// the expected entries for a sample of positions.
func TestToUnicodeCMapData_CP1252(t *testing.T) {
	data := toUnicodeCMapData(codepage.Idx_CP1252.Map())
	s := string(data)

	// Structure checks
	if !strings.Contains(s, "begincmap") {
		t.Error("missing begincmap")
	}
	if !strings.Contains(s, "endcmap") {
		t.Error("missing endcmap")
	}
	if !strings.Contains(s, "begincodespacerange") {
		t.Error("missing begincodespacerange")
	}
	if !strings.Contains(s, "<00> <FF>") {
		t.Error("codespace range should be <00> <FF>")
	}

	// byte 0x41 ('A') → U+0041
	if !strings.Contains(s, "<41> <0041>") {
		t.Error("missing mapping for 0x41 → U+0041")
	}
	// byte 0x80 → U+20AC (EURO SIGN, the CP1252-specific mapping)
	if !strings.Contains(s, "<80> <20AC>") {
		t.Errorf("missing CP1252 mapping for 0x80 → U+20AC;\ngot:\n%s", s)
	}
	// byte 0xFF → U+00FF (ÿ)
	if !strings.Contains(s, "<FF> <00FF>") {
		t.Error("missing mapping for 0xFF → U+00FF")
	}
}

// TestToUnicodeCMapData_BlockSize verifies that the CMap is split into
// beginbfchar/endbfchar blocks of at most 100 entries each.
func TestToUnicodeCMapData_BlockSize(t *testing.T) {
	// Build a 256-entry encoding where every byte maps to a non-zero rune.
	encoding := make([]rune, 256)
	for i := range encoding {
		encoding[i] = rune(0x0041 + i) // arbitrary non-zero runes
	}
	data := toUnicodeCMapData(encoding)
	s := string(data)

	// Count beginbfchar occurrences — should be ceil(256/100) = 3.
	count := strings.Count(s, "beginbfchar")
	if count != 3 {
		t.Errorf("expected 3 beginbfchar blocks for 256 entries, got %d", count)
	}
}

// TestToUnicodeCMapData_OmitsZero verifies that unmapped positions (rune 0)
// are not emitted in the CMap.
func TestToUnicodeCMapData_OmitsZero(t *testing.T) {
	encoding := make([]rune, 256) // all zero
	encoding[0x41] = 0x0041       // only one mapping
	data := toUnicodeCMapData(encoding)
	s := string(data)

	if !strings.Contains(s, "<41> <0041>") {
		t.Error("expected mapping for 0x41")
	}
	// Only one block with 1 entry.
	if !strings.Contains(s, "1 beginbfchar") {
		t.Error("expected exactly 1-entry beginbfchar block")
	}
}

// TestToUnicodeCMapData_ISO88592 verifies the CMap stream for ISO-8859-2
// correctly round-trips the positions that differ from ISO-8859-1.
func TestToUnicodeCMapData_ISO88592(t *testing.T) {
	data := toUnicodeCMapData(codepage.Idx_ISO_8859_2.Map())
	s := string(data)

	// 0xA1 → U+0104 (Ą, Latin capital A with ogonek) — differs from ISO-8859-1 (¡)
	if !strings.Contains(s, "<A1> <0104>") {
		t.Error("missing ISO-8859-2 mapping for 0xA1 → U+0104 (Ą)")
	}
	// 0xA3 → U+0141 (Ł, Latin capital L with stroke) — differs from ISO-8859-1 (£)
	if !strings.Contains(s, "<A3> <0141>") {
		t.Error("missing ISO-8859-2 mapping for 0xA3 → U+0141 (Ł)")
	}
	// 0xB1 → U+0105 (ą, Latin small a with ogonek) — differs from ISO-8859-1 (±)
	if !strings.Contains(s, "<B1> <0105>") {
		t.Error("missing ISO-8859-2 mapping for 0xB1 → U+0105 (ą)")
	}
	// 0xFE → U+0163 (ţ, Latin small t with cedilla) — differs from ISO-8859-1 (þ)
	if !strings.Contains(s, "<FE> <0163>") {
		t.Error("missing ISO-8859-2 mapping for 0xFE → U+0163 (ţ)")
	}
	// 0xFF → U+02D9 (dot above) — differs from ISO-8859-1 (ÿ)
	if !strings.Contains(s, "<FF> <02D9>") {
		t.Error("missing ISO-8859-2 mapping for 0xFF → U+02D9 (dot above)")
	}
	// Shared position: 0x41 ('A') should still be U+0041
	if !strings.Contains(s, "<41> <0041>") {
		t.Error("missing mapping for 0x41 → U+0041")
	}
}

// TestSimpleFont_ToUnicodeEntry_Fixture verifies that a generated PDF
// containing the minimal fixture font includes a /ToUnicode entry.
func TestSimpleFont_ToUnicodeEntry_Fixture(t *testing.T) {
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

	if !strings.Contains(pdf, "/ToUnicode") {
		t.Error("generated PDF missing /ToUnicode entry")
	}
	// The CMap stream content should also be present.
	if !strings.Contains(pdf, "begincmap") {
		t.Error("generated PDF missing CMap stream content")
	}
}

// TestSimpleFont_ToUnicodeEntry_SystemFont verifies that a generated PDF
// contains a /ToUnicode entry when system fonts are available.
func TestSimpleFont_ToUnicodeEntry_SystemFont(t *testing.T) {
	skipIfNoTTFFonts(t)

	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}

	dw := NewDocWriter()
	dw.AddFontSource(fc)

	pw := dw.NewPage()
	pw.SetFont("Arial", 12, options.Options{})
	pw.MoveTo(72, 720)
	pw.Print("Hello, World!")

	var buf bytes.Buffer
	dw.WriteTo(&buf)
	pdf := buf.String()

	if !strings.Contains(pdf, "/ToUnicode") {
		t.Error("generated PDF missing /ToUnicode entry")
	}
}
