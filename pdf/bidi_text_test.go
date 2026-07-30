// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/ttf_fonts"
	xbidi "golang.org/x/text/unicode/bidi"
)

func bidiTestFont(t *testing.T) *font.Font {
	t.Helper()
	source, err := ttf_fonts.New(filepath.Join("..", "shaping", "testdata", "Amiri-Regular.ttf"))
	if err != nil {
		t.Fatal(err)
	}
	if len(source.FontInfos) == 0 {
		t.Fatal("Amiri fixture contains no fonts")
	}
	face, err := font.New(source.FontInfos[0].Family(), options.Options{}, font.FontSources{source})
	if err != nil {
		t.Fatal(err)
	}
	return face
}

func displayPieceTexts(pieces []*rich_text.RichText) []string {
	texts := make([]string, len(pieces))
	for i, piece := range pieces {
		texts[i] = piece.Text
	}
	return texts
}

func TestBidiDisplayPieces_SingleLeafMixedRTLAndLTR(t *testing.T) {
	face := bidiTestFont(t)

	line := &rich_text.RichText{
		Text:     "للاستفسار: community@citylibrary.example",
		Font:     face,
		FontSize: 12,
	}

	pieces := bidiDisplayPieces(line, xbidi.RightToLeft)
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
	face := bidiTestFont(t)

	line := (&rich_text.RichText{
		Text:     "للاستفسار: ",
		Font:     face,
		FontSize: 12,
	}).AddPiece(&rich_text.RichText{
		Text:     "community@citylibrary.example",
		Font:     face,
		FontSize: 12,
	})

	pieces := bidiDisplayPieces(line, xbidi.RightToLeft)
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

func TestBidiDisplayPieces_ExplicitBaseDirectionControlsMixedPrompt(t *testing.T) {
	face := bidiTestFont(t)
	line := &rich_text.RichText{
		Text:     "Brent عادة ما يرى نفسه على أنه:",
		Font:     face,
		FontSize: 12,
	}

	tests := []struct {
		name string
		base xbidi.Direction
		want []string
	}{
		{
			name: "right to left",
			base: xbidi.RightToLeft,
			want: []string{" عادة ما يرى نفسه على أنه:", "Brent"},
		},
		{
			name: "left to right",
			base: xbidi.LeftToRight,
			want: []string{"Brent ", "عادة ما يرى نفسه على أنه", ":"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := displayPieceTexts(bidiDisplayPieces(line, tt.base))
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("display pieces = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestBidiDisplayPieces_LTRReportWithArabicName(t *testing.T) {
	face := bidiTestFont(t)
	line := &rich_text.RichText{
		Text:     "Report for برنت:",
		Font:     face,
		FontSize: 12,
	}

	got := displayPieceTexts(bidiDisplayPieces(line, xbidi.LeftToRight))
	want := []string{"Report for ", "برنت", ":"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("display pieces = %#v, want %#v", got, want)
	}
}

func TestBidiDisplayPieces_ExplicitRTLBaseSurvivesStyledLeafBoundary(t *testing.T) {
	face := bidiTestFont(t)
	line := (&rich_text.RichText{
		Text:     "Brent",
		Font:     face,
		FontSize: 14,
	}).AddPiece(&rich_text.RichText{
		Text:     " عادة ما يرى نفسه على أنه:",
		Font:     face,
		FontSize: 12,
	})

	got := displayPieceTexts(bidiDisplayPieces(line, xbidi.RightToLeft))
	want := []string{" عادة ما يرى نفسه على أنه:", "Brent"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("display pieces = %#v, want %#v", got, want)
	}
}
