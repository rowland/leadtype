package ltml

import (
	"fmt"
	"math"
	"slices"
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
	lines       []*rich_text.RichText
	intervals   []radialInterval
	curvedLines []sectorCurvedParagraphLine
	total       float64
	curved      bool
	err         error
	direction   pdf.CurvedTextDirection
	facing      pdf.CurvedTextFacing
}

type sectorCurvedParagraphLine struct {
	radius   float64
	angle    float64
	arcWidth float64
}

type sectorParagraphFlowMode uint8

const (
	sectorParagraphFlowRadial sectorParagraphFlowMode = iota
	sectorParagraphFlowHorizontal
	sectorParagraphFlowMixed
)

type sectorLabelLayout struct {
	anchorX   float64
	anchorY   float64
	radius    float64
	angle     float64
	boxLeft   float64
	boxTop    float64
	boxWidth  float64
	boxHeight float64
	straight  bool
	textAngle float64
	arcWidth  float64
}

type sectorFlowLabelAnchor struct {
	x        float64
	y        float64
	arcWidth float64
}

type sectorPositionedReference struct {
	pageX, pageY  float64
	radius, angle float64
}

type sectorAngularEdge int8

const (
	sectorAngularMidpoint sectorAngularEdge = iota
	sectorAngularStart
	sectorAngularEnd
)

type sectorRadialEdge int8

const (
	sectorRadialMidpoint sectorRadialEdge = iota
	sectorRadialOuter
	sectorRadialInner
)

type sectorPositionedChild struct {
	angularEdge  sectorAngularEdge
	angularInset float32
	radialEdge   sectorRadialEdge
	radialInset  float32
}

type sectorParagraphLayoutProvider interface {
	sectorParagraphLayoutFor(p *StdParagraph, w Writer) *sectorParagraphLayout
	drawSectorParagraph(p *StdParagraph, w Writer, layout *sectorParagraphLayout) error
}

const (
	sectorBorderOuter = iota
	sectorBorderEnd
	sectorBorderInner
	sectorBorderStart
)

type StdSector struct {
	StdContainer
	geometry           radialSectorGeometry
	localBounds        radialBounds
	localPolygon       []radialPoint
	contentBounds      radialBounds
	contentPolygon     []radialPoint
	flowRotation       float64
	labelLayouts       map[*StdLabel]*sectorLabelLayout
	paragraphLayouts   map[*StdParagraph]*sectorParagraphLayout
	sectorBorders      [4]*PenStyle
	sectorBorderSet    [4]bool
	flowSlots          map[Widget]radialBounds
	flowLabelAnchors   map[*StdLabel]sectorFlowLabelAnchor
	positionedChildren map[*StdWidget]sectorPositionedChild
	clipDisabled       bool
}

func (s *StdSector) DrawBorder(w Writer) error {
	x := s.geometry.CenterX - s.geometry.OuterRadius
	y := s.geometry.CenterY - s.geometry.OuterRadius
	size := s.geometry.OuterRadius * 2
	hasEdgeOverrides := false
	for i := range s.sectorBorders {
		hasEdgeOverrides = hasEdgeOverrides || s.sectorBorderSet[i] || s.borderSideSet[i] || s.sectorBorders[i] != nil || s.borders[i] != nil
	}
	if s.border != nil && !hasEdgeOverrides {
		if err := s.border.ApplyInRect(w, x, y, size, size); err != nil {
			return err
		}
		if s.geometry.InnerRadius > 0 {
			return w.Arch(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, s.geometry.InnerRadius, s.geometry.StartAngle, s.geometry.EndAngle, true, false, false)
		}
		return w.Pie(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, s.geometry.StartAngle, s.geometry.EndAngle, true, false, false)
	}

	outer, inner, start, end := s.resolvedSectorBorders()
	if err := s.strokeSectorArc(w, outer, s.geometry.OuterRadius, x, y, size); err != nil {
		return err
	}
	if s.geometry.InnerRadius > 0 {
		if err := s.strokeSectorArc(w, inner, s.geometry.InnerRadius, x, y, size); err != nil {
			return err
		}
	}
	if err := s.strokeSectorRadius(w, start, s.geometry.StartAngle, x, y, size); err != nil {
		return err
	}
	return s.strokeSectorRadius(w, end, s.geometry.EndAngle, x, y, size)
}

func (s *StdSector) resolvedSectorBorders() (outer, inner, start, end *PenStyle) {
	outer, inner, start, end = s.border, s.border, s.border, s.border
	if s.borderSideSet[topSide] || s.borders[topSide] != nil {
		outer = s.borders[topSide]
	}
	if s.borderSideSet[bottomSide] || s.borders[bottomSide] != nil {
		inner = s.borders[bottomSide]
	}
	if s.radialSweep() == radialSweepCW {
		if s.borderSideSet[rightSide] || s.borders[rightSide] != nil {
			start = s.borders[rightSide]
		}
		if s.borderSideSet[leftSide] || s.borders[leftSide] != nil {
			end = s.borders[leftSide]
		}
	} else {
		if s.borderSideSet[leftSide] || s.borders[leftSide] != nil {
			start = s.borders[leftSide]
		}
		if s.borderSideSet[rightSide] || s.borders[rightSide] != nil {
			end = s.borders[rightSide]
		}
	}
	if s.sectorBorderSet[sectorBorderOuter] || s.sectorBorders[sectorBorderOuter] != nil {
		outer = s.sectorBorders[sectorBorderOuter]
	}
	if s.sectorBorderSet[sectorBorderInner] || s.sectorBorders[sectorBorderInner] != nil {
		inner = s.sectorBorders[sectorBorderInner]
	}
	if s.sectorBorderSet[sectorBorderStart] || s.sectorBorders[sectorBorderStart] != nil {
		start = s.sectorBorders[sectorBorderStart]
	}
	if s.sectorBorderSet[sectorBorderEnd] || s.sectorBorders[sectorBorderEnd] != nil {
		end = s.sectorBorders[sectorBorderEnd]
	}
	return
}

func (s *StdSector) radialSweep() radialSweep {
	if container, ok := s.Container().(*StdContainer); ok {
		return container.RadialSweep()
	}
	return radialSweepCCW
}

