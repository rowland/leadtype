// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"fmt"
)

type Scope struct {
	parent  HasScope
	aliases map[string]*Alias
	styles  map[string]Styler
	layouts map[string]*LayoutStyle
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

func (scope *Scope) Alias(name string) (alias *Alias, ok bool) {
	alias, ok = scope.aliases[name]
	if !ok && scope.parent != nil {
		alias, ok = scope.parent.Alias(name)
	}
	return
}

func (scope *Scope) Layout(id string) (style *LayoutStyle, ok bool) {
	style, ok = scope.layouts[id]
	if !ok && scope.parent != nil {
		style, ok = scope.parent.Layout(id)
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

func (scope *Scope) Style(id string) (style Styler, ok bool) {
	style, ok = scope.styles[id]
	if !ok && scope.parent != nil {
		style, ok = scope.parent.Style(id)
	}
	return
}

func ScopeFor(parent interface{}) HasScope {
	for {
		if s, ok := parent.(HasScope); ok {
			return s
		}
		// panic here if tree is constructed with neither HasParent nor HasScope in ancestry
		parent = parent.(HasParent).Parent()
	}
}

var defaultScope = Scope{aliases: StdAliases}

var _ HasScope = (*Scope)(nil)
