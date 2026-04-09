// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package serveltml

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// uploadedFile pairs a validated relative path with its on-disk location
// inside the request temp directory.
type uploadedFile struct {
	relPath string // e.g. "assets/logo.png" — already fs.ValidPath-clean
	absPath string // absolute path inside tmpDir/uploads/
}

// fileOutputResponse is the JSON body returned on a successful file output request.
type fileOutputResponse struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	ElapsedMs int64  `json:"elapsed_ms"`
}

// fileOutputError is the JSON body returned when a file output request fails.
type fileOutputError struct {
	Error     string `json:"error"`
	ElapsedMs int64  `json:"elapsed_ms"`
}

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
	if name == "" || name == "." || !fs.ValidPath(name) {
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

// writeFileOutput creates {outputPath}/{ULID}/ and writes three kinds of files:
//
//   - {pdfName}      — the rendered PDF (copied from pdfFile)
//   - {stem}.ltml    — the original LTML source bytes
//   - {relPath}      — each uploaded attachment at its original relative path
//
// renderStart must be captured just before renderLTML() is called; elapsed_ms
// in the returned response covers the time from renderStart to when the PDF
// has been written to disk.
func writeFileOutput(
	outputPath string,
	pdfFile io.Reader,
	ltmlBytes []byte,
	uploads []uploadedFile,
	pdfName string,
	renderStart time.Time,
) (fileOutputResponse, error) {
	ulid, err := generateULID()
	if err != nil {
		return fileOutputResponse{}, err
	}

	outDir := filepath.Join(outputPath, ulid)
	if err := os.Mkdir(outDir, 0o755); err != nil {
		return fileOutputResponse{}, fmt.Errorf("creating output dir: %w", err)
	}

	// Write PDF and stop the clock immediately afterwards.
	pdfDest := filepath.Join(outDir, filepath.FromSlash(pdfName))
	if err := os.MkdirAll(filepath.Dir(pdfDest), 0o755); err != nil {
		return fileOutputResponse{}, fmt.Errorf("creating pdf parent dir: %w", err)
	}
	size, err := copyToFile(pdfFile, pdfDest)
	if err != nil {
		return fileOutputResponse{}, fmt.Errorf("writing pdf: %w", err)
	}
	elapsedMs := time.Since(renderStart).Milliseconds()

	// Write LTML sidecar.
	ltmlDest := filepath.Join(outDir, filepath.FromSlash(ltmlFilename(pdfName)))
	if err := os.MkdirAll(filepath.Dir(ltmlDest), 0o755); err != nil {
		return fileOutputResponse{}, fmt.Errorf("creating ltml parent dir: %w", err)
	}
	if err := os.WriteFile(ltmlDest, ltmlBytes, 0o644); err != nil {
		return fileOutputResponse{}, fmt.Errorf("writing ltml: %w", err)
	}

	// Copy uploads.
	for _, up := range uploads {
		dest := filepath.Join(outDir, filepath.FromSlash(up.relPath))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fileOutputResponse{}, fmt.Errorf("creating upload dir for %q: %w", up.relPath, err)
		}
		if err := copyPath(up.absPath, dest); err != nil {
			return fileOutputResponse{}, fmt.Errorf("copying upload %q: %w", up.relPath, err)
		}
	}

	return fileOutputResponse{
		Path:      ulid + "/" + pdfName,
		Size:      size,
		ElapsedMs: elapsedMs,
	}, nil
}

// copyToFile writes all bytes from src to a newly created file at destPath and
// returns the number of bytes written.
func copyToFile(src io.Reader, destPath string) (int64, error) {
	dst, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

// copyPath copies the file at srcPath to a newly created file at destPath.
func copyPath(srcPath, destPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()
	_, err = copyToFile(src, destPath)
	return err
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
