// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
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
		ttFonts, err := ttf_fonts.NewFromSystemFonts()
		if err != nil {
			return err
		}
		doc.AddFontSource(ttFonts)

		cx, cy := 3.5, 3.9
		r1, r2 := 1.25, 0.75

		if _, err := doc.SetFont("Courier", 18, options.Options{}); err != nil {
			return err
		}

		doc.SetLineColor(colors.LightGray)
		doc.SetLineWidth(1, "pt")
		_ = doc.Circle(cx, cy, r1, true, false, false)

		doc.SetFontColor(colors.DarkBlue)
		if err := doc.DrawTextOnCircle("LEADTYPE", cx, cy, r1, 90, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			Direction:   pdf.CurvedTextClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpright,
		}); err != nil {
			return err
		}

		doc.SetFontColor(colors.FireBrick)
		if err := doc.DrawTextOnCircle("CURVED TEXT", cx, cy, r1, -90, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			VAlign:      pdf.VTextAlignTop,
			Direction:   pdf.CurvedTextCounterClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpsideDown,
		}); err != nil {
			return err
		}

		doc.SetFontColor(colors.DarkGreen)
		if err := doc.DrawTextOnCircle("RIGHT SIDE", cx, cy, r1, 0, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			Direction:   pdf.CurvedTextClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpright,
		}); err != nil {
			return err
		}

		doc.SetFontColor(colors.Purple)
		if err := doc.DrawTextOnCircle("LEFT SIDE", cx, cy, r1, 180, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			Direction:   pdf.CurvedTextClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpright,
		}); err != nil {
			return err
		}

		if _, err := doc.SetFont("Arial Unicode MS", 18, options.Options{}); err != nil {
			return err
		}

		doc.SetLineColor(colors.LightGray)
		doc.SetLineWidth(1, "pt")
		_ = doc.Circle(cx, cy, r2, true, false, false)

		doc.SetFontColor(colors.Black)
		if err := doc.DrawTextOnCircle("Fortune", cx, cy, r2, 90, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			Direction:   pdf.CurvedTextClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpright,
		}); err != nil {
			return err
		}

		doc.SetFontColor(colors.Black)
		if err := doc.DrawTextOnCircle("Favors the Prepared Mind", cx, cy, r2, 270, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			VAlign:      pdf.VTextAlignTop,
			Direction:   pdf.CurvedTextCounterClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpsideDown,
		}); err != nil {
			return err
		}

		doc.SetLineColor(colors.LightGray)
		doc.SetLineWidth(1, "pt")
		metrics, err := doc.MeasureText("X")
		if err != nil {
			return err
		}
		_ = doc.Circle(cx, cy, r2+metrics.Height, true, false, false)

		if _, err := doc.SetFont("Amiri", 18, options.Options{}); err != nil {
			return err
		}

		r3 := 2.0
		_ = doc.Circle(cx, cy, r3, true, false, false)
		doc.SetFontColor(colors.Black)
		// "Sample Arabic text"
		if err := doc.DrawTextOnCircle("نص عربي تجريبي", cx, cy, r3, 30, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			VAlign:      pdf.VTextAlignBelow,
			Direction:   pdf.CurvedTextClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpright,
		}); err != nil {
			return err
		}
		// "Example of Arabic text"
		if err := doc.DrawTextOnCircle("مثال على نص عربي", cx, cy, r3, 150, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			VAlign:      pdf.VTextAlignBelow,
			Direction:   pdf.CurvedTextClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpright,
		}); err != nil {
			return err
		}
		// "Welcome"
		if err := doc.DrawTextOnCircle("مرحبا بكم", cx, cy, r3, 270, pdf.CurvedTextOptions{
			Align:       pdf.CurvedTextAlignCenter,
			VAlign:      pdf.VTextAlignAbove,
			Direction:   pdf.CurvedTextCounterClockwise,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      pdf.CurvedTextFacingUpsideDown,
		}); err != nil {
			return err
		}
		return nil
	})
}
