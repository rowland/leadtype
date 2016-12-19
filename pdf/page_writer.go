// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/rowland/leadtype/codepage"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/wordbreaking"
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
	fonts      []*font.Font
	gw         *graphWriter
	inGraph    bool
	inPath     bool
	inText     bool
	isClosed   bool
	keepOrigin bool
	last       drawState
	line       *rich_text.RichText
	lineHeight float64
	mw         *miscWriter
	options    options.Options
	origin     Location
	page       *page
	pageHeight float64
	pageWidth  float64
	stream     bytes.Buffer
	tw         *textWriter
	units      *units
	flushing   boolean
}

func newPageWriter(dw *DocWriter, options options.Options) *PageWriter {
	return new(PageWriter).init(dw, options)
}

func clonePageWriter(opw *PageWriter) *PageWriter {
	pw := new(PageWriter).init(opw.dw, opw.options)
	pw.drawState = opw.drawState
	pw.units = opw.units
	pw.fonts = append(pw.fonts, opw.fonts...)
	return pw
}

func (pw *PageWriter) init(dw *DocWriter, options options.Options) *PageWriter {
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

func (pw *PageWriter) AddFont(family string, options options.Options) ([]*font.Font, error) {
	if font, err := font.New(family, options, pw.dw.fontSources); err != nil {
		return nil, err
	} else {
		return pw.addFont(font), nil
	}
}

func (pw *PageWriter) addFont(font *font.Font) []*font.Font {
	pw.fonts = append(pw.fonts, font)
	return pw.fonts
}

func (pw *PageWriter) autoStrokeAndFill(stroke bool, fill bool) {
	if !pw.autoPath {
		return
	}
	if stroke && fill {
		pw.gw.fillAndStroke()
	} else if stroke {
		pw.gw.stroke()
	} else if fill {
		pw.gw.fill()
	} else {
		pw.gw.newPath()
	}
	pw.inPath = false
}

func (pw *PageWriter) carriageReturn() {
	pw.moveTo(pw.origin.X, pw.origin.Y)
}

func (pw *PageWriter) checkSetFillColor() {
	if pw.fillColor == pw.last.fillColor {
		return
	}
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pw.mw.setRgbColorFill(pw.fillColor.RGB64())
	pw.last.fillColor = pw.fillColor
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

var errTooFewPoints = errors.New("Need at least 4 points for curve")

func (pw *PageWriter) CurvePoints(points []Location) error {
	if len(points) < 4 {
		return errTooFewPoints
	}
	pw.MoveTo(points[0].X, points[0].Y)
	if !pw.last.loc.equal(pw.loc) {
		if pw.inPath && pw.autoPath {
			pw.gw.stroke()
			pw.inPath = false
		}
	}

	pw.checkSetLineColor()
	pw.checkSetLineWidth()
	pw.checkSetLineDashPattern()

	if !(pw.loc.equal(pw.last.loc) && pw.inPath) {
		pw.gw.moveTo(pw.units.toPts(pw.loc.X), pw.units.toPts(pw.loc.Y))
	}
	i := 1
	for i+2 < len(points) {
		pw.gw.curveTo(
			pw.units.toPts(points[i].X), pw.pageHeight-pw.units.toPts(points[i].Y),
			pw.units.toPts(points[i+1].X), pw.pageHeight-pw.units.toPts(points[i+1].Y),
			pw.units.toPts(points[i+2].X), pw.pageHeight-pw.units.toPts(points[i+2].Y))
		pw.MoveTo(points[i+2].X, points[i+2].Y)
		pw.last.loc = pw.loc
		i += 3
	}
	pw.inPath = true
	return nil
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
	pw.line.Merge().EachCodepage(func(cpi codepage.CodepageIndex, text string, p *rich_text.RichText) {
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
	pw.line.VisitAll(func(p *rich_text.RichText) {
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

func (pw *PageWriter) FontColor() colors.Color {
	return pw.fontColor
}

func (pw *PageWriter) Fonts() []*font.Font {
	return pw.fonts
}

func (pw *PageWriter) FontSize() float64 {
	return pw.fontSize
}

func (pw *PageWriter) FontStyle() string {
	if len(pw.fonts) > 0 {
		return pw.fonts[0].Style()
	}
	return ""
}

func (pw *PageWriter) LineCapStyle() LineCapStyle {
	return pw.lineCapStyle
}

func (pw *PageWriter) LineColor() colors.Color {
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

func (pw *PageWriter) PrintParagraph(para []*rich_text.RichText) {
	for _, p := range para {
		pw.PrintRichText(p)
		pw.newLine()
	}
}

func (pw *PageWriter) PrintRichText(text *rich_text.RichText) {
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

func (pw *PageWriter) PrintWithOptions(text string, options options.Options) (err error) {
	var para []*rich_text.RichText
	rt, err := pw.richTextForString(text)
	if err != nil {
		return
	}
	if width := options.FloatDefault("width", 0); width > 0 {
		flags := make([]wordbreaking.Flags, rt.Len())
		wordbreaking.MarkRuneAttributes(rt.String(), flags)
		para = rt.WrapToWidth(pw.units.toPts(width), flags, false)
	} else {
		para = []*rich_text.RichText{rt}
	}
	pw.PrintParagraph(para)
	return nil
}

func (pw *PageWriter) Rectangle(x, y, width, height float64, border bool, fill bool) {
	pw.rectangle(x, y, width, height, border, fill, nil, false, false)
}

func (pw *PageWriter) Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool) {
	pw.rectangle(x, y, width, height, border, fill, corners, path, reverse)
}

func (pw *PageWriter) rectangle(x, y, width, height float64, border, fill bool, corners []float64, path, reverse bool) {
	xpts, ypts := pw.units.toPts(x), pw.translate(pw.units.toPts(y+height))
	wpts, hpts := pw.units.toPts(width), pw.units.toPts(height)

	pw.startGraph()
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	if border {
		pw.checkSetLineColor()
		pw.checkSetLineWidth()
		pw.checkSetLineDashPattern()
	}
	if fill {
		pw.checkSetFillColor()
	}

	if len(corners) > 0 {
		pw.roundedRectangle(x, y, width, height, corners, reverse)
	} else if path || reverse {
		pw.rectanglePath(x, y, width, height, reverse)
	} else {
		pw.gw.rectangle(xpts, ypts, wpts, hpts)
	}
	pw.autoStrokeAndFill(border, fill)
	pw.MoveTo(x+width, y)
}

func (pw *PageWriter) rectanglePath(x, y, width, height float64, reverse bool) {
	pw.MoveTo(x, y)
	if reverse {
		pw.LineTo(x, y+height)
		pw.LineTo(x+width, y+height)
		pw.LineTo(x+width, y)
	} else {
		pw.LineTo(x+width, y)
		pw.LineTo(x+width, y+height)
		pw.LineTo(x, y+height)
	}
	pw.LineTo(x, y)
}

func (pw *PageWriter) ResetFonts() {
	pw.fonts = nil
}

func (pw *PageWriter) richTextForString(text string) (piece *rich_text.RichText, err error) {
	piece, err = rich_text.New(text, pw.fonts, pw.fontSize, options.Options{
		"color": pw.fontColor, "line_through": pw.lineThrough, "underline": pw.underline})
	return
}

func (pw *PageWriter) roundedRectangle(x, y, width, height float64, corners []float64, reverse bool) {
	var xr1, yr1, xr2, yr2, xr3, yr3, xr4, yr4 float64
	switch len(corners) {
	case 1:
		xr1, yr1, xr2, yr2, xr3, yr3, xr4, yr4 =
			corners[0], corners[0], corners[0], corners[0], corners[0], corners[0], corners[0], corners[0]
	case 2:
		xr1, yr1, xr2, yr2 = corners[0], corners[0], corners[0], corners[0]
		xr3, yr3, xr4, yr4 = corners[1], corners[1], corners[1], corners[1]
	case 4:
		xr1, yr1 = corners[0], corners[0]
		xr2, yr2 = corners[1], corners[1]
		xr3, yr3 = corners[2], corners[2]
		xr4, yr4 = corners[3], corners[3]
	case 8:
		xr1, yr1, xr2, yr2, xr3, yr3, xr4, yr4 =
			corners[0], corners[1], corners[2], corners[3], corners[4], corners[5], corners[6], corners[7]
	default:
		// xr1, yr1, xr2, yr2, xr3, yr3, xr4, yr4 = 0
	}

	q2p := quadrantBezierPoints(2, x+xr1, y+yr1, xr1, yr1)
	q1p := quadrantBezierPoints(1, x+width-xr2, y+yr2, xr2, yr2)
	q4p := quadrantBezierPoints(4, x+width-xr3, y+height-yr3, xr3, yr3)
	q3p := quadrantBezierPoints(3, x+xr4, y+height-yr4, xr4, yr4)
	qpa := [][]Location{q1p, q2p, q3p, q4p}

	if reverse {
		LocationSliceSlice(qpa).Reverse()
		for _, qp := range qpa {
			LocationSlice(qp).Reverse()
		}
	}

	pw.CurvePoints(qpa[0])
	pw.LineTo(qpa[1][0].X, qpa[1][0].Y)
	pw.CurvePoints(qpa[1])
	pw.LineTo(qpa[2][0].X, qpa[2][0].Y)
	pw.CurvePoints(qpa[2])
	pw.LineTo(qpa[3][0].X, qpa[3][0].Y)
	pw.CurvePoints(qpa[3])
	pw.LineTo(qpa[0][0].X, qpa[0][0].Y)
}

func (pw *PageWriter) setDefaultFont() {
	// TODO: Set Courier, Courier New or first font found.
}

func (pw *PageWriter) SetFont(name string, size float64, options options.Options) ([]*font.Font, error) {
	pw.ResetFonts()
	pw.SetFontSize(size)
	pw.SetFontColor(options["color"])
	return pw.AddFont(name, options)
}

func (pw *PageWriter) SetFillColor(value interface{}) (prev colors.Color) {
	prev = pw.fillColor

	switch value := value.(type) {
	case string:
		if c, err := colors.NamedColor(value); err == nil {
			pw.fillColor = c
		}
	case int:
		pw.fillColor = colors.Color(value)
	case int32:
		pw.fillColor = colors.Color(value)
	case colors.Color:
		pw.fillColor = value
	}

	return
}

func (pw *PageWriter) SetFontColor(value interface{}) (prev colors.Color) {
	prev = pw.fontColor

	switch value := value.(type) {
	case string:
		c, _ := colors.NamedColor(value)
		pw.fontColor = c
	case int:
		pw.fontColor = colors.Color(value)
	case int32:
		pw.fontColor = colors.Color(value)
	case colors.Color:
		pw.fontColor = value
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
	prev = prevFonts[0].Style()
	pw.ResetFonts()
	for _, font := range prevFonts {
		options := options.Options{
			"weight":       font.Weight,
			"style":        style,
			"relativeSize": font.RelativeSize,
		}
		if font.RuneSet != nil {
			options["ranges"] = font.RuneSet
		} else {
			options["ranges"] = font.Ranges
		}
		if _, err = pw.AddFont(font.Family(), options); err != nil {
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

func (pw *PageWriter) SetLineColor(value colors.Color) (prev colors.Color) {
	prev = pw.lineColor
	pw.lineColor = value
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
