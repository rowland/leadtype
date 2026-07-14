// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type TextStyle struct {
	id           string
	textAlign    HAlign
	textAlignSet bool
	vAlign       VAlign
}

func (ts *TextStyle) Apply(w Writer) {
	// fmt.Printf("Applying %s\n", ts)
}

func (ts *TextStyle) ID() string {
	return ts.id
}

func (ts *TextStyle) SetAttrs(attrs map[string]string) {
	if id, ok := attrs["id"]; ok {
		ts.id = id
	}
	if textAlign, ok := attrs["text-align"]; ok {
		ts.textAlignSet = true
		ts.textAlign = parseTextAlign(textAlign, true)
	}
	if vAlign, ok := attrs["valign"]; ok {
		switch vAlign {
		case "top":
			ts.vAlign = VAlignTop
		case "middle":
			ts.vAlign = VAlignMiddle
		case "bottom":
			ts.vAlign = VAlignBottom
		case "baseline":
			ts.vAlign = VAlignBaseline
		}
	}
}

func (ts *TextStyle) ResolvedTextAlign(c Container) HAlign {
	if ts != nil && ts.textAlignSet {
		return resolveTextAlign(ts.textAlign, c)
	}
	return resolveTextAlign(textAlignStart, c)
}

func (ts *TextStyle) String() string {
	return fmt.Sprintf("id=%s text-align=%s valign=%s", ts.id, ts.textAlign, ts.vAlign)
}

var _ HasAttrs = (*TextStyle)(nil)
var _ Styler = (*TextStyle)(nil)
