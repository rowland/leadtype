// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

// render-ltml renders LTML documents to PDF locally or submits them to a
// remote serve-ltml service.
//
// The reusable implementation lives in package
// github.com/rowland/leadtype/cmd/render-ltml/renderltml.
package main

import (
	"context"
	"os"

	"github.com/rowland/leadtype/cmd/render-ltml/renderltml"
)

func main() {
	os.Exit(renderltml.Main(context.Background(), os.Args[1:], os.Stderr, nil))
}
