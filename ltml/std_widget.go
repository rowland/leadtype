// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"
	"strings"
)

type StdWidget struct {
	units     Units
	container Container
	scope     HasScope
	Identity
	Dimensions
	border    *PenStyle
	borders   [4]*PenStyle
	colSpan   int
	fill      *BrushStyle
	font      *FontStyle
	position  Position
	rowSpan   int
	align     Align
	rotate    *float64
	originX   string
	originY   string
	shiftX    float64
	shiftY    float64
	zIndex    int
	display   DisplayMode
	printed   bool
	invisible bool
	disabled  bool
	path      string
}

func (widget *StdWidget) Align() Align {
	return widget.align
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
	if widget.fill != nil {
		widget.fill.Apply(w)
		w.Rectangle2(
			widget.Left()+widget.MarginLeft(),
			widget.Top()+widget.MarginTop(),
			widget.Width()-widget.MarginLeft()-widget.MarginRight(),
			widget.Height()-widget.MarginTop()-widget.MarginBottom(),
			false, true, widget.corners, false, false)
	}
	return nil
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
	if border, ok := attrs["border"]; ok {
		widget.border = PenStyleFor(border, widget.scope)
	}
	for i, side := range sideNames {
		if border, ok := attrs["border-"+side]; ok {
			widget.borders[i] = PenStyleFor(border, widget.scope)
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
		widget.fill.SetAttrs("fill.", attrs)
	}
	if font, ok := attrs["font"]; ok {
		widget.font = FontStyleFor(font, widget.scope)
	}
	if MapHasKeyPrefix(attrs, "font.") {
		widget.font = widget.Font().Clone()
		widget.font.SetAttrs("font.", attrs)
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
	if originX, ok := attrs["origin_x"]; ok {
		widget.originX = strings.TrimSpace(originX)
	}
	if originY, ok := attrs["origin_y"]; ok {
		widget.originY = strings.TrimSpace(originY)
	}
	if shift, ok := attrs["shift"]; ok {
		x, y := split2(shift, ",")
		widget.shiftX = ParseMeasurement(strings.TrimSpace(x), widget.Units())
		widget.shiftY = ParseMeasurement(strings.TrimSpace(y), widget.Units())
	}
	if zIndex, ok := attrs["z_index"]; ok {
		widget.zIndex, _ = strconv.Atoi(strings.TrimSpace(zIndex))
	}
	if display, ok := attrs["display"]; ok {
		widget.display = ParseDisplayMode(display)
	}
}

func (widget *StdWidget) SetContainer(container Container) error {
	widget.container = container
	return nil
}

func (widget *StdWidget) SetDisabled(value bool) {
	widget.disabled = value
}

func (widget *StdWidget) SetPrinted(value bool) {
	widget.printed = value
}

func (widget *StdWidget) SetScope(scope HasScope) {
	widget.scope = scope
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

func (widget *StdWidget) resolveLeft(value float64) float64 {
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
	registerTag(DefaultSpace, "widget", func() interface{} { return &StdWidget{} })
}

var _ HasAttrs = (*StdWidget)(nil)
var _ Printer = (*StdWidget)(nil)
var _ WantsContainer = (*StdWidget)(nil)
var _ WantsScope = (*StdWidget)(nil)
var _ Widget = (*StdWidget)(nil)
