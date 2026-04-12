// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

const (
	advancedGradientSVGPath = "pdf/testdata/test_scene_svg_advanced_gradients.svg"
	advancedMaskSVGPath     = "pdf/testdata/test_scene_svg_mask_blend.svg"
)

func init() {
	registerSample("test_023_svg_advanced", "demonstrate advanced SVG rendering through direct PDF writer APIs", runTest023SVGAdvanced)
}

func runTest023SVGAdvanced() (string, error) {
	return writeDoc("test_023_svg_advanced.pdf", func(doc *pdf.DocWriter) error {
		afm, err := afm_fonts.Default()
		if err != nil {
			return err
		}
		doc.AddFontSource(afm)
		doc.SetUnits("in")
		doc.NewPage()

		if _, err := doc.SetFont("Helvetica", 15, options.Options{}); err != nil {
			return err
		}
		doc.SetFontColor(colors.Black)
		doc.MoveTo(0.65, 0.48)
		fmt.Fprintln(doc, "Advanced SVG rendering")

		if _, err := doc.SetFont("Helvetica", 9.5, options.Options{}); err != nil {
			return err
		}
		doc.SetFontColor(colors.Color(0x4B5B67))
		doc.MoveTo(0.65, 0.72)
		fmt.Fprintln(doc, "Left panels load shared SVG fixtures from disk; right panels reuse the same assets from bytes at a smaller scale.")

		gradientData, err := os.ReadFile(advancedGradientSVGPath)
		if err != nil {
			return err
		}
		maskData, err := os.ReadFile(advancedMaskSVGPath)
		if err != nil {
			return err
		}

		if err := drawSVGRowHeading(doc, 0.65, 1.04, "Gradient fixture: userSpaceOnUse, gradientTransform, opacity, <use>, and clip-path."); err != nil {
			return err
		}
		if err := drawSVGPanel(doc, 0.75, 1.28, 3.15, "PrintSVGFile(...)", func(width float64) (float64, float64, error) {
			return doc.PrintSVGFile(advancedGradientSVGPath, 0.75, 1.56, &width, nil)
		}); err != nil {
			return err
		}
		if err := drawSVGPanel(doc, 4.55, 1.28, 2.35, "PrintSVG([]byte)", func(width float64) (float64, float64, error) {
			return doc.PrintSVG(gradientData, 4.55, 1.56, &width, nil)
		}); err != nil {
			return err
		}

		if err := drawSVGRowHeading(doc, 0.65, 4.08, "Mask fixture: userSpaceOnUse masking, filter-compatible mask markup, and hard-light blending."); err != nil {
			return err
		}
		if err := drawSVGPanel(doc, 0.75, 4.32, 3.15, "PrintSVGFile(...)", func(width float64) (float64, float64, error) {
			return doc.PrintSVGFile(advancedMaskSVGPath, 0.75, 4.60, &width, nil)
		}); err != nil {
			return err
		}
		if err := drawSVGPanel(doc, 4.55, 4.32, 2.35, "PrintSVG([]byte)", func(width float64) (float64, float64, error) {
			return doc.PrintSVG(maskData, 4.55, 4.60, &width, nil)
		}); err != nil {
			return err
		}

		return nil
	})
}

func drawSVGRowHeading(doc *pdf.DocWriter, x, y float64, text string) error {
	if _, err := doc.SetFont("Helvetica", 10, options.Options{}); err != nil {
		return err
	}
	doc.SetFontColor(colors.Color(0x263B48))
	doc.MoveTo(x, y)
	fmt.Fprintln(doc, text)
	return nil
}

func drawSVGPanel(doc *pdf.DocWriter, x, y, width float64, caption string, render func(width float64) (float64, float64, error)) error {
	if _, err := doc.SetFont("Helvetica", 9, options.Options{}); err != nil {
		return err
	}
	doc.SetFontColor(colors.Black)
	doc.MoveTo(x, y)
	fmt.Fprintln(doc, caption)

	actualWidth, actualHeight, err := render(width)
	if err != nil {
		return err
	}

	doc.SetLineColor(colors.LightGray)
	doc.SetLineWidth(0.75, "pt")
	doc.Rectangle(x-0.05, y+0.23, actualWidth+0.10, actualHeight+0.10, true, false)
	return nil
}
