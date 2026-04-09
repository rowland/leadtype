// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package serveltml

import (
	"errors"
	"fmt"
	"io"
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

var errOutputPathConflict = errors.New("output path conflict")

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
	if err := validateOutputUploads(pdfName, uploads); err != nil {
		return fileOutputResponse{}, err
	}

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
