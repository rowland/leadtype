package ltml

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/wordbreaking"
)

type sectorFacing uint8

const (
	sectorFacingAuto sectorFacing = iota
	sectorFacingUpright
	sectorFacingUpsideDown
)

type sectorParagraphLayout struct {
	lines     []*rich_text.RichText
	intervals []radialInterval
	total     float64
}

type sectorParagraphLayoutProvider interface {
	sectorParagraphLayoutFor(p *StdParagraph, w Writer) *sectorParagraphLayout
	drawSectorParagraph(p *StdParagraph, w Writer, layout *sectorParagraphLayout) error
}

type StdSector struct {
	StdContainer
	textPieces       []textPiece
	richText         *rich_text.RichText
	facing           sectorFacing
	angle            float64
	angleSet         bool
	textAlign        HAlign
	textAlignSet     bool
	geometry         radialSectorGeometry
	localBounds      radialBounds
	localPolygon     []radialPoint
	contentRotation  float64
	paragraphLayouts map[*StdParagraph]*sectorParagraphLayout
}

func (s *StdSector) AddText(text string) {
	s.AddTextWithFont(text, s.explicitFont())
}

func (s *StdSector) AddTextWithFont(text string, font *FontStyle) {
	addNormalizedTextPiece(&s.textPieces, &s.richText, text, font, normalizeLabelXMLText)
}

func (s *StdSector) AddInlineWithFont(content inlineText, font *FontStyle) {
	addInlineTextPiece(&s.textPieces, &s.richText, content, font)
}

func (s *StdSector) AccessibilityText() string {
	return resolvedTextPieces(s.textPieces, documentForContainer(s))
}

func (s *StdSector) DrawBorder(w Writer) error {
	if s.border == nil {
		return nil
	}
	x := s.geometry.CenterX - s.geometry.OuterRadius
	y := s.geometry.CenterY - s.geometry.OuterRadius
	if err := s.border.ApplyInRect(w, x, y, s.geometry.OuterRadius*2, s.geometry.OuterRadius*2); err != nil {
		return err
	}
	if s.geometry.InnerRadius > 0 {
		return w.Arch(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, s.geometry.InnerRadius, s.geometry.StartAngle, s.geometry.EndAngle, true, false, false)
	}
	return w.Pie(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, s.geometry.StartAngle, s.geometry.EndAngle, true, false, false)
}

func (s *StdSector) DrawContent(w Writer) error {
	return withWidgetRoleAccessibility(w, &s.StdWidget, "TD", s.AccessibilityText(), func() error {
		return s.withSectorClip(w, func() error {
			if rt := s.RichText(w); rt != nil && rt.Len() > 0 {
				if err := s.drawInlineText(w, rt); err != nil {
					return err
				}
			}
			return s.drawChildrenWithRotation(w)
		})
	})
}

func (s *StdSector) LayoutWidget(w Writer) {
	if s.layout != nil {
		LayoutContainer(s, w)
		return
	}

	static := make([]Widget, 0, len(s.Widgets()))
	for _, child := range s.Widgets() {
		if child.Position() == Static {
			static = append(static, child)
		}
	}
	if len(static) > 0 {
		seedWidth := s.seedContentWidth()
		gap := s.LayoutStyle().VPadding()
		centerX, centerY := s.contentLocalCenter()
		stable := false
		for pass := 0; pass < 6; pass++ {
			s.paragraphLayouts = make(map[*StdParagraph]*sectorParagraphLayout)
			totalHeight := 0.0
			totalHeight = 0.0
			for i, child := range static {
				if !child.WidthIsSet() {
					if _, ok := child.(*StdParagraph); ok {
						child.SetWidth(seedWidth)
					} else {
						pw := child.PreferredWidth(w)
						if pw == 0 {
							pw = seedWidth
						}
						child.SetWidth(min(pw, seedWidth))
					}
				}
				if !child.HeightIsSet() || pass > 0 {
					child.SetHeight(child.PreferredHeight(w))
				}
				totalHeight += child.Height()
				if i > 0 {
					totalHeight += gap
				}
			}
			bounds := s.contentLocalBounds()
			y := clampFloat(
				s.geometry.AnchorY+centerY-(totalHeight/2),
				s.geometry.AnchorY+bounds.MinY,
				s.geometry.AnchorY+bounds.MaxY-totalHeight,
			)
			changed := false
			for _, child := range static {
				centerLocalY := (y + child.Height()/2) - s.geometry.AnchorY
				interval := s.contentLineIntervalAt(centerLocalY)
				localLeft := clampFloat(centerX-child.Width()/2, interval.MinX, interval.MaxX-child.Width())
				left := s.geometry.AnchorX + localLeft
				if !floatEquals(child.Left(), left) || !floatEquals(child.Top(), y) {
					changed = true
				}
				child.SetLeft(left)
				child.SetTop(y)
				child.LayoutWidget(w)
				y += child.Height() + gap
			}
			if !changed {
				stable = true
				break
			}
		}
		if !stable {
			s.paragraphLayouts = make(map[*StdParagraph]*sectorParagraphLayout)
		}
	}
	layoutPositionedChildren(s, w)
}

