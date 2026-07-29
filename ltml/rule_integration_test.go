// Copyright 2017 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

// Integration tests for the rules system. Each test parses a small LTML
// document, then walks the element tree to verify that rules were (or were not)
// applied to the expected elements.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// parseDoc is a helper that parses an LTML document from a string and fatals
// the test on any error.
func parseDoc(t *testing.T, src string) *Doc {
	t.Helper()
	doc, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return doc
}

// firstPage returns the first page of the first ltml element in doc.
func firstPage(t *testing.T, doc *Doc) *StdPage {
	t.Helper()
	if doc.Root() == nil {
		t.Fatal("no ltml elements in document")
	}
	page := doc.Root().Page(0)
	if page == nil {
		t.Fatal("no pages in document")
	}
	return page
}

// firstParagraph returns the first paragraph on the first page of doc.
func firstParagraph(t *testing.T, doc *Doc) *StdParagraph {
	t.Helper()
	page := firstPage(t, doc)
	if len(page.children) == 0 {
		t.Fatal("no children on first page")
	}
	p, ok := page.children[0].(*StdParagraph)
	if !ok {
		t.Fatalf("first child is %T, not *StdParagraph", page.children[0])
	}
	return p
}

func firstContainer(t *testing.T, doc *Doc) *StdContainer {
	t.Helper()
	page := firstPage(t, doc)
	if len(page.children) == 0 {
		t.Fatal("no children on first page")
	}
	c, ok := page.children[0].(*StdContainer)
	if !ok {
		t.Fatalf("first child is %T, not *StdContainer", page.children[0])
	}
	return c
}

func childParagraph(t *testing.T, container *StdContainer, index int) *StdParagraph {
	t.Helper()
	if index >= len(container.children) {
		t.Fatalf("container child index %d out of range", index)
	}
	p, ok := container.children[index].(*StdParagraph)
	if !ok {
		t.Fatalf("child %d is %T, not *StdParagraph", index, container.children[index])
	}
	return p
}

func TestRules_integration_inline_style_reports_invalid_selector(t *testing.T) {
	_, err := Parse([]byte(`
		<ltml>
			<style>
				.bar-graph re2:scorebar { width: 66%; }
				.score-badge { width: 22pt; height: 22pt; }
			</style>
			<page />
		</ltml>`))
	if err == nil {
		t.Fatal("Parse error = nil, want invalid selector error")
	}
	if !strings.Contains(err.Error(), `unknown pseudo-class "scorebar"`) {
		t.Fatalf("Parse error = %q, want unknown pseudo-class error", err)
	}
}

// ----------------------------------------------------------------------------
// Font size set by a tag rule
// ----------------------------------------------------------------------------

func TestRules_integration_tag_rule_sets_font_size(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p { font.size: 14; }</style>
			<page><p>hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph by rule")
	}
	if p.font.size != 14 {
		t.Errorf("expected font size 14, got %v", p.font.size)
	}
}

func TestRules_integration_style_tag_sets_font_size(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p { font.size: 14; }</style>
			<page><p>hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph by style tag")
	}
	if p.font.size != 14 {
		t.Errorf("expected font size 14, got %v", p.font.size)
	}
}

// ----------------------------------------------------------------------------
// Font weight set by a class rule
// ----------------------------------------------------------------------------

func TestRules_integration_class_rule_sets_font_weight(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p.bold { font.weight: Bold; }</style>
			<page>
				<p class="bold">bold text</p>
			</page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph by rule")
	}
	if p.font.weight != "Bold" {
		t.Errorf("expected font weight Bold, got %q", p.font.weight)
	}
}

// ----------------------------------------------------------------------------
// Class rule does not affect elements without the matching class
// ----------------------------------------------------------------------------

func TestRules_integration_class_rule_does_not_affect_unmatched_element(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p.special { font.size: 20; }</style>
			<page><p>plain paragraph</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	// The rule should not have touched this paragraph's font.
	if p.font != nil && p.font.size == 20 {
		t.Error("class rule should not have applied to a paragraph without that class")
	}
}

// ----------------------------------------------------------------------------
// Direct XML attributes take precedence over rule attributes
// ----------------------------------------------------------------------------

