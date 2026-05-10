package pdf

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/svg"
)

func logSVGWarnings(warnings []svg.Warning) {
	for _, warning := range warnings {
		msg := warning.Message
		switch {
		case warning.Attribute != "":
			fmt.Fprintf(os.Stderr, "svg: <%s> %s: %s\n", warning.Element, warning.Attribute, msg)
		case warning.Element != "":
			fmt.Fprintf(os.Stderr, "svg: <%s>: %s\n", warning.Element, msg)
		default:
			fmt.Fprintf(os.Stderr, "svg: %s\n", msg)
		}
	}
}

func svgDimensions(data []byte) (width, height int, err error) {
	doc, warnings, err := svg.Parse(data)
	if err != nil {
		return 0, 0, err
	}
	logSVGWarnings(warnings)
	return int(doc.Width + 0.5), int(doc.Height + 0.5), nil
}

func (dw *DocWriter) newSVGForm(doc *svg.Document, renderOptions options.Options) (*renderedForm, error) {
	content := newContentWriter(dw, renderOptions.Merge(options.Options{"units": "pt"}), doc.Width, doc.Height)
	renderer := newSVGRenderer(doc, content, 0, 0, doc.Width, doc.Height)
	if err := renderer.render(); err != nil {
		return nil, err
	}
	content.endText()
	content.endGraph()

	formResources := dw.resources.clone(dw.nextSeq(), 0)
	form := newPDFForm(dw.nextSeq(), 0, content.stream.Bytes())
	form.setBBox(doc.Width, doc.Height)
	form.setResources(formResources)
	if dw.compressPages {
		if err := form.compress(); err != nil {
			return nil, err
		}
	}
	return &renderedForm{form: form, resources: formResources}, nil
}

type svgRenderer struct {
	doc           *svg.Document
	pw            *PageWriter
	x             float64
	y             float64
	width         float64
	height        float64
	scaleX        float64
	scaleY        float64
	viewBox       svg.ViewBox
	mode          svgRenderMode
	maskExtGState map[string]string
}

type svgRenderMode uint8

const (
	svgRenderColor svgRenderMode = iota
	svgRenderMask
)

func newSVGRenderer(doc *svg.Document, pw *PageWriter, x, y, width, height float64) *svgRenderer {
	viewBox := doc.ViewBox
	if viewBox.Width == 0 || viewBox.Height == 0 {
		viewBox = svg.ViewBox{Width: doc.Width, Height: doc.Height}
	}
	return &svgRenderer{
		doc:           doc,
		pw:            pw,
		x:             x,
		y:             y,
		width:         width,
		height:        height,
		scaleX:        width / viewBox.Width,
		scaleY:        height / viewBox.Height,
		viewBox:       viewBox,
		mode:          svgRenderColor,
		maskExtGState: map[string]string{},
	}
}

func (r *svgRenderer) mapPoint(point svg.Point) svg.Point {
	return svg.Point{
		X: r.x + ((point.X - r.viewBox.MinX) * r.scaleX),
		Y: r.y + ((point.Y - r.viewBox.MinY) * r.scaleY),
	}
}

func (r *svgRenderer) mapLengthX(value float64) float64 {
	return value * r.scaleX
}

func (r *svgRenderer) mapLengthY(value float64) float64 {
	return value * r.scaleY
}

func (r *svgRenderer) render() error {
	baseStyle := svg.DefaultStyle()
	for _, child := range r.doc.Children {
		if _, ok := child.(*svg.Defs); ok {
			continue
		}
		if _, ok := child.(*svg.ClipPath); ok {
			continue
		}
		if _, ok := child.(*svg.Mask); ok {
			continue
		}
		if err := r.renderNode(child, baseStyle, svg.IdentityTransform()); err != nil {
			return err
		}
	}
	return nil
}

func (r *svgRenderer) renderNode(node svg.Node, parentStyle svg.Style, parentTransform svg.Transform) error {
	common := node.NodeCommon()
	style := common.Style.Resolve(parentStyle)
	if style.Display == "none" {
		return nil
	}
	if common.FilterRef != "" {
		r.warnIgnoredFilter(node, common.FilterRef)
	}
	transform := parentTransform.Mul(common.Transform)
	return r.withClip(common.ClipPathRef, transform, style, func() error {
		if common.MaskRef != "" {
			return r.withMask(common.MaskRef, func() error {
				return r.renderNodeBody(node, style, transform)
			})
		}
		return r.renderNodeBody(node, style, transform)
	})
}

func (r *svgRenderer) warnIgnoredFilter(node svg.Node, filterRef string) {
	element := svgNodeElementName(node)
	if r.doc.Filters[filterRef] == nil {
		logSVGWarnings([]svg.Warning{{Element: element, Attribute: "filter", Message: fmt.Sprintf("missing filter #%s", filterRef)}})
		return
	}
	logSVGWarnings([]svg.Warning{{Element: element, Attribute: "filter", Message: fmt.Sprintf("filter #%s is parsed but not yet rendered", filterRef)}})
}

func svgNodeElementName(node svg.Node) string {
	switch node.(type) {
	case *svg.Group:
		return "g"
	case *svg.Defs:
		return "defs"
	case *svg.ClipPath:
		return "clipPath"
	case *svg.Mask:
		return "mask"
	case *svg.Use:
		return "use"
	case *svg.Line:
		return "line"
	case *svg.Rect:
		return "rect"
	case *svg.Circle:
		return "circle"
	case *svg.Ellipse:
		return "ellipse"
	case *svg.Polyline:
		return "polyline"
	case *svg.Polygon:
		return "polygon"
	case *svg.Path:
		return "path"
	case *svg.Text:
		return "text"
	default:
		return "svg"
	}
}

func (r *svgRenderer) renderNodeBody(node svg.Node, style svg.Style, transform svg.Transform) error {
	switch n := node.(type) {
	case *svg.Group:
		blendMode := ""
		if n.Style.BlendMode != nil {
			blendMode = mapSVGBlendMode(*n.Style.BlendMode)
		}
		return r.withBlendMode(blendMode, func() error {
			for _, child := range n.Children {
				if err := r.renderNode(child, style, transform); err != nil {
					return err
				}
			}
			return nil
		})
	case *svg.Defs, *svg.ClipPath, *svg.Mask:
		return nil
	case *svg.Use:
		return r.renderUse(n, style, transform)
	case *svg.Line:
		return r.drawSegments([]svg.Segment{
			{Type: svg.SegmentMoveTo, Points: []svg.Point{{X: n.X1, Y: n.Y1}}},
			{Type: svg.SegmentLineTo, Points: []svg.Point{{X: n.X2, Y: n.Y2}}},
		}, style, n.Style, transform)
	case *svg.Rect:
		return r.drawSegments(svgRectPath(n), style, n.Style, transform)
	case *svg.Circle:
		return r.drawSegments(svgEllipsePath(n.CX, n.CY, n.R, n.R), style, n.Style, transform)
	case *svg.Ellipse:
		return r.drawSegments(svgEllipsePath(n.CX, n.CY, n.RX, n.RY), style, n.Style, transform)
	case *svg.Polyline:
		return r.drawSegments(svgPolylinePath(n.Points, false), style, n.Style, transform)
	case *svg.Polygon:
		return r.drawSegments(svgPolylinePath(n.Points, true), style, n.Style, transform)
	case *svg.Path:
		return r.drawSegments(n.Segments, style, n.Style, transform)
	case *svg.Text:
		return r.drawText(n, style, transform)
	default:
		return nil
	}
}