func (s *StdSector) PaintBackground(w Writer) error {
	if s.fill == nil {
		return nil
	}
	s.fill.Apply(w)
	if s.geometry.InnerRadius > 0 {
		return w.Arch(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, s.geometry.InnerRadius, s.geometry.StartAngle, s.geometry.EndAngle, false, true, false)
	}
	return w.Pie(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, s.geometry.StartAngle, s.geometry.EndAngle, false, true, false)
}

func (s *StdSector) PreferredHeight(w Writer) float64 {
	if s.height != 0 {
		return float64(s.height)
	}
	if len(s.localPolygon) > 0 {
		return (s.localBounds.MaxY - s.localBounds.MinY) + NonContentHeight(s)
	}
	return s.StdContainer.PreferredHeight(w)
}

func (s *StdSector) PreferredWidth(w Writer) float64 {
	if s.width != 0 {
		return float64(s.width)
	}
	if len(s.localPolygon) > 0 {
		return (s.localBounds.MaxX - s.localBounds.MinX) + NonContentWidth(s)
	}
	return s.StdContainer.PreferredWidth(w)
}

func (s *StdSector) ResolveSectorReferenceX(widget Widget) float64 {
	x, _ := s.resolveSectorReference(widget)
	return s.geometry.AnchorX + x
}

func (s *StdSector) ResolveSectorReferenceY(widget Widget) float64 {
	_, y := s.resolveSectorReference(widget)
	return s.geometry.AnchorY + y
}

func (s *StdSector) RichText(w Writer) *rich_text.RichText {
	return richTextForTextPieces(w, s, s.textPieces, &s.richText, s.Font())
}

func (s *StdSector) SetAttrs(attrs map[string]string) {
	s.StdContainer.SetAttrs(attrs)
	s.facing = sectorFacingAuto
	switch attrs["facing"] {
	case "upright":
		s.facing = sectorFacingUpright
	case "upside-down":
		s.facing = sectorFacingUpsideDown
	}
	if angle, ok := attrs["angle"]; ok {
		if value, err := strconv.ParseFloat(strings.TrimSpace(angle), 64); err == nil {
			s.angle = value
			s.angleSet = true
		}
	}
	s.textAlign = HAlignCenter
	if textAlign, ok := attrs["text-align"]; ok {
		s.textAlignSet = true
		switch textAlign {
		case "left":
			s.textAlign = HAlignLeft
		case "right":
			s.textAlign = HAlignRight
		default:
			s.textAlign = HAlignCenter
		}
	}
}

func (s *StdSector) SetContainer(container Container) error {
	if _, ok := container.(*StdSector); ok {
		s.container = container
		return nil
	}
	if container == nil || !isRadialLayoutStyle(container.LayoutStyle()) {
		return fmt.Errorf("sector must be child of a radial or radial-out container")
	}
	s.container = container
	return nil
}

func (s *StdSector) String() string {
	return fmt.Sprintf("StdSector %s", &s.StdContainer)
}

func (s *StdSector) drawChildrenWithRotation(w Writer) error {
	if len(s.Widgets()) == 0 {
		return nil
	}
	var renderErr error
	draw := func() {
		renderErr = s.StdContainer.drawChildren(w)
	}
	if s.contentRotation == 0 {
		draw()
		return renderErr
	}
	if err := w.Rotate(s.contentRotation, s.geometry.AnchorX, s.geometry.AnchorY, draw); err != nil {
		return err
	}
	return renderErr
}

