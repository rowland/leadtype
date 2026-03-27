// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/rowland/leadtype/codepage"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/shaping"
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
	pathStates []pathState
	stream     bytes.Buffer
	tw         *textWriter
	units      *units
	vTextAlignPts float64
	flushing   boolean
}

type pathState struct {
	autoPath  bool
	fillColor colors.Color
	lineColor colors.Color
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
	pw.vTextAlign = options.StringDefault("v_text_align", "base")
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
	if pw.last.fontKey != pw.fontKey || pw.last.fontSize != pw.fontSize {
		pw.tw.setFontAndSize(pw.fontKey, pw.fontSize)
		pw.checkSetVTextAlign(true)
		pw.last.fontKey = pw.fontKey
		pw.last.fontSize = pw.fontSize
	}
}

func (pw *PageWriter) checkSetVTextAlign(force bool) {
	if !force && pw.vTextAlign == pw.last.vTextAlign {
		return
	}
	rise := 0.0
	if len(pw.fonts) > 0 {
		font := pw.fonts[0]
		scale := pw.fontSize * 0.001
		if upm := font.UnitsPerEm(); upm > 0 {
			scale = pw.fontSize / float64(upm)
		}
		top := float64(font.CapHeight()) * scale
		if top == 0 {
			top = float64(font.Ascent()) * scale
		}
		descent := float64(font.Descent()) * scale
		switch pw.vTextAlign {
		case "above":
			rise = -(top - descent)
		case "top":
			rise = -top
		case "middle":
			rise = -((top + descent) / 2.0)
		case "below":
			rise = -descent
		}
	}
	pw.vTextAlignPts = rise
	pw.tw.setRise(rise)
	pw.last.vTextAlign = pw.vTextAlign
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
	scale := math.Round(pw.lineWidth)
	if scale < 1 {
		scale = 1
	}
	if pat == nil {
		if scaled, ok := scaledDashPatternString(pw.lineDashPattern, scale); ok {
			pw.gw.setLineDashPattern(scaled)
		} else {
			pw.gw.setLineDashPattern(pw.lineDashPattern)
		}
	} else {
		scaled := make([]float64, len(pat.pattern))
		for i, value := range pat.pattern {
			scaled[i] = value * scale
		}
		pw.gw.setLineDashPattern((&linePattern{pattern: scaled, phase: pat.phase}).String())
	}
	pw.last.lineDashPattern = pw.lineDashPattern
	pw.gw.setLineCapStyle(int(pw.lineCapStyle))
	pw.last.lineCapStyle = pw.lineCapStyle
}

func scaledDashPatternString(pattern string, scale float64) (string, bool) {
	pattern = strings.TrimSpace(pattern)
	if !strings.HasPrefix(pattern, "[") {
		return "", false
	}
	end := strings.Index(pattern, "]")
	if end < 0 {
		return "", false
	}
	body := strings.TrimSpace(pattern[1:end])
	phaseText := strings.TrimSpace(pattern[end+1:])
	phase, err := strconv.Atoi(phaseText)
	if err != nil {
		return "", false
	}
	if body == "" {
		return (&linePattern{pattern: []float64{}, phase: phase}).String(), true
	}
	fields := strings.Fields(body)
	scaled := make([]float64, len(fields))
	for i, field := range fields {
		value, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return "", false
		}
		scaled[i] = value * scale
	}
	return (&linePattern{pattern: scaled, phase: phase}).String(), true
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

func (pw *PageWriter) checkSetSpacing() {
	if pw.charSpacing != pw.last.charSpacing {
		pw.tw.setCharSpacing(pw.charSpacing)
		pw.last.charSpacing = pw.charSpacing
	}
	if pw.wordSpacing != pw.last.wordSpacing {
		pw.tw.setWordSpacing(pw.wordSpacing)
		pw.last.wordSpacing = pw.wordSpacing
	}
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
	if pw.dw.compressPages {
		if err := pdfStream.compress(); err != nil {
			panic(err)
		}
	}
	pw.dw.file.body.add(pdfStream)
	// set annots
	pw.page.add(pdfStream)
	pw.dw.catalog.pages.add(pw.page) // unless reusing page
	pw.stream.Reset()
	pw.isClosed = true
}

