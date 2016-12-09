// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type ParagraphStyle struct {
	TextStyle
	bullet *BulletStyle
}

func (ps *ParagraphStyle) Apply(w Writer) {
	ps.TextStyle.Apply(w)
	fmt.Printf("Applying %s\n", ps)
}

func (ps *ParagraphStyle) SetAttrs(attrs map[string]string) {
	ps.TextStyle.SetAttrs(attrs)
	if bullet, ok := attrs["bullet"]; ok {
		scope := ScopeFor(ps.parent)
		ps.bullet = BulletStyleFor(bullet, scope)
	}
}

func (ps *ParagraphStyle) String() string {
	return fmt.Sprintf("ParagraphStyle %s bullet=%s", &ps.TextStyle, ps.bullet)
}

func ParagraphStyleFor(id string, scope HasScope) *ParagraphStyle {
	style, ok := scope.Style(id)
	if !ok {
		style, ok = scope.Style("para_" + id)
	}
	if ok {
		ps, _ := style.(*ParagraphStyle)
		return ps
	}
	return nil
}

func init() {
	registerTag(DefaultSpace, "para", func() interface{} { return &ParagraphStyle{} })
}
