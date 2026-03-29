package pdf

import (
	"fmt"
	"os"
	"strings"

	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
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

type svgRenderer struct {
	doc     *svg.Document
	pw      *PageWriter
	x       float64
	y       float64
	width   float64
	height  float64
	scaleX  float64
	scaleY  float64
	viewBox svg.ViewBox
}

func newSVGRenderer(doc *svg.Document, pw *PageWriter, x, y, width, height float64) *svgRenderer {
	viewBox := doc.ViewBox
	if viewBox.Width == 0 || viewBox.Height == 0 {
		viewBox = svg.ViewBox{Width: doc.Width, Height: doc.Height}
	}
	return &svgRenderer{
		doc:     doc,
		pw:      pw,
		x:       x,
		y:       y,
		width:   width,
		height:  height,
		scaleX:  width / viewBox.Width,
		scaleY:  height / viewBox.Height,
		viewBox: viewBox,
	}
}

func (r *svgRenderer) mapPoint(point svg.Point) svg.Point {
	return svg.Point{
		X: r.x + ((point.X-r.viewBox.MinX)*r.scaleX),
		Y: r.y + ((point.Y-r.viewBox.MinY)*r.scaleY),
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
		if err := r.renderNode(child, baseStyle, svg.IdentityTransform(), ""); err != nil {
			return err
		}
	}
	return nil
}

func (r *svgRenderer) renderNode(node svg.Node, parentStyle svg.Style, parentTransform svg.Transform, parentClip string) error {
	common := node.NodeCommon()
	style := common.Style.Resolve(parentStyle)
	if style.Display == "none" {
		return nil
	}
	transform := parentTransform.Mul(common.Transform)
	clipRef := parentClip
	if common.ClipPathRef != "" {
		clipRef = common.ClipPathRef
	}
	return r.withClip(clipRef, transform, style, func() error {
		switch n := node.(type) {
		case *svg.Group:
			for _, child := range n.Children {
				if err := r.renderNode(child, style, transform, clipRef); err != nil {
					return err
				}
			}
		case *svg.Defs, *svg.ClipPath:
			return nil
		case *svg.Line:
			return r.drawSegments([]svg.Segment{
				{Type: svg.SegmentMoveTo, Points: []svg.Point{{X: n.X1, Y: n.Y1}}},
				{Type: svg.SegmentLineTo, Points: []svg.Point{{X: n.X2, Y: n.Y2}}},
			}, style, transform)
		case *svg.Rect:
			return r.drawSegments(svgRectPath(n), style, transform)
		case *svg.Circle:
			return r.drawSegments(svgEllipsePath(n.CX, n.CY, n.R, n.R), style, transform)
		case *svg.Ellipse:
			return r.drawSegments(svgEllipsePath(n.CX, n.CY, n.RX, n.RY), style, transform)
		case *svg.Polyline:
			return r.drawSegments(svgPolylinePath(n.Points, false), style, transform)
		case *svg.Polygon:
			return r.drawSegments(svgPolylinePath(n.Points, true), style, transform)
		case *svg.Path:
			return r.drawSegments(n.Segments, style, transform)
		case *svg.Text:
			return r.drawText(n, style, transform)
		default:
			return nil
		}
		return nil
	})
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
	default:
		return nil
	}
}

func (r *svgRenderer) drawSegments(segments []svg.Segment, style svg.Style, transform svg.Transform) error {
	if len(segments) == 0 {
		return nil
	}
	fill := style.Fill.Set && !style.Fill.None
	stroke := style.Stroke.Set && !style.Stroke.None && style.StrokeWidth > 0
	if !fill && !stroke {
		return nil
	}
	if style.Opacity < 1 || style.FillOpacity < 1 || style.StrokeOpacity < 1 {
		logSVGWarnings([]svg.Warning{{Element: "style", Message: "opacity is not yet implemented; rendering as opaque"}})
	}
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
	if text.Body == "" {
		return nil
	}
	if transform.B != 0 || transform.C != 0 || transform.A != 1 || transform.D != 1 {
		logSVGWarnings([]svg.Warning{{Element: "text", Message: "text transforms beyond translation are not yet implemented"}})
	}
	point := transform.Apply(svg.Point{X: text.X, Y: text.Y})
	mapped := r.mapPoint(point)
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
	if _, err := r.pw.SetFont(style.FontFamily, style.FontSize*r.scaleY, options.Options{
		"style":  normalizeFontStyle(style),
		"weight": normalizeFontWeight(style),
	}); err != nil {
		logSVGWarnings([]svg.Warning{{Element: "text", Attribute: "font-family", Message: fmt.Sprintf("font %q unavailable: %v", style.FontFamily, err)}})
		return nil
	}
	startX := mapped.X
	if style.TextAnchor != "" && style.TextAnchor != "start" {
		if rich, err := r.pw.richTextForString(text.Body); err == nil {
			switch strings.ToLower(style.TextAnchor) {
			case "middle":
				startX -= rich.Width() / 2
			case "end":
				startX -= rich.Width()
			}
		}
	}
	r.pw.MoveTo(startX, mapped.Y)
	return r.pw.Print(text.Body)
}
