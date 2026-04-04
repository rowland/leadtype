// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_021_gradients", "demonstrate linear and radial gradient fills", runTest021Gradients)
}

func runTest021Gradients() (string, error) {
	return writeDoc("test_021_gradients.pdf", func(doc *pdf.DocWriter) error {
		doc.NewPage()
		doc.SetUnits("in")

		// Two-stop linear gradient fill.
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

		// Revert to solid fill.
		doc.ClearFillGradient()
		doc.SetFillColor(colors.LightGray)
		doc.Rectangle(1, 5.5, 3, 1, false, true)

		return nil
	})
}
