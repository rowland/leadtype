// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"os"

	"github.com/rowland/leadtype/cmd/serve-ltml/serveltml"
)

func main() {
	os.Exit(serveltml.Main(os.Stderr, nil))
}
