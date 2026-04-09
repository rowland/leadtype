package ltml

func LayoutHBox(container Container, style *LayoutStyle, writer Writer) {
	containerFull := false

	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		widget.SetVisible(false)
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

	for _, widget := range specified {
		widthAvail -= widget.Width()
		containerFull = widthAvail < 0
		widget.SetDisabled(containerFull)
		widthAvail -= style.HPadding()
	}

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
		if !widget.HeightIsSet() {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
		widget.SetTop(hboxCrossAxisTop(container, widget))
	}

	left := ContentLeft(container)
	right := ContentRight(container)

	if IsRTL(container) {
		for _, widget := range lpanels {
			if widget.Disabled() {
				continue
			}
			widget.SetLeft(right - widget.Width())
			right -= (widget.Width() + style.HPadding())
		}
		for i := len(rpanels) - 1; i >= 0; i-- {
			widget := rpanels[i]
			if widget.Disabled() {
				continue
			}
			widget.SetLeft(left)
			left += (widget.Width() + style.HPadding())
		}
		for _, widget := range unaligned {
			if widget.Disabled() {
				continue
			}
			widget.SetLeft(right - widget.Width())
			right -= (widget.Width() + style.HPadding())
		}
	} else {
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
	layoutPositionedChildren(container, writer)
}

func LayoutVBox(container Container, style *LayoutStyle, writer Writer) {
	containerFull := false
	static, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		widget.SetVisible(false)
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
	rtl := IsRTL(container)
	for _, widget := range static {
		if !widget.WidthIsSet() {
			cw := ContentWidth(container)
			pw := 0.0
			if _, ok := widget.(*StdParagraph); ok {
				pw = cw
			} else {
				pw = widget.PreferredWidth(writer)
			}
			if pw == 0 {
				pw = cw
			}
			w := min(pw, cw)
			widget.SetWidth(w)
		}
		widget.SetLeft(vboxCrossAxisLeft(container, widget, rtl))
	}
	top, dy := ContentTop(container), 0.0
	bottom := ContentTop(container) + MaxContentHeight(container)

	for i, widget := range headers {
		widget.SetTop(top)
		widget.LayoutWidget(writer)
		if !widget.HeightIsSet() {
			widget.SetHeight(widget.PreferredHeight(writer))
		}
		if widgetZeroFootprint(widget) {
			widget.SetVisible(widget.Top() <= bottom)
			continue
		}
		top += widget.Height() + style.VPadding()
		dy += widget.Height()
		if i > 0 {
			dy += style.VPadding()
		}
		widget.SetVisible(widget.Bottom() <= bottom)
	}

	if len(footers) > 0 {
		footerBottom := bottom
		for i := len(footers) - 1; i >= 0; i-- {
			widget := footers[i]
			if !widget.HeightIsSet() {
				widget.SetHeight(widget.PreferredHeight(writer))
			}
			widget.SetBottom(footerBottom)
			widget.LayoutWidget(writer)
			if widgetZeroFootprint(widget) {
				widget.SetVisible(widget.Top() >= top)
				continue
			}
			footerBottom = widget.Top() - style.VPadding()
			widget.SetVisible(widget.Top() >= top)
		}
	}

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
		if widgetZeroFootprint(widget) {
			widget.SetVisible(widget.Top() <= bottom)
			continue
		}
		top += widget.Height()
		dy += widget.Height()
		if i > 0 {
			dy += style.VPadding()
		}
		if top > bottom {
			containerFull = true
			widget.SetVisible(false)
		}
		top += style.VPadding()
	}
	if !container.HeightIsSet() {
		contentHeight := dy
		if len(headers) > 0 {
			contentHeight = max(contentHeight, top-ContentTop(container)-style.VPadding())
		}
		if len(footers) > 0 {
			footerTop := bottom
			for _, widget := range footers {
				if widget.Visible() {
					footerTop = min(footerTop, widget.Top())
				}
			}
			if footerTop < bottom {
				contentHeight = max(contentHeight, footerTop-ContentTop(container))
			}
		}
		container.SetHeight(max(contentHeight, 0) + NonContentHeight(container))
	}
	layoutPositionedChildren(container, writer)
}

func hboxCrossAxisTop(container Container, widget Widget) float64 {
	switch widget.SelfAlign() {
	case SelfAlignEnd:
		return ContentBottom(container) - widget.Height()
	case SelfAlignCenter:
		return ContentTop(container) + max(ContentHeight(container)-widget.Height(), 0)/2
	default:
		if container.Align() == AlignBottom {
			return ContentBottom(container) - widget.Height()
		}
		return ContentTop(container)
	}
}

func vboxCrossAxisLeft(container Container, widget Widget, rtl bool) float64 {
	switch widget.SelfAlign() {
	case SelfAlignCenter:
		return ContentLeft(container) + max(ContentWidth(container)-widget.Width(), 0)/2
	case SelfAlignEnd:
		if rtl {
			return ContentLeft(container)
		}
		return ContentRight(container) - widget.Width()
	default:
		if rtl {
			return ContentRight(container) - widget.Width()
		}
		return ContentLeft(container)
	}
}
