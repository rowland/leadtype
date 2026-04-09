package ltml

func LayoutFlow(container Container, style *LayoutStyle, writer Writer) {
	var cx, cy, maxY float64
	rtl := IsRTL(container)
	containerFull := false
	bottom := ContentTop(container) + MaxContentHeight(container)
	widgets, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		widget.SetVisible(false)
	}
	for _, widget := range widgets {
		widget.SetVisible(!containerFull)
		if containerFull {
			continue
		}
		if widgetZeroFootprint(widget) {
			widget.SetWidth(0)
			widget.SetHeight(0)
			if rtl {
				widget.SetLeft(ContentRight(container))
			} else {
				widget.SetLeft(ContentLeft(container) + cx)
			}
			widget.SetTop(ContentTop(container) + cy)
			widget.LayoutWidget(writer)
			widget.SetVisible(widget.Top() <= bottom)
			continue
		}
		if w := widget.Width(); w == 0 {
			pw := widget.PreferredWidth(writer)
			cw := ContentWidth(container)
			if pw == 0 {
				pw = cw
			}
			w = min(pw, cw)
			widget.SetWidth(w)
		}
		if cx != 0 && (cx+widget.Width()) > ContentWidth(container) {
			cy += maxY + style.VPadding()
			cx, maxY = 0, 0
		}
		if rtl {
			widget.SetLeft(ContentRight(container) - cx - widget.Width())
		} else {
			widget.SetLeft(ContentLeft(container) + cx)
		}
		widget.SetTop(ContentTop(container) + cy)
		if h := widget.Height(); h == 0 {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
		widget.LayoutWidget(writer)
		if widget.Bottom() > bottom {
			containerFull = true
			widget.SetVisible(false)
			continue
		}
		cx += widget.Width() + style.HPadding()
		maxY = max(maxY, widget.Height())
	}
	if container.Height() == 0 && maxY > 0 {
		container.SetHeight(cy + maxY + NonContentHeight(container))
	}
	layoutPositionedChildren(container, writer)
}
