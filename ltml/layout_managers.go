package ltml

import (
	"fmt"
	"math"
)

type LayoutFunc func(container Container, style *LayoutStyle, writer Writer)

var layoutManagers = make(map[string]LayoutFunc)

func RegisterLayoutManager(name string, f LayoutFunc) {
	layoutManagers[name] = f
}

func LayoutManagerFor(name string) LayoutFunc {
	fmt.Println("In LayoutManagerFor")
	if f, ok := layoutManagers[name]; ok {
		// fmt.Printf("%#v", f)
		return f
	}
	fmt.Println("couldn't find ", name)
	return LayoutVBox
}

func LayoutAbsolute(container Container, style *LayoutStyle, writer Writer) {

}

func LayoutFlow(container Container, style *LayoutStyle, writer Writer) {

}

func LayoutHBox(container Container, style *LayoutStyle, writer Writer) {

}

func LayoutRelative(container Container, style *LayoutStyle, writer Writer) {

}

func LayoutVBox(container Container, style *LayoutStyle, writer Writer) {
	fmt.Println("In LayoutVBox")
	for _, widget := range container.Widgets() {
		pw := widget.PreferredWidth()
		fmt.Println("pw:", pw)
		cw := container.ContentWidth()
		fmt.Println("cw:", cw)
		if pw == 0 {
			pw = cw
		}
		w := math.Min(pw, cw)
		fmt.Println("w:", w)
		widget.SetWidth(w)
		widget.SetLeft(container.ContentLeft())
	}
}

func init() {
	RegisterLayoutManager("absolute", LayoutAbsolute)
	RegisterLayoutManager("flow", LayoutFlow)
	RegisterLayoutManager("hbox", LayoutHBox)
	RegisterLayoutManager("relative", LayoutRelative)
	RegisterLayoutManager("vbox", LayoutVBox)
}
