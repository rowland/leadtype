// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_014_images", "demonstrate JPEG image placement from file and memory", runTest014Images)
}

func drawImageFrame(doc *pdf.DocWriter, x, y, width, height float64, label string) {
	doc.SetLineColor(colors.LightGray)
	doc.Rectangle(x, y, width, height, true, false)
	doc.SetFontColor(colors.DimGray)
	doc.MoveTo(x, y-0.12)
	fmt.Fprintln(doc, label)
}

func runTest014Images() (string, error) {
	return writeDoc("test_014_images.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err == nil && ttfc.Len() > 0 {
			doc.AddFontSource(ttfc)
			if _, err := doc.SetFont("Arial", 12, options.Options{}); err == nil {
				doc.MoveTo(0.7, 0.6)
				fmt.Fprintln(doc, "JPEG image placement demo")
			}
		}

		imagePath := "pdf/testdata/testimg.jpg"
		data, err := os.ReadFile(imagePath)
		if err != nil {
			return err
		}

		doc.SetFont("Arial", 10, options.Options{})

		drawImageFrame(doc, 0.8, 1.3, 2.1, 2.0, "File input, width only")
		widthOnly := 1.8
		if _, _, err := doc.PrintImageFile(imagePath, 0.95, 1.45, &widthOnly, nil); err != nil {
			return err
		}

		drawImageFrame(doc, 3.3, 1.3, 2.1, 2.0, "In-memory input, height only")
		heightOnly := 1.4
		if _, _, err := doc.PrintImage(data, 3.55, 1.45, nil, &heightOnly); err != nil {
			return err
		}

		drawImageFrame(doc, 0.8, 4.0, 2.1, 2.2, "Intrinsic size")
		if _, _, err := doc.PrintImageFile(imagePath, 1.05, 4.25, nil, nil); err != nil {
			return err
		}

		drawImageFrame(doc, 3.3, 4.0, 2.1, 2.2, "Explicit width and height")
		explicitWidth, explicitHeight := 1.7, 1.1
		if _, _, err := doc.PrintImage(data, 3.55, 4.4, &explicitWidth, &explicitHeight); err != nil {
			return err
		}

		return nil
	})
}
