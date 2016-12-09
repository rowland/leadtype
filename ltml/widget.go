// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type Widget struct {
	AParent
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
		widget.border = PenStyleFor(border, ScopeFor(widget))
	}
	for i, side := range sideNames {
		if border, ok := attrs["border-"+side]; ok {
			widget.borders[i] = PenStyleFor(border, ScopeFor(widget))
		}
	}
}

func (widget *Widget) String() string {
	return fmt.Sprintf("Widget %s %s %s %s", &widget.Identity, &widget.Dimensions, widget.border, widget.borders)
}