func svgRectPath(n *svg.Rect) []svg.Segment {
	return svgRectSegments(n.X, n.Y, n.Width, n.Height, n.RX, n.RY)
}

func svgPolylinePath(points []svg.Point, close bool) []svg.Segment {
	if len(points) == 0 {
		return nil
	}
	segments := []svg.Segment{{Type: svg.SegmentMoveTo, Points: []svg.Point{points[0]}}}
	for _, point := range points[1:] {
		segments = append(segments, svg.Segment{Type: svg.SegmentLineTo, Points: []svg.Point{point}})
	}
	if close {
		segments = append(segments, svg.Segment{Type: svg.SegmentClosePath})
	}
	return segments
}

func svgEllipsePath(cx, cy, rx, ry float64) []svg.Segment {
	const kappa = 0.5522847498307936
	return []svg.Segment{
		{Type: svg.SegmentMoveTo, Points: []svg.Point{{X: cx + rx, Y: cy}}},
		{Type: svg.SegmentCubicTo, Points: []svg.Point{{X: cx + rx, Y: cy + kappa*ry}, {X: cx + kappa*rx, Y: cy + ry}, {X: cx, Y: cy + ry}}},
		{Type: svg.SegmentCubicTo, Points: []svg.Point{{X: cx - kappa*rx, Y: cy + ry}, {X: cx - rx, Y: cy + kappa*ry}, {X: cx - rx, Y: cy}}},
		{Type: svg.SegmentCubicTo, Points: []svg.Point{{X: cx - rx, Y: cy - kappa*ry}, {X: cx - kappa*rx, Y: cy - ry}, {X: cx, Y: cy - ry}}},
		{Type: svg.SegmentCubicTo, Points: []svg.Point{{X: cx + kappa*rx, Y: cy - ry}, {X: cx + rx, Y: cy - kappa*ry}, {X: cx + rx, Y: cy}}},
		{Type: svg.SegmentClosePath},
	}
}

func svgRectSegments(x, y, width, height, rx, ry float64) []svg.Segment {
	if rx <= 0 && ry <= 0 {
		return []svg.Segment{
			{Type: svg.SegmentMoveTo, Points: []svg.Point{{X: x, Y: y}}},
			{Type: svg.SegmentLineTo, Points: []svg.Point{{X: x + width, Y: y}}},
			{Type: svg.SegmentLineTo, Points: []svg.Point{{X: x + width, Y: y + height}}},
			{Type: svg.SegmentLineTo, Points: []svg.Point{{X: x, Y: y + height}}},
			{Type: svg.SegmentClosePath},
		}
	}
	if rx <= 0 {
		rx = ry
	}
	if ry <= 0 {
		ry = rx
	}
	if rx > width/2 {
		rx = width / 2
	}
	if ry > height/2 {
		ry = height / 2
	}
	kx := rx * 0.5522847498307936
	ky := ry * 0.5522847498307936
	return []svg.Segment{
		{Type: svg.SegmentMoveTo, Points: []svg.Point{{X: x + rx, Y: y}}},
		{Type: svg.SegmentLineTo, Points: []svg.Point{{X: x + width - rx, Y: y}}},
		{Type: svg.SegmentCubicTo, Points: []svg.Point{{X: x + width - rx + kx, Y: y}, {X: x + width, Y: y + ry - ky}, {X: x + width, Y: y + ry}}},
		{Type: svg.SegmentLineTo, Points: []svg.Point{{X: x + width, Y: y + height - ry}}},
		{Type: svg.SegmentCubicTo, Points: []svg.Point{{X: x + width, Y: y + height - ry + ky}, {X: x + width - rx + kx, Y: y + height}, {X: x + width - rx, Y: y + height}}},
		{Type: svg.SegmentLineTo, Points: []svg.Point{{X: x + rx, Y: y + height}}},
		{Type: svg.SegmentCubicTo, Points: []svg.Point{{X: x + rx - kx, Y: y + height}, {X: x, Y: y + height - ry + ky}, {X: x, Y: y + height - ry}}},
		{Type: svg.SegmentLineTo, Points: []svg.Point{{X: x, Y: y + ry}}},
		{Type: svg.SegmentCubicTo, Points: []svg.Point{{X: x, Y: y + ry - ky}, {X: x + rx - kx, Y: y}, {X: x + rx, Y: y}}},
		{Type: svg.SegmentClosePath},
	}
}

func (r *svgRenderer) withClip(clipRef string, parentTransform svg.Transform, style svg.Style, fn func() error) error {
	if clipRef == "" {
		return fn()
	}
	clip := r.doc.ClipPaths[clipRef]
	if clip == nil {
		logSVGWarnings([]svg.Warning{{Element: "clip-path", Message: fmt.Sprintf("missing clipPath #%s", clipRef)}})
		return fn()
	}
	clipStyle := clip.Style.Resolve(style)
	if len(clip.Children) == 1 {
		if textNode, ok := clip.Children[0].(*svg.Text); ok {
			textStyle := textNode.Style.Resolve(clipStyle)
			textTransform := parentTransform.Mul(textNode.NodeCommon().Transform)
			return r.clipWithText(textNode, textStyle, textTransform, fn)
		}
	}
	var renderErr error
	err := r.pw.Path(func() {
		for _, child := range clip.Children {
			segments := r.pathForNode(child)
			if len(segments) == 0 {
				continue
			}
			r.traceSegments(segments, parentTransform.Mul(child.NodeCommon().Transform))
		}
		if strings.EqualFold(clipStyle.FillRule, "evenodd") {
			logSVGWarnings([]svg.Warning{{Element: "clipPath", Attribute: "fill-rule", Message: "evenodd clip falls back to nonzero"}})
		}
		if clipErr := r.pw.Clip(func() {
			renderErr = fn()
		}); clipErr != nil && renderErr == nil {
			renderErr = clipErr
		}
	})
	if err != nil {
		return err
	}
	return renderErr
}

func (r *svgRenderer) clipWithText(text *svg.Text, style svg.Style, transform svg.Transform, fn func() error) error {
	layout, err := r.layoutText(text, style, transform)
	if err != nil || layout == nil {
		return err
	}
	return r.withPreparedTextState(style, func() error {
		r.pw.MoveTo(layout.startX, layout.mappedY)
		var renderErr error
		err := r.pw.ClipRichText(layout.rich, func() {
			if fn != nil {
				renderErr = fn()
			}
		})
		if err != nil {
			return err
		}
		return renderErr
	})
}

