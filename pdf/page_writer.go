// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"fmt"
	"leadtype/codepage"
	"leadtype/wordbreaking"
	"math"
	"strings"
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
	inPath     bool
	inText     bool
	isClosed   bool
	keepOrigin bool
	last       drawState
	line       *RichText
	lineHeight float64
	mw         *miscWriter
	options    Options
	origin     Location
	page       *page
	pageHeight float64
	pageWidth  float64
	stream     bytes.Buffer
	tw         *textWriter
	units      *units
	flushing   boolean
}

func newPageWriter(dw *DocWriter, options Options) *PageWriter {
	return new(PageWriter).init(dw, options)
}

func clonePageWriter(opw *PageWriter) *PageWriter {
	pw := new(PageWriter).init(opw.dw, opw.options)
	pw.drawState = opw.drawState
	pw.units = opw.units
	pw.fonts = append(pw.fonts, opw.fonts...)
	return pw
}

func (pw *PageWriter) init(dw *DocWriter, options Options) *PageWriter {
	pw.dw = dw
	pw.options = options
	pw.lineSpacing = options.FloatDefault("line_spacing", 1.0)
	pw.units = UnitConversions[options.StringDefault("units", "pt")]
	ps := newPageStyle(options)
	pw.pageHeight = ps.pageSize.y2
	pw.pageWidth = ps.pageSize.x2
	pw.page = newPage(pw.dw.nextSeq(), 0, pw.dw.catalog.pages)
	pw.page.setMediaBox(ps.pageSize)
	pw.page.setCropBox(ps.cropSize)
	pw.page.setRotate(ps.rotate)
	pw.page.setResources(pw.dw.resources)
	pw.dw.file.body.add(pw.page)
	pw.autoPath = true
	pw.mw = newMiscWriter(&pw.stream)
	pw.tw = newTextWriter(&pw.stream)
	pw.gw = newGraphWriter(&pw.stream)

	return pw
}

func (pw *PageWriter) AddFont(family string, options Options) ([]*Font, error) {
	if font, err := NewFont(family, options, pw.dw.fontSources); err != nil {
		return nil, err
	} else {
		return pw.addFont(font), nil
	}
}

func (pw *PageWriter) addFont(font *Font) []*Font {
	pw.fonts = append(pw.fonts, font)
	return pw.fonts
}

func (pw *PageWriter) carriageReturn() {
	pw.moveTo(pw.origin.X, pw.origin.Y)
}

func (pw *PageWriter) checkSetFont() {
	if len(pw.fonts) == 0 {
		pw.setDefaultFont()
	}
	if pw.last.fontKey != pw.fontKey {
		pw.tw.setFontAndSize(pw.fontKey, pw.fontSize)
		// TODO: check_set_v_text_align(true)
		pw.last.fontKey = pw.fontKey
	}
}

