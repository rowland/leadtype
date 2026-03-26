// Copyright 2017 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

// Integration tests for the rules system. Each test parses a small LTML
// document, then walks the element tree to verify that rules were (or were not)
// applied to the expected elements.

import "testing"

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
	if len(doc.ltmls) == 0 {
		t.Fatal("no ltml elements in document")
	}
	page := doc.ltmls[0].Page(0)
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

// ----------------------------------------------------------------------------
// Font size set by a tag rule
// ----------------------------------------------------------------------------

func TestRules_integration_tag_rule_sets_font_size(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<rules>p { font.size: 14; }</rules>
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

// ----------------------------------------------------------------------------
// Font weight set by a class rule
// ----------------------------------------------------------------------------

func TestRules_integration_class_rule_sets_font_weight(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<rules>p.bold { font.weight: Bold; }</rules>
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
			<rules>p.special { font.size: 20; }</rules>
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
			<rules>p { font.size: 14; }</rules>
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

// ----------------------------------------------------------------------------
// Rules inside XML comments are also parsed and applied
// ----------------------------------------------------------------------------

func TestRules_integration_comment_rule_applies(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<rules><!-- p { font.size: 16; } --></rules>
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
			<rules>p { font.size: 10; } p { font.size: 18; }</rules>
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

// ----------------------------------------------------------------------------
// Descendant selector: rule only applies when element is inside the right parent
// ----------------------------------------------------------------------------

func TestRules_integration_descendant_rule_applies(t *testing.T) {
	// The rule targets "div p" — a <p> inside a <div>.
	doc := parseDoc(t, `
		<ltml>
			<rules>div p { font.size: 15; }</rules>
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
			<rules>div p { font.size: 15; }</rules>
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
			<rules>div>p { font.size: 17; }</rules>
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
			<rules>div>p { font.size: 17; }</rules>
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
			<rules>p#intro.highlight { font.size: 22; }</rules>
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
			<rules>p#intro.highlight { font.size: 22; }</rules>
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
			<rules>h { font.size: 24; }</rules>
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
			<rules>p { font.size: 14; }</rules>
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
			<rules>caption { font.size: 10; }</rules>
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

// ----------------------------------------------------------------------------
// Multiple properties in a single rule are all applied
// ----------------------------------------------------------------------------

func TestRules_integration_rule_with_multiple_attrs(t *testing.T) {
	doc := parseDoc(t, `
		<ltml>
			<rules>p { font.size: 13; font.weight: Bold; }</rules>
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
