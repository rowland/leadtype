package pdf

import (
	"bytes"
)

type PageWriter struct {
	autoPath      bool
	dw            *DocWriter
	gw            *graphWriter
	inGraph       bool
	inMisc        bool
	inPath        bool
	inText        bool
	lastLineColor Color
	lastLineWidth float64
	lastLoc       location
	lineColor     Color
	lineWidth     float64
	loc           location
	mw            *miscWriter
	page          *page
	pageHeight    float64
	stream        bytes.Buffer
	units         *units
}

func newPageWriter(dw *DocWriter, options Options) *PageWriter {
	return new(PageWriter).init(dw, options)
}

func (pw *PageWriter) init(dw *DocWriter, options Options) *PageWriter {
	pw.dw = dw
	pw.units = UnitConversions[options.StringDefault("units", "pt")]
	ps := newPageStyle(options)
	pw.pageHeight = ps.pageSize.y2
	pw.page = newPage(pw.dw.nextSeq(), 0, pw.dw.catalog.pages)
	pw.page.setMediaBox(ps.pageSize)
	pw.page.setCropBox(ps.cropSize)
	pw.page.setRotate(ps.rotate)
	pw.page.setResources(pw.dw.resources)
	pw.dw.file.body.add(pw.page)
	pw.startMisc()
	return pw
}

func (pw *PageWriter) checkSetLineColor() {
	if pw.lineColor == pw.lastLineColor {
		return
	}
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pw.mw.setRgbColorStroke(pw.lineColor.RGB64())
	pw.lastLineColor = pw.lineColor
}

func (pw *PageWriter) checkSetLineWidth() {
	if pw.lineWidth == pw.lastLineWidth {
		return
	}
	pw.startGraph()
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pw.gw.setLineWidth(pw.lineWidth)
	pw.lastLineWidth = pw.lineWidth
}

func (pw *PageWriter) Close() {
	// end margins
	// end sub page
	pw.endText()
	// end graph
	pw.endMisc()
	// compress stream
	pdf_stream := newStream(pw.dw.nextSeq(), 0, pw.stream.Bytes())
	pw.dw.file.body.add(pdf_stream)
	// set annots
	pw.page.add(pdf_stream)
	pw.dw.catalog.pages.add(pw.page) // unless reusing page
	pw.stream.Reset()
}

func (pw *PageWriter) endMisc() {
	pw.mw = nil
	pw.inMisc = false
}

func (pw *PageWriter) endText() {

}

func (pw *PageWriter) LineColor() Color {
	return pw.lineColor
}

func (pw *PageWriter) LineWidth(units string) float64 {
	return unitsFromPts(units, pw.lineWidth)
}

func (pw *PageWriter) MoveTo(x, y float64) {
	xpts, ypts := pw.units.toPts(x), pw.units.toPts(y)
	pw.loc = pw.translate(xpts, ypts)
}

func (pw *PageWriter) PageHeight() float64 {
	return pw.units.fromPts(pw.pageHeight)
}

func (pw *PageWriter) SetLineColor(color Color) (prev Color) {
	prev = pw.lineColor
	pw.lineColor = color
	return
}

func (pw *PageWriter) SetLineWidth(width float64, units string) (prev float64) {
	prev = unitsFromPts(units, pw.lineWidth)
	pw.lineWidth = unitsToPts(units, width)
	return
}

func (pw *PageWriter) SetUnits(units string) {
	pw.units = UnitConversions[units]
}

func (pw *PageWriter) startGraph() {
	if pw.inGraph {
		return
	}
	if pw.inText {
		pw.endText()
	}
	pw.lastLoc = location{0, 0}
	pw.gw = newGraphWriter(&pw.stream)
	pw.inGraph = true
}

func (pw *PageWriter) startMisc() {
	if pw.inMisc {
		return
	}
	pw.mw = newMiscWriter(&pw.stream)
	pw.inMisc = true
}

func (pw *PageWriter) translate(x, y float64) location {
	return location{x, pw.pageHeight - y}
}

func (pw *PageWriter) Units() string {
	return pw.units.name
}
