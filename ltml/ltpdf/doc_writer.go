// Copyright 2016, 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltpdf

import (
	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

type DocWriter struct {
	*pdf.DocWriter
}

func (dw *DocWriter) NewPage() {
	dw.DocWriter.NewPage()
}

func (dw *DocWriter) EnableTaggedPDF(value bool) {
	dw.DocWriter.EnableTaggedPDF(value)
}

func (dw *DocWriter) LayoutProbeWriter() any {
	probe := newDocWriterWithFontSources(dw.FontSources())
	probe.SetAssetFS(dw.AssetFS())
	if dw.TaggedPDFEnabled() {
		probe.EnableTaggedPDF(true)
	}
	return probe
}

func (dw *DocWriter) SetLineWidth(width float64) {
	dw.DocWriter.SetLineWidth(width, "pt")
}

// PaintImageFile is the LTML image-as-fill hook. It supports brush-specific
// options such as uniform opacity while preserving explicit LTML sizing.
func (dw *DocWriter) PaintImageFile(filename string, x, y, width, height, opacity float64) error {
	return dw.DocWriter.PaintImageFile(filename, x, y, width, height, opacity)
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
	dw, err := NewDocWriterWithFontDirs(nil)
	if err != nil {
		panic(err)
	}
	return dw
}

// NewDocWriterWithFontDirs creates a DocWriter that searches system fonts plus
// each directory in dirs. It returns an error if any entry in dirs is invalid.
func NewDocWriterWithFontDirs(dirs []string) (*DocWriter, error) {
	ttFonts, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		if err := ttFonts.AddDir(dir); err != nil {
			return nil, err
		}
	}
	afmFonts, err := afm_fonts.Default()
	if err != nil {
		return nil, err
	}

	return newDocWriterWithFontSources(font.FontSources{ttFonts, afmFonts}), nil
}

func newDocWriterWithFontSources(sources font.FontSources) *DocWriter {
	dw := pdf.NewDocWriter()
	for _, source := range sources {
		dw.AddFontSource(source)
	}
	return &DocWriter{dw}
}
