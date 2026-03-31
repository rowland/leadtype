// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"testing"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/ttf_fonts"
)

func TestBidiDisplayPieces_SingleLeafMixedRTLAndLTR(t *testing.T) {
	skipIfNoTTFFonts(t)

	dw := NewDocWriter()
	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)

	face, err := font.New("Arial", options.Options{}, dw.fontSources)
	if err != nil {
		t.Fatal(err)
	}

	line := &rich_text.RichText{
		Text:     "للاستفسار: community@citylibrary.example",
		Font:     face,
		FontSize: 12,
	}

	pieces := bidiDisplayPieces(line)
	if len(pieces) != 2 {
		t.Fatalf("expected 2 display pieces, got %d", len(pieces))
	}
	if pieces[0].Text != "community@citylibrary.example" {
		t.Fatalf("expected email first in display order, got %q", pieces[0].Text)
	}
	if pieces[1].Text != "للاستفسار: " {
		t.Fatalf("expected Arabic prompt second in display order, got %q", pieces[1].Text)
	}
}

func TestBidiDisplayPieces_MultiLeafMixedRTLAndLTR(t *testing.T) {
	skipIfNoTTFFonts(t)

	dw := NewDocWriter()
	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)

	face, err := font.New("Arial", options.Options{}, dw.fontSources)
	if err != nil {
		t.Fatal(err)
	}

	line := (&rich_text.RichText{
		Text:     "للاستفسار: ",
		Font:     face,
		FontSize: 12,
	}).AddPiece(&rich_text.RichText{
		Text:     "community@citylibrary.example",
		Font:     face,
		FontSize: 12,
	})

	pieces := bidiDisplayPieces(line)
	if len(pieces) != 2 {
		t.Fatalf("expected 2 display pieces, got %d", len(pieces))
	}
	if pieces[0].Text != "community@citylibrary.example" {
		t.Fatalf("expected email first in display order, got %q", pieces[0].Text)
	}
	if pieces[1].Text != "للاستفسار: " {
		t.Fatalf("expected Arabic prompt second in display order, got %q", pieces[1].Text)
	}
}
