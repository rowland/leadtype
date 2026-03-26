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
	registerSample("test_010_path_api", "demonstrate manual path fill, stroke, and clipping", runTest010PathAPI)
}

func runTest010PathAPI() (string, error) {
	return writeDoc("test_010_path_api.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err == nil && ttfc.Len() > 0 {
			doc.AddFontSource(ttfc)
			if _, err := doc.SetFont("Arial", 12, options.Options{}); err == nil {
				doc.MoveTo(0.75, 0.6)
				fmt.Fprintln(doc, "Path API demo: fill, stroke, and clipping")
			}
		}

		doc.SetLineWidth(2, "pt")
		doc.SetLineColor(colors.DarkBlue)
		doc.SetFillColor(colors.LightBlue)

		// Filled triangle on the left.
		if err := doc.Path(func() {
			doc.MoveTo(0.9, 1.6)
			doc.LineTo(2.2, 3.8)
			doc.LineTo(0.9, 3.8)
			doc.LineTo(0.9, 1.6)
			if err := doc.Fill(); err != nil {
				panic(err)
			}
		}); err != nil {
			return err
		}

		// Stroked diamond in the middle.
		doc.SetLineColor(colors.FireBrick)
		if err := doc.Path(func() {
			doc.MoveTo(3.2, 1.7)
			doc.LineTo(4.0, 2.7)
			doc.LineTo(3.2, 3.7)
			doc.LineTo(2.4, 2.7)
			doc.LineTo(3.2, 1.7)
			if err := doc.Stroke(); err != nil {
				panic(err)
			}
		}); err != nil {
			return err
		}

		// Clipped starburst stripes on the right.
		doc.SetLineColor(colors.DarkGreen)
		doc.SetFillColor(colors.LightYellow)
		if err := doc.Path(func() {
			doc.MoveTo(5.0, 1.6)
			doc.LineTo(5.9, 2.2)
			doc.LineTo(5.6, 3.4)
			doc.LineTo(4.4, 3.4)
			doc.LineTo(4.1, 2.2)
			doc.LineTo(5.0, 1.6)
			if err := doc.Clip(func() {
				doc.SetLineWidth(4, "pt")
				doc.SetLineColor(colors.Gold)
				for i := 0; i < 8; i++ {
					y := 1.7 + float64(i)*0.25
					doc.MoveTo(4.1, y)
					doc.LineTo(5.9, y+0.8)
				}
				doc.SetLineWidth(2, "pt")
				doc.SetLineColor(colors.DarkGreen)
				doc.Rectangle(4.1, 1.6, 1.8, 1.8, true, false)
			}); err != nil {
				panic(err)
			}
		}); err != nil {
			return err
		}

		return nil
	})
}
