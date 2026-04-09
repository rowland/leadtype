// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package serveltml

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

// ltmlContentTypes lists the accepted Content-Type values for the LTML part.
// The first is the preferred private media type; the others are compatibility
// fallbacks (and an empty value, meaning the field was omitted).
var ltmlContentTypes = map[string]bool{
	"application/vnd.rowland.leadtype.ltml+xml": true,
	"application/xml": true,
	"text/xml":        true,
	"":                true,
}

var nextRequestID uint64

// renderHandler is an http.Handler for POST /render.
type renderHandler struct {
	cfg *Config
}

func newRenderHandler(cfg *Config) *renderHandler {
	return &renderHandler{cfg: cfg}
}

func (h *renderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := newRequestID()
	start := time.Now()
	w.Header().Set("X-Request-Id", requestID)

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Determine file output mode from the custom header.
	outputFilename := r.Header.Get("X-Output-File")
	fileOutputMode := outputFilename != ""

	if fileOutputMode {
		if err := validateOutputFilename(outputFilename); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, fileOutputError{Error: err.Error()})
			return
		}
		if h.cfg.OutputPath == "" {
			writeJSONResponse(w, http.StatusInternalServerError, fileOutputError{
				Error: "file output not configured (output-path not set)",
			})
			return
		}
	}

	requestLogf(requestID, "started: method=%s path=%s remote=%s", r.Method, r.URL.Path, r.RemoteAddr)

	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadBytes)

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/form-data" {
		httpError(w, fileOutputMode, "Content-Type must be multipart/form-data", http.StatusBadRequest, 0)
		return
	}
	boundary := params["boundary"]
	if boundary == "" {
		httpError(w, fileOutputMode, "missing multipart boundary", http.StatusBadRequest, 0)
		return
	}

	// Create the request-scoped temp directory. Everything lives here; a
	// deferred RemoveAll cleans it up regardless of how the request ends.
	tmpDir, err := os.MkdirTemp("", "serve-ltml-*")
	if err != nil {
		requestLogf(requestID, "creating temp dir: %v", err)
		httpError(w, fileOutputMode, "internal error", http.StatusInternalServerError, 0)
		return
	}
	defer os.RemoveAll(tmpDir)

	uploadDir := filepath.Join(tmpDir, "uploads")
	if err := os.Mkdir(uploadDir, 0o700); err != nil {
		requestLogf(requestID, "creating upload dir: %v", err)
		httpError(w, fileOutputMode, "internal error", http.StatusInternalServerError, 0)
		return
	}

	mr := multipart.NewReader(r.Body, boundary)

	// --- First part: LTML document ---
	firstPart, err := mr.NextPart()
	if err != nil {
		if isMaxBytesError(err) {
			httpError(w, fileOutputMode, "request too large", http.StatusRequestEntityTooLarge, 0)
		} else {
			httpError(w, fileOutputMode, "missing LTML part", http.StatusBadRequest, 0)
		}
		return
	}

	if firstPart.FormName() != "ltml" {
		firstPart.Close()
		httpError(w, fileOutputMode, `first multipart part must use field name "ltml"`, http.StatusBadRequest, 0)
		return
	}

	partCT := firstPart.Header.Get("Content-Type")
	partMediaType := ""
	if partCT != "" {
		partMediaType, _, _ = mime.ParseMediaType(partCT)
	}
	if !ltmlContentTypes[partMediaType] {
		firstPart.Close()
		httpError(w, fileOutputMode, fmt.Sprintf("unsupported LTML part content type: %q", partCT), http.StatusBadRequest, 0)
		return
	}

	ltmlBytes, err := io.ReadAll(firstPart)
	firstPart.Close()
	if err != nil {
		if isMaxBytesError(err) {
			httpError(w, fileOutputMode, "request too large", http.StatusRequestEntityTooLarge, 0)
		} else {
			requestLogf(requestID, "reading LTML part: %v", err)
			httpError(w, fileOutputMode, "error reading LTML", http.StatusBadRequest, 0)
		}
		return
	}

	if len(ltmlBytes) == 0 {
		httpError(w, fileOutputMode, "LTML part is empty", http.StatusBadRequest, 0)
		return
	}

	// --- Subsequent parts: uploaded asset files ---
	var uploads []uploadedFile
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			if isMaxBytesError(err) {
				httpError(w, fileOutputMode, "request too large", http.StatusRequestEntityTooLarge, 0)
			} else {
				requestLogf(requestID, "reading multipart: %v", err)
				httpError(w, fileOutputMode, "bad multipart request", http.StatusBadRequest, 0)
			}
			return
		}

		if part.FormName() != "file" {
			part.Close()
			httpError(w, fileOutputMode, `uploaded file parts must use field name "file"`, http.StatusBadRequest, 0)
			return
		}

		// Read the raw filename directly from the Content-Disposition header
		// rather than using part.FileName(), which calls filepath.Base and
		// would strip path components like "assets/" from "assets/logo.png".
		filename := rawFilename(part.Header.Get("Content-Disposition"))
		destPath, validErr := validateUploadFilename(filename, uploadDir)
		if validErr != nil {
			part.Close()
			httpError(w, fileOutputMode, fmt.Sprintf("invalid filename: %v", validErr), http.StatusBadRequest, 0)
			return
		}

		if err := saveUploadedFile(part, destPath); err != nil {
			part.Close()
			if isMaxBytesError(err) {
				httpError(w, fileOutputMode, "request too large", http.StatusRequestEntityTooLarge, 0)
			} else {
				requestLogf(requestID, "saving upload %q: %v", filename, err)
				httpError(w, fileOutputMode, "error storing uploaded file", http.StatusInternalServerError, 0)
			}
			return
		}
		part.Close()
		uploads = append(uploads, uploadedFile{relPath: filename, absPath: destPath})
	}
	uploadCount := len(uploads)
	requestLogf(requestID, "parsed ltml: bytes=%d uploads=%d", len(ltmlBytes), uploadCount)

	// --- Render ---
	baseFSys := os.DirFS(h.cfg.BasePath)
	uploadFSys := os.DirFS(uploadDir)
	overlay := newOverlayFS(uploadFSys, baseFSys)
	requestLogf(requestID, "rendering ltml: bytes=%d uploads=%d", len(ltmlBytes), uploadCount)

	renderStart := time.Now()
	pdfFile, err := renderLTML(ltmlBytes, overlay, tmpDir)
	if err != nil {
		elapsedMs := time.Since(renderStart).Milliseconds()
		requestLogf(requestID, "render: %v", err)
		if errors.Is(err, errInvalidLTML) {
			httpError(w, fileOutputMode, err.Error(), http.StatusBadRequest, elapsedMs)
		} else {
			httpError(w, fileOutputMode, err.Error(), http.StatusInternalServerError, elapsedMs)
		}
		return
	}
	defer pdfFile.Close()

	// --- File output mode: persist to disk and return JSON ---
	if fileOutputMode {
		result, err := writeFileOutput(h.cfg.OutputPath, pdfFile, ltmlBytes, uploads, outputFilename, renderStart)
		if err != nil {
			requestLogf(requestID, "file output: %v", err)
			if errors.Is(err, errOutputPathConflict) {
				writeJSONResponse(w, http.StatusBadRequest, fileOutputError{
					Error:     err.Error(),
					ElapsedMs: time.Since(renderStart).Milliseconds(),
				})
				return
			}
			writeJSONResponse(w, http.StatusInternalServerError, fileOutputError{
				Error:     "error writing output files",
				ElapsedMs: time.Since(renderStart).Milliseconds(),
			})
			return
		}
		writeJSONResponse(w, http.StatusOK, result)
		requestLogf(requestID, "completed file output: path=%s size=%d uploads=%d elapsed=%dms",
			result.Path, result.Size, uploadCount, time.Since(start).Milliseconds())
		return
	}

	// --- Stream response ---
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `inline; filename="output.pdf"`)
	n, err := io.Copy(w, pdfFile)
	if err != nil {
		// Headers already sent; can only log.
		requestLogf(requestID, "streaming PDF: %v", err)
		return
	}
	requestLogf(requestID, "completed: status=%d pdf_bytes=%d uploads=%d elapsed=%dms", http.StatusOK, n, uploadCount, time.Since(start).Milliseconds())
}

