// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

// fakeFS is a simple in-memory filesystem for testing asset FS threading.
func fakeFS(files map[string][]byte) fs.FS {
	m := make(fstest.MapFS)
	for name, data := range files {
		m[name] = &fstest.MapFile{Data: data}
	}
	return m
}

// TestDoc_SetAssetFS_IsDocumentLocal verifies that each document carries its
// own asset filesystem and that one document's FS does not affect another.
func TestDoc_SetAssetFS_IsDocumentLocal(t *testing.T) {
	fsA := fakeFS(map[string][]byte{"logo.png": []byte("A")})
	fsB := fakeFS(map[string][]byte{"logo.png": []byte("B")})

	docA, err := Parse([]byte("<ltml></ltml>"))
	if err != nil {
		t.Fatal(err)
	}
	docB, err := Parse([]byte("<ltml></ltml>"))
	if err != nil {
		t.Fatal(err)
	}

	docA.SetAssetFS(fsA)
	docB.SetAssetFS(fsB)

	// Doc A's scope should see fsA.
	gotA := docA.scope().AssetFS()
	if gotA == nil {
		t.Fatal("docA assetFS is nil")
	}
	dataA, err := fs.ReadFile(gotA, "logo.png")
	if err != nil {
		t.Fatalf("docA: reading logo.png: %v", err)
	}
	if string(dataA) != "A" {
		t.Errorf("docA logo.png = %q, want \"A\"", dataA)
	}

	// Doc B's scope should see fsB.
	gotB := docB.scope().AssetFS()
	if gotB == nil {
		t.Fatal("docB assetFS is nil")
	}
	dataB, err := fs.ReadFile(gotB, "logo.png")
	if err != nil {
		t.Fatalf("docB: reading logo.png: %v", err)
	}
	if string(dataB) != "B" {
		t.Errorf("docB logo.png = %q, want \"B\"", dataB)
	}
}

// TestScope_AssetFS_InheritedByNestedScope verifies that a nested scope
// inherits the asset filesystem from the nearest ancestor that has one set.
func TestScope_AssetFS_InheritedByNestedScope(t *testing.T) {
	fsys := fakeFS(map[string][]byte{"img.png": []byte("root")})

	root := &Scope{}
	root.SetAssetFS(fsys)

	child := &Scope{}
	child.SetParentScope(root)

	grandchild := &Scope{}
	grandchild.SetParentScope(child)

	// Verify inheritance by reading a file through the inherited FS.
	got := grandchild.AssetFS()
	if got == nil {
		t.Fatal("grandchild AssetFS() is nil; expected to inherit from root")
	}
	data, err := fs.ReadFile(got, "img.png")
	if err != nil {
		t.Fatalf("reading via inherited FS: %v", err)
	}
	if string(data) != "root" {
		t.Errorf("data = %q, want \"root\"", data)
	}
}

// TestScope_AssetFS_NilWhenNotSet verifies that AssetFS returns nil when no
// ancestor has an asset filesystem configured.
func TestScope_AssetFS_NilWhenNotSet(t *testing.T) {
	root := &Scope{}
	child := &Scope{}
	child.SetParentScope(root)

	if child.AssetFS() != nil {
		t.Error("expected nil assetFS when none is set")
	}
}

// dataWriterSpy records calls to PrintImage and ImageDimensions so tests can
// verify that the FS-based code path is taken.
type dataWriterSpy struct {
	imageTestWriter
	imageDimensionsCalls int
	printImageCalls      int
	lastData             []byte
}

func (s *dataWriterSpy) ImageDimensions(data []byte) (width, height int, err error) {
	s.imageDimensionsCalls++
	s.lastData = data
	return 100, 50, nil
}

func (s *dataWriterSpy) PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	s.printImageCalls++
	s.lastData = data
	return 0, 0, nil
}

// TestStdImage_DrawContent_UsesAssetFS verifies that when a scope has an asset
// filesystem and the writer supports data-based image methods, DrawContent
// resolves the image via the FS rather than via PrintImageFile.
func TestStdImage_DrawContent_UsesAssetFS(t *testing.T) {
	pngBytes := []byte("fake-png-data")
	fsys := fakeFS(map[string][]byte{"logo.png": pngBytes})

	scope := &Scope{}
	scope.SetAssetFS(fsys)

	img := &StdImage{src: "logo.png"}
	img.SetScope(scope)

	spy := &dataWriterSpy{}
	if err := img.DrawContent(spy); err != nil {
		t.Fatalf("DrawContent: %v", err)
	}

	if spy.printImageCalls != 1 {
		t.Errorf("PrintImage calls = %d, want 1", spy.printImageCalls)
	}
	if len(spy.calls) != 0 {
		t.Errorf("PrintImageFile unexpectedly called %d time(s)", len(spy.calls))
	}
	if string(spy.lastData) != string(pngBytes) {
		t.Errorf("data passed to PrintImage = %q, want %q", spy.lastData, pngBytes)
	}
}

// TestStdImage_DrawContent_FallsBackToFilePath verifies that when no asset
// filesystem is set, DrawContent uses the Writer's file-path–based method.
func TestStdImage_DrawContent_FallsBackToFilePath(t *testing.T) {
	img := &StdImage{src: "fixture.jpg"}
	w := &imageTestWriter{}

	if err := img.DrawContent(w); err != nil {
		t.Fatalf("DrawContent: %v", err)
	}

	if len(w.calls) != 1 {
		t.Fatalf("PrintImageFile calls = %d, want 1", len(w.calls))
	}
	if w.calls[0].filename != "fixture.jpg" {
		t.Errorf("filename = %q, want fixture.jpg", w.calls[0].filename)
	}
}
