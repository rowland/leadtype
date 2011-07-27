package pdf

import (
	"strconv"
)

func g(value float64) string {
	s := strconv.Ftoa64(value, 'f', 4)
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

type Size struct {
	Width, Height float64
}

type SizeMap map[string]Size

func stringSliceFromIntSlice(values []int) (result []string) {
	result = make([]string, len(values))
	for i, v := range values {
		result[i] = strconv.Itoa(v)
	}
	return
}