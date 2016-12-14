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

var _ HasAttrs = (*StdWidget)(nil)
var _ Printer = (*StdWidget)(nil)
var _ WantsContainer = (*StdWidget)(nil)
var _ WantsScope = (*StdWidget)(nil)