func (pw *PageWriter) checkSetFontColor() {
	if pw.fontColor == pw.last.fillColor {
		return
	}
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pw.mw.setRgbColorFill(pw.fontColor.RGB64())
	pw.last.fillColor = pw.fontColor
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

func (pw *PageWriter) close() {
	if pw.isClosed {
		return
	}
	// end margins
	// end sub page
	pw.endText()
	pw.endGraph()
	// compress stream
	pdfStream := newStream(pw.dw.nextSeq(), 0, pw.stream.Bytes())
	pw.dw.file.body.add(pdfStream)
	// set annots
	pw.page.add(pdfStream)
	pw.dw.catalog.pages.add(pw.page) // unless reusing page
	pw.stream.Reset()
	pw.isClosed = true
}

func (pw *PageWriter) drawUnderline(loc1 Location, loc2 Location, position float64, thickness float64) {
	saveWidth := pw.setLineWidth(thickness)
	// TODO: rotate coordiates given angle
	pw.moveTo(loc1.X, loc1.Y+position)
	pw.lineTo(loc2.X, loc2.Y+position)
	pw.setLineWidth(saveWidth)
}

func (pw *PageWriter) endGraph() {
	if pw.inPath {
		pw.endPath()
	}
	pw.inGraph = false
}

func (pw *PageWriter) endPath() {
	if pw.autoPath {
		pw.gw.stroke()
	}
	pw.inPath = false
}

func (pw *PageWriter) endText() {
	pw.flushText()
	pw.tw.close()
	pw.inText = false
}

func (pw *PageWriter) flushText() {
	if pw.line == nil || pw.flushing {
		return
	}
	pw.flushing = true
	pw.startText()
	if pw.loc != pw.last.loc {
		pw.tw.moveBy(pw.loc.X-pw.last.loc.X, pw.loc.Y-pw.last.loc.Y)
	}
	loc1 := pw.loc
	var buf bytes.Buffer
	pw.line.Merge().EachCodepage(func(cpi codepage.CodepageIndex, text string, p *RichText) {
		if p.Font == nil {
			fmt.Println(cpi)
			fmt.Println(text)
			panic("EachCodepage calling back with nil p")
		}
		buf.Reset()
		// fmt.Println(cpi)
		if cpi < 0 {
			// buf.WriteString(text)
		} else {
			cp := cpi.Codepage()
			for _, r := range text {
				ch, _ := cp.CharForCodepoint(r)
				buf.WriteByte(byte(ch))
			}
		}
		pw.SetFontColor(p.Color)
		pw.checkSetFontColor()
		pw.fontKey = pw.dw.fontKey(p.Font, cpi)
		pw.SetFontSize(p.FontSize)
		pw.checkSetFont()
		pw.tw.show(buf.Bytes())
	})
	pw.line.VisitAll(func(p *RichText) {
		if !p.IsLeaf() {
			return
		}
		loc2 := Location{loc1.X + p.Width(), loc1.Y} // TODO: Adjust if print at an angle.
		if p.Underline {
			pw.drawUnderline(loc1, loc2, p.UnderlinePosition, p.UnderlineThickness)
		}
		loc1 = loc2
	})
	pw.last.loc = pw.loc
	pw.lineHeight = math.Max(pw.lineHeight, pw.line.Leading()*pw.lineSpacing)
	pw.loc.X += pw.line.Width()
	// TODO: Adjust pw.loc.y if printing at an angle.

	pw.line = nil
	pw.flushing = false
}

func (pw *PageWriter) FontColor() Color {
	return pw.fontColor
}

func (pw *PageWriter) Fonts() []*Font {
	return pw.fonts
}

func (pw *PageWriter) FontSize() float64 {
	return pw.fontSize
}

func (pw *PageWriter) FontStyle() string {
	if len(pw.fonts) > 0 {
		return pw.fonts[0].style
	}
	return ""
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

func (pw *PageWriter) LineThrough() bool {
	return pw.lineThrough
}

func (pw *PageWriter) LineTo(x, y float64) {
	xpts, ypts := pw.units.toPts(x), pw.translate(pw.units.toPts(y))
	pw.lineTo(xpts, ypts)
}

func (pw *PageWriter) lineTo(x, y float64) {
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
		pw.gw.moveTo(pw.loc.X, pw.loc.Y)
	}
	pw.moveTo(x, y)
	pw.gw.lineTo(pw.loc.X, pw.loc.Y)
	pw.inPath = true
	pw.last.loc = pw.loc
}

func (pw *PageWriter) LineWidth(units string) float64 {
	return unitsFromPts(units, pw.lineWidth)
}

func (pw *PageWriter) Loc() Location {
	return Location{pw.X(), pw.Y()}
}

func (pw *PageWriter) MoveTo(x, y float64) {
	xpts, ypts := pw.units.toPts(x), pw.translate(pw.units.toPts(y))
	pw.moveTo(xpts, ypts)
}

func (pw *PageWriter) moveTo(x, y float64) {
	pw.flushText()
	pw.loc = Location{x, y}
	pw.lineHeight = 0
}

func (pw *PageWriter) newLine() {
	pw.flushText()
	if pw.lineHeight == 0 {
		if rt, err := pw.richTextForString("X"); err == nil {
			pw.lineHeight = rt.Leading() * pw.lineSpacing
		}
	}
	pw.moveTo(pw.origin.X, pw.origin.Y-pw.lineHeight)
}

func (pw *PageWriter) PageHeight() float64 {
	return pw.units.fromPts(pw.pageHeight)
}

func (pw *PageWriter) PageWidth() float64 {
	return pw.units.fromPts(pw.pageWidth)
}

func (pw *PageWriter) Print(text string) (err error) {
	i := strings.IndexAny(text, "\t\r\n")
	for i >= 0 {
		if err = pw.print(text[:i]); err != nil {
			return
		}
		switch text[i] {
		case '\t':
			pw.tab()
		case '\r':
			pw.carriageReturn()
		case '\n':
			pw.newLine()
		}
		text = text[i+1:]
		i = strings.IndexAny(text, "\t\r\n")
	}
	return pw.print(text)
}

func (pw *PageWriter) print(text string) (err error) {
	piece, err := pw.richTextForString(text)
	if err != nil {
		return
	}
	pw.startText()
	pw.PrintRichText(piece)
	return
}

func (pw *PageWriter) PrintParagraph(para []*RichText) {
	for _, p := range para {
		pw.PrintRichText(p)
		pw.newLine()
	}
}

func (pw *PageWriter) PrintRichText(text *RichText) {
	if pw.line == nil {
		if pw.keepOrigin {
			pw.keepOrigin = false
		} else {
			pw.origin = pw.loc
		}
		pw.line = text.Clone() // Avoid mutating text.
	} else {
		pw.line = pw.line.AddPiece(text)
	}
}

func (pw *PageWriter) PrintWithOptions(text string, options Options) (err error) {
	var para []*RichText
	rt, err := pw.richTextForString(text)
	if err != nil {
		return
	}
	if width := options.FloatDefault("width", 0); width > 0 {
		flags := make([]wordbreaking.Flags, rt.Len())
		wordbreaking.MarkRuneAttributes(rt.String(), flags)
		para = rt.WrapToWidth(pw.units.toPts(width), flags, false)
	} else {
		para = []*RichText{rt}
	}
	pw.PrintParagraph(para)
	return nil
}

func (pw *PageWriter) ResetFonts() {
	pw.fonts = nil
}

func (pw *PageWriter) richTextForString(text string) (piece *RichText, err error) {
	piece, err = NewRichText(text, pw.fonts, pw.fontSize, Options{
		"color": pw.fontColor, "line_through": pw.lineThrough, "underline": pw.underline})
	return
}

func (pw *PageWriter) setDefaultFont() {
	// TODO: Set Courier, Courier New or first font found.
}

func (pw *PageWriter) SetFont(name string, size float64, options Options) ([]*Font, error) {
	pw.ResetFonts()
	pw.SetFontSize(size)
	pw.SetFontColor(options["color"])
	return pw.AddFont(name, options)
}

func (pw *PageWriter) SetFontColor(color interface{}) (prev Color) {
	prev = pw.fontColor

	switch color := color.(type) {
	case string:
		if c, err := NamedColor(color); err == nil {
			pw.fontColor = c
		}
	case int:
		pw.fontColor = Color(color)
	case int32:
		pw.fontColor = Color(color)
	case Color:
		pw.fontColor = color
	}

	return
}

func (pw *PageWriter) SetFontSize(size float64) (prev float64) {
	prev = pw.fontSize
	pw.fontSize = size
	return
}

func (pw *PageWriter) SetFontStyle(style string) (prev string, err error) {
	prevFonts := pw.Fonts()
	if len(prevFonts) < 1 {
		return "", fmt.Errorf("No current font to apply style %s to.", style)
	}
	prev = prevFonts[0].style
	pw.ResetFonts()
	for _, font := range prevFonts {
		options := Options{
			"weight":       font.weight,
			"style":        style,
			"relativeSize": font.relativeSize,
		}
		if font.runeSet != nil {
			options["ranges"] = font.runeSet
		} else {
			options["ranges"] = font.ranges
		}
		if _, err = pw.AddFont(font.family, options); err != nil {
			break
		}
	}
	return
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

func (pw *PageWriter) SetLineSpacing(lineSpacing float64) (prev float64) {
	prev = pw.lineSpacing
	pw.lineSpacing = lineSpacing
	return
}

func (pw *PageWriter) SetLineThrough(lineThrough bool) (prev bool) {
	prev = pw.lineThrough
	pw.lineThrough = lineThrough
	return
}

func (pw *PageWriter) SetLineWidth(width float64, units string) (prev float64) {
	prev = unitsFromPts(units, pw.setLineWidth(unitsToPts(units, width)))
	return
}

func (pw *PageWriter) setLineWidth(width float64) (prev float64) {
	prev = pw.lineWidth
	pw.lineWidth = width
	return
}

func (pw *PageWriter) SetUnderline(underline bool) (prev bool) {
	prev = pw.underline
	pw.underline = underline
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
	pw.last.loc = Location{0, 0}
	pw.inGraph = true
}

func (pw *PageWriter) startText() {
	if pw.inText {
		return
	}
	if pw.inGraph {
		pw.endGraph()
	}
	pw.last.loc = Location{0, 0}
	pw.tw.open()
	pw.inText = true
}

func (pw *PageWriter) tab() {
	pw.keepOrigin = true
	// TODO: move to next horizontal tab position or print space
}

func (pw *PageWriter) translate(y float64) float64 {
	return pw.pageHeight - y
}

func (pw *PageWriter) Underline() bool {
	return pw.underline
}

func (pw *PageWriter) Units() string {
	return pw.units.name
}

func (pw *PageWriter) Write(text []byte) (n int, err error) {
	return len(text), pw.Print(string(text))
}

func (pw *PageWriter) X() float64 {
	return pw.units.fromPts(pw.loc.X)
}

func (pw *PageWriter) Y() float64 {
	return pw.units.fromPts(pw.translate(pw.loc.Y))
}
