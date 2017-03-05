// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package haru

import (
	"bytes"
	"math"
	"strconv"
)

func g(value float64) string {
	s := strconv.FormatFloat(value, 'f', 4, 64)
	n := len(s)
	for n > 0 && s[n-1] == '0' {
		n--
	}
	if n > 0 && s[n-1] == '.' {
		n--
	}
	return s[:n]
}

type Location struct {
	X, Y float64
}

func (this Location) equal(other Location) bool {
	return this.X == other.X && this.Y == other.Y
}

type LocationSlice []Location

func (slice LocationSlice) Reverse() {
	for left, right := 0, len(slice)-1; left < right; left, right = left+1, right-1 {
		slice[left], slice[right] = slice[right], slice[left]
	}
}

type LocationSliceSlice [][]Location

func (slice LocationSliceSlice) Reverse() {
	for left, right := 0, len(slice)-1; left < right; left, right = left+1, right-1 {
		slice[left], slice[right] = slice[right], slice[left]
	}
}

type Size struct {
	Width, Height float64
}

type SizeMap map[string]Size

type intSlice []int

func (s intSlice) join(separator string) string {
	var buf bytes.Buffer
	for i, v := range s {
		if i > 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(strconv.Itoa(v))
	}
	return buf.String()
}

type float64Slice []float64

func (s float64Slice) join(separator string) string {
	var buf bytes.Buffer
	for i, v := range s {
		if i > 0 {
			buf.WriteString(separator)
		}
		buf.WriteString(g(v))
	}
	return buf.String()
}

var signs = []Location{{1, -1}, {-1, -1}, {-1, 1}, {1, 1}}

func quadrantBezierPoints(quadrant int, x, y, rx, ry float64) (bp []Location) {
	const a = 4.0 / 3.0 * (math.Sqrt2 - 1.0)
	if quadrant == 1 || quadrant == 3 {
		// (1, 0)
		bp = append(bp, Location{x + (rx * signs[quadrant-1].X), y})
		// (1, a)
		bp = append(bp, Location{bp[0].X, y + (a * ry * signs[quadrant-1].Y)})
		// (a, 1)
		bp = append(bp, Location{x + (a * rx * signs[quadrant-1].X), y + (ry * signs[quadrant-1].Y)})
		// (0, 1)
		bp = append(bp, Location{x, bp[2].Y})
	} else {
		// (0,1)
		bp = append(bp, Location{x, y + (ry * signs[quadrant-1].Y)})
		// (a,1)
		bp = append(bp, Location{x + (a * rx * signs[quadrant-1].Y), bp[0].Y})
		// (1,a)
		bp = append(bp, Location{x + (rx * signs[quadrant-1].X), y + (a * ry * signs[quadrant-1].Y)})
		// (1,0)
		bp = append(bp, Location{bp[2].X, y})
	}
	return
}
