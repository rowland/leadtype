// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

const maxCorners = 8

type Corners []float32

func (corners *Corners) SetAll(value string, units Units) {
	values := strings.SplitN(value, " ", maxCorners)
	switch len(values) {
	case 8, 4, 2, 1:
		*corners = make([]float32, len(values))
		for i := range values {
			(*corners)[i] = float32(ParseMeasurement(values[i], units))
		}
	}
}

func (corners *Corners) String() string {
	return fmt.Sprintf("%v", corners.Float64s())
}

func (corners Corners) Float64s() []float64 {
	if len(corners) == 0 {
		return nil
	}
	values := make([]float64, len(corners))
	for i, value := range corners {
		values[i] = float64(value)
	}
	return values
}