func (s *StdSector) drawInlineText(w Writer, rt *rich_text.RichText) error {
	if s.angleSet {
		return s.drawStraightText(w, rt)
	}
	opts := pdf.CurvedTextOptions{
		Align:       s.curvedTextAlign(),
		VAlign:      pdf.VTextAlignMiddle,
		Direction:   s.curvedTextDirection(),
		Orientation: pdf.CurvedTextOrientationOutside,
		Facing:      s.curvedTextFacing(),
	}
	return w.DrawRichTextOnCircle(rt, s.geometry.CenterX, s.geometry.CenterY, (s.geometry.InnerRadius+s.geometry.OuterRadius)/2, s.inlineStartAngle(), opts)
}

func (s *StdSector) drawStraightText(w Writer, rt *rich_text.RichText) error {
	applyContainerFont(w, s)
	anchorX, anchorY := s.geometry.AnchorX, s.geometry.AnchorY
	startX := anchorX
	switch s.resolvedTextAlign() {
	case HAlignCenter:
		startX -= rt.Width() / 2
	case HAlignRight:
		startX -= rt.Width()
	}
	var renderErr error
	if err := w.Rotate(s.angle, anchorX, anchorY, func() {
		w.MoveTo(startX, anchorY+rt.Ascent()/2)
		w.PrintRichText(rt)
	}); err != nil {
		return err
	}
	return renderErr
}

func (s *StdSector) withSectorClip(w Writer, fn func() error) error {
	var clipErr error
	var renderErr error
	err := w.Path(func() {
		if clipErr = s.buildSectorPath(w); clipErr != nil {
			return
		}
		clipErr = w.Clip(func() {
			renderErr = fn()
		})
	})
	if err != nil {
		return err
	}
	if clipErr != nil {
		return clipErr
	}
	return renderErr
}

func (s *StdSector) buildSectorPath(w Writer) error {
	if s.geometry.InnerRadius > 0 {
		startX, startY := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, s.geometry.StartAngle)
		endInnerX, endInnerY := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, s.geometry.InnerRadius, s.geometry.EndAngle)
		w.MoveTo(startX, startY)
		if err := w.Arc(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, s.geometry.StartAngle, s.geometry.EndAngle, false); err != nil {
			return err
		}
		w.LineTo(endInnerX, endInnerY)
		if err := w.Arc(s.geometry.CenterX, s.geometry.CenterY, s.geometry.InnerRadius, s.geometry.EndAngle, s.geometry.StartAngle, false); err != nil {
			return err
		}
		w.LineTo(startX, startY)
		return nil
	}
	centerX, centerY := s.geometry.CenterX, s.geometry.CenterY
	startX, startY := radialPointAt(centerX, centerY, s.geometry.OuterRadius, s.geometry.StartAngle)
	w.MoveTo(centerX, centerY)
	w.LineTo(startX, startY)
	if err := w.Arc(centerX, centerY, s.geometry.OuterRadius, s.geometry.StartAngle, s.geometry.EndAngle, false); err != nil {
		return err
	}
	w.LineTo(centerX, centerY)
	return nil
}

func (s *StdSector) curvedTextAlign() pdf.CurvedTextHAlign {
	switch s.resolvedTextAlign() {
	case HAlignLeft:
		return pdf.CurvedTextAlignLeft
	case HAlignRight:
		return pdf.CurvedTextAlignRight
	default:
		return pdf.CurvedTextAlignCenter
	}
}

func (s *StdSector) curvedTextDirection() pdf.CurvedTextDirection {
	if s.geometry.AnchorY > s.geometry.CenterY {
		return pdf.CurvedTextCounterClockwise
	}
	return pdf.CurvedTextClockwise
}

func (s *StdSector) curvedTextFacing() pdf.CurvedTextFacing {
	switch s.facing {
	case sectorFacingUpright:
		return pdf.CurvedTextFacingUpright
	case sectorFacingUpsideDown:
		return pdf.CurvedTextFacingUpsideDown
	default:
		if s.geometry.AnchorY > s.geometry.CenterY {
			return pdf.CurvedTextFacingUpsideDown
		}
		return pdf.CurvedTextFacingUpright
	}
}

func (s *StdSector) inlineStartAngle() float64 {
	switch s.resolvedTextAlign() {
	case HAlignLeft:
		return s.geometry.StartAngle
	case HAlignRight:
		return s.geometry.EndAngle
	default:
		return s.geometry.AnchorAngle
	}
}

func (s *StdSector) resolvedTextAlign() HAlign {
	if s.textAlignSet {
		return s.textAlign
	}
	return HAlignCenter
}

