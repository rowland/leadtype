// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	topSide = iota
	rightSide
	bottomSide
	leftSide
)

type Sides [4]float64

func (sides *Sides) SetAll(value, units string) {
	values := strings.SplitN(value, " ", len(sides))
	switch len(values) {
	case 4, 3:
		for i := 0; i < len(values); i++ {
			sides[i] = ParseMeasurement(values[i], units)
		}
	case 2:
		topBottom, rightLeft := ParseMeasurement(values[0], units), ParseMeasurement(values[1], units)
		sides[topSide] = topBottom
		sides[rightSide] = rightLeft
		sides[bottomSide] = topBottom
		sides[leftSide] = rightLeft
	case 1:
		all := ParseMeasurement(values[0], units)
		for i := 0; i < 4; i++ {
			sides[i] = all
		}
	}
}

var sideNames = [4]string{"top", "right", "bottom", "left"}

func (sides *Sides) SetAttrs(prefix string, attrs map[string]string, units string) {
	for i, name := range sideNames {
		if value, ok := attrs[prefix+name]; ok {
			sides[i] = ParseMeasurement(value, units)
		}
	}
}

func (sides *Sides) String() string {
	var buf bytes.Buffer
	buf.WriteString("[")
	for i := 0; i < len(sides); i++ {
		if i > 0 {
			buf.WriteString(" ")
		}
		fmt.Fprintf(&buf, "%s=%f", sideNames[i], sides[i])
	}
	buf.WriteString("]")
	return buf.String()
}
