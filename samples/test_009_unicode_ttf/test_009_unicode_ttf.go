// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

// Command test_009_unicode_ttf demonstrates multi-script Unicode rendering
// using the Type0/CIDFontType2 composite font path introduced in the unicode
// mode of DocWriter. A single font call per line renders Latin, Greek,
// Cyrillic, and CJK text without any codepage splitting.
//
// Run from the module root:
//
//	go run samples/test_009_unicode_ttf/test_009_unicode_ttf.go
package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

const outputName = "test_009_unicode_ttf.pdf"

// scripts lists the text samples to render, each on its own line.
// Fonts that lack a script will show .notdef boxes for those glyphs.
var scripts = []struct {
	label string
	text  string
}{
	{"Latin", "The quick brown fox jumps over the lazy dog."},
	{"Latin + diacritics", "Ça va? Ñoño. Üniversität. Ångström."},
	{"Greek capitals", "ΑΒΓΔΕΖΗΘΙΚΛΜΝΞΟΠΡΣΤΥΦΧΨΩ"},
	{"Greek lowercase", "αβγδεζηθικλμνξοπρστυφχψω"},
	{"Cyrillic", "Съешь же ещё этих мягких французских булок."},
	{"Mixed Latin + Greek", "E = mc² and α + β = γ"},
	{"CJK sample", "中文 日本語 한국어"},
	{"Emoji (may show boxes)", "Hello 🌍 world 🎉"},
}

func main() {
	ttfc, err := ttf_fonts.New("/Library/Fonts/*.ttf")
	if err != nil || ttfc.Len() == 0 {
		fmt.Fprintln(os.Stderr, "No TTF fonts found in /Library/Fonts — try installing system fonts.")
		os.Exit(1)
	}

	// Prefer a broad-coverage font; fall back to whatever is available.
	preferredFonts := []string{"Arial Unicode MS", "Hiragino Sans", "Noto Sans", "Arial"}
	family := ""
	for _, name := range preferredFonts {
		if _, err := ttfc.Select(name, "", "", nil); err == nil {
			family = name
			break
		}
	}
	if family == "" {
		fmt.Fprintln(os.Stderr, "None of the preferred fonts found; using first available.")
		if len(ttfc.FontInfos) > 0 {
			family = ttfc.FontInfos[0].Family()
		}
	}
	fmt.Printf("Using font: %s\n", family)

	dw := pdf.NewDocWriterUnicode()
	dw.AddFontSource(ttfc)
	dw.SetUnits("in")

	dw.NewPage()
	dw.MoveTo(1, 1)

	// Title
	if _, err := dw.SetFont(family, 18, options.Options{}); err != nil {
		fmt.Fprintf(os.Stderr, "SetFont(%s, 18): %v\n", family, err)
		os.Exit(1)
	}
	fmt.Fprintln(dw, "Unicode Multi-Script Rendering (Type0/CIDFontType2)")
	fmt.Fprintln(dw)

	if _, err := dw.SetFont(family, 13, options.Options{}); err != nil {
		fmt.Fprintf(os.Stderr, "SetFont(%s, 13): %v\n", family, err)
		os.Exit(1)
	}
	dw.SetLineSpacing(1.6)

	for _, s := range scripts {
		if dw.Y() > 10 {
			dw.NewPage()
			dw.MoveTo(1, 1)
		}
		// Script label in a smaller size
		dw.SetFont(family, 9, options.Options{})
		fmt.Fprintf(dw, "%s:\n", s.label)
		// Script text in the main size
		dw.SetFont(family, 13, options.Options{})
		fmt.Fprintln(dw, s.text)
	}

	f, err := os.Create(outputName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dw.WriteTo(f)
	f.Close()
	fmt.Printf("Wrote %s\n", outputName)
	exec.Command("open", outputName).Start()
}
