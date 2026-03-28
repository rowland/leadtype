// Copyright 2016, 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltpdf

import (
	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

type DocWriter struct {
	*pdf.DocWriter
}

func (dw *DocWriter) NewPage() {
	dw.DocWriter.NewPage()
}

func (dw *DocWriter) LayoutProbeWriter() interface{} {
	probe := NewDocWriter()
	probe.SetAssetFS(dw.AssetFS())
	return probe
}

func (dw *DocWriter) SetLineWidth(width float64) {
	dw.DocWriter.SetLineWidth(width, "pt")
}

func (dw *DocWriter) SetLineCapStyle(style string) (prev string) {
	prevStyle := dw.DocWriter.CurPage().LineCapStyle()
	switch style {
	case "round_cap":
		dw.DocWriter.CurPage().SetLineCapStyle(pdf.RoundCap)
	case "projecting_square_cap":
		dw.DocWriter.CurPage().SetLineCapStyle(pdf.ProjectingSquareCap)
	default:
		dw.DocWriter.CurPage().SetLineCapStyle(pdf.ButtCap)
	}
	switch prevStyle {
	case pdf.RoundCap:
		return "round_cap"
	case pdf.ProjectingSquareCap:
		return "projecting_square_cap"
	default:
		return "butt_cap"
	}
}

func NewDocWriter() *DocWriter {
	dw := pdf.NewDocWriter()
	ttFonts, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		panic(err)
	}
	dw.AddFontSource(ttFonts)

	afmFonts, err := afm_fonts.Default()
	if err != nil {
		panic(err)
	}
	dw.AddFontSource(afmFonts)

	return &DocWriter{dw}
}
