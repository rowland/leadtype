package ltml

import "fmt"

type LayoutFunc func(container Container, style *LayoutStyle, writer Writer) error

var layoutManagers = make(map[string]LayoutFunc)

func RegisterLayoutManager(name string, f LayoutFunc) {
	layoutManagers[name] = f
}

func LayoutManagerFor(name string) (LayoutFunc, error) {
	if f, ok := layoutManagers[name]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("unknown layout manager %q", name)
}

func init() {
	RegisterLayoutManager("absolute", LayoutAbsolute)
	RegisterLayoutManager("flow", LayoutFlow)
	RegisterLayoutManager("hbox", LayoutHBox)
	RegisterLayoutManager("relative", LayoutRelative)
	RegisterLayoutManager("table", LayoutTable)
	RegisterLayoutManager("vbox", LayoutVBox)
}
