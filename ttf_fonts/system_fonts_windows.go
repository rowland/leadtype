// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf_fonts

import (
	"os"
	"path/filepath"
)

// SystemFontDirs returns the standard font directories on Windows.
func SystemFontDirs() []string {
	dirs := []string{
		filepath.Join(os.Getenv("SystemRoot"), "Fonts"),
	}
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		dirs = append(dirs, filepath.Join(localAppData, "Microsoft", "Windows", "Fonts"))
	}
	return dirs
}
