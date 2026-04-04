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
		const (
			discRed    = colors.Color(0xE06556)
			discYellow = colors.Color(0xEDD468)
			discGreen  = colors.Color(0x8DB967)
			discBlue   = colors.Color(0x608FAA)
		)

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

		// Pseudo-sweep annular band built from explicit solid and gradient slices.
		label(4.7, 7.35, "Pseudo-sweep annular band")
		if err := doc.PaintSweepBand(&pdf.SweepBand{
			X:           6.25,
			Y:           8.5,
			InnerRadius: 0.58,
			OuterRadius: 1.02,
			Segments: segmentsForCircle(22.5, []colorSteps{
				{startColor: discRed, endColor: discRed, steps: 0},
				{startColor: discRed, endColor: discBlue, steps: 12},
				{startColor: discBlue, endColor: discBlue, steps: 0},
				{startColor: discBlue, endColor: discGreen, steps: 12},
				{startColor: discGreen, endColor: discGreen, steps: 0},
				{startColor: discGreen, endColor: discYellow, steps: 12},
				{startColor: discYellow, endColor: discYellow, steps: 0},
				{startColor: discYellow, endColor: discRed, steps: 12},
			}),
		}); err != nil {
			return err
		}
		doc.SetLineColor(colors.Gray)
		doc.SetLineWidth(2, "pt")
		if err := doc.Circle(6.25, 8.5, 1.02, true, false, false); err != nil {
			return err
		}
		if err := doc.Circle(6.25, 8.5, 0.58, true, false, false); err != nil {
			return err
		}

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

type colorSteps struct {
	startColor colors.Color
	endColor   colors.Color
	steps      int
}

func segmentsForCircle(startAngle float64, colorSteps []colorSteps) []pdf.SweepBandSegment {
	angleStep := 360.0 / float64(len(colorSteps))
	segments := make([]pdf.SweepBandSegment, 0, len(colorSteps))
	for i := range colorSteps {
		startAngle := startAngle + (angleStep * float64(i))
		endAngle := startAngle + angleStep
		segments = append(segments, pdf.SweepBandSegment{
			StartAngle: startAngle,
			EndAngle:   endAngle,
			StartColor: colorSteps[i].startColor,
			EndColor:   colorSteps[i].endColor,
			Steps:      colorSteps[i].steps,
		})
	}
	return segments
}
