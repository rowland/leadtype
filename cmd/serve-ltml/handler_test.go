// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// minimalLTML is a tiny valid LTML document used across handler tests.
const minimalLTML = `<ltml></ltml>`

// buildMultipart constructs a multipart/form-data body with the given parts.
// Each element of parts is [fieldName, filename, contentType, body].
// Set filename to "" to omit it; set contentType to "" to omit it.
func buildMultipart(parts [][4]string) (body *bytes.Buffer, contentType string) {
	body = &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	for _, p := range parts {
		fieldName, filename, ct, data := p[0], p[1], p[2], p[3]

		h := make(map[string][]string)
		cd := fmt.Sprintf(`form-data; name="%s"`, fieldName)
		if filename != "" {
			cd += fmt.Sprintf(`; filename="%s"`, filename)
		}
		h["Content-Disposition"] = []string{cd}
		if ct != "" {
			h["Content-Type"] = []string{ct}
		}

		pw, err := mw.CreatePart(h)
		if err != nil {
			panic(err)
		}
		io.WriteString(pw, data)
	}
	mw.Close()
	return body, mw.FormDataContentType()
}

// newHandler returns a renderHandler backed by a temp directory used as base-path.
func newHandler(t *testing.T) (*renderHandler, string) {
	t.Helper()
	base := t.TempDir()
	cfg := &Config{
		Listen:         ":0",
		BasePath:       base,
		MaxUploadBytes: 32 << 20,
	}
	return newRenderHandler(cfg), base
}

// TestHandler_ValidLTMLOnly tests a well-formed request with only an LTML part.
func TestHandler_ValidLTMLOnly(t *testing.T) {
	h, _ := newHandler(t)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "application/vnd.rowland.leadtype.ltml+xml", minimalLTML},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Content-Type = %q, want application/pdf", ct)
	}
	if cd := rr.Header().Get("Content-Disposition"); cd != `inline; filename="output.pdf"` {
		t.Errorf("Content-Disposition = %q", cd)
	}
	if rr.Body.Len() == 0 {
		t.Error("response body is empty, expected PDF bytes")
	}
}

// TestHandler_ValidLTMLWithUploadedAsset tests a request that includes an
// uploaded file alongside the LTML document.
func TestHandler_ValidLTMLWithUploadedAsset(t *testing.T) {
	h, _ := newHandler(t)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "application/vnd.rowland.leadtype.ltml+xml", minimalLTML},
		{"file", "logo.txt", "text/plain", "fake-asset-data"},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
}

// TestHandler_MissingLTMLPart tests that a missing LTML part yields 400.
func TestHandler_MissingLTMLPart(t *testing.T) {
	h, _ := newHandler(t)

	// No parts at all.
	body, ct := buildMultipart([][4]string{})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

// TestHandler_LTMLNotFirstPart tests that putting LTML after another part yields 400.
func TestHandler_LTMLNotFirstPart(t *testing.T) {
	h, _ := newHandler(t)

	body, ct := buildMultipart([][4]string{
		{"file", "asset.txt", "text/plain", "data"},
		{"ltml", "", "", minimalLTML},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

// TestHandler_InvalidUploadFilename tests that an invalid uploaded filename yields 400.
func TestHandler_InvalidUploadFilename(t *testing.T) {
	h, _ := newHandler(t)

	cases := []struct {
		name     string
		filename string
	}{
		{"empty", ""},
		{"absolute", "/etc/passwd"},
		{"dot", "."},
		{"dotSlash", "./logo.png"},
		{"dotdot", "../secret"},
		{"normalizedDotDot", "a/../logo.png"},
		{"normalizedDir", "a/.."},
		{"escapingDotDot", "foo/../../secret"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, ct := buildMultipart([][4]string{
				{"ltml", "", "", minimalLTML},
				{"file", tc.filename, "text/plain", "data"},
			})
			req := httptest.NewRequest(http.MethodPost, "/render", body)
			req.Header.Set("Content-Type", ct)

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("filename=%q: status = %d, want 400", tc.filename, rr.Code)
			}
		})
	}
}

// TestHandler_OversizedRequest tests that an oversized request yields 413.
func TestHandler_OversizedRequest(t *testing.T) {
	base := t.TempDir()
	cfg := &Config{
		BasePath:       base,
		MaxUploadBytes: 10, // tiny limit to trigger the error
	}
	h := newRenderHandler(cfg)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "", strings.Repeat("X", 100)},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", rr.Code)
	}
}

// TestHandler_NonPostMethod tests that GET requests to /render yield 405.
func TestHandler_NonPostMethod(t *testing.T) {
	h, _ := newHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/render", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rr.Code)
	}
}

// TestHandler_TempDirRemovedAfterSuccess verifies that the request temp
// directory is cleaned up after a successful render.
func TestHandler_TempDirRemovedAfterSuccess(t *testing.T) {
	// Intercept os.MkdirTemp by checking that no "serve-ltml-*" dirs remain.
	// We check all temp dirs before and after the request.
	tmpBase := os.TempDir()
	before := countServeLTMLDirs(t, tmpBase)

	h, _ := newHandler(t)
	body, ct := buildMultipart([][4]string{
		{"ltml", "", "", minimalLTML},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	after := countServeLTMLDirs(t, tmpBase)
	if after != before {
		t.Errorf("serve-ltml-* temp dirs after request = %d, want %d (leaked)", after, before)
	}
}

// TestHandler_TempDirRemovedAfterRenderFailure verifies that the request temp
// directory is cleaned up even when rendering fails.
func TestHandler_TempDirRemovedAfterRenderFailure(t *testing.T) {
	tmpBase := os.TempDir()
	before := countServeLTMLDirs(t, tmpBase)

	h, _ := newHandler(t)
	body, ct := buildMultipart([][4]string{
		{"ltml", "", "", "<not-valid-xml>"},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	// Render should fail for malformed LTML.
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}

	after := countServeLTMLDirs(t, tmpBase)
	if after != before {
		t.Errorf("serve-ltml-* temp dirs after failed request = %d, want %d (leaked)", after, before)
	}
}

func TestHandler_MalformedLTMLReturnsBadRequest(t *testing.T) {
	h, _ := newHandler(t)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "", "<not-valid-xml>"},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "invalid LTML") {
		t.Fatalf("body = %q, want invalid LTML message", rr.Body.String())
	}
}

func countServeLTMLDirs(t *testing.T, base string) int {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(base, "serve-ltml-*"))
	if err != nil {
		t.Fatalf("globbing temp dirs: %v", err)
	}
	return len(matches)
}

// TestValidateUploadFilename covers the filename validation helper directly.
func TestValidateUploadFilename(t *testing.T) {
	tmpDir := t.TempDir()

	good := []string{
		"logo.png",
		"assets/logo.png",
		"a/b/c.txt",
	}
	for _, name := range good {
		t.Run("valid:"+name, func(t *testing.T) {
			_, err := validateUploadFilename(name, tmpDir)
			if err != nil {
				t.Errorf("unexpected error for %q: %v", name, err)
			}
		})
	}

	bad := []string{
		"",
		".",
		"./logo.png",
		"/etc/passwd",
		"../escape",
		"a/../logo.png",
		"a/..",
		"a/../../escape",
	}
	for _, name := range bad {
		t.Run("invalid:"+name, func(t *testing.T) {
			_, err := validateUploadFilename(name, tmpDir)
			if err == nil {
				t.Errorf("expected error for %q, got nil", name)
			}
		})
	}
}
