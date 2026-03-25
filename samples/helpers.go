// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/pdf"
)

func sampleOutputPath(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return name
	}
	return filepath.Join(filepath.Dir(file), name)
}

func writeDoc(outputName string, build func(*pdf.DocWriter) error) (string, error) {
	outputPath := sampleOutputPath(outputName)

	f, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	doc := pdf.NewDocWriter()
	if err := build(doc); err != nil {
		return "", err
	}
	if _, err := doc.WriteTo(f); err != nil {
		return "", err
	}
	return outputPath, nil
}

func openFile(path string, openArgs ...string) error {
	args := append(append([]string(nil), openArgs...), path)
	cmd := exec.Command("open", args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func mustHaveFonts(fonts []*font.Font, err error) error {
	if err != nil {
		return err
	}
	if len(fonts) == 0 {
		return fmt.Errorf("no fonts returned")
	}
	return nil
}
