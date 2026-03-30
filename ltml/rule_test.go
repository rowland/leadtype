// Copyright 2017 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import "testing"

// ----------------------------------------------------------------------------
// attrsMapFromString
// ----------------------------------------------------------------------------

func TestAttrsMapFromString_single_pair(t *testing.T) {
	attrs := attrsMapFromString("font.size: 12;")
	if attrs["font.size"] != "12" {
		t.Errorf("expected font.size=12, got %q", attrs["font.size"])
	}
}

func TestAttrsMapFromString_multiple_pairs(t *testing.T) {
	attrs := attrsMapFromString("font.size: 14; font.weight: Bold;")
	if attrs["font.size"] != "14" {
		t.Errorf("expected font.size=14, got %q", attrs["font.size"])
	}
	if attrs["font.weight"] != "Bold" {
		t.Errorf("expected font.weight=Bold, got %q", attrs["font.weight"])
	}
}

// Values are trimmed of surrounding whitespace; keys are not — leading or
// trailing spaces in a key name would produce a key with those spaces included.
// In practice, rule declarations in LTML do not have extra whitespace in key names.
func TestAttrsMapFromString_trims_value_whitespace(t *testing.T) {
	attrs := attrsMapFromString("font.size:  14  ; margin:  5  ")
	if attrs["font.size"] != "14" {
		t.Errorf("expected trimmed value 14, got %q", attrs["font.size"])
	}
	if attrs["margin"] != "5" {
		t.Errorf("expected trimmed value 5, got %q", attrs["margin"])
	}
}

func TestAttrsMapFromString_trailing_semicolon_optional(t *testing.T) {
	withSemi := attrsMapFromString("font.size: 12;")
	withoutSemi := attrsMapFromString("font.size: 12")
	if withSemi["font.size"] != withoutSemi["font.size"] {
		t.Errorf("trailing semicolon should be optional: %q vs %q", withSemi["font.size"], withoutSemi["font.size"])
	}
}

// ----------------------------------------------------------------------------
// NewRule
// ----------------------------------------------------------------------------

func TestNewRule_stores_selector(t *testing.T) {
	rule := NewRule("p", map[string]string{"font.size": "12"}, 0, 0)
	if rule.Selector != "p" {
		t.Errorf("expected selector %q, got %q", "p", rule.Selector)
	}
}

func TestNewRule_stores_attrs(t *testing.T) {
	attrs := map[string]string{"font.size": "12", "font.weight": "Bold"}
	rule := NewRule("p", attrs, 0, 0)
	if rule.Attrs["font.size"] != "12" {
		t.Errorf("expected font.size=12, got %q", rule.Attrs["font.size"])
	}
	if rule.Attrs["font.weight"] != "Bold" {
		t.Errorf("expected font.weight=Bold, got %q", rule.Attrs["font.weight"])
	}
}

func TestNewRule_compiles_selector_regexp(t *testing.T) {
	rule := NewRule("p", nil, 0, 0)
	if rule.SelectorRegexp == nil {
		t.Error("expected non-nil SelectorRegexp")
	}
}

func TestNewRule_regexp_matches_selector(t *testing.T) {
	rule := NewRule("p", nil, 0, 0)
	if !rule.SelectorRegexp.MatchString("p") {
		t.Error("rule regexp should match its own selector tag")
	}
}

func TestNewRule_stores_specificity_tier_and_order(t *testing.T) {
	rule := NewRule("div.notice > p#hero.summary", nil, 4, 7)
	if rule.Tier != 4 {
		t.Fatalf("expected tier 4, got %d", rule.Tier)
	}
	if rule.Order != 7 {
		t.Fatalf("expected order 7, got %d", rule.Order)
	}
	want := Specificity{IDs: 1, Classes: 2, Tags: 2}
	if rule.Specificity != want {
		t.Fatalf("expected specificity %+v, got %+v", want, rule.Specificity)
	}
}

// ----------------------------------------------------------------------------
// Rules.AddText
// ----------------------------------------------------------------------------

func TestRules_AddText_parses_single_rule(t *testing.T) {
	var r Rules
	r.AddText("p { font.size: 12; }")
	if len(r.rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(r.rules))
	}
	if r.rules[0].Selector != "p" {
		t.Errorf("expected selector %q, got %q", "p", r.rules[0].Selector)
	}
	if r.rules[0].Attrs["font.size"] != "12" {
		t.Errorf("expected font.size=12, got %q", r.rules[0].Attrs["font.size"])
	}
}

