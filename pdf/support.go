package pdf

type Options map[string]interface{}

func (this Options) StringDefault(key, def string) string {
	if value, ok := this[key]; ok {
		return value.(string)
	}
	return def
}

func (this Options) Merge(other Options) Options {
	result := make(Options, len(this) + len(other))
	for k, v := range this {
		result[k] = v
	}
	for k, v := range other {
		result[k] = v
	}
	return result
}

type Size struct {
	Width, Height float32
}

type SizeMap map[string]Size
