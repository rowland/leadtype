// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import "testing"

func TestStdWidget_Path(t *testing.T) {
	i1 := map[string]string{"tag": "foo", "id": "bar", "class": "boom baz"}
	e1 := "foo#bar.baz.boom"

	var w1 StdWidget
	w1.SetIentifiers(i1)
	if w1.Path() != e1 {
		t.Errorf("Expected <%s>, got <%s>", e1, w1.Path())
	}
}
