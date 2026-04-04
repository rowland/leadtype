// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_022_emoji", "render emoji-style symbols using an outline-capable system font", runTest022Emoji)
}

var emojiLines = []struct {
	label string
	text  string
}{
	{"Faces", "☺ ☹ ☺ ❤ ☎"},
	{"Travel", "✈ ✉ ☀ ✓ ❤"},
	{"Symbols", "☀ ✈ ✉ ✓ ☺"},
	{"Mixed text", "Leadtype says hello ☺ with hearts ❤ and mail ✉"},
}

func runTest022Emoji() (string, error) {
	return writeDoc("test_022_emoji.pdf", func(dw *pdf.DocWriter) error {
		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err != nil {
			return err
		}
		if ttfc.Len() == 0 {
			return fmt.Errorf("no TTF fonts found in system font directories")
		}

		preferredFamilies := []string{
			"Arial Unicode MS",
			"Menlo",
			"Apple Symbols",
			"Zapf Dingbats",
		}
		family := ""
		hasEmoji := func(name string) bool {
			font, err := ttfc.Select(name, "", "", nil)
			if err != nil {
				return false
			}
			for _, r := range []rune{'☺', '❤', '✈', '✉', '☀'} {
				if font.GlyphIndex(r) == 0 {
					return false
				}
			}
			return true
		}
		for _, name := range preferredFamilies {
			if hasEmoji(name) {
				family = name
				break
			}
		}
		if family == "" {
			for _, fi := range ttfc.FontInfos {
				if fi == nil || fi.Family() == "" {
					continue
				}
				if hasEmoji(fi.Family()) {
					family = fi.Family()
					break
				}
			}
		}
		if family == "" {
			return fmt.Errorf("no outline-capable emoji-symbol font found; try `go run ./ttquery emoji` to inspect available fonts")
		}

		dw.AddFontSource(ttfc)
		dw.SetUnits("in")
		dw.NewPage()
		dw.MoveTo(1, 1)

		if _, err := dw.SetFont(family, 18, options.Options{}); err != nil {
			return err
		}
		fmt.Fprintln(dw, "Emoji Rendering Sample")
		fmt.Fprintln(dw)

		if _, err := dw.SetFont(family, 11, options.Options{}); err != nil {
			return err
		}
		fmt.Fprintf(dw, "Font: %s\n", family)
		fmt.Fprintln(dw, "Using monochrome emoji-style symbols supported by the current outline font path.")
		fmt.Fprintln(dw)

		dw.SetLineSpacing(1.5)
		for _, line := range emojiLines {
			if _, err := dw.SetFont(family, 10, options.Options{}); err != nil {
				return err
			}
			fmt.Fprintf(dw, "%s:\n", line.label)
			if _, err := dw.SetFont(family, 20, options.Options{}); err != nil {
				return err
			}
			fmt.Fprintln(dw, line.text)
			fmt.Fprintln(dw)
		}
		return nil
	})
}
