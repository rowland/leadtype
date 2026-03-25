// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_003_hello_world", "render hello world text with AFM and TTF fonts", runTest003HelloWorld, "-a", "Firefox")
}

func runTest003HelloWorld() (string, error) {
	return writeDoc("test_003_hello_world.pdf", func(doc *pdf.DocWriter) error {
		afmfc, err := afm_fonts.Default()
		if err != nil {
			return err
		}
		doc.AddFontSource(afmfc)

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err != nil {
			return err
		}
		doc.AddFontSource(ttfc)

		doc.NewPage()
		doc.SetUnits("in")

		if _, err := doc.SetFont("Helvetica", 12, options.Options{}); err != nil {
			return err
		}
		doc.MoveTo(1, 1)
		doc.Print("Hello, World!")

		if _, err := doc.SetFont("Arial", 14, options.Options{}); err != nil {
			return err
		}
		doc.MoveTo(2, 2)
		doc.Print("Goodbye, cruel world...")
		return nil
	})
}
