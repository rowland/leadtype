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

func (this Options) BoolDefault(key string, def bool) bool {
	if value, ok := this[key]; ok {
		switch value := value.(type) {
		case bool:
			return value
		case string:
			if b, err := strconv.ParseBool(value); err == nil {
				return b
			}
			return def
		case int:
			return value != 0
		case float64:
			return value != 0
		}
	}
	return def
}

func (this Options) ColorDefault(key string, def Color) (color Color) {
	color = def
	if value, ok := this[key]; ok {
		switch value := value.(type) {
		case Color:
			color = value
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
		switch value := value.(type) {
		case string:
			return value
		case int:
			return strconv.Itoa(value)
		case float64:
			return strconv.FormatFloat(value, 'f', -1, 64)
		}
	}
	return def
}

func (this Options) FloatDefault(key string, def float64) float64 {
	if value, ok := this[key]; ok {
		switch value := value.(type) {
		case string:
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				return f
			}
			return def
		case int:
			return float64(value)
		case float64:
			return value
		}
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
