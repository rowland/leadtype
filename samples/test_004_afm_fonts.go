// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_004_afm_fonts", "list built-in AFM fonts", runTest004AfmFonts, "-a", "Firefox")
}

func runTest004AfmFonts() (string, error) {
	return writeDoc("test_004_afm_fonts.pdf", func(doc *pdf.DocWriter) error {
		afmfc, err := afm_fonts.Default()
		if err != nil {
			return err
		}
		doc.AddFontSource(afmfc)

		doc.SetUnits("in")
		doc.SetLineSpacing(1.5)
		doc.MoveTo(1, 1)

		for _, info := range afmfc.FontInfos {
			if doc.Y() > 10 {
				doc.NewPage()
				doc.MoveTo(1, 1)
			}
			if _, err := doc.SetFont(info.PostScriptName(), 12, options.Options{}); err != nil {
				return err
			}
			fmt.Fprintln(doc, info.FullName())
		}
		return nil
	})
}
