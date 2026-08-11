// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"errors"
	"testing"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

// mockWriter records SetFont and AddFont calls for inspection.
type mockWriter struct {
	NoopWriter
	setFontName   string
	setFontSize   float64
	addFontNames  []string
	fonts         []*font.Font
	setFontCalls  []string
	setFontOpts   []options.Options
	addFontCalls  []string
	setFontErrs   map[string]error
	setFontErrSeq map[string][]error
	addFontErrs   map[string]error
	t             testing.TB
}

type testFontProvider struct {
	font *FontStyle
}

func (p *testFontProvider) Font() *FontStyle {
	return p.font
}

func TestSetFontStyleAssignsNamedStyle(t *testing.T) {
	scope := &Scope{}
	base := &FontStyle{id: "body", entries: []fontEntry{{name: "Helvetica"}}}
	if err := scope.AddStyle(base); err != nil {
		t.Fatal(err)
	}

	var field *FontStyle
	SetFontStyle(&field, "font", map[string]string{"font": "body"}, scope, "pt", nil)

	if field != base {
		t.Fatalf("font = %p, want named style %p", field, base)
	}
}

func TestSetFontStyleClonesNamedStyleBeforeOverrides(t *testing.T) {
	oldScope := &Scope{}
	scope := &Scope{}
	base := &FontStyle{
		scope:   oldScope,
		id:      "body",
		entries: []fontEntry{{name: "Helvetica"}},
		weight:  "Regular",
	}
	if err := scope.AddStyle(base); err != nil {
		t.Fatal(err)
	}

	var field *FontStyle
	SetFontStyle(&field, "font", map[string]string{
		"font":        "body",
		"font.weight": "Bold",
	}, scope, "pt", nil)

	if field == base {
		t.Fatal("font reused named style, want clone")
	}
	if field.weight != "Bold" {
		t.Fatalf("font weight = %q, want Bold", field.weight)
	}
	if field.scope != scope {
		t.Fatalf("font scope = %p, want %p", field.scope, scope)
	}
	if base.weight != "Regular" || base.scope != oldScope {
		t.Fatalf("named style mutated: %#v", base)
	}
}

func TestSetFontStyleClonesParentAndAppliesWidgetUnits(t *testing.T) {
	base := &FontStyle{entries: []fontEntry{{name: "Helvetica"}}, weight: "Regular"}
	parent := &testFontProvider{font: base}
	var field *FontStyle

	SetFontStyle(&field, "font", map[string]string{
		"font.weight":       "Bold",
		"font.stroke-width": "2",
	}, nil, "mm", parent)

	if field == base {
		t.Fatal("font reused parent style, want clone")
	}
	if field.strokeWidth == nil || *field.strokeWidth != FromUnits(2, "mm") {
		t.Fatalf("stroke width = %v, want 2mm", field.strokeWidth)
	}
	if base.weight != "Regular" || base.strokeWidth != nil {
		t.Fatalf("parent font mutated: %#v", base)
	}
}

func TestSetFontStyleMissingNameFallsBackToParent(t *testing.T) {
	scope := &Scope{}
	base := &FontStyle{entries: []fontEntry{{name: "Helvetica"}}, weight: "Regular"}
	parent := &testFontProvider{font: base}
	field := &FontStyle{weight: "Old"}

	SetFontStyle(&field, "font", map[string]string{
		"font":        "missing",
		"font.weight": "Bold",
	}, scope, "pt", parent)

	if field == base || field.weight != "Bold" {
		t.Fatalf("font = %#v, want overridden clone of parent", field)
	}
}

func TestSetFontStyleNilParentUsesDefaultAndExplicitUnits(t *testing.T) {
	var field *FontStyle

	SetFontStyle(&field, "text-font", map[string]string{
		"text-font.units":        "cm",
		"text-font.stroke-width": "2",
	}, nil, "mm", nil)

	if field == defaultFont {
		t.Fatal("font reused default style, want clone")
	}
	if field.strokeWidth == nil || *field.strokeWidth != FromUnits(2, "cm") {
		t.Fatalf("stroke width = %v, want 2cm", field.strokeWidth)
	}
	if defaultFont.strokeWidth != nil {
		t.Fatalf("default font stroke width = %v, want unchanged nil", defaultFont.strokeWidth)
	}
}

