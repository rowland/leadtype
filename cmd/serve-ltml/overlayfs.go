// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"io/fs"

	"github.com/rowland/leadtype/internal/overlayfs"
)

// overlayFS is an alias for the shared overlay filesystem type. Upper is the
// per-request upload directory; lower is the server-wide static assets directory.
type overlayFS = overlayfs.FS

// newOverlayFS constructs a per-request overlay. upper is typically an
// os.DirFS rooted at the request upload directory; lower is the assets FS.
func newOverlayFS(upper, lower fs.FS) *overlayFS {
	return overlayfs.New(upper, lower)
}
