// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"testing"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
)

// mockWriter records SetFont and AddFont calls for inspection.
type mockWriter struct {
	setFontName  string
	setFontSize  float64
	addFontNames []string
	fonts        []*font.Font
}

func (m *mockWriter) AddFont(family string, opts options.Options) ([]*font.Font, error) {
	m.addFontNames = append(m.addFontNames, family)
	return m.fonts, nil
}
func (m *mockWriter) SetFont(name string, size float64, opts options.Options) ([]*font.Font, error) {
	m.setFontName = name
	m.setFontSize = size
	m.addFontNames = nil
	return m.fonts, nil
}
func (m *mockWriter) FontColor() interface{}                            { return nil }
func (m *mockWriter) Fonts() []*font.Font                              { return m.fonts }
func (m *mockWriter) FontSize() float64                                { return m.setFontSize }
func (m *mockWriter) LineSpacing() float64                             { return 1.0 }
func (m *mockWriter) LineTo(x, y float64)                              {}
func (m *mockWriter) Loc() (float64, float64)                          { return 0, 0 }
func (m *mockWriter) MoveTo(x, y float64)                              {}
func (m *mockWriter) NewPage()                                         {}
func (m *mockWriter) Print(text string) error                          { return nil }
func (m *mockWriter) PrintParagraph(para interface{}, opts interface{}) {}
func (m *mockWriter) PrintRichText(text interface{})                   {}
func (m *mockWriter) Rectangle(x, y, w, h float64, b, f bool)         {}
func (m *mockWriter) Rectangle2(x, y, w, h float64, b, f bool, c []float64, p, r bool) {}
func (m *mockWriter) SetFillColor(v interface{}) interface{}           { return nil }
func (m *mockWriter) SetLineColor(v interface{}) interface{}           { return nil }
func (m *mockWriter) SetLineDashPattern(p string) string               { return "" }
func (m *mockWriter) SetLineSpacing(ls float64) float64                { return 0 }
func (m *mockWriter) SetLineWidth(w float64)                           {}
func (m *mockWriter) SetStrikeout(s bool) bool                         { return false }
func (m *mockWriter) SetUnderline(u bool) bool                         { return false }
func (m *mockWriter) Strikeout() bool                                  { return false }
func (m *mockWriter) Underline() bool                                  { return false }

func TestFontStyle_SetAttrs_SingleName(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs("", map[string]string{"name": "Helvetica", "size": "12"})
	if len(fs.entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(fs.entries))
	}
	if fs.entries[0].name != "Helvetica" {
		t.Errorf("expected Helvetica, got %s", fs.entries[0].name)
	}
	if fs.size != 12 {
		t.Errorf("expected size 12, got %f", fs.size)
	}
}

func TestFontStyle_SetAttrs_MultipleNames(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs("", map[string]string{"name": "Helvetica, Arial Unicode MS, Courier"})
	if len(fs.entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(fs.entries))
	}
	if fs.entries[0].name != "Helvetica" {
		t.Errorf("entry 0: expected Helvetica, got %s", fs.entries[0].name)
	}
	if fs.entries[1].name != "Arial Unicode MS" {
		t.Errorf("entry 1: expected Arial Unicode MS, got %s", fs.entries[1].name)
	}
	if fs.entries[2].name != "Courier" {
		t.Errorf("entry 2: expected Courier, got %s", fs.entries[2].name)
	}
}

func TestFontStyle_SetAttrs_Ranges(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs("", map[string]string{
		"name":   "Helvetica, NotoSansCJK",
		"ranges": " | CJK Unified Ideographs",
	})
	if len(fs.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(fs.entries))
	}
	if len(fs.entries[0].ranges) != 0 {
		t.Errorf("entry 0: expected no ranges, got %v", fs.entries[0].ranges)
	}
	if len(fs.entries[1].ranges) != 1 || fs.entries[1].ranges[0] != "CJK Unified Ideographs" {
		t.Errorf("entry 1: expected [CJK Unified Ideographs], got %v", fs.entries[1].ranges)
	}
}

func TestFontStyle_SetAttrs_Sizes(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs("", map[string]string{
		"name":  "Helvetica, NotoSansCJK",
		"sizes": "1.0 | 0.9",
	})
	if len(fs.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(fs.entries))
	}
	if fs.entries[0].relativeSize != 1.0 {
		t.Errorf("entry 0: expected relativeSize 1.0, got %f", fs.entries[0].relativeSize)
	}
	if fs.entries[1].relativeSize != 0.9 {
		t.Errorf("entry 1: expected relativeSize 0.9, got %f", fs.entries[1].relativeSize)
	}
}

func TestFontStyle_SetAttrs_Prefix(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs("font.", map[string]string{
		"font.name":  "Helvetica, Arial",
		"font.size":  "14",
		"font.sizes": "1.0 | 0.85",
	})
	if len(fs.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(fs.entries))
	}
	if fs.entries[1].relativeSize != 0.85 {
		t.Errorf("entry 1: expected relativeSize 0.85, got %f", fs.entries[1].relativeSize)
	}
	if fs.size != 14 {
		t.Errorf("expected size 14, got %f", fs.size)
	}
}

func TestFontStyle_Clone_DeepCopiesEntries(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs("", map[string]string{
		"name":   "Helvetica, NotoSansCJK",
		"ranges": " | CJK Unified Ideographs",
	})
	clone := fs.Clone()
	// Mutate original; clone should be unaffected.
	fs.entries[1].ranges[0] = "CHANGED"
	if clone.entries[1].ranges[0] != "CJK Unified Ideographs" {
		t.Errorf("Clone shares ranges slice with original")
	}
}

func TestFontStyle_String_MultipleEntries(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs("", map[string]string{"name": "Helvetica, Arial"})
	s := fs.String()
	if s == "" {
		t.Error("String() returned empty")
	}
	// Should contain both names joined by comma.
	if !contains(s, "Helvetica,Arial") {
		t.Errorf("String() missing font names: %s", s)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsRune(s, substr))
}

func containsRune(s, substr string) bool {
	for i := range s {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFontStyle_DefaultFont(t *testing.T) {
	if len(defaultFont.entries) != 1 {
		t.Fatalf("defaultFont should have 1 entry, got %d", len(defaultFont.entries))
	}
	if defaultFont.entries[0].name != defaultFontName {
		t.Errorf("defaultFont entry name: expected %s, got %s", defaultFontName, defaultFont.entries[0].name)
	}
}
