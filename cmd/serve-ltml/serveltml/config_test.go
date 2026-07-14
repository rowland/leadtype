// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package serveltml

import (
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/rowland/leadtype/ttf_fonts"
)

func TestParseConfigFontDirectories(t *testing.T) {
	if os.Getenv("LEADTYPE_CONFIG_HELPER") == "1" {
		runParseConfigFontDirectoriesHelper(t)
		return
	}

	assets := t.TempDir()
	first := t.TempDir()
	second := t.TempDir()
	for _, mode := range []string{"flag", "environment", "default"} {
		cmd := exec.Command(os.Args[0], "-test.run=TestParseConfigFontDirectories$")
		cmd.Env = append(os.Environ(),
			"LEADTYPE_CONFIG_HELPER=1",
			"LEADTYPE_CONFIG_MODE="+mode,
			"LEADTYPE_CONFIG_ASSETS="+assets,
			"LEADTYPE_CONFIG_FIRST="+first,
			"LEADTYPE_CONFIG_SECOND="+second,
		)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s configuration failed: %v\n%s", mode, err, output)
		}
	}
}

func runParseConfigFontDirectoriesHelper(t *testing.T) {
	assets := os.Getenv("LEADTYPE_CONFIG_ASSETS")
	first := os.Getenv("LEADTYPE_CONFIG_FIRST")
	second := os.Getenv("LEADTYPE_CONFIG_SECOND")
	mode := os.Getenv("LEADTYPE_CONFIG_MODE")

	var spec string
	switch mode {
	case "flag":
		spec = first + ",auto," + second
		os.Args = []string{"serve-ltml", "-assets", assets, "-font-dir", spec}
		os.Unsetenv("FONT_DIR")
	case "environment":
		spec = second + "," + first
		os.Args = []string{"serve-ltml", "-assets", assets}
		os.Setenv("FONT_DIR", spec)
	case "default":
		spec = "auto"
		os.Args = []string{"serve-ltml", "-assets", assets}
		os.Unsetenv("FONT_DIR")
	default:
		t.Fatalf("unknown helper mode %q", mode)
	}

	want, err := ttf_fonts.ResolveFontDirs(spec)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := parseConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.FontDirs, want) {
		t.Fatalf("FontDirs = %#v, want %#v", cfg.FontDirs, want)
	}
}
