// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rowland/leadtype/internal/pdfsubset"
)

func TestPDFSubsetSessionSharesParsedCFF(t *testing.T) {
	first, err := LoadFont(minimalCIDCFF)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LoadFont(minimalCIDCFF)
	if err != nil {
		t.Fatal(err)
	}

	originalReader := readCFFSubsetTable
	reads := 0
	readCFFSubsetTable = func(font *Font, entry *tableDirEntry) ([]byte, error) {
		reads++
		return originalReader(font, entry)
	}
	defer func() { readCFFSubsetTable = originalReader }()

	session := pdfsubset.NewSession()
	aGID := first.GlyphIndex('A')
	aSubset, subtype, err := first.PDFSubsetWithSession(session, []uint16{aGID})
	if err != nil {
		t.Fatal(err)
	}
	if subtype != "CIDFontType0C" {
		t.Fatalf("first subtype = %q, want CIDFontType0C", subtype)
	}
	bGID := second.GlyphIndex('B')
	bSubset, _, err := second.PDFSubsetWithSession(session, []uint16{bGID})
	if err != nil {
		t.Fatal(err)
	}
	if reads != 1 {
		t.Fatalf("shared-session CFF reads = %d, want 1", reads)
	}
	if bytes.Equal(aSubset, bSubset) {
		t.Fatal("different glyph sets produced identical CFF subsets")
	}
	assertCIDSubsetGlyph(t, aSubset, uint16(1000)+aGID)
	assertCIDSubsetGlyph(t, bSubset, uint16(1000)+bGID)

	if _, _, err := second.PDFSubsetWithSession(pdfsubset.NewSession(), []uint16{bGID}); err != nil {
		t.Fatal(err)
	}
	if reads != 2 {
		t.Fatalf("new-session CFF reads = %d, want 2", reads)
	}
}

func TestPDFSubsetSessionCachesCFFLoadErrors(t *testing.T) {
	first, err := LoadFont(minimalCIDCFF)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LoadFont(minimalCIDCFF)
	if err != nil {
		t.Fatal(err)
	}

	originalReader := readCFFSubsetTable
	wantErr := errors.New("read failed")
	reads := 0
	readCFFSubsetTable = func(*Font, *tableDirEntry) ([]byte, error) {
		reads++
		return nil, wantErr
	}
	defer func() { readCFFSubsetTable = originalReader }()

	session := pdfsubset.NewSession()
	for _, candidate := range []*Font{first, second} {
		if _, _, err := candidate.PDFSubsetWithSession(session, []uint16{1}); !errors.Is(err, wantErr) {
			t.Fatalf("PDFSubsetWithSession error = %v, want %v", err, wantErr)
		}
	}
	if reads != 1 {
		t.Fatalf("failed CFF reads = %d, want 1 cached error", reads)
	}
}

func TestPDFSubsetSessionRejectsTruncatedCFFTable(t *testing.T) {
	font, err := LoadFont(minimalCIDCFF)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(minimalCIDCFF)
	if err != nil {
		t.Fatal(err)
	}
	entry := font.tableDir.table("CFF ")
	end := int(entry.offset + entry.length)
	path := filepath.Join(t.TempDir(), "truncated.otf")
	if err := os.WriteFile(path, raw[:end-1], 0o600); err != nil {
		t.Fatal(err)
	}
	font.filename = path
	font.rawBytes = nil

	if _, _, err := font.PDFSubsetWithSession(pdfsubset.NewSession(), []uint16{font.GlyphIndex('A')}); err == nil {
		t.Fatal("truncated CFF table unexpectedly produced a PDF subset")
	}
}

func TestCFFSubsetResourceKeySeparatesSourceTableAndGlyphCount(t *testing.T) {
	memory := &Font{}
	base := cffSubsetResourceKey{filename: "font.ttc", offset: 10, length: 20, numGlyphs: 30}
	variants := []cffSubsetResourceKey{
		{filename: "other.ttc", offset: 10, length: 20, numGlyphs: 30},
		{filename: "font.ttc", offset: 11, length: 20, numGlyphs: 30},
		{filename: "font.ttc", offset: 10, length: 21, numGlyphs: 30},
		{filename: "font.ttc", offset: 10, length: 20, numGlyphs: 31},
		{memory: memory, offset: 10, length: 20, numGlyphs: 30},
	}
	for _, variant := range variants {
		if base == variant {
			t.Fatalf("cache key did not distinguish %#v", variant)
		}
	}
}

func assertCIDSubsetGlyph(t *testing.T, data []byte, wantCID uint16) {
	t.Helper()
	parsed, err := parseCFFForSubset(data, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.charsets) != 2 || parsed.charsets[1] != wantCID {
		t.Fatalf("subset charsets = %v, want [0 %d]", parsed.charsets, wantCID)
	}
}
