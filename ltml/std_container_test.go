// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import "testing"

func TestStdContainer_Path(t *testing.T) {
	i1 := map[string]string{"tag": "foo", "id": "bar", "class": "boom baz"}
	e1 := "foo#bar.baz.boom"

	var c1 StdContainer
	c1.SetIentifiers(i1)
	if c1.Path() != e1 {
		t.Errorf("Expected <%s>, got <%s>", e1, c1.Path())
	}

	i2 := map[string]string{"tag": "abc", "id": "def", "class": "jkl ghi"}
	e2 := "foo#bar.baz.boom/abc#def.ghi.jkl"

	var c2 StdContainer
	c2.SetIentifiers(i2)
	c2.container = &c1
	if c2.Path() != e2 {
		t.Errorf("Expected <%s>, got <%s>", e2, c2.Path())
	}
}

func TestStdContainer_SetAttrs_ClonesLayoutForLayoutPrefixOverrides(t *testing.T) {
	scope := &Scope{}
	scope.SetParentScope(&defaultScope)

	container := &StdContainer{}
	container.SetScope(scope)
	container.SetAttrs(map[string]string{
		"layout":          "vbox",
		"layout.vpadding": "9pt",
	})

	if container.LayoutStyle() == defaultLayouts["vbox"] {
		t.Fatal("layout style reused shared vbox layout, want clone")
	}
	if got := container.LayoutStyle().VPadding(); got != 9 {
		t.Fatalf("layout vpadding = %v, want 9", got)
	}
	if got := defaultLayouts["vbox"].VPadding(); got != 0 {
		t.Fatalf("shared vbox vpadding = %v, want 0", got)
	}
}