func TestRules_integration_direct_attrs_override_rule(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p { font.size: 14; }</style>
			<page><p font.size="20">hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph")
	}
	if p.font.size != 20 {
		t.Errorf("direct attribute (font.size=20) should override rule (font.size=14), got %v", p.font.size)
	}
}

func TestRules_integration_page_default_tier_beats_document_default_tier(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p { font.size: 14; }</style>
			<page>
				<style>p { font.size: 20; }</style>
				<p>hello</p>
			</page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph")
	}
	if p.font.size != 20 {
		t.Errorf("expected page default tier to win with font.size=20, got %v", p.font.size)
	}
}

func TestRules_integration_document_override_tier_beats_page_default_tier(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style tier="4">p { font.size: 22; }</style>
			<page>
				<style>p { font.size: 18; }</style>
				<p>hello</p>
			</page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph")
	}
	if p.font.size != 22 {
		t.Errorf("expected document override tier to win with font.size=22, got %v", p.font.size)
	}
}

func TestRules_integration_style_tag_sets_font_size_without_tier(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p { font.size: 16; }</style>
			<page><p>hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph by style tag")
	}
	if p.font.size != 16 {
		t.Errorf("expected font size 16, got %v", p.font.size)
	}
}

func TestRules_integration_style_src_loadsFromAssetFS(t *testing.T) {
	doc, err := Parse([]byte(`
		<ltml>
			<style src="styles.ltml.css" />
			<page><p>hello</p></page>
		</ltml>`), WithAssetFS(fstest.MapFS{
		"styles.ltml.css": &fstest.MapFile{Data: []byte("p { font.size: 17; }")},
	}))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph by style src")
	}
	if p.font.size != 17 {
		t.Fatalf("expected font size 17, got %v", p.font.size)
	}
}

func TestRules_integration_style_src_loadsRelativeToParsedFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "styles.ltml.css"), []byte("p { font.size: 19; }"), 0o644); err != nil {
		t.Fatal(err)
	}
	inputPath := filepath.Join(dir, "doc.ltml")
	if err := os.WriteFile(inputPath, []byte(`
<ltml>
  <style src="styles.ltml.css" />
  <page><p>hello</p></page>
</ltml>`), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := ParseFile(inputPath)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph by relative style src")
	}
	if p.font.size != 19 {
		t.Fatalf("expected font size 19, got %v", p.font.size)
	}
}

func TestRules_integration_style_src_overridesInlineBody(t *testing.T) {
	doc, err := Parse([]byte(`
		<ltml>
			<style src="styles.ltml.css">p { font.size: 11; }</style>
			<page><p>hello</p></page>
		</ltml>`), WithAssetFS(fstest.MapFS{
		"styles.ltml.css": &fstest.MapFile{Data: []byte("p { font.size: 21; }")},
	}))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph by style src")
	}
	if p.font.size != 21 {
		t.Fatalf("expected src to win with font size 21, got %v", p.font.size)
	}
}

func TestRules_integration_style_src_missingFileReturnsHelpfulError(t *testing.T) {
	_, err := Parse([]byte(`
		<ltml>
			<style src="missing.ltml.css" />
			<page><p>hello</p></page>
		</ltml>`), WithAssetFS(fstest.MapFS{}))
	if err == nil {
		t.Fatal("Parse succeeded, want error")
	}
	if !strings.Contains(err.Error(), `loading style src "missing.ltml.css"`) {
		t.Fatalf("error = %q, want style src context", err)
	}
}

func TestRules_styleSrcMemoizesLoadedText(t *testing.T) {
	doc := newDocWithOptions(WithAssetFS(fstest.MapFS{
		"styles.ltml.css": &fstest.MapFile{Data: []byte("p { font.size: 23; }")},
	}))
	doc.root = &StdDocument{}

	rules := &Rules{}
	rules.SetDoc(doc)
	rules.SetAttrs(map[string]string{"src": "styles.ltml.css"})

	var scope Scope
	if err := scope.AddRules(rules); err != nil {
		t.Fatalf("AddRules: %v", err)
	}
	if !rules.sourceLoaded {
		t.Fatal("expected source text to be loaded")
	}
	if rules.sourceText == "" {
		t.Fatal("expected style src text to be cached")
	}

	doc.SetAssetFS(fstest.MapFS{})
	if err := rules.ensureSourceLoaded(); err != nil {
		t.Fatalf("ensureSourceLoaded after asset removal: %v", err)
	}
	if rules.sourceText != "p { font.size: 23; }" {
		t.Fatalf("cached sourceText = %q, want original content", rules.sourceText)
	}
}

