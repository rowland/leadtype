// Copyright 2016, 2017 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"fmt"
	"sort"
)

type Scope struct {
	parent          HasScope
	aliases         map[string]*Alias
	styles          map[string]Styler
	layouts         map[string]*LayoutStyle
	pageStyles      map[string]*PageStyle
	rules           []*Rules
	defaultRuleTier int
}

func (scope *Scope) AddAlias(alias *Alias) error {
	if alias.ID == "" {
		return fmt.Errorf("id required for alias: %s", alias)
	}
	if scope.aliases == nil {
		scope.aliases = make(map[string]*Alias)
	}
	scope.aliases[alias.ID] = alias
	return nil
}

func (scope *Scope) AddLayout(layout *LayoutStyle) error {
	if layout.ID() == "" {
		return fmt.Errorf("id required for layout: %s", layout)
	}
	if scope.layouts == nil {
		scope.layouts = make(map[string]*LayoutStyle)
	}
	scope.layouts[layout.ID()] = layout
	return nil
}

func (scope *Scope) AddPageStyle(style *PageStyle) error {
	if style.ID() == "" {
		return fmt.Errorf("id required for page style: %v", style)
	}
	if scope.pageStyles == nil {
		scope.pageStyles = make(map[string]*PageStyle)
	}
	scope.pageStyles[style.ID()] = style
	return nil
}

// AddRules registers a parsed Rules collection with the scope. Multiple Rules
// sets may be added; they are consulted in the order they were added.
func (scope *Scope) AddRules(rules *Rules) error {
	if err := rules.ensureTier(scope.defaultRuleTier); err != nil {
		return err
	}
	scope.rules = append(scope.rules, rules)
	return nil
}

func (scope *Scope) AddStyle(style Styler) error {
	if style.ID() == "" {
		return fmt.Errorf("id required for style: %s", style)
	}
	if scope.styles == nil {
		scope.styles = make(map[string]Styler)
	}
	scope.styles[style.ID()] = style
	return nil
}

func (scope *Scope) AliasFor(name string) (alias *Alias, ok bool) {
	alias, ok = scope.aliases[name]
	if !ok && scope.parent != nil {
		alias, ok = scope.parent.AliasFor(name)
	}
	return
}

// EachRuleFor calls f for every Rule whose selector matches path in cascade
// order: lower tiers first, then lower specificity, then earlier declarations.
// Direct XML attributes are still applied after rules and therefore win last.
func (scope *Scope) EachRuleFor(path string, f func(rule *Rule)) {
	matched := scope.matchingRules(path)
	sort.SliceStable(matched, func(i, j int) bool {
		left, right := matched[i], matched[j]
		if left.Tier != right.Tier {
			return left.Tier < right.Tier
		}
		if cmp := left.Specificity.Compare(right.Specificity); cmp != 0 {
			return cmp < 0
		}
		return left.Order < right.Order
	})
	for _, rule := range matched {
		f(rule)
	}
}

func (scope *Scope) matchingRules(path string) []*Rule {
	var matched []*Rule
	if parent, ok := scope.parent.(interface{ matchingRules(string) []*Rule }); ok {
		matched = append(matched, parent.matchingRules(path)...)
	}
	for _, rz := range scope.rules {
		for _, rule := range rz.rules {
			if rule.Compiled != nil && rule.Compiled.MatchesPath(path) {
				matched = append(matched, rule)
			}
		}
	}
	return matched
}

// EachPseudoRuleForWidget calls f for every pseudo-class Rule matching widget
// in cascade order: lower tiers first, then lower specificity, then earlier
// declarations.
func (scope *Scope) EachPseudoRuleForWidget(widget Widget, resolver *selectorStructureResolver, f func(rule *Rule)) {
	matched := scope.matchingPseudoRules(widget, resolver)
	sort.SliceStable(matched, func(i, j int) bool {
		left, right := matched[i], matched[j]
		if left.Tier != right.Tier {
			return left.Tier < right.Tier
		}
		if cmp := left.Specificity.Compare(right.Specificity); cmp != 0 {
			return cmp < 0
		}
		return left.Order < right.Order
	})
	for _, rule := range matched {
		f(rule)
	}
}

func (scope *Scope) matchingPseudoRules(widget Widget, resolver *selectorStructureResolver) []*Rule {
	var matched []*Rule
	if parent, ok := scope.parent.(interface {
		matchingPseudoRules(Widget, *selectorStructureResolver) []*Rule
	}); ok {
		matched = append(matched, parent.matchingPseudoRules(widget, resolver)...)
	}
	for _, rz := range scope.rules {
		for _, rule := range rz.rules {
			if !rule.HasPseudo || rule.Compiled == nil {
				continue
			}
			if rule.Compiled.MatchesWidget(widget, resolver) {
				matched = append(matched, rule)
			}
		}
	}
	return matched
}

func (scope *Scope) LayoutFor(id string) (style *LayoutStyle, ok bool) {
	style, ok = scope.layouts[id]
	if !ok && scope.parent != nil {
		style, ok = scope.parent.LayoutFor(id)
	}
	return
}

func (scope *Scope) PageStyleFor(id string) (style *PageStyle, ok bool) {
	style, ok = scope.pageStyles[id]
	if !ok && scope.parent != nil {
		style, ok = scope.parent.PageStyleFor(id)
	}
	return
}

func (scope *Scope) SetParentScope(parent HasScope) {
	scope.parent = parent
}

func (scope *Scope) String() string {
	var buf bytes.Buffer
	for name, alias := range scope.aliases {
		fmt.Fprintf(&buf, "%s=%s\n", name, alias)
	}
	for id, style := range scope.styles {
		fmt.Fprintf(&buf, "%s=%s\n", id, style)
	}
	return buf.String()
}

func (scope *Scope) StyleFor(id string) (style Styler, ok bool) {
	style, ok = scope.styles[id]
	if !ok && scope.parent != nil {
		style, ok = scope.parent.StyleFor(id)
	}
	return
}

var defaultStyles = map[string]Styler{
	"solid":  &PenStyle{id: "solid", color: NamedColor("black"), width: 0.001, pattern: "solid"},
	"dotted": &PenStyle{id: "dotted", color: NamedColor("black"), width: 0.001, pattern: "dotted"},
	"dashed": &PenStyle{id: "dashed", color: NamedColor("black"), width: 0.001, pattern: "dashed"},
	"fixed":  &FontStyle{id: "fixed", entries: []fontEntry{{name: "Courier New"}}, size: 12},
	"ordered": &BulletStyle{
		id:     "ordered",
		format: "%d.",
	},
	"unordered": &BulletStyle{
		id:     "unordered",
		shape:  "circle",
		width:  defaultListBulletSize,
		alignY: "middle",
		r:      3,
	},
}

var defaultScope = Scope{
	aliases:    StdAliases,
	styles:     defaultStyles,
	layouts:    defaultLayouts,
	pageStyles: defaultPageStyles,
}

var _ HasScope = (*Scope)(nil)
