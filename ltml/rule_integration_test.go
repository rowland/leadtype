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