var errTooFewPoints = errors.New("Need at least 4 points for curve")
var errNoActivePath = errors.New("No active manual path.")
var errPathAlreadyActive = errors.New("Manual path already active.")
var errInvalidPolygonSides = errors.New("Polygon requires at least 3 sides.")
var errInvalidStarPoints = errors.New("Star requires at least 5 points.")
var errTransformInsideManualPath = errors.New("Transform not allowed during active manual path.")

func (pw *PageWriter) beginManualPath() error {
	if len(pw.pathStates) > 0 {
		return errPathAlreadyActive
	}
	pw.flushText()
	pw.pathStates = append(pw.pathStates, pathState{
		autoPath:  pw.autoPath,
		fillColor: pw.fillColor,
		lineColor: pw.lineColor,
	})
	pw.autoPath = false
	return nil
}

func (pw *PageWriter) discardActivePath() {
	if !pw.inPath {
		return
	}
	pw.startGraph()
	pw.gw.newPath()
	pw.inPath = false
}

func (pw *PageWriter) restorePathState() {
	if len(pw.pathStates) == 0 {
		return
	}
	last := pw.pathStates[len(pw.pathStates)-1]
	pw.autoPath = last.autoPath
	pw.fillColor = last.fillColor
	pw.lineColor = last.lineColor
	pw.pathStates = pw.pathStates[:len(pw.pathStates)-1]
}

func (pw *PageWriter) requireActivePath() error {
	if len(pw.pathStates) == 0 || !pw.inPath {
		return errNoActivePath
	}
	return nil
}

// Path opens a scoped manual path session. During the session, auto-path
// stroking is disabled so callers can build compound paths and explicitly
// finalize them with Fill, Stroke, FillAndStroke, or Clip. If fn returns
// without finalizing the path, the unfinished path is discarded.
func (pw *PageWriter) Path(fn func()) error {
	if err := pw.beginManualPath(); err != nil {
		return err
	}
	defer func() {
		if len(pw.pathStates) > 0 {
			pw.discardActivePath()
			pw.restorePathState()
		}
	}()
	if fn != nil {
		fn()
	}
	return nil
}

// Fill fills the current manual path and restores the pre-path drawing state.
func (pw *PageWriter) Fill() error {
	if err := pw.requireActivePath(); err != nil {
		return err
	}
	pw.startGraph()
	pw.checkSetFillColor()
	pw.gw.fill()
	pw.inPath = false
	pw.restorePathState()
	return nil
}

// Stroke strokes the current manual path and restores the pre-path drawing state.
func (pw *PageWriter) Stroke() error {
	if err := pw.requireActivePath(); err != nil {
		return err
	}
	pw.startGraph()
	pw.checkSetLineColor()
	pw.checkSetLineWidth()
	pw.checkSetLineDashPattern()
	pw.gw.stroke()
	pw.inPath = false
	pw.restorePathState()
	return nil
}

// FillAndStroke fills and strokes the current manual path and restores the
// pre-path drawing state.
func (pw *PageWriter) FillAndStroke() error {
	if err := pw.requireActivePath(); err != nil {
		return err
	}
	pw.startGraph()
	pw.checkSetFillColor()
	pw.checkSetLineColor()
	pw.checkSetLineWidth()
	pw.checkSetLineDashPattern()
	pw.gw.fillAndStroke()
	pw.inPath = false
	pw.restorePathState()
	return nil
}

// Clip uses the current manual path as a clipping boundary for drawing within
// fn. The clipping region is scoped with save/restore graphics state.
func (pw *PageWriter) Clip(fn func()) error {
	if err := pw.requireActivePath(); err != nil {
		return err
	}
	pw.startGraph()
	pw.gw.saveGraphicsState()
	pw.gw.clip()
	pw.gw.newPath()
	pw.inPath = false
	pw.restorePathState()
	if fn != nil {
		fn()
	}
	if pw.inText {
		pw.endText()
	}
	if pw.inGraph {
		pw.endGraph()
	}
	pw.gw.restoreGraphicsState()
	return nil
}

