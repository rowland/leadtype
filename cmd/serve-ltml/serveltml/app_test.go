// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package serveltml

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	leadtype "github.com/rowland/leadtype"
)

func TestMainVersion(t *testing.T) {
	originalArgs := os.Args
	os.Args = []string{"serve-ltml", "--version"}
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	var output bytes.Buffer
	if code := Main(&output, nil); code != 0 {
		t.Fatalf("Main() = %d, want 0", code)
	}
	want := fmt.Sprintf("serve-ltml %s\n", leadtype.Version)
	if got := output.String(); got != want {
		t.Fatalf("version output = %q, want %q", got, want)
	}
}
