// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import "testing"

func TestStdWidget_SetAttrs_ParsesSideSpecificBorders(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)
	widget := &StdWidget{}
	widget.SetScope(scope)

	widget.SetAttrs(map[string]string{"border-right": "dashed"})

	if widget.borders[rightSide] == nil {
		t.Fatal("right border is nil, want parsed pen style")
	}
	if got := widget.borders[rightSide].pattern; got != "dashed" {
		t.Fatalf("right border pattern = %q, want dashed", got)
	}
}

func TestStdWidget_SetAttrs_ClonesBorderForBorderPrefixOverrides(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)

	widget := &StdWidget{}
	widget.SetScope(scope)
	widget.SetAttrs(map[string]string{
		"border":       "dashed",
		"border.color": "red",
	})

	if widget.border == nil {
		t.Fatal("border is nil, want cloned pen style")
	}
	base := PenStyleFor("dashed", scope)
	if widget.border == base {
		t.Fatal("border reused shared pen style, want clone")
	}
	if got := widget.border.color; got != NamedColor("red") {
		t.Fatalf("border color = %v, want red", got)
	}
	if got := base.color; got == NamedColor("red") {
		t.Fatalf("shared dashed pen color = %v, want unchanged", got)
	}
}

func TestStdWidget_SetAttrs_ClonesSideBorderForSidePrefixOverrides(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)

	widget := &StdWidget{}
	widget.SetScope(scope)
	widget.SetAttrs(map[string]string{
		"border-right":       "dashed",
		"border-right.color": "red",
	})

	if widget.borders[rightSide] == nil {
		t.Fatal("right border is nil, want cloned pen style")
	}
	base := PenStyleFor("dashed", scope)
	if widget.borders[rightSide] == base {
		t.Fatal("right border reused shared dashed pen, want clone")
	}
	if got := widget.borders[rightSide].color; got != NamedColor("red") {
		t.Fatalf("right border color = %v, want red", got)
	}
	if got := base.color; got == NamedColor("red") {
		t.Fatalf("shared dashed pen color = %v, want unchanged", got)
	}
}

func TestStdWidget_SetAttrs_SideBorderPrefixOverridesCloneMainBorderWhenNeeded(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)

	widget := &StdWidget{}
	widget.SetScope(scope)
	widget.SetAttrs(map[string]string{
		"border":           "dashed",
		"border-top.color": "red",
	})

	if widget.border == nil {
		t.Fatal("main border is nil, want parsed pen style")
	}
	if widget.borders[topSide] == nil {
		t.Fatal("top border is nil, want derived clone")
	}
	if widget.borders[topSide] == widget.border {
		t.Fatal("top border reused main border, want clone")
	}
	if got := widget.borders[topSide].color; got != NamedColor("red") {
		t.Fatalf("top border color = %v, want red", got)
	}
	if got := widget.border.color; got == NamedColor("red") {
		t.Fatalf("main border color = %v, want unchanged", got)
	}
	if got := widget.borders[topSide].pattern; got != widget.border.pattern {
		t.Fatalf("top border pattern = %q, want %q", got, widget.border.pattern)
	}
}
