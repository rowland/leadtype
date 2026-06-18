// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import "testing"

type testParagraphStyleOwner struct {
	style    *ParagraphStyle
	fallback *ParagraphStyle
}

func (o *testParagraphStyleOwner) ParagraphStyle() *ParagraphStyle {
	if o.style != nil {
		return o.style
	}
	return o.fallback
}

func TestSetParagraphStyleAssignsNamedStyle(t *testing.T) {
	scope := &Scope{}
	base := &ParagraphStyle{TextStyle: TextStyle{id: "body"}}
	if err := scope.AddStyle(base); err != nil {
		t.Fatal(err)
	}
	owner := &testParagraphStyleOwner{fallback: defaultParagraphStyle}

	SetParagraphStyle(&owner.style, "style", map[string]string{
		"style": "body",
	}, scope, owner)

	if owner.style != base {
		t.Fatalf("style = %p, want named style %p", owner.style, base)
	}
}

func TestSetParagraphStyleClonesNamedStyleBeforeOverrides(t *testing.T) {
	scope := &Scope{}
	base := &ParagraphStyle{TextStyle: TextStyle{id: "body", textAlign: HAlignLeft, textAlignSet: true}}
	if err := scope.AddStyle(base); err != nil {
		t.Fatal(err)
	}
	owner := &testParagraphStyleOwner{fallback: defaultParagraphStyle}

	SetParagraphStyle(&owner.style, "style", map[string]string{
		"style":            "body",
		"style.text-align": "right",
	}, scope, owner)

	if owner.style == base {
		t.Fatal("style reused named style, want clone")
	}
	if owner.style.textAlign != HAlignRight {
		t.Fatalf("style text-align = %q, want right", owner.style.textAlign)
	}
	if base.textAlign != HAlignLeft {
		t.Fatalf("named style text-align = %q, want unchanged left", base.textAlign)
	}
}

func TestSetParagraphStyleClonesEffectiveStyleForArbitraryAttribute(t *testing.T) {
	fallback := &ParagraphStyle{TextStyle: TextStyle{textAlign: HAlignLeft, textAlignSet: true}}
	owner := &testParagraphStyleOwner{fallback: fallback}

	SetParagraphStyle(&owner.style, "paragraph-style", map[string]string{
		"paragraph-style.text-align": "center",
	}, nil, owner)

	if owner.style == fallback {
		t.Fatal("style reused effective style, want clone")
	}
	if owner.style.textAlign != HAlignCenter {
		t.Fatalf("style text-align = %q, want center", owner.style.textAlign)
	}
	if fallback.textAlign != HAlignLeft {
		t.Fatalf("effective style text-align = %q, want unchanged left", fallback.textAlign)
	}
}

func TestSetParagraphStyleUnknownNameOverridesEffectiveFallback(t *testing.T) {
	fallback := &ParagraphStyle{TextStyle: TextStyle{textAlign: HAlignLeft, textAlignSet: true}}
	owner := &testParagraphStyleOwner{style: &ParagraphStyle{}, fallback: fallback}
	scope := &Scope{}

	SetParagraphStyle(&owner.style, "style", map[string]string{
		"style":            "missing",
		"style.text-align": "right",
	}, scope, owner)

	if owner.style == fallback || owner.style.textAlign != HAlignRight {
		t.Fatalf("style = %#v, want overridden clone of fallback", owner.style)
	}
}

func TestSetParagraphStyleLeavesFieldUnchangedWithoutMatchingAttrs(t *testing.T) {
	base := &ParagraphStyle{}
	owner := &testParagraphStyleOwner{style: base, fallback: defaultParagraphStyle}

	SetParagraphStyle(&owner.style, "style", map[string]string{
		"paragraph-style.text-align": "center",
	}, nil, owner)

	if owner.style != base {
		t.Fatalf("style = %p, want unchanged %p", owner.style, base)
	}
}
