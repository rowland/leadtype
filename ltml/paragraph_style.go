// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

type ParagraphStyle struct {
	scope HasScope
	TextStyle
	bullets []*BulletStyle
}

// HasParagraphStyle provides the effective paragraph style used as the base
// for attribute overrides.
type HasParagraphStyle interface {
	ParagraphStyle() *ParagraphStyle
}

func (ps *ParagraphStyle) Apply(w Writer) {
	ps.TextStyle.Apply(w)
	// fmt.Printf("Applying %s\n", ps)
}

func (ps *ParagraphStyle) Bullet() *BulletStyle {
	if bullets := ps.Bullets(); len(bullets) > 0 {
		return bullets[0]
	}
	return nil
}

func (ps *ParagraphStyle) Bullets() []*BulletStyle {
	return ps.bullets
}

func (ps *ParagraphStyle) Clone() *ParagraphStyle {
	clone := *ps
	return &clone
}

func (ps *ParagraphStyle) SetAttrs(attrs map[string]string) {
	ps.TextStyle.SetAttrs(attrs)
	if bullet, ok := attrs["bullet"]; ok {
		ps.bullets = bulletStylesFor(bullet, ps.scope)
	}
}

func (ps *ParagraphStyle) SetScope(scope HasScope) {
	ps.scope = scope
}

func (ps *ParagraphStyle) String() string {
	return fmt.Sprintf("ParagraphStyle %s bullet=%v", &ps.TextStyle, ps.bullets)
}

func bulletStylesFor(value string, scope HasScope) []*BulletStyle {
	if scope == nil {
		return nil
	}
	var bullets []*BulletStyle
	for _, id := range strings.Fields(value) {
		if bullet := BulletStyleFor(id, scope); bullet != nil {
			bullets = append(bullets, bullet)
		}
	}
	return bullets
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

// SetParagraphStyle sets a paragraph style field from attrName and any
// prefixed overrides in attrs. Overrides are applied to a clone of owner's
// effective paragraph style so named or inherited styles are not mutated.
//
// A third-party widget can use this from SetAttrs with its own style field:
//
//	SetParagraphStyle(&w.paragraphStyle, "style", attrs, w.Scope(), w)
func SetParagraphStyle(field **ParagraphStyle, attrName string, attrs map[string]string, scope HasScope, owner HasParagraphStyle) {
	if id, ok := attrs[attrName]; ok {
		*field = ParagraphStyleFor(id, scope)
	}
	prefix := attrName + "."
	if !MapHasKeyPrefix(attrs, prefix) {
		return
	}
	*field = owner.ParagraphStyle().Clone()
	(*field).SetAttrs(filterMapAttrs(prefix, attrs))
}

var defaultParagraphStyle = &ParagraphStyle{}

var _ HasAttrs = (*ParagraphStyle)(nil)
var _ Styler = (*ParagraphStyle)(nil)
var _ WantsScope = (*ParagraphStyle)(nil)

func init() {
	registerTag(DefaultSpace, "para", func() any { return &ParagraphStyle{} })
}
