// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"errors"
	"strings"
	"testing"

	"github.com/rowland/leadtype/rich_text"
	xbidi "golang.org/x/text/unicode/bidi"
)

func TestPageWriter_WithTextDirectionNestsAndRestores(t *testing.T) {
	pw := NewDocWriter().NewPage()
	if pw.textDirectionSet {
		t.Fatal("new page unexpectedly has an explicit text direction")
	}
	if err := pw.WithTextDirection(TextDirectionRTL, nil); err != nil {
		t.Fatal(err)
	}
	if pw.textDirectionSet {
		t.Fatal("nil callback unexpectedly changed the text direction")
	}

	sentinel := errors.New("sentinel")
	err := pw.WithTextDirection(TextDirectionRTL, func() error {
		if !pw.textDirectionSet || pw.textDirection != TextDirectionRTL {
			t.Fatalf("outer direction = %v/%v, want explicit RTL", pw.textDirection, pw.textDirectionSet)
		}
		if got := pw.bidiBaseDirection("Brent مرحبا:"); got != xbidi.RightToLeft {
			t.Fatalf("outer bidi base = %v, want RTL", got)
		}
		if err := pw.WithTextDirection(TextDirectionRTL, func() error {
			if pw.textDirection != TextDirectionRTL {
				t.Fatalf("same-direction nested scope = %v, want RTL", pw.textDirection)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if err := pw.WithTextDirection(TextDirectionLTR, func() error {
			if pw.textDirection != TextDirectionLTR {
				t.Fatalf("opposite nested scope = %v, want LTR", pw.textDirection)
			}
			return sentinel
		}); !errors.Is(err, sentinel) {
			t.Fatalf("nested error = %v, want sentinel", err)
		}
		if pw.textDirection != TextDirectionRTL {
			t.Fatalf("direction after nested scope = %v, want RTL", pw.textDirection)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if pw.textDirectionSet {
		t.Fatal("direction scope did not restore the legacy unset state")
	}
	if got := pw.bidiBaseDirection("Brent مرحبا:"); got != xbidi.LeftToRight {
		t.Fatalf("legacy first-strong base = %v, want LTR", got)
	}
}

func TestPageWriter_CurvedTextUsesScopedBaseDirection(t *testing.T) {
	face := bidiTestFont(t)
	pw := NewDocWriter().NewPage()
	pw.supportsArabicShaping = true
	line := &rich_text.RichText{
		Text:     "Brent عادة ما يرى نفسه على أنه:",
		Font:     face,
		FontSize: 12,
	}

	firstLatinIndex := func(direction TextDirection) int {
		t.Helper()
		index := -1
		err := pw.WithTextDirection(direction, func() error {
			glyphs, err := pw.curvedTextGlyphsForRichText(line)
			if err != nil {
				return err
			}
			for i, glyph := range glyphs {
				if strings.Contains(glyph.Text, "B") {
					index = i
					break
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return index
	}

	if got := firstLatinIndex(TextDirectionLTR); got != 0 {
		t.Fatalf("LTR first Latin glyph index = %d, want 0", got)
	}
	if got := firstLatinIndex(TextDirectionRTL); got <= 0 {
		t.Fatalf("RTL first Latin glyph index = %d, want it after the Arabic run", got)
	}
}