func TestRules_integration_rules_tag_is_not_registered(t *testing.T) {
	if got := makeElement(DefaultSpace, "rules"); got != nil {
		t.Fatalf("makeElement(%q, %q) = %T, want nil", DefaultSpace, "rules", got)
	}
}

func TestRules_integration_higher_document_tier_beats_lower_document_tier(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style tier="4">p { font.size: 18; }</style>
			<style tier="5">p { font.size: 24; }</style>
			<page><p>hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph")
	}
	if p.font.size != 24 {
		t.Errorf("expected higher document tier to win with font.size=24, got %v", p.font.size)
	}
}

// ----------------------------------------------------------------------------
// Rules inside XML comments are also parsed and applied
// ----------------------------------------------------------------------------

func TestRules_integration_comment_rule_applies(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style><!-- p { font.size: 16; } --></style>
			<page><p>hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set; comment-based rule was not applied")
	}
	if p.font.size != 16 {
		t.Errorf("expected font size 16 from comment rule, got %v", p.font.size)
	}
}

// ----------------------------------------------------------------------------
// Later rule declarations override earlier ones for the same property
// ----------------------------------------------------------------------------

func TestRules_integration_later_rule_wins(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p { font.size: 10; } p { font.size: 18; }</style>
			<page><p>hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph")
	}
	if p.font.size != 18 {
		t.Errorf("expected later rule (font.size=18) to win over earlier rule (font.size=10), got %v", p.font.size)
	}
}

func TestRules_integration_first_and_last_child_pseudos_apply(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>
				p:first-child { font.size: 14; }
				p:last-child { font.weight: Bold; }
			</style>
			<page>
				<div>
					<p>first</p>
					<p>middle</p>
					<p>last</p>
				</div>
			</page>
		</ltml>`)

	container := firstContainer(t, doc)
	first := childParagraph(t, container, 0)
	middle := childParagraph(t, container, 1)
	last := childParagraph(t, container, 2)

	if first.font == nil || first.font.size != 14 {
		t.Fatalf("first child font size = %#v, want 14", first.font)
	}
	if middle.font != nil && middle.font.size == 14 {
		t.Fatal("middle child should not match :first-child")
	}
	if last.font == nil || last.font.weight != "Bold" {
		t.Fatalf("last child font = %#v, want Bold weight", last.font)
	}
}

func TestRules_integration_direction_pseudos_match_effective_direction(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>
				page.ar { dir: rtl; }
				page:dir(rtl) > .outer { alt: rtl-parent; }
				.outer:dir(rtl) > label:dir(rtl) { alt: inherited-rtl; z-index: 2; }
				.override:dir(ltr) > label:dir(ltr) { alt: nested-ltr; z-index: 3; }
			</style>
			<page class="ar">
				<div class="outer">
					<label>rtl leaf</label>
					<div class="override" dir="ltr"><label>ltr leaf</label></div>
				</div>
			</page>
		</ltml>`)

	page := firstPage(t, doc)
	outer, ok := page.children[0].(*StdContainer)
	if !ok {
		t.Fatalf("outer child = %T, want *StdContainer", page.children[0])
	}
	if page.Dir() != DirRTL || outer.Dir() != DirRTL || outer.alt != "rtl-parent" {
		t.Fatalf("page/outer direction and match = %s/%s/%q, want rtl/rtl/rtl-parent", page.Dir(), outer.Dir(), outer.alt)
	}
	rtlLeaf, ok := outer.children[0].(*StdLabel)
	if !ok {
		t.Fatalf("RTL leaf = %T, want *StdLabel", outer.children[0])
	}
	if rtlLeaf.alt != "inherited-rtl" || rtlLeaf.zIndex != 2 {
		t.Fatalf("RTL leaf match = %q/%d, want inherited-rtl/2", rtlLeaf.alt, rtlLeaf.zIndex)
	}
	override, ok := outer.children[1].(*StdContainer)
	if !ok {
		t.Fatalf("override child = %T, want *StdContainer", outer.children[1])
	}
	ltrLeaf, ok := override.children[0].(*StdLabel)
	if !ok {
		t.Fatalf("LTR leaf = %T, want *StdLabel", override.children[0])
	}
	if override.Dir() != DirLTR || ltrLeaf.alt != "nested-ltr" || ltrLeaf.zIndex != 3 {
		t.Fatalf("LTR override match = %s/%q/%d, want ltr/nested-ltr/3", override.Dir(), ltrLeaf.alt, ltrLeaf.zIndex)
	}
}

