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
	registerSample("test_013_v_text_align", "demonstrate vertical text alignment modes", runTest013VTextAlign)
}

func runTest013VTextAlign() (string, error) {
	return writeDoc("test_013_v_text_align.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err == nil && ttfc.Len() > 0 {
			doc.AddFontSource(ttfc)
			if _, err := doc.SetFont("Arial", 12, options.Options{}); err == nil {
				doc.MoveTo(0.7, 0.6)
				fmt.Fprintln(doc, "Vertical text alignment demo")
			}
		}

		doc.ResetFonts()
		doc.SetLineWidth(1, "pt")
		doc.SetLineColor(colors.LightGray)
		doc.Line(0.8, 1.6, 0, 5.6)
		doc.Line(0.8, 2.8, 0, 5.6)
		doc.Line(0.8, 4.0, 0, 5.6)
		doc.Line(0.8, 5.2, 0, 5.6)
		doc.Line(0.8, 6.4, 0, 5.6)

		doc.SetLineColor(colors.Black)
		doc.SetFont("Arial", 24, options.Options{})

		rows := []struct {
			y     float64
			align string
			color colors.Color
		}{
			{1.6, "above", colors.FireBrick},
			{2.8, "top", colors.DarkBlue},
			{4.0, "middle", colors.DarkGreen},
			{5.2, "base", colors.Purple},
			{6.4, "below", colors.OrangeRed},
		}

		for _, row := range rows {
			doc.SetFontColor(row.color)
			doc.SetVTextAlign(row.align)
			doc.MoveTo(1.1, row.y)
			fmt.Fprintf(doc, "%s  AgjpQy", row.align)
		}

		return nil
	})
}
