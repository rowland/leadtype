// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"math"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_027_form_memo_canvas", "reuse one logical-canvas form at multiple placement sizes", runTest027FormMemoCanvas)
}

type targetBoardMode uint8

const (
	targetBoardDetailed targetBoardMode = iota
	targetBoardSimplified
)

func runTest027FormMemoCanvas() (string, error) {
	return writeDoc("test_027_form_memo_canvas.pdf", func(doc *pdf.DocWriter) error {
		doc.SetUnits("pt")
		if fonts, err := afm_fonts.Default(); err == nil {
			doc.AddFontSource(fonts)
		}

		page := doc.NewPage()
		if _, err := page.SetFont("Helvetica", 18, options.Options{"weight": "bold"}); err == nil {
			page.MoveTo(48, 44)
			_ = page.Print("Memoized form with a logical canvas")
		}
		if _, err := page.SetFont("Helvetica", 10, options.Options{}); err == nil {
			page.MoveTo(48, 66)
			_ = page.Print("The detailed board below is drawn directly. The smaller boards reuse a simplified form captured once on a 240pt canvas.")
			page.MoveTo(48, 84)
			_ = page.Print("That simplified version is never placed at full size, but it keeps a consistent coordinate system for every reuse.")
		}

		if err := drawTargetBoard(page, 164, 250, 108, targetBoardDetailed); err != nil {
			return err
		}
		if _, err := page.SetFont("Helvetica", 9, options.Options{"weight": "bold"}); err == nil {
			page.MoveTo(82, 382)
			_ = page.Print("Detailed geometry drawn directly")
		}

		placements := []struct {
			x     float64
			y     float64
			size  float64
			label string
		}{
			{x: 340, y: 154, size: 96, label: "96pt placement"},
			{x: 468, y: 172, size: 64, label: "64pt placement"},
			{x: 340, y: 292, size: 48, label: "48pt placement"},
			{x: 428, y: 280, size: 34, label: "34pt placement"},
			{x: 492, y: 296, size: 24, label: "24pt placement"},
		}
		for _, placement := range placements {
			if err := page.MemoizeFormOnCanvas(
				"target-board:simplified",
				placement.x,
				placement.y,
				placement.size,
				placement.size,
				240,
				240,
				func(form *pdf.PageWriter) error {
					return drawTargetBoard(form, 120, 120, 108, targetBoardSimplified)
				},
			); err != nil {
				return err
			}
			if _, err := page.SetFont("Helvetica", 8, options.Options{}); err == nil {
				page.MoveTo(placement.x, placement.y+placement.size+14)
				_ = page.Print(placement.label)
			}
		}

		page.SetLineColor(colors.Gainsboro)
		page.SetLineWidth(1, "pt")
		page.Line(300, 120, 90, 270)

		return nil
	})
}

func drawTargetBoard(pw *pdf.PageWriter, cx, cy, radius float64, mode targetBoardMode) error {
	pw.SetLineColor(colors.DarkSlateBlue)
	pw.SetLineWidth(1.25, "pt")
	pw.SetFillColor(colors.WhiteSmoke)
	if err := pw.Circle(cx, cy, radius, true, true, false); err != nil {
		return err
	}

	if err := fillTargetBoardBands(pw, cx, cy, radius, mode); err != nil {
		return err
	}

	pw.SetFillColor(colors.Gold)
	if err := pw.Circle(cx, cy, radius*0.2, true, true, false); err != nil {
		return err
	}
	pw.SetFillColor(colors.FireBrick)
	if err := pw.Circle(cx, cy, radius*0.09, true, true, false); err != nil {
		return err
	}

	for _, ring := range []float64{1.0, 0.82, 0.58, 0.34, 0.2, 0.09} {
		if err := pw.Circle(cx, cy, radius*ring, true, false, false); err != nil {
			return err
		}
	}

	spokes := []float64{0, 90, 180, 270}
	if mode == targetBoardDetailed {
		spokes = []float64{0, 22.5, 45, 67.5, 90, 112.5, 135, 157.5, 180, 202.5, 225, 247.5, 270, 292.5, 315, 337.5}
	}
	for _, angle := range spokes {
		x1, y1 := pointOnCircle(cx, cy, radius*0.2, angle)
		x2, y2 := pointOnCircle(cx, cy, radius, angle)
		pw.MoveTo(x1, y1)
		pw.LineTo(x2, y2)
	}

	return nil
}

func fillTargetBoardBands(pw *pdf.PageWriter, cx, cy, radius float64, mode targetBoardMode) error {
	if mode == targetBoardSimplified {
		palette := []colors.Color{colors.LightSteelBlue, colors.PeachPuff, colors.LightCyan, colors.LightGoldenRodYellow}
		for i := 0; i < 4; i++ {
			pw.SetFillColor(palette[i%len(palette)])
			if err := pw.Pie(cx, cy, radius*0.82, float64(i)*90, float64(i+1)*90, false, true, false); err != nil {
				return err
			}
		}
		pw.SetFillColor(colors.WhiteSmoke)
		return pw.Circle(cx, cy, radius*0.34, false, true, false)
	}

	palette := []colors.Color{colors.LightSteelBlue, colors.LightGoldenRodYellow, colors.MistyRose, colors.HoneyDew}
	for i := 0; i < 16; i++ {
		pw.SetFillColor(palette[i%len(palette)])
		if err := pw.Arch(cx, cy, radius*0.58, radius*0.82, float64(i)*22.5, float64(i+1)*22.5, false, true, false); err != nil {
			return err
		}
	}
	for i := 0; i < 8; i++ {
		pw.SetFillColor(palette[(i+1)%len(palette)])
		if err := pw.Arch(cx, cy, radius*0.2, radius*0.34, float64(i)*45, float64(i+1)*45, false, true, false); err != nil {
			return err
		}
	}
	return nil
}

func pointOnCircle(cx, cy, radius, angle float64) (x, y float64) {
	radians := angle * math.Pi / 180.0
	return cx + (radius * math.Cos(radians)), cy - (radius * math.Sin(radians))
}
