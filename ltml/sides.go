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

type Side struct {
	Value float64
	IsSet bool
}

func (side *Side) Set(value float64) {
	side.Value = value
	side.IsSet = true
}

type Sides [4]Side

func (sides *Sides) SetAll(value string, units Units) {
	values := strings.SplitN(value, " ", len(sides))
	switch len(values) {
	case 4, 3:
		for i := 0; i < len(values); i++ {
			sides[i].Set(ParseMeasurement(values[i], units))
		}
	case 2:
		topBottom, rightLeft := ParseMeasurement(values[0], units), ParseMeasurement(values[1], units)
		sides[topSide].Set(topBottom)
		sides[rightSide].Set(rightLeft)
		sides[bottomSide].Set(topBottom)
		sides[leftSide].Set(rightLeft)
	case 1:
		all := ParseMeasurement(values[0], units)
		for i := 0; i < 4; i++ {
			sides[i].Set(all)
		}
	}
}

var sideNames = [4]string{"top", "right", "bottom", "left"}

func (sides *Sides) SetAttrs(prefix string, attrs map[string]string, units Units) {
	for i, name := range sideNames {
		if value, ok := attrs[prefix+name]; ok {
			sides[i].Set(ParseMeasurement(value, units))
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
		fmt.Fprintf(&buf, "%s=%f/%v", sideNames[i], sides[i].Value, sides[i].IsSet)
	}
	buf.WriteString("]")
	return buf.String()
}
