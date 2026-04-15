// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"maps"
	"math"
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
	borders         [4]*PenStyle
	colSpan         int
	fill            *BrushStyle
	font            *FontStyle
	position        Position
	rowSpan         int
	align           Align
	selfAlign       SelfAlign
	rotate          *float64
	originX         string
	originY         string
	shiftX          float64
	shiftY          float64
	zIndex          int
	display         DisplayMode
	printed         bool
	invisible       bool
	disabled        bool
	path            string
	rawAttrs        map[string]string
}

type sectorReferenceResolver interface {
	ResolveSectorReferenceX(widget Widget) float64
	ResolveSectorReferenceY(widget Widget) float64
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
	x2 := widget.Right() - widget.MarginRight()
	y2 := widget.Bottom() - widget.MarginBottom()
	if widget.border != nil {
		widget.border.Apply(w)
		w.Rectangle2(x1, y1,
			widget.Width()-widget.MarginLeft()-widget.MarginRight(),
			widget.Height()-widget.MarginTop()-widget.MarginBottom(),
			true, false, widget.corners, false, false)
	}
	if widget.borders[topSide] != nil {
		widget.borders[topSide].Apply(w)
		w.MoveTo(x1, y1)
		w.LineTo(x2, y1)
	}
	if widget.borders[rightSide] != nil {
		widget.borders[rightSide].Apply(w)
		w.MoveTo(x2, y1)
		w.LineTo(x2, y2)
	}
	if widget.borders[bottomSide] != nil {
		widget.borders[bottomSide].Apply(w)
		w.MoveTo(x2, y2)
		w.LineTo(x1, y2)
	}
	if widget.borders[leftSide] != nil {
		widget.borders[leftSide].Apply(w)
		w.MoveTo(x1, y2)
		w.LineTo(x1, y1)
	}
	return nil
}

func (widget *StdWidget) Font() *FontStyle {
	if widget.font == nil {
		return widget.container.Font()
	}
	return widget.font
}

func (widget *StdWidget) LayoutWidget(w Writer) {
	// to be overridden
}

