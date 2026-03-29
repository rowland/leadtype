// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_016_tmp_svgs", "throwaway sample that prints every SVG file in tmp", runTest016TmpSVGs)
}

func runTest016TmpSVGs() (string, error) {
	return writeDoc("test_016_tmp_svgs.pdf", func(doc *pdf.DocWriter) error {
		paths, err := filepath.Glob("tmp/*.svg")
		if err != nil {
			return err
		}
		sort.Strings(paths)
		if len(paths) == 0 {
			return fmt.Errorf("no svg files found in tmp")
		}

		doc.SetUnits("in")
		doc.NewPage()

		afm, err := afm_fonts.Default()
		if err == nil {
			doc.AddFontSource(afm)
			_, _ = doc.SetFont("Helvetica", 10, options.Options{})
		}

		const (
			marginX     = 0.55
			marginY     = 0.75
			panelW      = 2.25
			panelH      = 1.7
			gapX        = 0.25
			gapY        = 0.3
			imageW      = 1.65
			imageH      = 0.9
			cols        = 3
			rowsPerPage = 5
		)

		drawHeader := func(page int) {
			doc.SetFontColor(colors.Black)
			doc.MoveTo(marginX, 0.45)
			fmt.Fprintf(doc, "tmp SVG sweep, page %d", page)
		}

		page := 1
		drawHeader(page)

		for i, path := range paths {
			panelIndex := i % (cols * rowsPerPage)
			if i > 0 && panelIndex == 0 {
				doc.NewPage()
				page++
				drawHeader(page)
			}

			col := panelIndex % cols
			row := panelIndex / cols
			x := marginX + (float64(col) * (panelW + gapX))
			y := marginY + (float64(row) * (panelH + gapY))

			doc.SetLineColor(colors.LightGray)
			doc.SetLineWidth(0.75, "pt")
			doc.Rectangle(x, y, panelW, panelH, true, false)

			doc.SetFontColor(colors.Black)
			doc.MoveTo(x+0.06, y+0.18)
			fmt.Fprint(doc, filepath.Base(path))

			doc.SetLineColor(colors.Gainsboro)
			doc.SetLineWidth(0.5, "pt")
			doc.Rectangle(x+0.25, y+0.45, imageW, imageH, true, false)

			intrinsicW, intrinsicH, err := doc.SVGDimensionsFromFile(path)
			if err != nil {
				doc.SetFontColor(colors.FireBrick)
				doc.MoveTo(x+0.08, y+1.5)
				fmt.Fprintf(doc, "size error: %v", err)
				continue
			}
			targetW := imageW
			targetH := imageH
			if intrinsicW > 0 && intrinsicH > 0 {
				scale := targetW / float64(intrinsicW)
				heightFromWidth := float64(intrinsicH) * scale
				if heightFromWidth > targetH {
					scale = targetH / float64(intrinsicH)
				}
				targetW = float64(intrinsicW) * scale
				targetH = float64(intrinsicH) * scale
			}
			drawX := x + 0.25 + ((imageW - targetW) / 2)
			drawY := y + 0.45 + ((imageH - targetH) / 2)

			if _, _, err := doc.PrintSVGFile(path, drawX, drawY, &targetW, &targetH); err != nil {
				doc.SetFontColor(colors.FireBrick)
				doc.MoveTo(x+0.08, y+1.5)
				fmt.Fprintf(doc, "render error: %v", err)
			}
		}

		return nil
	})
}
