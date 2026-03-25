// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_005_ttf_fonts", "list system TrueType fonts that can render their own label", runTest005TtfFonts)
}

func runTest005TtfFonts() (string, error) {
	return writeDoc("test_005_ttf_fonts.pdf", func(doc *pdf.DocWriter) error {
		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err != nil {
			return err
		}
		doc.AddFontSource(ttfc)

		doc.SetUnits("in")
		doc.SetLineSpacing(1.5)
		doc.MoveTo(1, 1)

		for _, info := range ttfc.FontInfos {
			if doc.Y() > 10 {
				doc.NewPage()
				doc.MoveTo(1, 1)
			}
			if _, err := doc.SetFont(info.Family(), 12, options.Options{"style": info.Style()}); err != nil {
				fmt.Fprintf(os.Stderr, "Setting font to %s/%s: %v\n", info.Family(), info.Style(), err)
				continue
			}
			label := info.FullName()
			fonts, err := doc.SetFont(info.Family(), 12, options.Options{"style": info.Style()})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Setting font to %s/%s: %v\n", info.Family(), info.Style(), err)
				continue
			}
			if !canRender(label, fonts...) {
				continue
			}
			fmt.Fprintln(doc, label)
		}
		return nil
	})
}

func canRender(text string, fonts ...*font.Font) bool {
	if len(fonts) == 0 || fonts[0] == nil {
		return false
	}
	for _, r := range text {
		if !fonts[0].HasRune(r) {
			return false
		}
	}
	return true
}
