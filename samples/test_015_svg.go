// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_015_svg", "demonstrate reusable SVG placement through direct and image-compatible PDF APIs", runTest015SVG)
}

func runTest015SVG() (string, error) {
	return writeDoc("test_015_svg.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err == nil && ttfc.Len() > 0 {
			doc.AddFontSource(ttfc)
			if _, err := doc.SetFont("Arial", 12, options.Options{}); err == nil {
				doc.MoveTo(0.7, 0.6)
				fmt.Fprintln(doc, "SVG rendering demo")
			}
		}

		data, err := os.ReadFile("pdf/testdata/test_scene.svg")
		if err != nil {
			return err
		}

		widthLarge := 2.4
		if _, _, err := doc.PrintSVG(data, 0.8, 1.1, &widthLarge, nil); err != nil {
			return err
		}
		widthMedium := 1.6
		if _, _, err := doc.PrintSVG(data, 4.0, 1.3, &widthMedium, nil); err != nil {
			return err
		}
		heightSmall := 0.95
		if _, _, err := doc.PrintImageFile("pdf/testdata/test_scene.svg", 0.9, 4.6, nil, &heightSmall); err != nil {
			return err
		}
		widthSmall := 1.15
		if _, _, err := doc.PrintImageFile("pdf/testdata/test_scene.svg", 2.4, 4.5, &widthSmall, nil); err != nil {
			return err
		}
		widthTiny := 0.9
		if _, _, err := doc.PrintSVG(data, 4.1, 4.55, &widthTiny, nil); err != nil {
			return err
		}

		doc.NewPage()
		if _, err := doc.SetFont("Arial", 12, options.Options{}); err == nil {
			doc.MoveTo(0.7, 0.6)
			fmt.Fprintln(doc, "Same SVG asset reused again on page 2")
		}
		widthHero := 3.6
		if _, _, err := doc.PrintImageFile("pdf/testdata/test_scene.svg", 1.2, 1.2, &widthHero, nil); err != nil {
			return err
		}

		return nil
	})
}