func (pw *PageWriter) scopedTransform(a, b, c, d, x, y float64, fn func()) error {
	if len(pw.pathStates) > 0 {
		return errTransformInsideManualPath
	}
	if pw.inText {
		pw.endText()
	} else if pw.line != nil {
		pw.flushText()
		if pw.inText {
			pw.endText()
		}
	}
	if pw.inGraph {
		pw.endGraph()
	}
	pw.gw.saveGraphicsState()
	pw.gw.concatMatrix(a, b, c, d, x, y)
	if fn != nil {
		fn()
	}
	if pw.inText {
		pw.endText()
	} else if pw.line != nil {
		pw.flushText()
		if pw.inText {
			pw.endText()
		}
	}
	if pw.inGraph {
		pw.endGraph()
	}
	pw.gw.restoreGraphicsState()
	return nil
}

// Rotate applies a scoped rotation around (x, y) for all drawing performed
// within fn. Coordinates use the PageWriter's current unit system.
func (pw *PageWriter) Rotate(angle, x, y float64, fn func()) error {
	theta := angle * math.Pi / 180.0
	vCos := math.Cos(theta)
	vSin := math.Sin(theta)
	xpts := pw.units.toPts(x)
	ypts := pw.translate(pw.units.toPts(y))
	return pw.scopedTransform(
		vCos, vSin, -vSin, vCos,
		xpts-(xpts*vCos)+(ypts*vSin),
		ypts-(xpts*vSin)-(ypts*vCos),
		fn)
}

// Scale applies a scoped scale around (x, y) for all drawing performed within
// fn. Coordinates use the PageWriter's current unit system.
func (pw *PageWriter) Scale(x, y, scaleX, scaleY float64, fn func()) error {
	xpts := pw.units.toPts(x)
	ypts := pw.translate(pw.units.toPts(y))
	return pw.scopedTransform(
		scaleX, 0, 0, scaleY,
		xpts-(xpts*scaleX),
		ypts-(ypts*scaleY),
		fn)
}

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
		pw.gw.moveTo(pw.loc.X, pw.loc.Y)
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

func (pw *PageWriter) drawClosedShape(border, fill bool, build func()) (err error) {
	if border && fill {
		err = pw.Path(func() {
			build()
			err = pw.FillAndStroke()
		})
	} else if border {
		err = pw.Path(func() {
			build()
			err = pw.Stroke()
		})
	} else if fill {
		err = pw.Path(func() {
			build()
			err = pw.Fill()
		})
	} else {
		err = pw.Path(build)
	}
	return
}

// Line draws a line segment of the given length beginning at (x, y) and
// extending at angle degrees, where 0 is to the right and positive angles turn
// counter-clockwise in mathematical space.
func (pw *PageWriter) Line(x, y, angle, length float64) {
	dx, dy := rotateXY(1, 0, angle)
	pw.MoveTo(x, y)
	pw.LineTo(x+(dx*length), y-(dy*length))
	pw.autoStrokeAndFill(true, false)
}

func (pw *PageWriter) PointsForCircle(x, y, r float64) []Location {
	points := make([]Location, 0, 13)
	for q := 1; q <= 4; q++ {
		points = append(points, quadrantBezierPoints(q, x, y, r, r)...)
	}
	for _, i := range []int{12, 8, 4} {
		points = append(points[:i], points[i+1:]...)
	}
	return points
}

func (pw *PageWriter) Circle(x, y, r float64, border, fill, reverse bool) error {
	return pw.drawClosedShape(border, fill, func() {
		points := pw.PointsForCircle(x, y, r)
		if reverse {
			points = reverseCurvePoints(points)
		}
		_ = pw.CurvePoints(points)
	})
}

