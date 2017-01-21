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
