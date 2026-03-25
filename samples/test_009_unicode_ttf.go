// Copyright 2024 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_009_unicode_ttf", "demonstrate multi-script Unicode rendering with TTF fonts", runTest009UnicodeTtf)
}

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

func runTest009UnicodeTtf() (string, error) {
	return writeDoc("test_009_unicode_ttf.pdf", func(dw *pdf.DocWriter) error {
		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err != nil {
			return err
		}
		if ttfc.Len() == 0 {
			return fmt.Errorf("no TTF fonts found in system font directories")
		}

		preferredFonts := []string{"Arial Unicode MS", "Hiragino Sans", "Noto Sans", "Arial"}
		family := ""
		for _, name := range preferredFonts {
			if _, err := ttfc.Select(name, "", "", nil); err == nil {
				family = name
				break
			}
		}
		if family == "" && len(ttfc.FontInfos) > 0 {
			family = ttfc.FontInfos[0].Family()
		}
		if family == "" {
			return fmt.Errorf("no usable TTF font family found")
		}

		dw.AddFontSource(ttfc)
		dw.SetUnits("in")
		dw.NewPage()
		dw.MoveTo(1, 1)

		if _, err := dw.SetFont(family, 18, options.Options{}); err != nil {
			return err
		}
		fmt.Fprintln(dw, "Unicode Multi-Script Rendering (Type0/CIDFontType2)")
		fmt.Fprintln(dw)

		if _, err := dw.SetFont(family, 13, options.Options{}); err != nil {
			return err
		}
		dw.SetLineSpacing(1.6)

		for _, s := range scripts {
			if dw.Y() > 10 {
				dw.NewPage()
				dw.MoveTo(1, 1)
			}
			if _, err := dw.SetFont(family, 9, options.Options{}); err != nil {
				return err
			}
			fmt.Fprintf(dw, "%s:\n", s.label)
			if _, err := dw.SetFont(family, 13, options.Options{}); err != nil {
				return err
			}
			fmt.Fprintln(dw, s.text)
		}
		return nil
	})
}
