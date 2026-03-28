// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

// Package overlayfs provides a read-only composite fs.FS that resolves names
// against an upper filesystem first and falls back to a lower filesystem.
package overlayfs

import (
	"errors"
	"io/fs"
	"sort"
)

// FS is a read-only composite filesystem. Lookups are tried against upper
// first; on fs.ErrNotExist the lookup falls through to lower.
type FS struct {
	upper fs.FS
	lower fs.FS
}

// New returns an FS that consults upper before lower for every lookup.
func New(upper, lower fs.FS) *FS {
	return &FS{upper: upper, lower: lower}
}

// Open implements fs.FS.
func (o *FS) Open(name string) (fs.File, error) {
	f, err := o.upper.Open(name)
	if err == nil {
		return f, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return o.lower.Open(name)
	}
	return nil, err
}

// ReadFile implements fs.ReadFileFS.
func (o *FS) ReadFile(name string) ([]byte, error) {
	data, err := fs.ReadFile(o.upper, name)
	if err == nil {
		return data, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return fs.ReadFile(o.lower, name)
	}
	return nil, err
}

// Stat implements fs.StatFS.
func (o *FS) Stat(name string) (fs.FileInfo, error) {
	info, err := fs.Stat(o.upper, name)
	if err == nil {
		return info, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return fs.Stat(o.lower, name)
	}
	return nil, err
}

// ReadDir implements fs.ReadDirFS. Entries from upper shadow same-named entries
// from lower; the merged result is sorted by name as required by the contract.
func (o *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	upperEntries, upperErr := fs.ReadDir(o.upper, name)
	lowerEntries, lowerErr := fs.ReadDir(o.lower, name)

	if upperErr != nil && lowerErr != nil {
		if errors.Is(upperErr, fs.ErrNotExist) {
			return nil, lowerErr
		}
		return nil, upperErr
	}

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
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Name() < merged[j].Name()
	})
	return merged, nil
}

var (
	_ fs.FS         = (*FS)(nil)
	_ fs.ReadFileFS = (*FS)(nil)
	_ fs.StatFS     = (*FS)(nil)
	_ fs.ReadDirFS  = (*FS)(nil)
)
