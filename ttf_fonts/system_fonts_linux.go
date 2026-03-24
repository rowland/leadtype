// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ttf_fonts

import (
	"os"
	"path/filepath"
)

// SystemFontDirs returns the standard font directories on Linux.
func SystemFontDirs() []string {
	dirs := []string{
		"/usr/share/fonts",
		"/usr/local/share/fonts",
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs,
			filepath.Join(home, ".local", "share", "fonts"),
			filepath.Join(home, ".fonts"),
		)
	}
	return dirs
}
