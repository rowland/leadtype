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
	left := ContentLeft(container)
	fmt.Println("left:", left)
	for _, widget := range container.Widgets() {
		pw := widget.PreferredWidth(writer)
		fmt.Println("pw:", pw, widget)
		cw := ContentWidth(container)
		fmt.Println("cw:", cw, container)
		if pw == 0 {
			pw = cw
		}
		w := math.Min(pw, cw)
		fmt.Println("w:", w)
		// panic("foo")
		widget.SetWidth(w)
		widget.SetLeft(left)
	}
	top := ContentTop(container)
	fmt.Println("top:", top)
	for _, widget := range container.Widgets() {
		widget.SetTop(top)
		widget.LayoutWidget(writer)
		if widget.Height() == 0 {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
		top += widget.Height()
	}
}

func init() {
	RegisterLayoutManager("absolute", LayoutAbsolute)
	RegisterLayoutManager("flow", LayoutFlow)
	RegisterLayoutManager("hbox", LayoutHBox)
	RegisterLayoutManager("relative", LayoutRelative)
	RegisterLayoutManager("vbox", LayoutVBox)
}
