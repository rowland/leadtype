// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func makeMapFS(files map[string]string) fs.FS {
	m := make(fstest.MapFS)
	for name, data := range files {
		m[name] = &fstest.MapFile{Data: []byte(data)}
	}
	return m
}

// TestNewOverlayFS_Wiring verifies that newOverlayFS correctly wires upper and
// lower so that the upper layer takes precedence. Detailed overlay semantics
// are tested in internal/overlayfs.
func TestNewOverlayFS_Wiring(t *testing.T) {
	upper := makeMapFS(map[string]string{"f.txt": "upper"})
	lower := makeMapFS(map[string]string{"f.txt": "lower", "g.txt": "lower-only"})

	o := newOverlayFS(upper, lower)

	if data, err := fs.ReadFile(o, "f.txt"); err != nil || string(data) != "upper" {
		t.Errorf("f.txt: got %q, %v; want upper, nil", data, err)
	}
	if data, err := fs.ReadFile(o, "g.txt"); err != nil || string(data) != "lower-only" {
		t.Errorf("g.txt: got %q, %v; want lower-only, nil", data, err)
	}
}
