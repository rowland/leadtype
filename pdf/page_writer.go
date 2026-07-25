// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/rowland/leadtype/codepage"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/shaping"
	"github.com/rowland/leadtype/svg"
	"github.com/rowland/leadtype/wordbreaking"
)

type LineCapStyle int

const (
	ButtCap             = LineCapStyle(iota)
	RoundCap            = LineCapStyle(iota)
	ProjectingSquareCap = LineCapStyle(iota)
)

type LineJoinStyle int

const (
	MiterJoin = LineJoinStyle(iota)
	RoundJoin
	BevelJoin
)

type TextMetrics struct {
	Width   float64
	Height  float64
	Ascent  float64
	Descent float64
}

type PageWriter struct {
	drawState
	autoPath              bool
	dw                    *DocWriter
	fonts                 []*font.Font
	gw                    *graphWriter
	inGraph               bool
	inPath                bool
	inText                bool
	isClosed              bool
	keepOrigin            bool
	last                  drawState
	line                  *rich_text.RichText
	lineHeight            float64
	mw                    *miscWriter
	options               options.Options
	origin                Location
	page                  *page
	pageHeight            float64
	pageWidth             float64
	pathStates            []pathState
	supportsArabicShaping bool
	stream                bytes.Buffer
	tw                    *textWriter
	units                 *units
	vTextAlignPts         float64
	flushing              boolean
	artifactDepth         int
	accessibilityStack    []*structElem
	memoCapture           bool
}

type pathState struct {
	autoPath     bool
	fillColor    colors.Color
	fillGradient string
	lineColor    colors.Color
	lineGradient string
	lineWidth    float64
	miter        float64
	lineCap      LineCapStyle
	lineJoin     LineJoinStyle
	lineDash     string
}

func newPageWriter(dw *DocWriter, options options.Options) *PageWriter {
	return new(PageWriter).init(dw, options)
}

func newContentWriter(dw *DocWriter, options options.Options, pageWidth, pageHeight float64) *PageWriter {
	return new(PageWriter).initContent(dw, options, pageWidth, pageHeight)
}

func newContentWriterFromPage(base *PageWriter, pageWidth, pageHeight float64) *PageWriter {
	opts := base.options
	if opts == nil {
		opts = options.Options{}
	}
	content := newContentWriter(base.dw, opts.Merge(options.Options{"units": base.Units()}), pageWidth, pageHeight)
	content.drawState = base.drawState
	content.loc = Location{}
	content.last = drawState{}
	content.fonts = append([]*font.Font(nil), base.fonts...)
	content.supportsArabicShaping = base.supportsArabicShaping
	content.memoCapture = true
	return content
}

func clonePageWriter(opw *PageWriter) *PageWriter {
	pw := new(PageWriter).init(opw.dw, opw.options)
	pw.drawState = opw.drawState
	pw.units = opw.units
	pw.fonts = append(pw.fonts, opw.fonts...)
	pw.supportsArabicShaping = opw.supportsArabicShaping
	return pw
}

func (pw *PageWriter) init(dw *DocWriter, options options.Options) *PageWriter {
	ps := newPageStyle(options)
	pw.initContent(dw, options, ps.pageSize.x2, ps.pageSize.y2)
	pw.page = newPage(pw.dw.nextSeq(), 0, pw.dw.catalog.pages)
	pw.page.setMediaBox(ps.pageSize)
	pw.page.setCropBox(ps.cropSize)
	pw.page.setRotate(ps.rotate)
	pw.page.setResources(pw.dw.resources)
	pw.dw.file.body.add(pw.page)
	return pw
}

func (pw *PageWriter) initContent(dw *DocWriter, options options.Options, pageWidth, pageHeight float64) *PageWriter {
	pw.dw = dw
	pw.options = options
	pw.lineSpacing = options.FloatDefault("line_spacing", 1.0)
	pw.units = UnitConversions[options.StringDefault("units", "pt")]
	pw.vTextAlign = parseVerticalTextAlign(options.StringDefault("v_text_align", "base"))
	pw.pageHeight = pageHeight
	pw.pageWidth = pageWidth
	pw.autoPath = true
	pw.lineJoinStyle = MiterJoin
	pw.last.lineJoinStyle = MiterJoin
	pw.miterLimit = 10
	pw.last.miterLimit = 10
	pw.mw = newMiscWriter(&pw.stream)
	pw.tw = newTextWriter(&pw.stream)
	pw.gw = newGraphWriter(&pw.stream)

	return pw
}

func (pw *PageWriter) AddFont(family string, options options.Options) ([]*font.Font, error) {
	if pw.dw.profiler != nil {
		defer pw.dw.profiler.Begin("pdf.font.select").End()
	}
	if font, err := pw.dw.selectFont(family, options); err != nil {
		return nil, err
	} else {
		return pw.addFont(font), nil
	}
}

func (pw *PageWriter) addFont(font *font.Font) []*font.Font {
	pw.fonts = append(pw.fonts, font)
	if font.SupportsArabic() {
		pw.supportsArabicShaping = true
	}
	return pw.fonts
}

func fontsSupportArabicShaping(fonts []*font.Font) bool {
	for _, font := range fonts {
		if font.SupportsArabic() {
			return true
		}
	}
	return false
}

func (pw *PageWriter) WithAccessibilityTag(tag string, opts AccessibilityOptions, fn func()) error {
	if pw.memoCapture {
		if fn != nil {
			fn()
		}
		return nil
	}
	if tag == "" {
		if fn != nil {
			fn()
		}
		return nil
	}
	pw.dw.EnableTaggedPDF(true)
	pw.flushText()
	elem := pw.dw.accessibility.resolveElement(pw.currentStructElem(), tag, opts)
	if elem == nil {
		if fn != nil {
			fn()
		}
		return nil
	}
	pw.accessibilityStack = append(pw.accessibilityStack, elem)
	pw.dw.accessibility.stack = append(pw.dw.accessibility.stack, elem)
	defer func() {
		pw.flushText()
		if len(pw.accessibilityStack) > 0 {
			pw.accessibilityStack = pw.accessibilityStack[:len(pw.accessibilityStack)-1]
		}
		if len(pw.dw.accessibility.stack) > 0 {
			pw.dw.accessibility.stack = pw.dw.accessibility.stack[:len(pw.dw.accessibility.stack)-1]
		}
	}()
	if fn != nil {
		fn()
	}
	return nil
}

func (pw *PageWriter) WithAccessibilityArtifact(fn func()) error {
	if pw.memoCapture {
		if fn != nil {
			fn()
		}
		return nil
	}
	if !pw.dw.taggedPDFEnabled() {
		if fn != nil {
			fn()
		}
		return nil
	}
	pw.flushText()
	if pw.artifactDepth == 0 {
		pw.mw.beginMarkedContent("Artifact")
		defer pw.mw.endMarkedContent()
	}
	pw.artifactDepth++
	defer func() {
		pw.flushText()
		pw.artifactDepth--
	}()
	if fn != nil {
		fn()
	}
	return nil
}

func (pw *PageWriter) beginTaggedContent(tag string, elem *structElem) bool {
	if !pw.dw.taggedPDFEnabled() || elem == nil || tag == "" {
		return false
	}
	mcid := pw.dw.accessibility.associateMarkedContent(pw.page, elem)
	if mcid < 0 {
		return false
	}
	pw.mw.beginMarkedContentWithProperties(tag, dictionary{
		"MCID": integer(mcid),
	})
	return true
}

func (pw *PageWriter) beginActualTextContent(actualText string) bool {
	if actualText == "" {
		return false
	}
	pw.mw.beginMarkedContentWithProperties("Span", dictionary{
		"ActualText": textString(actualText),
	})
	return true
}

func (pw *PageWriter) currentStructElem() *structElem {
	if pw.memoCapture {
		return nil
	}
	if len(pw.accessibilityStack) > 0 {
		return pw.accessibilityStack[len(pw.accessibilityStack)-1]
	}
	if pw.dw.taggedPDFEnabled() {
		return pw.dw.accessibility.currentStructElem()
	}
	return nil
}

