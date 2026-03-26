// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_012_transforms", "demonstrate scoped PDF rotate and scale transforms", runTest012Transforms)
}

func drawPivot(doc *pdf.DocWriter, x, y float64) {
	saveWidth := doc.SetLineWidth(1, "pt")
	doc.SetLineColor(colors.Gray)
	_ = doc.Path(func() {
		doc.MoveTo(x-0.09, y)
		doc.LineTo(x+0.09, y)
		_ = doc.Stroke()
	})
	_ = doc.Path(func() {
		doc.MoveTo(x, y-0.09)
		doc.LineTo(x, y+0.09)
		_ = doc.Stroke()
	})
	doc.SetLineWidth(saveWidth, "pt")
}

func runTest012Transforms() (string, error) {
	return writeDoc("test_012_transforms.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err == nil && ttfc.Len() > 0 {
			doc.AddFontSource(ttfc)
			if _, err := doc.SetFont("Arial", 12, options.Options{}); err == nil {
				doc.MoveTo(0.7, 0.6)
				fmt.Fprintln(doc, "Scoped transform demo")
			}
		}

		doc.SetLineWidth(2, "pt")
		doc.SetFont("Arial", 11, options.Options{})

		doc.SetFontColor(colors.DimGray)
		doc.MoveTo(0.9, 1.05)
		fmt.Fprintln(doc, "Rotate around pivot")
		doc.MoveTo(3.8, 1.05)
		fmt.Fprintln(doc, "Scale around pivot")

		doc.SetLineColor(colors.LightGray)
		doc.SetFillColor(colors.White)
		doc.Rectangle(0.95, 1.35, 1.3, 0.55, true, false)
		doc.Rectangle(3.95, 1.35, 1.1, 0.55, true, false)

		doc.SetFontColor(colors.LightGray)
		doc.MoveTo(1.2, 1.7)
		fmt.Fprintln(doc, "Reference")
		doc.MoveTo(4.15, 1.7)
		fmt.Fprintln(doc, "Reference")

		doc.SetLineColor(colors.FireBrick)
		doc.SetFillColor(colors.MistyRose)
		doc.SetFontColor(colors.FireBrick)
		if err := doc.Rotate(25, 1.6, 1.625, func() {
			doc.Rectangle(0.95, 1.35, 1.3, 0.55, true, true)
			doc.MoveTo(1.16, 1.69)
			fmt.Fprintln(doc, "Rotated")
		}); err != nil {
			return err
		}

		doc.SetLineColor(colors.DarkBlue)
		doc.SetFillColor(colors.LightCyan)
		doc.SetFontColor(colors.DarkBlue)
		if err := doc.Scale(4.5, 1.625, 1.35, 0.7, func() {
			doc.Rectangle(3.95, 1.35, 1.1, 0.55, true, true)
			doc.MoveTo(4.15, 1.69)
			fmt.Fprintln(doc, "Scaled")
		}); err != nil {
			return err
		}
		drawPivot(doc, 1.6, 1.625)
		drawPivot(doc, 4.5, 1.625)

		doc.SetFontColor(colors.DimGray)
		doc.MoveTo(0.9, 3.0)
		fmt.Fprintln(doc, "Reference and transformed star")

		doc.SetLineColor(colors.LightGray)
		doc.SetFillColor(colors.White)
		_ = doc.Star(1.7, 4.2, 0.42, 0.19, 5, true, false, false, 0)

		doc.SetLineColor(colors.Teal)
		doc.SetFillColor(colors.PaleTurquoise)
		if err := doc.Scale(4.8, 4.2, 1.45, 1.45, func() {
			_ = doc.Star(4.8, 4.2, 0.42, 0.19, 5, true, true, false, 0)
		}); err != nil {
			return err
		}
		drawPivot(doc, 1.7, 4.2)
		drawPivot(doc, 4.8, 4.2)

		doc.SetLineColor(colors.Purple)
		doc.SetFontColor(colors.Purple)
		drawPivot(doc, 3.25, 6.25)
		if err := doc.Rotate(-28, 3.25, 6.25, func() {
			doc.Rectangle(2.6, 5.95, 1.3, 0.6, true, false)
			doc.MoveTo(2.83, 6.3)
			fmt.Fprintln(doc, "Rotated text")
		}); err != nil {
			return err
		}
		drawPivot(doc, 3.25, 6.25)

		return nil
	})
}
