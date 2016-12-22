// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type TextStyle struct {
	id        string
	textAlign HAlign
	vAlign    VAlign
}

func (ts *TextStyle) Apply(w Writer) {
	fmt.Printf("Applying %s\n", ts)
}

func (ts *TextStyle) ID() string {
	return ts.id
}

func (ts *TextStyle) SetAttrs(prefix string, attrs map[string]string) {
	if id, ok := attrs[prefix+"id"]; ok {
		ts.id = id
	}
	if textAlign, ok := attrs[prefix+"text-align"]; ok {
		switch textAlign {
		case "left":
			ts.textAlign = HAlignLeft
		case "center":
			ts.textAlign = HAlignCenter
		case "right":
			ts.textAlign = HAlignRight
		case "justify":
			ts.textAlign = HAlignJustify
		}
	}
	if vAlign, ok := attrs[prefix+"valign"]; ok {
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

func (ts *TextStyle) String() string {
	return fmt.Sprintf("id=%s text-align=%s valign=%s", ts.id, ts.textAlign, ts.vAlign)
}

var _ HasAttrsPrefix = (*TextStyle)(nil)
var _ Styler = (*TextStyle)(nil)