func (pw *PageWriter) structElemForLeaf(p *rich_text.RichText) (*structElem, string) {
	if !pw.dw.taggedPDFEnabled() {
		return nil, ""
	}
	if pw.artifactDepth > 0 {
		return nil, ""
	}
	current := pw.currentStructElem()
	if current == nil {
		return nil, ""
	}
	if p != nil && (p.LinkURI != "" || p.LinkTarget != "") {
		elem := pw.dw.accessibility.resolveElement(current, "Link", AccessibilityOptions{ID: p.LinkID})
		return elem, "Link"
	}
	return current, current.s
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
	if pw.fillGradient != "" {
		pw.checkSetFillGradient()
		return
	}
	if pw.last.fillGradient != "" {
		// Reverting from gradient to solid color.
		pw.mw.setRgbColorFill(pw.fillColor.RGB64())
		pw.last.fillColor = pw.fillColor
		pw.last.fillGradient = ""
		return
	}
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

func (pw *PageWriter) checkSetFillGradient() {
	if pw.fillGradient == pw.last.fillGradient {
		return
	}
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pw.mw.setColorSpaceFill("Pattern")
	pw.mw.setColorFillPattern(pw.fillGradient)
	pw.last.fillGradient = pw.fillGradient
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
		case VTextAlignAbove:
			rise = -(top - descent)
		case VTextAlignTop:
			rise = -top
		case VTextAlignMiddle:
			rise = -((top + descent) / 2.0)
		case VTextAlignBelow:
			rise = -descent
		}
	}
	pw.vTextAlignPts = rise
	pw.tw.setRise(rise)
	pw.last.vTextAlign = pw.vTextAlign
}

func (pw *PageWriter) checkSetFontColor() {
	if pw.last.fillGradient == "" && pw.fontColor == pw.last.fillColor {
		return
	}
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pw.mw.setRgbColorFill(pw.fontColor.RGB64())
	pw.last.fillColor = pw.fontColor
	pw.last.fillGradient = ""
}

