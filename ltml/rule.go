// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"regexp"
	"strings"
)

type Rule struct {
	Selector       string
	SelectorRegexp *regexp.Regexp
	Attrs          map[string]string
}

func NewRule(selector string, attrs map[string]string) *Rule {
	return &Rule{
		Selector:       selector,
		SelectorRegexp: regexpForSelector(selector),
		Attrs:          attrs,
	}
}

type Rules struct {
	rules []*Rule
}

func (r *Rules) AddComment(comment string) {
	r.AddText(comment)
}

var reRule = regexp.MustCompile(`\s*([^\{]+?)\s*\{([^\}]+)\}`)

func (r *Rules) AddText(text string) {
	matches := reRule.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		r.rules = append(r.rules, NewRule(m[1], attrsMapFromString(m[2])))
	}
}

var reAttrs = regexp.MustCompile(`\s*([^:]+)\s*:\s*([^;]+)\s*;?`)

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