func (s *StdSector) strokeSectorArc(w Writer, pen *PenStyle, radius, x, y, size float64) error {
	if pen == nil || radius <= 0 {
		return nil
	}
	if err := pen.ApplyInRect(w, x, y, size, size); err != nil {
		return err
	}
	var strokeErr error
	if err := w.Path(func() {
		if strokeErr = w.Arc(s.geometry.CenterX, s.geometry.CenterY, radius, s.geometry.StartAngle, s.geometry.EndAngle, true); strokeErr != nil {
			return
		}
		strokeErr = w.Stroke()
	}); err != nil {
		return err
	}
	return strokeErr
}

func (s *StdSector) strokeSectorRadius(w Writer, pen *PenStyle, angle, x, y, size float64) error {
	if pen == nil {
		return nil
	}
	if err := pen.ApplyInRect(w, x, y, size, size); err != nil {
		return err
	}
	innerX, innerY := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, s.geometry.InnerRadius, angle)
	outerX, outerY := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, angle)
	var strokeErr error
	if err := w.Path(func() {
		w.MoveTo(innerX, innerY)
		w.LineTo(outerX, outerY)
		strokeErr = w.Stroke()
	}); err != nil {
		return err
	}
	return strokeErr
}

func (s *StdSector) DrawContent(w Writer) error {
	return withWidgetRoleAccessibility(w, &s.StdWidget, "TD", "", func() error {
		return s.withSectorClip(w, func() error {
			return s.drawChildren(w)
		})
	})
}

func (s *StdSector) LayoutWidget(w Writer) error {
	flowMode := s.staticParagraphFlowMode()
	if flowMode == sectorParagraphFlowMixed {
		return fmt.Errorf("%s: static curved and angle=\"0\" paragraphs cannot share a sector; position one paragraph or move it to another container", s.Path())
	}
	s.setFlowRotationForFlowMode(flowMode)
	s.rebuildContentGeometry()
	if len(s.contentPolygon) < 3 {
		s.flowSlots = nil
		return nil
	}
	if err := s.layoutStaticFlow(w); err != nil {
		return err
	}
	return s.layoutPositionedChildren(w)
}

func (s *StdSector) layoutPositionedChildren(w Writer) error {
	absolute, _ := printableWidgets(s, Absolute)
	if err := layoutWidgetsWithPosition(w, absolute, Absolute); err != nil {
		return err
	}
	relative, _ := printableWidgets(s, Relative)
	for _, child := range relative {
		if child.Printed() {
			child.SetVisible(false)
			continue
		}
		child.SetVisible(true)
		child.SetPosition(Relative)
		if !child.WidthIsSet() {
			width, err := child.PreferredWidth(w)
			if err != nil {
				return err
			}
			child.ResolveWidth(width)
		}
		if err := child.LayoutWidget(w); err != nil {
			return err
		}
		if !child.HeightIsSet() {
			height, err := child.PreferredHeight(w)
			if err != nil {
				return err
			}
			child.ResolveHeight(height)
		}
	}
	return nil
}

func (s *StdSector) PaintBackground(w Writer) error {
	if s.fill == nil {
		return nil
	}
	if s.fill.Kind() == BrushKindSweepGradient {
		band, err := resolveSectorSweepBand(s.fill, s.geometry)
		if err != nil || band == nil {
			return err
		}
		return w.PaintSweepBand(band)
	}
	s.fill.Apply(w)
	if s.geometry.InnerRadius > 0 {
		return w.Arch(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, s.geometry.InnerRadius, s.geometry.StartAngle, s.geometry.EndAngle, false, true, false)
	}
	return w.Pie(s.geometry.CenterX, s.geometry.CenterY, s.geometry.OuterRadius, s.geometry.StartAngle, s.geometry.EndAngle, false, true, false)
}

func (s *StdSector) PreferredHeight(w Writer) (float64, error) {
	if s.height != 0 {
		return float64(s.height), nil
	}
	if len(s.localPolygon) > 0 {
		return (s.localBounds.MaxY - s.localBounds.MinY) + NonContentHeight(s), nil
	}
	return s.StdContainer.PreferredHeight(w)
}

func (s *StdSector) PreferredWidth(w Writer) (float64, error) {
	if s.width != 0 {
		return float64(s.width), nil
	}
	if len(s.localPolygon) > 0 {
		return (s.localBounds.MaxX - s.localBounds.MinX) + NonContentWidth(s), nil
	}
	return s.StdContainer.PreferredWidth(w)
}

func (s *StdSector) ResolveSectorPlacement(widget *StdWidget) sectorPositionedPlacement {
	reference := s.resolvePositionedReference(widget)
	xFactor := originXAttachmentFactor(widget.originX)
	yFactor := originYAttachmentFactor(widget.originY)
	return sectorPositionedPlacement{
		boxLeft: reference.pageX - widget.Width()*xFactor,
		boxTop:  reference.pageY - widget.Height()*yFactor,
		anchorX: reference.pageX,
		anchorY: reference.pageY,
	}
}

func (s *StdSector) SetAttrs(attrs map[string]string) {
	s.StdContainer.SetAttrs(attrs)
	s.setSectorBorderResourceAttrs(attrs, s.Units())
	if clip, ok := attrs["clip"]; ok {
		s.clipDisabled = strings.TrimSpace(clip) == "false"
	}
}

func (s *StdSector) cachedLabelLayout(label *StdLabel) *sectorLabelLayout {
	if s.labelLayouts == nil {
		return nil
	}
	return s.labelLayouts[label]
}

func (s *StdSector) setLabelLayout(label *StdLabel, layout *sectorLabelLayout) {
	if s.labelLayouts == nil {
		s.labelLayouts = make(map[*StdLabel]*sectorLabelLayout)
	}
	s.labelLayouts[label] = layout
}

func (s *StdSector) invalidateLabelLayout(label *StdLabel) {
	delete(s.labelLayouts, label)
}

func (s *StdSector) invalidateParagraphLayout(paragraph *StdParagraph) {
	delete(s.paragraphLayouts, paragraph)
}

func (s *StdSector) invalidateTextLayouts() {
	s.labelLayouts = nil
	s.paragraphLayouts = nil
}

func (s *StdSector) resetResourceAttrs() {
	s.StdContainer.resetResourceAttrs()
	s.sectorBorders = [4]*PenStyle{}
	s.sectorBorderSet = [4]bool{}
}

func (s *StdSector) setResourceAttrs(attrs map[string]string, units Units) {
	s.StdContainer.setResourceAttrs(attrs, units)
	s.setSectorBorderResourceAttrs(attrs, units)
}

