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

type Options map[string]interface{}

func (this Options) ColorDefault(key string, def Color) (color Color) {
	color = def
	if value, ok := this[key]; ok {
		switch value := value.(type) {
		case string:
			if c, err := NamedColor(value); err == nil {
				color = c
			}
		case int:
			color = Color(value)
		case int32:
			color = Color(value)
		}
	}
	return
}

func (this Options) StringDefault(key, def string) string {
	if value, ok := this[key]; ok {
		return value.(string)
	}
	return def
}

func (this Options) Merge(other Options) Options {
	result := make(Options, len(this)+len(other))
	for k, v := range this {
		result[k] = v
	}
	for k, v := range other {
		result[k] = v
	}
	return result
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