func TestRules_AddText_parses_multiple_rules(t *testing.T) {
	var r Rules
	r.AddText("p { font.size: 12; } h1 { font.size: 24; }")
	if len(r.rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(r.rules))
	}
}

func TestRules_AddText_accumulates_across_calls(t *testing.T) {
	var r Rules
	r.AddText("p { font.size: 12; }")
	r.AddText("h1 { font.size: 24; }")
	if len(r.rules) != 2 {
		t.Fatalf("expected 2 accumulated rules, got %d", len(r.rules))
	}
}

func TestRules_AddText_empty_string_adds_no_rules(t *testing.T) {
	var r Rules
	r.AddText("")
	if len(r.rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(r.rules))
	}
}

func TestRules_AddText_ignores_css_comments(t *testing.T) {
	var r Rules
	r.AddText("/* heading styles */ p { font.size: 12; } /* tail */")
	if len(r.rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(r.rules))
	}
	if r.rules[0].Selector != "p" {
		t.Errorf("expected selector %q, got %q", "p", r.rules[0].Selector)
	}
	if r.rules[0].Attrs["font.size"] != "12" {
		t.Errorf("expected font.size=12, got %q", r.rules[0].Attrs["font.size"])
	}
}

func TestRules_AddText_ignores_css_comments_inside_rule_body(t *testing.T) {
	var r Rules
	r.AddText("p { /* size */ font.size: 12; /* weight */ font.weight: Bold; }")
	if len(r.rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(r.rules))
	}
	if r.rules[0].Attrs["font.size"] != "12" {
		t.Errorf("expected font.size=12, got %q", r.rules[0].Attrs["font.size"])
	}
	if r.rules[0].Attrs["font.weight"] != "Bold" {
		t.Errorf("expected font.weight=Bold, got %q", r.rules[0].Attrs["font.weight"])
	}
}

func TestRules_AddText_splits_grouped_selectors(t *testing.T) {
	var r Rules
	r.AddText("p, span { font.size: 12; }")
	if len(r.rules) != 2 {
		t.Fatalf("expected 2 rules from grouped selector, got %d", len(r.rules))
	}
	if r.rules[0].Selector != "p" || r.rules[1].Selector != "span" {
		t.Fatalf("expected selectors p and span, got %q and %q", r.rules[0].Selector, r.rules[1].Selector)
	}
}

func TestRules_SetAttrs_parses_tier(t *testing.T) {
	var r Rules
	r.SetAttrs(map[string]string{"tier": "4"})
	if err := r.ensureTier(0); err != nil {
		t.Fatalf("ensureTier: %v", err)
	}
	if r.tier != 4 {
		t.Fatalf("expected tier 4, got %d", r.tier)
	}
}

func TestRules_SetAttrs_rejects_negative_tier(t *testing.T) {
	var r Rules
	r.SetAttrs(map[string]string{"tier": "-1"})
	if err := r.ensureTier(0); err == nil {
		t.Fatal("expected invalid tier error")
	}
}

func TestRules_EnsureTier_defaults_when_unspecified(t *testing.T) {
	var r Rules
	r.AddText("p { font.size: 12; }")
	if err := r.ensureTier(3); err != nil {
		t.Fatalf("ensureTier: %v", err)
	}
	if r.tier != 3 {
		t.Fatalf("expected default tier 3, got %d", r.tier)
	}
	if r.rules[0].Tier != 3 {
		t.Fatalf("expected rule tier 3, got %d", r.rules[0].Tier)
	}
}

// ----------------------------------------------------------------------------
// Rules.AddComment
// ----------------------------------------------------------------------------

// AddComment delegates to AddText so that rule declarations can appear inside
// XML comments (<!-- p { font.size: 12 } -->) and still be parsed.
func TestRules_AddComment_parses_rules_from_xml_comment(t *testing.T) {
	var r Rules
	r.AddComment("p { font.size: 12; }")
	if len(r.rules) != 1 {
		t.Fatalf("expected 1 rule from comment, got %d", len(r.rules))
	}
	if r.rules[0].Selector != "p" {
		t.Errorf("expected selector %q, got %q", "p", r.rules[0].Selector)
	}
}

// ----------------------------------------------------------------------------
// Scope.EachRuleFor
// ----------------------------------------------------------------------------

