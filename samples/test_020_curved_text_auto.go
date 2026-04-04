// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_020_curved_text_auto", "compare curved text auto direction and facing against explicit settings", runTest020CurvedTextAuto)
}

func runTest020CurvedTextAuto() (string, error) {
	return writeDoc("test_020_curved_text_auto.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		afm, err := afm_fonts.Default()
		if err != nil {
			return err
		}
		doc.AddFontSource(afm)

		if _, err := doc.SetFont("Courier", 16, options.Options{}); err != nil {
			return err
		}
		if _, err := doc.SetFont("Helvetica", 10, options.Options{}); err != nil {
			return err
		}

		doc.MoveTo(0.6, 0.45)
		fmt.Fprintln(doc, "Curved text auto policy demo")
		doc.MoveTo(0.6, 0.67)
		fmt.Fprintln(doc, "Blue is explicit. Red uses Auto for direction and facing.")

		type sample struct {
			title      string
			x, y, r    float64
			text       string
			startAngle float64
			explicit   pdf.CurvedTextOptions
			automatic  pdf.CurvedTextOptions
		}

		samples := []sample{
			{
				title: "Top span",
				x:     2.2, y: 2.8, r: 1.0,
				text:       "TOP SAMPLE",
				startAngle: 90,
				explicit: pdf.CurvedTextOptions{
					Align:       pdf.CurvedTextAlignCenter,
					Direction:   pdf.CurvedTextClockwise,
					Orientation: pdf.CurvedTextOrientationOutside,
					Facing:      pdf.CurvedTextFacingUpright,
				},
				automatic: pdf.CurvedTextOptions{
					Align: pdf.CurvedTextAlignCenter,
					// Direction:   pdf.CurvedTextDirectionAuto,
					Orientation: pdf.CurvedTextOrientationOutside,
					// Facing:      pdf.CurvedTextFacingAuto,
				},
			},
			{
				title: "Right span",
				x:     5.8, y: 2.8, r: 1.0,
				text:       "RIGHT SAMPLE",
				startAngle: 0,
				explicit: pdf.CurvedTextOptions{
					Align:       pdf.CurvedTextAlignCenter,
					Direction:   pdf.CurvedTextClockwise,
					Orientation: pdf.CurvedTextOrientationOutside,
					Facing:      pdf.CurvedTextFacingUpright,
				},
				automatic: pdf.CurvedTextOptions{
					Align: pdf.CurvedTextAlignCenter,
					// Direction:   pdf.CurvedTextDirectionAuto,
					Orientation: pdf.CurvedTextOrientationOutside,
					// Facing:      pdf.CurvedTextFacingAuto,
				},
			},
			{
				title: "Bottom span",
				x:     2.2, y: 6.2, r: 1.0,
				text:       "BOTTOM SAMPLE",
				startAngle: -90,
				explicit: pdf.CurvedTextOptions{
					Align:       pdf.CurvedTextAlignCenter,
					VAlign:      pdf.VTextAlignTop,
					Direction:   pdf.CurvedTextCounterClockwise,
					Orientation: pdf.CurvedTextOrientationOutside,
					Facing:      pdf.CurvedTextFacingUpsideDown,
				},
				automatic: pdf.CurvedTextOptions{
					Align:  pdf.CurvedTextAlignCenter,
					VAlign: pdf.VTextAlignTop,
					// Direction:   pdf.CurvedTextDirectionAuto,
					Orientation: pdf.CurvedTextOrientationOutside,
					// Facing:      pdf.CurvedTextFacingAuto,
				},
			},
			{
				title: "Left span",
				x:     5.8, y: 6.2, r: 1.0,
				text:       "LEFT SAMPLE",
				startAngle: 180,
				explicit: pdf.CurvedTextOptions{
					Align:       pdf.CurvedTextAlignCenter,
					Direction:   pdf.CurvedTextClockwise,
					Orientation: pdf.CurvedTextOrientationOutside,
					Facing:      pdf.CurvedTextFacingUpright,
				},
				automatic: pdf.CurvedTextOptions{
					Align: pdf.CurvedTextAlignCenter,
					// Direction:   pdf.CurvedTextDirectionAuto,
					Orientation: pdf.CurvedTextOrientationOutside,
					// Facing:      pdf.CurvedTextFacingAuto,
				},
			},
		}

		for _, sample := range samples {
			doc.SetFontColor(colors.Black)
			if _, err := doc.SetFont("Helvetica", 10, options.Options{}); err != nil {
				return err
			}
			doc.MoveTo(sample.x-0.45, sample.y-1.45)
			fmt.Fprintln(doc, sample.title)

			doc.SetLineColor(colors.LightGray)
			doc.SetLineWidth(2, "pt")
			if err = doc.Circle(sample.x, sample.y, sample.r, true, false, false); err != nil {
				return err
			}

			if _, err := doc.SetFont("Courier", 16, options.Options{}); err != nil {
				return err
			}

			doc.SetFontColor(colors.DarkBlue)
			if err := doc.DrawTextOnCircle(sample.text, sample.x, sample.y, sample.r, sample.startAngle, sample.explicit); err != nil {
				return err
			}

			doc.SetFontColor(colors.FireBrick)
			if err := doc.DrawTextOnCircle(sample.text, sample.x, sample.y, sample.r+0.18, sample.startAngle, sample.automatic); err != nil {
				return err
			}
		}
		return nil
	})
}