func newRequestID() string {
	id := atomic.AddUint64(&nextRequestID, 1)
	return fmt.Sprintf("%06d", id)
}

func requestLogf(requestID, format string, args ...any) {
	log.Printf("serve-ltml: req=%s "+format, append([]any{requestID}, args...)...)
}

// validateUploadFilename checks that filename is a clean fs.FS-relative path
// and returns the absolute destination path under uploadDir.
func validateUploadFilename(filename, uploadDir string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("filename must not be empty")
	}
	if filename == "." || !fs.ValidPath(filename) {
		return "", fmt.Errorf("filename must be a clean relative asset path")
	}
	return filepath.Join(uploadDir, filepath.FromSlash(filename)), nil
}

// saveUploadedFile writes the contents of part to destPath, creating parent
// directories as needed.
func saveUploadedFile(r io.Reader, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o700); err != nil {
		return err
	}
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

// rawFilename extracts the "filename" parameter from a Content-Disposition
// header without stripping directory components. The standard library's
// Part.FileName() calls filepath.Base, which would discard path prefixes like
// "assets/" that we want to preserve for nested asset placement.
func rawFilename(contentDisposition string) string {
	if contentDisposition == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return ""
	}
	return params["filename"]
}

// isMaxBytesError reports whether err signals that the request body size limit
// was exceeded (http.MaxBytesReader).
func isMaxBytesError(err error) bool {
	var mbe *http.MaxBytesError
	return errors.As(err, &mbe)
}
