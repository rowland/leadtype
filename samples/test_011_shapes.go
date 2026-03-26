// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_011_shapes", "demonstrate higher-level PDF shape primitives", runTest011Shapes)
}

func runTest011Shapes() (string, error) {
	return writeDoc("test_011_shapes.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err == nil && ttfc.Len() > 0 {
			doc.AddFontSource(ttfc)
			if _, err := doc.SetFont("Arial", 12, options.Options{}); err == nil {
				doc.MoveTo(0.75, 0.6)
				fmt.Fprintln(doc, "Shape primitives demo")
			}
		}

		doc.SetLineWidth(2, "pt")

		doc.SetLineColor(colors.DarkBlue)
		if err := doc.Circle(1.5, 1.8, 0.55, true, false, false); err != nil {
			return err
		}

		doc.SetLineColor(colors.FireBrick)
		if err := doc.Ellipse(3.5, 1.8, 0.8, 0.45, true, false, false); err != nil {
			return err
		}

		doc.SetLineColor(colors.DarkGreen)
		doc.SetFillColor(colors.LightYellow)
		if err := doc.Pie(5.5, 1.8, 0.65, 20, 140, true, false, false); err != nil {
			return err
		}

		doc.SetLineColor(colors.Purple)
		if err := doc.Arch(1.6, 4.2, 0.4, 0.8, 220, 20, true, false, false); err != nil {
			return err
		}

		doc.SetLineColor(colors.Teal)
		if err := doc.Polygon(3.5, 4.1, 0.7, 6, true, false, false, 0); err != nil {
			return err
		}

		doc.SetLineColor(colors.OrangeRed)
		if err := doc.Star(5.5, 4.1, 0.75, 0.32, 5, true, false, false, 0); err != nil {
			return err
		}

		doc.SetLineColor(colors.DarkBlue)
		doc.SetFillColor(colors.LightCyan)
		if err := doc.Circle(1.5, 6.0, 0.55, true, true, false); err != nil {
			return err
		}

		doc.SetLineColor(colors.FireBrick)
		doc.SetFillColor(colors.MistyRose)
		if err := doc.Ellipse(3.5, 6.0, 0.8, 0.45, true, true, false); err != nil {
			return err
		}

		doc.SetLineColor(colors.DarkGreen)
		doc.SetFillColor(colors.LightYellow)
		if err := doc.Pie(5.5, 6.0, 0.65, 20, 140, true, true, false); err != nil {
			return err
		}

		doc.SetLineColor(colors.Purple)
		doc.SetFillColor(colors.Lavender)
		if err := doc.Arch(1.6, 8.15, 0.4, 0.8, 220, 20, true, true, false); err != nil {
			return err
		}

		doc.SetLineColor(colors.Teal)
		doc.SetFillColor(colors.PaleTurquoise)
		if err := doc.Polygon(3.5, 8.0, 0.7, 6, true, true, false, 0); err != nil {
			return err
		}

		doc.SetLineColor(colors.OrangeRed)
		doc.SetFillColor(colors.LightGoldenRodYellow)
		if err := doc.Star(5.5, 8.0, 0.75, 0.32, 5, true, true, false, 0); err != nil {
			return err
		}

		doc.SetLineColor(colors.Black)
		doc.Line(0.8, 10.0, 20, 1.6)
		doc.Line(0.8, 10.0, -20, 1.6)

		return nil
	})
}
