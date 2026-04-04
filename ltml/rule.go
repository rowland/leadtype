// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Specificity struct {
	IDs     int
	Classes int
	Tags    int
}

func (s Specificity) Compare(other Specificity) int {
	if s.IDs != other.IDs {
		return cmpInt(s.IDs, other.IDs)
	}
	if s.Classes != other.Classes {
		return cmpInt(s.Classes, other.Classes)
	}
	return cmpInt(s.Tags, other.Tags)
}

// Rule associates a CSS-like selector with a set of attributes. When an element's
// path matches the selector, the attributes are applied to that element before any
// directly-specified attributes (so direct attributes take precedence).
type Rule struct {
	Selector       string
	SelectorRegexp *regexp.Regexp
	Attrs          map[string]string
	Tier           int
	Specificity    Specificity
	Order          int
}

// NewRule creates a Rule for the given selector string and attribute map. The
// selector is compiled into a regexp once so matching during document layout is
// fast. See selectors.go for supported selector syntax.
func NewRule(selector string, attrs map[string]string, tier, order int) *Rule {
	return &Rule{
		Selector:       selector,
		SelectorRegexp: regexpForSelector(selector),
		Attrs:          attrs,
		Tier:           tier,
		Specificity:    specificityForSelector(selector),
		Order:          order,
	}
}

// Rules is the in-memory representation of a <style> block. It holds zero or
// more Rule values parsed from CSS-like text and is registered with the
// enclosing Scope so that elements can be matched against them during parsing.
type Rules struct {
	rules         []*Rule
	tier          int
	tierExplicit  bool
	nextRuleOrder int
	parseErr      error
}

// AddComment satisfies the XML comment handler interface. Rule declarations may
// appear inside XML comments (<!-- p { font.size: 12 } -->) so that the LTML
// source remains valid XML.
func (r *Rules) AddComment(comment string) {
	r.AddText(comment)
}

var reCSSComment = regexp.MustCompile(`(?s)/\*.*?\*/`)
var reRule = regexp.MustCompile(`\s*([^\{]+?)\s*\{([^\}]+)\}`)

// AddText parses one or more CSS-like rule declarations from text and appends
// them to the Rules collection. Each declaration has the form:
//
//	selector { key: value; key: value }
//
// Multiple declarations may appear in a single call. Whitespace around selectors
// and values is trimmed automatically. CSS-style block comments are ignored.
func (r *Rules) AddText(text string) {
	if r.parseErr != nil {
		return
	}
	text = stripCSSComments(text)
	matches := reRule.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		for _, selector := range splitRuleSelectors(m[1]) {
			r.rules = append(r.rules, NewRule(selector, attrsMapFromString(m[2]), r.tier, r.nextRuleOrder))
			r.nextRuleOrder++
		}
	}
}

func (r *Rules) SetAttrs(attrs map[string]string) {
	tierText, ok := attrs["tier"]
	if !ok || tierText == "" {
		return
	}
	tier, err := strconv.Atoi(strings.TrimSpace(tierText))
	if err != nil {
		r.parseErr = fmt.Errorf("invalid style tier %q", tierText)
		return
	}
	if tier < 0 {
		r.parseErr = fmt.Errorf("invalid style tier %q: tier must be >= 0", tierText)
		return
	}
	r.tier = tier
	r.tierExplicit = true
}

func (r *Rules) ensureTier(defaultTier int) error {
	if r.parseErr != nil {
		return r.parseErr
	}
	if !r.tierExplicit {
		r.tier = defaultTier
		for _, rule := range r.rules {
			rule.Tier = defaultTier
		}
	}
	return nil
}

var reAttrs = regexp.MustCompile(`\s*([^:]+)\s*:\s*([^;]+)\s*;?`)

// attrsMapFromString parses a semicolon-separated list of "key: value" pairs
// into a map. Trailing semicolons and surrounding whitespace are handled
// gracefully.
func attrsMapFromString(s string) map[string]string {
	attrs := make(map[string]string)
	pairs := reAttrs.FindAllStringSubmatch(s, -1)
	for _, pair := range pairs {
		attrs[pair[1]] = strings.TrimSpace(pair[2])
	}
	return attrs
}

func stripCSSComments(s string) string {
	return reCSSComment.ReplaceAllString(s, "")
}

func init() {
	registerTag(DefaultSpace, "style", func() any { return &Rules{} })
}

var _ HasAttrs = (*Rules)(nil)
