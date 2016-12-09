// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"regexp"
)

var (
	re6DigitHexColor = regexp.MustCompile(`^#([0-9A-Fa-f]{6})$`)
	re3DigitHexColor = regexp.MustCompile(`^#([0-9A-Fa-f]{3})$`)
)

func validColor(color string) bool {
	return re6DigitHexColor.MatchString(color) || re3DigitHexColor.MatchString(color)
}

type Color string

func (c *Color) SetAttrs(attrs map[string]string) {
	if color, ok := attrs["color"]; ok {
		if validColor(color) {
			*c = Color(color)
		}
	}
}