func (pw *PageWriter) PointsForEllipse(x, y, rx, ry float64) []Location {
	points := make([]Location, 0, 13)
	for q := 1; q <= 4; q++ {
		points = append(points, quadrantBezierPoints(q, x, y, rx, ry)...)
	}
	for _, i := range []int{12, 8, 4} {
		points = append(points[:i], points[i+1:]...)
	}
	return points
}

func (pw *PageWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	return pw.drawClosedShape(border, fill, func() {
		points := pw.PointsForEllipse(x, y, rx, ry)
		if reverse {
			points = reverseCurvePoints(points)
		}
		_ = pw.CurvePoints(points)
	})
}

func (pw *PageWriter) PointsForArc(x, y, r, startAngle, endAngle float64) []Location {
	if startAngle == endAngle {
		return nil
	}
	numArcs := 1
	ccwcw := 1.0
	arcSpan := endAngle - startAngle
	if endAngle < startAngle {
		ccwcw = -1.0
	}
	for math.Abs(arcSpan)/float64(numArcs) > 90.0 {
		numArcs++
	}
	angleBump := arcSpan / float64(numArcs)
	halfBump := 0.5 * angleBump
	curAngle := startAngle + halfBump
	points := make([]Location, 0, numArcs*3+1)
	for i := 0; i < numArcs; i++ {
		segment := calcArcSmall(r, curAngle, halfBump, ccwcw)
		for j, point := range segment {
			if i > 0 && j == 0 {
				continue
			}
			points = append(points, Location{x + point.X, y - point.Y})
		}
		curAngle += angleBump
	}
	return points
}

func (pw *PageWriter) Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error {
	points := pw.PointsForArc(x, y, r, startAngle, endAngle)
	if len(points) == 0 {
		return nil
	}
	if !moveToStart && pw.inPath {
		pw.LineTo(points[0].X, points[0].Y)
	}
	return pw.CurvePoints(points)
}

func (pw *PageWriter) Pie(x, y, r, startAngle, endAngle float64, border, fill, reverse bool) error {
	return pw.drawClosedShape(border, fill, func() {
		if reverse {
			startAngle, endAngle = endAngle, startAngle
		}
		pw.MoveTo(x, y)
		pw.LineTo(x+r*math.Cos(startAngle*math.Pi/180.0), y-r*math.Sin(startAngle*math.Pi/180.0))
		_ = pw.Arc(x, y, r, startAngle, endAngle, false)
		pw.LineTo(x, y)
	})
}

func (pw *PageWriter) Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error {
	if startAngle == endAngle {
		return nil
	}
	if reverse {
		startAngle, endAngle = endAngle, startAngle
	}
	arc1 := pw.PointsForArc(x, y, r1, startAngle, endAngle)
	arc2 := pw.PointsForArc(x, y, r2, endAngle, startAngle)
	if len(arc1) == 0 || len(arc2) == 0 {
		return nil
	}
	return pw.drawClosedShape(border, fill, func() {
		pw.MoveTo(arc1[0].X, arc1[0].Y)
		_ = pw.CurvePoints(arc1)
		pw.LineTo(arc2[0].X, arc2[0].Y)
		_ = pw.CurvePoints(arc2)
		pw.LineTo(arc1[0].X, arc1[0].Y)
	})
}

func (pw *PageWriter) PointsForPolygon(x, y, r float64, sides int, rotation float64) []Location {
	if sides < 3 {
		return nil
	}
	step := 360.0 / float64(sides)
	angle := step/2.0 + 90.0
	points := make([]Location, 0, sides+1)
	for i := 0; i <= sides; i++ {
		dx, dy := rotateXY(1, 0, angle)
		points = append(points, Location{x + dx*r, y - dy*r})
		angle += step
	}
	if rotation != 0 {
		center := Location{x, y}
		for i := range points {
			points[i] = rotatePoint(center, points[i], -rotation)
		}
	}
	return points
}