func TestRules_integration_table_row_and_col_pseudos_apply(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>
				p:first-row { font.size: 10; }
				p:last-row { font.weight: Bold; }
				p:first-col { font.style: Italic; }
				p:last-col { font.size: 20; }
				p:row-even { z-index: 2; }
				p:row-odd { z-index: 3; }
				p:col-even { alt: col-even; }
				p:col-odd { alt: col-odd; }
				p:row-1 { display: odd; }
				p:col-1 { display: always; }
			</style>
			<page>
				<div layout="table" cols="2" width="4">
					<p>a</p>
					<p>b</p>
					<p>c</p>
					<p>d</p>
				</div>
			</page>
		</ltml>`)

	container := firstContainer(t, doc)
	a := childParagraph(t, container, 0)
	b := childParagraph(t, container, 1)
	c := childParagraph(t, container, 2)
	d := childParagraph(t, container, 3)

	if a.font == nil || a.font.size != 10 || a.font.style != "Italic" {
		t.Fatalf("cell a font = %#v, want first-row + first-col styles", a.font)
	}
	if a.zIndex != 2 || a.alt != "col-even" {
		t.Fatalf("cell a z-index/alt = %d/%q, want 2/col-even", a.zIndex, a.alt)
	}
	if b.font == nil || b.font.size != 20 {
		t.Fatalf("cell b font = %#v, want last-col size 20", b.font)
	}
	if b.display != DisplayAlways || b.alt != "col-odd" {
		t.Fatalf("cell b display/alt = %s/%q, want always/col-odd", b.display, b.alt)
	}
	if c.zIndex != 3 || c.display != DisplayOdd {
		t.Fatalf("cell c z-index/display = %d/%s, want 3/odd", c.zIndex, c.display)
	}
	if d.font == nil || d.font.weight != "Bold" {
		t.Fatalf("cell d font = %#v, want Bold from :last-row", d.font)
	}
}

func TestRules_integration_display_none_overrides_repeating_display(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>
				.repeating { display: succeeding; }
				.hidden { display: none; }
			</style>
			<page>
				<div>
					<p class="repeating hidden">hidden</p>
				</div>
			</page>
		</ltml>`)

	paragraph := childParagraph(t, firstContainer(t, doc), 0)
	if paragraph.display != DisplayNone {
		t.Fatalf("display = %s, want none", paragraph.display)
	}
	if widgetDisplayForRender(paragraph, false, 2, 2) {
		t.Fatal("display:none widget should not participate in layout or rendering")
	}
}

func TestRules_integration_direct_attrs_override_pseudo_rules(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p:first-child { font.size: 14; }</style>
			<page>
				<div>
					<p font.size="22">hello</p>
					<p>other</p>
				</div>
			</page>
		</ltml>`)

	container := firstContainer(t, doc)
	first := childParagraph(t, container, 0)
	if first.font == nil || first.font.size != 22 {
		t.Fatalf("first child font size = %#v, want direct attr override 22", first.font)
	}
}

func TestRules_integration_row_and_col_pseudos_do_not_apply_outside_tables(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>
				p:first-row { font.size: 14; }
				p:col-0 { font.weight: Bold; }
			</style>
			<page>
				<div>
					<p>hello</p>
				</div>
			</page>
		</ltml>`)

	container := firstContainer(t, doc)
	first := childParagraph(t, container, 0)
	if first.font != nil && (first.font.size == 14 || first.font.weight == "Bold") {
		t.Fatalf("non-table child font = %#v, row/col pseudos should not apply", first.font)
	}
}

