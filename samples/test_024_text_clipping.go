// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_024_text_clipping", "recreate the classic EideticPDF text clipping demo with a photo inside large text", runTest024TextClipping)
}

func runTest024TextClipping() (string, error) {
	return writeDoc("test_024_text_clipping.pdf", func(doc *pdf.DocWriter) error {
		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err != nil {
			return err
		}
		doc.AddFontSource(ttfc)

		doc.NewPage()
		doc.SetUnits("in")

		doc.SetVTextAlign("top")
		if _, err := doc.SetFont("Arial", 14, options.Options{}); err != nil {
			return err
		}
		doc.SetUnderline(true) // TODO: make work with VTextAlign=top
		doc.MoveTo(0.5, 0.5)
		if err := doc.Print("Text Clipping"); err != nil {
			return err
		}
		doc.SetUnderline(false)

		imagePath := "pdf/testdata/testimg.jpg"

		if _, err := doc.SetFont("Helvetica", 144, options.Options{"weight": "bold"}); err != nil {
			return err
		}
		doc.SetLineWidth(1, "pt")

		x, y, w := 1.0, 2.0, 6.5
		doc.MoveTo(x+0.25, y+1.5)
		var clipErr error
		if err := doc.FillStrokeClipText("ROSE", func() {
			_, _, clipErr = doc.PrintImageFile(imagePath, x, y, &w, nil)
		}); err != nil {
			return err
		}
		if clipErr != nil {
			return clipErr
		}

		y += 4.0
		doc.PrintImageFile(imagePath, x, y, &w, nil)
		doc.MoveTo(x+0.25, y+1.5)
		doc.Print("ROSE")

		return nil
	})
}
