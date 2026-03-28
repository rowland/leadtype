// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package overlayfs_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/rowland/leadtype/internal/overlayfs"
)

func makeMapFS(files map[string]string) fs.FS {
	m := make(fstest.MapFS)
	for name, data := range files {
		m[name] = &fstest.MapFile{Data: []byte(data)}
	}
	return m
}

func TestOverlayFS_UpperShadowsLower(t *testing.T) {
	upper := makeMapFS(map[string]string{"logo.png": "upper-version"})
	lower := makeMapFS(map[string]string{"logo.png": "lower-version"})

	o := overlayfs.New(upper, lower)

	data, err := fs.ReadFile(o, "logo.png")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "upper-version" {
		t.Errorf("got %q, want upper-version", data)
	}
}

func TestOverlayFS_FallsBackToLower(t *testing.T) {
	upper := makeMapFS(map[string]string{})
	lower := makeMapFS(map[string]string{"static.css": "lower-content"})

	o := overlayfs.New(upper, lower)

	data, err := fs.ReadFile(o, "static.css")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "lower-content" {
		t.Errorf("got %q, want lower-content", data)
	}
}

func TestOverlayFS_IndependentInstances(t *testing.T) {
	upperA := makeMapFS(map[string]string{"logo.png": "A"})
	upperB := makeMapFS(map[string]string{"logo.png": "B"})
	lower := makeMapFS(map[string]string{})

	ovA := overlayfs.New(upperA, lower)
	ovB := overlayfs.New(upperB, lower)

	dataA, _ := fs.ReadFile(ovA, "logo.png")
	dataB, _ := fs.ReadFile(ovB, "logo.png")

	if string(dataA) != "A" {
		t.Errorf("instance A: got %q, want A", dataA)
	}
	if string(dataB) != "B" {
		t.Errorf("instance B: got %q, want B", dataB)
	}
}

func TestOverlayFS_ReadDir_MergesAndSortsEntries(t *testing.T) {
	upper := makeMapFS(map[string]string{"a.txt": "upper-a", "b.txt": "upper-b"})
	lower := makeMapFS(map[string]string{"b.txt": "lower-b", "c.txt": "lower-c"})

	o := overlayfs.New(upper, lower)

	entries, err := fs.ReadDir(o, ".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("entry count = %d, want 3", len(entries))
	}

	want := []string{"a.txt", "b.txt", "c.txt"}
	for i, e := range entries {
		if e.Name() != want[i] {
			t.Errorf("entries[%d].Name() = %q, want %q", i, e.Name(), want[i])
		}
	}
}
