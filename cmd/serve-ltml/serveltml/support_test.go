// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package serveltml

import (
	"strings"
	"testing"
)

// TestGenerateULID_Length verifies that every ULID is exactly 26 characters.
func TestGenerateULID_Length(t *testing.T) {
	for i := 0; i < 10; i++ {
		ulid, err := generateULID()
		if err != nil {
			t.Fatalf("generateULID() error: %v", err)
		}
		if len(ulid) != 26 {
			t.Errorf("len=%d, want 26; got %q", len(ulid), ulid)
		}
	}
}

// TestGenerateULID_Charset verifies that every character is in the Crockford
// base-32 alphabet.
func TestGenerateULID_Charset(t *testing.T) {
	const alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	for i := 0; i < 20; i++ {
		ulid, err := generateULID()
		if err != nil {
			t.Fatalf("generateULID() error: %v", err)
		}
		for j, ch := range ulid {
			if !strings.ContainsRune(alphabet, ch) {
				t.Errorf("ULID[%d]=%q is not in Crockford alphabet; ULID=%q", j, ch, ulid)
			}
		}
	}
}

// TestGenerateULID_Uniqueness verifies that 1000 generated ULIDs are all distinct.
func TestGenerateULID_Uniqueness(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		ulid, err := generateULID()
		if err != nil {
			t.Fatalf("generateULID() error: %v", err)
		}
		if seen[ulid] {
			t.Fatalf("duplicate ULID generated: %q", ulid)
		}
		seen[ulid] = true
	}
}

// TestLtmlFilename covers the extension-replacement helper.
func TestLtmlFilename(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"report.pdf", "report.ltml"},
		{"output.PDF", "output.ltml"},
		{"archive.tar.gz", "archive.tar.ltml"},
		{"noext", "noext.ltml"},
		{"sub/report.pdf", "sub/report.ltml"},
	}
	for _, tc := range cases {
		got := ltmlFilename(tc.input)
		if got != tc.want {
			t.Errorf("ltmlFilename(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// TestValidateOutputFilename covers valid and invalid X-Output-File values.
func TestValidateOutputFilename(t *testing.T) {
	valid := []string{
		"report.pdf",
		"output.pdf",
		"sub/report.pdf",
		"a/b/c.pdf",
	}
	for _, name := range valid {
		t.Run("valid:"+name, func(t *testing.T) {
			if err := validateOutputFilename(name); err != nil {
				t.Errorf("unexpected error for %q: %v", name, err)
			}
		})
	}

	invalid := []string{
		"",
		".",
		"./report.pdf",
		"/etc/passwd",
		"../escape",
		"a/../report.pdf",
		"a/..",
	}
	for _, name := range invalid {
		t.Run("invalid:"+name, func(t *testing.T) {
			if err := validateOutputFilename(name); err == nil {
				t.Errorf("expected error for %q, got nil", name)
			}
		})
	}
}

// TestValidateUploadFilename covers the filename validation helper directly.
func TestValidateUploadFilename(t *testing.T) {
	tmpDir := t.TempDir()

	good := []string{
		"logo.png",
		"assets/logo.png",
		"a/b/c.txt",
	}
	for _, name := range good {
		t.Run("valid:"+name, func(t *testing.T) {
			_, err := validateUploadFilename(name, tmpDir)
			if err != nil {
				t.Errorf("unexpected error for %q: %v", name, err)
			}
		})
	}

	bad := []string{
		"",
		".",
		"./logo.png",
		"/etc/passwd",
		"../escape",
		"a/../logo.png",
		"a/..",
		"a/../../escape",
	}
	for _, name := range bad {
		t.Run("invalid:"+name, func(t *testing.T) {
			_, err := validateUploadFilename(name, tmpDir)
			if err == nil {
				t.Errorf("expected error for %q, got nil", name)
			}
		})
	}
}