func (widget *StdWidget) PaintBackground(w Writer) error {
	if widget.fill == nil {
		return nil
	}
	x, y, width, height := widget.backgroundRect()
	if width <= 0 || height <= 0 {
		return nil
	}
	return widget.paintBrushInRect(w, widget.fill, x, y, width, height)
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

func (widget *StdWidget) OriginXToken() string {
	return widget.originX
}

func (widget *StdWidget) OriginYToken() string {
	return widget.originY
}

func (widget *StdWidget) Position() Position {
	return widget.position
}

func (widget *StdWidget) SetPosition(value Position) {
	widget.position = value
}

func (widget *StdWidget) PreferredHeight(Writer) float64 {
	return widget.Height()
}

func (widget *StdWidget) PreferredWidth(Writer) float64 {
	return widget.Width()
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

	if position, ok := attrs["position"]; ok {
		switch position {
		case "static":
			widget.position = Static
		case "relative":
			widget.position = Relative
		case "absolute":
			widget.position = Absolute
		}
	} else if MapHasAnyKey(attrs, "top", "right", "bottom", "left") {
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
	if border, ok := attrs["border"]; ok {
		widget.border = PenStyleFor(border, widget.scope)
	}
	if MapHasKeyPrefix(attrs, "border.") {
		if widget.border == nil {
			widget.border = &PenStyle{pattern: defaultPenPattern, cap: defaultPenCap}
		} else {
			widget.border = widget.border.Clone()
		}
		widget.border.SetAttrs("border.", attrs)
	}
	for i, side := range sideNames {
		if border, ok := attrs["border-"+side]; ok {
			widget.borders[i] = PenStyleFor(border, widget.scope)
		}
		prefix := "border-" + side + "."
		if MapHasKeyPrefix(attrs, prefix) {
			base := widget.borders[i]
			if base == nil {
				base = widget.border
			}
			if base == nil {
				widget.borders[i] = &PenStyle{pattern: defaultPenPattern, cap: defaultPenCap}
			} else {
				widget.borders[i] = base.Clone()
			}
			widget.borders[i].SetAttrs(prefix, attrs)
		}
	}
	if fill, ok := attrs["fill"]; ok {
		widget.fill = BrushStyleFor(fill, widget.scope)
	}
	if MapHasKeyPrefix(attrs, "fill.") {
		if widget.fill == nil {
			widget.fill = &BrushStyle{}
		} else {
			widget.fill = widget.fill.Clone()
		}
		widget.fill.SetAttrs(addUnits(filterMapAttrs("fill.", attrs), widget.Units()))
	}
	if font, ok := attrs["font"]; ok {
		widget.font = FontStyleFor(font, widget.scope)
	}
	if MapHasKeyPrefix(attrs, "font.") {
		baseFont := widget.font
		if baseFont == nil {
			if widget.container != nil {
				baseFont = widget.container.Font()
			} else {
				baseFont = defaultFont
			}
		}
		widget.font = baseFont.Clone()
		widget.font.SetScope(widget.scope)
		widget.font.SetAttrs("font.", normalizeFontDecorationMeasurementAttrs(attrs, "font.", widget.Units()))
	}
	if colSpan, ok := attrs["colspan"]; ok {
		widget.colSpan, _ = strconv.Atoi(colSpan)
	}
	if rowSpan, ok := attrs["rowspan"]; ok {
		widget.rowSpan, _ = strconv.Atoi(rowSpan)
	}
	if rotate, ok := attrs["rotate"]; ok {
		if value, err := strconv.ParseFloat(rotate, 64); err == nil {
			widget.rotate = &value
		}
	}
	if originX, ok := attrs["origin-x"]; ok {
		widget.originX = strings.TrimSpace(originX)
	}
	if originY, ok := attrs["origin-y"]; ok {
		widget.originY = strings.TrimSpace(originY)
	}
	if shift, ok := attrs["shift"]; ok {
		x, y := split2(shift, ",")
		widget.shiftX = ParseMeasurement(strings.TrimSpace(x), widget.Units())
		widget.shiftY = ParseMeasurement(strings.TrimSpace(y), widget.Units())
	}
	if zIndex, ok := attrs["z-index"]; ok {
		widget.zIndex, _ = strconv.Atoi(strings.TrimSpace(zIndex))
	}
	if display, ok := attrs["display"]; ok {
		widget.display = ParseDisplayMode(display)
	}
	if value, ok := attrs["alt"]; ok {
		widget.alt = value
	}
	if value, ok := attrs["role"]; ok {
		widget.role = strings.TrimSpace(value)
	}
}

func (widget *StdWidget) backgroundRect() (x, y, width, height float64) {
	return widget.Left() + widget.MarginLeft(),
		widget.Top() + widget.MarginTop(),
		widget.Width() - widget.MarginLeft() - widget.MarginRight(),
		widget.Height() - widget.MarginTop() - widget.MarginBottom()
}

// paintBrushInRect is the shared widget-box brush painter used by standard
// widgets that rely on StdWidget.PaintBackground. Widgets with custom fill
// geometry, such as sectors or shape primitives, continue to own their own
// background logic.
func (widget *StdWidget) paintBrushInRect(w Writer, brush *BrushStyle, x, y, width, height float64) error {
	if brush == nil {
		return nil
	}
	switch brush.Kind() {
	case BrushKindLinearGradient:
		if brush.linearGradient == nil {
			return nil
		}
		gradient := offsetLinearGradient(brush.linearGradient, x, y)
		return widget.paintClippedRect(w, x, y, width, height, func() error {
			return w.PaintLinearGradient(gradient)
		})
	case BrushKindRadialGradient:
		if brush.radialGradient == nil {
			return nil
		}
		gradient := offsetRadialGradient(brush.radialGradient, x, y)
		return widget.paintClippedRect(w, x, y, width, height, func() error {
			return w.PaintRadialGradient(gradient)
		})
	case BrushKindImage:
		return widget.paintImageBrushInRect(w, brush.image, x, y, width, height)
	default:
		brush.Apply(w)
		w.Rectangle2(x, y, width, height, false, true, widget.corners, false, false)
		return nil
	}
}

func (widget *StdWidget) paintClippedRect(w Writer, x, y, width, height float64, paint func() error) error {
	var paintErr error
	if err := w.Path(func() {
		w.Rectangle2(x, y, width, height, false, false, widget.corners, true, false)
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
			tileWidth, tileHeight = resolveBrushImageSize(fitMode, width, height, float64(imageWidth), float64(imageHeight))
			tileX, tileY = resolveBrushAnchor(normalizeBrushAnchor(image), x, y, width, height, tileWidth, tileHeight)
		}
	}
	if fitMode == "tile" && repeatMode == "no-repeat" {
		repeatMode = "repeat"
	}

	return widget.paintClippedRect(w, x, y, width, height, func() error {
		return paintRepeatedBrushImage(w, ref.identifier, tileX, tileY, tileWidth, tileHeight, x, y, width, height, repeatMode)
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

func paintRepeatedBrushImage(w Writer, filename string, tileX, tileY, tileWidth, tileHeight, clipX, clipY, clipWidth, clipHeight float64, repeatMode string) error {
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
			if err := w.PaintImageFile(filename, drawX, drawY, tileWidth, tileHeight); err != nil {
				return err
			}
		}
	}
	return nil
}

func offsetLinearGradient(source *pdf.LinearGradient, x, y float64) *pdf.LinearGradient {
	if source == nil {
		return nil
	}
	clone := *source
	clone.X0 += x
	clone.Y0 += y
	clone.X1 += x
	clone.Y1 += y
	clone.Stops = append([]pdf.GradientStop(nil), source.Stops...)
	return &clone
}

func offsetRadialGradient(source *pdf.RadialGradient, x, y float64) *pdf.RadialGradient {
	if source == nil {
		return nil
	}
	clone := *source
	clone.X0 += x
	clone.Y0 += y
	clone.X1 += x
	clone.Y1 += y
	clone.Stops = append([]pdf.GradientStop(nil), source.Stops...)
	return &clone
}

func normalizeFontDecorationMeasurementAttrs(attrs map[string]string, prefix string, units Units) map[string]string {
	normalized := maps.Clone(attrs)
	for _, key := range []string{"underline-pos", "strikeout-pos"} {
		fullKey := prefix + key
		value, ok := normalized[fullKey]
		if !ok {
			continue
		}
		if points := parseOptionalMeasurement(strings.TrimSpace(value), units); points != nil {
			normalized[fullKey] = strconv.FormatFloat(*points, 'f', -1, 64) + "pt"
		}
	}
	return normalized
}

func normalizeBrushMeasurementAttrs(attrs map[string]string, prefix string, units Units) map[string]string {
	normalized := maps.Clone(attrs)
	for _, key := range []string{"x0", "y0", "x1", "y1", "r0", "r1"} {
		fullKey := prefix + key
		value, ok := normalized[fullKey]
		if !ok {
			continue
		}
		normalized[fullKey] = strconv.FormatFloat(ParseMeasurement(strings.TrimSpace(value), units), 'f', -1, 64)
	}
	return normalized
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
	if widget.sides[topSide].IsSet {
		return widget.resolveTop(widget.sides[topSide].Value)
	}
	if !widget.sides[bottomSide].IsSet || widget.Height() == 0 {
		return 0
	}
	return widget.resolveBottom(widget.sides[bottomSide].Value) - widget.Height()
}

func (widget *StdWidget) Right() float64 {
	if widget.sides[rightSide].IsSet {
		return widget.resolveRight(widget.sides[rightSide].Value)
	}
	if !widget.sides[leftSide].IsSet || widget.Width() == 0 {
		return 0
	}
	return widget.resolveLeft(widget.sides[leftSide].Value) + widget.Width()
}

func (widget *StdWidget) Bottom() float64 {
	if widget.sides[bottomSide].IsSet {
		return widget.resolveBottom(widget.sides[bottomSide].Value)
	}
	if !widget.sides[topSide].IsSet || widget.Height() == 0 {
		return 0
	}
	return widget.resolveTop(widget.sides[topSide].Value) + widget.Height()
}

func (widget *StdWidget) Left() float64 {
	if widget.sides[leftSide].IsSet {
		return widget.resolveLeft(widget.sides[leftSide].Value)
	}
	if !widget.sides[rightSide].IsSet || widget.Width() == 0 {
		return 0
	}
	return widget.resolveRight(widget.sides[rightSide].Value) - widget.Width()
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
	if widget.widthPct > 0 {
		return widget.widthPct / 100.0 * ContentWidth(widget.container)
	}
	if widget.widthRel != 0 {
		return ContentWidth(widget.container) + widget.widthRel
	}
	if widget.widthSet {
		return widget.width
	}
	if widget.sides[leftSide].IsSet && widget.sides[rightSide].IsSet {
		return widget.resolveRight(widget.sides[rightSide].Value) - widget.resolveLeft(widget.sides[leftSide].Value)
	}
	return widget.width
}

func (widget *StdWidget) Height() float64 {
	if widget.heightPct > 0 {
		return widget.heightPct / 100.0 * ContentHeight(widget.container)
	}
	if widget.heightRel != 0 {
		return ContentHeight(widget.container) + widget.heightRel
	}
	if widget.heightSet {
		return widget.height
	}
	if widget.sides[topSide].IsSet && widget.sides[bottomSide].IsSet {
		return widget.resolveBottom(widget.sides[bottomSide].Value) - widget.resolveTop(widget.sides[topSide].Value)
	}
	return widget.height
}

func (widget *StdWidget) HeightIsSet() bool {
	return widget.heightSet || (widget.sides[topSide].IsSet && widget.sides[bottomSide].IsSet)
}

func (widget *StdWidget) WidthIsSet() bool {
	return widget.widthSet || (widget.sides[leftSide].IsSet && widget.sides[rightSide].IsSet)
}

func (widget *StdWidget) Visible() bool {
	return !widget.invisible
}

func (widget *StdWidget) ZIndex() int {
	return widget.zIndex
}

func (widget *StdWidget) Display() DisplayMode {
	if widget.display == "" {
		return DisplayOnce
	}
	return widget.display
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
	if widget.position == Relative && widget.container != nil {
		if resolver, ok := widget.container.(sectorReferenceResolver); ok {
			return resolver.ResolveSectorReferenceX(widget) + value + widget.shiftX
		}
	}
	if widget.container != nil && value < 0 {
		value = widget.container.Width() + value
	}
	if widget.position == Relative && widget.container != nil {
		value += widget.container.Left()
	}
	value += widget.shiftX
	return value
}

func (widget *StdWidget) resolveRight(value float64) float64 {
	if widget.position == Relative && widget.container != nil {
		if resolver, ok := widget.container.(sectorReferenceResolver); ok {
			return resolver.ResolveSectorReferenceX(widget) + value + widget.shiftX
		}
	}
	if widget.container != nil && value <= 0 {
		value = widget.container.Width() + value
	}
	if widget.position == Relative && widget.container != nil {
		value += widget.container.Left()
	}
	value += widget.shiftX
	return value
}

func (widget *StdWidget) resolveTop(value float64) float64 {
	if widget.position == Relative && widget.container != nil {
		if resolver, ok := widget.container.(sectorReferenceResolver); ok {
			return resolver.ResolveSectorReferenceY(widget) + value + widget.shiftY
		}
	}
	if widget.container != nil && value < 0 {
		value = widget.container.Height() + value
	}
	if widget.position == Relative && widget.container != nil {
		value += widget.container.Top()
	}
	value += widget.shiftY
	return value
}

func (widget *StdWidget) resolveBottom(value float64) float64 {
	if widget.position == Relative && widget.container != nil {
		if resolver, ok := widget.container.(sectorReferenceResolver); ok {
			return resolver.ResolveSectorReferenceY(widget) + value + widget.shiftY
		}
	}
	if widget.container != nil && value <= 0 {
		value = widget.container.Height() + value
	}
	if widget.position == Relative && widget.container != nil {
		value += widget.container.Top()
	}
	value += widget.shiftY
	return value
}

func (widget *StdWidget) paintWithTransform(w Writer, fn func() error) error {
	if widget.rotate == nil {
		return fn()
	}
	var renderErr error
	if err := w.Rotate(*widget.rotate, widget.OriginX(), widget.OriginY(), func() {
		renderErr = fn()
	}); err != nil {
		return err
	}
	return renderErr
}

func (widget *StdWidget) OriginX() float64 {
	if resolver, ok := widget.container.(sectorReferenceResolver); ok {
		switch widget.originX {
		case "start", "center", "end":
			return resolver.ResolveSectorReferenceX(widget)
		}
	}
	switch widget.originX {
	case "center":
		return (widget.Left() + widget.Right()) / 2
	case "right":
		return widget.Right()
	case "":
		return widget.Left()
	default:
		return ParseMeasurement(widget.originX, widget.Units())
	}
}

func (widget *StdWidget) OriginY() float64 {
	if resolver, ok := widget.container.(sectorReferenceResolver); ok {
		switch widget.originY {
		case "inner", "middle", "outer":
			return resolver.ResolveSectorReferenceY(widget)
		}
	}
	switch widget.originY {
	case "middle":
		return (widget.Top() + widget.Bottom()) / 2
	case "bottom":
		return widget.Bottom()
	case "":
		return widget.Top()
	default:
		return ParseMeasurement(widget.originY, widget.Units())
	}
}

func init() {
	registerTag(DefaultSpace, "widget", func() any { return &StdWidget{} })
}

var _ HasAttrs = (*StdWidget)(nil)
var _ Printer = (*StdWidget)(nil)
var _ WantsContainer = (*StdWidget)(nil)
var _ WantsScope = (*StdWidget)(nil)
var _ Widget = (*StdWidget)(nil)