func (s *StdSector) resolveSectorReference(widget Widget) (float64, float64) {
	xOrigin, yOrigin := widget.OriginX(), widget.OriginY()
	midRadius := (s.geometry.InnerRadius + s.geometry.OuterRadius) / 2
	angleForOrigin := func(origin OriginX) float64 {
		switch origin {
		case OriginXStart:
			return s.geometry.StartAngle
		case OriginXEnd:
			return s.geometry.EndAngle
		default:
			return s.geometry.AnchorAngle
		}
	}
	radiusForOrigin := func(origin OriginY) float64 {
		switch origin {
		case OriginYTop:
			return s.geometry.InnerRadius
		case OriginYBottom:
			return s.geometry.OuterRadius
		default:
			return midRadius
		}
	}
	if xOrigin != OriginXCustom && yOrigin != OriginYCustom {
		actualX, actualY := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, radiusForOrigin(yOrigin), angleForOrigin(xOrigin))
		return s.toLocal(actualX, actualY)
	}

	localX := s.localBounds.MinX
	if xOrigin == OriginXCustom {
		localX += widget.OriginXValue()
	}
	if xOrigin != OriginXCustom {
		actualX, actualY := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, midRadius, angleForOrigin(xOrigin))
		localX, _ = s.toLocal(actualX, actualY)
	}

	localY := s.localBounds.MinY
	if yOrigin == OriginYCustom {
		localY += widget.OriginYValue()
	}
	if yOrigin != OriginYCustom {
		actualX, actualY := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, radiusForOrigin(yOrigin), s.geometry.AnchorAngle)
		_, localY = s.toLocal(actualX, actualY)
	}
	return localX, localY
}

func (s *StdSector) sectorParagraphLayoutFor(p *StdParagraph, w Writer) *sectorParagraphLayout {
	if layout := s.paragraphLayouts[p]; layout != nil {
		return layout
	}
	rt := p.RichText(w)
	flags := make([]wordbreaking.Flags, rt.Len())
	wordbreaking.MarkRuneAttributes(rt.String(), flags)
	seed := rt.WrapToWidth(max(ContentWidth(p)-p.textIndent(), 1), flags, false)
	if len(seed) == 0 {
		seed = []*rich_text.RichText{rt}
	}
	lines := seed
	var intervals []radialInterval
	for range 4 {
		intervals = s.paragraphIntervalsForLines(p, w, lines)
		widths := make([]float64, 0, len(intervals))
		for _, interval := range intervals {
			widths = append(widths, max(interval.MaxX-interval.MinX-p.textIndent(), 1))
		}
		nextLines := wrapRichTextToWidths(rt, flags, widths)
		if richTextLinesEqual(lines, nextLines) {
			lines = nextLines
			break
		}
		lines = nextLines
	}
	layout := &sectorParagraphLayout{
		lines:     lines,
		intervals: s.paragraphIntervalsForLines(p, w, lines),
		total:     p.contentHeightForLines(lines, w),
	}
	s.paragraphLayouts[p] = layout
	return layout
}

func (s *StdSector) drawSectorParagraph(p *StdParagraph, w Writer, layout *sectorParagraphLayout) error {
	if layout == nil {
		return nil
	}
	indent := p.textIndent()
	y := ContentTop(p)
	if bullets := p.Bullets(); len(bullets) > 0 && !p.suppressBullet && len(layout.intervals) > 0 {
		interval := layout.intervals[0]
		bandStart := s.geometry.AnchorX + interval.MinX
		bandEnd := bandStart + indent
		if IsRTL(p) {
			bandEnd = s.geometry.AnchorX + interval.MaxX
			bandStart = bandEnd - indent
		}
		if err := withAccessibilityArtifact(w, func() error {
			return p.drawBulletsInBand(w, bullets, layout.lines[0], bandStart, bandEnd, y+layout.lines[0].Ascent(), layout.total)
		}); err != nil {
			return err
		}
	}
	for i, line := range layout.lines {
		interval := layout.intervals[min(i, len(layout.intervals)-1)]
		widthAvail := interval.MaxX - interval.MinX - indent
		x := s.geometry.AnchorX + interval.MinX + indent
		switch p.ParagraphStyle().ResolvedTextAlign(p) {
		case HAlignCenter:
			x += max((widthAvail-line.Width())/2, 0)
		case HAlignRight:
			x += max(widthAvail-line.Width(), 0)
		}
		w.MoveTo(x, y+line.Ascent())
		w.PrintRichText(line)
		y += line.Leading() * w.LineSpacing()
	}
	return nil
}

