// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package assetpath

import "testing"

func TestValid(t *testing.T) {
	valid := []string{
		"logo.png",
		"assets/logo.png",
		"a/b/c.txt",
	}
	for _, name := range valid {
		t.Run("valid:"+name, func(t *testing.T) {
			if !Valid(name) {
				t.Fatalf("Valid(%q) = false, want true", name)
			}
		})
	}

	invalid := []string{
		"",
		".",
		"./logo.png",
		"/etc/passwd",
		"../escape",
		"a/../logo.png",
		"a/..",
	}
	for _, name := range invalid {
		t.Run("invalid:"+name, func(t *testing.T) {
			if Valid(name) {
				t.Fatalf("Valid(%q) = true, want false", name)
			}
		})
	}
}
