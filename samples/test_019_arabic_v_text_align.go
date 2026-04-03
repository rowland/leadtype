// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/shaping"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_019_arabic_v_text_align", "compare Arabic vertical text alignment modes on straight baselines", runTest019ArabicVTextAlign)
}

func runTest019ArabicVTextAlign() (string, error) {
	return writeDoc("test_019_arabic_v_text_align.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		afm, err := afm_fonts.Default()
		if err != nil {
			return err
		}
		doc.AddFontSource(afm)

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err != nil {
			return err
		}
		ttfc.SetShaper(shaping.NewShaper())
		doc.AddFontSource(ttfc)

		const specimen = "مثال عربي"
		alignments := []struct {
			label string
			color colors.Color
		}{
			{label: "above", color: colors.FireBrick},
			{label: "top", color: colors.DarkBlue},
			{label: "middle", color: colors.DarkGreen},
			{label: "base", color: colors.Purple},
			{label: "below", color: colors.OrangeRed},
		}

		fonts := []struct {
			family string
			style  string
			label  string
		}{
			{family: "Amiri", style: "Regular", label: "Amiri Regular"},
			{family: "Arial", style: "Bold", label: "Arial Bold"},
			{family: "Courier New", style: "Bold", label: "Courier New Bold"},
			{family: "Damascus", style: "Regular", label: "Damascus Regular"},
			{family: "Times New Roman", style: "Bold", label: "Times New Roman Bold"},
		}

		if _, err := doc.SetFont("Helvetica", 11, options.Options{}); err != nil {
			return err
		}
		doc.MoveTo(0.55, 0.45)
		fmt.Fprintln(doc, "Arabic vertical alignment comparison")
		doc.MoveTo(0.55, 0.68)
		fmt.Fprintln(doc, "Guide line is the shared baseline reference for each row.")

		xs := []float64{1.8, 3.15, 4.5, 5.85, 7.2}
		for i, align := range alignments {
			doc.SetFontColor(align.color)
			doc.SetVTextAlign("base")
			doc.MoveTo(xs[i], 1.05)
			fmt.Fprintln(doc, align.label)
		}

		doc.SetLineWidth(1, "pt")
		y := 2.0
		for _, face := range fonts {
			doc.SetVTextAlign("base")
			doc.SetLineColor(colors.LightGray)
			doc.Line(0.75, y, 0, 7.1)

			doc.SetFontColor(colors.Black)
			if _, err := doc.SetFont("Helvetica", 10, options.Options{}); err != nil {
				return err
			}
			doc.MoveTo(0.15, y)
			fmt.Fprintln(doc, face.label)

			for i, align := range alignments {
				doc.SetFontColor(align.color)
				doc.SetVTextAlign(align.label)
				fonts, err := doc.SetFont(face.family, 12, options.Options{"style": face.style})
				if err != nil {
					return err
				}
				if !canRender(specimen, fonts...) {
					doc.SetVTextAlign("base")
					if _, err := doc.SetFont("Helvetica", 8, options.Options{}); err != nil {
						return err
					}
					doc.MoveTo(xs[i], y)
					fmt.Fprintln(doc, "missing glyphs")
					continue
				}
				doc.MoveTo(xs[i], y)
				fmt.Fprintln(doc, specimen)
			}

			y += 1.55
		}
		doc.SetVTextAlign("base")
		return nil
	})
}
