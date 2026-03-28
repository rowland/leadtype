// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rowland/leadtype/internal/overlayfs"
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
	flag.StringVar(&assetsDir, "a", "", "path to asset `directory` (shorthand)")
	flag.StringVar(&outputPath, "output", "", "output `file` (default: stdout)")
	flag.StringVar(&outputPath, "o", "", "output `file` (shorthand)")
	flag.Var(&extraFiles, "extra", "additional asset `file` (may be repeated)")
	flag.Var(&extraFiles, "e", "additional asset `file` (shorthand)")
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

	// Parse LTML.
	doc, err := ltml.ParseFile(absInput)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", inputFile, err)
	}

	// Build and attach an asset filesystem when assets or extra files are provided.
	assetFS, cleanup, err := buildOptionalAssetFS(assetsDir, extraFiles)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Render to PDF.
	w := ltpdf.NewDocWriter()
	if assetFS != nil {
		w.SetAssetFS(assetFS)
	}
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

// buildOptionalAssetFS constructs an optional fs.FS that covers assetsDir
// (lower layer) and the named extraFiles (upper layer, each stored under its
// base name). Extra files shadow same-named entries in assetsDir rather than
// erroring on conflict.
//
// Returns nil, nil, nil when neither assetsDir nor extraFiles are provided.
// When a non-nil cleanup function is returned, the caller must invoke it after
// rendering is complete.
func buildOptionalAssetFS(assetsDir string, extraFiles []string) (fs.FS, func(), error) {
	hasAssets := assetsDir != ""
	hasExtras := len(extraFiles) > 0

	if !hasAssets && !hasExtras {
		return nil, nil, nil
	}

	// Place extra files in a temp directory so they can be addressed as an
	// os.DirFS. Symlinks keep memory use low for large files.
	var extraDir string
	var cleanup func()
	if hasExtras {
		tmpDir, err := os.MkdirTemp("", "render-ltml-*")
		if err != nil {
			return nil, nil, fmt.Errorf("creating work dir: %w", err)
		}
		cleanup = func() { os.RemoveAll(tmpDir) }
		extraDir = tmpDir

		for _, f := range extraFiles {
			abs, err := filepath.Abs(f)
			if err != nil {
				cleanup()
				return nil, nil, fmt.Errorf("resolving extra file %s: %w", f, err)
			}
			dst := filepath.Join(extraDir, filepath.Base(f))
			// If two extra files share a base name, the last one wins.
			os.Remove(dst)
			if err := os.Symlink(abs, dst); err != nil {
				cleanup()
				return nil, nil, fmt.Errorf("linking extra file %s: %w", filepath.Base(f), err)
			}
		}
	}

	switch {
	case hasExtras && hasAssets:
		absAssets, err := filepath.Abs(assetsDir)
		if err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("resolving assets dir: %w", err)
		}
		return overlayfs.New(os.DirFS(extraDir), os.DirFS(absAssets)), cleanup, nil

	case hasExtras:
		return os.DirFS(extraDir), cleanup, nil

	default: // hasAssets only
		absAssets, err := filepath.Abs(assetsDir)
		if err != nil {
			return nil, nil, fmt.Errorf("resolving assets dir: %w", err)
		}
		return os.DirFS(absAssets), nil, nil
	}
}
