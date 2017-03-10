// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package haru

import (
	"reflect"
	"testing"
)

func TestNewLinePatternFromString(t *testing.T) {
	LinePatterns["other"] = &linePattern{[]int{3, 5}, 6}
	expected := []struct {
		name    string
		pattern string
	}{
		{"solid", "[] 0"},
		{"dotted", "[1 2] 0"},
		{"dashed", "[4 2] 0"},
		{"other", "[3 5] 6"},
	}
	for _, e := range expected {
		lp := newLinePatternFromString(e.pattern)
		pat := LinePatterns[e.name]
		if !reflect.DeepEqual(lp, pat) {
			t.Errorf("Expected %#v, got %#v", pat, lp)
		}
	}
}