func (r *svgRenderer) pathForNode(node svg.Node) []svg.Segment {
	switch n := node.(type) {
	case *svg.Line:
		return []svg.Segment{
			{Type: svg.SegmentMoveTo, Points: []svg.Point{{X: n.X1, Y: n.Y1}}},
			{Type: svg.SegmentLineTo, Points: []svg.Point{{X: n.X2, Y: n.Y2}}},
		}
	case *svg.Rect:
		return svgRectPath(n)
	case *svg.Circle:
		return svgEllipsePath(n.CX, n.CY, n.R, n.R)
	case *svg.Ellipse:
		return svgEllipsePath(n.CX, n.CY, n.RX, n.RY)
	case *svg.Polyline:
		return svgPolylinePath(n.Points, false)
	case *svg.Polygon:
		return svgPolylinePath(n.Points, true)
	case *svg.Path:
		return n.Segments
	case *svg.Use:
		referenced := r.doc.Refs[n.Href]
		if referenced == nil {
			return nil
		}
		segments := r.pathForNode(referenced)
		return translateSegments(segments, n.X, n.Y)
	default:
		return nil
	}
}

func translateSegments(segments []svg.Segment, dx, dy float64) []svg.Segment {
	if len(segments) == 0 || (dx == 0 && dy == 0) {
		return segments
	}
	out := make([]svg.Segment, len(segments))
	for i, segment := range segments {
		out[i] = segment
		if len(segment.Points) == 0 {
			continue
		}
		out[i].Points = make([]svg.Point, len(segment.Points))
		for j, point := range segment.Points {
			out[i].Points[j] = svg.Point{X: point.X + dx, Y: point.Y + dy}
		}
	}
	return out
}

func (r *svgRenderer) drawSegments(segments []svg.Segment, style svg.Style, styleSpec svg.StyleSpec, transform svg.Transform) error {
	if len(segments) == 0 {
		return nil
	}
	fill := style.Fill.Set && !style.Fill.None
	stroke := style.Stroke.Set && !style.Stroke.None && style.StrokeWidth > 0
	if !fill && !stroke {
		return nil
	}
	if r.mode == svgRenderColor &&
		!style.Fill.IsGradient() &&
		!style.Stroke.IsGradient() &&
		style.Opacity >= 0.9999 &&
		style.FillOpacity >= 0.9999 &&
		style.StrokeOpacity >= 0.9999 &&
		styleSpec.BlendMode == nil {
		var err error
		err = r.pw.Path(func() {
			r.traceSegments(segments, transform)
			switch {
			case fill && stroke:
				err = r.finishPath(style, true, true)
			case fill:
				err = r.finishPath(style, true, false)
			case stroke:
				err = r.finishPath(style, false, true)
			}
		})
		return err
	}
	if fill {
		if err := r.paintSegments(segments, style, styleSpec, transform, true); err != nil {
			return err
		}
	}
	if stroke {
		if err := r.paintSegments(segments, style, styleSpec, transform, false); err != nil {
			return err
		}
	}
	return nil
}

type svgBounds struct {
	MinX  float64
	MinY  float64
	MaxX  float64
	MaxY  float64
	Valid bool
}

func (b svgBounds) Width() float64 {
	if !b.Valid {
		return 0
	}
	return b.MaxX - b.MinX
}

func (b svgBounds) Height() float64 {
	if !b.Valid {
		return 0
	}
	return b.MaxY - b.MinY
}

type resolvedSVGGradient struct {
	linear       *LinearGradient
	radial       *RadialGradient
	colorStops   []GradientStop
	alphaStops   []alphaGradientStop
	maskStops    []alphaGradientStop
	uniformAlpha float64
	flatAlpha    float64
	varyingAlpha bool
	firstColor   colors.Color
}

func (r *svgRenderer) renderUse(node *svg.Use, style svg.Style, transform svg.Transform) error {
	referenced := r.doc.Refs[node.Href]
	if referenced == nil {
		logSVGWarnings([]svg.Warning{{Element: "use", Attribute: "href", Message: fmt.Sprintf("missing reference #%s", node.Href)}})
		return nil
	}
	useTransform := transform.Mul(svg.Transform{A: 1, D: 1, E: node.X, F: node.Y})
	return r.renderNode(referenced, style, useTransform)
}

func (r *svgRenderer) withBlendMode(blendMode string, fn func() error) error {
	if blendMode == "" || blendMode == "Normal" {
		return fn()
	}
	return r.withScopedGraphicsState(1, 1, blendMode, nil, "", nil, fn)
}

func (r *svgRenderer) withMask(maskRef string, fn func() error) error {
	name, err := r.maskExtGStateName(maskRef)
	if err != nil || name == "" {
		return err
	}
	savedLast := r.pw.last
	r.pw.startGraph()
	r.pw.gw.saveGraphicsState()
	r.pw.mw.setExtGState(name)
	if fn != nil {
		err = fn()
	}
	r.pw.endText()
	if r.pw.inGraph {
		r.pw.endGraph()
	}
	r.pw.gw.restoreGraphicsState()
	r.pw.last = savedLast
	return err
}

func (r *svgRenderer) withScopedGraphicsState(fillAlpha, strokeAlpha float64, blendMode string, maskForm seqGen, maskSubtype string, maskBackdrop []float64, fn func() error) error {
	if fillAlpha >= 1 && strokeAlpha >= 1 && (blendMode == "" || blendMode == "Normal") && maskForm == nil {
		return fn()
	}
	name, err := r.pw.dw.registerExtGState(fillAlpha, strokeAlpha, blendMode, maskForm, maskSubtype, maskBackdrop)
	if err != nil {
		return err
	}
	savedLast := r.pw.last
	r.pw.startGraph()
	r.pw.gw.saveGraphicsState()
	r.pw.mw.setExtGState(name)
	if fn != nil {
		err = fn()
	}
	r.pw.endText()
	if r.pw.inGraph {
		r.pw.endGraph()
	}
	r.pw.gw.restoreGraphicsState()
	r.pw.last = savedLast
	return err
}

func clipPathWithRule(pw *PageWriter, rule string, fn func() error) error {
	if err := pw.requireActivePath(); err != nil {
		return err
	}
	savedLast := pw.last
	pw.startGraph()
	pw.gw.saveGraphicsState()
	if strings.EqualFold(rule, "evenodd") {
		pw.gw.eoClip()
	} else {
		pw.gw.clip()
	}
	pw.gw.newPath()
	pw.inPath = false
	pw.restorePathState()
	var err error
	if fn != nil {
		err = fn()
	}
	pw.endText()
	if pw.inGraph {
		pw.endGraph()
	}
	pw.gw.restoreGraphicsState()
	pw.last = savedLast
	return err
}

func (r *svgRenderer) maskExtGStateName(maskRef string) (string, error) {
	if name := r.maskExtGState[maskRef]; name != "" {
		return name, nil
	}
	mask := r.doc.Masks[maskRef]
	if mask == nil {
		logSVGWarnings([]svg.Warning{{Element: "mask", Message: fmt.Sprintf("missing mask #%s", maskRef)}})
		return "", nil
	}
	if mask.Units != "" && !strings.EqualFold(mask.Units, "userSpaceOnUse") {
		logSVGWarnings([]svg.Warning{{Element: "mask", Attribute: "maskUnits", Message: "only userSpaceOnUse masks are supported"}})
		return "", nil
	}
	form, _, err := r.renderMaskForm(mask)
	if err != nil {
		return "", err
	}
	name, err := r.pw.dw.registerExtGState(1, 1, "", form, "Luminosity", []float64{0})
	if err != nil {
		return "", err
	}
	r.maskExtGState[maskRef] = name
	return name, nil
}

