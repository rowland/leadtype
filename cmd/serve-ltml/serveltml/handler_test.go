// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package serveltml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/rowland/leadtype/ltml"
)

// minimalLTML is a tiny valid LTML document used across handler tests.
const minimalLTML = `<ltml></ltml>`

type serveComponent struct {
	ltml.StdComponent
}

var (
	registerServeComponentOnce sync.Once
	serveComponentMu           sync.Mutex
	serveComponentBody         string
)

func (c *serveComponent) DrawContent(w ltml.Writer) error {
	serveComponentMu.Lock()
	defer serveComponentMu.Unlock()
	serveComponentBody = c.Body()
	if body := c.Body(); body == "" {
		return fmt.Errorf("component body is empty")
	}
	return nil
}

func registerServeComponent(t *testing.T) {
	t.Helper()
	registerServeComponentOnce.Do(func() {
		if err := ltml.RegisterTagExt("servecomponenttest", "card", func() any { return &serveComponent{} }); err != nil {
			t.Fatalf("register component tag: %v", err)
		}
	})
	serveComponentMu.Lock()
	serveComponentBody = ""
	serveComponentMu.Unlock()
}

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

func TestHandler_LogsSuccessfulRenderLifecycle(t *testing.T) {
	h, _ := newHandler(t)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "application/vnd.rowland.leadtype.ltml+xml", minimalLTML},
		{"file", "logo.txt", "text/plain", "fake-asset-data"},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)

	var logs bytes.Buffer
	prevWriter := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(prevWriter)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	requestID := rr.Header().Get("X-Request-Id")
	if requestID == "" {
		t.Fatal("X-Request-Id header is empty")
	}

	got := logs.String()
	for _, want := range []string{
		"serve-ltml: req=" + requestID + " started:",
		"serve-ltml: req=" + requestID + " parsed ltml: bytes=",
		"serve-ltml: req=" + requestID + " rendering ltml: bytes=",
		"serve-ltml: req=" + requestID + " completed: status=200 pdf_bytes=",
		"elapsed=",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("logs = %q, want substring %q", got, want)
		}
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

func TestRenderLTML_SetsParserAssetFSForUploadedComponentSrc(t *testing.T) {
	registerServeComponent(t)

	tmpDir := t.TempDir()
	uploadDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(uploadDir, "snippet.xml"), []byte("<p>from uploaded asset</p>"), 0o600); err != nil {
		t.Fatal(err)
	}

	overlay := newOverlayFS(os.DirFS(uploadDir), os.DirFS(t.TempDir()))
	pdfFile, err := renderLTML([]byte(`
<ltml xmlns:xt="servecomponenttest">
  <page>
    <xt:card src="snippet.xml" />
  </page>
</ltml>`), overlay, tmpDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	pdfFile.Close()

	serveComponentMu.Lock()
	defer serveComponentMu.Unlock()
	if got, want := serveComponentBody, "<p>from uploaded asset</p>"; got != want {
		t.Fatalf("component body = %q, want %q", got, want)
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
	if ct := rr.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("Content-Type = %q, want text/plain", ct)
	}
	if body := rr.Body.String(); !strings.Contains(body, "method not allowed") {
		t.Errorf("body = %q, want method-not-allowed text", body)
	}
}

// TestHandler_FileOutputMode_NonPostMethod_IsJSON verifies that file-output
// requests get the JSON error shape even for method validation failures.
func TestHandler_FileOutputMode_NonPostMethod_IsJSON(t *testing.T) {
	h, _, _ := newHandlerWithOutput(t)

	req := httptest.NewRequest(http.MethodGet, "/render", nil)
	req.Header.Set("X-Output-File", "report.pdf")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405; body: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp fileOutputError
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding JSON error response: %v", err)
	}
	if resp.Error != "method not allowed" {
		t.Errorf("error = %q, want %q", resp.Error, "method not allowed")
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
	respBody := rr.Body.String()
	if !strings.Contains(respBody, "invalid LTML:") || !strings.Contains(respBody, "XML syntax error") {
		t.Fatalf("body = %q, want detailed invalid LTML parse message", respBody)
	}
}