func TestSetFontStyleLeavesFieldUnchangedWithoutMatchingAttrs(t *testing.T) {
	base := &FontStyle{}
	field := base

	SetFontStyle(&field, "font", map[string]string{"text-font.weight": "Bold"}, nil, "pt", nil)

	if field != base {
		t.Fatalf("font = %p, want unchanged %p", field, base)
	}
}

func (m *mockWriter) AddFont(family string, opts options.Options) ([]*font.Font, error) {
	m.addFontCalls = append(m.addFontCalls, family)
	m.addFontNames = append(m.addFontNames, family)
	if err := m.addFontErrs[family]; err != nil {
		return nil, err
	}
	return m.Fonts(), nil
}

func (m *mockWriter) SetFont(name string, size float64, opts options.Options) ([]*font.Font, error) {
	m.setFontCalls = append(m.setFontCalls, name)
	m.setFontOpts = append(m.setFontOpts, opts)
	m.setFontName = name
	m.setFontSize = size
	m.addFontNames = nil
	if errs := m.setFontErrSeq[name]; len(errs) > 0 {
		err := errs[0]
		m.setFontErrSeq[name] = errs[1:]
		if err != nil {
			m.fonts = nil
			return nil, err
		}
	}
	if err := m.setFontErrs[name]; err != nil {
		m.fonts = nil
		return nil, err
	}
	return m.Fonts(), nil
}

func (m *mockWriter) Fonts() []*font.Font {
	if len(m.fonts) == 0 && m.t != nil {
		m.fonts = defaultTestFonts(m.t)
	}
	return m.fonts
}

func (m *mockWriter) FontSize() float64 { return m.setFontSize }

// Preserve the historical mock behavior for callers that inspect the previous
// line spacing returned by SetLineSpacing.
func (m *mockWriter) SetLineSpacing(float64) float64 { return 0 }