func (r *svgRenderer) renderMaskForm(mask *svg.Mask) (*pdfForm, *resources, error) {
	content := newContentWriter(r.pw.dw, options.Options{"units": "pt"}, r.doc.Width, r.doc.Height)
	maskRenderer := newSVGRenderer(r.doc, content, 0, 0, r.doc.Width, r.doc.Height)
	maskRenderer.mode = svgRenderMask
	for _, child := range mask.Children {
		if err := maskRenderer.renderNode(child, svg.DefaultStyle(), svg.IdentityTransform()); err != nil {
			return nil, nil, err
		}
	}
	content.endText()
	content.endGraph()

	formResources := r.pw.dw.resources.clone(r.pw.dw.nextSeq(), 0)
	form := newPDFForm(r.pw.dw.nextSeq(), 0, content.stream.Bytes())
	form.setBBox(r.doc.Width, r.doc.Height)
	form.setResources(formResources)
	form.setTransparencyGroup("DeviceGray")
	if r.pw.dw.compressPages {
		if err := form.compress(); err != nil {
			return nil, nil, err
		}
	}
	r.pw.dw.file.body.add(formResources, form)
	return form, formResources, nil
}

func (r *svgRenderer) paintSegments(segments []svg.Segment, style svg.Style, styleSpec svg.StyleSpec, transform svg.Transform, fill bool) error {
	var err error
	err = r.pw.Path(func() {
		r.traceSegments(segments, transform)
		if r.mode == svgRenderMask {
			err = r.paintMaskPath(segments, style, transform, fill)
			return
		}
		err = r.paintColorPath(segments, style, styleSpec, transform, fill)
	})
	return err
}

func (r *svgRenderer) paintColorPath(segments []svg.Segment, style svg.Style, styleSpec svg.StyleSpec, transform svg.Transform, fill bool) error {
	blendMode := ""
	if styleSpec.BlendMode != nil {
		blendMode = mapSVGBlendMode(*styleSpec.BlendMode)
	}
	alpha := clamp01(style.Opacity)
	if fill {
		alpha = clamp01(alpha * style.FillOpacity)
		if style.Fill.IsGradient() {
			resolved, err := r.resolveGradient(style.Fill.Ref, alpha, segments, transform)
			if err != nil {
				logSVGWarnings([]svg.Warning{{Element: "linearGradient", Message: err.Error()}})
				return r.fillSolidPath(style.Fill.Color, alpha, style.FillRule, blendMode)
			}
			if resolved.varyingAlpha {
				if svgGradientStopOpacityMode(r.pw.options) == svgGradientStopOpacityModeFlat {
					return r.withScopedGraphicsState(resolved.flatAlpha, 1, blendMode, nil, "", nil, func() error {
						return r.fillGradientPath(style, resolved)
					})
				}
				maskForm, err := r.makeGradientAlphaMaskForm(segments, transform, style.FillRule, resolved)
				if err != nil {
					return err
				}
				return r.withScopedGraphicsState(1, 1, blendMode, maskForm, "Luminosity", []float64{0, 0, 0}, func() error {
					return r.fillGradientByShading(style.FillRule, resolved)
				})
			}
			return r.withScopedGraphicsState(resolved.uniformAlpha, 1, blendMode, nil, "", nil, func() error {
				return r.fillGradientPath(style, resolved)
			})
		}
		return r.fillSolidPath(style.Fill.Color, alpha, style.FillRule, blendMode)
	}

	alpha = clamp01(alpha * style.StrokeOpacity)
	if style.Stroke.IsGradient() {
		resolved, err := r.resolveGradient(style.Stroke.Ref, alpha, segments, transform)
		if err != nil {
			logSVGWarnings([]svg.Warning{{Element: "linearGradient", Message: err.Error()}})
			return r.strokeSolidPath(style, style.Stroke.Color, alpha, blendMode)
		}
		if resolved.varyingAlpha {
			if svgGradientStopOpacityMode(r.pw.options) == svgGradientStopOpacityModeFlat {
				alpha = resolved.flatAlpha
			} else {
				logSVGWarnings([]svg.Warning{{Element: "style", Message: "gradient stroke stop-opacity falls back to constant stroke opacity"}})
				alpha = resolved.uniformAlpha
			}
		} else {
			alpha = resolved.uniformAlpha
		}
		return r.withScopedGraphicsState(1, alpha, blendMode, nil, "", nil, func() error {
			return r.strokeGradientPath(style, resolved)
		})
	}
	return r.strokeSolidPath(style, style.Stroke.Color, alpha, blendMode)
}

func (r *svgRenderer) paintMaskPath(segments []svg.Segment, style svg.Style, transform svg.Transform, fill bool) error {
	if fill {
		if style.Fill.IsGradient() {
			resolved, err := r.resolveGradient(style.Fill.Ref, clamp01(style.Opacity*style.FillOpacity), segments, transform)
			if err != nil {
				logSVGWarnings([]svg.Warning{{Element: "mask", Message: err.Error()}})
				return nil
			}
			return r.fillMaskGradientPath(style.FillRule, resolved)
		}
		gray := colorGray(style.Fill.Color) * clamp01(style.Opacity*style.FillOpacity)
		return r.fillSolidMaskPath(gray, style.FillRule)
	}
	if style.Stroke.IsGradient() {
		logSVGWarnings([]svg.Warning{{Element: "mask", Message: "gradient strokes in masks fall back to the first stop color"}})
	}
	gray := colorGray(style.Stroke.Color) * clamp01(style.Opacity*style.StrokeOpacity)
	return r.strokeSolidMaskPath(style, gray)
}

func (r *svgRenderer) fillGradientPath(style svg.Style, resolved *resolvedSVGGradient) error {
	if resolved == nil {
		return nil
	}
	prevFill := r.pw.fillGradient
	defer func() {
		r.pw.fillGradient = prevFill
	}()
	switch {
	case resolved.linear != nil:
		if err := r.pw.SetFillLinearGradient(resolved.linear); err != nil {
			return err
		}
	case resolved.radial != nil:
		if err := r.pw.SetFillRadialGradient(resolved.radial); err != nil {
			return err
		}
	}
	return fillWithRule(r.pw, style.FillRule)
}

func (r *svgRenderer) fillGradientByShading(rule string, resolved *resolvedSVGGradient) error {
	if resolved == nil {
		return nil
	}
	return clipPathWithRule(r.pw, rule, func() error {
		switch {
		case resolved.linear != nil:
			return r.pw.PaintLinearGradient(resolved.linear)
		case resolved.radial != nil:
			return r.pw.PaintRadialGradient(resolved.radial)
		default:
			return nil
		}
	})
}

