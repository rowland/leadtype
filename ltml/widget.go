// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type Widget struct {
	units     Units
	container Container
	scope     HasScope
	Identity
	Dimensions
	border  *PenStyle
	borders [4]*PenStyle
	font    *FontStyle
}

const defaultPenPattern = "solid"

func (widget *Widget) ContentTop() float64 {
	return widget.Top() + widget.MarginTop() + widget.PaddingTop()
}

func (widget *Widget) ContentRight() float64 {
	return widget.Right() - widget.MarginRight() - widget.PaddingRight()
}

func (widget *Widget) ContentBottom() float64 {
	return widget.Bottom() - widget.MarginBottom() - widget.PaddingBottom()
}

func (widget *Widget) ContentLeft() float64 {
	return widget.Left() + widget.MarginLeft() + widget.PaddingLeft()
}

func (widget *Widget) Font() *FontStyle {
	if widget.font == nil {
		return widget.container.Font()
	}
	return widget.font
}

func (widget *Widget) Print(w Writer) error {
	fmt.Printf("Printing %s\n", widget)
	return nil
}

func (widget *Widget) SetAttrs(attrs map[string]string) {
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

func (widget *Widget) SetContainer(container Container) error {
	widget.container = container
	return nil
}

func (widget *Widget) SetScope(scope HasScope) {
	widget.scope = scope
}

func (widget *Widget) String() string {
	return fmt.Sprintf("Widget %s units=%s %s %s %s", &widget.Identity, widget.units, &widget.Dimensions, widget.border, widget.borders)
}

func (widget *Widget) Units() Units {
	if widget.units == "" {
		return widget.container.Units()
	}
	return widget.units
}

var _ HasAttrs = (*Widget)(nil)
var _ Printer = (*Widget)(nil)
var _ WantsContainer = (*Widget)(nil)
var _ WantsScope = (*Widget)(nil)