func (pw *PageWriter) checkSetLineColor() {
	if pw.lineGradient != "" {
		pw.checkSetLineGradient()
		return
	}
	if pw.last.lineGradient != "" {
		pw.mw.setRgbColorStroke(pw.lineColor.RGB64())
		pw.last.lineColor = pw.lineColor
		pw.last.lineGradient = ""
		return
	}
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

func (pw *PageWriter) checkSetLineGradient() {
	if pw.lineGradient == pw.last.lineGradient {
		return
	}
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pw.mw.setColorSpaceStroke("Pattern")
	pw.mw.setColorStrokePattern(pw.lineGradient)
	pw.last.lineGradient = pw.lineGradient
}

func (pw *PageWriter) checkSetLineDashPattern() {
	if pw.lineDashPattern == pw.last.lineDashPattern &&
		pw.lineCapStyle == pw.last.lineCapStyle &&
		pw.lineJoinStyle == pw.last.lineJoinStyle &&
		pw.miterLimit == pw.last.miterLimit {
		return
	}
	pw.startGraph()
	if pw.inPath && pw.autoPath {
		pw.gw.stroke()
		pw.inPath = false
	}
	pat := LinePatterns[pw.lineDashPattern]
	if pw.lineDashPattern == "" {
		pat = LinePatterns["solid"]
	}
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
	if pw.lineJoinStyle != pw.last.lineJoinStyle {
		pw.gw.setLineJoinStyle(int(pw.lineJoinStyle))
		pw.last.lineJoinStyle = pw.lineJoinStyle
	}
	if pw.miterLimit != pw.last.miterLimit {
		pw.gw.setMiterLimit(pw.miterLimit)
		pw.last.miterLimit = pw.miterLimit
	}
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
var errInvalidStarPoints = errors.New("Star requires at least 2 points.")
var errTransformInsideManualPath = errors.New("Transform not allowed during active manual path.")
var errTextClipInsideManualPath = errors.New("Text clipping not allowed during active manual path.")

type textEmissionOptions struct {
	renderMode        int
	applyFillColor    bool
	applyStrokeState  bool
	emitDecorations   bool
	emitLinks         bool
	emitAccessibility bool
}

var visibleTextEmission = textEmissionOptions{
	applyFillColor:    true,
	applyStrokeState:  true,
	emitDecorations:   true,
	emitLinks:         true,
	emitAccessibility: true,
}

var clipTextEmission = textEmissionOptions{
	renderMode: 7,
}

var fillStrokeClipTextEmission = textEmissionOptions{
	renderMode:       6,
	applyFillColor:   true,
	applyStrokeState: true,
}

func (pw *PageWriter) beginManualPath() error {
	if len(pw.pathStates) > 0 {
		return errPathAlreadyActive
	}
	pw.flushText()
	pw.pathStates = append(pw.pathStates, pathState{
		autoPath:     pw.autoPath,
		fillColor:    pw.fillColor,
		fillGradient: pw.fillGradient,
		lineColor:    pw.lineColor,
		lineGradient: pw.lineGradient,
		lineWidth:    pw.lineWidth,
		miter:        pw.miterLimit,
		lineCap:      pw.lineCapStyle,
		lineJoin:     pw.lineJoinStyle,
		lineDash:     pw.lineDashPattern,
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
	pw.fillGradient = last.fillGradient
	pw.lineColor = last.lineColor
	pw.lineGradient = last.lineGradient
	pw.lineWidth = last.lineWidth
	pw.miterLimit = last.miter
	pw.lineCapStyle = last.lineCap
	pw.lineJoinStyle = last.lineJoin
	pw.lineDashPattern = last.lineDash
	pw.pathStates = pw.pathStates[:len(pw.pathStates)-1]
}

func (pw *PageWriter) requireActivePath() error {
	if len(pw.pathStates) == 0 || !pw.inPath {
		return errNoActivePath
	}
	return nil
}

func (pw *PageWriter) requirePathSession() error {
	if len(pw.pathStates) == 0 {
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
	savedLast := pw.last
	pw.startGraph()
	pw.gw.saveGraphicsState()
	pw.gw.clip()
	pw.gw.newPath()
	pw.inPath = false
	pw.restorePathState()
	if fn != nil {
		fn()
	}
	pw.endText()
	if pw.inGraph {
		pw.endGraph()
	}
	pw.gw.restoreGraphicsState()
	pw.last = savedLast
	return nil
}

func (pw *PageWriter) scopedTransform(a, b, c, d, x, y float64, fn func()) error {
	if len(pw.pathStates) > 0 {
		return errTransformInsideManualPath
	}
	savedLast := pw.last
	pw.endText()
	if pw.inGraph {
		pw.endGraph()
	}
	pw.gw.saveGraphicsState()
	pw.gw.concatMatrix(a, b, c, d, x, y)
	if fn != nil {
		fn()
	}
	pw.endText()
	if pw.inGraph {
		pw.endGraph()
	}
	pw.gw.restoreGraphicsState()
	pw.last = savedLast
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
	pw.startGraph()
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

func (pw *PageWriter) appendCurvePoints(points []Location) error {
	if len(points) < 4 {
		return errTooFewPoints
	}
	if !pw.loc.equal(points[0]) {
		pw.LineTo(points[0].X, points[0].Y)
	}
	if !pw.inPath {
		pw.startGraph()
		pw.checkSetLineColor()
		pw.checkSetLineWidth()
		pw.checkSetLineDashPattern()
		pw.gw.moveTo(pw.loc.X, pw.loc.Y)
		pw.inPath = true
		pw.last.loc = pw.loc
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

func (pw *PageWriter) annularArcPath(x, y, r1, r2, startAngle, endAngle float64) error {
	arc1 := pw.PointsForArc(x, y, r1, startAngle, endAngle)
	arc2 := pw.PointsForArc(x, y, r2, endAngle, startAngle)
	if len(arc1) == 0 || len(arc2) == 0 {
		return nil
	}
	pw.MoveTo(arc1[0].X, arc1[0].Y)
	if err := pw.appendCurvePoints(arc1); err != nil {
		return err
	}
	pw.LineTo(arc2[0].X, arc2[0].Y)
	if err := pw.appendCurvePoints(arc2); err != nil {
		return err
	}
	pw.LineTo(arc1[0].X, arc1[0].Y)
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

func (pw *PageWriter) ClosedShapeBounds(shape ClosedShape) (Bounds, error) {
	return shape.Bounds()
}

func (pw *PageWriter) buildClosedShapePath(shape ClosedShape) error {
	shape = shape.normalized()
	if err := shape.validate(); err != nil {
		return err
	}
	switch shape.Kind {
	case ClosedShapeCircle:
		points := circlePoints(shape.Center.X, shape.Center.Y, shape.RadiusX)
		if shape.Reverse {
			points = reverseCurvePoints(points)
		}
		if err := pw.CurvePoints(points); err != nil {
			return err
		}
		pw.gw.closePath()
		return nil
	case ClosedShapeEllipse:
		points := ellipsePoints(shape.Center.X, shape.Center.Y, shape.RadiusX, shape.RadiusY)
		if shape.Rotation != 0 {
			center := shape.Center
			for i := range points {
				points[i] = rotatePoint(center, points[i], -shape.Rotation)
			}
		}
		if shape.Reverse {
			points = reverseCurvePoints(points)
		}
		if err := pw.CurvePoints(points); err != nil {
			return err
		}
		pw.gw.closePath()
		return nil
	case ClosedShapePolygon:
		points := polygonPoints(shape.Center.X, shape.Center.Y, shape.Radius, shape.Sides, shape.Rotation)
		if shape.Reverse {
			LocationSlice(points).Reverse()
		}
		for i, point := range points {
			if i == 0 {
				pw.MoveTo(point.X, point.Y)
			} else {
				pw.LineTo(point.X, point.Y)
			}
		}
		return nil
	case ClosedShapeStar:
		points := starPoints(shape.Center.X, shape.Center.Y, shape.Radius, shape.InnerRadius, shape.Points, shape.Rotation)
		if len(points) == 0 {
			return errInvalidStarPoints
		}
		if shape.Reverse {
			LocationSlice(points).Reverse()
		}
		for i, point := range points {
			if i == 0 {
				pw.MoveTo(point.X, point.Y)
			} else {
				pw.LineTo(point.X, point.Y)
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported closed shape %q", shape.Kind)
	}
}

func (pw *PageWriter) DrawClosedShape(shape ClosedShape, border, fill bool) error {
	shape = shape.normalized()
	if err := shape.validate(); err != nil {
		return err
	}
	return pw.drawClosedShape(border, fill, func() {
		_ = pw.buildClosedShapePath(shape)
	})
}

func (pw *PageWriter) AppendClosedShapePath(shape ClosedShape) error {
	if err := pw.requirePathSession(); err != nil {
		return err
	}
	shape = shape.normalized()
	if err := shape.validate(); err != nil {
		return err
	}
	return pw.buildClosedShapePath(shape)
}

func (pw *PageWriter) ClipClosedShape(shape ClosedShape, fn func()) error {
	shape = shape.normalized()
	if err := shape.validate(); err != nil {
		return err
	}
	var clipErr error
	if err := pw.Path(func() {
		if e := pw.buildClosedShapePath(shape); e != nil {
			clipErr = e
			return
		}
		if e := pw.Clip(fn); e != nil && clipErr == nil {
			clipErr = e
		}
	}); err != nil {
		return err
	}
	return clipErr
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
	return pw.DrawClosedShape(ClosedShape{
		Kind:    ClosedShapeCircle,
		Center:  Location{X: x, Y: y},
		Radius:  r,
		Reverse: reverse,
	}, border, fill)
}

func (pw *PageWriter) CirclePath(x, y, r float64, reverse bool) error {
	if err := pw.requirePathSession(); err != nil {
		return err
	}
	return pw.buildClosedShapePath(ClosedShape{
		Kind:    ClosedShapeCircle,
		Center:  Location{X: x, Y: y},
		Radius:  r,
		Reverse: reverse,
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
	return pw.DrawClosedShape(ClosedShape{
		Kind:    ClosedShapeEllipse,
		Center:  Location{X: x, Y: y},
		RadiusX: rx,
		RadiusY: ry,
		Reverse: reverse,
	}, border, fill)
}

func (pw *PageWriter) EllipsePath(x, y, rx, ry float64, reverse bool) error {
	if err := pw.requirePathSession(); err != nil {
		return err
	}
	return pw.buildClosedShapePath(ClosedShape{
		Kind:    ClosedShapeEllipse,
		Center:  Location{X: x, Y: y},
		RadiusX: rx,
		RadiusY: ry,
		Reverse: reverse,
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

func (pw *PageWriter) appendPiePath(x, y, r, startAngle, endAngle float64, reverse bool) error {
	if reverse {
		startAngle, endAngle = endAngle, startAngle
	}
	pw.MoveTo(x, y)
	pw.LineTo(x+r*math.Cos(startAngle*math.Pi/180.0), y-r*math.Sin(startAngle*math.Pi/180.0))
	if err := pw.Arc(x, y, r, startAngle, endAngle, false); err != nil {
		return err
	}
	pw.LineTo(x, y)
	return nil
}

func (pw *PageWriter) Pie(x, y, r, startAngle, endAngle float64, border, fill, reverse bool) error {
	return pw.drawClosedShape(border, fill, func() {
		_ = pw.appendPiePath(x, y, r, startAngle, endAngle, reverse)
	})
}

func (pw *PageWriter) AppendPiePath(x, y, r, startAngle, endAngle float64, reverse bool) error {
	if err := pw.requirePathSession(); err != nil {
		return err
	}
	return pw.appendPiePath(x, y, r, startAngle, endAngle, reverse)
}

func (pw *PageWriter) Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error {
	if startAngle == endAngle {
		return nil
	}
	if reverse {
		startAngle, endAngle = endAngle, startAngle
	}
	return pw.drawClosedShape(border, fill, func() {
		_ = pw.annularArcPath(x, y, r1, r2, startAngle, endAngle)
	})
}

func (pw *PageWriter) AppendArchPath(x, y, r1, r2, startAngle, endAngle float64, reverse bool) error {
	if startAngle == endAngle {
		return nil
	}
	if err := pw.requirePathSession(); err != nil {
		return err
	}
	if reverse {
		startAngle, endAngle = endAngle, startAngle
	}
	return pw.annularArcPath(x, y, r1, r2, startAngle, endAngle)
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
	return pw.DrawClosedShape(ClosedShape{
		Kind:     ClosedShapePolygon,
		Center:   Location{X: x, Y: y},
		Radius:   r,
		Sides:    sides,
		Rotation: rotation,
		Reverse:  reverse,
	}, border, fill)
}

func (pw *PageWriter) PolygonPath(x, y, r float64, sides int, reverse bool, rotation float64) error {
	if err := pw.requirePathSession(); err != nil {
		return err
	}
	return pw.buildClosedShapePath(ClosedShape{
		Kind:     ClosedShapePolygon,
		Center:   Location{X: x, Y: y},
		Radius:   r,
		Sides:    sides,
		Rotation: rotation,
		Reverse:  reverse,
	})
}

func (pw *PageWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	return pw.DrawClosedShape(ClosedShape{
		Kind:        ClosedShapeStar,
		Center:      Location{X: x, Y: y},
		Radius:      r1,
		InnerRadius: r2,
		Points:      points,
		Rotation:    rotation,
		Reverse:     reverse,
	}, border, fill)
}

func (pw *PageWriter) StarPath(x, y, r1, r2 float64, points int, reverse bool, rotation float64) error {
	if err := pw.requirePathSession(); err != nil {
		return err
	}
	return pw.buildClosedShapePath(ClosedShape{
		Kind:        ClosedShapeStar,
		Center:      Location{X: x, Y: y},
		Radius:      r1,
		InnerRadius: r2,
		Points:      points,
		Rotation:    rotation,
		Reverse:     reverse,
	})
}

func lineCapStyleFromString(style string) (LineCapStyle, bool) {
	switch style {
	case "round_cap":
		return RoundCap, true
	case "projecting_square_cap":
		return ProjectingSquareCap, true
	case "butt_cap":
		return ButtCap, true
	default:
		return ButtCap, false
	}
}

func (pw *PageWriter) drawUnderline(
	loc1 Location,
	loc2 Location,
	rise float64,
	position float64,
	thickness float64,
	lineColor colors.Color,
	hasLineColor bool,
	lineCapStyleName string,
	linePattern string,
	hasLinePattern bool,
) {
	saveWidth := pw.setLineWidth(thickness)
	saveCap := pw.LineCapStyle()
	if lineCapStyle, ok := lineCapStyleFromString(lineCapStyleName); ok {
		pw.SetLineCapStyle(lineCapStyle)
	}
	savePattern := ""
	if hasLinePattern {
		savePattern = pw.SetLineDashPattern(linePattern)
	}
	var saveLineColor colors.Color
	saveLineGradient := pw.lineGradient
	if hasLineColor {
		pw.lineGradient = ""
		saveLineColor = pw.SetLineColor(lineColor)
	}
	// TODO: rotate coordiates given angle
	offsetY := position + rise
	pw.moveTo(loc1.X, loc1.Y+offsetY)
	pw.lineTo(loc2.X, loc2.Y+offsetY)
	if hasLineColor {
		pw.SetLineColor(saveLineColor)
		pw.lineGradient = saveLineGradient
	}
	if hasLinePattern {
		pw.SetLineDashPattern(savePattern)
	}
	pw.SetLineCapStyle(saveCap)
	pw.setLineWidth(saveWidth)
}

func pieceUnderlineStyle(p *rich_text.RichText) (position, thickness float64, lineColor colors.Color, hasLineColor bool, capStyle string, linePattern string, hasLinePattern bool) {
	position = float64(p.UnderlinePosition)
	thickness = float64(p.UnderlineThickness)
	if p.Decoration != nil {
		line := p.Decoration.Underline
		if line.HasPosition {
			position = line.Position
		}
		if line.HasWidth {
			thickness = line.Width
		}
		if line.HasColor {
			lineColor = line.Color
			hasLineColor = true
		}
		if line.CapStyle != "" {
			capStyle = line.CapStyle
		}
		if line.HasPattern {
			linePattern = line.Pattern
			hasLinePattern = true
		}
	}
	return
}

func pieceStrikeoutStyle(p *rich_text.RichText) (position, thickness float64, lineColor colors.Color, hasLineColor bool, capStyle string, linePattern string, hasLinePattern bool) {
	position = float64(p.StrikeoutPosition)
	thickness = float64(p.StrikeoutThickness)
	if p.Decoration != nil {
		line := p.Decoration.Strikeout
		if line.HasPosition {
			position = line.Position
		}
		if line.HasWidth {
			thickness = line.Width
		}
		if line.HasColor {
			lineColor = line.Color
			hasLineColor = true
		}
		if line.CapStyle != "" {
			capStyle = line.CapStyle
		}
		if line.HasPattern {
			linePattern = line.Pattern
			hasLinePattern = true
		}
	}
	return
}

func pieceTextStrokeStyle(p *rich_text.RichText) (lineColor colors.Color, lineWidth float64, hasStroke bool) {
	if p == nil || p.Decoration == nil {
		return 0, 0, false
	}
	stroke := p.Decoration.TextStroke
	if !stroke.HasWidth || stroke.Width <= 0 {
		return 0, 0, false
	}
	lineWidth = stroke.Width
	lineColor = colors.Black
	if stroke.HasColor {
		lineColor = stroke.Color
	}
	return lineColor, lineWidth, true
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
	if !pw.inText {
		return
	}
	pw.tw.close()
	pw.inText = false
}

func (pw *PageWriter) flushText() {
	if pw.line == nil || pw.flushing {
		return
	}
	pw.flushing = true
	pw.startText()
	line := pw.line
	pw.line = nil
	pw.emitRichTextLine(line, pw.visibleTextEmission())
	pw.flushing = false
}

func (pw *PageWriter) visibleTextEmission() textEmissionOptions {
	emit := visibleTextEmission
	if pw.memoCapture {
		emit.emitLinks = false
		emit.emitAccessibility = false
	}
	return emit
}

func (pw *PageWriter) emitRichTextLine(line *rich_text.RichText, emit textEmissionOptions) {
	if line == nil {
		return
	}
	savedFontColor := pw.fontColor
	savedFontKey := pw.fontKey
	savedFontSize := pw.fontSize
	savedLineColor := pw.lineColor
	savedLineWidth := pw.lineWidth
	savedCharSpacing := pw.charSpacing
	savedWordSpacing := pw.wordSpacing
	savedVTextAlign := pw.vTextAlign
	savedVTextAlignPts := pw.vTextAlignPts
	defer func() {
		pw.fontColor = savedFontColor
		pw.fontKey = savedFontKey
		pw.fontSize = savedFontSize
		pw.lineColor = savedLineColor
		pw.lineWidth = savedLineWidth
		pw.charSpacing = savedCharSpacing
		pw.wordSpacing = savedWordSpacing
		pw.vTextAlign = savedVTextAlign
		pw.vTextAlignPts = savedVTextAlignPts
	}()
	pw.startText()
	if pw.loc != pw.last.loc {
		pw.tw.moveBy(pw.loc.X-pw.last.loc.X, pw.loc.Y-pw.last.loc.Y)
	}
	if emit.renderMode != 0 {
		pw.tw.setRenderingMode(emit.renderMode)
	}
	currentRenderMode := emit.renderMode
	loc1 := pw.loc
	textLoc := loc1
	usedPositionedText := false
	var buf bytes.Buffer
	merged := line.Merge()
	displayPieces := bidiDisplayPieces(merged)
	// Iterate leaf pieces directly. TrueType leaves are encoded as big-endian
	// uint16 glyph ID pairs. AFM/Type1 leaves use codepage-based encoding.
	for _, p := range displayPieces {
		leafStart := textLoc
		textLoc.X += p.Width()
		elem, tag := pw.structElemForLeaf(p)
		closeMarkedContent := false
		if emit.emitAccessibility && elem != nil {
			closeMarkedContent = pw.beginTaggedContent(tag, elem)
		}
		closeActualText := false
		renderMode := emit.renderMode
		lineColor, lineWidth, hasTextStroke := pieceTextStrokeStyle(p)
		if emit.applyStrokeState && hasTextStroke {
			renderMode = 2
		}
		if emit.applyStrokeState && hasTextStroke {
			pw.SetLineColor(lineColor)
			pw.checkSetLineColor()
			pw.setLineWidth(lineWidth)
			pw.checkSetLineWidth()
		}
		if renderMode != currentRenderMode {
			pw.tw.setRenderingMode(renderMode)
			currentRenderMode = renderMode
		}
		if p.Font.SubType() == "TrueType" {
			fk := pw.dw.fontKeyUnicode(p.Font)
			psName := p.Font.PostScriptName()
			gr := pw.dw.glyphRecorders[psName]
			buf.Reset()

			var shaped []shaping.GlyphPosition
			var runes []rune // allocated only when shaping is attempted
			var glyphRuneAssignments map[int][]rune
			var glyphEmissionOrder []int
			usePositionedGlyphs := false
			if pw.supportsArabicShaping &&
				p.Font.SupportsArabic() &&
				shaping.ContainsArabic(p.Text) {
				runes = []rune(p.Text)
				var err error
				shaped, err = p.Font.Shaper.Shape(runes, p.Font, float32(p.FontSize))
				if err != nil {
					fmt.Fprintf(os.Stderr, "leadtype: shaping failed during PDF emission for %q (%s): %v\n", p.Text, p.Font.PostScriptName(), err)
				} else {
					glyphRuneAssignments = shapedGlyphRuneAssignments(shaped, runes)
					glyphEmissionOrder = shapedGlyphEmissionOrder(shaped, runes)
				}
			}

			if shaped != nil {
				usePositionedGlyphs = true
			} else {
				usePositionedGlyphs = p.CharSpacing != 0 || p.WordSpacing != 0
			}

			if emit.applyFillColor {
				pw.SetFontColor(p.Color)
				pw.checkSetFontColor()
			}
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

			if shaped != nil {
				// ActualText is replacement text and must always be stored in
				// logical Unicode order. PDF consumers are responsible for bidi
				// presentation; reversing it here merely happens to compensate for
				// one extractor and causes conforming consumers to double-reverse.
				if emit.emitAccessibility {
					closeActualText = pw.beginActualTextContent(p.Text)
				}
				usedPositionedText = true
				fontSize := p.FontSize
				scale := fontSize * 0.001
				if upm := p.Font.UnitsPerEm(); upm > 0 {
					scale = fontSize / float64(upm)
				}
				glyphOrigins := make([]Location, len(shaped))
				visualPenX := 0.0
				for i, gp := range shaped {
					glyphOrigins[i] = Location{
						X: leafStart.X + visualPenX + (float64(gp.XOffset) / 64.0),
						Y: leafStart.Y + (float64(gp.YOffset) / 64.0),
					}
					visualPenX += float64(gp.XAdvance) / 64.0
				}
				if len(glyphEmissionOrder) == 0 {
					glyphEmissionOrder = make([]int, len(shaped))
					for i := range shaped {
						glyphEmissionOrder[i] = i
					}
				}

				// Assign CIDs for all shaped glyphs.
				codes := make([]uint16, len(shaped))
				for i, gp := range shaped {
					code := gp.GlyphID
					if gr != nil {
						if seq := glyphRuneAssignments[i]; len(seq) > 0 {
							code = gr.recordRunes(gp.GlyphID, seq)
						} else {
							code = gr.recordEmpty(gp.GlyphID)
						}
					}
					codes[i] = code
				}

				// Check if all glyphs share the same YOffset (commonly 0).
				// If so, use a single TJ array for better text extraction.
				baseY := shaped[0].YOffset
				allSameY := true
				for _, gp := range shaped[1:] {
					if gp.YOffset != baseY {
						allSameY = false
						break
					}
				}

				if allSameY {
					first := glyphEmissionOrder[0]
					// Emit all glyphs in a single TJ array.
					pw.tw.setMatrix(
						1, 0, 0, 1,
						glyphOrigins[first].X,
						glyphOrigins[first].Y,
					)
					var tjElems []interface{}
					var segment bytes.Buffer
					rawCursor := glyphOrigins[first].X
					flushSegment := func() {
						if segment.Len() == 0 {
							return
						}
						tjElems = append(tjElems, append([]byte(nil), segment.Bytes()...))
						segment.Reset()
					}
					for i, glyphIndex := range glyphEmissionOrder {
						gp := shaped[glyphIndex]
						segment.WriteByte(byte(codes[glyphIndex] >> 8))
						segment.WriteByte(byte(codes[glyphIndex] & 0xFF))

						rawCursor += scale * float64(p.Font.AdvanceWidthForGlyph(gp.GlyphID))
						if i == len(glyphEmissionOrder)-1 {
							continue
						}

						desired := glyphOrigins[glyphEmissionOrder[i+1]].X
						// TJ adjustment: positive moves left, in 1/1000 text units.
						adj := (rawCursor - desired) * 1000.0 / fontSize
						if adj < -0.5 || adj > 0.5 {
							flushSegment()
							tjElems = append(tjElems, adj)
							rawCursor = desired
						}
					}
					flushSegment()
					pw.tw.showHexTJ(tjElems)
				} else {
					// Fall back to per-glyph Tm+Tj for runs with varying YOffset.
					for _, glyphIndex := range glyphEmissionOrder {
						buf.WriteByte(byte(codes[glyphIndex] >> 8))
						buf.WriteByte(byte(codes[glyphIndex] & 0xFF))
						pw.tw.setMatrix(
							1, 0, 0, 1,
							glyphOrigins[glyphIndex].X,
							glyphOrigins[glyphIndex].Y,
						)
						pw.tw.showHex(buf.Bytes())
						buf.Reset()
					}
				}
			} else if usePositionedGlyphs {
				usedPositionedText = true
				penX := 0.0
				fsize := p.FontSize / float64(p.Font.UnitsPerEm())
				for _, r := range p.Text {
					gid := p.Font.GlyphIndex(r)
					code := gid
					if gr != nil {
						code = gr.record(gid, r)
					}
					advanceWidth, _ := p.Font.AdvanceWidth(r)
					buf.WriteByte(byte(code >> 8))
					buf.WriteByte(byte(code & 0xFF))
					pw.tw.setMatrix(1, 0, 0, 1, leafStart.X+penX, leafStart.Y)
					closeAliasText := gr != nil && gr.requiresActualText(code, []rune{r}) && pw.beginActualTextContent(string(r))
					pw.tw.showHex(buf.Bytes())
					if closeAliasText {
						pw.mw.endMarkedContent()
					}
					buf.Reset()
					penX += (fsize * float64(advanceWidth)) + p.CharSpacing
					if r == ' ' {
						penX += p.WordSpacing
					}
				}
			} else {
				for _, r := range p.Text {
					gid := p.Font.GlyphIndex(r)
					code := gid
					if gr != nil {
						code = gr.record(gid, r)
					}
					if gr != nil && gr.requiresActualText(code, []rune{r}) {
						if buf.Len() > 0 {
							pw.tw.showHex(buf.Bytes())
							buf.Reset()
						}
						closeAliasText := pw.beginActualTextContent(string(r))
						pw.tw.showHex([]byte{byte(code >> 8), byte(code & 0xFF)})
						if closeAliasText {
							pw.mw.endMarkedContent()
						}
						continue
					}
					buf.WriteByte(byte(code >> 8))
					buf.WriteByte(byte(code & 0xFF))
				}
				// Composite-font CIDs are arbitrary binary bytes. Emit them as a
				// hex string instead of a literal PDF string so control-byte CIDs
				// do not rely on reader-specific string parsing behavior.
				if buf.Len() > 0 {
					pw.tw.showHex(buf.Bytes())
				}
			}
			if usePositionedGlyphs {
				pw.tw.setMatrix(1, 0, 0, 1, leafStart.X+p.Width(), leafStart.Y)
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
				if emit.applyFillColor {
					pw.SetFontColor(piece.Color)
					pw.checkSetFontColor()
				}
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
		if closeActualText {
			pw.mw.endMarkedContent()
		}
		if closeMarkedContent {
			pw.mw.endMarkedContent()
		}
	}
	if usedPositionedText {
		pw.tw.setMatrix(1, 0, 0, 1, pw.loc.X, pw.loc.Y)
	}
	// Link rectangles are derived from the final laid-out leaf pieces, so a
	// wrapped or bidi-reordered link naturally becomes one annotation per line.
	line.VisitAll(func(p *rich_text.RichText) {
		if !p.IsLeaf() {
			return
		}
		rise := pw.textRiseForPiece(p, savedVTextAlign)
		loc2 := Location{loc1.X + p.Width(), loc1.Y} // TODO: Adjust if print at an angle.
		if emit.emitDecorations && p.Underline {
			position, thickness, lineColor, hasLineColor, capStyle, linePattern, hasLinePattern := pieceUnderlineStyle(p)
			pw.drawUnderline(loc1, loc2, rise, position, thickness, lineColor, hasLineColor, capStyle, linePattern, hasLinePattern)
		}
		if emit.emitDecorations && p.Strikeout {
			position, thickness, lineColor, hasLineColor, capStyle, linePattern, hasLinePattern := pieceStrikeoutStyle(p)
			pw.drawUnderline(loc1, loc2, rise, position, thickness, lineColor, hasLineColor, capStyle, linePattern, hasLinePattern)
		}
		if emit.emitLinks && (p.LinkURI != "" || p.LinkTarget != "") {
			elem, _ := pw.structElemForLeaf(p)
			pw.addTextLinkAnnotation(rectangle{
				x1: loc1.X,
				y1: loc1.Y + rise + p.Descent(),
				x2: loc2.X,
				y2: loc1.Y + rise + p.Ascent(),
			}, p.LinkURI, p.LinkTarget, elem)
		}
		loc1 = loc2
	})
	if currentRenderMode != 0 {
		pw.tw.setRenderingMode(0)
	}
	pw.last.loc = pw.loc
	pw.lineHeight = max(pw.lineHeight, line.Leading()*pw.lineSpacing)
	pw.loc.X += line.Width()
	// TODO: Adjust pw.loc.y if printing at an angle.
}

func (pw *PageWriter) textRiseForPiece(p *rich_text.RichText, vTextAlign VerticalTextAlign) float64 {
	if p == nil || p.Font == nil {
		return 0
	}
	return textRiseForFont(p.Font, p.FontSize, vTextAlign)
}

func textRiseForFont(f *font.Font, fontSize float64, vTextAlign VerticalTextAlign) float64 {
	if f == nil {
		return 0
	}
	scale := fontSize * 0.001
	if upm := f.UnitsPerEm(); upm > 0 {
		scale = fontSize / float64(upm)
	}
	top := float64(f.CapHeight()) * scale
	if top == 0 {
		top = float64(f.Ascent()) * scale
	}
	descent := float64(f.Descent()) * scale
	switch vTextAlign {
	case VTextAlignAbove:
		return -(top - descent)
	case VTextAlignTop:
		return -top
	case VTextAlignMiddle:
		return -((top + descent) / 2.0)
	case VTextAlignBelow:
		return -descent
	default:
		return 0
	}
}

func shapedGlyphRuneAssignments(glyphs []shaping.GlyphPosition, runes []rune) map[int][]rune {
	sortedStarts, clusterSequences, clusterGlyphs := shapedGlyphClusterData(glyphs, runes)
	if len(sortedStarts) == 0 {
		return nil
	}

	assignments := make(map[int][]rune, len(glyphs))
	for _, clusterIdx := range sortedStarts {
		seq := clusterSequences[clusterIdx]
		gidxs := clusterGlyphs[clusterIdx]
		if len(gidxs) == 0 || len(seq) == 0 {
			continue
		}

		if len(gidxs) == 1 {
			// Single glyph for this cluster: assign all runes.
			assignments[gidxs[0]] = seq
		} else if len(gidxs) == len(seq) {
			// Equal number of glyphs and runes: distribute one rune per
			// glyph. Glyphs are in visual (L-to-R) order; runes are in
			// logical order. For RTL text the visual order is the reverse
			// of logical, so assign rune[N-1-j] to glyph[j].
			for j, gi := range gidxs {
				ri := len(seq) - 1 - j
				assignments[gi] = []rune{seq[ri]}
			}
		} else if len(gidxs) > len(seq) {
			// More glyphs than runes: distribute runes to glyphs in
			// reverse order, leave remaining glyphs unassigned.
			for j := 0; j < len(seq) && j < len(gidxs); j++ {
				gi := gidxs[j]
				ri := len(seq) - 1 - j
				assignments[gi] = []rune{seq[ri]}
			}
		} else {
			// More runes than glyphs: assign all runes to the first glyph.
			assignments[gidxs[0]] = seq
		}
	}
	return assignments
}

func shapedGlyphEmissionOrder(glyphs []shaping.GlyphPosition, runes []rune) []int {
	sortedStarts, clusterSequences, clusterGlyphs := shapedGlyphClusterData(glyphs, runes)
	if len(sortedStarts) == 0 {
		return nil
	}

	order := make([]int, 0, len(glyphs))
	seen := make(map[int]struct{}, len(glyphs))
	for _, clusterIdx := range sortedStarts {
		seq := clusterSequences[clusterIdx]
		gidxs := clusterGlyphs[clusterIdx]
		if len(gidxs) == 0 {
			continue
		}

		switch {
		case len(gidxs) == 1:
			order = append(order, gidxs[0])
			seen[gidxs[0]] = struct{}{}
		case len(gidxs) == len(seq):
			for i := len(gidxs) - 1; i >= 0; i-- {
				order = append(order, gidxs[i])
				seen[gidxs[i]] = struct{}{}
			}
		case len(gidxs) > len(seq):
			mapped := len(seq)
			for i := mapped - 1; i >= 0; i-- {
				order = append(order, gidxs[i])
				seen[gidxs[i]] = struct{}{}
			}
			for _, gi := range gidxs[mapped:] {
				order = append(order, gi)
				seen[gi] = struct{}{}
			}
		default:
			order = append(order, gidxs[0])
			seen[gidxs[0]] = struct{}{}
			for _, gi := range gidxs[1:] {
				order = append(order, gi)
				seen[gi] = struct{}{}
			}
		}
	}
	for i := range glyphs {
		if _, ok := seen[i]; ok {
			continue
		}
		order = append(order, i)
	}
	return order
}

func shapedGlyphClusterData(glyphs []shaping.GlyphPosition, runes []rune) ([]int, map[int][]rune, map[int][]int) {
	if len(glyphs) == 0 || len(runes) == 0 {
		return nil, nil, nil
	}

	// Identify cluster boundaries in the rune array.
	clusterStarts := make(map[int]struct{}, len(glyphs))
	for _, gp := range glyphs {
		if gp.ClusterIndex >= 0 && gp.ClusterIndex < len(runes) {
			clusterStarts[gp.ClusterIndex] = struct{}{}
		}
	}
	if len(clusterStarts) == 0 {
		return nil, nil, nil
	}

	sortedStarts := make([]int, 0, len(clusterStarts))
	for start := range clusterStarts {
		sortedStarts = append(sortedStarts, start)
	}
	sort.Ints(sortedStarts)

	// Build the rune sequence for each cluster.
	clusterSequences := make(map[int][]rune, len(sortedStarts))
	for i, start := range sortedStarts {
		end := len(runes)
		if i+1 < len(sortedStarts) {
			end = sortedStarts[i+1]
		}
		if start < end {
			clusterSequences[start] = append([]rune(nil), runes[start:end]...)
		}
	}

	// Group glyph indices by cluster.
	clusterGlyphs := make(map[int][]int, len(clusterSequences))
	for i, gp := range glyphs {
		clusterGlyphs[gp.ClusterIndex] = append(clusterGlyphs[gp.ClusterIndex], i)
	}

	return sortedStarts, clusterSequences, clusterGlyphs
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

func (pw *PageWriter) LineJoinStyle() LineJoinStyle {
	return pw.lineJoinStyle
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

func (pw *PageWriter) MiterLimit() float64 {
	return pw.miterLimit
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

func (pw *PageWriter) MemoizeForm(key string, x, y, width, height float64, render func(*PageWriter) error) error {
	return pw.MemoizeFormOnCanvas(key, x, y, width, height, width, height, render)
}

// MemoizeFormOnCanvas captures a form once on the supplied logical canvas size,
// then places that form at the requested destination size.
func (pw *PageWriter) MemoizeFormOnCanvas(key string, x, y, width, height, canvasWidth, canvasHeight float64, render func(*PageWriter) error) error {
	if key == "" {
		return fmt.Errorf("memo form key must not be empty")
	}
	if render == nil || width <= 0 || height <= 0 || canvasWidth <= 0 || canvasHeight <= 0 {
		return nil
	}

	xPts := pw.units.toPts(x)
	yPts := pw.units.toPts(y)
	widthPts := pw.units.toPts(width)
	heightPts := pw.units.toPts(height)
	canvasWidthPts := pw.units.toPts(canvasWidth)
	canvasHeightPts := pw.units.toPts(canvasHeight)
	cacheKey := pw.memoFormCacheKey(key, canvasWidthPts, canvasHeightPts)
	form, err := pw.dw.loadMemoForm(cacheKey, pw, canvasWidthPts, canvasHeightPts, render)
	if err != nil {
		return err
	}
	return pw.placeFormXObject(
		form.name,
		xPts,
		yPts,
		widthPts,
		heightPts,
		form.width,
		form.height,
		1,
	)
}

func (pw *PageWriter) memoFormCacheKey(key string, widthPts, heightPts float64) string {
	return fmt.Sprintf(
		"memo:%s;w=%s;h=%s;state=%s;opts=%s",
		key,
		g(widthPts),
		g(heightPts),
		pw.memoFormStateKey(),
		memoFormOptionsKey(pw.options, pw.Units()),
	)
}

func (pw *PageWriter) memoFormStateKey() string {
	var b strings.Builder
	fmt.Fprintf(&b, "fillColor=%06x;", int32(pw.fillColor))
	fmt.Fprintf(&b, "fillGradient=%s;", pw.fillGradient)
	fmt.Fprintf(&b, "fontColor=%06x;", int32(pw.fontColor))
	fmt.Fprintf(&b, "fontKey=%s;", pw.fontKey)
	fmt.Fprintf(&b, "fontSize=%s;", g(pw.fontSize))
	fmt.Fprintf(&b, "lineCap=%d;", pw.lineCapStyle)
	fmt.Fprintf(&b, "lineJoin=%d;", pw.lineJoinStyle)
	fmt.Fprintf(&b, "lineColor=%06x;", int32(pw.lineColor))
	fmt.Fprintf(&b, "lineGradient=%s;", pw.lineGradient)
	fmt.Fprintf(&b, "lineDash=%s;", pw.lineDashPattern)
	fmt.Fprintf(&b, "lineSpacing=%s;", g(pw.lineSpacing))
	fmt.Fprintf(&b, "lineWidth=%s;", g(pw.lineWidth))
	fmt.Fprintf(&b, "miter=%s;", g(pw.miterLimit))
	fmt.Fprintf(&b, "strikeout=%t;", pw.strikeout)
	fmt.Fprintf(&b, "underline=%t;", pw.underline)
	fmt.Fprintf(&b, "vTextAlign=%s;", pw.vTextAlign.String())
	fmt.Fprintf(&b, "charSpacing=%s;", g(pw.charSpacing))
	fmt.Fprintf(&b, "wordSpacing=%s;", g(pw.wordSpacing))
	if len(pw.fonts) > 0 {
		b.WriteString("fonts=")
		for _, current := range pw.fonts {
			if current == nil {
				b.WriteString("<nil>,")
				continue
			}
			b.WriteString(current.PostScriptName())
			b.WriteByte(',')
		}
		b.WriteByte(';')
	}
	return b.String()
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

func (pw *PageWriter) paragraphNewLine(next *rich_text.RichText) {
	pw.flushText()
	if next != nil {
		pw.lineHeight = next.Leading() * pw.lineSpacing
	} else if pw.lineHeight == 0 {
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

func (pw *PageWriter) ClipRichText(text *rich_text.RichText, fn func()) error {
	return pw.clipRichTextWithOptions(text, clipTextEmission, fn)
}

func (pw *PageWriter) FillStrokeClipRichText(text *rich_text.RichText, fn func()) error {
	return pw.clipRichTextWithOptions(text, fillStrokeClipTextEmission, fn)
}

func (pw *PageWriter) clipRichTextWithOptions(text *rich_text.RichText, emit textEmissionOptions, fn func()) error {
	if text == nil || text.Len() == 0 {
		return nil
	}
	if len(pw.pathStates) > 0 {
		return errTextClipInsideManualPath
	}
	pw.endText()
	if pw.inGraph {
		pw.endGraph()
	}
	savedLast := pw.last
	pw.gw.saveGraphicsState()
	if emit.applyStrokeState {
		pw.checkSetLineColor()
		pw.checkSetLineWidth()
		pw.checkSetLineDashPattern()
	}
	pw.startText()
	pw.emitRichTextLine(text, emit)
	pw.endText()
	if fn != nil {
		fn()
	}
	pw.endText()
	if pw.inGraph {
		pw.endGraph()
	}
	pw.gw.restoreGraphicsState()
	savedLast.loc = pw.loc
	pw.last = savedLast
	return nil
}

func (pw *PageWriter) ClipText(text string, fn func()) error {
	piece, err := pw.richTextForString(text)
	if err != nil {
		return err
	}
	return pw.ClipRichText(piece, fn)
}

func (pw *PageWriter) FillStrokeClipText(text string, fn func()) error {
	piece, err := pw.richTextForString(text)
	if err != nil {
		return err
	}
	return pw.FillStrokeClipRichText(piece, fn)
}

func (pw *PageWriter) normalizedPaintOpacity(opacity float64) float64 {
	switch {
	case math.IsNaN(opacity):
		return 1
	case opacity <= 0:
		return 0
	case opacity >= 1:
		return 1
	default:
		return opacity
	}
}

func (pw *PageWriter) prepareXObjectDraw() {
	if pw.inPath {
		pw.endPath()
	}
	pw.endText()
	if pw.inGraph {
		pw.endGraph()
	}
}

func (pw *PageWriter) withPlacedXObjectGraphicsState(opacity float64, draw func(extGStateName string)) error {
	pw.prepareXObjectDraw()
	extGStateName := ""
	if opacity < 1 {
		name, err := pw.dw.registerExtGState(opacity, 1, "", nil, "", nil)
		if err != nil {
			return err
		}
		extGStateName = name
	}
	if pw.dw.taggedPDFEnabled() && pw.artifactDepth == 0 {
		if elem := pw.currentStructElem(); elem != nil && pw.beginTaggedContent(elem.s, elem) {
			draw(extGStateName)
			pw.mw.endMarkedContent()
			return nil
		}
	}
	draw(extGStateName)
	return nil
}

func (pw *PageWriter) placeImageXObject(name string, xpts, ypts, wpts, hpts, opacity float64) error {
	return pw.withPlacedXObjectGraphicsState(opacity, func(extGStateName string) {
		writeImageXObject(pw.mw, pw.gw, name, xpts, ypts, wpts, hpts, pw.pageHeight, extGStateName)
	})
}

func (pw *PageWriter) placeFormXObject(name string, xpts, ypts, wpts, hpts, sourceWidth, sourceHeight, opacity float64) error {
	return pw.withPlacedXObjectGraphicsState(opacity, func(extGStateName string) {
		writeFormXObject(pw.mw, pw.gw, name, xpts, ypts, wpts, hpts, sourceWidth, sourceHeight, pw.pageHeight, extGStateName)
	})
}

func (pw *PageWriter) PaintImage(data []byte, x, y, width, height, opacity float64) error {
	opacity = pw.normalizedPaintOpacity(opacity)
	if opacity <= 0 || width <= 0 || height <= 0 {
		return nil
	}
	if svg.LooksLikeSVG(data) {
		return pw.paintSVG(data, x, y, width, height, opacity)
	}
	key := imageKey(data)
	_, name, err := pw.dw.loadImage(data, key)
	if err != nil {
		return err
	}
	return pw.placeImageXObject(
		name,
		pw.units.toPts(x),
		pw.units.toPts(y),
		pw.units.toPts(width),
		pw.units.toPts(height),
		opacity,
	)
}

func (pw *PageWriter) PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	if svg.LooksLikeSVG(data) {
		return pw.PrintSVG(data, x, y, width, height)
	}
	key := imageKey(data)
	image, name, err := pw.dw.loadImage(data, key)
	if err != nil {
		return 0, 0, err
	}
	xpts := pw.units.toPts(x)
	ypts := pw.units.toPts(y)
	info := imageInfo{
		width:            image.width,
		height:           image.height,
		bitsPerComponent: image.bitsPerComponent,
	}
	wpts, hpts := imageSizeInPoints(info, pw.units, width, height)
	if err := pw.placeImageXObject(name, xpts, ypts, wpts, hpts, 1); err != nil {
		return 0, 0, err
	}
	return pw.units.fromPts(wpts), pw.units.fromPts(hpts), nil
}

func (pw *PageWriter) PaintImageFile(filename string, x, y, width, height, opacity float64) error {
	data, err := pw.dw.readImageFile(filename)
	if err != nil {
		return err
	}
	return pw.PaintImage(data, x, y, width, height, opacity)
}

func (pw *PageWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	data, err := pw.dw.readImageFile(filename)
	if err != nil {
		return 0, 0, err
	}
	return pw.PrintImage(data, x, y, width, height)
}

func (pw *PageWriter) paintSVG(data []byte, x, y, width, height, opacity float64) error {
	key := pw.svgFormCacheKey(data)
	form, err := pw.dw.loadSVGForm(data, key, pw.options)
	if err != nil {
		return err
	}
	return pw.placeFormXObject(
		form.name,
		pw.units.toPts(x),
		pw.units.toPts(y),
		pw.units.toPts(width),
		pw.units.toPts(height),
		form.width,
		form.height,
		opacity,
	)
}

func (pw *PageWriter) PrintSVG(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	key := pw.svgFormCacheKey(data)
	form, err := pw.dw.loadSVGForm(data, key, pw.options)
	if err != nil {
		return 0, 0, err
	}
	info := imageInfo{width: int(form.width + 0.5), height: int(form.height + 0.5)}
	wpts, hpts := imageSizeInPoints(info, pw.units, width, height)
	xpts := pw.units.toPts(x)
	ypts := pw.units.toPts(y)
	if err := pw.placeFormXObject(form.name, xpts, ypts, wpts, hpts, form.width, form.height, 1); err != nil {
		return 0, 0, err
	}
	return pw.units.fromPts(wpts), pw.units.fromPts(hpts), nil
}

func (pw *PageWriter) svgFormCacheKey(data []byte) string {
	key := imageKey(data)
	key += ";stop-opacity=" + svgGradientStopOpacityMode(pw.options).String()
	key += ";blend-mode=" + svgBlendMode(pw.options).String()
	return key
}

func (pw *PageWriter) PrintSVGFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	data, err := pw.dw.readImageFile(filename)
	if err != nil {
		return 0, 0, err
	}
	return pw.PrintSVG(data, x, y, width, height)
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
	width := pw.units.toPts(options.FloatDefault("width", pw.units.fromPts(pw.pageWidth-pw.loc.X)))
	for i, p := range para {
		pw.origin = pw.loc
		switch options.StringDefault("text-align", "left") {
		case "center":
			pw.keepOrigin = true
			pw.loc = Location{pw.loc.X + (width-p.Width())/2, pw.loc.Y}
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
		var next *rich_text.RichText
		if i+1 < len(para) {
			next = para[i+1]
		}
		pw.paragraphNewLine(next)
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

var errEmptyLinkURI = errors.New("uri link requires a non-empty URI")
var errEmptyLinkTarget = errors.New("target link requires a non-empty destination name")

func (pw *PageWriter) RegisterDestination(name string, x, y float64) {
	if pw.memoCapture {
		return
	}
	if name == "" {
		return
	}
	xpts := pw.units.toPts(x)
	ypts := pw.translate(pw.units.toPts(y))
	pw.dw.registerDestination(name, pw.page, xpts, ypts)
}

func (pw *PageWriter) AddURILink(x, y, width, height float64, uri string) error {
	if pw.memoCapture {
		return nil
	}
	if uri == "" {
		return errEmptyLinkURI
	}
	var elem *structElem
	if pw.dw.taggedPDFEnabled() && pw.artifactDepth == 0 {
		elem = pw.currentStructElem()
	}
	pw.addTextLinkAnnotation(pw.annotationRect(x, y, width, height), uri, "", elem)
	return nil
}

func (pw *PageWriter) AddTargetLink(x, y, width, height float64, target string) error {
	if pw.memoCapture {
		return nil
	}
	if target == "" {
		return errEmptyLinkTarget
	}
	var elem *structElem
	if pw.dw.taggedPDFEnabled() && pw.artifactDepth == 0 {
		elem = pw.currentStructElem()
	}
	pw.addTextLinkAnnotation(pw.annotationRect(x, y, width, height), "", target, elem)
	return nil
}

func (pw *PageWriter) annotationRect(x, y, width, height float64) rectangle {
	xpts := pw.units.toPts(x)
	wpts := pw.units.toPts(width)
	yTop := pw.translate(pw.units.toPts(y))
	hpts := pw.units.toPts(height)
	return rectangle{x1: xpts, y1: yTop - hpts, x2: xpts + wpts, y2: yTop}
}

func (pw *PageWriter) addTextLinkAnnotation(rect rectangle, uri, target string, elem *structElem) {
	if pw.memoCapture {
		return
	}
	if rect.x2 <= rect.x1 || rect.y2 <= rect.y1 {
		return
	}
	annot := newLinkAnnotation(pw.dw.nextSeq(), 0, rect)
	switch {
	case uri != "":
		annot.setURI(uri)
	case target != "":
		annot.setTarget(target)
		pw.dw.registerPendingTargetLink(annot)
	default:
		return
	}
	pw.dw.file.body.add(annot)
	pw.page.addAnnot(annot)
	if elem != nil && pw.dw.taggedPDFEnabled() {
		pw.dw.accessibility.associateObject(pw.page, annot, elem)
	}
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
	pw.supportsArabicShaping = false
}

func (pw *PageWriter) MeasureText(text string) (metrics TextMetrics, err error) {
	piece, err := pw.richTextForString(text)
	if err != nil {
		return metrics, err
	}
	if piece == nil {
		return metrics, nil
	}
	metrics = TextMetrics{
		Width:   pw.units.fromPts(piece.Width()),
		Height:  pw.units.fromPts(piece.Height()),
		Ascent:  pw.units.fromPts(piece.Ascent()),
		Descent: pw.units.fromPts(piece.Descent()),
	}
	return metrics, nil
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

func (pw *PageWriter) SetFillColor(value any) (prev colors.Color) {
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

// SetFillLinearGradient sets the fill to a linear gradient. Coordinates are
// in the current unit system.
func (pw *PageWriter) SetFillLinearGradient(lg *LinearGradient) error {
	if err := lg.validate(); err != nil {
		return err
	}
	patName, err := pw.dw.registerLinearGradient(lg, pw.units, pw.pageHeight)
	if err != nil {
		return err
	}
	pw.fillGradient = patName
	return nil
}

// SetFillRadialGradient sets the fill to a radial gradient. Coordinates are
// in the current unit system.
func (pw *PageWriter) SetFillRadialGradient(rg *RadialGradient) error {
	if err := rg.validate(); err != nil {
		return err
	}
	patName, err := pw.dw.registerRadialGradient(rg, pw.units, pw.pageHeight)
	if err != nil {
		return err
	}
	pw.fillGradient = patName
	return nil
}

// ClearFillGradient reverts the fill to solid color.
func (pw *PageWriter) ClearFillGradient() {
	pw.fillGradient = ""
}

// SetLineLinearGradient sets the stroke to a linear gradient.
func (pw *PageWriter) SetLineLinearGradient(lg *LinearGradient) error {
	if err := lg.validate(); err != nil {
		return err
	}
	patName, err := pw.dw.registerLinearGradient(lg, pw.units, pw.pageHeight)
	if err != nil {
		return err
	}
	pw.lineGradient = patName
	return nil
}

// SetLineRadialGradient sets the stroke to a radial gradient.
func (pw *PageWriter) SetLineRadialGradient(rg *RadialGradient) error {
	if err := rg.validate(); err != nil {
		return err
	}
	patName, err := pw.dw.registerRadialGradient(rg, pw.units, pw.pageHeight)
	if err != nil {
		return err
	}
	pw.lineGradient = patName
	return nil
}

// ClearLineGradient reverts the stroke to solid color.
func (pw *PageWriter) ClearLineGradient() {
	pw.lineGradient = ""
}

// PaintLinearGradient paints a linear gradient directly into the current
// clipping region using the sh operator.
func (pw *PageWriter) PaintLinearGradient(lg *LinearGradient) error {
	if err := lg.validate(); err != nil {
		return err
	}
	opacity := pw.normalizedGradientOpacity(lg.Opacity)
	if opacity <= 0 {
		return nil
	}
	shName, err := pw.dw.registerLinearShading(lg, pw.units, pw.pageHeight)
	if err != nil {
		return err
	}
	pw.startGraph()
	if opacity < 1 {
		gsName, err := pw.dw.registerExtGState(opacity, 1, "", nil, "", nil)
		if err != nil {
			return err
		}
		pw.gw.saveGraphicsState()
		pw.mw.setExtGState(gsName)
		pw.gw.paintShading(shName)
		pw.gw.restoreGraphicsState()
		return nil
	}
	pw.gw.paintShading(shName)
	return nil
}

// PaintRadialGradient paints a radial gradient directly into the current
// clipping region using the sh operator.
func (pw *PageWriter) PaintRadialGradient(rg *RadialGradient) error {
	if err := rg.validate(); err != nil {
		return err
	}
	opacity := pw.normalizedGradientOpacity(rg.Opacity)
	if opacity <= 0 {
		return nil
	}
	shName, err := pw.dw.registerRadialShading(rg, pw.units, pw.pageHeight)
	if err != nil {
		return err
	}
	pw.startGraph()
	if opacity < 1 {
		gsName, err := pw.dw.registerExtGState(opacity, 1, "", nil, "", nil)
		if err != nil {
			return err
		}
		pw.gw.saveGraphicsState()
		pw.mw.setExtGState(gsName)
		pw.gw.paintShading(shName)
		pw.gw.restoreGraphicsState()
		return nil
	}
	pw.gw.paintShading(shName)
	return nil
}

func (pw *PageWriter) normalizedGradientOpacity(opacity float64) float64 {
	switch {
	case math.IsNaN(opacity):
		return 1
	case opacity < 0:
		return 0
	case opacity == 0:
		return 1
	case opacity > 1:
		return 1
	default:
		return opacity
	}
}

func (pw *PageWriter) SetFontColor(value any) (prev colors.Color) {
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

func (pw *PageWriter) SetLineJoinStyle(lineJoinStyle LineJoinStyle) (prev LineJoinStyle) {
	prev = pw.lineJoinStyle
	pw.lineJoinStyle = lineJoinStyle
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

func (pw *PageWriter) SetMiterLimit(limit float64) (prev float64) {
	prev = pw.miterLimit
	pw.miterLimit = limit
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

func (pw *PageWriter) SetSVGGradientStopOpacityMode(mode SVGGradientStopOpacityMode) (prev SVGGradientStopOpacityMode) {
	prev = svgGradientStopOpacityMode(pw.options)
	if pw.options == nil {
		pw.options = options.Options{}
	}
	pw.options[svgGradientStopOpacityModeOption] = svgGradientStopOpacityMode(options.Options{svgGradientStopOpacityModeOption: mode})
	return prev
}

func (pw *PageWriter) SetSVGBlendMode(mode SVGBlendMode) (prev SVGBlendMode) {
	prev = svgBlendMode(pw.options)
	if pw.options == nil {
		pw.options = options.Options{}
	}
	pw.options[svgBlendModeOption] = svgBlendMode(options.Options{svgBlendModeOption: mode})
	return prev
}

func (pw *PageWriter) SetUnits(units string) {
	pw.units = UnitConversions[units]
}

func (pw *PageWriter) SetVTextAlign(vTextAlign string) (prev string) {
	next := parseVerticalTextAlign(vTextAlign)
	prev = pw.vTextAlign.String()
	if pw.line != nil && pw.vTextAlign != next {
		pw.flushText()
	}
	pw.vTextAlign = next
	return
}

func (pw *PageWriter) startGraph() {
	pw.endText()
	if pw.inGraph {
		return
	}
	pw.last.loc = Location{0, 0}
	pw.inGraph = true
}

func (pw *PageWriter) startText() {
	if pw.inGraph {
		pw.endGraph()
	}
	if pw.inText {
		return
	}
	pw.last.loc = Location{0, 0}
	pw.resetTextStateCache()
	pw.tw.open()
	pw.inText = true
}

func (pw *PageWriter) resetTextStateCache() {
	pw.last.fontKey = ""
	pw.last.fontSize = 0
	pw.last.charSpacing = 0
	pw.last.wordSpacing = 0
	pw.last.vTextAlign = VTextAlignBase
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
	return pw.vTextAlign.String()
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