func TestScope_EachRuleFor_yields_matching_rules(t *testing.T) {
	var r Rules
	r.AddText("p { font.size: 12; }")

	var scope Scope
	scope.AddRules(&r)

	var matched []*Rule
	scope.EachRuleFor("p", func(rule *Rule) { matched = append(matched, rule) })

	if len(matched) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matched))
	}
	if matched[0].Attrs["font.size"] != "12" {
		t.Errorf("expected font.size=12, got %q", matched[0].Attrs["font.size"])
	}
}

func TestScope_EachRuleFor_no_match(t *testing.T) {
	var r Rules
	r.AddText("p { font.size: 12; }")

	var scope Scope
	scope.AddRules(&r)

	var matched []*Rule
	scope.EachRuleFor("h1", func(rule *Rule) { matched = append(matched, rule) })

	if len(matched) != 0 {
		t.Errorf("expected no matches for h1, got %d", len(matched))
	}
}

func TestScope_EachRuleFor_multiple_matching_rules(t *testing.T) {
	var r Rules
	r.AddText("p { font.size: 12; } p { font.weight: Bold; }")

	var scope Scope
	scope.AddRules(&r)

	var matched []*Rule
	scope.EachRuleFor("p", func(rule *Rule) { matched = append(matched, rule) })

	if len(matched) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matched))
	}
}

// Parent scope rules remain visible to child scopes.
func TestScope_EachRuleFor_parent_rules_are_visible(t *testing.T) {
	var parentRules Rules
	parentRules.AddText("p { font.size: 10; }")

	var childRules Rules
	childRules.AddText("p { font.size: 14; }")

	var parent Scope
	parent.AddRules(&parentRules)

	var child Scope
	child.SetParentScope(&parent)
	child.AddRules(&childRules)

	var sizes []string
	child.EachRuleFor("p", func(rule *Rule) {
		sizes = append(sizes, rule.Attrs["font.size"])
	})

	if len(sizes) != 2 {
		t.Fatalf("expected 2 matches (parent + child), got %d", len(sizes))
	}
	if sizes[0] != "10" || sizes[1] != "14" {
		t.Fatalf("expected matching rule values 10 then 14, got %v", sizes)
	}
}

func TestScope_EachRuleFor_sorts_by_tier_then_specificity_then_order(t *testing.T) {
	var parentRules Rules
	parentRules.SetAttrs(map[string]string{"tier": "4"})
	parentRules.AddText("p { font.size: 10; }")

	var childRules Rules
	childRules.SetAttrs(map[string]string{"tier": "4"})
	childRules.AddText("p.notice { font.size: 14; } p.notice { font.size: 18; }")

	var parent Scope
	parent.defaultRuleTier = 0
	if err := parent.AddRules(&parentRules); err != nil {
		t.Fatalf("AddRules parent: %v", err)
	}

	var child Scope
	child.defaultRuleTier = 1
	child.SetParentScope(&parent)
	if err := child.AddRules(&childRules); err != nil {
		t.Fatalf("AddRules child: %v", err)
	}

	var sizes []string
	child.EachRuleFor("p.notice", func(rule *Rule) {
		sizes = append(sizes, rule.Attrs["font.size"])
	})

	if len(sizes) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(sizes))
	}
	if sizes[0] != "10" || sizes[1] != "14" || sizes[2] != "18" {
		t.Fatalf("expected cascade order [10 14 18], got %v", sizes)
	}
}

func TestScope_EachRuleFor_class_selector(t *testing.T) {
	var r Rules
	r.AddText("p.intro { font.weight: Bold; }")

	var scope Scope
	scope.AddRules(&r)

	var matched []*Rule
	scope.EachRuleFor("p.intro", func(rule *Rule) { matched = append(matched, rule) })
	if len(matched) != 1 {
		t.Fatalf("expected match for p.intro, got %d", len(matched))
	}

	matched = nil
	scope.EachRuleFor("p.other", func(rule *Rule) { matched = append(matched, rule) })
	if len(matched) != 0 {
		t.Errorf("expected no match for p.other, got %d", len(matched))
	}
}

func TestScope_EachRuleFor_descendant_selector(t *testing.T) {
	var r Rules
	r.AddText("div p { font.size: 11; }")

	var scope Scope
	scope.AddRules(&r)

	var matched []*Rule
	scope.EachRuleFor("div/p", func(rule *Rule) { matched = append(matched, rule) })
	if len(matched) != 1 {
		t.Errorf("expected descendant match for div/p, got %d", len(matched))
	}

	matched = nil
	scope.EachRuleFor("p", func(rule *Rule) { matched = append(matched, rule) })
	if len(matched) != 0 {
		t.Errorf("expected no match for bare p (not inside div), got %d", len(matched))
	}
}
