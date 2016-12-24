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

type Position int

const (
	Static = Position(iota)
	Relative
	Absolute
)

func LayoutAbsolute(container Container, style *LayoutStyle, writer Writer) {

}

func LayoutFlow(container Container, style *LayoutStyle, writer Writer) {
	var cx, cy, maxY float64
	containerFull := false
	bottom := ContentTop(container) + MaxContentHeight(container)
	widgets, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		if widget.Printed() {
			widget.SetVisible(false)
		}
	}
	for _, widget := range widgets {
		widget.SetVisible(!containerFull)
		if containerFull {
			continue
		}
		//   widget.before_layout
		if w := widget.Width(); w == 0 {
			pw := widget.PreferredWidth(writer)
			cw := ContentWidth(container)
			if pw == 0 {
				pw = cw
			}
			w = math.Min(pw, cw)
			widget.SetWidth(w)
		}
		if cx != 0 && (cx+widget.Width()) > ContentWidth(container) {
			cy += maxY + style.VPadding()
			cx, maxY = 0, 0
		}
		widget.SetLeft(ContentLeft(container) + cx)
		widget.SetTop(ContentTop(container) + cy)
		if h := widget.Height(); h == 0 {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
		widget.LayoutWidget(writer)
		if widget.Bottom() > bottom {
			containerFull = true
			//     # widget.visible = (cy == 0)
			//     widget.visible = container.root_page.positioned_widgets[:static] == 0
			//     # $stderr.puts "+++flow+++ #{container.root_page.positioned_widgets[:static]}, visible: #{widget.visible}"
			continue
		}
		//   container.root_page.positioned_widgets[widget.position] += 1
		cx += widget.Width() + style.HPadding()
		maxY = math.Max(maxY, widget.Height())
	}
	// container.more(true) if container_full and container.overflow
	if container.Height() == 0 && maxY > 0 {
		container.SetHeight(cy + maxY + NonContentHeight(container))
	}
	// super(container, writer)
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
		// TODO: skip calculations if width has already been set.
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
		top += style.VPadding()
	}
}

func printableWidgets(c Container, p Position) (widgets, remaining []Widget) {
	for _, w := range c.Widgets() {
		if w.Position() == p {
			widgets = append(widgets, w)
		} else {
			remaining = append(remaining, w)
		}
	}
	return
}

func init() {
	RegisterLayoutManager("absolute", LayoutAbsolute)
	RegisterLayoutManager("flow", LayoutFlow)
	RegisterLayoutManager("hbox", LayoutHBox)
	RegisterLayoutManager("relative", LayoutRelative)
	RegisterLayoutManager("vbox", LayoutVBox)
}
