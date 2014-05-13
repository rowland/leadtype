// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"strconv"
)

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