func (r *svgRenderer) strokeGradientPath(style svg.Style, resolved *resolvedSVGGradient) error {
	if resolved == nil {
		return nil
	}
	prevGradient := r.pw.lineGradient
	prevWidth := r.pw.setLineWidth(r.mapLengthX(style.StrokeWidth))
	prevCap := r.pw.SetLineCapStyle(mapSVGLineCap(style.LineCap))
	prevJoin := r.pw.SetLineJoinStyle(mapSVGLineJoin(style.LineJoin))
	prevMiter := r.pw.SetMiterLimit(style.MiterLimit)
	prevDash := r.pw.SetLineDashPattern(dashPatternForStyle(style, r.scaleX))
	defer func() {
		r.pw.lineGradient = prevGradient
		r.pw.setLineWidth(prevWidth)
		r.pw.SetLineCapStyle(prevCap)
		r.pw.SetLineJoinStyle(prevJoin)
		r.pw.SetMiterLimit(prevMiter)
		r.pw.SetLineDashPattern(prevDash)
	}()
	switch {
	case resolved.linear != nil:
		if err := r.pw.SetLineLinearGradient(resolved.linear); err != nil {
			return err
		}
	case resolved.radial != nil:
		if err := r.pw.SetLineRadialGradient(resolved.radial); err != nil {
			return err
		}
	}
	return r.pw.Stroke()
}

func (r *svgRenderer) fillSolidPath(color colors.Color, alpha float64, rule string, blendMode string) error {
	return r.withScopedGraphicsState(alpha, 1, blendMode, nil, "", nil, func() error {
		prevFill := r.pw.SetFillColor(color)
		defer r.pw.SetFillColor(prevFill)
		return fillWithRule(r.pw, rule)
	})
}

func (r *svgRenderer) strokeSolidPath(style svg.Style, color colors.Color, alpha float64, blendMode string) error {
	return r.withScopedGraphicsState(1, alpha, blendMode, nil, "", nil, func() error {
		prevStroke := r.pw.SetLineColor(color)
		prevWidth := r.pw.setLineWidth(r.mapLengthX(style.StrokeWidth))
		prevCap := r.pw.SetLineCapStyle(mapSVGLineCap(style.LineCap))
		prevJoin := r.pw.SetLineJoinStyle(mapSVGLineJoin(style.LineJoin))
		prevMiter := r.pw.SetMiterLimit(style.MiterLimit)
		prevDash := r.pw.SetLineDashPattern(dashPatternForStyle(style, r.scaleX))
		defer func() {
			r.pw.SetLineColor(prevStroke)
			r.pw.setLineWidth(prevWidth)
			r.pw.SetLineCapStyle(prevCap)
			r.pw.SetLineJoinStyle(prevJoin)
			r.pw.SetMiterLimit(prevMiter)
			r.pw.SetLineDashPattern(prevDash)
		}()
		return r.pw.Stroke()
	})
}

func (r *svgRenderer) fillSolidMaskPath(gray float64, rule string) error {
	prevFill := r.pw.SetFillColor(colors.Color(0))
	defer r.pw.SetFillColor(prevFill)
	r.pw.mw.setGrayFill(clamp01(gray))
	r.pw.last.fillColor = 0
	r.pw.last.fillGradient = ""
	return fillWithRule(r.pw, rule)
}

func (r *svgRenderer) strokeSolidMaskPath(style svg.Style, gray float64) error {
	prevStroke := r.pw.SetLineColor(colors.Color(0))
	prevWidth := r.pw.setLineWidth(r.mapLengthX(style.StrokeWidth))
	prevCap := r.pw.SetLineCapStyle(mapSVGLineCap(style.LineCap))
	prevJoin := r.pw.SetLineJoinStyle(mapSVGLineJoin(style.LineJoin))
	prevMiter := r.pw.SetMiterLimit(style.MiterLimit)
	prevDash := r.pw.SetLineDashPattern(dashPatternForStyle(style, r.scaleX))
	defer func() {
		r.pw.SetLineColor(prevStroke)
		r.pw.setLineWidth(prevWidth)
		r.pw.SetLineCapStyle(prevCap)
		r.pw.SetLineJoinStyle(prevJoin)
		r.pw.SetMiterLimit(prevMiter)
		r.pw.SetLineDashPattern(prevDash)
	}()
	r.pw.mw.setGrayStroke(clamp01(gray))
	r.pw.last.lineColor = 0
	r.pw.last.lineGradient = ""
	return r.pw.Stroke()
}

func (r *svgRenderer) fillMaskGradientPath(rule string, resolved *resolvedSVGGradient) error {
	if resolved == nil || resolved.linear == nil {
		return r.fillSolidMaskPath(0, rule)
	}
	shName, err := r.pw.dw.registerLinearAlphaShading(resolved.linear, resolved.maskStops, r.pw.units, r.pw.pageHeight)
	if err != nil {
		return err
	}
	var renderErr error
	if err := r.pw.Clip(func() {
		r.pw.startGraph()
		r.pw.gw.paintShading(shName)
	}); err != nil {
		return err
	}
	return renderErr
}

func (r *svgRenderer) makeGradientAlphaMaskForm(segments []svg.Segment, transform svg.Transform, rule string, resolved *resolvedSVGGradient) (*pdfForm, error) {
	content := newContentWriter(r.pw.dw, options.Options{"units": "pt"}, r.doc.Width, r.doc.Height)
	maskRenderer := newSVGRenderer(r.doc, content, 0, 0, r.doc.Width, r.doc.Height)
	var renderErr error
	err := content.Path(func() {
		maskRenderer.traceSegments(segments, transform)
		if strings.EqualFold(rule, "evenodd") {
			logSVGWarnings([]svg.Warning{{Element: "style", Attribute: "fill-rule", Message: "evenodd alpha mask falls back to nonzero"}})
		}
		if clipErr := content.Clip(func() {
			maskGradient := alphaMaskGradient(resolved)
			if maskGradient == nil {
				return
			}
			var (
				shName string
				shErr  error
			)
			switch gradient := maskGradient.(type) {
			case *LinearGradient:
				shName, shErr = content.dw.registerLinearShading(gradient, content.units, content.pageHeight)
			case *RadialGradient:
				shName, shErr = content.dw.registerRadialShading(gradient, content.units, content.pageHeight)
			}
			if shErr != nil {
				renderErr = shErr
				return
			}
			content.startGraph()
			content.gw.paintShading(shName)
		}); clipErr != nil && renderErr == nil {
			renderErr = clipErr
		}
	})
	if err != nil {
		return nil, err
	}
	if renderErr != nil {
		return nil, renderErr
	}
	content.endText()
	content.endGraph()
	formResources := r.pw.dw.resources.clone(r.pw.dw.nextSeq(), 0)
	form := newPDFForm(r.pw.dw.nextSeq(), 0, content.stream.Bytes())
	form.setBBox(r.doc.Width, r.doc.Height)
	form.setResources(formResources)
	form.setTransparencyGroup("DeviceRGB")
	if r.pw.dw.compressPages {
		if err := form.compress(); err != nil {
			return nil, err
		}
	}
	r.pw.dw.file.body.add(formResources, form)
	return form, nil
}

