// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_017_curved_text", "demonstrate circle text placement", runTest017CurvedText)
}

func runTest017CurvedText() (string, error) {
	return writeDoc("test_017_curved_text.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		afm, err := afm_fonts.Default()
		if err != nil {
			return err
		}
		doc.AddFontSource(afm)

		if _, err := doc.SetFont("Courier", 18, options.Options{}); err != nil {
			return err
		}

		doc.SetLineColor(colors.LightGray)
		doc.SetLineWidth(1, "pt")
		_ = doc.Circle(3.5, 3.9, 1.25, true, false, false)

		doc.SetFontColor(colors.DarkBlue)
		if err := doc.DrawTextOnCircle("LEADTYPE", 3.5, 3.9, 1.25, 90, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			Direction:   pdf.CurvedTextClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpright,
		}); err != nil {
			return err
		}

		doc.SetFontColor(colors.FireBrick)
		if err := doc.DrawTextOnCircle("CURVED TEXT", 3.5, 3.9, 1.25, -90, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			VAlign:      pdf.VTextAlignTop,
			Direction:   pdf.CurvedTextCounterClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpsideDown,
		}); err != nil {
			return err
		}

		doc.SetFontColor(colors.DarkGreen)
		if err := doc.DrawTextOnCircle("RIGHT SIDE", 3.5, 3.9, 1.25, 0, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			Direction:   pdf.CurvedTextClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpright,
		}); err != nil {
			return err
		}

		doc.SetFontColor(colors.Purple)
		if err := doc.DrawTextOnCircle("LEFT SIDE", 3.5, 3.9, 1.25, 180, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			Direction:   pdf.CurvedTextClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpright,
		}); err != nil {
			return err
		}

		return nil
	})
}
