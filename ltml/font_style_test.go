// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"errors"
	"testing"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/rich_text"
)

// mockWriter records SetFont and AddFont calls for inspection.
type mockWriter struct {
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

func (m *mockWriter) Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error {
	return nil
}

func (m *mockWriter) Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error {
	return nil
}

func (m *mockWriter) Clip(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}

func (m *mockWriter) ClipClosedShape(shape pdf.ClosedShape, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}

func (m *mockWriter) ClipRichText(text *rich_text.RichText, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}

func (m *mockWriter) ClipText(text string, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}

func (m *mockWriter) Circle(x, y, r float64, border, fill, reverse bool) error { return nil }
func (m *mockWriter) ClosedShapeBounds(shape pdf.ClosedShape) (pdf.Bounds, error) {
	return shape.Bounds()
}
func (m *mockWriter) CompressEmbeddedFonts(bool) *pdf.DocWriter { return nil }
func (m *mockWriter) CompressPages(bool) *pdf.DocWriter         { return nil }
func (m *mockWriter) CompressToUnicode(bool) *pdf.DocWriter     { return nil }

func (m *mockWriter) DrawRichTextOnCircle(text *rich_text.RichText, x, y, r, startAngle float64, opts pdf.CurvedTextOptions) error {
	return nil
}

func (m *mockWriter) DrawTextOnCircle(text string, x, y, r, startAngle float64, opts pdf.CurvedTextOptions) error {
	return nil
}

func (m *mockWriter) DrawClosedShape(shape pdf.ClosedShape, border, fill bool) error {
	return nil
}

func (m *mockWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	return nil
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

func (m *mockWriter) FontColor() colors.Color { return 0 }

func (m *mockWriter) Fonts() []*font.Font {
	if len(m.fonts) == 0 && m.t != nil {
		m.fonts = defaultTestFonts(m.t)
	}
	return m.fonts
}

func (m *mockWriter) FontSize() float64 { return m.setFontSize }

func (m *mockWriter) ImageDimensions(data []byte) (int, int, error) {
	return 0, 0, nil
}

func (m *mockWriter) SVGDimensions(data []byte) (int, int, error) {
	return 0, 0, nil
}

func (m *mockWriter) SVGDimensionsFromFile(filename string) (int, int, error) {
	return 0, 0, nil
}

func (m *mockWriter) ImageDimensionsFromFile(filename string) (int, int, error) {
	return 0, 0, nil
}

func (m *mockWriter) LineSpacing() float64                { return 1.0 }
func (m *mockWriter) SetLineCapStyle(style string) string { return "" }
func (m *mockWriter) Line(x, y, angle, length float64)    {}
func (m *mockWriter) LineTo(x, y float64)                 {}
func (m *mockWriter) Loc() (float64, float64)             { return 0, 0 }
func (m *mockWriter) MoveTo(x, y float64)                 {}
func (m *mockWriter) NewPage()                            {}
func (m *mockWriter) Print(text string) error             { return nil }

func (m *mockWriter) PrintImage(data []byte, x, y float64, width, height *float64) (float64, float64, error) {
	return 0, 0, nil
}

func (m *mockWriter) PrintSVG(data []byte, x, y float64, width, height *float64) (float64, float64, error) {
	return 0, 0, nil
}

func (m *mockWriter) PrintSVGFile(filename string, x, y float64, width, height *float64) (float64, float64, error) {
	return 0, 0, nil
}

func (m *mockWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (float64, float64, error) {
	return 0, 0, nil
}

func (m *mockWriter) PaintImageFile(filename string, x, y, width, height, opacity float64) error {
	return nil
}
func (m *mockWriter) PaintLinearGradient(lg *pdf.LinearGradient) error                { return nil }
func (m *mockWriter) PaintRadialGradient(rg *pdf.RadialGradient) error                { return nil }
func (m *mockWriter) PaintSweepBand(sb *pdf.SweepBand) error                          { return nil }
func (m *mockWriter) PrintParagraph(para []*rich_text.RichText, opts options.Options) {}
func (m *mockWriter) PrintRichText(text *rich_text.RichText)                          {}

func (m *mockWriter) Pie(x, y, r, startAngle, endAngle float64, border, fill, reverse bool) error {
	return nil
}

func (m *mockWriter) Path(fn func()) error {
	fn()
	return nil
}

func (m *mockWriter) CurvePoints(points []pdf.Location) error { return nil }

func (m *mockWriter) Rotate(angle, x, y float64, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}

func (m *mockWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	return nil
}

func (m *mockWriter) Rectangle(x, y, w, h float64, b, f bool)                          {}
func (m *mockWriter) Rectangle2(x, y, w, h float64, b, f bool, c []float64, p, r bool) {}
func (m *mockWriter) SetFillColor(v any) colors.Color                                  { return 0 }
func (m *mockWriter) SetFillLinearGradient(lg *pdf.LinearGradient) error               { return nil }
func (m *mockWriter) SetFillRadialGradient(rg *pdf.RadialGradient) error               { return nil }
func (m *mockWriter) ClearFillGradient()                                               {}
func (m *mockWriter) SetLineLinearGradient(lg *pdf.LinearGradient) error               { return nil }
func (m *mockWriter) SetLineRadialGradient(rg *pdf.RadialGradient) error               { return nil }
func (m *mockWriter) ClearLineGradient()                                               {}
func (m *mockWriter) SetLineColor(v colors.Color) colors.Color                         { return 0 }
func (m *mockWriter) SetLineDashPattern(p string) string                               { return "" }
func (m *mockWriter) SetLineSpacing(ls float64) float64                                { return 0 }
func (m *mockWriter) SetLineWidth(w float64)                                           {}
func (m *mockWriter) SetLanguage(language string)                                      {}
func (m *mockWriter) SetSVGBlendMode(pdf.SVGBlendMode) pdf.SVGBlendMode {
	return pdf.SVGBlendModeRespect
}
func (m *mockWriter) SetSVGGradientStopOpacityMode(pdf.SVGGradientStopOpacityMode) pdf.SVGGradientStopOpacityMode {
	return pdf.SVGGradientStopOpacityModeSoftMask
}
func (m *mockWriter) SetStrikeout(s bool) bool { return false }
func (m *mockWriter) SetUnderline(u bool) bool { return false }

func (m *mockWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	return nil
}

func (m *mockWriter) Stroke() error              { return nil }
func (m *mockWriter) Strikeout() bool            { return false }
func (m *mockWriter) Underline() bool            { return false }
func (m *mockWriter) EnableTaggedPDF(value bool) {}
func (m *mockWriter) TaggedPDFEnabled() bool     { return false }

func (m *mockWriter) WithAccessibilityArtifact(fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}

func (m *mockWriter) WithAccessibilityTag(tag string, opts pdf.AccessibilityOptions, fn func()) error {
	if fn != nil {
		fn()
	}
	return nil
}

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
