// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import "testing"
import "reflect"

func TestIdentity_SelectorTag(t *testing.T) {
	i1 := Identity{}
	e1 := ""
	if value := i1.SelectorTag(); value != e1 {
		t.Errorf("Expected <%s>, got <%s>", e1, value)
	}

	i2 := Identity{Tag: "foo"}
	e2 := "foo"
	if value := i2.SelectorTag(); value != e2 {
		t.Errorf("Expected <%s>, got <%s>", e2, value)
	}

	i3 := Identity{Tag: "foo", ID: "bar"}
	e3 := "foo#bar"
	if value := i3.SelectorTag(); value != e3 {
		t.Errorf("Expected <%s>, got <%s>", e3, value)
	}

	i4 := Identity{Tag: "foo", Classes: []string{"baz"}}
	e4 := "foo.baz"
	if value := i4.SelectorTag(); value != e4 {
		t.Errorf("Expected <%s>, got <%s>", e4, value)
	}

	i5 := Identity{Tag: "foo", ID: "bar", Classes: []string{"baz"}}
	e5 := "foo#bar.baz"
	if value := i5.SelectorTag(); value != e5 {
		t.Errorf("Expected <%s>, got <%s>", e5, value)
	}
}

func TestIdentity_SetIdentifiers(t *testing.T) {
	var i1 Identity
	i1.SetIentifiers(map[string]string{"tag": "foo"})
	if i1.Tag != "foo" {
		t.Errorf("Expected <foo>, got <%s>", i1.Tag)
	}
	if i1.ID != "" {
		t.Errorf("Expected <>, got <%s>", i1.ID)
	}
	if len(i1.Classes) != 0 {
		t.Errorf("Expected empty slice, got <%v>", i1.Classes)
	}

	var i2 Identity
	i2.SetIentifiers(map[string]string{"tag": "foo", "id": "bar"})
	if i2.Tag != "foo" {
		t.Errorf("Expected <foo>, got <%s>", i2.Tag)
	}
	if i2.ID != "bar" {
		t.Errorf("Expected <bar>, got <%s>", i2.ID)
	}
	if len(i2.Classes) != 0 {
		t.Errorf("Expected empty slice, got <%v>", i2.Classes)
	}

	ec := []string{"baz", "boom"}

	var i3 Identity
	i3.SetIentifiers(map[string]string{"tag": "foo", "class": "boom baz"})
	if i3.Tag != "foo" {
		t.Errorf("Expected <foo>, got <%s>", i3.Tag)
	}
	if i3.ID != "" {
		t.Errorf("Expected <>, got <%s>", i3.ID)
	}
	if !reflect.DeepEqual(ec, i3.Classes) {
		t.Errorf("Expected <%v>, got <%v>", ec, i3.Classes)
	}

	var i4 Identity
	i4.SetIentifiers(map[string]string{"tag": "foo", "id": "bar", "class": "boom baz"})
	if i4.Tag != "foo" {
		t.Errorf("Expected <foo>, got <%s>", i4.Tag)
	}
	if i4.ID != "bar" {
		t.Errorf("Expected <bar>, got <%s>", i4.ID)
	}
	if !reflect.DeepEqual(ec, i4.Classes) {
		t.Errorf("Expected <%v>, got <%v>", ec, i4.Classes)
	}
}
