// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"bytes"
	"io"
	"io/fs"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestBuildRemoteRequestBody_IncludesLTMLAndExtraFiles(t *testing.T) {
	inputDir := t.TempDir()
	inputFile := filepath.Join(inputDir, "report.ltml")
	ltmlBytes := []byte(`<ltml></ltml>`)
	if err := os.WriteFile(inputFile, ltmlBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	extraDir := t.TempDir()
	logoFile := filepath.Join(extraDir, "logo.txt")
	iconFile := filepath.Join(extraDir, "icon.dat")
	if err := os.WriteFile(logoFile, []byte("logo-data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(iconFile, []byte("icon-data"), 0o600); err != nil {
		t.Fatal(err)
	}

	body, contentType, err := buildRemoteRequestBody(inputFile, ltmlBytes, []string{logoFile, iconFile})
	if err != nil {
		t.Fatal(err)
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatal(err)
	}
	if mediaType != "multipart/form-data" {
		t.Fatalf("content type = %q, want multipart/form-data", mediaType)
	}

	mr := multipart.NewReader(bytes.NewReader(body), params["boundary"])

	part, err := mr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if got := part.FormName(); got != "ltml" {
		t.Fatalf("first part form name = %q, want ltml", got)
	}
	if got := part.FileName(); got != "" {
		t.Fatalf("ltml part filename = %q, want empty", got)
	}
	if got := part.Header.Get("Content-Type"); got != "application/vnd.rowland.leadtype.ltml+xml" {
		t.Fatalf("ltml content type = %q", got)
	}
	data, err := io.ReadAll(part)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(ltmlBytes) {
		t.Fatalf("ltml bytes = %q, want %q", data, ltmlBytes)
	}

	part, err = mr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if got := part.FormName(); got != "file" {
		t.Fatalf("second part form name = %q, want file", got)
	}
	if got := part.FileName(); got != "logo.txt" {
		t.Fatalf("second part filename = %q, want logo.txt", got)
	}
	data, err = io.ReadAll(part)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "logo-data" {
		t.Fatalf("second part body = %q", data)
	}

	part, err = mr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if got := part.FormName(); got != "file" {
		t.Fatalf("third part form name = %q, want file", got)
	}
	if got := part.FileName(); got != "icon.dat" {
		t.Fatalf("third part filename = %q, want icon.dat", got)
	}
	data, err = io.ReadAll(part)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "icon-data" {
		t.Fatalf("third part body = %q", data)
	}
}

func TestBuildRemoteRequestBody_RejectsDuplicateExtraBaseNames(t *testing.T) {
	extraRoot := t.TempDir()
	firstDir := filepath.Join(extraRoot, "one")
	secondDir := filepath.Join(extraRoot, "two")
	if err := os.MkdirAll(firstDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(secondDir, 0o700); err != nil {
		t.Fatal(err)
	}

	first := filepath.Join(firstDir, "logo.png")
	second := filepath.Join(secondDir, "logo.png")
	if err := os.WriteFile(first, []byte("a"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("b"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, _, err := buildRemoteRequestBody("report.ltml", []byte(`<ltml></ltml>`), []string{first, second})
	if err == nil {
		t.Fatal("expected duplicate base name error")
	}
	if !strings.Contains(err.Error(), `duplicate -extra base name "logo.png"`) {
		t.Fatalf("error = %v", err)
	}
}

func TestSubmitRemote_WritesResponseBody(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	extraFile := filepath.Join(t.TempDir(), "logo.txt")
	if err := os.WriteFile(extraFile, []byte("logo-data"), 0o600); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.URL.Path; got != "/render" {
			t.Fatalf("path = %s, want /render", got)
		}

		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			t.Fatal(err)
		}
		if mediaType != "multipart/form-data" {
			t.Fatalf("content type = %q", mediaType)
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		if got := part.FormName(); got != "ltml" {
			t.Fatalf("first part form name = %q", got)
		}
		data, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != `<ltml></ltml>` {
			t.Fatalf("ltml body = %q", data)
		}

		part, err = mr.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		if got := part.FormName(); got != "file" {
			t.Fatalf("second part form name = %q", got)
		}
		if got := part.FileName(); got != "logo.txt" {
			t.Fatalf("second part filename = %q", got)
		}
		data, err = io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "logo-data" {
			t.Fatalf("file body = %q", data)
		}

		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-remote"))
	}))
	defer srv.Close()

	var out bytes.Buffer
	if err := submitRemote(inputFile, "", srv.URL+"/render", []string{extraFile}, &out); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "%PDF-remote" {
		t.Fatalf("output = %q, want %%PDF-remote", got)
	}
}

func TestSubmitRemote_RejectsAssetsDir(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := submitRemote(inputFile, t.TempDir(), "http://example.com/render", nil, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "does not support -assets") {
		t.Fatalf("error = %v", err)
	}
}

func TestSubmitRemote_SurfacesNon2xxResponse(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid LTML", http.StatusBadRequest)
	}))
	defer srv.Close()

	var out bytes.Buffer
	err := submitRemote(inputFile, "", srv.URL, nil, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400 Bad Request") || !strings.Contains(err.Error(), "invalid LTML") {
		t.Fatalf("error = %v", err)
	}
}

func TestRun_LocalModeStillRendersWithoutSubmit(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "report.ltml")
	outputFile := filepath.Join(t.TempDir(), "report.pdf")
	if err := os.WriteFile(inputFile, []byte(`<ltml></ltml>`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := run(inputFile, "", outputFile, "", nil); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected rendered PDF bytes")
	}
}
