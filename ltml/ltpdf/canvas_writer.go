// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltpdf

import "github.com/rowland/leadtype/pdf"

type canvasWriter struct {
	*pdf.PageWriter
	docWriter *pdf.DocWriter
}

func (w *canvasWriter) EnableTaggedPDF(bool) {}

func (w *canvasWriter) NewPage() {}

func (w *canvasWriter) TaggedPDFEnabled() bool {
	return false
}

func (w *canvasWriter) ImageDimensions(data []byte) (width, height int, err error) {
	return w.docWriter.ImageDimensions(data)
}

func (w *canvasWriter) SVGDimensions(data []byte) (width, height int, err error) {
	return w.docWriter.SVGDimensions(data)
}

func (w *canvasWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
	return w.docWriter.ImageDimensionsFromFile(filename)
}

func (w *canvasWriter) SVGDimensionsFromFile(filename string) (width, height int, err error) {
	return w.docWriter.SVGDimensionsFromFile(filename)
}

func (w *canvasWriter) SetLineCapStyle(style string) (prev string) {
	prevStyle := w.PageWriter.LineCapStyle()
	switch style {
	case "round_cap":
		w.PageWriter.SetLineCapStyle(pdf.RoundCap)
	case "projecting_square_cap":
		w.PageWriter.SetLineCapStyle(pdf.ProjectingSquareCap)
	default:
		w.PageWriter.SetLineCapStyle(pdf.ButtCap)
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

func (w *canvasWriter) SetLineWidth(width float64) {
	w.PageWriter.SetLineWidth(width, "pt")
}

func (w *canvasWriter) DrawCanvas(key string, x, y, width, height, canvasWidth, canvasHeight float64, draw func(any) error) error {
	if draw == nil {
		return nil
	}
	return w.PageWriter.MemoizeFormOnCanvas(canvasMemoKey(key), x, y, width, height, canvasWidth, canvasHeight, func(pw *pdf.PageWriter) error {
		return draw(&canvasWriter{PageWriter: pw, docWriter: w.docWriter})
	})
}

func canvasMemoKey(key string) string {
	return "ltml-canvas:" + key
}