// newHandlerWithOutput returns a renderHandler with both BasePath and
// OutputPath configured to separate temp directories.
func newHandlerWithOutput(t *testing.T) (*renderHandler, string, string) {
	t.Helper()
	base := t.TempDir()
	outputPath := t.TempDir()
	cfg := &Config{
		Listen:         ":0",
		BasePath:       base,
		OutputPath:     outputPath,
		MaxUploadBytes: 32 << 20,
	}
	return newRenderHandler(cfg), base, outputPath
}

// TestHandler_FileOutputMode_Success verifies that X-Output-File triggers
// JSON response and writes PDF + LTML to the output directory.
func TestHandler_FileOutputMode_Success(t *testing.T) {
	h, _, outputPath := newHandlerWithOutput(t)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "application/vnd.rowland.leadtype.ltml+xml", minimalLTML},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-Output-File", "report.pdf")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp fileOutputResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding JSON response: %v", err)
	}
	if resp.Size <= 0 {
		t.Errorf("size = %d, want > 0", resp.Size)
	}
	if resp.ElapsedMs < 0 {
		t.Errorf("elapsed_ms = %d, want >= 0", resp.ElapsedMs)
	}
	if !strings.HasSuffix(resp.Path, "/report.pdf") {
		t.Errorf("path = %q, want suffix /report.pdf", resp.Path)
	}

	// Verify files on disk.
	pdfPath := filepath.Join(outputPath, filepath.FromSlash(resp.Path))
	if info, err := os.Stat(pdfPath); err != nil {
		t.Errorf("PDF not found at %q: %v", pdfPath, err)
	} else if info.Size() != resp.Size {
		t.Errorf("PDF on-disk size %d != JSON size %d", info.Size(), resp.Size)
	}

	ltmlPath := strings.TrimSuffix(pdfPath, ".pdf") + ".ltml"
	if _, err := os.Stat(ltmlPath); err != nil {
		t.Errorf("LTML not found at %q: %v", ltmlPath, err)
	}
}

// TestHandler_FileOutputMode_WithUploads verifies that uploaded files are
// copied to the output directory alongside the PDF and LTML.
func TestHandler_FileOutputMode_WithUploads(t *testing.T) {
	h, _, outputPath := newHandlerWithOutput(t)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "application/vnd.rowland.leadtype.ltml+xml", minimalLTML},
		{"file", "assets/logo.txt", "text/plain", "fake-asset-data"},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-Output-File", "report.pdf")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body.String())
	}

	var resp fileOutputResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding JSON response: %v", err)
	}

	// Derive ULID directory from the response path.
	parts := strings.SplitN(resp.Path, "/", 2)
	if len(parts) != 2 {
		t.Fatalf("unexpected path format: %q", resp.Path)
	}
	ulidDir := filepath.Join(outputPath, parts[0])

	uploadDest := filepath.Join(ulidDir, "assets", "logo.txt")
	data, err := os.ReadFile(uploadDest)
	if err != nil {
		t.Errorf("upload not found at %q: %v", uploadDest, err)
	} else if string(data) != "fake-asset-data" {
		t.Errorf("upload content = %q, want %q", data, "fake-asset-data")
	}
}

// TestHandler_FileOutputMode_RejectsUploadCollisionWithPDF verifies that an
// uploaded file cannot overwrite the generated PDF output.
func TestHandler_FileOutputMode_RejectsUploadCollisionWithPDF(t *testing.T) {
	h, _, outputPath := newHandlerWithOutput(t)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "application/vnd.rowland.leadtype.ltml+xml", minimalLTML},
		{"file", "report.pdf", "application/pdf", "not-a-rendered-pdf"},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-Output-File", "report.pdf")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp fileOutputError
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding JSON error response: %v", err)
	}
	if !strings.Contains(resp.Error, `uploaded file "report.pdf" conflicts with generated output`) {
		t.Fatalf("error = %q, want collision message", resp.Error)
	}

	entries, err := os.ReadDir(outputPath)
	if err != nil {
		t.Fatalf("reading output path: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("output path contains %d entries, want none", len(entries))
	}
}