func (pw *PageWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	if sides < 3 {
		return errInvalidPolygonSides
	}
	return pw.drawClosedShape(border, fill, func() {
		points := pw.PointsForPolygon(x, y, r, sides, rotation)
		if reverse {
			LocationSlice(points).Reverse()
		}
		for i, point := range points {
			if i == 0 {
				pw.MoveTo(point.X, point.Y)
			} else {
				pw.LineTo(point.X, point.Y)
			}
		}
	})
}

func (pw *PageWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	if points < 5 {
		return errInvalidStarPoints
	}
	outer := pw.PointsForPolygon(x, y, r1, points, rotation)
	inner := pw.PointsForPolygon(x, y, r2, points, rotation+(360.0/float64(points)/2.0))
	if len(outer) == 0 || len(inner) == 0 {
		return errInvalidStarPoints
	}
	return pw.drawClosedShape(border, fill, func() {
		if reverse {
			LocationSlice(outer).Reverse()
			LocationSlice(inner).Reverse()
		}
		pw.MoveTo(inner[0].X, inner[0].Y)
		for i := 0; i < points; i++ {
			pw.LineTo(outer[i].X, outer[i].Y)
			pw.LineTo(inner[i+1].X, inner[i+1].Y)
		}
	})
}

func (pw *PageWriter) drawUnderline(loc1 Location, loc2 Location, position float64, thickness float64) {
	saveWidth := pw.setLineWidth(thickness)
	// TODO: rotate coordiates given angle
	offsetY := position - pw.vTextAlignPts
	pw.moveTo(loc1.X, loc1.Y+offsetY)
	pw.lineTo(loc2.X, loc2.Y+offsetY)
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
	savedFontColor := pw.fontColor
	savedFontKey := pw.fontKey
	savedFontSize := pw.fontSize
	savedCharSpacing := pw.charSpacing
	savedWordSpacing := pw.wordSpacing
	savedVTextAlign := pw.vTextAlign
	savedVTextAlignPts := pw.vTextAlignPts
	pw.startText()
	if pw.loc != pw.last.loc {
		pw.tw.moveBy(pw.loc.X-pw.last.loc.X, pw.loc.Y-pw.last.loc.Y)
	}
	loc1 := pw.loc
	textLoc := loc1
	var buf bytes.Buffer
	merged := pw.line.Merge()
	// Iterate leaf pieces directly. TrueType leaves are encoded as big-endian
	// uint16 glyph ID pairs. AFM/Type1 leaves use codepage-based encoding.
	merged.VisitAll(func(p *rich_text.RichText) {
		if !p.IsLeaf() || p.Text == "" || p.Font == nil {
			return
		}
		leafStart := textLoc
		textLoc.X += p.Width()
		if p.Font.SubType() == "TrueType" {
			fk := pw.dw.fontKeyUnicode(p.Font)
			psName := p.Font.PostScriptName()
			gr := pw.dw.glyphRecorders[psName]
			buf.Reset()

			var shaped []shaping.GlyphPosition
			var runes []rune // allocated only when shaping is attempted
			var glyphSequences map[int][]rune
			usePositionedGlyphs := false
			if p.Font.Shaper != nil && shaping.ContainsArabic(p.Text) {
				runes = []rune(p.Text)
				shaped, _ = p.Font.Shaper.Shape(runes, p.Font, float32(p.FontSize))
				glyphSequences = shapedGlyphSequences(shaped, runes)
			}

			if shaped != nil {
				usePositionedGlyphs = true
				penX := 0.0
				// Emit shaped glyphs in reverse visual order so that the
				// left-to-right PDF text stream matches the left-to-right
				// visual layout of the pre-shaped Arabic glyphs.
				for i := len(shaped) - 1; i >= 0; i-- {
					gp := shaped[i]
					if gr != nil {
						if seq := glyphSequences[gp.ClusterIndex]; len(seq) > 0 {
							gr.recordRunes(gp.GlyphID, seq)
						} else {
							gr.record(gp.GlyphID, runes[gp.ClusterIndex])
						}
					}
					buf.WriteByte(byte(gp.GlyphID >> 8))
					buf.WriteByte(byte(gp.GlyphID & 0xFF))
					pw.tw.setMatrix(
						1, 0, 0, 1,
						leafStart.X+penX+(float64(gp.XOffset)/64.0),
						leafStart.Y+(float64(gp.YOffset)/64.0),
					)
					pw.tw.showHex(buf.Bytes())
					buf.Reset()
					penX += float64(gp.XAdvance) / 64.0
				}
			} else {
				usePositionedGlyphs = p.CharSpacing != 0 || p.WordSpacing != 0
				if usePositionedGlyphs {
					penX := 0.0
					fsize := p.FontSize / float64(p.Font.UnitsPerEm())
					for _, r := range p.Text {
						gid := p.Font.GlyphIndex(r)
						if gr != nil {
							gr.record(gid, r)
						}
						advanceWidth, _ := p.Font.AdvanceWidth(r)
						buf.WriteByte(byte(gid >> 8))
						buf.WriteByte(byte(gid & 0xFF))
						pw.tw.setMatrix(1, 0, 0, 1, leafStart.X+penX, leafStart.Y)
						pw.tw.showHex(buf.Bytes())
						buf.Reset()
						penX += (fsize * float64(advanceWidth)) + p.CharSpacing
						if r == ' ' {
							penX += p.WordSpacing
						}
					}
				} else {
					for _, r := range p.Text {
						gid := p.Font.GlyphIndex(r)
						if gr != nil {
							gr.record(gid, r)
						}
						buf.WriteByte(byte(gid >> 8))
						buf.WriteByte(byte(gid & 0xFF))
					}
				}
			}
			if usePositionedGlyphs {
				pw.tw.setMatrix(1, 0, 0, 1, leafStart.X+p.Width(), leafStart.Y)
			}

			pw.SetFontColor(p.Color)
			pw.checkSetFontColor()
			pw.fontKey = fk
			pw.SetFontSize(p.FontSize)
			pw.checkSetFont()
			pw.checkSetVTextAlign(false)
			if usePositionedGlyphs {
				pw.charSpacing = 0
				pw.wordSpacing = 0
			} else {
				pw.charSpacing = p.CharSpacing
				pw.wordSpacing = p.WordSpacing
			}
			pw.checkSetSpacing()
			if !usePositionedGlyphs {
				pw.tw.show(buf.Bytes())
				// pw.tw.showHex(buf.Bytes())
			}
		} else {
			// AFM/Type1: still needs codepage-based encoding.
			p.EachCodepage(func(cpi codepage.CodepageIndex, text string, piece *rich_text.RichText) {
				buf.Reset()
				if cpi >= 0 {
					cp := cpi.Codepage()
					for _, r := range text {
						ch, _ := cp.CharForCodepoint(r)
						buf.WriteByte(byte(ch))
					}
				}
				pw.SetFontColor(piece.Color)
				pw.checkSetFontColor()
				pw.fontKey = pw.dw.fontKey(piece.Font, cpi)
				pw.SetFontSize(piece.FontSize)
				pw.checkSetFont()
				pw.checkSetVTextAlign(false)
				pw.charSpacing = piece.CharSpacing
				pw.wordSpacing = piece.WordSpacing
				pw.checkSetSpacing()
				pw.tw.show(buf.Bytes())
			})
		}
	})
	pw.line.VisitAll(func(p *rich_text.RichText) {
		if !p.IsLeaf() {
			return
		}
		loc2 := Location{loc1.X + p.Width(), loc1.Y} // TODO: Adjust if print at an angle.
		if p.Underline {
			pw.drawUnderline(loc1, loc2, p.UnderlinePosition, p.UnderlineThickness)
		}
		if p.Strikeout {
			pw.drawUnderline(loc1, loc2, p.StrikeoutPosition, p.StrikeoutThickness)
		}
		loc1 = loc2
	})
	pw.last.loc = pw.loc
	pw.lineHeight = math.Max(pw.lineHeight, pw.line.Leading()*pw.lineSpacing)
	pw.loc.X += pw.line.Width()
	// TODO: Adjust pw.loc.y if printing at an angle.

	pw.fontColor = savedFontColor
	pw.fontKey = savedFontKey
	pw.fontSize = savedFontSize
	pw.charSpacing = savedCharSpacing
	pw.wordSpacing = savedWordSpacing
	pw.vTextAlign = savedVTextAlign
	pw.vTextAlignPts = savedVTextAlignPts
	pw.line = nil
	pw.flushing = false
}

