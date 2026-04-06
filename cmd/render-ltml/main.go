// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"context"
	"os"

	"github.com/rowland/leadtype/cmd/render-ltml/renderltml"
)

func main() {
	os.Exit(renderltml.Main(context.Background(), os.Args[1:], os.Stderr, nil))
}
