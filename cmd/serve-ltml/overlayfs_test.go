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

// TestOverlayFS_UploadShadowsBase verifies that a file present in both the
// upload dir and the base path is resolved from the upload dir.
func TestOverlayFS_UploadShadowsBase(t *testing.T) {
	upper := makeMapFS(map[string]string{"logo.png": "upload-version"})
	lower := makeMapFS(map[string]string{"logo.png": "base-version"})

	o := newOverlayFS(upper, lower)

	data, err := fs.ReadFile(o, "logo.png")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "upload-version" {
		t.Errorf("got %q, want upload-version", data)
	}
}

// TestOverlayFS_FallsBackToBase verifies that a file absent from the upload
// dir is resolved from the base path.
func TestOverlayFS_FallsBackToBase(t *testing.T) {
	upper := makeMapFS(map[string]string{})
	lower := makeMapFS(map[string]string{"static.css": "base-content"})

	o := newOverlayFS(upper, lower)

	data, err := fs.ReadFile(o, "static.css")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "base-content" {
		t.Errorf("got %q, want base-content", data)
	}
}

// TestOverlayFS_ConcurrentRequestsDoNotShare verifies that two overlay
// instances created for different requests do not share state.
func TestOverlayFS_ConcurrentRequestsDoNotShare(t *testing.T) {
	upperA := makeMapFS(map[string]string{"logo.png": "request-A"})
	upperB := makeMapFS(map[string]string{"logo.png": "request-B"})
	lower := makeMapFS(map[string]string{})

	ovA := newOverlayFS(upperA, lower)
	ovB := newOverlayFS(upperB, lower)

	dataA, err := fs.ReadFile(ovA, "logo.png")
	if err != nil {
		t.Fatalf("request A ReadFile: %v", err)
	}
	dataB, err := fs.ReadFile(ovB, "logo.png")
	if err != nil {
		t.Fatalf("request B ReadFile: %v", err)
	}

	if string(dataA) != "request-A" {
		t.Errorf("request A: got %q, want request-A", dataA)
	}
	if string(dataB) != "request-B" {
		t.Errorf("request B: got %q, want request-B", dataB)
	}
}

// TestOverlayFS_ReadDir_MergesAndSortsEntries verifies that ReadDir includes
// entries from both upper and lower (with upper shadowing lower) and that the
// result is sorted by name, as required by the fs.ReadDirFS contract.
func TestOverlayFS_ReadDir_MergesAndSortsEntries(t *testing.T) {
	upper := makeMapFS(map[string]string{"a.txt": "upper-a", "b.txt": "upper-b"})
	lower := makeMapFS(map[string]string{"b.txt": "lower-b", "c.txt": "lower-c"})

	o := newOverlayFS(upper, lower)

	entries, err := fs.ReadDir(o, ".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	// b.txt from upper shadows lower's b.txt, so exactly 3 entries expected.
	if len(entries) != 3 {
		t.Fatalf("entry count = %d, want 3", len(entries))
	}

	// Entries must be sorted by name (fs.ReadDirFS contract).
	want := []string{"a.txt", "b.txt", "c.txt"}
	for i, e := range entries {
		if e.Name() != want[i] {
			t.Errorf("entries[%d].Name() = %q, want %q", i, e.Name(), want[i])
		}
	}
}
