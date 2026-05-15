// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"
	"strings"
)

const maxCorners = 8

type Corners struct {
	values []float32
	pct    []bool
}

func (corners *Corners) SetAll(value string, units Units) {
	values := strings.Fields(value)
	if len(values) > maxCorners {
		values = values[:maxCorners]
	}
	switch len(values) {
	case 8, 4, 2, 1:
		corners.values = make([]float32, len(values))
		corners.pct = make([]bool, len(values))
		for i := range values {
			part := strings.TrimSpace(values[i])
			if strings.HasSuffix(part, "%") {
				v, _ := strconv.ParseFloat(strings.TrimSuffix(part, "%"), 64)
				corners.values[i] = float32(v)
				corners.pct[i] = true
				continue
			}
			corners.values[i] = float32(ParseMeasurement(part, units))
		}
	}
}

func (corners *Corners) String() string {
	return fmt.Sprintf("%v", corners.Float64s())
}

func (corners Corners) Float64s() []float64 {
	return corners.Float64sFor(0, 0)
}

func (corners Corners) Float64sFor(width, height float64) []float64 {
	if len(corners.values) == 0 {
		return nil
	}
	values := make([]float64, len(corners.values))
	pctBase := min(width, height)
	for i, value := range corners.values {
		if corners.pct[i] {
			values[i] = pctBase * float64(value) / 100.0
		} else {
			values[i] = float64(value)
		}
	}
	return values
}

func (corners Corners) Len() int {
	return len(corners.values)
}