func TestRules_integration_row_and_col_pseudos_use_anchor_cell_for_spans(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>
				p:first-col { font.style: Italic; }
				p:col-1 { font.weight: Bold; }
				p:last-row { font.size: 18; }
			</style>
			<page>
				<div layout="table" cols="2" width="4">
					<p rowspan="2">a</p>
					<p>b</p>
					<p>c</p>
				</div>
			</page>
		</ltml>`)

	container := firstContainer(t, doc)
	a := childParagraph(t, container, 0)
	b := childParagraph(t, container, 1)
	c := childParagraph(t, container, 2)

	if a.font == nil || a.font.style != "Italic" {
		t.Fatalf("spanning cell font = %#v, want first-col match from anchor cell", a.font)
	}
	if a.font.weight == "Bold" || a.font.size == 18 {
		t.Fatalf("spanning cell font = %#v, should not match col-1 or last-row from covered cells", a.font)
	}
	if b.font == nil || b.font.weight != "Bold" {
		t.Fatalf("cell b font = %#v, want col-1 match", b.font)
	}
	if c.font == nil || c.font.size != 18 {
		t.Fatalf("cell c font = %#v, want last-row match", c.font)
	}
}

func TestRules_integration_more_specific_rule_wins_within_tier(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p { font.size: 10; } p.intro { font.size: 16; }</style>
			<page><p class="intro">hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph")
	}
	if p.font.size != 16 {
		t.Errorf("expected more specific p.intro rule to win with font.size=16, got %v", p.font.size)
	}
}

func TestRules_integration_id_rule_wins_over_class_rule_within_tier(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p.intro { font.size: 16; } p#hero { font.size: 19; }</style>
			<page><p id="hero" class="intro">hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph")
	}
	if p.font.size != 19 {
		t.Errorf("expected id selector to win with font.size=19, got %v", p.font.size)
	}
}

func TestRules_integration_descendant_rule_beats_tag_rule_on_tag_count_tiebreak(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p { font.size: 10; } div p { font.size: 15; }</style>
			<page>
				<div><p>nested paragraph</p></div>
			</page>
		</ltml>`)

	page := firstPage(t, doc)
	div := page.children[0].(*StdContainer)
	p := div.children[0].(*StdParagraph)
	if p.font == nil {
		t.Fatal("font was not set on paragraph")
	}
	if p.font.size != 15 {
		t.Errorf("expected descendant selector to win with font.size=15, got %v", p.font.size)
	}
}

// ----------------------------------------------------------------------------
// Descendant selector: rule only applies when element is inside the right parent
// ----------------------------------------------------------------------------

func TestRules_integration_descendant_rule_applies(t *testing.T) {
	// The rule targets "div p" — a <p> inside a <div>.
	doc := parseDoc(t, `
		<ltml>
			<style>div p { font.size: 15; }</style>
			<page>
				<div><p>nested paragraph</p></div>
			</page>
		</ltml>`)

	page := firstPage(t, doc)
	if len(page.children) == 0 {
		t.Fatal("no children on page")
	}
	div, ok := page.children[0].(*StdContainer)
	if !ok {
		t.Fatalf("expected *StdContainer (div), got %T", page.children[0])
	}
	if len(div.children) == 0 {
		t.Fatal("no children in div")
	}
	p, ok := div.children[0].(*StdParagraph)
	if !ok {
		t.Fatalf("expected *StdParagraph inside div, got %T", div.children[0])
	}
	if p.font == nil {
		t.Fatal("font was not set on nested paragraph by descendant rule")
	}
	if p.font.size != 15 {
		t.Errorf("expected font size 15 from descendant rule, got %v", p.font.size)
	}
}

func TestRules_integration_descendant_rule_does_not_apply_without_ancestor(t *testing.T) {
	// The rule targets "div p" — a bare <p> at page level should not match.
	doc := parseDoc(t, `
		<ltml>
			<style>div p { font.size: 15; }</style>
			<page><p>bare paragraph</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font != nil && p.font.size == 15 {
		t.Error("descendant rule should not apply to a <p> not inside a <div>")
	}
}

// ----------------------------------------------------------------------------
// Direct child selector (>) — only matches an immediate child, not a deeper
// descendant
// ----------------------------------------------------------------------------

func TestRules_integration_direct_child_rule_applies_to_immediate_child(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>div>p { font.size: 17; }</style>
			<page>
				<div><p>direct child</p></div>
			</page>
		</ltml>`)

	page := firstPage(t, doc)
	div, ok := page.children[0].(*StdContainer)
	if !ok {
		t.Fatalf("expected *StdContainer (div), got %T", page.children[0])
	}
	p, ok := div.children[0].(*StdParagraph)
	if !ok {
		t.Fatalf("expected *StdParagraph, got %T", div.children[0])
	}
	if p.font == nil {
		t.Fatal("font was not set on direct child paragraph")
	}
	if p.font.size != 17 {
		t.Errorf("expected font.size=17 from direct child rule, got %v", p.font.size)
	}
}

