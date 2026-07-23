// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"maps"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/rowland/leadtype/pdf"
)

type StdWidget struct {
	doc       *Doc
	units     Units
	container Container
	scope     HasScope
	Identity
	Dimensions
	alt             string
	role            string
	accessibilityID string
	border          *PenStyle
	borderSet       bool
	borders         [4]*PenStyle
	borderSideSet   [4]bool
	colSpan         int
	fill            *BrushStyle
	font            *FontStyle
	position        Position
	rowSpan         int
	align           Align
	selfAlign       SelfAlign
	rotate          float32
	originX         OriginX
	originY         OriginY
	originXValue    float32
	originYValue    float32
	shiftX          float32
	shiftY          float32
	shiftXMode      DimensionMode
	shiftYMode      DimensionMode
	zIndex          int
	display         DisplayMode
	displaySet      bool
	printed         bool
	invisible       bool
	disabled        bool
	path            string
	rawAttrs        map[string]string
}

type sectorPositionedPlacement struct {
	boxLeft, boxTop  float64
	anchorX, anchorY float64
}

type sectorPlacementResolver interface {
	ResolveSectorPlacement(widget *StdWidget) sectorPositionedPlacement
}

func (widget *StdWidget) Align() Align {
	return widget.align
}

func (widget *StdWidget) SelfAlign() SelfAlign {
	return widget.selfAlign
}

func (widget *StdWidget) BeforePrint(Writer) error {
	// to be overridden
	return nil
}

func (widget *StdWidget) ColSpan() int {
	if widget.colSpan < 1 {
		return 1
	}
	return widget.colSpan
}

func (widget *StdWidget) Container() Container {
	return widget.container
}

func (widget *StdWidget) sectorWidget() *StdWidget {
	return widget
}

func (widget *StdWidget) RowSpan() int {
	if widget.rowSpan < 1 {
		return 1
	}
	return widget.rowSpan
}

func (widget *StdWidget) Disabled() bool {
	return widget.disabled
}

func (widget *StdWidget) DrawContent(w Writer) error {
	return nil
}

func (widget *StdWidget) GetID() string {
	return widget.ID
}

func (widget *StdWidget) DrawBorder(w Writer) error {
	x1 := widget.Left() + widget.MarginLeft()
	y1 := widget.Top() + widget.MarginTop()
	width := widget.Width() - widget.MarginLeft() - widget.MarginRight()
	height := widget.Height() - widget.MarginTop() - widget.MarginBottom()
	return drawRectBorders(w, x1, y1, width, height, widget.corners.Float64sFor(width, height),
		widget.border, widget.borders, widget.borderSideSet)
}

func drawRectBorders(w Writer, x, y, width, height float64, corners []float64, aggregate *PenStyle, sides [4]*PenStyle, sideSet [4]bool) error {
	hasSideOverrides := false
	for i := range sides {
		hasSideOverrides = hasSideOverrides || sideSet[i] || sides[i] != nil
	}
	if !hasSideOverrides {
		if aggregate == nil {
			return nil
		}
		if err := aggregate.ApplyInRect(w, x, y, width, height); err != nil {
			return err
		}
		w.Rectangle2(x, y, width, height, true, false, corners, false, false)
		return nil
	}

	effective := [4]*PenStyle{}
	for i := range sides {
		effective[i] = aggregate
		if sideSet[i] || sides[i] != nil {
			effective[i] = sides[i]
		}
	}

	if effective[0] != nil && sameRenderedPen(effective[0], effective[1]) &&
		sameRenderedPen(effective[0], effective[2]) && sameRenderedPen(effective[0], effective[3]) {
		if err := effective[0].ApplyInRect(w, x, y, width, height); err != nil {
			return err
		}
		w.Rectangle2(x, y, width, height, true, false, corners, false, false)
		return nil
	}

	curves := rectBorderCurves(x, y, width, height, corners)
	visited := [4]bool{}
	for i := range effective {
		if visited[i] || effective[i] == nil || sameRenderedPen(effective[(i+3)%4], effective[i]) {
			continue
		}
		run := []int{i}
		visited[i] = true
		for next := (i + 1) % 4; next != i && sameRenderedPen(effective[next], effective[i]); next = (next + 1) % 4 {
			run = append(run, next)
			visited[next] = true
		}
		if err := drawRectBorderRun(w, x, y, width, height, effective, curves, run); err != nil {
			return err
		}
	}
	return nil
}

func sameRenderedPen(a, b *PenStyle) bool {
	if a == nil || b == nil {
		return a == b
	}
	aValue, bValue := *a, *b
	aValue.id, bValue.id = "", ""
	return reflect.DeepEqual(aValue, bValue)
}

type rectBorderCurve [4]pdf.Location

