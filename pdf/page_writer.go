// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"fmt"
)

type LineCapStyle int

const (
	ButtCap             = LineCapStyle(iota)
	RoundCap            = LineCapStyle(iota)
	ProjectingSquareCap = LineCapStyle(iota)
)

type PageWriter struct {
	drawState
	autoPath   bool
	dw         *DocWriter
	fonts      []*Font
	gw         *graphWriter
	inGraph    bool
	inMisc     bool
	inPath     bool
	inText     bool
	isClosed   bool
	last       drawState
	mw         *miscWriter
	options    Options
	page       *page
	pageHeight float64
	stream     bytes.Buffer
	units      *units
}

func newPageWriter(dw *DocWriter, options Options) *PageWriter {
	return new(PageWriter).init(dw, options)
}

func clonePageWriter(opw *PageWriter) *PageWriter {
	pw := new(PageWriter).init(opw.dw, opw.options)
	pw.drawState = opw.drawState
	pw.units = opw.units
	return pw
}

func (pw *PageWriter) init(dw *DocWriter, options Options) *PageWriter {
	pw.dw = dw
	pw.options = options
	pw.units = UnitConversions[options.StringDefault("units", "pt")]
	ps := newPageStyle(options)
	pw.pageHeight = ps.pageSize.y2
	pw.page = newPage(pw.dw.nextSeq(), 0, pw.dw.catalog.pages)
	pw.page.setMediaBox(ps.pageSize)
	pw.page.setCropBox(ps.cropSize)
	pw.page.setRotate(ps.rotate)
	pw.page.setResources(pw.dw.resources)
	pw.dw.file.body.add(pw.page)
	pw.autoPath = true
	pw.startMisc()
	return pw
}

func (pw *PageWriter) AddFont(family string, size float64, options Options) ([]*Font, error) {
	f := &Font{
		family:  family,
		size:    size,
		weight:  options.StringDefault("weight", ""),
		style:   options.StringDefault("style", ""),
		color:   options.ColorDefault("color", Black),
		subType: options.StringDefault("sub_type", "TrueType"),
	}
	var fontSource FontSource
	var ok bool
	if fontSource, ok = pw.dw.fontSources[f.subType]; !ok {
		return nil, fmt.Errorf("Font subtype %s not found", f.subType)
	}
	if ranges, ok := options["ranges"]; ok {
		switch ranges := ranges.(type) {
		case []string:
			f.ranges = ranges
		case RuneSet:
			f.runeSet = ranges
		}
	}
	var err error
	if f.metrics, err = fontSource.Select(f.family, f.weight, f.style, f.ranges); err != nil {
		return nil, err
	}
	pw.fonts = append(pw.fonts, f)
	return pw.fonts, nil
}

func (pw *PageWriter) checkSetLineColor() {
	if pw.lineColor == pw.last.lineColor {
		return
	}
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pw.mw.setRgbColorStroke(pw.lineColor.RGB64())
	pw.last.lineColor = pw.lineColor
}

func (pw *PageWriter) checkSetLineDashPattern() {
	if pw.lineDashPattern == pw.last.lineDashPattern && pw.lineCapStyle == pw.last.lineCapStyle {
		return
	}
	pw.startGraph()
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pat := LinePatterns[pw.lineDashPattern]
	if pat == nil {
		pw.gw.setLineDashPattern(pw.lineDashPattern)
	} else {
		pw.gw.setLineDashPattern(pat.String())
	}
	pw.last.lineDashPattern = pw.lineDashPattern
	pw.gw.setLineCapStyle(int(pw.lineCapStyle))
	pw.last.lineCapStyle = pw.lineCapStyle
}

func (pw *PageWriter) checkSetLineWidth() {
	if pw.lineWidth == pw.last.lineWidth {
		return
	}
	pw.startGraph()
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pw.gw.setLineWidth(pw.lineWidth)
	pw.last.lineWidth = pw.lineWidth
}

func (pw *PageWriter) Close() {
	if pw.isClosed {
		return
	}
	// end margins
	// end sub page
	pw.endText()
	pw.endGraph()
	pw.endMisc()
	// compress stream
	pdf_stream := newStream(pw.dw.nextSeq(), 0, pw.stream.Bytes())
	pw.dw.file.body.add(pdf_stream)
	// set annots
	pw.page.add(pdf_stream)
	pw.dw.catalog.pages.add(pw.page) // unless reusing page
	pw.stream.Reset()
	pw.isClosed = true
}

func (pw *PageWriter) endGraph() {
	if pw.inPath {
		pw.endPath()
	}
	pw.gw = nil
	pw.inGraph = false
}

func (pw *PageWriter) endMisc() {
	pw.mw = nil
	pw.inMisc = false
}

func (pw *PageWriter) endPath() {
	if pw.autoPath {
		pw.gw.stroke()
	}
	pw.inPath = false
}

func (pw *PageWriter) endText() {

}

func (pw *PageWriter) Fonts() []*Font {
	return pw.fonts
}

func (pw *PageWriter) LineCapStyle() LineCapStyle {
	return pw.lineCapStyle
}

func (pw *PageWriter) LineColor() Color {
	return pw.lineColor
}

func (pw *PageWriter) LineDashPattern() string {
	return pw.lineDashPattern
}

func (pw *PageWriter) LineTo(x, y float64) {
	pw.startGraph()
	if !pw.last.loc.equal(pw.loc) {
		if pw.inPath && pw.autoPath {
			pw.gw.stroke()
		}
		pw.inPath = false
	}
	pw.checkSetLineColor()
	pw.checkSetLineWidth()
	pw.checkSetLineDashPattern()

	if !pw.inPath {
		pw.gw.moveTo(pw.loc.x, pw.loc.y)
	}
	pw.MoveTo(x, y)
	pw.gw.lineTo(pw.loc.x, pw.loc.y)
	pw.inPath = true
	pw.last.loc = pw.loc
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

func (pw *PageWriter) ResetFonts() {
	pw.fonts = nil
}

func (pw *PageWriter) SetFont(name string, size float64, options Options) ([]*Font, error) {
	pw.ResetFonts()
	return pw.AddFont(name, size, options)
}

func (pw *PageWriter) SetLineCapStyle(lineCapStyle LineCapStyle) (prev LineCapStyle) {
	prev = pw.lineCapStyle
	pw.lineCapStyle = lineCapStyle
	return
}

func (pw *PageWriter) SetLineColor(color Color) (prev Color) {
	prev = pw.lineColor
	pw.lineColor = color
	return
}

func (pw *PageWriter) SetLineDashPattern(lineDashPattern string) (prev string) {
	prev = pw.lineDashPattern
	pw.lineDashPattern = lineDashPattern
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
	pw.last.loc = location{0, 0}
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