func (s *StdSector) paragraphIntervalsForLines(p *StdParagraph, w Writer, lines []*rich_text.RichText) []radialInterval {
	intervals := make([]radialInterval, 0, len(lines))
	y := ContentTop(p)
	for _, line := range lines {
		centerY := y + (line.Leading()*w.LineSpacing())/2
		localY := centerY - s.geometry.AnchorY
		intervals = append(intervals, s.contentLineIntervalAt(localY))
		y += line.Leading() * w.LineSpacing()
	}
	return intervals
}

func (s *StdSector) lineIntervalAt(localY float64) radialInterval {
	xs := make([]float64, 0, 8)
	for i := 0; i < len(s.localPolygon); i++ {
		a := s.localPolygon[i]
		b := s.localPolygon[(i+1)%len(s.localPolygon)]
		if (localY < min(a.Y, b.Y)) || (localY > max(a.Y, b.Y)) || a.Y == b.Y {
			continue
		}
		t := (localY - a.Y) / (b.Y - a.Y)
		xs = append(xs, a.X+t*(b.X-a.X))
	}
	if len(xs) < 2 {
		return radialInterval{MinX: s.localBounds.MinX, MaxX: s.localBounds.MaxX}
	}
	slices.Sort(xs)
	intervals := make([]radialInterval, 0, len(xs)/2)
	for i := 0; i+1 < len(xs); i += 2 {
		intervals = append(intervals, radialInterval{MinX: xs[i], MaxX: xs[i+1]})
	}
	for _, interval := range intervals {
		if interval.MinX <= 0 && 0 <= interval.MaxX {
			return interval
		}
	}
	best := intervals[0]
	bestWidth := best.MaxX - best.MinX
	for _, interval := range intervals[1:] {
		if width := interval.MaxX - interval.MinX; width > bestWidth {
			best = interval
			bestWidth = width
		}
	}
	return best
}

func (s *StdSector) setGeometry(geometry radialSectorGeometry) {
	s.geometry = geometry
	s.contentRotation = s.contentRotationAngle()
	s.localPolygon = s.buildLocalPolygon()
	s.localBounds = boundsForPoints(s.localPolygon)
	s.SetLeft(s.geometry.AnchorX + s.localBounds.MinX)
	s.SetTop(s.geometry.AnchorY + s.localBounds.MinY)
	s.SetWidth(s.localBounds.MaxX - s.localBounds.MinX)
	s.SetHeight(s.localBounds.MaxY - s.localBounds.MinY)
}

func (s *StdSector) buildLocalPolygon() []radialPoint {
	points := make([]radialPoint, 0, 48)
	span := s.geometry.EndAngle - s.geometry.StartAngle
	steps := max(int(math.Ceil(math.Abs(span)/15)), 4)
	for i := 0; i <= steps; i++ {
		angle := s.geometry.StartAngle + (span * float64(i) / float64(steps))
		x, y := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, angle)
		lx, ly := s.toLocal(x, y)
		points = append(points, radialPoint{X: lx, Y: ly})
	}
	if s.geometry.InnerRadius > 0 {
		for i := steps; i >= 0; i-- {
			angle := s.geometry.StartAngle + (span * float64(i) / float64(steps))
			x, y := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, s.geometry.InnerRadius, angle)
			lx, ly := s.toLocal(x, y)
			points = append(points, radialPoint{X: lx, Y: ly})
		}
	} else {
		points = append(points, radialPoint{0, 0})
	}
	return points
}

func (s *StdSector) contentRotationAngle() float64 {
	if s.angleSet {
		return s.angle
	}
	if s.hasParagraphChild() {
		return 0
	}
	theta := s.geometry.AnchorAngle * math.Pi / 180.0
	tangentX := math.Sin(theta)
	tangentY := math.Cos(theta)
	if s.geometry.AnchorY > s.geometry.CenterY {
		tangentX = -math.Sin(theta)
		tangentY = -math.Cos(theta)
	}
	if s.facing == sectorFacingUpsideDown {
		tangentX = -tangentX
		tangentY = -tangentY
	}
	if s.facing == sectorFacingUpright && s.geometry.AnchorY > s.geometry.CenterY {
		tangentX = -tangentX
		tangentY = -tangentY
	}
	return math.Atan2(tangentY, tangentX) * 180 / math.Pi
}

func (s *StdSector) toLocal(x, y float64) (float64, float64) {
	lx, ly := rotatePagePoint(x, y, s.geometry.AnchorX, s.geometry.AnchorY, -s.contentRotation)
	return lx - s.geometry.AnchorX, ly - s.geometry.AnchorY
}

