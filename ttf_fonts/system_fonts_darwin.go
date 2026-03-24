// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf_fonts

import (
	"os"
	"path/filepath"
)

// SystemFontDirs returns the standard font directories on macOS.
func SystemFontDirs() []string {
	dirs := []string{}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "Library", "Fonts"))
	}
	dirs = append(dirs,
		"/Library/Fonts",
		"/System/Library/Fonts/Supplemental",
		"/System/Library/Fonts",
	)
	return dirs
}
