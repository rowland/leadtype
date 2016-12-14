// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"regexp"
)

var (
	re6DigitHexColor = regexp.MustCompile(`#([0-9A-Fa-f]{6})`)
)

type Color string

func (c *Color) SetAttrs(attrs map[string]string) {
	if color, ok := attrs["color"]; ok {
		if matches := re6DigitHexColor.FindStringSubmatch(color); len(matches) >= 2 {
			color = matches[1]
		}
		*c = Color(color)
	}
}
