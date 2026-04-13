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
	registerSample("test_025_svg_text_clip", "demonstrate SVG text rendered under an SVG clip path", runTest025SVGTextClip)
}

func runTest025SVGTextClip() (string, error) {
	return writeDoc("test_025_svg_text_clip.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err == nil && ttfc.Len() > 0 {
			doc.AddFontSource(ttfc)
			if _, err := doc.SetFont("Arial", 12, options.Options{}); err == nil {
				doc.MoveTo(0.7, 0.6)
				fmt.Fprintln(doc, "SVG text clip-path demo")
			}
		}

		widthLarge := 4.8
		if _, _, err := doc.PrintSVGFile("pdf/testdata/test_scene_svg_text_clip.svg", 0.8, 1.0, &widthLarge, nil); err != nil {
			return err
		}

		if _, err := doc.SetFont("Arial", 10, options.Options{}); err == nil {
			doc.MoveTo(0.95, 3.95)
			fmt.Fprintln(doc, "Top line is clipped by the rounded rectangle; bottom line is the unclipped reference.")
		}

		widthSmall := 2.9
		if _, _, err := doc.PrintImageFile("pdf/testdata/test_scene_svg_text_clip.svg", 1.75, 4.45, &widthSmall, nil); err != nil {
			return err
		}

		if _, err := doc.SetFont("Arial", 10, options.Options{}); err == nil {
			doc.MoveTo(1.9, 6.75)
			fmt.Fprintln(doc, "Same asset via SVG-compatible image placement.")
		}

		return nil
	})
}