func TestRules_integration_direct_child_rule_does_not_match_p_inside_p(t *testing.T) {
	// div>p should NOT match a <p> whose direct parent is another <p> (path
	// div/p/p). The inner <p> is a descendant of the <div> but its immediate
	// parent is <p>, so the direct-child relationship with <div> doesn't hold.
	// Note: div>p DOES still match the outer <p> (path div/p) — only the inner
	// one (path div/p/p) should be unaffected.
	doc := parseDoc(t, `
		<ltml>
			<style>div>p { font.size: 17; }</style>
			<page>
				<div><p>outer<p>inner</p></p></div>
			</page>
		</ltml>`)

	page := firstPage(t, doc)
	div, ok := page.children[0].(*StdContainer)
	if !ok {
		t.Fatalf("expected *StdContainer (div), got %T", page.children[0])
	}
	outerP, ok := div.children[0].(*StdParagraph)
	if !ok {
		t.Fatalf("expected outer *StdParagraph, got %T", div.children[0])
	}
	// The outer <p> (div/p) should match.
	if outerP.font == nil || outerP.font.size != 17 {
		t.Errorf("outer p (div/p) should match div>p rule, got font %v", outerP.font)
	}
	// The inner <p> (div/p/p) should NOT match.
	if len(outerP.children) == 0 {
		t.Fatal("outer <p> has no children")
	}
	innerP, ok := outerP.children[0].(*StdParagraph)
	if !ok {
		t.Fatalf("expected inner *StdParagraph, got %T", outerP.children[0])
	}
	if innerP.font != nil && innerP.font.size == 17 {
		t.Error("direct child rule (div>p) should not apply to inner <p> whose direct parent is <p>, not <div>")
	}
}

// ----------------------------------------------------------------------------
// Selector with both id and class — e.g. p#intro.highlight
// ----------------------------------------------------------------------------

func TestRules_integration_id_and_class_selector_matches_exact_element(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p#intro.highlight { font.size: 22; }</style>
			<page>
				<p id="intro" class="highlight">targeted</p>
			</page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on element matching id+class selector")
	}
	if p.font.size != 22 {
		t.Errorf("expected font.size=22, got %v", p.font.size)
	}
}

func TestRules_integration_id_and_class_selector_requires_both(t *testing.T) {
	// Same rule but the element only has the id, not the class — should not match.
	doc := parseDoc(t, `
		<ltml>
			<style>p#intro.highlight { font.size: 22; }</style>
			<page>
				<p id="intro">id only, no class</p>
			</page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font != nil && p.font.size == 22 {
		t.Error("id+class selector should not match an element that has only the id")
	}
}

// ----------------------------------------------------------------------------
// Selectors use the alias name, not the underlying type
//
// Aliases map a user-facing name to a built-in tag (e.g. <h> → <p>).
// The element's Path() is built from the alias name as written in the source,
// so rules must also use that name.  A rule targeting the underlying type does
// not apply to elements written with an alias name.
// ----------------------------------------------------------------------------

// A rule written with the alias name applies to elements using that alias.
func TestRules_integration_builtin_alias_targeted_by_alias_name(t *testing.T) {
	// <h> is a built-in alias for <p>. The rule uses "h", not "p".
	doc := parseDoc(t, `
		<ltml>
			<style>h { font.size: 24; }</style>
			<page><h>heading text</h></page>
		</ltml>`)

	page := firstPage(t, doc)
	if len(page.children) == 0 {
		t.Fatal("no children on page")
	}
	p, ok := page.children[0].(*StdParagraph)
	if !ok {
		t.Fatalf("expected *StdParagraph (h alias), got %T", page.children[0])
	}
	if p.font == nil {
		t.Fatal("font was not set on <h> element by alias-name rule")
	}
	if p.font.size != 24 {
		t.Errorf("expected font.size=24 from h rule, got %v", p.font.size)
	}
}

// A rule targeting the underlying type does NOT apply to elements written with
// an alias name — the selector matches against the name as written in source.
func TestRules_integration_underlying_type_rule_does_not_apply_to_alias(t *testing.T) {
	// <h> is an alias for <p>, but a rule for "p" should not match <h>.
	doc := parseDoc(t, `
		<ltml>
			<style>p { font.size: 14; }</style>
			<page><h>heading text</h></page>
		</ltml>`)

	page := firstPage(t, doc)
	if len(page.children) == 0 {
		t.Fatal("no children on page")
	}
	p, ok := page.children[0].(*StdParagraph)
	if !ok {
		t.Fatalf("expected *StdParagraph (h alias), got %T", page.children[0])
	}
	// The "p" rule must not have set font.size=14 on the <h> element.
	if p.font != nil && p.font.size == 14 {
		t.Error("rule for 'p' should not apply to an element written as <h>")
	}
}

// A user-defined alias (via <define>) can be targeted by its alias name.
func TestRules_integration_user_defined_alias_targeted_by_alias_name(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<define id="caption" tag="p" />
			<style>caption { font.size: 10; }</style>
			<page><caption>fig. 1</caption></page>
		</ltml>`)

	page := firstPage(t, doc)
	if len(page.children) == 0 {
		t.Fatal("no children on page")
	}
	p, ok := page.children[0].(*StdParagraph)
	if !ok {
		t.Fatalf("expected *StdParagraph (caption alias), got %T", page.children[0])
	}
	if p.font == nil {
		t.Fatal("font was not set on <caption> element by user-defined alias rule")
	}
	if p.font.size != 10 {
		t.Errorf("expected font.size=10 from caption rule, got %v", p.font.size)
	}
}

