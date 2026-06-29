// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package serveltml

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/rowland/leadtype/internal/assetpath"
)

var nextRequestID uint64

// crockford is the Crockford base-32 alphabet used for ULID encoding.
const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// generateULID returns a 26-character ULID string consisting of a 48-bit
// millisecond timestamp followed by 80 bits of cryptographic randomness,
// encoded in Crockford base-32.
func generateULID() (string, error) {
	now := uint64(time.Now().UnixMilli())

	var rnd [10]byte
	if _, err := io.ReadFull(rand.Reader, rnd[:]); err != nil {
		return "", fmt.Errorf("ulid: %w", err)
	}

	var enc [26]byte

	// Timestamp: 10 characters, 48 bits, big-endian 5-bit groups.
	for i := 9; i >= 0; i-- {
		enc[i] = crockford[now&0x1F]
		now >>= 5
	}

	// Random: 16 characters, 80 bits.
	// Slide a uint16 accumulator over the 10 random bytes, draining 5-bit
	// groups as soon as they are available.
	var acc uint16
	bits := 0
	j := 10
	for _, b := range rnd {
		acc = (acc << 8) | uint16(b)
		bits += 8
		for bits >= 5 {
			bits -= 5
			enc[j] = crockford[(acc>>uint(bits))&0x1F]
			j++
		}
	}

	return string(enc[:]), nil
}

// validateOutputFilename checks that the X-Output-File header value satisfies
// the same fs.FS path constraints used for upload filenames: it must be a
// clean, relative path with no ".." components.
func validateOutputFilename(name string) error {
	if !assetpath.Valid(name) {
		return fmt.Errorf("X-Output-File %q is not a valid filename", name)
	}
	return nil
}

// ltmlFilename derives the .ltml sidecar filename from the PDF filename by
// replacing its extension with ".ltml", or appending ".ltml" if there is none.
func ltmlFilename(pdfName string) string {
	ext := filepath.Ext(pdfName)
	if ext == "" {
		return pdfName + ".ltml"
	}
	return pdfName[:len(pdfName)-len(ext)] + ".ltml"
}

// reservedOutputPaths returns the output-relative paths that are owned by the
// generated artifacts for a file-output request.
func reservedOutputPaths(pdfName string) map[string]struct{} {
	return map[string]struct{}{
		pdfName:               {},
		ltmlFilename(pdfName): {},
	}
}

// validateOutputUploads rejects uploaded files that would overwrite the
// generated PDF or LTML sidecar in file-output mode.
func validateOutputUploads(pdfName string, uploads []uploadedFile) error {
	reserved := reservedOutputPaths(pdfName)
	for _, up := range uploads {
		if _, exists := reserved[up.relPath]; exists {
			return fmt.Errorf("%w: uploaded file %q conflicts with generated output", errOutputPathConflict, up.relPath)
		}
	}
	return nil
}

// writeJSONResponse sets Content-Type to application/json, writes the given
// HTTP status code, and encodes v as JSON into the response body.
func writeJSONResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck // headers already sent; nothing we can do
}

// httpError writes either a plain-text error (when fileOutputMode is false) or
// a JSON error object (when fileOutputMode is true).
func httpError(w http.ResponseWriter, fileOutputMode bool, msg string, status int, elapsedMs int64) {
	if fileOutputMode {
		writeJSONResponse(w, status, fileOutputError{Error: msg, ElapsedMs: elapsedMs})
	} else {
		http.Error(w, msg, status)
	}
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
	if !assetpath.Valid(filename) {
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
