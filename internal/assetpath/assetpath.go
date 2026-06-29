// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

// Package assetpath validates virtual asset paths used with fs.FS-backed assets.
package assetpath

import "io/fs"

// Valid reports whether name is a non-empty, clean, relative fs.FS path
// suitable for a virtual asset. It rejects "." because it names the filesystem
// root.
func Valid(name string) bool {
	return name != "" && name != "." && fs.ValidPath(name)
}
