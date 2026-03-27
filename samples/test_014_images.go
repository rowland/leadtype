// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_014_images", "demonstrate JPEG and PNG image placement from file and memory", runTest014Images)
}

func drawPanel(doc *pdf.DocWriter, x, y, width, height float64, title string) {
	doc.SetLineColor(colors.LightGray)
	doc.SetLineWidth(1, "pt")
	doc.Rectangle(x, y, width, height, true, false)
	doc.SetFontColor(colors.Black)
	doc.MoveTo(x, y-0.12)
	fmt.Fprintln(doc, title)
}

func drawFootprint(doc *pdf.DocWriter, x, y, width, height float64, color colors.Color, label string) {
	doc.SetLineColor(color)
	doc.SetLineWidth(1, "pt")
	doc.Rectangle(x, y, width, height, true, false)
	doc.SetFontColor(color)
	doc.MoveTo(x, y-0.08)
	fmt.Fprintln(doc, label)
}

func runTest014Images() (string, error) {
	return writeDoc("test_014_images.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("in")
		doc.NewPage()

		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err == nil && ttfc.Len() > 0 {
			doc.AddFontSource(ttfc)
			if _, err := doc.SetFont("Arial", 12, options.Options{}); err == nil {
				doc.MoveTo(0.7, 0.6)
				fmt.Fprintln(doc, "JPEG and PNG image placement demo")
			}
		}

		jpegPath := "pdf/testdata/testimg.jpg"
		jpegData, err := os.ReadFile(jpegPath)
		if err != nil {
			return err
		}
		jpegIntrinsicW, jpegIntrinsicH, err := doc.ImageDimensions(jpegData)
		if err != nil {
			return err
		}

		pngPath := "pdf/testdata/eidetic.png"
		pngData, err := os.ReadFile(pngPath)
		if err != nil {
			return err
		}
		pngIntrinsicW, pngIntrinsicH, err := doc.ImageDimensions(pngData)
		if err != nil {
			return err
		}

		doc.SetFont("Arial", 10, options.Options{})

		drawPanel(doc, 0.7, 1.4, 3.0, 2.4, "JPEG width fixed to 1.8 in")
		widthOnly := 1.8
		if _, _, err := doc.PrintImageFile(jpegPath, 1.1, 1.85, &widthOnly, nil); err != nil {
			return err
		}
		drawFootprint(doc, 1.1, 1.85, 1.8, 1.8*float64(jpegIntrinsicH)/float64(jpegIntrinsicW), colors.SteelBlue, "height inferred")

		drawPanel(doc, 4.0, 1.4, 3.0, 2.4, "JPEG height fixed to 1.4 in")
		heightOnly := 1.4
		widthFromHeight := 1.4 * float64(jpegIntrinsicW) / float64(jpegIntrinsicH)
		if _, _, err := doc.PrintImage(jpegData, 4.35, 1.85, nil, &heightOnly); err != nil {
			return err
		}
		drawFootprint(doc, 4.35, 1.85, widthFromHeight, 1.4, colors.SeaGreen, "width inferred")

		jpegIntrinsicWidth := float64(jpegIntrinsicW) / 72.0
		jpegIntrinsicHeight := float64(jpegIntrinsicH) / 72.0

		drawPanel(doc, 0.7, 4.3, 3.2, 3.0, "Native size from JPEG metadata")
		if _, _, err := doc.PrintImageFile(jpegPath, 1.0, 4.8, nil, nil); err != nil {
			return err
		}
		drawFootprint(doc, 1.0, 4.8, jpegIntrinsicWidth, jpegIntrinsicHeight, colors.OrangeRed, "227 pt x 149 pt")

		drawPanel(doc, 4.2, 4.3, 2.8, 3.0, "JPEG forced to 1.7 x 1.1 in")
		explicitJPEGWidth, explicitJPEGHeight := 1.7, 1.1
		if _, _, err := doc.PrintImage(jpegData, 4.55, 4.95, &explicitJPEGWidth, &explicitJPEGHeight); err != nil {
			return err
		}
		drawFootprint(doc, 4.55, 4.95, explicitJPEGWidth, explicitJPEGHeight, colors.Purple, "aspect ratio ignored")

		drawPanel(doc, 0.7, 7.45, 3.2, 2.45, "PNG width fixed to 2.4 in")
		pngWidth := 2.4
		pngHeight := pngWidth * float64(pngIntrinsicH) / float64(pngIntrinsicW)
		if _, _, err := doc.PrintImageFile(pngPath, 1.0, 7.85, &pngWidth, nil); err != nil {
			return err
		}
		drawFootprint(doc, 1.0, 7.85, pngWidth, pngHeight, colors.MediumVioletRed, "height inferred")

		drawPanel(doc, 4.2, 7.45, 2.8, 2.45, "PNG forced to 1.7 x 1.1 in")
		explicitWidth, explicitHeight := 1.7, 1.1
		if _, _, err := doc.PrintImage(pngData, 4.55, 7.95, &explicitWidth, &explicitHeight); err != nil {
			return err
		}
		drawFootprint(doc, 4.55, 7.95, explicitWidth, explicitHeight, colors.DarkCyan, "aspect ratio ignored")

		return nil
	})
}
