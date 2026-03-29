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
