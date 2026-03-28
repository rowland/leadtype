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

// TestOverlayFS_ReadDir_MergesEntries verifies that ReadDir includes entries
// from both upper and lower, with upper entries shadowing lower.
func TestOverlayFS_ReadDir_MergesEntries(t *testing.T) {
	upper := makeMapFS(map[string]string{"a.txt": "upper-a", "b.txt": "upper-b"})
	lower := makeMapFS(map[string]string{"b.txt": "lower-b", "c.txt": "lower-c"})

	o := newOverlayFS(upper, lower)

	entries, err := fs.ReadDir(o, ".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}

	for _, want := range []string{"a.txt", "b.txt", "c.txt"} {
		if !names[want] {
			t.Errorf("missing entry %q", want)
		}
	}

	// b.txt should come from upper (shadow), so count should be 3 not 4.
	if len(entries) != 3 {
		t.Errorf("entry count = %d, want 3", len(entries))
	}
}
