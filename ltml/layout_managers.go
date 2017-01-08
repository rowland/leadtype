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
	fmt.Println("In LayoutHBox")
	containerFull := false

	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		if widget.Printed() {
			widget.SetVisible(false)
		}
	}

	var lpanels, rpanels, unaligned []Widget
	for _, widget := range static {
		switch widget.Align() {
		case AlignLeft:
			lpanels = append(lpanels, widget)
		case AlignRight:
			rpanels = append(rpanels, widget)
		default:
			unaligned = append(unaligned, widget)
		}
	}

	var percents, specified, others []Widget
	for _, widget := range static {
		if widget.WidthPctIsSet() {
			percents = append(percents, widget)
		} else if widget.WidthIsSet() {
			specified = append(specified, widget)
		} else {
			others = append(others, widget)
		}
	}

	widthAvail := ContentWidth(container)

	// allocate specified widths first
	for _, widget := range specified {
		widthAvail -= widget.Width()
		containerFull = widthAvail < 0
		widget.SetDisabled(containerFull)
		widthAvail -= style.HPadding()
	}

	// allocate percent widths next, with a minimum width of 1 point
	if widthAvail-float64(len(percents)-1)*style.HPadding() >= float64(len(percents)) {
		widthAvail -= float64(len(percents)-1) * style.HPadding()
		totalPercents := 0.0
		for _, widget := range percents {
			totalPercents += widget.Width()
		}
		ratio := widthAvail / totalPercents
		for _, widget := range percents {
			if ratio < 1.0 {
				widget.SetWidth(widget.Width() * ratio)
			}
			widthAvail -= widget.Width()
		}
	} else {
		containerFull = true
		for _, widget := range percents {
			widget.SetDisabled(true)
		}
	}
	widthAvail -= style.HPadding()

	// divide remaining width equally among widgets with unspecified widths
	if widthAvail-float64(len(others)-1)*style.HPadding() >= float64(len(others)) {
		widthAvail -= float64(len(others)-1) * style.HPadding()
		othersWidth := widthAvail / float64(len(others))
		for _, widget := range others {
			widget.SetWidth(othersWidth)
		}
	} else {
		containerFull = true
		for _, widget := range others {
			widget.SetDisabled(true)
		}
	}

	for _, widget := range static {
		if container.Align() == AlignBottom {
			widget.SetBottom(ContentBottom(container))
		} else {
			containerFull = true
			widget.SetTop(ContentTop(container))
		}
		if !widget.HeightIsSet() {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
	}

	left := ContentLeft(container)
	right := ContentRight(container)
	fmt.Println("left:", left, "right:", right)

	for _, widget := range lpanels {
		if widget.Disabled() {
			continue
		}
		widget.SetLeft(left)
		left += (widget.Width() + style.HPadding())
	}
	for i := len(rpanels) - 1; i >= 0; i-- {
		widget := rpanels[i]
		if widget.Disabled() {
			continue
		}
		widget.SetRight(right)
		right -= (widget.Width() + style.HPadding())
	}
	for _, widget := range unaligned {
		if widget.Disabled() {
			continue
		}
		widget.SetLeft(left)
		left += (widget.Width() + style.HPadding())
	}

	if !container.HeightIsSet() {
		contentHeight := 0.0
		for _, widget := range static {
			if widget.Height() > contentHeight {
				contentHeight = widget.Height()
			}
		}
		container.SetHeight(contentHeight + NonContentHeight(container))
	}
	for _, widget := range static {
		if widget.Visible() && !widget.Disabled() {
			widget.LayoutWidget(writer)
		}
	}
	// super(container, writer)
}

func LayoutRelative(container Container, style *LayoutStyle, writer Writer) {

}

func LayoutVBox(container Container, style *LayoutStyle, writer Writer) {
	fmt.Println("In LayoutVBox")
	containerFull := false
	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		if widget.Printed() {
			widget.SetVisible(false)
		}
	}
	var headers, footers, unaligned []Widget
	for _, widget := range static {
		switch widget.Align() {
		case AlignTop:
			headers = append(headers, widget)
		case AlignBottom:
			footers = append(footers, widget)
		default:
			unaligned = append(unaligned, widget)
		}
	}
	left := ContentLeft(container)
	fmt.Println("left:", left)
	for _, widget := range static {
		if !widget.WidthIsSet() {
			pw := widget.PreferredWidth(writer)
			// fmt.Println("pw:", pw, widget)
			cw := ContentWidth(container)
			// fmt.Println("cw:", cw, container)
			if pw == 0 {
				pw = cw
			}
			w := math.Min(pw, cw)
			// fmt.Println("w:", w)
			// panic("foo")
			widget.SetWidth(w)
		}
		widget.SetLeft(left)
	}
	top, dy := ContentTop(container), 0.0
	fmt.Println("top:", top)
	bottom := ContentTop(container) + MaxContentHeight(container)

	for i, widget := range headers {
		widget.SetTop(top)
		widget.LayoutWidget(writer)
		if !widget.HeightIsSet() {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
		top += widget.Height() + style.VPadding()
		dy += widget.Height()
		if i > 0 {
			dy += style.VPadding()
		}
		widget.SetVisible(widget.Bottom() <= bottom)
	}

	if len(footers) > 0 {
		if !container.HeightIsSet() {
			container.SetHeightPct(100)
		}
		for i := len(footers) - 1; i >= 0; i-- {
			widget := footers[i]
			widget.SetBottom(bottom)
			widget.LayoutWidget(writer)
			if !widget.HeightIsSet() {
				widget.SetHeight(widget.PreferredHeight(writer))
			}
			widget.SetVisible(widget.Top() >= top)
		}
	}

	widgetsVisible := 0
	for i, widget := range unaligned {
		widget.SetVisible(!containerFull)
		if containerFull {
			continue
		}
		widget.SetTop(top)
		widget.LayoutWidget(writer)
		if !widget.HeightIsSet() {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
		top += widget.Height()
		dy += widget.Height()
		if i > 0 {
			dy += style.VPadding()
		}
		if top > bottom {
			containerFull = true
			widget.SetVisible(widgetsVisible == 0)
			// widget.visible = widget.leaves > 0 and container.root_page.positioned_widgets[:static] == 0
		}
		if widget.Visible() {
			widgetsVisible += 1
		}
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
