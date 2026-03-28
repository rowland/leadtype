// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildOptionalAssetFS_ExtraOverridesAssetsDir(t *testing.T) {
	assetsDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(assetsDir, "logo.txt"), []byte("lower"), 0o600); err != nil {
		t.Fatal(err)
	}

	extraDir := t.TempDir()
	extraFile := filepath.Join(extraDir, "logo.txt")
	if err := os.WriteFile(extraFile, []byte("upper"), 0o600); err != nil {
		t.Fatal(err)
	}

	assetFS, cleanup, err := buildOptionalAssetFS(assetsDir, []string{extraFile})
	if err != nil {
		t.Fatal(err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	data, err := fs.ReadFile(assetFS, "logo.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "upper" {
		t.Fatalf("expected upper override, got %q", data)
	}
}

func TestBuildOptionalAssetFS_PreservesNestedAssetPathsFromAssetsDir(t *testing.T) {
	assetsDir := t.TempDir()
	nestedDir := filepath.Join(assetsDir, "assets")
	if err := os.MkdirAll(nestedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "logo.txt"), []byte("nested"), 0o600); err != nil {
		t.Fatal(err)
	}

	assetFS, cleanup, err := buildOptionalAssetFS(assetsDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	data, err := fs.ReadFile(assetFS, "assets/logo.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "nested" {
		t.Fatalf("expected nested asset, got %q", data)
	}
}
