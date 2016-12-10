// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type Widget struct {
	container Container
	scope     HasScope
	Identity
	Dimensions
	border  *PenStyle
	borders [4]*PenStyle
}

const defaultPenPattern = "solid"

func (widget *Widget) Print(w Writer) error {
	fmt.Printf("Printing %s\n", widget)
	return nil
}

func (widget *Widget) SetAttrs(attrs map[string]string) {
	widget.Dimensions.SetAttrs(attrs)

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
	return fmt.Sprintf("Widget %s %s %s %s", &widget.Identity, &widget.Dimensions, widget.border, widget.borders)
}

var _ HasAttrs = (*Widget)(nil)
var _ Printer = (*Widget)(nil)
var _ WantsContainer = (*Widget)(nil)
var _ WantsScope = (*Widget)(nil)
