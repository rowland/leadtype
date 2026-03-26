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

		doc.SetLineColor(colors.LightGray)
		doc.Rectangle(0.9, 1.3, 1.4, 0.8, true, false)
		doc.Rectangle(3.8, 1.3, 1.4, 0.8, true, false)

		doc.SetLineColor(colors.FireBrick)
		doc.SetFillColor(colors.MistyRose)
		if err := doc.Rotate(25, 1.6, 1.7, func() {
			doc.Rectangle(0.95, 1.45, 1.3, 0.5, true, true)
			doc.MoveTo(1.15, 1.7)
			fmt.Fprintln(doc, "Rotate")
		}); err != nil {
			return err
		}

		doc.SetLineColor(colors.DarkBlue)
		doc.SetFillColor(colors.LightCyan)
		if err := doc.Scale(4.5, 1.7, 1.35, 0.7, func() {
			doc.Rectangle(3.95, 1.45, 1.1, 0.5, true, true)
			doc.MoveTo(4.15, 1.7)
			fmt.Fprintln(doc, "Scale")
		}); err != nil {
			return err
		}

		doc.SetLineColor(colors.Purple)
		if err := doc.Rotate(-30, 2.0, 4.1, func() {
			doc.Line(1.2, 4.1, 0, 1.6)
			doc.Line(2.0, 3.3, 90, 1.6)
		}); err != nil {
			return err
		}

		doc.SetLineColor(colors.Teal)
		doc.SetFillColor(colors.PaleTurquoise)
		if err := doc.Scale(4.8, 4.0, 1.4, 1.4, func() {
			_ = doc.Star(4.8, 4.0, 0.45, 0.2, 5, true, true, false, 0)
		}); err != nil {
			return err
		}

		return nil
	})
}