func (s *StdSector) contentLocalBounds() radialBounds {
	bounds := s.localBounds
	bounds.MinX += s.PaddingLeft()
	bounds.MaxX -= s.PaddingRight()
	bounds.MinY += s.PaddingTop()
	bounds.MaxY -= s.PaddingBottom()
	if bounds.MaxX < bounds.MinX {
		midX := (bounds.MinX + bounds.MaxX) / 2
		bounds.MinX, bounds.MaxX = midX, midX
	}
	if bounds.MaxY < bounds.MinY {
		midY := (bounds.MinY + bounds.MaxY) / 2
		bounds.MinY, bounds.MaxY = midY, midY
	}
	return bounds
}

func (s *StdSector) contentLineIntervalAt(localY float64) radialInterval {
	interval := s.lineIntervalAt(localY)
	interval.MinX += s.PaddingLeft()
	interval.MaxX -= s.PaddingRight()
	if interval.MaxX < interval.MinX {
		midX := (interval.MinX + interval.MaxX) / 2
		interval.MinX, interval.MaxX = midX, midX
	}
	return interval
}

func (s *StdSector) contentLocalCenter() (float64, float64) {
	centroidX, centroidY, ok := polygonCentroid(s.localPolygon)
	if !ok {
		bounds := s.contentLocalBounds()
		return (bounds.MinX + bounds.MaxX) / 2, (bounds.MinY + bounds.MaxY) / 2
	}
	bounds := s.contentLocalBounds()
	return clampFloat(centroidX, bounds.MinX, bounds.MaxX), clampFloat(centroidY, bounds.MinY, bounds.MaxY)
}

func (s *StdSector) seedContentWidth() float64 {
	interval := s.contentLineIntervalAt(0)
	width := interval.MaxX - interval.MinX
	if width > 0 {
		return width
	}
	bounds := s.contentLocalBounds()
	return max(bounds.MaxX-bounds.MinX, 1)
}

func clampFloat(value, minValue, maxValue float64) float64 {
	if maxValue < minValue {
		return (minValue + maxValue) / 2
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func floatEquals(a, b float64) bool {
	return math.Abs(a-b) <= 0.001
}

func (s *StdSector) hasParagraphChild() bool {
	for _, child := range s.Widgets() {
		if _, ok := child.(*StdParagraph); ok {
			return true
		}
	}
	return false
}

func polygonCentroid(points []radialPoint) (float64, float64, bool) {
	if len(points) < 3 {
		return 0, 0, false
	}
	doubleArea := 0.0
	cx := 0.0
	cy := 0.0
	for i := 0; i < len(points); i++ {
		a := points[i]
		b := points[(i+1)%len(points)]
		cross := (a.X * b.Y) - (b.X * a.Y)
		doubleArea += cross
		cx += (a.X + b.X) * cross
		cy += (a.Y + b.Y) * cross
	}
	if math.Abs(doubleArea) < 0.000001 {
		return 0, 0, false
	}
	scale := 1 / (3 * doubleArea)
	return cx * scale, cy * scale, true
}

func wrapRichTextToWidths(rt *rich_text.RichText, flags []wordbreaking.Flags, widths []float64) []*rich_text.RichText {
	if rt == nil {
		return nil
	}
	current := rt
	currentFlags := flags
	lines := make([]*rich_text.RichText, 0, max(len(widths), 1))
	for i := 0; current != nil; i++ {
		width := widths[min(i, len(widths)-1)]
		line, remainder, remainderFlags := current.WordsToWidth(width, currentFlags, false)
		lines = append(lines, line.TrimSpace())
		current = remainder
		currentFlags = remainderFlags
	}
	return lines
}

func richTextLinesEqual(a, b []*rich_text.RichText) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].String() != b[i].String() {
			return false
		}
	}
	return true
}

func init() {
	registerTag(DefaultSpace, "sector", func() any { return &StdSector{} })
}

var _ AddInlineWithFonter = (*StdSector)(nil)
var _ AddTextWithFonter = (*StdSector)(nil)
var _ Container = (*StdSector)(nil)
var _ HasAttrs = (*StdSector)(nil)
var _ HasText = (*StdSector)(nil)
var _ Identifier = (*StdSector)(nil)
var _ Printer = (*StdSector)(nil)
var _ WantsContainer = (*StdSector)(nil)
var _ sectorParagraphLayoutProvider = (*StdSector)(nil)