func drawRectBorderRun(w Writer, x, y, width, height float64, pens [4]*PenStyle, curves [4]rectBorderCurve, run []int) error {
	pen := pens[run[0]]
	if err := pen.ApplyInRect(w, x, y, width, height); err != nil {
		return err
	}

	var drawErr, strokeErr error
	if err := w.Path(func() {
		first := run[0]
		previous := (first + 3) % 4
		leading := curves[previous]
		if pens[previous] == nil {
			w.MoveTo(leading[0].X, leading[0].Y)
			drawErr = appendRectBorderCurve(w, leading)
		} else {
			_, second := splitRectBorderCurve(leading)
			w.MoveTo(second[0].X, second[0].Y)
			drawErr = appendRectBorderCurve(w, second)
		}
		if drawErr != nil {
			return
		}

		for position, edge := range run {
			curve := curves[edge]
			w.LineTo(curve[0].X, curve[0].Y)
			if position+1 < len(run) || pens[(edge+1)%4] == nil {
				drawErr = appendRectBorderCurve(w, curve)
			} else {
				firstHalf, _ := splitRectBorderCurve(curve)
				drawErr = appendRectBorderCurve(w, firstHalf)
			}
			if drawErr != nil {
				return
			}
		}
		strokeErr = w.Stroke()
	}); err != nil {
		return err
	}
	if drawErr != nil {
		return drawErr
	}
	return strokeErr
}

func appendRectBorderCurve(w Writer, curve rectBorderCurve) error {
	if curve[0] == curve[3] {
		w.LineTo(curve[3].X, curve[3].Y)
		return nil
	}
	return w.CurvePoints(curve[:])
}

