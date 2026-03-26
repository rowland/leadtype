// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

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

func rotateXY(x, y, angle float64) (float64, float64) {
	theta := angle * math.Pi / 180.0
	vCos := math.Cos(theta)
	vSin := math.Sin(theta)
	return (x * vCos) - (y * vSin), (x * vSin) + (y * vCos)
}

func rotatePoint(center, point Location, angle float64) Location {
	x, y := rotateXY(point.X-center.X, center.Y-point.Y, angle)
	return Location{center.X + x, center.Y - y}
}

func calcArcSmall(r, midTheta, halfAngle, ccwcw float64) (points [4]Location) {
	halfTheta := math.Abs(halfAngle) * math.Pi / 180.0
	vCos := math.Cos(halfTheta)
	vSin := math.Sin(halfTheta)

	x0 := r * vCos
	y0 := -ccwcw * r * vSin
	x1 := r * (4.0 - vCos) / 3.0
	x2 := x1
	y1 := r * ccwcw * (1.0 - vCos) * (vCos - 3.0) / (3.0 * vSin)
	y2 := -y1
	x3 := r * vCos
	y3 := ccwcw * r * vSin

	x0, y0 = rotateXY(x0, y0, midTheta)
	x1, y1 = rotateXY(x1, y1, midTheta)
	x2, y2 = rotateXY(x2, y2, midTheta)
	x3, y3 = rotateXY(x3, y3, midTheta)

	points[0] = Location{x0, y0}
	points[1] = Location{x1, y1}
	points[2] = Location{x2, y2}
	points[3] = Location{x3, y3}
	return
}

func reverseCurvePoints(points []Location) []Location {
	if len(points) == 0 {
		return nil
	}
	if len(points) < 4 || (len(points)-1)%3 != 0 {
		reversed := append([]Location(nil), points...)
		LocationSlice(reversed).Reverse()
		return reversed
	}
	reversed := make([]Location, 0, len(points))
	reversed = append(reversed, points[len(points)-1])
	for i := len(points) - 3; i >= 1; i -= 3 {
		reversed = append(reversed, points[i+1], points[i], points[i-1])
	}
	return reversed
}