// TestHandler_FileOutputMode_RejectsUploadCollisionWithLTML verifies that an
// uploaded file cannot overwrite the generated LTML sidecar.
func TestHandler_FileOutputMode_RejectsUploadCollisionWithLTML(t *testing.T) {
	h, _, outputPath := newHandlerWithOutput(t)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "application/vnd.rowland.leadtype.ltml+xml", minimalLTML},
		{"file", "report.ltml", "application/xml", "<unexpected/>"},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-Output-File", "report.pdf")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rr.Code, rr.Body.String())
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp fileOutputError
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding JSON error response: %v", err)
	}
	if !strings.Contains(resp.Error, `uploaded file "report.ltml" conflicts with generated output`) {
		t.Fatalf("error = %q, want collision message", resp.Error)
	}

	entries, err := os.ReadDir(outputPath)
	if err != nil {
		t.Fatalf("reading output path: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("output path contains %d entries, want none", len(entries))
	}
}

// TestHandler_FileOutputMode_NoOutputPath verifies that sending X-Output-File
// when OutputPath is not configured returns a JSON 500 error.
func TestHandler_FileOutputMode_NoOutputPath(t *testing.T) {
	h, _ := newHandler(t) // OutputPath is empty

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "application/vnd.rowland.leadtype.ltml+xml", minimalLTML},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-Output-File", "report.pdf")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var resp fileOutputError
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding JSON error response: %v", err)
	}
	if resp.Error == "" {
		t.Error("error field is empty")
	}
}

// TestHandler_FileOutputMode_InvalidFilename verifies that an invalid
// X-Output-File value returns a JSON 400 error.
func TestHandler_FileOutputMode_InvalidFilename(t *testing.T) {
	h, _, _ := newHandlerWithOutput(t)

	cases := []string{".", "../escape", "/abs"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			body, ct := buildMultipart([][4]string{
				{"ltml", "", "application/vnd.rowland.leadtype.ltml+xml", minimalLTML},
			})
			req := httptest.NewRequest(http.MethodPost, "/render", body)
			req.Header.Set("Content-Type", ct)
			req.Header.Set("X-Output-File", name)

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("filename=%q: status = %d, want 400", name, rr.Code)
			}
			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("filename=%q: Content-Type = %q, want application/json", name, ct)
			}
		})
	}
}

// TestHandler_FileOutputMode_RenderError_IsJSON verifies that render errors
// are returned as JSON when X-Output-File is set.
func TestHandler_FileOutputMode_RenderError_IsJSON(t *testing.T) {
	h, _, _ := newHandlerWithOutput(t)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "", "<not-valid-xml>"},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-Output-File", "report.pdf")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var resp fileOutputError
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding JSON error response: %v", err)
	}
	if resp.Error == "" {
		t.Error("error field is empty")
	}
}

// TestHandler_FileOutputMode_OversizedRequest_IsJSON verifies that 413 errors
// are returned as JSON when X-Output-File is set.
func TestHandler_FileOutputMode_OversizedRequest_IsJSON(t *testing.T) {
	outputPath := t.TempDir()
	cfg := &Config{
		BasePath:       t.TempDir(),
		OutputPath:     outputPath,
		MaxUploadBytes: 10,
	}
	h := newRenderHandler(cfg)

	body, ct := buildMultipart([][4]string{
		{"ltml", "", "", strings.Repeat("X", 100)},
	})
	req := httptest.NewRequest(http.MethodPost, "/render", body)
	req.Header.Set("Content-Type", ct)
	req.Header.Set("X-Output-File", "report.pdf")

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
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
