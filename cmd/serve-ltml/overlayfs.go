// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"errors"
	"io/fs"
	"sort"
)

// overlayFS is a read-only composite filesystem that consults the upper
// filesystem first and falls back to the lower filesystem. It is created
// per request; each request's uploads populate the upper FS while the
// server-wide static content lives in the lower FS.
type overlayFS struct {
	upper fs.FS // request upload dir
	lower fs.FS // configured base path
}

// newOverlayFS constructs a per-request overlay. upper is typically an
// os.DirFS rooted at the request temp directory; lower is the static base-path
// FS configured at startup.
func newOverlayFS(upper, lower fs.FS) *overlayFS {
	return &overlayFS{upper: upper, lower: lower}
}

// Open implements fs.FS. It tries upper first and falls back to lower on
// fs.ErrNotExist.
func (o *overlayFS) Open(name string) (fs.File, error) {
	f, err := o.upper.Open(name)
	if err == nil {
		return f, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return o.lower.Open(name)
	}
	return nil, err
}

// ReadFile implements fs.ReadFileFS. It tries upper first.
func (o *overlayFS) ReadFile(name string) ([]byte, error) {
	data, err := fs.ReadFile(o.upper, name)
	if err == nil {
		return data, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return fs.ReadFile(o.lower, name)
	}
	return nil, err
}

// Stat implements fs.StatFS. It tries upper first.
func (o *overlayFS) Stat(name string) (fs.FileInfo, error) {
	info, err := fs.Stat(o.upper, name)
	if err == nil {
		return info, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return fs.Stat(o.lower, name)
	}
	return nil, err
}

// ReadDir implements fs.ReadDirFS. It merges entries from both filesystems;
// entries in upper shadow same-named entries in lower.
func (o *overlayFS) ReadDir(name string) ([]fs.DirEntry, error) {
	upperEntries, upperErr := fs.ReadDir(o.upper, name)
	lowerEntries, lowerErr := fs.ReadDir(o.lower, name)

	// If neither has the directory, return whichever error is more informative.
	if upperErr != nil && lowerErr != nil {
		if errors.Is(upperErr, fs.ErrNotExist) {
			return nil, lowerErr
		}
		return nil, upperErr
	}

	// Build a merged set; upper entries shadow lower.
	seen := make(map[string]struct{})
	var merged []fs.DirEntry
	for _, e := range upperEntries {
		seen[e.Name()] = struct{}{}
		merged = append(merged, e)
	}
	for _, e := range lowerEntries {
		if _, shadowed := seen[e.Name()]; !shadowed {
			merged = append(merged, e)
		}
	}
	// fs.ReadDirFS requires entries to be sorted by name.
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Name() < merged[j].Name()
	})
	return merged, nil
}

var (
	_ fs.FS         = (*overlayFS)(nil)
	_ fs.ReadFileFS = (*overlayFS)(nil)
	_ fs.StatFS     = (*overlayFS)(nil)
	_ fs.ReadDirFS  = (*overlayFS)(nil)
)
