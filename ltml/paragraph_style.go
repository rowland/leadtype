// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type ParagraphStyle struct {
	scope HasScope
	TextStyle
	bullet *BulletStyle
}

func (ps *ParagraphStyle) Apply(w Writer) {
	ps.TextStyle.Apply(w)
	// fmt.Printf("Applying %s\n", ps)
}

func (ps *ParagraphStyle) Bullet() *BulletStyle {
	return ps.bullet
}

func (ps *ParagraphStyle) Clone() *ParagraphStyle {
	clone := *ps
	return &clone
}

func (ps *ParagraphStyle) SetAttrs(prefix string, attrs map[string]string) {
	ps.TextStyle.SetAttrs(prefix, attrs)
	if bullet, ok := attrs[prefix+"bullet"]; ok {
		ps.bullet = BulletStyleFor(bullet, ps.scope)
	}
}

func (ps *ParagraphStyle) SetScope(scope HasScope) {
	ps.scope = scope
}

func (ps *ParagraphStyle) String() string {
	return fmt.Sprintf("ParagraphStyle %s bullet=%s", &ps.TextStyle, ps.bullet)
}

func ParagraphStyleFor(id string, scope HasScope) *ParagraphStyle {
	if id == "" {
		return defaultParagraphStyle
	}
	if style, ok := scope.StyleFor(id); ok {
		ps, _ := style.(*ParagraphStyle)
		return ps
	}
	return nil
}

var defaultParagraphStyle = &ParagraphStyle{}

var _ HasAttrsPrefix = (*ParagraphStyle)(nil)
var _ Styler = (*ParagraphStyle)(nil)
var _ WantsScope = (*ParagraphStyle)(nil)

func init() {
	registerTag(DefaultSpace, "para", func() interface{} { return &ParagraphStyle{} })
}
