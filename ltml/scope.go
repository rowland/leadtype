// Copyright 2016, 2017 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"fmt"
)

type Scope struct {
	parent     HasScope
	aliases    map[string]*Alias
	styles     map[string]Styler
	layouts    map[string]*LayoutStyle
	pageStyles map[string]*PageStyle
	rules      []*Rules
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

func (scope *Scope) AddRules(rules *Rules) error {
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

func (scope *Scope) EachRuleFor(path string, f func(rule *Rule)) {
	if scope.parent != nil {
		scope.parent.EachRuleFor(path, f)
	}
	for _, rz := range scope.rules {
		for _, rule := range rz.rules {
			if rule.SelectorRegexp.MatchString(path) {
				f(rule)
			}
		}
	}
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
	"fixed":  &FontStyle{id: "fixed", name: "Courier New", size: 12},
}

var defaultScope = Scope{
	aliases:    StdAliases,
	styles:     defaultStyles,
	layouts:    defaultLayouts,
	pageStyles: defaultPageStyles,
}

var _ HasScope = (*Scope)(nil)
