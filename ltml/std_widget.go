// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdWidget struct {
	units     Units
	container Container
	scope     HasScope
	Identity
	Dimensions
	border  *PenStyle
	borders [4]*PenStyle
	fill    *BrushStyle
	font    *FontStyle
}

func (widget *StdWidget) BeforePrint(Writer) error {
	// to be overridden
	return nil
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

func (widget *StdWidget) Height() float64 {
	if widget.heightPct > 0 {
		return widget.heightPct * ContentHeight(widget.container)
	}
	return widget.height
}

func (widget *StdWidget) LayoutWidget(Writer) {
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

func (widget *StdWidget) PreferredHeight(Writer) float64 {
	return widget.Height()
}

func (widget *StdWidget) PreferredWidth(Writer) float64 {
	return widget.Width()
}

func (widget *StdWidget) Print(w Writer) error {
	fmt.Printf("Printing %s\n", widget)
	return nil
}

func (widget *StdWidget) SetAttrs(attrs map[string]string) {
	widget.units.SetAttrs(attrs)
	widget.Dimensions.SetAttrs(attrs, widget.Units())

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
	if font, ok := attrs["font"]; ok {
		widget.font = FontStyleFor(font, widget.scope)
	}
	if MapHasKeyPrefix(attrs, "font.") {
		widget.font = widget.Font().Clone()
		widget.font.SetAttrs("font.", attrs)
	}
}

func (widget *StdWidget) SetContainer(container Container) error {
	widget.container = container
	return nil
}

func (widget *StdWidget) SetScope(scope HasScope) {
	widget.scope = scope
}

func (widget *StdWidget) String() string {
	return fmt.Sprintf("Widget %s units=%s %s %s %s", &widget.Identity, widget.units, &widget.Dimensions, widget.border, widget.borders)
}

func (widget *StdWidget) Top() float64 {
	if widget.sides[topSide] != 0 {
		return widget.sides[topSide]
	}
	if widget.sides[bottomSide] == 0 || widget.Height() == 0 {
		return 0
	}
	return widget.sides[bottomSide] - widget.Height()
}

func (widget *StdWidget) Right() float64 {
	if widget.sides[rightSide] != 0 {
		return widget.sides[rightSide]
	}
	if widget.sides[leftSide] == 0 || widget.Width() == 0 {
		return 0
	}
	return widget.sides[leftSide] + widget.Width()
}

func (widget *StdWidget) Bottom() float64 {
	if widget.sides[bottomSide] != 0 {
		return widget.sides[bottomSide]
	}
	if widget.sides[topSide] == 0 || widget.Height() == 0 {
		return 0
	}
	return widget.sides[topSide] + widget.Height()
}

func (widget *StdWidget) Left() float64 {
	if widget.sides[leftSide] != 0 {
		return widget.sides[leftSide]
	}
	if widget.sides[rightSide] == 0 || widget.Width() == 0 {
		return 0
	}
	return widget.sides[rightSide] - widget.Width()
}

func (widget *StdWidget) Units() Units {
	if widget.units == "" && widget.container != nil {
		return widget.container.Units()
	}
	return widget.units
}

func (widget *StdWidget) Width() float64 {
	if widget.widthPct > 0 {
		return widget.widthPct * ContentWidth(widget.container)
	}
	return widget.width
}

func init() {
	registerTag(DefaultSpace, "widget", func() interface{} { return &StdWidget{} })
}

var _ HasAttrs = (*StdWidget)(nil)
var _ Printer = (*StdWidget)(nil)
var _ WantsContainer = (*StdWidget)(nil)
var _ WantsScope = (*StdWidget)(nil)
