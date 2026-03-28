// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rowland/leadtype/ltml"
	"github.com/rowland/leadtype/ltml/ltpdf"
)

type multiFlag []string

func (m *multiFlag) String() string {
	if m == nil || len(*m) == 0 {
		return ""
	}
	return fmt.Sprintf("%v", []string(*m))
}

func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func main() {
	var assetsDir string
	var outputPath string
	var extraFiles multiFlag

	flag.StringVar(&assetsDir, "assets", "", "path to asset `directory`")
	flag.StringVar(&outputPath, "output", "", "output `file` (default: stdout)")
	flag.Var(&extraFiles, "extra", "additional asset `file` (may be repeated)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: render-ltml [flags] <file.ltml>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(flag.Arg(0), assetsDir, outputPath, []string(extraFiles)); err != nil {
		fmt.Fprintf(os.Stderr, "render-ltml: %v\n", err)
		os.Exit(1)
	}
}

func run(inputFile, assetsDir, outputPath string, extraFiles []string) error {
	// Resolve paths before potentially changing directory.
	absInput, err := filepath.Abs(inputFile)
	if err != nil {
		return fmt.Errorf("resolving input: %w", err)
	}

	var absOutput string
	if outputPath != "" {
		absOutput, err = filepath.Abs(outputPath)
		if err != nil {
			return fmt.Errorf("resolving output: %w", err)
		}
	}

	// Set up asset working directory when assets or extra files are provided.
	if assetsDir != "" || len(extraFiles) > 0 {
		workDir, cleanup, err := setupWorkDir(assetsDir, extraFiles)
		if err != nil {
			return err
		}
		defer cleanup()
		if err := os.Chdir(workDir); err != nil {
			return fmt.Errorf("chdir to work dir: %w", err)
		}
	}

	// Parse LTML.
	doc, err := ltml.ParseFile(absInput)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", inputFile, err)
	}

	// Render to PDF.
	w := ltpdf.NewDocWriter()
	if err := doc.Print(w); err != nil {
		return fmt.Errorf("rendering: %w", err)
	}

	// Write output.
	var out io.Writer
	if absOutput != "" {
		f, err := os.Create(absOutput)
		if err != nil {
			return fmt.Errorf("creating output: %w", err)
		}
		defer f.Close()
		out = f
	} else {
		out = os.Stdout
	}

	if _, err := w.WriteTo(out); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

// setupWorkDir creates a temporary directory populated with symlinks to the
// contents of assetsDir and each file in extraFiles. The caller must invoke
// the returned cleanup function when rendering is complete.
func setupWorkDir(assetsDir string, extraFiles []string) (string, func(), error) {
	tmpDir, err := os.MkdirTemp("", "render-ltml-*")
	if err != nil {
		return "", nil, fmt.Errorf("creating work dir: %w", err)
	}
	cleanup := func() { os.RemoveAll(tmpDir) }

	if assetsDir != "" {
		absAssets, err := filepath.Abs(assetsDir)
		if err != nil {
			cleanup()
			return "", nil, fmt.Errorf("resolving assets dir: %w", err)
		}
		entries, err := os.ReadDir(absAssets)
		if err != nil {
			cleanup()
			return "", nil, fmt.Errorf("reading assets dir: %w", err)
		}
		for _, entry := range entries {
			src := filepath.Join(absAssets, entry.Name())
			dst := filepath.Join(tmpDir, entry.Name())
			if err := os.Symlink(src, dst); err != nil {
				cleanup()
				return "", nil, fmt.Errorf("linking asset %s: %w", entry.Name(), err)
			}
		}
	}

	for _, f := range extraFiles {
		abs, err := filepath.Abs(f)
		if err != nil {
			cleanup()
			return "", nil, fmt.Errorf("resolving extra file %s: %w", f, err)
		}
		dst := filepath.Join(tmpDir, filepath.Base(f))
		if err := os.Symlink(abs, dst); err != nil {
			cleanup()
			return "", nil, fmt.Errorf("linking extra file %s: %w", filepath.Base(f), err)
		}
	}

	return tmpDir, cleanup, nil
}