func TestRules_integration_builtin_heading_alias_targeted_by_alias_name(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>h3 { font.size: 2rem; }</style>
			<page font.name="Helvetica" font.size="12"><h3>heading text</h3></page>
		</ltml>`)

	page := firstPage(t, doc)
	if len(page.children) == 0 {
		t.Fatal("no children on page")
	}
	label, ok := page.children[0].(*StdLabel)
	if !ok {
		t.Fatalf("expected *StdLabel (h3 alias), got %T", page.children[0])
	}
	if label.font == nil {
		t.Fatal("font was not set on <h3> element by alias-name rule")
	}
	w := &labelTestWriter{t: t, lineSpacing: 1.0}
	assertAllLeafFontSizesEqual(t, label.RichText(w), 24)
}

// ----------------------------------------------------------------------------
// Multiple properties in a single rule are all applied
// ----------------------------------------------------------------------------

func TestRules_integration_rule_with_multiple_attrs(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p { font.size: 13; font.weight: Bold; }</style>
			<page><p>hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph")
	}
	if p.font.size != 13 {
		t.Errorf("expected font.size=13, got %v", p.font.size)
	}
	if p.font.weight != "Bold" {
		t.Errorf("expected font.weight=Bold, got %q", p.font.weight)
	}
}

func TestRules_integration_grouped_selectors_apply_independently(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>p, span { font.size: 13; }</style>
			<page><p>hello</p></page>
		</ltml>`)

	p := firstParagraph(t, doc)
	if p.font == nil {
		t.Fatal("font was not set on paragraph")
	}
	if p.font.size != 13 {
		t.Errorf("expected grouped selector rule to apply with font.size=13, got %v", p.font.size)
	}
}

func TestRules_integration_universal_child_selector_applies_to_multiple_widget_types(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<style>.box > * { z-index: 7; }</style>
			<page>
				<vbox class="box">
					<p>paragraph</p>
					<label>label</label>
				</vbox>
			</page>
		</ltml>`)

	box := firstContainer(t, doc)
	if len(box.children) != 2 {
		t.Fatalf("box child count = %d, want 2", len(box.children))
	}
	for i, child := range box.children {
		if child.ZIndex() != 7 {
			t.Errorf("child %d (%T) z-index = %d, want 7", i, child, child.ZIndex())
		}
	}
}

func TestRules_integration_invalid_tier_returns_parse_error(t *testing.T) {
	if _, err := Parse([]byte(`
		<ltml>
			<style tier="-1">p { font.size: 14; }</style>
			<page><p>hello</p></page>
		</ltml>`)); err == nil {
		t.Fatal("expected invalid tier to return parse error")
	}
}