func TestFontStyle_SetAttrs_SingleName(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs(map[string]string{"name": "Helvetica", "size": "12"})
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
	fs.SetAttrs(map[string]string{"name": "Helvetica, Arial Unicode MS, Courier"})
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
	fs.SetAttrs(map[string]string{
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
	fs.SetAttrs(map[string]string{
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
	fs.SetAttrs(map[string]string{
		"name":  "Helvetica, Arial",
		"size":  "14",
		"sizes": "1.0 | 0.85",
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

func TestFontStyle_SetAttrs_RemSize(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs(map[string]string{"size": "1.25rem"})
	if fs.sizeSpec.kind != fontSizeRem {
		t.Fatalf("expected rem size spec, got %#v", fs.sizeSpec)
	}
	if fs.sizeSpec.value != 1.25 {
		t.Fatalf("expected rem multiplier 1.25, got %v", fs.sizeSpec.value)
	}
	if got := fs.ResolveAgainstBase(20); got != 25 {
		t.Fatalf("ResolveAgainstBase(20) = %v, want 25", got)
	}
}

func TestFontStyle_SetAttrs_PointSizeSuffix(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs(map[string]string{"size": "14pt"})
	if fs.sizeSpec.kind != fontSizeAbsolute {
		t.Fatalf("expected absolute size spec, got %#v", fs.sizeSpec)
	}
	if fs.size != 14 {
		t.Fatalf("size = %v, want 14", fs.size)
	}
	if got := fs.ResolveAgainstBase(20); got != 14 {
		t.Fatalf("ResolveAgainstBase(20) = %v, want 14", got)
	}
}

func TestFontStyle_SetAttrs_InvalidRemFallsBackToDefault(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs(map[string]string{"size": "bogusrem"})
	if fs.size != defaultFontSize {
		t.Fatalf("size = %v, want default %v", fs.size, defaultFontSize)
	}
	if fs.sizeSpec.kind != fontSizeAbsolute || fs.sizeSpec.value != defaultFontSize {
		t.Fatalf("sizeSpec = %#v, want absolute default %v", fs.sizeSpec, defaultFontSize)
	}
}

func TestFontStyle_Clone_DeepCopiesEntries(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs(map[string]string{
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
	fs.SetAttrs(map[string]string{"name": "Helvetica, Arial"})
	s := fs.String()
	if s == "" {
		t.Error("String() returned empty")
	}
	// Should contain both names joined by comma.
	if !contains(s, "Helvetica,Arial") {
		t.Errorf("String() missing font names: %s", s)
	}
}

func TestFontStyle_SetAttrs_DecorationOverrides(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)
	if err := scope.AddStyle(&PenStyle{id: "accent", color: colors.Red, width: 1.5, pattern: "dashed", cap: "round_cap"}); err != nil {
		t.Fatal(err)
	}

	var fs FontStyle
	fs.SetScope(scope)
	fs.SetAttrs(map[string]string{
		"underline-pen": "accent",
		"strikeout-pos": "0.25in",
		"underline-pos": "1cm",
		"strikeout-pen": "missing",
	})

	if fs.underlinePenID != "accent" {
		t.Fatalf("expected underline pen id to be recorded")
	}
	if fs.strikeoutPenID != "missing" {
		t.Fatalf("expected strikeout pen id to be recorded")
	}
	if fs.underlinePos == nil || *fs.underlinePos != FromUnits(1, "cm") {
		t.Fatalf("expected underline position in points, got %v", fs.underlinePos)
	}
	if fs.strikeoutPos == nil || *fs.strikeoutPos != 18 {
		t.Fatalf("expected strikeout position in points, got %v", fs.strikeoutPos)
	}

	opts := fs.RichTextOptions()
	decoration, _ := opts["decoration"].(*rich_text.DecorationOverrides)
	if decoration == nil {
		t.Fatalf("expected decoration payload")
	}
	opts2 := fs.RichTextOptions()
	decoration2, _ := opts2["decoration"].(*rich_text.DecorationOverrides)
	if decoration2 != decoration {
		t.Fatalf("expected decoration payload to be cached and reused")
	}
	if !decoration.Underline.HasColor || decoration.Underline.Color != colors.Red {
		t.Fatalf("expected underline color override from pen")
	}
	if !decoration.Underline.HasWidth || decoration.Underline.Width != 1.5 {
		t.Fatalf("expected underline width override from pen")
	}
	if !decoration.Underline.HasPattern || decoration.Underline.Pattern != "dashed" {
		t.Fatalf("expected underline dash override from pen")
	}
	if decoration.Underline.CapStyle != "round_cap" {
		t.Fatalf("expected underline cap override from pen")
	}
	if !decoration.Underline.HasPosition || decoration.Underline.Position != FromUnits(1, "cm") {
		t.Fatalf("expected underline position override")
	}
	if decoration.Strikeout.HasColor || decoration.Strikeout.HasWidth || decoration.Strikeout.HasPattern || decoration.Strikeout.CapStyle != "" {
		t.Fatalf("did not expect missing pen id to synthesize style overrides")
	}
	if !decoration.Strikeout.HasPosition || decoration.Strikeout.Position != 18 {
		t.Fatalf("expected strikeout position override")
	}
}

func TestFontStyle_SetAttrs_AdHocDecorationPenUsesDeclarationUnits(t *testing.T) {
	var fs FontStyle
	fs.SetScope(&Scope{})
	fs.SetAttrs(map[string]string{
		"units":         "mm",
		"underline-pen": "2 dashed #08f",
	})
	// A later layer changing units must not reinterpret the earlier pen value.
	fs.SetAttrs(map[string]string{"units": "in", "weight": "Bold"})

	decoration := fs.RichTextOptions()["decoration"].(*rich_text.DecorationOverrides)
	line := decoration.Underline
	if !line.HasWidth || line.Width != FromUnits(2, "mm") || !line.HasPattern || line.Pattern != "dashed" || !line.HasColor || line.Color != NamedColor("#0088ff") {
		t.Fatalf("underline = %#v, want 2mm dashed #0088ff", line)
	}
}

func TestFontStyle_AdHocDecorationPenWithoutWidthPreservesMetricThickness(t *testing.T) {
	var fs FontStyle
	fs.SetScope(&Scope{})
	fs.SetAttrs(map[string]string{
		"underline-pen": "dashed #08f",
		"strikeout-pen": "#c33",
	})

	decoration := fs.RichTextOptions()["decoration"].(*rich_text.DecorationOverrides)
	if decoration.Underline.HasWidth || decoration.Strikeout.HasWidth {
		t.Fatalf("decoration = %#v, omitted widths must preserve metric-derived thickness", decoration)
	}
	if !decoration.Underline.HasColor || !decoration.Underline.HasPattern || !decoration.Strikeout.HasColor {
		t.Fatalf("decoration = %#v, want non-width pen overrides", decoration)
	}
}

func TestFontStyle_AdHocDecorationPenExplicitZeroWidthOverridesMetricThickness(t *testing.T) {
	var fs FontStyle
	fs.SetScope(&Scope{})
	fs.SetAttrs(map[string]string{"underline-pen": "0 solid #08f"})

	line := fs.RichTextOptions()["decoration"].(*rich_text.DecorationOverrides).Underline
	if !line.HasWidth || line.Width != 0 {
		t.Fatalf("underline = %#v, explicit zero width must request PDF hairline", line)
	}
}

func TestFontStyle_NamedDecorationPenWithoutWidthPreservesMetricThickness(t *testing.T) {
	scope := &Scope{}
	named := &PenStyle{id: "color-only", color: NamedColor("Gold"), pattern: "dotted"}
	if err := scope.AddStyle(named); err != nil {
		t.Fatal(err)
	}
	var fs FontStyle
	fs.SetScope(scope)
	fs.SetAttrs(map[string]string{"underline-pen": "color-only"})

	line := fs.RichTextOptions()["decoration"].(*rich_text.DecorationOverrides).Underline
	if line.HasWidth {
		t.Fatalf("underline = %#v, named pen without width must preserve metric-derived thickness", line)
	}
}

func TestFontStyle_DecorationPenPreservesForwardNamedReference(t *testing.T) {
	scope := &Scope{}
	var fs FontStyle
	fs.SetScope(scope)
	fs.SetAttrs(map[string]string{"underline-pen": "later"})

	named := &PenStyle{id: "later", width: 3, pattern: "dotted", color: NamedColor("Gold")}
	if err := scope.AddStyle(named); err != nil {
		t.Fatal(err)
	}

	decoration := fs.RichTextOptions()["decoration"].(*rich_text.DecorationOverrides)
	line := decoration.Underline
	if !line.HasWidth || line.Width != 3 || !line.HasPattern || line.Pattern != "dotted" || !line.HasColor || line.Color != NamedColor("Gold") {
		t.Fatalf("underline = %#v, want forward named pen", line)
	}
}

func TestFontStyle_RichTextOptions_IncludesTextStroke(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs(map[string]string{
		"color":        "White",
		"stroke-color": "Black",
		"stroke-width": "0.75pt",
	})

	opts := fs.RichTextOptions()
	decoration, _ := opts["decoration"].(*rich_text.DecorationOverrides)
	if decoration == nil {
		t.Fatal("expected decoration payload")
	}
	if !decoration.TextStroke.HasColor || decoration.TextStroke.Color != colors.Black {
		t.Fatalf("stroke color = %v, want %v", decoration.TextStroke.Color, colors.Black)
	}
	if !decoration.TextStroke.HasWidth || decoration.TextStroke.Width != 0.75 {
		t.Fatalf("stroke width = %v, want 0.75", decoration.TextStroke.Width)
	}
}

func TestFontStyle_SetAttrs_InvalidDecorationMeasurementIsUnset(t *testing.T) {
	var fs FontStyle
	fs.SetAttrs(map[string]string{"underline-pos": "bogus"})
	if fs.underlinePos != nil {
		t.Fatalf("expected invalid underline-pos to remain unset")
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

func TestFontStyle_Apply_UsesFirstAvailableFontInChain(t *testing.T) {
	fs := &FontStyle{
		entries: []fontEntry{
			{name: "Missing Primary"},
			{name: "Helvetica"},
			{name: "Courier"},
		},
		size: 12,
	}
	w := &mockWriter{
		t:           t,
		setFontErrs: map[string]error{"Missing Primary": errors.New("not found")},
		addFontErrs: map[string]error{"Missing Primary": errors.New("not found")},
	}

	fs.Apply(w)

	if len(w.setFontCalls) != 2 || w.setFontCalls[0] != "Missing Primary" || w.setFontCalls[1] != "Helvetica" {
		t.Fatalf("SetFont calls = %v, want exact attempts in fallback order", w.setFontCalls)
	}
	if w.setFontName != "Helvetica" {
		t.Fatalf("final SetFont name = %q, want %q", w.setFontName, "Helvetica")
	}
	if len(w.addFontCalls) != 3 || w.addFontCalls[0] != "Missing Primary" || w.addFontCalls[1] != "Missing Primary" || w.addFontCalls[2] != "Courier" {
		t.Fatalf("AddFont calls = %v, want per-family exact and nearest fallbacks", w.addFontCalls)
	}
}

func TestFontStyle_Apply_PrefersExactFaceBeforeNearest(t *testing.T) {
	fs := &FontStyle{
		entries: []fontEntry{
			{name: "Primary"},
			{name: "Earlier Fallback"},
			{name: "Later Exact"},
		},
		size:   12,
		weight: "Bold",
	}
	w := &mockWriter{
		t: t,
		addFontErrs: map[string]error{
			"Earlier Fallback": errors.New("bold not found"),
		},
	}

	fs.Apply(w)

	if len(w.addFontCalls) != 3 || w.addFontCalls[0] != "Earlier Fallback" || w.addFontCalls[1] != "Earlier Fallback" || w.addFontCalls[2] != "Later Exact" {
		t.Fatalf("AddFont calls = %v, want per-family exact and nearest fallbacks", w.addFontCalls)
	}
}

func TestFontStyle_Apply_FallsBackToDefaultFontWhenChainMissing(t *testing.T) {
	fs := &FontStyle{
		entries: []fontEntry{
			{name: "Missing One"},
			{name: "Missing Two"},
		},
		size: 12,
	}
	w := &mockWriter{
		t: t,
		setFontErrs: map[string]error{
			"Missing One": errors.New("not found"),
			"Missing Two": errors.New("not found"),
		},
	}

	fs.Apply(w)

	if got := w.setFontName; got != defaultFontName {
		t.Fatalf("final SetFont name = %q, want %q", got, defaultFontName)
	}
	if len(w.setFontCalls) != 3 || w.setFontCalls[2] != defaultFontName {
		t.Fatalf("SetFont calls = %v, want fallback to %q", w.setFontCalls, defaultFontName)
	}
}

func TestFontStyle_Apply_UsesNearestDefaultFaceWhenRequestedWeightMissing(t *testing.T) {
	fs := &FontStyle{
		entries: []fontEntry{
			{name: "Missing One"},
			{name: "Missing Two"},
		},
		size:   12,
		weight: "Black",
	}
	w := &mockWriter{
		t: t,
		setFontErrs: map[string]error{
			"Missing One": errors.New("not found"),
			"Missing Two": errors.New("not found"),
		},
		setFontErrSeq: map[string][]error{
			defaultFontName: {errors.New("black not found"), nil},
		},
	}

	fs.Apply(w)

	if len(w.setFontCalls) != 4 || w.setFontCalls[2] != defaultFontName || w.setFontCalls[3] != defaultFontName {
		t.Fatalf("SetFont calls = %v, want final default retry", w.setFontCalls)
	}
	if got := w.setFontOpts[2].StringDefault("weight", ""); got != "Black" {
		t.Fatalf("first default fallback weight = %q, want Black", got)
	}
	if got := w.setFontOpts[3].StringDefault("match", ""); got != "nearest" {
		t.Fatalf("default fallback match = %q, want nearest", got)
	}
}
