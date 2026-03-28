// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/rowland/leadtype/ltml"
	"github.com/rowland/leadtype/ltml/ltpdf"
)

var errInvalidLTML = errors.New("invalid LTML")

// renderLTML parses ltmlBytes, renders the document using overlay as the
// writer's asset filesystem, and returns a temp PDF file inside tmpDir.
// The returned file is positioned at the beginning of the PDF.
// open *os.File positioned at the beginning of the PDF. The caller is
// responsible for closing the file; the file itself lives inside tmpDir
// and will be removed when tmpDir is cleaned up.
func renderLTML(ltmlBytes []byte, overlay *overlayFS, tmpDir string) (*os.File, error) {
	if overlay == nil {
		return nil, fmt.Errorf("missing asset filesystem")
	}

	doc, err := ltml.Parse(ltmlBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errInvalidLTML, err)
	}

	w := ltpdf.NewDocWriter()
	w.SetAssetFS(overlay)

	if err := doc.Print(w); err != nil {
		return nil, fmt.Errorf("rendering LTML: %w", err)
	}

	tmpPDF, err := os.CreateTemp(tmpDir, "output-*.pdf")
	if err != nil {
		return nil, fmt.Errorf("creating temp PDF: %w", err)
	}

	if _, err := w.WriteTo(tmpPDF); err != nil {
		tmpPDF.Close()
		return nil, fmt.Errorf("writing PDF: %w", err)
	}

	if _, err := tmpPDF.Seek(0, io.SeekStart); err != nil {
		tmpPDF.Close()
		return nil, fmt.Errorf("seeking PDF: %w", err)
	}

	return tmpPDF, nil
}