func (s *StdSector) setSectorBorderResourceAttrs(attrs map[string]string, units Units) {
	s.setSectorBorderStyle(sectorBorderOuter, "border-outer", attrs, units)
	s.setSectorBorderStyle(sectorBorderEnd, "border-end", attrs, units)
	s.setSectorBorderStyle(sectorBorderInner, "border-inner", attrs, units)
	s.setSectorBorderStyle(sectorBorderStart, "border-start", attrs, units)
}

func (s *StdSector) setChildPositionAttrs(widget *StdWidget, attrs map[string]string, units Units) {
	if !MapHasAnyKey(attrs, "start", "end", "outer", "inner") {
		return
	}
	if s.positionedChildren == nil {
		s.positionedChildren = make(map[*StdWidget]sectorPositionedChild)
	}
	state := s.positionedChildren[widget]
	// Apply the lower-priority member first so start and outer win
	// deterministically when opposing attributes occur in the same layer.
	if value, ok := attrs["end"]; ok {
		if strings.TrimSpace(value) == "auto" {
			state.angularEdge = sectorAngularMidpoint
			state.angularInset = 0
		} else {
			state.angularEdge = sectorAngularEnd
			state.angularInset = float32(ParseMeasurement(value, units))
		}
	}
	if value, ok := attrs["start"]; ok {
		if strings.TrimSpace(value) == "auto" {
			state.angularEdge = sectorAngularMidpoint
			state.angularInset = 0
		} else {
			state.angularEdge = sectorAngularStart
			state.angularInset = float32(ParseMeasurement(value, units))
		}
	}
	if value, ok := attrs["inner"]; ok {
		if strings.TrimSpace(value) == "auto" {
			state.radialEdge = sectorRadialMidpoint
			state.radialInset = 0
		} else {
			state.radialEdge = sectorRadialInner
			state.radialInset = float32(ParseMeasurement(value, units))
		}
	}
	if value, ok := attrs["outer"]; ok {
		if strings.TrimSpace(value) == "auto" {
			state.radialEdge = sectorRadialMidpoint
			state.radialInset = 0
		} else {
			state.radialEdge = sectorRadialOuter
			state.radialInset = float32(ParseMeasurement(value, units))
		}
	}
	s.positionedChildren[widget] = state
}

func isInlineOnlyWidget(widget Widget) bool {
	switch widget.(type) {
	case *StdA, *StdIndexPage, *StdIndexTitle, *StdLeader, *StdPageNo, *StdSpan:
		return true
	default:
		return false
	}
}

