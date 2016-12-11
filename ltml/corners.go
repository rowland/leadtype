// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

const maxCorners = 8

type Corners []float64

func (corners *Corners) SetAll(value string, units Units) {
	values := strings.SplitN(value, " ", maxCorners)
	switch len(values) {
	case 8, 4, 2, 1:
		*corners = make([]float64, len(values))
		for i := 0; i < len(values); i++ {
			(*corners)[i] = ParseMeasurement(values[i], units)
		}
	}
}

func (corners *Corners) String() string {
	return fmt.Sprintf("%v", []float64(*corners))
}