func (r *svgRenderer) resolveGradient(ref string, opacityScale float64, segments []svg.Segment, transform svg.Transform) (*resolvedSVGGradient, error) {
	def := r.doc.Gradients[ref]
	if def == nil {
		return nil, fmt.Errorf("missing gradient #%s", ref)
	}
	bounds := pathBounds(segments)
	if !bounds.Valid {
		return nil, fmt.Errorf("gradient target has no bounds")
	}
	resolved := &resolvedSVGGradient{
		colorStops: make([]GradientStop, len(def.Stops)),
		alphaStops: make([]alphaGradientStop, len(def.Stops)),
		maskStops:  make([]alphaGradientStop, len(def.Stops)),
	}
	switch def.Kind {
	case svg.GradientLinear:
		x0 := r.resolveGradientLength(def.Linear.X1, bounds.MinX, bounds.Width(), true, def.Units)
		y0 := r.resolveGradientLength(def.Linear.Y1, bounds.MinY, bounds.Height(), false, def.Units)
		x1 := r.resolveGradientLength(def.Linear.X2, bounds.MinX, bounds.Width(), true, def.Units)
		y1 := r.resolveGradientLength(def.Linear.Y2, bounds.MinY, bounds.Height(), false, def.Units)
		p0 := transform.Apply(def.Transform.Apply(svg.Point{X: x0, Y: y0}))
		p1 := transform.Apply(def.Transform.Apply(svg.Point{X: x1, Y: y1}))
		mp0 := r.mapPoint(p0)
		mp1 := r.mapPoint(p1)
		resolved.linear = &LinearGradient{X0: mp0.X, Y0: mp0.Y, X1: mp1.X, Y1: mp1.Y}
	case svg.GradientRadial:
		radial, err := r.resolveRadialGradient(def, bounds, transform)
		if err != nil {
			return nil, err
		}
		resolved.radial = radial
	}
	alphaMin := -1.0
	alphaSum := 0.0
	for i, stop := range def.Stops {
		resolved.colorStops[i] = GradientStop{Position: stop.Position, Color: stop.Color}
		alpha := clamp01(opacityScale * stop.Opacity)
		alphaSum += alpha
		resolved.alphaStops[i] = alphaGradientStop{Position: stop.Position, Alpha: alpha}
		resolved.maskStops[i] = alphaGradientStop{Position: stop.Position, Alpha: clamp01(colorGray(stop.Color) * alpha)}
		if i == 0 {
			resolved.firstColor = stop.Color
			alphaMin = alpha
		} else if math.Abs(alpha-alphaMin) > 0.0001 {
			resolved.varyingAlpha = true
		}
	}
	if resolved.linear != nil {
		resolved.linear.Stops = resolved.colorStops
	}
	if resolved.radial != nil {
		resolved.radial.Stops = resolved.colorStops
	}
	if len(resolved.alphaStops) > 0 {
		resolved.uniformAlpha = resolved.alphaStops[0].Alpha
		resolved.flatAlpha = clamp01(alphaSum / float64(len(resolved.alphaStops)))
	}
	return resolved, nil
}

func (r *svgRenderer) resolveRadialGradient(def *svg.Gradient, bounds svgBounds, transform svg.Transform) (*RadialGradient, error) {
	cx0 := r.resolveGradientLength(def.Radial.CX, bounds.MinX, bounds.Width(), true, def.Units)
	cy0 := r.resolveGradientLength(def.Radial.CY, bounds.MinY, bounds.Height(), false, def.Units)
	cx1 := r.resolveGradientLength(def.Radial.FX, bounds.MinX, bounds.Width(), true, def.Units)
	cy1 := r.resolveGradientLength(def.Radial.FY, bounds.MinY, bounds.Height(), false, def.Units)
	r0 := r.resolveGradientRadius(def.Radial.FR, bounds, def.Units)
	r1 := r.resolveGradientRadius(def.Radial.R, bounds, def.Units)
	center0 := r.mapPoint(transform.Apply(def.Transform.Apply(svg.Point{X: cx0, Y: cy0})))
	center1 := r.mapPoint(transform.Apply(def.Transform.Apply(svg.Point{X: cx1, Y: cy1})))
	rx0 := r.mappedRadius(def.Transform, transform, cx0, cy0, r0, true)
	ry0 := r.mappedRadius(def.Transform, transform, cx0, cy0, r0, false)
	rx1 := r.mappedRadius(def.Transform, transform, cx1, cy1, r1, true)
	ry1 := r.mappedRadius(def.Transform, transform, cx1, cy1, r1, false)
	if math.Abs(rx0-ry0) > 0.01 || math.Abs(rx1-ry1) > 0.01 {
		return nil, fmt.Errorf("radial gradient becomes elliptical under the current transform")
	}
	return &RadialGradient{
		X0: center0.X, Y0: center0.Y, R0: (rx0 + ry0) / 2,
		X1: center1.X, Y1: center1.Y, R1: (rx1 + ry1) / 2,
	}, nil
}

func (r *svgRenderer) mappedRadius(gradientTransform, transform svg.Transform, cx, cy, radius float64, xAxis bool) float64 {
	if radius == 0 {
		return 0
	}
	p0 := r.mapPoint(transform.Apply(gradientTransform.Apply(svg.Point{X: cx, Y: cy})))
	p1 := svg.Point{X: cx, Y: cy + radius}
	if xAxis {
		p1 = svg.Point{X: cx + radius, Y: cy}
	}
	p1 = r.mapPoint(transform.Apply(gradientTransform.Apply(p1)))
	return math.Hypot(p1.X-p0.X, p1.Y-p0.Y)
}

func (r *svgRenderer) resolveGradientLength(length svg.Length, origin, size float64, horizontal bool, units string) float64 {
	if strings.EqualFold(units, "userSpaceOnUse") {
		if length.Unit == svg.LengthPercent {
			if horizontal {
				return r.viewBox.MinX + (r.viewBox.Width * length.Value / 100.0)
			}
			return r.viewBox.MinY + (r.viewBox.Height * length.Value / 100.0)
		}
		return length.Value
	}
	if length.Unit == svg.LengthPercent {
		return origin + (size * length.Value / 100.0)
	}
	return origin + (size * length.Value)
}

func (r *svgRenderer) resolveGradientRadius(length svg.Length, bounds svgBounds, units string) float64 {
	if strings.EqualFold(units, "userSpaceOnUse") {
		if length.Unit == svg.LengthPercent {
			return min(r.viewBox.Width, r.viewBox.Height) * length.Value / 100.0
		}
		return length.Value
	}
	ref := min(bounds.Width(), bounds.Height())
	if length.Unit == svg.LengthPercent {
		return ref * length.Value / 100.0
	}
	return ref * length.Value
}

func pathBounds(segments []svg.Segment) svgBounds {
	var bounds svgBounds
	extend := func(point svg.Point) {
		if !bounds.Valid {
			bounds = svgBounds{MinX: point.X, MinY: point.Y, MaxX: point.X, MaxY: point.Y, Valid: true}
			return
		}
		bounds.MinX = math.Min(bounds.MinX, point.X)
		bounds.MinY = math.Min(bounds.MinY, point.Y)
		bounds.MaxX = math.Max(bounds.MaxX, point.X)
		bounds.MaxY = math.Max(bounds.MaxY, point.Y)
	}
	for _, segment := range segments {
		for _, point := range segment.Points {
			extend(point)
		}
	}
	return bounds
}

func colorGray(color colors.Color) float64 {
	r, g, b := color.RGB64()
	return clamp01((0.299 * r) + (0.587 * g) + (0.114 * b))
}

func grayColorValue(value float64) colors.Color {
	level := int(math.Round(clamp01(value) * 255))
	return colors.Color((level << 16) | (level << 8) | level)
}

