// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"regexp"
	"strings"

	"github.com/rowland/leadtype/colors"
)

var (
	reHexColor = regexp.MustCompile(`^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`)
)

func NamedColor(color string) colors.Color {
	c, _ := parseLTMLColor(color)
	return c
}

func parseLTMLColor(value string) (colors.Color, bool) {
	value = strings.TrimSpace(value)
	if matches := reHexColor.FindStringSubmatch(value); len(matches) >= 2 {
		hex := matches[1]
		if len(hex) == 3 {
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		value = hex
	}
	c, err := colors.NamedColor(value)
	return c, err == nil
}
