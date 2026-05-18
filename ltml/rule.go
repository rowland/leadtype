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
	Selector    string
	Compiled    *compiledSelectorList
	Attrs       map[string]string
	Tier        int
	Specificity Specificity
	Order       int
	HasPseudo   bool
}

// NewRule creates a Rule for the given selector string and attribute map. The
// selector is compiled into a regexp once so matching during document layout is
// fast. See selectors.go for supported selector syntax.
func NewRule(selector string, attrs map[string]string, tier, order int) (*Rule, error) {
	compiled, err := compileSelectorList(selector)
	if err != nil {
		return nil, err
	}
	hasPseudo := false
	for _, item := range compiled.selectors {
		if item.hasPseudo {
			hasPseudo = true
			break
		}
	}
	return &Rule{
		Selector:    selector,
		Compiled:    compiled,
		Attrs:       attrs,
		Tier:        tier,
		Specificity: specificityForSelector(selector),
		Order:       order,
		HasPseudo:   hasPseudo,
	}, nil
}

// Rules is the in-memory representation of a <style> block. It holds zero or
// more Rule values parsed from CSS-like text and is registered with the
// enclosing Scope so that elements can be matched against them during parsing.
type Rules struct {
	rules         []*Rule
	source        textSource
	doc           *Doc
	sourceText    string
	sourceLoaded  bool
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
	if r.source.Explicit() {
		return
	}
	r.addText(text, "")
}

func (r *Rules) addText(text, sourceName string) {
	if r.parseErr != nil {
		return
	}
	text = stripCSSComments(text)
	matches := reRule.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		for _, selector := range splitRuleSelectors(m[1]) {
			rule, err := NewRule(selector, attrsMapFromString(m[2]), r.tier, r.nextRuleOrder)
			if err != nil {
				if sourceName != "" {
					r.parseErr = fmt.Errorf("parsing style src %q: %w", sourceName, err)
					return
				}
				r.parseErr = err
				return
			}
			r.rules = append(r.rules, rule)
			r.nextRuleOrder++
		}
	}
}

func (r *Rules) SetAttrs(attrs map[string]string) {
	r.source.SetAttrs(attrs)
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
	if err := r.ensureSourceLoaded(); err != nil {
		return err
	}
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

func (r *Rules) SetDoc(doc *Doc) {
	r.doc = doc
}

func (r *Rules) ensureSourceLoaded() error {
	if !r.source.Explicit() || r.sourceLoaded {
		return r.parseErr
	}
	r.sourceLoaded = true
	if r.doc == nil {
		r.parseErr = fmt.Errorf("style document is not set")
		return r.parseErr
	}
	container := Container(nil)
	if r.doc.root != nil {
		container = r.doc.root
	}
	text, err := r.source.Text(r.doc, container, "", "style")
	if err != nil {
		r.parseErr = fmt.Errorf("loading style src %q: %w", r.source.src, err)
		return r.parseErr
	}
	r.sourceText = text
	r.addText(text, r.source.src)
	return r.parseErr
}

var reAttrs = regexp.MustCompile(`\s*([^:]+)\s*:\s*([^;]+)\s*;?`)

// attrsMapFromString parses a semicolon-separated list of "key: value" pairs
// into a map. Trailing semicolons and surrounding whitespace are handled
// gracefully.
func attrsMapFromString(s string) map[string]string {
	attrs := make(map[string]string)
	pairs := reAttrs.FindAllStringSubmatch(s, -1)
	for _, pair := range pairs {
		attrs[pair[1]] = unquoteRuleValue(strings.TrimSpace(pair[2]))
	}
	return attrs
}

func unquoteRuleValue(value string) string {
	if len(value) < 2 {
		return value
	}
	first, last := value[0], value[len(value)-1]
	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		return value[1 : len(value)-1]
	}
	return value
}

func stripCSSComments(s string) string {
	return reCSSComment.ReplaceAllString(s, "")
}

func init() {
	registerTag(DefaultSpace, "style", func() any { return &Rules{} })
}

var _ HasAttrs = (*Rules)(nil)
var _ WantsDoc = (*Rules)(nil)