func alphaMaskGradient(resolved *resolvedSVGGradient) interface{} {
	if resolved == nil {
		return nil
	}
	stops := make([]GradientStop, len(resolved.alphaStops))
	for i, stop := range resolved.alphaStops {
		stops[i] = GradientStop{Position: stop.Position, Color: grayColorValue(stop.Alpha)}
	}
	switch {
	case resolved.linear != nil:
		return &LinearGradient{
			X0:    resolved.linear.X0,
			Y0:    resolved.linear.Y0,
			X1:    resolved.linear.X1,
			Y1:    resolved.linear.Y1,
			Stops: stops,
		}
	case resolved.radial != nil:
		return &RadialGradient{
			X0:    resolved.radial.X0,
			Y0:    resolved.radial.Y0,
			R0:    resolved.radial.R0,
			X1:    resolved.radial.X1,
			Y1:    resolved.radial.Y1,
			R1:    resolved.radial.R1,
			Stops: stops,
		}
	default:
		return nil
	}
}

func clamp01(value float64) float64 {
	return math.Max(0, math.Min(1, value))
}

func (r *svgRenderer) traceSegments(segments []svg.Segment, transform svg.Transform) {
	var start svg.Point
	var started bool
	for _, segment := range segments {
		switch segment.Type {
		case svg.SegmentMoveTo:
			p := r.mapPoint(transform.Apply(segment.Points[0]))
			r.pw.MoveTo(p.X, p.Y)
			start = p
			started = true
		case svg.SegmentLineTo:
			p := r.mapPoint(transform.Apply(segment.Points[0]))
			r.pw.LineTo(p.X, p.Y)
		case svg.SegmentCubicTo:
			if len(segment.Points) != 3 {
				continue
			}
			curX, curY := r.pw.Loc()
			points := []Location{
				{X: curX, Y: curY},
				toLocation(r.mapPoint(transform.Apply(segment.Points[0]))),
				toLocation(r.mapPoint(transform.Apply(segment.Points[1]))),
				toLocation(r.mapPoint(transform.Apply(segment.Points[2]))),
			}
			_ = r.pw.CurvePoints(points)
		case svg.SegmentClosePath:
			if started {
				r.pw.LineTo(start.X, start.Y)
			}
		}
	}
}

func toLocation(point svg.Point) Location {
	return Location{X: point.X, Y: point.Y}
}

func (r *svgRenderer) finishPath(style svg.Style, fill, stroke bool) error {
	prevFill := r.pw.SetFillColor(style.Fill.Color)
	prevStroke := r.pw.SetLineColor(style.Stroke.Color)
	prevWidth := r.pw.setLineWidth(r.mapLengthX(style.StrokeWidth))
	prevCap := r.pw.SetLineCapStyle(mapSVGLineCap(style.LineCap))
	prevJoin := r.pw.SetLineJoinStyle(mapSVGLineJoin(style.LineJoin))
	prevMiter := r.pw.SetMiterLimit(style.MiterLimit)
	prevDash := r.pw.SetLineDashPattern(dashPatternForStyle(style, r.scaleX))
	defer func() {
		r.pw.SetFillColor(prevFill)
		r.pw.SetLineColor(prevStroke)
		r.pw.setLineWidth(prevWidth)
		r.pw.SetLineCapStyle(prevCap)
		r.pw.SetLineJoinStyle(prevJoin)
		r.pw.SetMiterLimit(prevMiter)
		r.pw.SetLineDashPattern(prevDash)
	}()
	switch {
	case fill && stroke:
		return fillAndStrokeWithRule(r.pw, style.FillRule)
	case fill:
		return fillWithRule(r.pw, style.FillRule)
	default:
		return r.pw.Stroke()
	}
}

func fillWithRule(pw *PageWriter, rule string) error {
	if !strings.EqualFold(rule, "evenodd") {
		return pw.Fill()
	}
	if err := pw.requireActivePath(); err != nil {
		return err
	}
	pw.startGraph()
	pw.checkSetFillColor()
	pw.gw.eoFill()
	pw.inPath = false
	pw.restorePathState()
	return nil
}

func fillAndStrokeWithRule(pw *PageWriter, rule string) error {
	if !strings.EqualFold(rule, "evenodd") {
		return pw.FillAndStroke()
	}
	if err := pw.requireActivePath(); err != nil {
		return err
	}
	pw.startGraph()
	pw.checkSetFillColor()
	pw.checkSetLineColor()
	pw.checkSetLineWidth()
	pw.checkSetLineDashPattern()
	pw.gw.eoFillAndStroke()
	pw.inPath = false
	pw.restorePathState()
	return nil
}

func dashPatternForStyle(style svg.Style, scale float64) string {
	if len(style.DashArray) == 0 {
		return "solid"
	}
	parts := make([]string, len(style.DashArray))
	for i, value := range style.DashArray {
		parts[i] = g(value * scale)
	}
	offset := int(style.DashOffset * scale)
	return fmt.Sprintf("[%s] %d", strings.Join(parts, " "), offset)
}

func mapSVGLineCap(value string) LineCapStyle {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "round":
		return RoundCap
	case "square":
		return ProjectingSquareCap
	default:
		return ButtCap
	}
}

func mapSVGLineJoin(value string) LineJoinStyle {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "round":
		return RoundJoin
	case "bevel":
		return BevelJoin
	default:
		return MiterJoin
	}
}

func mapSVGBlendMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "normal":
		return "Normal"
	case "hard-light":
		return "HardLight"
	default:
		logSVGWarnings([]svg.Warning{{Element: "style", Attribute: "mix-blend-mode", Message: fmt.Sprintf("%q is unsupported; falling back to normal", value)}})
		return "Normal"
	}
}

func normalizeFontStyle(style svg.Style) string {
	if strings.EqualFold(style.FontStyle, "italic") || strings.EqualFold(style.FontStyle, "oblique") {
		return "italic"
	}
	return ""
}

func normalizeFontWeight(style svg.Style) string {
	if strings.EqualFold(style.FontWeight, "bold") || style.FontWeight == "700" {
		return "bold"
	}
	return ""
}

