// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_026_svg_text_gradient_clip", "demonstrate SVG text used as a clip path for gradient paint", runTest026SVGTextGradientClip)
}

func runTest026SVGTextGradientClip() (string, error) {
	return writeDoc("test_026_svg_text_gradient_clip.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err == nil && ttfc.Len() > 0 {
			doc.AddFontSource(ttfc)
			if _, err := doc.SetFont("Arial", 12, options.Options{}); err == nil {
				doc.MoveTo(0.7, 0.6)
				fmt.Fprintln(doc, "SVG text as a clip path")
			}
		}

		widthLarge := 4.8
		if _, _, err := doc.PrintSVGFile("pdf/testdata/test_scene_svg_text_gradient_clip.svg", 0.8, 1.0, &widthLarge, nil); err != nil {
			return err
		}

		if _, err := doc.SetFont("Arial", 10, options.Options{}); err == nil {
			doc.MoveTo(1.0, 5.35)
			fmt.Fprintln(doc, "Top panel uses <text> inside <clipPath> to reveal the gradient; bottom panel uses the same-sized word without clipping.")
		}

		widthSmall := 2.9
		if _, _, err := doc.PrintImageFile("pdf/testdata/test_scene_svg_text_gradient_clip.svg", 1.75, 5.8, &widthSmall, nil); err != nil {
			return err
		}

		if _, err := doc.SetFont("Arial", 10, options.Options{}); err == nil {
			doc.MoveTo(1.9, 8.65)
			fmt.Fprintln(doc, "Same asset via SVG-compatible image placement.")
		}

		return nil
	})
}
