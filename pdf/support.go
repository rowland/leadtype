// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
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

type location struct {
	x, y float64
}

func (this location) equal(other location) bool {
	return this.x == other.x && this.y == other.y
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

func stringSlicesEqual(sl1, sl2 []string) bool {
	if len(sl1) != len(sl2) {
		return false
	}
	for i, s := range sl1 {
		if s != sl2[i] {
			return false
		}
	}
	return true
}
