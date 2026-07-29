// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

// Package leadtype contains module-wide metadata for Leadtype.
package leadtype

import (
	_ "embed"
	"strings"
)

// Version is the semantic version of this Leadtype build.
//
//go:embed VERSION
var Version string

func init() {
	Version = strings.TrimSpace(Version)
}
