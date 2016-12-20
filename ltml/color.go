// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"regexp"

	"github.com/rowland/leadtype/colors"
)

var (
	re6DigitHexColor = regexp.MustCompile(`#([0-9A-Fa-f]{6})`)
)

func NamedColor(color string) colors.Color {
	if matches := re6DigitHexColor.FindStringSubmatch(color); len(matches) >= 2 {
		color = matches[1]
	}
	c, _ := colors.NamedColor(color)
	return c
}