func splitRectBorderCurve(curve rectBorderCurve) (rectBorderCurve, rectBorderCurve) {
	midpoint := func(a, b pdf.Location) pdf.Location {
		return pdf.Location{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
	}
	p01 := midpoint(curve[0], curve[1])
	p12 := midpoint(curve[1], curve[2])
	p23 := midpoint(curve[2], curve[3])
	p012 := midpoint(p01, p12)
	p123 := midpoint(p12, p23)
	p0123 := midpoint(p012, p123)
	return rectBorderCurve{curve[0], p01, p012, p0123},
		rectBorderCurve{p0123, p123, p23, curve[3]}
}

func rectBorderCurves(x, y, width, height float64, corners []float64) [4]rectBorderCurve {
	radii := rectBorderCornerRadii(corners)
	kappa := 4.0 / 3.0 * (math.Sqrt2 - 1.0)
	x2, y2 := x+width, y+height
	topLeft, topRight, bottomRight, bottomLeft := radii[0], radii[1], radii[2], radii[3]
	return [4]rectBorderCurve{
		{{X: x2 - topRight[0], Y: y}, {X: x2 - topRight[0] + kappa*topRight[0], Y: y}, {X: x2, Y: y + topRight[1] - kappa*topRight[1]}, {X: x2, Y: y + topRight[1]}},
		{{X: x2, Y: y2 - bottomRight[1]}, {X: x2, Y: y2 - bottomRight[1] + kappa*bottomRight[1]}, {X: x2 - bottomRight[0] + kappa*bottomRight[0], Y: y2}, {X: x2 - bottomRight[0], Y: y2}},
		{{X: x + bottomLeft[0], Y: y2}, {X: x + bottomLeft[0] - kappa*bottomLeft[0], Y: y2}, {X: x, Y: y2 - bottomLeft[1] + kappa*bottomLeft[1]}, {X: x, Y: y2 - bottomLeft[1]}},
		{{X: x, Y: y + topLeft[1]}, {X: x, Y: y + topLeft[1] - kappa*topLeft[1]}, {X: x + topLeft[0] - kappa*topLeft[0], Y: y}, {X: x + topLeft[0], Y: y}},
	}
}

func rectBorderCornerRadii(corners []float64) [4][2]float64 {
	var radii [4][2]float64
	switch len(corners) {
	case 1:
		for i := range radii {
			radii[i] = [2]float64{corners[0], corners[0]}
		}
	case 2:
		radii[0], radii[1] = [2]float64{corners[0], corners[0]}, [2]float64{corners[0], corners[0]}
		radii[2], radii[3] = [2]float64{corners[1], corners[1]}, [2]float64{corners[1], corners[1]}
	case 4:
		for i := range radii {
			radii[i] = [2]float64{corners[i], corners[i]}
		}
	case 8:
		for i := range radii {
			radii[i] = [2]float64{corners[i*2], corners[i*2+1]}
		}
	}
	return radii
}

func (widget *StdWidget) Font() *FontStyle {
	if widget.font == nil {
		return widget.container.Font()
	}
	return widget.font
}

func (widget *StdWidget) explicitFont() *FontStyle {
	return widget.font
}

func (widget *StdWidget) LayoutWidget(w Writer) error {
	// to be overridden
	return nil
}

func (widget *StdWidget) PaintBackground(w Writer) error {
	if widget.fill == nil {
		return nil
	}
	x, y, width, height := widget.backgroundRect()
	if width <= 0 || height <= 0 {
		return nil
	}
	return widget.PaintBrushInRect(w, widget.fill, x, y, width, height)
}

func (widget *StdWidget) Path() string {
	if widget.path == "" {
		if widget.container != nil {
			widget.path = widget.container.Path() + "/"
		}
		widget.path += widget.SelectorTag()
	}
	return widget.path
}

func (widget *StdWidget) SetPath(path string) {
	widget.path = path
}

func (widget *StdWidget) RawAttrs() map[string]string {
	return widget.rawAttrs
}

func (widget *StdWidget) SetRawAttrs(attrs map[string]string) {
	if len(attrs) == 0 {
		widget.rawAttrs = nil
		return
	}
	widget.rawAttrs = maps.Clone(attrs)
}

func (widget *StdWidget) OriginX() OriginX {
	return widget.originX
}

func (widget *StdWidget) OriginY() OriginY {
	return widget.originY
}

func (widget *StdWidget) Position() Position {
	return widget.position
}

func (widget *StdWidget) SetPosition(value Position) {
	widget.position = value
}

func (widget *StdWidget) PreferredHeight(Writer) (float64, error) {
	return widget.Height(), nil
}

func (widget *StdWidget) PreferredWidth(Writer) (float64, error) {
	return widget.Width(), nil
}

func (widget *StdWidget) Print(w Writer) error {
	return nil
}

func (widget *StdWidget) Printed() bool {
	return widget.printed
}

func (widget *StdWidget) SetAttrs(attrs map[string]string) {
	widget.units.SetAttrs(attrs)
	widget.Dimensions.SetAttrs(attrs, widget.Units())
	if sector, ok := widget.container.(*StdSector); ok {
		sector.setChildPositionAttrs(widget, attrs, widget.Units())
	}

	if position, ok := attrs["position"]; ok {
		switch position {
		case "static":
			widget.position = Static
		case "relative":
			widget.position = Relative
		case "absolute":
			widget.position = Absolute
		}
	} else if widget.isDirectSectorChild() &&
		MapHasAnyKey(attrs, "start", "end", "outer", "inner") {
		widget.position = Relative
	} else if !widget.isDirectSectorChild() && MapHasAnyKey(attrs, "top", "right", "bottom", "left") {
		// Match ERML continuity: positional attrs implicitly opt a widget into
		// positioned layout when position is otherwise omitted.
		widget.position = Relative
	}

	if align, ok := attrs["align"]; ok {
		switch align {
		case "left":
			widget.align = AlignLeft
		case "right":
			widget.align = AlignRight
		case "top":
			widget.align = AlignTop
		case "bottom":
			widget.align = AlignBottom
		}
	}
	if alignSelf, ok := attrs["align-self"]; ok {
		switch alignSelf {
		case "start":
			widget.selfAlign = SelfAlignStart
		case "center":
			widget.selfAlign = SelfAlignCenter
		case "end":
			widget.selfAlign = SelfAlignEnd
		}
	}
	widget.setResourceAttrs(attrs, widget.Units())
	if colSpan, ok := attrs["colspan"]; ok {
		widget.colSpan, _ = strconv.Atoi(colSpan)
	}
	if rowSpan, ok := attrs["rowspan"]; ok {
		widget.rowSpan, _ = strconv.Atoi(rowSpan)
	}
	if rotate, ok := attrs["rotate"]; ok {
		if value, err := strconv.ParseFloat(rotate, 64); err == nil {
			widget.rotate = float32(value)
		}
	}
	if originX, ok := attrs["origin-x"]; ok {
		widget.originX, widget.originXValue = parseOriginX(strings.TrimSpace(originX), widget.Units())
	}
	if originY, ok := attrs["origin-y"]; ok {
		widget.originY, widget.originYValue = parseOriginY(strings.TrimSpace(originY), widget.Units())
	}
	if shiftX, ok := attrs["shift-x"]; ok {
		widget.shiftX, widget.shiftXMode = parseShift(strings.TrimSpace(shiftX), widget.Units())
	}
	if shiftY, ok := attrs["shift-y"]; ok {
		widget.shiftY, widget.shiftYMode = parseShift(strings.TrimSpace(shiftY), widget.Units())
	}
	if zIndex, ok := attrs["z-index"]; ok {
		widget.zIndex, _ = strconv.Atoi(strings.TrimSpace(zIndex))
	}
	if display, ok := attrs["display"]; ok {
		widget.display = ParseDisplayMode(display)
		widget.displaySet = true
	}
	if value, ok := attrs["alt"]; ok {
		widget.alt = value
	}
	if value, ok := attrs["role"]; ok {
		widget.role = strings.TrimSpace(value)
	}
}

func (widget *StdWidget) resetResourceAttrs() {
	widget.border = nil
	widget.borderSet = false
	widget.borders = [4]*PenStyle{}
	widget.borderSideSet = [4]bool{}
	widget.fill = nil
	widget.font = nil
}

func (widget *StdWidget) setResourceAttrs(attrs map[string]string, units Units) {
	setOptionalPenStyle(&widget.border, &widget.borderSet, "border", attrs, widget.scope, units, nil)
	for i, side := range sideNames {
		setOptionalPenStyle(&widget.borders[i], &widget.borderSideSet[i], "border-"+side,
			attrs, widget.scope, units, widget.border)
	}
	SetBrushStyle(&widget.fill, "fill", attrs, widget.scope, units)
	SetFontStyle(&widget.font, "font", attrs, widget.scope, units, widget.container)
}

func (widget *StdWidget) backgroundRect() (x, y, width, height float64) {
	return widget.Left() + widget.MarginLeft(),
		widget.Top() + widget.MarginTop(),
		widget.Width() - widget.MarginLeft() - widget.MarginRight(),
		widget.Height() - widget.MarginTop() - widget.MarginBottom()
}

// PaintBrushInRect paints brush into the rectangle, including solid, gradient,
// and image kinds. Custom widgets embedded in other packages should use this
// instead of BrushStyle.Apply (which only handles solid fills). It is the
// shared widget-box brush painter used by standard widgets that rely on
// StdWidget.PaintBackground. Widgets with custom fill geometry, such as
// sectors or shape primitives, continue to own their own background logic.
func (widget *StdWidget) PaintBrushInRect(w Writer, brush *BrushStyle, x, y, width, height float64) error {
	if brush == nil {
		return nil
	}
	switch brush.Kind() {
	case BrushKindLinearGradient:
		if brush.linearGradient == nil {
			return nil
		}
		opacity := normalizeBrushOpacityValue(brush.opacity)
		if opacity <= 0 {
			return nil
		}
		gradient := resolveLinearGradient(brush.linearGradient, brush.linearPct, x, y, width, height)
		gradient.Opacity = opacity
		return widget.paintClippedRect(w, x, y, width, height, func() error {
			return w.PaintLinearGradient(gradient)
		})
	case BrushKindRadialGradient:
		if brush.radialGradient == nil {
			return nil
		}
		opacity := normalizeBrushOpacityValue(brush.opacity)
		if opacity <= 0 {
			return nil
		}
		gradient := resolveRadialGradient(brush.radialGradient, brush.radialPct, x, y, width, height)
		gradient.Opacity = opacity
		return widget.paintClippedRect(w, x, y, width, height, func() error {
			return w.PaintRadialGradient(gradient)
		})
	case BrushKindSweepGradient:
		return fmt.Errorf("sweep-gradient brushes are supported only on sectors")
	case BrushKindImage:
		return widget.paintImageBrushInRect(w, brush.image, x, y, width, height)
	default:
		brush.Apply(w)
		w.Rectangle2(x, y, width, height, false, true, widget.corners.Float64sFor(width, height), false, false)
		return nil
	}
}

func (widget *StdWidget) paintClippedRect(w Writer, x, y, width, height float64, paint func() error) error {
	var paintErr error
	if err := w.Path(func() {
		w.Rectangle2(x, y, width, height, false, false, widget.corners.Float64sFor(width, height), true, false)
		if err := w.Clip(func() {
			paintErr = paint()
		}); err != nil && paintErr == nil {
			paintErr = err
		}
	}); err != nil {
		return err
	}
	if paintErr != nil {
		return paintErr
	}
	return nil
}

func (widget *StdWidget) paintImageBrushInRect(w Writer, image *BrushImageStyle, x, y, width, height float64) error {
	if image == nil || strings.TrimSpace(image.Src) == "" {
		return fmt.Errorf("brush image src must be specified")
	}
	opacity := normalizeBrushOpacity(image)
	if opacity <= 0 {
		return nil
	}
	ref, err := widget.resolveBrushImageSource(image.Src)
	if err != nil {
		return err
	}
	if ref.identifier == "" {
		return fmt.Errorf("brush image src %q did not resolve to an asset", image.Src)
	}

	tileX, tileY, tileWidth, tileHeight := x, y, width, height
	repeatMode := normalizeBrushRepeat(image)
	fitMode := normalizeBrushFit(image)

	if fitMode != "stretch" || repeatMode != "no-repeat" {
		imageWidth, imageHeight, err := w.ImageDimensionsFromFile(ref.identifier)
		if err == nil && imageWidth > 0 && imageHeight > 0 {
			if fitMode == "tile" {
				if explicitWidth, explicitHeight, ok := resolveBrushTileSize(image, width, height, float64(imageWidth), float64(imageHeight)); ok {
					tileWidth, tileHeight = explicitWidth, explicitHeight
				} else {
					tileWidth, tileHeight = resolveBrushImageSize(fitMode, width, height, float64(imageWidth), float64(imageHeight))
				}
			} else {
				tileWidth, tileHeight = resolveBrushImageSize(fitMode, width, height, float64(imageWidth), float64(imageHeight))
			}
			tileX, tileY = resolveBrushAnchor(normalizeBrushAnchor(image), x, y, width, height, tileWidth, tileHeight)
		}
	}
	if fitMode == "tile" && repeatMode == "no-repeat" {
		repeatMode = "repeat"
	}

	return widget.paintClippedRect(w, x, y, width, height, func() error {
		return paintRepeatedBrushImage(w, ref.identifier, tileX, tileY, tileWidth, tileHeight, x, y, width, height, repeatMode, opacity)
	})
}

func (widget *StdWidget) resolveBrushImageSource(src string) (assetSourceRef, error) {
	if widget.doc != nil {
		return widget.doc.resolveAssetSource(widget.container, src)
	}
	return resolveAssetSourceRef(nil, widget.container, src)
}

func normalizeBrushFit(image *BrushImageStyle) string {
	if image == nil || strings.TrimSpace(image.Fit) == "" {
		return "stretch"
	}
	return strings.TrimSpace(strings.ToLower(image.Fit))
}

func normalizeBrushAnchor(image *BrushImageStyle) string {
	if image == nil || strings.TrimSpace(image.Anchor) == "" {
		return "center"
	}
	return strings.TrimSpace(strings.ToLower(image.Anchor))
}

func normalizeBrushRepeat(image *BrushImageStyle) string {
	if image == nil || strings.TrimSpace(image.Repeat) == "" {
		return "no-repeat"
	}
	return strings.TrimSpace(strings.ToLower(image.Repeat))
}

func normalizeBrushOpacity(image *BrushImageStyle) float64 {
	if image == nil {
		return 1
	}
	return normalizeBrushOpacityValue(&image.Opacity)
}

func normalizeBrushOpacityValue(opacity *float64) float64 {
	if opacity == nil {
		return 1
	}
	switch {
	case *opacity <= 0:
		return 0
	case *opacity >= 1:
		return 1
	default:
		return *opacity
	}
}

func resolveBrushImageSize(fit string, boxWidth, boxHeight, imageWidth, imageHeight float64) (width, height float64) {
	if imageWidth <= 0 || imageHeight <= 0 {
		return boxWidth, boxHeight
	}
	switch fit {
	case "contain":
		scale := math.Min(boxWidth/imageWidth, boxHeight/imageHeight)
		return imageWidth * scale, imageHeight * scale
	case "cover":
		scale := math.Max(boxWidth/imageWidth, boxHeight/imageHeight)
		return imageWidth * scale, imageHeight * scale
	case "tile":
		return imageWidth, imageHeight
	default:
		return boxWidth, boxHeight
	}
}

func resolveBrushTileSize(image *BrushImageStyle, boxWidth, boxHeight, imageWidth, imageHeight float64) (width, height float64, ok bool) {
	if image == nil {
		return 0, 0, false
	}
	resolvedWidth := image.TileWidth
	if image.TileWidthPct > 0 {
		resolvedWidth = boxWidth * image.TileWidthPct / 100.0
	}
	resolvedHeight := image.TileHeight
	if image.TileHeightPct > 0 {
		resolvedHeight = boxHeight * image.TileHeightPct / 100.0
	}
	switch {
	case resolvedWidth > 0 && resolvedHeight > 0:
		return resolvedWidth, resolvedHeight, true
	case resolvedWidth > 0 && imageHeight > 0:
		return resolvedWidth, resolvedWidth * imageHeight / imageWidth, true
	case resolvedHeight > 0 && imageWidth > 0:
		return resolvedHeight * imageWidth / imageHeight, resolvedHeight, true
	default:
		return 0, 0, false
	}
}

func resolveBrushAnchor(anchor string, x, y, width, height, contentWidth, contentHeight float64) (drawX, drawY float64) {
	switch anchor {
	case "top":
		return x + (width-contentWidth)/2, y
	case "bottom":
		return x + (width-contentWidth)/2, y + height - contentHeight
	case "left":
		return x, y + (height-contentHeight)/2
	case "right":
		return x + width - contentWidth, y + (height-contentHeight)/2
	case "top-left":
		return x, y
	case "top-right":
		return x + width - contentWidth, y
	case "bottom-left":
		return x, y + height - contentHeight
	case "bottom-right":
		return x + width - contentWidth, y + height - contentHeight
	default:
		return x + (width-contentWidth)/2, y + (height-contentHeight)/2
	}
}

func paintRepeatedBrushImage(w Writer, filename string, tileX, tileY, tileWidth, tileHeight, clipX, clipY, clipWidth, clipHeight float64, repeatMode string, opacity float64) error {
	if tileWidth <= 0 || tileHeight <= 0 {
		return nil
	}
	repeatX := repeatMode == "repeat" || repeatMode == "repeat-x"
	repeatY := repeatMode == "repeat" || repeatMode == "repeat-y"

	startX := tileX
	startY := tileY
	if repeatX {
		for startX > clipX {
			startX -= tileWidth
		}
	}
	if repeatY {
		for startY > clipY {
			startY -= tileHeight
		}
	}

	endX := tileX + tileWidth
	endY := tileY + tileHeight
	if repeatX {
		endX = clipX + clipWidth
	}
	if repeatY {
		endY = clipY + clipHeight
	}

	for drawY := startY; drawY < endY; drawY += tileHeight {
		if !repeatY && drawY != startY {
			break
		}
		for drawX := startX; drawX < endX; drawX += tileWidth {
			if !repeatX && drawX != startX {
				break
			}
			if err := w.PaintImageFile(filename, drawX, drawY, tileWidth, tileHeight, opacity); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveLinearGradient(source *pdf.LinearGradient, pct *linearGradientPct, x, y, width, height float64) *pdf.LinearGradient {
	if source == nil {
		return nil
	}
	clone := *source
	clone.X0 += x
	clone.Y0 += y
	clone.X1 += x
	clone.Y1 += y
	if pct != nil {
		if pct.X0 != nil {
			clone.X0 = x + width*(*pct.X0)/100.0
		}
		if pct.Y0 != nil {
			clone.Y0 = y + height*(*pct.Y0)/100.0
		}
		if pct.X1 != nil {
			clone.X1 = x + width*(*pct.X1)/100.0
		}
		if pct.Y1 != nil {
			clone.Y1 = y + height*(*pct.Y1)/100.0
		}
	}
	clone.Stops = append([]pdf.GradientStop(nil), source.Stops...)
	return &clone
}

func resolveRadialGradient(source *pdf.RadialGradient, pct *radialGradientPct, x, y, width, height float64) *pdf.RadialGradient {
	if source == nil {
		return nil
	}
	clone := *source
	clone.X0 += x
	clone.Y0 += y
	clone.X1 += x
	clone.Y1 += y
	if pct != nil {
		if pct.X0 != nil {
			clone.X0 = x + width*(*pct.X0)/100.0
		}
		if pct.Y0 != nil {
			clone.Y0 = y + height*(*pct.Y0)/100.0
		}
		if pct.X1 != nil {
			clone.X1 = x + width*(*pct.X1)/100.0
		}
		if pct.Y1 != nil {
			clone.Y1 = y + height*(*pct.Y1)/100.0
		}
		boxMin := math.Min(width, height)
		if pct.R0 != nil {
			clone.R0 = boxMin * (*pct.R0) / 100.0
		}
		if pct.R1 != nil {
			clone.R1 = boxMin * (*pct.R1) / 100.0
		}
	}
	clone.Stops = append([]pdf.GradientStop(nil), source.Stops...)
	return &clone
}

func (widget *StdWidget) SetContainer(container Container) error {
	widget.container = container
	return nil
}

func (widget *StdWidget) SetDisabled(value bool) {
	widget.disabled = value
}

func (widget *StdWidget) SetDoc(doc *Doc) {
	widget.doc = doc
}

func (widget *StdWidget) SetPrinted(value bool) {
	widget.printed = value
}

func (widget *StdWidget) SetScope(scope HasScope) {
	widget.scope = scope
}

func (widget *StdWidget) Scope() HasScope {
	return widget.scope
}

func (widget *StdWidget) SetVisible(value bool) {
	widget.invisible = !value
}

func (widget *StdWidget) String() string {
	return fmt.Sprintf("Widget %s units=%s %s %s %s", &widget.Identity, widget.units, &widget.Dimensions, widget.border, widget.borders)
}

func (widget *StdWidget) Top() float64 {
	if placement, ok := widget.relativeSectorPlacement(); ok {
		return placement.boxTop
	}
	if widget.sides[topSide].IsSet {
		return widget.resolveTop(widget.sides[topSide].Float64())
	}
	if !widget.sides[bottomSide].IsSet || widget.Height() == 0 {
		return 0
	}
	return widget.resolveBottom(widget.sides[bottomSide].Float64()) - widget.Height()
}

func (widget *StdWidget) Right() float64 {
	if placement, ok := widget.relativeSectorPlacement(); ok {
		return placement.boxLeft + widget.Width()
	}
	if widget.sides[rightSide].IsSet {
		return widget.resolveRight(widget.sides[rightSide].Float64())
	}
	if !widget.sides[leftSide].IsSet || widget.Width() == 0 {
		return 0
	}
	return widget.resolveLeft(widget.sides[leftSide].Float64()) + widget.Width()
}

func (widget *StdWidget) Bottom() float64 {
	if placement, ok := widget.relativeSectorPlacement(); ok {
		return placement.boxTop + widget.Height()
	}
	if widget.sides[bottomSide].IsSet {
		return widget.resolveBottom(widget.sides[bottomSide].Float64())
	}
	if !widget.sides[topSide].IsSet || widget.Height() == 0 {
		return 0
	}
	return widget.resolveTop(widget.sides[topSide].Float64()) + widget.Height()
}

func (widget *StdWidget) Left() float64 {
	if placement, ok := widget.relativeSectorPlacement(); ok {
		return placement.boxLeft
	}
	if widget.sides[leftSide].IsSet {
		return widget.resolveLeft(widget.sides[leftSide].Float64())
	}
	if !widget.sides[rightSide].IsSet || widget.Width() == 0 {
		return 0
	}
	return widget.resolveRight(widget.sides[rightSide].Float64()) - widget.Width()
}

func (widget *StdWidget) TopIsSet() bool {
	return widget.sides[topSide].IsSet
}

func (widget *StdWidget) RightIsSet() bool {
	return widget.sides[rightSide].IsSet
}

func (widget *StdWidget) BottomIsSet() bool {
	return widget.sides[bottomSide].IsSet
}

func (widget *StdWidget) LeftIsSet() bool {
	return widget.sides[leftSide].IsSet
}

func (widget *StdWidget) Units() Units {
	if widget.units == "" && widget.container != nil {
		return widget.container.Units()
	}
	return widget.units
}

func (widget *StdWidget) Width() float64 {
	return widget.capWidth(widget.uncappedWidth())
}

func (widget *StdWidget) uncappedWidth() float64 {
	return widget.uncappedWidthWithResolver(widget.resolveLeft, widget.resolveRight)
}

func (widget *StdWidget) uncappedWidthWithoutShift() float64 {
	return widget.uncappedWidthWithResolver(widget.resolveLeftWithoutShift, widget.resolveRightWithoutShift)
}

func (widget *StdWidget) uncappedWidthWithResolver(resolveLeft, resolveRight func(float64) float64) float64 {
	if widget.widthValid {
		return float64(widget.width)
	}
	switch widget.widthMode {
	case DimPct:
		return float64(widget.widthValue) / 100.0 * ContentWidth(widget.container)
	case DimRel:
		return ContentWidth(widget.container) + float64(widget.widthValue)
	case DimLiteral:
		return float64(widget.widthValue)
	}
	if !widget.isRelativeSectorChild() &&
		widget.sides[leftSide].IsSet && widget.sides[rightSide].IsSet {
		return resolveRight(widget.sides[rightSide].Float64()) - resolveLeft(widget.sides[leftSide].Float64())
	}
	return 0
}

func (widget *StdWidget) Height() float64 {
	return widget.capHeight(widget.uncappedHeight())
}

func (widget *StdWidget) uncappedHeight() float64 {
	return widget.uncappedHeightWithResolver(widget.resolveTop, widget.resolveBottom)
}

func (widget *StdWidget) uncappedHeightWithoutShift() float64 {
	return widget.uncappedHeightWithResolver(widget.resolveTopWithoutShift, widget.resolveBottomWithoutShift)
}

func (widget *StdWidget) uncappedHeightWithResolver(resolveTop, resolveBottom func(float64) float64) float64 {
	if widget.heightValid {
		return float64(widget.height)
	}
	switch widget.heightMode {
	case DimPct:
		return float64(widget.heightValue) / 100.0 * ContentHeight(widget.container)
	case DimRel:
		return ContentHeight(widget.container) + float64(widget.heightValue)
	case DimLiteral:
		return float64(widget.heightValue)
	}
	if !widget.isRelativeSectorChild() &&
		widget.sides[topSide].IsSet && widget.sides[bottomSide].IsSet {
		return resolveBottom(widget.sides[bottomSide].Float64()) - resolveTop(widget.sides[topSide].Float64())
	}
	return 0
}

func (widget *StdWidget) MaxWidth() float64 {
	switch widget.max.widthMode {
	case DimPct:
		if widget.container == nil {
			return 0
		}
		return float64(widget.max.widthValue) / 100.0 * ContentWidth(widget.container)
	case DimRel:
		if widget.container == nil {
			return float64(widget.max.widthValue)
		}
		return ContentWidth(widget.container) + float64(widget.max.widthValue)
	case DimLiteral:
		return float64(widget.max.widthValue)
	default:
		return 0
	}
}

func (widget *StdWidget) MaxHeight() float64 {
	switch widget.max.heightMode {
	case DimPct:
		if widget.container == nil {
			return 0
		}
		return float64(widget.max.heightValue) / 100.0 * ContentHeight(widget.container)
	case DimRel:
		if widget.container == nil {
			return float64(widget.max.heightValue)
		}
		return ContentHeight(widget.container) + float64(widget.max.heightValue)
	case DimLiteral:
		return float64(widget.max.heightValue)
	default:
		return 0
	}
}

func (widget *StdWidget) capWidth(width float64) float64 {
	if widget.MaxWidthIsSet() {
		width = min(width, widget.MaxWidth())
	}
	return max(width, 0)
}

func (widget *StdWidget) capHeight(height float64) float64 {
	if widget.MaxHeightIsSet() {
		height = min(height, widget.MaxHeight())
	}
	return max(height, 0)
}

func (widget *StdWidget) HeightIsSet() bool {
	return widget.Dimensions.HeightIsSet() || (widget.sides[topSide].IsSet && widget.sides[bottomSide].IsSet)
}

func (widget *StdWidget) WidthIsSet() bool {
	return widget.Dimensions.WidthIsSet() || (widget.sides[leftSide].IsSet && widget.sides[rightSide].IsSet)
}

func (widget *StdWidget) Visible() bool {
	return !widget.invisible
}

func (widget *StdWidget) ZIndex() int {
	return widget.zIndex
}

func (widget *StdWidget) Display() DisplayMode {
	return widget.display
}

func (widget *StdWidget) DisplayExplicit() bool {
	return widget.displaySet
}

func (widget *StdWidget) AccessibilityLogicalID() string {
	if widget.accessibilityID == "" {
		widget.accessibilityID = fmt.Sprintf("%T:%p", widget, widget)
	}
	return widget.accessibilityID
}

func (widget *StdWidget) AccessibilityAlt() string {
	return widget.alt
}

func (widget *StdWidget) AccessibilityRole() string {
	return widget.role
}

func (widget *StdWidget) resolveLeft(value float64) float64 {
	return widget.resolveLeftWithoutShift(value) + widget.shiftXOffset()
}

func (widget *StdWidget) resolveLeftWithoutShift(value float64) float64 {
	if widget.container != nil && value < 0 {
		value = widget.container.Width() + value
	}
	if widget.position == Relative && widget.container != nil {
		value += widget.container.Left()
	}
	return value
}

func (widget *StdWidget) resolveRight(value float64) float64 {
	return widget.resolveRightWithoutShift(value) + widget.shiftXOffset()
}

func (widget *StdWidget) resolveRightWithoutShift(value float64) float64 {
	if widget.container != nil && value <= 0 {
		value = widget.container.Width() + value
	}
	if widget.position == Relative && widget.container != nil {
		value += widget.container.Left()
	}
	return value
}

func (widget *StdWidget) resolveTop(value float64) float64 {
	return widget.resolveTopWithoutShift(value) + widget.shiftYOffset()
}

func (widget *StdWidget) resolveTopWithoutShift(value float64) float64 {
	if widget.container != nil && value < 0 {
		value = widget.container.Height() + value
	}
	if widget.position == Relative && widget.container != nil {
		value += widget.container.Top()
	}
	return value
}

func (widget *StdWidget) resolveBottom(value float64) float64 {
	return widget.resolveBottomWithoutShift(value) + widget.shiftYOffset()
}

func (widget *StdWidget) resolveBottomWithoutShift(value float64) float64 {
	if widget.container != nil && value <= 0 {
		value = widget.container.Height() + value
	}
	if widget.position == Relative && widget.container != nil {
		value += widget.container.Top()
	}
	return value
}

func (widget *StdWidget) shiftXOffset() float64 {
	if widget.shiftXMode == DimPct {
		return float64(widget.shiftX) / 100.0 * widget.capWidth(widget.uncappedWidthWithoutShift())
	}
	return float64(widget.shiftX)
}

func (widget *StdWidget) shiftYOffset() float64 {
	if widget.shiftYMode == DimPct {
		return float64(widget.shiftY) / 100.0 * widget.capHeight(widget.uncappedHeightWithoutShift())
	}
	return float64(widget.shiftY)
}

func (widget *StdWidget) paintWithTransform(w Writer, fn func() error) error {
	if widget.rotate == 0 {
		return fn()
	}
	var renderErr error
	if err := w.Rotate(float64(widget.rotate), widget.OriginXValue(), widget.OriginYValue(), func() {
		renderErr = fn()
	}); err != nil {
		return err
	}
	return renderErr
}

func (widget *StdWidget) OriginXValue() float64 {
	if widget.originX == OriginXCustom {
		return float64(widget.originXValue)
	}
	if placement, ok := widget.relativeSectorPlacement(); ok {
		return placement.anchorX
	}
	switch widget.originX {
	case OriginXUnspecified, OriginXStart:
		return widget.Left()
	case OriginXCenter:
		return (widget.Left() + widget.Right()) / 2
	case OriginXEnd:
		return widget.Right()
	}
	return widget.Left()
}

func (widget *StdWidget) OriginYValue() float64 {
	if widget.originY == OriginYCustom {
		return float64(widget.originYValue)
	}
	if placement, ok := widget.relativeSectorPlacement(); ok {
		return placement.anchorY
	}
	switch widget.originY {
	case OriginYUnspecified, OriginYTop:
		return widget.Top()
	case OriginYMiddle:
		return (widget.Top() + widget.Bottom()) / 2
	case OriginYBottom:
		return widget.Bottom()
	}
	return widget.Top()
}

func (widget *StdWidget) relativeSectorPlacement() (sectorPositionedPlacement, bool) {
	if !widget.isRelativeSectorChild() {
		return sectorPositionedPlacement{}, false
	}
	resolver := widget.container.(sectorPlacementResolver)
	return resolver.ResolveSectorPlacement(widget), true
}

func (widget *StdWidget) isRelativeSectorChild() bool {
	if widget.position != Relative || widget.container == nil {
		return false
	}
	_, ok := widget.container.(sectorPlacementResolver)
	return ok
}

func (widget *StdWidget) isDirectSectorChild() bool {
	_, ok := widget.container.(*StdSector)
	return ok
}

func parseOriginX(token string, units Units) (OriginX, float32) {
	switch token {
	case "":
		return OriginXUnspecified, 0
	case "start":
		return OriginXStart, 0
	case "center":
		return OriginXCenter, 0
	case "end":
		return OriginXEnd, 0
	default:
		return OriginXCustom, float32(ParseMeasurement(token, units))
	}
}

func parseOriginY(token string, units Units) (OriginY, float32) {
	switch token {
	case "":
		return OriginYUnspecified, 0
	case "top", "inner":
		return OriginYTop, 0
	case "middle":
		return OriginYMiddle, 0
	case "bottom", "outer":
		return OriginYBottom, 0
	default:
		return OriginYCustom, float32(ParseMeasurement(token, units))
	}
}

func parseShift(token string, units Units) (float32, DimensionMode) {
	if strings.HasSuffix(token, "%") {
		value, err := strconv.ParseFloat(strings.TrimSpace(token[:len(token)-1]), 64)
		if err == nil {
			return float32(value), DimPct
		}
		return 0, DimLiteral
	}
	return float32(ParseMeasurement(token, units)), DimLiteral
}

func init() {
	registerTag(DefaultSpace, "widget", func() any { return &StdWidget{} })
}

var _ HasAttrs = (*StdWidget)(nil)
var _ Printer = (*StdWidget)(nil)
var _ WantsContainer = (*StdWidget)(nil)
var _ WantsScope = (*StdWidget)(nil)
var _ Widget = (*StdWidget)(nil)