func (r *svgRenderer) drawText(text *svg.Text, style svg.Style, transform svg.Transform) error {
	layout, err := r.layoutText(text, style, transform)
	if err != nil || layout == nil {
		return err
	}
	blendMode := mapSVGBlendMode(style.BlendMode)
	alpha := clamp01(style.Opacity * style.FillOpacity)
	if style.Stroke.IsGradient() {
		logSVGWarnings([]svg.Warning{{Element: "text", Attribute: "stroke", Message: "gradient stroke on text is not yet implemented"}})
	}
	if style.Fill.IsGradient() {
		resolved, err := r.resolveTextGradient(style.Fill.Ref, alpha, layout.rich, layout.startXSVG, layout.baselineY, transform)
		if err != nil {
			logSVGWarnings([]svg.Warning{{Element: "linearGradient", Message: err.Error()}})
			if style.Fill.None {
				return nil
			}
			return r.withScopedGraphicsState(alpha, 1, blendMode, nil, "", nil, func() error {
				return r.withPreparedTextState(style, func() error {
					r.pw.SetFontColor(style.Fill.Color)
					r.pw.MoveTo(layout.startX, layout.mappedY)
					return r.pw.Print(text.Body)
				})
			})
		}
		if resolved.varyingAlpha {
			if svgGradientStopOpacityMode(r.pw.options) == svgGradientStopOpacityModeFlat {
				alpha = resolved.flatAlpha
			} else {
				logSVGWarnings([]svg.Warning{{Element: "text", Attribute: "fill", Message: "gradient text stop-opacity falls back to constant text opacity"}})
				alpha = resolved.uniformAlpha
			}
		} else {
			alpha = resolved.uniformAlpha
		}
		return r.withScopedGraphicsState(alpha, 1, blendMode, nil, "", nil, func() error {
			return r.withPreparedTextState(style, func() error {
				r.pw.MoveTo(layout.startX, layout.mappedY)
				var renderErr error
				err := r.pw.ClipRichText(layout.rich, func() {
					switch {
					case resolved.linear != nil:
						renderErr = r.pw.PaintLinearGradient(resolved.linear)
					case resolved.radial != nil:
						renderErr = r.pw.PaintRadialGradient(resolved.radial)
					}
				})
				if err != nil {
					return err
				}
				return renderErr
			})
		})
	}
	if style.Fill.None {
		return nil
	}
	return r.withScopedGraphicsState(alpha, 1, blendMode, nil, "", nil, func() error {
		return r.withPreparedTextState(style, func() error {
			r.pw.SetFontColor(style.Fill.Color)
			r.pw.MoveTo(layout.startX, layout.mappedY)
			return r.pw.Print(text.Body)
		})
	})
}

type svgTextLayout struct {
	rich      *rich_text.RichText
	startX    float64
	mappedY   float64
	startXSVG float64
	baselineY float64
}

func (r *svgRenderer) layoutText(text *svg.Text, style svg.Style, transform svg.Transform) (*svgTextLayout, error) {
	if text.Body == "" {
		return nil, nil
	}
	if transform.B != 0 || transform.C != 0 || transform.A != 1 || transform.D != 1 {
		logSVGWarnings([]svg.Warning{{Element: "text", Message: "text transforms beyond translation are not yet implemented"}})
	}
	point := transform.Apply(svg.Point{X: text.X, Y: text.Y})
	mapped := r.mapPoint(point)
	var rich *rich_text.RichText
	err := r.withPreparedTextState(style, func() error {
		var err error
		rich, err = r.pw.richTextForString(text.Body)
		return err
	})
	if err != nil {
		return nil, err
	}
	if rich == nil || rich.Len() == 0 {
		return nil, nil
	}
	startX := mapped.X
	startXSVG := point.X
	if style.TextAnchor != "" && style.TextAnchor != "start" {
		switch strings.ToLower(style.TextAnchor) {
		case "middle":
			startX -= rich.Width() / 2
			startXSVG -= r.textWidthSVG(rich) / 2
		case "end":
			startX -= rich.Width()
			startXSVG -= r.textWidthSVG(rich)
		}
	}
	return &svgTextLayout{
		rich:      rich,
		startX:    startX,
		mappedY:   mapped.Y,
		startXSVG: startXSVG,
		baselineY: point.Y,
	}, nil
}

func (r *svgRenderer) withPreparedTextState(style svg.Style, fn func() error) error {
	savedFonts := append([]*font.Font(nil), r.pw.fonts...)
	savedFontSize := r.pw.fontSize
	savedFontColor := r.pw.fontColor
	prevUnits := r.pw.units
	defer func() {
		r.pw.fonts = savedFonts
		r.pw.fontSize = savedFontSize
		r.pw.fontColor = savedFontColor
		r.pw.units = prevUnits
	}()
	r.pw.SetFontColor(style.Fill.Color)
	r.pw.units = UnitConversions["pt"]
	opts := options.Options{
		"style":  normalizeFontStyle(style),
		"weight": normalizeFontWeight(style),
	}
	candidates := svgFontFamilyCandidates(style)
	var lastErr error
	for i, candidate := range candidates {
		if _, err := r.pw.SetFont(candidate.family, style.FontSize*r.scaleY, opts); err != nil {
			lastErr = err
			continue
		}
		if i > 0 {
			logSVGWarnings([]svg.Warning{{Element: "text", Attribute: "font-family", Message: fmt.Sprintf("font %q unavailable; using fallback %q", candidates[0].label, candidate.label)}})
		}
		if fn == nil {
			return nil
		}
		return fn()
	}
	if lastErr != nil {
		labels := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			labels = append(labels, fmt.Sprintf("%q", candidate.label))
		}
		logSVGWarnings([]svg.Warning{{Element: "text", Attribute: "font-family", Message: fmt.Sprintf("fonts unavailable: tried %s: %v", strings.Join(labels, ", "), lastErr)}})
	} else {
		logSVGWarnings([]svg.Warning{{Element: "text", Attribute: "font-family", Message: "no usable font families"}})
	}
	return nil
}

type svgFontFamilyCandidate struct {
	label  string
	family string
}

func svgFontFamilyCandidates(style svg.Style) []svgFontFamilyCandidate {
	families := style.FontFamilies
	if len(families) == 0 && strings.TrimSpace(style.FontFamily) != "" {
		families = []string{style.FontFamily}
	}
	candidates := make([]svgFontFamilyCandidate, 0, len(families))
	for _, family := range families {
		family = strings.TrimSpace(family)
		if family == "" {
			continue
		}
		candidates = append(candidates, svgFontFamilyCandidate{
			label:  family,
			family: mapSVGGenericFontFamily(family),
		})
	}
	if len(candidates) == 0 {
		candidates = append(candidates, svgFontFamilyCandidate{label: "Helvetica", family: "Helvetica"})
	}
	return candidates
}

func mapSVGGenericFontFamily(family string) string {
	switch strings.ToLower(strings.TrimSpace(family)) {
	case "sans-serif":
		return "Helvetica"
	case "serif":
		return "Times"
	case "monospace":
		return "Courier"
	default:
		return family
	}
}

func (r *svgRenderer) resolveTextGradient(ref string, opacityScale float64, text *rich_text.RichText, startX, baselineY float64, transform svg.Transform) (*resolvedSVGGradient, error) {
	if text == nil {
		return nil, fmt.Errorf("gradient text has no content")
	}
	width := r.textWidthSVG(text)
	height := r.textHeightSVG(text)
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("gradient text has no bounds")
	}
	minY := baselineY - r.textAscentSVG(text)
	segments := svgRectSegments(startX, minY, width, height, 0, 0)
	return r.resolveGradient(ref, opacityScale, segments, transform)
}

func (r *svgRenderer) textWidthSVG(text *rich_text.RichText) float64 {
	if text == nil || r.scaleX == 0 {
		return 0
	}
	return text.Width() / r.scaleX
}

func (r *svgRenderer) textAscentSVG(text *rich_text.RichText) float64 {
	if text == nil || r.scaleY == 0 {
		return 0
	}
	return text.Ascent() / r.scaleY
}

func (r *svgRenderer) textHeightSVG(text *rich_text.RichText) float64 {
	if text == nil || r.scaleY == 0 {
		return 0
	}
	return (text.Ascent() - text.Descent()) / r.scaleY
}
