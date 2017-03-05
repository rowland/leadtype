// Copyright 2016, 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package harupdf

import (
	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/haru"
	"github.com/rowland/leadtype/ttf_fonts"
)

type DocWriter struct {
	*haru.DocWriter
}

func (dw *DocWriter) NewPage() {
	dw.DocWriter.NewPage()
}

func (dw *DocWriter) SetLineWidth(width float64) {
	dw.DocWriter.SetLineWidth(width, "pt")
}

func NewDocWriter() *DocWriter {
	dw := haru.NewDocWriter()
	ttFonts, err := ttf_fonts.New("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	dw.AddFontSource(ttFonts)

	afmFonts, err := afm_fonts.New("../afm/data/fonts/*.afm")
	if err != nil {
		panic(err)
	}
	dw.AddFontSource(afmFonts)

	return &DocWriter{dw}
}
