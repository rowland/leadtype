// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"sort"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_021_named_color_swatches", "render a paginated named-color reference with labeled swatches", runTest021NamedColorSwatches)
}

func runTest021NamedColorSwatches() (string, error) {
	return writeDoc("test_021_named_color_swatches.pdf", func(doc *pdf.DocWriter) error {
		swatches := sortedNamedSwatches()

		doc.SetUnits("in")
		doc.NewPage()

		afm, err := afm_fonts.Default()
		if err != nil {
			return err
		}
		doc.AddFontSource(afm)
		if _, err := doc.SetFont("Helvetica", 10, options.Options{}); err != nil {
			return err
		}

		const (
			marginX     = 0.55
			startY      = 0.95
			cellW       = 1.82
			cellH       = 1.88
			gapX        = 0.14
			gapY        = 0.14
			cols        = 4
			rowsPerPage = 5
			swatchR     = 0.5
		)

		perPage := cols * rowsPerPage
		totalPages := (len(swatches) + perPage - 1) / perPage

		centerText := func(text string, centerX, y float64) error {
			metrics, err := doc.MeasureText(text)
			if err != nil {
				return err
			}
			doc.MoveTo(centerX-(metrics.Width/2), y)
			fmt.Fprint(doc, text)
			return nil
		}

		drawPageChrome := func(page int) error {
			doc.SetFont("Helvetica", 14, options.Options{})
			doc.SetFontColor(colors.Black)
			doc.MoveTo(marginX, 0.42)
			fmt.Fprint(doc, "Named color reference")

			doc.SetFont("Helvetica", 9, options.Options{})
			doc.SetFontColor(colors.DimGray)
			doc.MoveTo(marginX, 0.64)
			fmt.Fprintf(doc, "%d accepted names from colors.NamedColors", len(swatches))

			doc.MoveTo(7.05, 0.42)
			fmt.Fprintf(doc, "page %d/%d", page, totalPages)

			doc.SetLineColor(colors.Gainsboro)
			doc.SetLineWidth(0.6, "pt")
			doc.Line(marginX, 0.74, 7.35, 0)

			return nil
		}

		for page := 0; page < totalPages; page++ {
			if page > 0 {
				doc.NewPage()
			}
			if err := drawPageChrome(page + 1); err != nil {
				return err
			}

			start := page * perPage
			end := start + perPage
			if end > len(swatches) {
				end = len(swatches)
			}

			for i, swatch := range swatches[start:end] {
				panelIndex := i
				col := panelIndex % cols
				row := panelIndex / cols

				x := marginX + (float64(col) * (cellW + gapX))
				y := startY + (float64(row) * (cellH + gapY))

				doc.SetLineColor(colors.Gainsboro)
				doc.SetLineWidth(0.5, "pt")
				doc.Rectangle(x, y, cellW, cellH, true, false)

				centerX := x + (cellW / 2)

				doc.SetFont("Helvetica", 8.5, options.Options{})
				nameMetrics, err := doc.MeasureText(swatch.Name)
				if err != nil {
					return err
				}

				hexLabel := fmt.Sprintf("#%06X", int32(swatch.Color))
				doc.SetFont("Helvetica", 8, options.Options{})
				hexMetrics, err := doc.MeasureText(hexLabel)
				if err != nil {
					return err
				}

				stackGap := (cellH - (nameMetrics.Height + (2 * swatchR) + hexMetrics.Height)) / 4
				nameBaselineY := y + stackGap + nameMetrics.Ascent
				centerY := y + stackGap + nameMetrics.Height + stackGap + swatchR
				hexBaselineY := y + stackGap + nameMetrics.Height + stackGap + (2 * swatchR) + stackGap + hexMetrics.Ascent

				doc.SetLineColor(colors.DarkSlateGray)
				doc.SetLineWidth(0.45, "pt")
				doc.SetFillColor(swatch.Color)
				if err := doc.Circle(centerX, centerY, swatchR, true, true, false); err != nil {
					return err
				}

				doc.SetFont("Helvetica", 8.5, options.Options{})
				doc.SetFontColor(colors.Black)
				if err := centerText(swatch.Name, centerX, nameBaselineY); err != nil {
					return err
				}

				doc.SetFont("Helvetica", 8, options.Options{})
				doc.SetFontColor(colors.DimGray)
				if err := centerText(hexLabel, centerX, hexBaselineY); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func sortedNamedSwatches() []colors.NamedColorEntry {
	swatches := append([]colors.NamedColorEntry(nil), colors.NamedColorEntries...)
	sort.Slice(swatches, func(i, j int) bool { return swatches[i].Name < swatches[j].Name })
	return swatches
}