func shapedGlyphSequences(glyphs []shaping.GlyphPosition, runes []rune) map[int][]rune {
	if len(glyphs) == 0 || len(runes) == 0 {
		return nil
	}

	clusterStarts := make(map[int]struct{}, len(glyphs))
	for _, gp := range glyphs {
		if gp.ClusterIndex >= 0 && gp.ClusterIndex < len(runes) {
			clusterStarts[gp.ClusterIndex] = struct{}{}
		}
	}
	if len(clusterStarts) == 0 {
		return nil
	}

	sortedStarts := make([]int, 0, len(clusterStarts))
	for start := range clusterStarts {
		sortedStarts = append(sortedStarts, start)
	}
	sort.Ints(sortedStarts)

	sequences := make(map[int][]rune, len(sortedStarts))
	for i, start := range sortedStarts {
		end := len(runes)
		if i+1 < len(sortedStarts) {
			end = sortedStarts[i+1]
		}
		if start < end {
			sequences[start] = append([]rune(nil), runes[start:end]...)
		}
	}
	return sequences
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

func (pw *PageWriter) LineSpacing() float64 {
	return pw.lineSpacing
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

func (pw *PageWriter) Loc() (x, y float64) {
	return pw.X(), pw.Y()
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

func (pw *PageWriter) PrintParagraph(para []*rich_text.RichText, options options.Options) {
	pw.flushText()
	width := options.FloatDefault("width", pw.PageWidth()-pw.loc.X)
	for _, p := range para {
		pw.origin = pw.loc
		switch options.StringDefault("text-align", "left") {
		case "center":
			pw.keepOrigin = true
			pw.loc = Location{pw.loc.X + (width-p.Width())/2, pw.loc.Y}
			fmt.Println("center", pw.loc)
		case "right":
			pw.keepOrigin = true
			pw.loc = Location{pw.loc.X + width - p.Width(), pw.loc.Y}
		case "justify":
			delta := width - p.Width()
			spaces := 0
			for _, r := range p.String() {
				if r == rune(32) {
					spaces++
				}
			}
			words := spaces + 1
			if math.Abs(delta)/width < 0.4 {
				var charSpacing, wordSpacing float64
				if words == 1 {
					wordSpacing = 0
					charSpacing = delta / float64(p.Chars()-1)
				} else if math.Abs(delta)/float64(words) > 3 {
					wordSpacing = 3
					delta -= float64(words-1) * wordSpacing
					charSpacing = delta / float64(p.Chars()-1)
				} else {
					wordSpacing = delta / float64(words-1)
					charSpacing = 0
				}
				p = p.DeepClone()
				p.VisitAll(func(p *rich_text.RichText) {
					p.CharSpacing = charSpacing
					p.WordSpacing = wordSpacing
				})
			}
		}
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
	pw.PrintParagraph(para, options)
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
		"color": pw.fontColor, "strikeout": pw.strikeout, "underline": pw.underline})
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
	pw.flushText()
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

func (pw *PageWriter) SetStrikeout(strikeout bool) (prev bool) {
	prev = pw.strikeout
	pw.strikeout = strikeout
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

func (pw *PageWriter) SetVTextAlign(vTextAlign string) (prev string) {
	prev = pw.vTextAlign
	if pw.line != nil && prev != vTextAlign {
		pw.flushText()
	}
	pw.vTextAlign = vTextAlign
	return
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

func (pw *PageWriter) Strikeout() bool {
	return pw.strikeout
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

func (pw *PageWriter) VTextAlign() string {
	return pw.vTextAlign
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
