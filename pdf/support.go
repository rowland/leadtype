package pdf

type Options map[string]interface{}

func (this Options) StringDefault(key, def string) string {
	if value, ok := this[key]; ok {
		return value.(string)
	}
	return def
}

type Size struct {
	Width, Height float32
}

type SizeMap map[string]Size