func (s *StdSector) setSectorBorderStyle(index int, attrName string, attrs map[string]string, units Units) {
	setOptionalPenStyle(&s.sectorBorders[index], &s.sectorBorderSet[index], attrName,
		attrs, s.Scope(), units, s.border)
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

func (s *StdSector) drawChildren(w Writer) error {
	if len(s.Widgets()) == 0 {
		return nil
	}
	children := slices.Clone(s.Widgets())
	slices.SortStableFunc(children, func(a, b Widget) int { return a.ZIndex() - b.ZIndex() })
	for _, child := range children {
		if !child.Visible() || child.Disabled() {
			continue
		}
		if err := Print(child, w); err != nil {
			return err
		}
	}
	return nil
}

func (s *StdSector) layoutSectorLabel(label *StdLabel, w Writer) error {
	textAngle, straight := label.sectorTextAngle()
	boxWidth, boxHeight := label.Width(), label.Height()
	if straight {
		if !label.WidthIsSet() {
			width, err := label.PreferredWidth(w)
			if err != nil {
				return err
			}
			label.ResolveWidth(width)
		}
		if !label.HeightIsSet() {
			height, err := label.PreferredHeight(w)
			if err != nil {
				return err
			}
			label.ResolveHeight(height)
		}
		boxWidth, boxHeight = label.Width(), label.Height()
	} else {
		rt := label.RichText(w)
		if rt != nil {
			boxWidth = rt.Width()
			boxHeight = rt.Leading() * w.LineSpacing()
		}
		if boxHeight <= 0 {
			boxHeight = effectiveFontSizeForContainer(label) * w.LineSpacing()
		}
	}

	align := label.sectorTextAlign()
	xFactor := 0.5
	switch align {
	case HAlignLeft:
		xFactor = 0
	case HAlignRight:
		xFactor = 1
	}
	if straight {
		xFactor = originXAttachmentFactor(label.StdWidget.originX)
	}
	yFactor := 0.5
	if straight {
		yFactor = originYAttachmentFactor(label.StdWidget.originY)
	} else {
		switch label.sectorTextVAlign() {
		case VAlignTop:
			yFactor = 0
		case VAlignBottom:
			yFactor = 1
		}
	}

	var anchorX, anchorY, boxLeft, boxTop, arcWidth float64
	if flowAnchor, flowing := s.flowLabelAnchors[label]; flowing && label.Position() == Static && !straight {
		anchorX, anchorY, arcWidth = flowAnchor.x, flowAnchor.y, flowAnchor.arcWidth
		boxLeft, boxTop = anchorX-boxWidth*xFactor, anchorY-boxHeight*yFactor
	} else if slot, flowing := s.flowSlots[label]; flowing && label.Position() == Static {
		localAnchorX := slot.MinX + (slot.MaxX-slot.MinX)*xFactor
		localAnchorY := slot.MinY + (slot.MaxY-slot.MinY)*yFactor
		anchorX, anchorY = rotatePagePoint(s.geometry.AnchorX+localAnchorX, s.geometry.AnchorY+localAnchorY,
			s.geometry.AnchorX, s.geometry.AnchorY, s.flowRotation)
		centerX, centerY := rotatePagePoint(s.geometry.AnchorX+(slot.MinX+slot.MaxX)/2, s.geometry.AnchorY+(slot.MinY+slot.MaxY)/2,
			s.geometry.AnchorX, s.geometry.AnchorY, s.flowRotation)
		boxLeft, boxTop = centerX-boxWidth/2, centerY-boxHeight/2
		band := s.contentBandForHeight(slot.MinY, slot.MaxY-slot.MinY)
		arcWidth = max(min(slot.MaxX, band.MaxX)-max(slot.MinX, band.MinX), 0)
	} else {
		reference := s.resolvePositionedReference(&label.StdWidget)
		anchorX, anchorY = reference.pageX, reference.pageY
		boxLeft, boxTop = anchorX-boxWidth*xFactor, anchorY-boxHeight*yFactor
	}
	dx := anchorX - s.geometry.CenterX
	dy := s.geometry.CenterY - anchorY
	radius := math.Hypot(dx, dy)
	angle := math.Atan2(dy, dx) * 180 / math.Pi
	if !straight && (radius <= radialAngleEpsilon || math.IsNaN(radius) || math.IsInf(radius, 0)) {
		return fmt.Errorf("%s: curved sector label requires a positive finite radius", label.Path())
	}

	if arcWidth <= 0 {
		arcWidth = s.availableArcWidth(radius, angle, align)
	}
	s.setLabelLayout(label, &sectorLabelLayout{
		anchorX:   anchorX,
		anchorY:   anchorY,
		radius:    radius,
		angle:     angle,
		boxLeft:   boxLeft,
		boxTop:    boxTop,
		boxWidth:  boxWidth,
		boxHeight: boxHeight,
		straight:  straight,
		textAngle: textAngle,
		arcWidth:  arcWidth,
	})
	return nil
}

func (s *StdSector) drawSectorLabel(label *StdLabel, w Writer) error {
	placement := s.cachedLabelLayout(label)
	textAngle, straight := label.sectorTextAngle()
	if placement == nil || placement.straight != straight || (straight && !floatEquals(placement.textAngle, textAngle)) {
		if err := s.layoutSectorLabel(label, w); err != nil {
			return err
		}
		placement = s.cachedLabelLayout(label)
	}
	if placement.straight {
		return label.drawBoxLabelContent(w, placement.textAngle)
	}
	return withWidgetRoleAccessibility(w, &label.StdWidget, "P", label.AccessibilityText(), func() error {
		rt := label.RichText(w)
		if rt == nil || rt.Len() == 0 {
			return nil
		}
		if label.shrinkToFit {
			arcWidth := placement.arcWidth
			if arcWidth > 0 && rt.Width() > arcWidth {
				rt = rt.Scale(arcWidth/rt.Width(), 6.0)
			}
		}
		direction, facing := sectorCurvedTextOrientation(label.sectorTextFacing(), placement.anchorY > s.geometry.CenterY)
		opts := pdf.CurvedTextOptions{
			Align:       s.labelCurvedTextAlign(label),
			VAlign:      s.labelCurvedTextVAlign(label),
			Direction:   direction,
			Orientation: pdf.CurvedTextOrientationOutside,
			Facing:      facing,
		}
		return w.DrawRichTextOnCircle(rt, s.geometry.CenterX, s.geometry.CenterY, placement.radius, placement.angle, opts)
	})
}

func (s *StdSector) availableArcWidth(radius, anchorAngle float64, align HAlign) float64 {
	if radius <= 0 {
		return 0
	}
	start := s.contentBoundaryAngle(true, radius)
	end := s.contentBoundaryAngle(false, radius)
	startDistance := s.angularDistanceAlongSweep(start, anchorAngle)
	endDistance := s.angularDistanceAlongSweep(anchorAngle, end)
	degrees := startDistance + endDistance
	switch align {
	case HAlignLeft:
		degrees = endDistance
	case HAlignRight:
		degrees = startDistance
	default:
		degrees = 2 * min(startDistance, endDistance)
	}
	return radius * degrees * math.Pi / 180
}

func (s *StdSector) angularDistanceAlongSweep(from, to float64) float64 {
	span := s.geometry.EndAngle - s.geometry.StartAngle
	distance := to - from
	if span < 0 {
		distance = -distance
	}
	for distance < 0 {
		distance += 360
	}
	return min(distance, math.Abs(span))
}

func (s *StdSector) labelCurvedTextAlign(label *StdLabel) pdf.CurvedTextHAlign {
	switch label.sectorTextAlign() {
	case HAlignLeft:
		return pdf.CurvedTextAlignLeft
	case HAlignRight:
		return pdf.CurvedTextAlignRight
	default:
		return pdf.CurvedTextAlignCenter
	}
}

func (s *StdSector) labelCurvedTextVAlign(label *StdLabel) pdf.VerticalTextAlign {
	switch label.sectorTextVAlign() {
	case VAlignTop:
		return pdf.VTextAlignTop
	case VAlignBottom:
		return pdf.VTextAlignBelow
	case VAlignBaseline:
		return pdf.VTextAlignBase
	case VAlignCapMiddle:
		return pdf.VTextAlignCapMiddle
	default:
		return pdf.VTextAlignMiddle
	}
}

func sectorCurvedTextOrientation(facing sectorFacing, belowCenter bool) (pdf.CurvedTextDirection, pdf.CurvedTextFacing) {
	switch facing {
	case sectorFacingUpright:
		return pdf.CurvedTextClockwise, pdf.CurvedTextFacingUpright
	case sectorFacingUpsideDown:
		return pdf.CurvedTextCounterClockwise, pdf.CurvedTextFacingUpsideDown
	default:
		if belowCenter {
			return pdf.CurvedTextCounterClockwise, pdf.CurvedTextFacingUpsideDown
		}
		return pdf.CurvedTextClockwise, pdf.CurvedTextFacingUpright
	}
}

func (s *StdSector) withSectorClip(w Writer, fn func() error) error {
	if len(s.contentPolygon) < 3 {
		return nil
	}
	if s.clipDisabled {
		return fn()
	}
	var clipErr error
	var renderErr error
	err := w.Path(func() {
		if clipErr = s.buildContentPath(w); clipErr != nil {
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

func (s *StdSector) buildContentPath(w Writer) error {
	for i, point := range s.contentPolygon {
		x, y := rotatePagePoint(s.geometry.AnchorX+point.X, s.geometry.AnchorY+point.Y,
			s.geometry.AnchorX, s.geometry.AnchorY, s.flowRotation)
		if i == 0 {
			w.MoveTo(x, y)
		} else {
			w.LineTo(x, y)
		}
	}
	first := s.contentPolygon[0]
	x, y := rotatePagePoint(s.geometry.AnchorX+first.X, s.geometry.AnchorY+first.Y,
		s.geometry.AnchorX, s.geometry.AnchorY, s.flowRotation)
	w.LineTo(x, y)
	return nil
}

func originXAttachmentFactor(origin OriginX) float64 {
	switch origin {
	case OriginXCenter:
		return 0.5
	case OriginXEnd:
		return 1
	default:
		return 0
	}
}

func originYAttachmentFactor(origin OriginY) float64 {
	switch origin {
	case OriginYMiddle:
		return 0.5
	case OriginYBottom:
		return 1
	default:
		return 0
	}
}

func (s *StdSector) resolvePositionedReference(widget *StdWidget) sectorPositionedReference {
	state := s.positionedChildren[widget]
	innerRadius, outerRadius := s.contentRadii()
	radius := (innerRadius + outerRadius) / 2
	switch state.radialEdge {
	case sectorRadialOuter:
		radius = outerRadius - float64(state.radialInset)
	case sectorRadialInner:
		radius = innerRadius + float64(state.radialInset)
	}

	angle := s.geometry.AnchorAngle
	switch state.angularEdge {
	case sectorAngularStart:
		angle = s.positionedBoundaryAngle(true, radius, float64(state.angularInset))
	case sectorAngularEnd:
		angle = s.positionedBoundaryAngle(false, radius, float64(state.angularInset))
	}

	pageX, pageY := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, radius, angle)
	pageX += widget.shiftXOffset()
	pageY += widget.shiftYOffset()
	return sectorPositionedReference{
		pageX: pageX, pageY: pageY,
		radius: radius, angle: angle,
	}
}

func (s *StdSector) positionedBoundaryAngle(start bool, radius, childInset float64) float64 {
	paddingStart, paddingEnd := s.radialSidePadding()
	padding := paddingEnd
	if start {
		padding = paddingStart
	}
	if math.Abs(s.geometry.EndAngle-s.geometry.StartAngle) >= 360-radialAngleEpsilon {
		padding = 0
	}
	return s.radialBoundaryAngle(start, radius, padding+childInset)
}

func (s *StdSector) sectorParagraphLayoutFor(p *StdParagraph, w Writer) *sectorParagraphLayout {
	if layout := s.paragraphLayouts[p]; layout != nil {
		return layout
	}
	var layout *sectorParagraphLayout
	if p.curvedInSector() {
		layout = s.curvedSectorParagraphLayoutFor(p, w)
	} else {
		layout = s.horizontalSectorParagraphLayoutFor(p, w)
	}
	if s.paragraphLayouts == nil {
		s.paragraphLayouts = make(map[*StdParagraph]*sectorParagraphLayout)
	}
	s.paragraphLayouts[p] = layout
	return layout
}

func (s *StdSector) horizontalSectorParagraphLayoutFor(p *StdParagraph, w Writer) *sectorParagraphLayout {
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
	return &sectorParagraphLayout{
		lines:     lines,
		intervals: s.paragraphIntervalsForLines(p, w, lines),
		total:     p.contentHeightForLines(lines, w),
	}
}

func (s *StdSector) curvedSectorParagraphLayoutFor(p *StdParagraph, w Writer) *sectorParagraphLayout {
	rt := p.RichText(w)
	if rt == nil || rt.Len() == 0 {
		return &sectorParagraphLayout{curved: true}
	}
	flags := make([]wordbreaking.Flags, rt.Len())
	wordbreaking.MarkRuneAttributes(rt.String(), flags)

	anchorRadius, anchorAngle, anchorY, static := s.curvedParagraphAnchor(p)
	direction, facing := sectorCurvedTextOrientation(p.sectorTextFacing(), anchorY > s.geometry.CenterY)

	initialWidth := s.curvedParagraphArcWidth(p, anchorRadius, anchorAngle, static)
	lines := wrapRichTextToWidths(rt, flags, []float64{max(initialWidth, 1)})
	var placements []sectorCurvedParagraphLine
	total := 0.0
	var layoutErr error
	for range 8 {
		placements, total, layoutErr = s.curvedParagraphLinePlacements(p, w, lines, anchorRadius, anchorAngle, static, direction, facing)
		if layoutErr != nil {
			break
		}
		widths := make([]float64, len(placements))
		for i, placement := range placements {
			widths[i] = max(placement.arcWidth, 1)
		}
		next := wrapRichTextToWidths(rt, flags, widths)
		if richTextLinesEqual(lines, next) {
			lines = next
			break
		}
		lines = next
	}
	if layoutErr == nil {
		placements, total, layoutErr = s.curvedParagraphLinePlacements(p, w, lines, anchorRadius, anchorAngle, static, direction, facing)
	}
	return &sectorParagraphLayout{
		lines: lines, curvedLines: placements, total: total, curved: true,
		direction: direction, facing: facing, err: layoutErr,
	}
}

func (s *StdSector) curvedParagraphAnchor(p *StdParagraph) (radius, angle, y float64, static bool) {
	if slot, ok := s.flowSlots[p]; ok && p.Position() == Static {
		anchor := s.flowLabelAnchorAt((slot.MinY+slot.MaxY)/2, 0)
		dx, dy := anchor.x-s.geometry.CenterX, s.geometry.CenterY-anchor.y
		return math.Hypot(dx, dy), math.Atan2(dy, dx) * 180 / math.Pi, anchor.y, true
	}
	if p.Position() == Static {
		innerRadius, outerRadius := s.contentRadii()
		radius = (innerRadius + outerRadius) / 2
		_, anchorY := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, radius, s.geometry.AnchorAngle)
		return radius, s.geometry.AnchorAngle, anchorY, true
	}
	reference := s.resolvePositionedReference(&p.StdWidget)
	x, y := reference.pageX, reference.pageY
	dx, dy := x-s.geometry.CenterX, s.geometry.CenterY-y
	return math.Hypot(dx, dy), math.Atan2(dy, dx) * 180 / math.Pi, y, false
}

func (s *StdSector) curvedParagraphLinePlacements(p *StdParagraph, w Writer, lines []*rich_text.RichText, anchorRadius, anchorAngle float64, static bool, direction pdf.CurvedTextDirection, facing pdf.CurvedTextFacing) ([]sectorCurvedParagraphLine, float64, error) {
	heights := make([]float64, len(lines))
	total := 0.0
	for i, line := range lines {
		height := line.Leading() * w.LineSpacing()
		if height <= 0 {
			height = effectiveFontSizeForContainer(p) * w.LineSpacing()
		}
		heights[i] = height
		total += height
	}
	centerRadius := anchorRadius
	if !static {
		switch p.OriginY() {
		case OriginYTop:
			centerRadius += total / 2
		case OriginYBottom:
			centerRadius -= total / 2
		}
	}
	progression := -1.0
	if facing == pdf.CurvedTextFacingUpsideDown {
		progression = 1
	}
	placements := make([]sectorCurvedParagraphLine, len(lines))
	position := 0.0
	for i, height := range heights {
		lineCenter := position + height/2
		radius := centerRadius + progression*(lineCenter-total/2)
		if radius <= radialAngleEpsilon || math.IsNaN(radius) || math.IsInf(radius, 0) {
			return nil, total, fmt.Errorf("%s: curved sector paragraph requires positive finite line radii", p.Path())
		}
		angle := anchorAngle
		align := p.ParagraphStyle().ResolvedTextAlign(p)
		if static {
			pathStart, pathEnd := s.curvedParagraphArcEndpoints(radius, direction)
			switch align {
			case HAlignCenter:
				angle = s.geometry.AnchorAngle
			case HAlignRight:
				angle = pathEnd
			default:
				angle = pathStart
			}
		} else {
			switch p.OriginX() {
			case OriginXStart:
				angle = s.contentBoundaryAngle(true, radius)
			case OriginXEnd:
				angle = s.contentBoundaryAngle(false, radius)
			}
		}
		placements[i] = sectorCurvedParagraphLine{
			radius: radius, angle: angle,
			arcWidth: s.curvedParagraphArcWidth(p, radius, angle, static),
		}
		position += height
	}
	return placements, total, nil
}

func (s *StdSector) curvedParagraphArcEndpoints(radius float64, direction pdf.CurvedTextDirection) (start, end float64) {
	start = s.contentBoundaryAngle(true, radius)
	end = s.contentBoundaryAngle(false, radius)
	sectorRunsCounterClockwise := s.geometry.EndAngle-s.geometry.StartAngle >= 0
	textRunsCounterClockwise := direction == pdf.CurvedTextCounterClockwise
	if sectorRunsCounterClockwise != textRunsCounterClockwise {
		return end, start
	}
	return start, end
}

func (s *StdSector) curvedParagraphArcWidth(p *StdParagraph, radius, angle float64, static bool) float64 {
	if radius <= 0 {
		return 0
	}
	if static || math.Abs(s.geometry.EndAngle-s.geometry.StartAngle) >= 360-radialAngleEpsilon {
		return s.contentArcWidth(radius)
	}
	align := p.ParagraphStyle().ResolvedTextAlign(p)
	if align == HAlignJustify {
		align = HAlignLeft
	}
	return s.availableArcWidth(radius, angle, align)
}

func (s *StdSector) contentArcWidth(radius float64) float64 {
	if radius <= 0 {
		return 0
	}
	span := math.Abs(s.geometry.EndAngle - s.geometry.StartAngle)
	if span >= 360-radialAngleEpsilon {
		return 2 * math.Pi * radius
	}
	start := s.contentBoundaryAngle(true, radius)
	end := s.contentBoundaryAngle(false, radius)
	return radius * s.angularDistanceAlongSweep(start, end) * math.Pi / 180
}

func (s *StdSector) drawSectorParagraph(p *StdParagraph, w Writer, layout *sectorParagraphLayout) error {
	if layout == nil {
		return nil
	}
	if layout.err != nil {
		return layout.err
	}
	if layout.curved {
		return s.drawCurvedSectorParagraph(p, w, layout)
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
		align := p.ParagraphStyle().ResolvedTextAlign(p)
		if align == HAlignJustify && i+1 == len(layout.lines) {
			align = HAlignLeft
		}
		xOffset, drawLine := paragraphAlignedLine(line, widthAvail, align)
		if align == HAlignCenter || align == HAlignRight {
			xOffset = max(xOffset, 0)
		}
		w.MoveTo(x+xOffset, y+line.Ascent())
		if p.textFill != nil {
			backgroundX, backgroundY, backgroundWidth, backgroundHeight := p.backgroundRect()
			var paintErr error
			if err := w.ClipRichText(drawLine, func() {
				paintErr = p.PaintBrushInRect(w, p.textFill, backgroundX, backgroundY, backgroundWidth, backgroundHeight)
			}); err != nil {
				return err
			}
			if paintErr != nil {
				return paintErr
			}
		} else {
			w.PrintRichText(drawLine)
		}
		y += line.Leading() * w.LineSpacing()
	}
	return nil
}

func (s *StdSector) drawCurvedSectorParagraph(p *StdParagraph, w Writer, layout *sectorParagraphLayout) error {
	align := p.ParagraphStyle().ResolvedTextAlign(p)
	for i, line := range layout.lines {
		if i >= len(layout.curvedLines) {
			break
		}
		placement := layout.curvedLines[i]
		drawLine := line
		curvedAlign := pdf.CurvedTextAlignLeft
		switch align {
		case HAlignCenter:
			curvedAlign = pdf.CurvedTextAlignCenter
		case HAlignRight:
			curvedAlign = pdf.CurvedTextAlignRight
		case HAlignJustify:
			if i+1 < len(layout.lines) {
				_, drawLine = paragraphAlignedLine(line, placement.arcWidth, HAlignJustify)
			}
		}
		if err := w.DrawRichTextOnCircle(drawLine, s.geometry.CenterX, s.geometry.CenterY, placement.radius, placement.angle, pdf.CurvedTextOptions{
			Align: curvedAlign, VAlign: pdf.VTextAlignMiddle,
			Direction: layout.direction, Facing: layout.facing,
			Orientation: pdf.CurvedTextOrientationOutside,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *StdSector) paragraphIntervalsForLines(p *StdParagraph, w Writer, lines []*rich_text.RichText) []radialInterval {
	intervals := make([]radialInterval, 0, len(lines))
	y := ContentTop(p)
	for _, line := range lines {
		lineHeight := line.Leading() * w.LineSpacing()
		localTop := y - s.geometry.AnchorY
		// A centerline chord can be wider than the sector both above and below
		// it. Use the intersection across the complete line box so ascenders,
		// descenders, fills, and accessibility clipping stay inside the padded
		// sector shape.
		if p.horizontalInSector() && s.flowRotation != 0 {
			polygon, bounds := s.horizontalContentGeometry()
			intervals = append(intervals, polygonBandForHeight(polygon, bounds, localTop, lineHeight))
		} else {
			intervals = append(intervals, s.contentBandForHeight(localTop, lineHeight))
		}
		y += lineHeight
	}
	return intervals
}

func polygonLineIntervalAt(polygon []radialPoint, bounds radialBounds, localY, preferredX float64) radialInterval {
	if len(polygon) < 3 || localY < bounds.MinY-radialAngleEpsilon || localY > bounds.MaxY+radialAngleEpsilon {
		return radialInterval{MinX: preferredX, MaxX: preferredX}
	}
	xs := make([]float64, 0, 8)
	for i := 0; i < len(polygon); i++ {
		a := polygon[i]
		b := polygon[(i+1)%len(polygon)]
		if a.Y == b.Y || !((a.Y <= localY && localY < b.Y) || (b.Y <= localY && localY < a.Y)) {
			continue
		}
		t := (localY - a.Y) / (b.Y - a.Y)
		xs = append(xs, a.X+t*(b.X-a.X))
	}
	if len(xs) < 2 {
		return radialInterval{MinX: preferredX, MaxX: preferredX}
	}
	slices.Sort(xs)
	intervals := make([]radialInterval, 0, len(xs)/2)
	for i := 0; i+1 < len(xs); i += 2 {
		intervals = append(intervals, radialInterval{MinX: xs[i], MaxX: xs[i+1]})
	}
	for _, interval := range intervals {
		if interval.MinX <= preferredX && preferredX <= interval.MaxX {
			return interval
		}
	}
	best := intervals[0]
	bestDistance := min(math.Abs(preferredX-best.MinX), math.Abs(preferredX-best.MaxX))
	for _, interval := range intervals[1:] {
		distance := min(math.Abs(preferredX-interval.MinX), math.Abs(preferredX-interval.MaxX))
		if distance < bestDistance-radialAngleEpsilon ||
			(math.Abs(distance-bestDistance) <= radialAngleEpsilon && interval.MaxX-interval.MinX > best.MaxX-best.MinX) {
			best = interval
			bestDistance = distance
		}
	}
	return best
}

func (s *StdSector) setGeometry(geometry radialSectorGeometry) {
	s.geometry = geometry
	s.flowRotation = s.flowRotationAngle(s.staticParagraphFlowMode())
	s.localPolygon = s.buildLocalPolygon()
	s.localBounds = boundsForPoints(s.localPolygon)
	s.rebuildContentGeometry()
	s.SetLeft(s.geometry.AnchorX + s.localBounds.MinX)
	s.SetTop(s.geometry.AnchorY + s.localBounds.MinY)
	s.SetWidth(s.localBounds.MaxX - s.localBounds.MinX)
	s.SetHeight(s.localBounds.MaxY - s.localBounds.MinY)
	s.invalidateTextLayouts()
	s.flowSlots = nil
	s.flowLabelAnchors = nil
}

func (s *StdSector) buildLocalPolygon() []radialPoint {
	return s.localizePolygon(s.buildPagePolygon(s.geometry.InnerRadius, s.geometry.OuterRadius))
}

func (s *StdSector) buildPagePolygon(innerRadius, outerRadius float64) []radialPoint {
	if outerRadius <= innerRadius || outerRadius <= 0 {
		return nil
	}
	points := make([]radialPoint, 0, 370)
	span := s.geometry.EndAngle - s.geometry.StartAngle
	steps := max(int(math.Ceil(math.Abs(span)/2)), 4)
	for i := 0; i <= steps; i++ {
		angle := s.geometry.StartAngle + (span * float64(i) / float64(steps))
		x, y := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, outerRadius, angle)
		points = append(points, radialPoint{X: x, Y: y})
	}
	if innerRadius > 0 {
		for i := steps; i >= 0; i-- {
			angle := s.geometry.StartAngle + (span * float64(i) / float64(steps))
			x, y := radialPointAt(s.geometry.CenterX, s.geometry.CenterY, innerRadius, angle)
			points = append(points, radialPoint{X: x, Y: y})
		}
	} else {
		points = append(points, radialPoint{X: s.geometry.CenterX, Y: s.geometry.CenterY})
	}
	return points
}

func (s *StdSector) localizePolygon(points []radialPoint) []radialPoint {
	result := make([]radialPoint, 0, len(points))
	for _, point := range points {
		x, y := s.toLocal(point.X, point.Y)
		result = append(result, radialPoint{X: x, Y: y})
	}
	return result
}

func (s *StdSector) rebuildContentGeometry() {
	points := s.buildContentPagePolygon()
	s.contentPolygon = s.localizePolygon(points)
	s.contentBounds = boundsForPoints(s.contentPolygon)
	s.invalidateTextLayouts()
}

func (s *StdSector) buildContentPagePolygon() []radialPoint {
	innerRadius, outerRadius := s.contentRadii()
	points := s.buildPagePolygon(innerRadius, outerRadius)
	if len(points) > 0 && math.Abs(s.geometry.EndAngle-s.geometry.StartAngle) < 360-radialAngleEpsilon {
		startPadding, endPadding := s.radialSidePadding()
		points = clipPolygonToRadialEdge(points, s.geometry, s.geometry.StartAngle, startPadding)
		points = clipPolygonToRadialEdge(points, s.geometry, s.geometry.EndAngle, endPadding)
	}
	return points
}

func (s *StdSector) horizontalContentGeometry() ([]radialPoint, radialBounds) {
	pagePoints := s.buildContentPagePolygon()
	points := make([]radialPoint, 0, len(pagePoints))
	for _, point := range pagePoints {
		points = append(points, radialPoint{X: point.X - s.geometry.AnchorX, Y: point.Y - s.geometry.AnchorY})
	}
	return points, boundsForPoints(points)
}

func (s *StdSector) contentRadii() (inner, outer float64) {
	inner = s.geometry.InnerRadius + s.PaddingBottom()
	outer = s.geometry.OuterRadius - s.PaddingTop()
	inner = max(inner, 0)
	outer = max(outer, 0)
	if outer < inner {
		outer = inner
	}
	return inner, outer
}

func (s *StdSector) radialSidePadding() (start, end float64) {
	if s.radialSweep() == radialSweepCW {
		return s.PaddingRight(), s.PaddingLeft()
	}
	return s.PaddingLeft(), s.PaddingRight()
}

func (s *StdSector) contentBoundaryAngle(start bool, radius float64) float64 {
	paddingStart, paddingEnd := s.radialSidePadding()
	padding := paddingEnd
	if start {
		padding = paddingStart
	}
	if math.Abs(s.geometry.EndAngle-s.geometry.StartAngle) >= 360-radialAngleEpsilon {
		padding = 0
	}
	return s.radialBoundaryAngle(start, radius, padding)
}

func (s *StdSector) radialBoundaryAngle(start bool, radius, inset float64) float64 {
	angle := s.geometry.EndAngle
	direction := -1.0
	if start {
		angle = s.geometry.StartAngle
		direction = 1
	}
	span := s.geometry.EndAngle - s.geometry.StartAngle
	if span < 0 {
		direction = -direction
	}
	if radius <= radialAngleEpsilon || inset == 0 {
		return angle
	}
	ratio := clampFloat(inset/radius, -1, 1)
	delta := math.Asin(ratio) * 180 / math.Pi
	return angle + direction*delta
}

func clipPolygonToRadialEdge(points []radialPoint, geometry radialSectorGeometry, edgeAngle, padding float64) []radialPoint {
	if len(points) == 0 || padding <= 0 {
		return points
	}
	theta := edgeAngle * math.Pi / 180
	dx, dy := math.Cos(theta), -math.Sin(theta)
	midRadius := (geometry.InnerRadius + geometry.OuterRadius) / 2
	midX, midY := radialPointAt(geometry.CenterX, geometry.CenterY, midRadius, geometry.AnchorAngle)
	signedDistance := func(point radialPoint) float64 {
		return dx*(point.Y-geometry.CenterY) - dy*(point.X-geometry.CenterX)
	}
	sign := 1.0
	if signedDistance(radialPoint{X: midX, Y: midY}) < 0 {
		sign = -1
	}
	distance := func(point radialPoint) float64 { return sign*signedDistance(point) - padding }
	result := make([]radialPoint, 0, len(points)+2)
	previous := points[len(points)-1]
	previousDistance := distance(previous)
	for _, current := range points {
		currentDistance := distance(current)
		previousInside, currentInside := previousDistance >= -radialAngleEpsilon, currentDistance >= -radialAngleEpsilon
		if previousInside != currentInside {
			t := previousDistance / (previousDistance - currentDistance)
			result = append(result, radialPoint{
				X: previous.X + t*(current.X-previous.X),
				Y: previous.Y + t*(current.Y-previous.Y),
			})
		}
		if currentInside {
			result = append(result, current)
		}
		previous, previousDistance = current, currentDistance
	}
	return result
}

func (s *StdSector) flowRotationAngle(mode sectorParagraphFlowMode) float64 {
	if mode == sectorParagraphFlowHorizontal || mode == sectorParagraphFlowMixed {
		return 0
	}
	return s.tangentContentRotationAngle()
}

func (s *StdSector) tangentContentRotationAngle() float64 {
	theta := s.geometry.AnchorAngle * math.Pi / 180.0
	tangentX := math.Sin(theta)
	tangentY := math.Cos(theta)
	if s.geometry.AnchorY > s.geometry.CenterY {
		tangentX = -math.Sin(theta)
		tangentY = -math.Cos(theta)
	}
	return math.Atan2(tangentY, tangentX) * 180 / math.Pi
}

func (s *StdSector) setFlowRotationForFlowMode(mode sectorParagraphFlowMode) {
	rotation := s.flowRotationAngle(mode)
	if floatEquals(rotation, s.flowRotation) {
		return
	}
	s.flowRotation = rotation
	s.localPolygon = s.buildLocalPolygon()
	s.localBounds = boundsForPoints(s.localPolygon)
	s.SetLeft(s.geometry.AnchorX + s.localBounds.MinX)
	s.SetTop(s.geometry.AnchorY + s.localBounds.MinY)
	s.SetWidth(s.localBounds.MaxX - s.localBounds.MinX)
	s.SetHeight(s.localBounds.MaxY - s.localBounds.MinY)
}

func (s *StdSector) toLocal(x, y float64) (float64, float64) {
	lx, ly := rotatePagePoint(x, y, s.geometry.AnchorX, s.geometry.AnchorY, -s.flowRotation)
	return lx - s.geometry.AnchorX, ly - s.geometry.AnchorY
}

func (s *StdSector) contentLocalBounds() radialBounds {
	return s.contentBounds
}

func (s *StdSector) contentLineIntervalAt(localY float64) radialInterval {
	if len(s.contentPolygon) < 3 {
		return radialInterval{}
	}
	centerX, _, ok := polygonCentroid(s.contentPolygon)
	if !ok {
		centerX = (s.contentBounds.MinX + s.contentBounds.MaxX) / 2
	}
	return polygonLineIntervalAt(s.contentPolygon, s.contentBounds, localY, centerX)
}

func (s *StdSector) contentLocalCenter() (float64, float64) {
	centroidX, centroidY, ok := polygonCentroid(s.contentPolygon)
	if !ok {
		bounds := s.contentLocalBounds()
		return (bounds.MinX + bounds.MaxX) / 2, (bounds.MinY + bounds.MaxY) / 2
	}
	bounds := s.contentLocalBounds()
	return clampFloat(centroidX, bounds.MinX, bounds.MaxX), clampFloat(centroidY, bounds.MinY, bounds.MaxY)
}

func (s *StdSector) contentPolarCenterLocalY() float64 {
	innerRadius, outerRadius := s.contentRadii()
	x, y := radialPointAt(s.geometry.CenterX, s.geometry.CenterY,
		(innerRadius+outerRadius)/2, s.geometry.AnchorAngle)
	_, localY := s.toLocal(x, y)
	return localY
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

func (s *StdSector) staticParagraphFlowMode() sectorParagraphFlowMode {
	curved, horizontal := false, false
	for _, child := range s.Widgets() {
		paragraph, ok := child.(*StdParagraph)
		if !ok || paragraph.Position() != Static || paragraph.Display() == DisplayNone {
			continue
		}
		if paragraph.horizontalInSector() {
			horizontal = true
		} else {
			curved = true
		}
	}
	if curved && horizontal {
		return sectorParagraphFlowMixed
	}
	if horizontal {
		return sectorParagraphFlowHorizontal
	}
	return sectorParagraphFlowRadial
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

var _ Container = (*StdSector)(nil)
var _ HasAttrs = (*StdSector)(nil)
var _ Identifier = (*StdSector)(nil)
var _ Printer = (*StdSector)(nil)
var _ WantsContainer = (*StdSector)(nil)
var _ sectorParagraphLayoutProvider = (*StdSector)(nil)
