// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_021_gradients", "demonstrate linear and radial gradient fills", runTest021Gradients)
}

func runTest021Gradients() (string, error) {
	return writeDoc("test_021_gradients.pdf", func(doc *pdf.DocWriter) error {
		afmfc, err := afm_fonts.Default()
		if err != nil {
			return err
		}
		doc.AddFontSource(afmfc)

		doc.NewPage()
		doc.SetUnits("in")
		if _, err := doc.SetFont("Helvetica", 10, options.Options{"color": colors.Black}); err != nil {
			return err
		}

		label := func(x, y float64, text string) {
			doc.ClearFillGradient()
			doc.ClearLineGradient()
			doc.SetFillColor(colors.Black)
			doc.SetLineColor(colors.Black)
			doc.MoveTo(x, y)
			_ = doc.Print(text)
		}

		// Two-stop linear gradient fill.
		label(1, 0.8, "Two-stop linear fill")
		if err := doc.SetFillLinearGradient(&pdf.LinearGradient{
			X0: 1, Y0: 1, X1: 4, Y1: 1,
			Stops: []pdf.GradientStop{
				{Position: 0, Color: colors.Red},
				{Position: 1, Color: colors.Blue},
			},
		}); err != nil {
			return err
		}
		doc.Rectangle(1, 1, 3, 1.5, false, true)

		// Multi-stop linear gradient fill.
		label(1, 2.8, "Multi-stop linear fill")
		if err := doc.SetFillLinearGradient(&pdf.LinearGradient{
			X0: 1, Y0: 3, X1: 4, Y1: 3,
			Stops: []pdf.GradientStop{
				{Position: 0, Color: colors.Red},
				{Position: 0.5, Color: colors.White},
				{Position: 1, Color: colors.Blue},
			},
		}); err != nil {
			return err
		}
		doc.Rectangle(1, 3, 3, 1.5, false, true)

		// Radial gradient fill.
		label(5, 0.8, "Radial fill")
		doc.ClearFillGradient()
		if err := doc.SetFillRadialGradient(&pdf.RadialGradient{
			X0: 6.25, Y0: 1.75, R0: 0,
			X1: 6.25, Y1: 1.75, R1: 1.25,
			Stops: []pdf.GradientStop{
				{Position: 0, Color: colors.Yellow},
				{Position: 1, Color: colors.DarkGreen},
			},
		}); err != nil {
			return err
		}
		doc.Rectangle(5, 1, 2.5, 1.5, false, true)

		// Gradient with border.
		label(5, 2.8, "Linear fill with border")
		if err := doc.SetFillLinearGradient(&pdf.LinearGradient{
			X0: 5, Y0: 3, X1: 7.5, Y1: 4.5,
			Stops: []pdf.GradientStop{
				{Position: 0, Color: colors.Orange},
				{Position: 1, Color: colors.Purple},
			},
		}); err != nil {
			return err
		}
		doc.SetLineColor(colors.Black)
		doc.SetLineWidth(1, "pt")
		doc.Rectangle(5, 3, 2.5, 1.5, true, true)

		// Line gradient without fill.
		label(1, 6.8, "Line gradient stroke")
		doc.ClearFillGradient()
		doc.SetFillColor(colors.White)
		if err := doc.SetLineLinearGradient(&pdf.LinearGradient{
			X0: 1, Y0: 7, X1: 4, Y1: 8,
			Stops: []pdf.GradientStop{
				{Position: 0, Color: colors.DarkBlue},
				{Position: 1, Color: colors.Orange},
			},
		}); err != nil {
			return err
		}
		doc.SetLineWidth(6, "pt")
		doc.Rectangle(1, 7, 3, 1, true, false)
		doc.ClearLineGradient()

		// Thick circle border with a repeating multi-stop line gradient.
		label(5, 7.4, "Circle border with repeating line gradient")
		if err := doc.SetLineLinearGradient(&pdf.LinearGradient{
			X0: 5, Y0: 8.5, X1: 7.5, Y1: 8.5,
			Stops: []pdf.GradientStop{
				{Position: 0, Color: colors.Red},
				{Position: 0.25, Color: colors.Yellow},
				{Position: 0.5, Color: colors.Green},
				{Position: 0.75, Color: colors.Blue},
				{Position: 1, Color: colors.Red},
			},
		}); err != nil {
			return err
		}
		doc.SetLineWidth(18, "pt")
		if err := doc.Circle(6.25, 8.5, 0.85, true, false, false); err != nil {
			return err
		}
		doc.ClearLineGradient()

		// Direct shading into a clipped region.
		label(5, 5.3, "Clipped direct shading")
		if err := doc.Path(func() {
			doc.MoveTo(5, 5.5)
			doc.LineTo(7.5, 5.5)
			doc.LineTo(7.5, 7)
			doc.LineTo(5, 7)
			doc.LineTo(5, 5.5)
			_ = doc.Clip(func() {
				_ = doc.PaintLinearGradient(&pdf.LinearGradient{
					X0: 5, Y0: 5.5, X1: 7.5, Y1: 7,
					Stops: []pdf.GradientStop{
						{Position: 0, Color: colors.White},
						{Position: 1, Color: colors.DarkBlue},
					},
				})
			})
		}); err != nil {
			return err
		}

		// Revert to solid fill.
		label(1, 5.3, "Solid fill after clearing gradient")
		doc.ClearFillGradient()
		doc.SetFillColor(colors.LightGray)
		doc.Rectangle(1, 5.5, 3, 1, false, true)

		return nil
	})
}
