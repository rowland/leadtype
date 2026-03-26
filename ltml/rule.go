// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"regexp"
	"strings"
)

// Rule associates a CSS-like selector with a set of attributes. When an element's
// path matches the selector, the attributes are applied to that element before any
// directly-specified attributes (so direct attributes take precedence).
type Rule struct {
	Selector       string
	SelectorRegexp *regexp.Regexp
	Attrs          map[string]string
}

// NewRule creates a Rule for the given selector string and attribute map. The
// selector is compiled into a regexp once so matching during document layout is
// fast. See selectors.go for supported selector syntax.
func NewRule(selector string, attrs map[string]string) *Rule {
	return &Rule{
		Selector:       selector,
		SelectorRegexp: regexpForSelector(selector),
		Attrs:          attrs,
	}
}

// Rules is the in-memory representation of a <rules> block. It holds zero or
// more Rule values parsed from CSS-like text and is registered with the enclosing
// Scope so that elements can be matched against them during parsing.
type Rules struct {
	rules []*Rule
}

// AddComment satisfies the XML comment handler interface. Rule declarations may
// appear inside XML comments (<!-- p { font.size: 12 } -->) so that the LTML
// source remains valid XML.
func (r *Rules) AddComment(comment string) {
	r.AddText(comment)
}

var reRule = regexp.MustCompile(`\s*([^\{]+?)\s*\{([^\}]+)\}`)

// AddText parses one or more CSS-like rule declarations from text and appends
// them to the Rules collection. Each declaration has the form:
//
//	selector { key: value; key: value }
//
// Multiple declarations may appear in a single call. Whitespace around selectors
// and values is trimmed automatically.
func (r *Rules) AddText(text string) {
	matches := reRule.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		r.rules = append(r.rules, NewRule(m[1], attrsMapFromString(m[2])))
	}
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

func init() {
	registerTag(DefaultSpace, "rules", func() interface{} { return &Rules{} })
}
