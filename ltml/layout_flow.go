package ltml

import "math"

func LayoutFlow(container Container, style *LayoutStyle, writer Writer) (err error) {
	defer func() { err = wrapLayoutError("flow", containerPath(container), err) }()
	if err := validateLayoutInputs(container, style); err != nil {
		return err
	}
	var cx, cy, maxY float64
	rtl := IsRTL(container)
	containerFull := false
	continues := containerHasEffectiveContinuation(container)
	bottom := math.Inf(1)
	if container.Height() != 0 {
		bottom = ContentTop(container) + MaxContentHeight(container)
	}
	widgets, remaining := printableWidgets(container, Static)
	for _, widget := range remaining {
		widget.SetVisible(false)
	}
	contentPlaced := false
	for i, widget := range widgets {
		widget.SetVisible(!containerFull)
		if containerFull {
			continue
		}
		if isPageBreak(widget) {
			// Flow consumes pgbr at the current cursor without advancing it. A
			// break is only actionable between ordinary one-time items, which
			// prevents edge markers from creating empty pages.
			widget.ResolveWidth(0)
			widget.ResolveHeight(0)
			widget.SetLeft(ContentLeft(container) + cx)
			widget.SetTop(ContentTop(container) + cy)
			if err := widget.LayoutWidget(writer); err != nil {
				return err
			}
			if continues && contentPlaced && hasOrdinaryFlowContentAfter(widgets, i+1) {
				containerFull = true
			}
			continue
		}
		if widgetZeroFootprint(widget) {
			widget.ResolveWidth(0)
			widget.ResolveHeight(0)
			if rtl {
				widget.SetLeft(ContentRight(container))
			} else {
				widget.SetLeft(ContentLeft(container) + cx)
			}
			widget.SetTop(ContentTop(container) + cy)
			if err := widget.LayoutWidget(writer); err != nil {
				return err
			}
			widget.SetVisible(!continues || widget.Top() <= bottom)
			continue
		}
		if widgetAutoWidth(widget) || !widgetWidthSpecified(widget) {
			pw, err := widget.PreferredWidth(writer)
			if err != nil {
				return err
			}
			cw := ContentWidth(container)
			if pw == 0 {
				pw = cw
			}
			w := min(pw, cw)
			widget.ResolveWidth(w)
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
		if widgetAutoHeight(widget) || !widgetHeightSpecified(widget) {
			height, err := widget.PreferredHeight(writer)
			if err != nil {
				return err
			}
			widget.ResolveHeight(height)
		}
		if err := widget.LayoutWidget(writer); err != nil {
			return err
		}
		// As with vbox, a nested continuing child may need splitting because of
		// its own pgbr even though its outer dimensions fit this flow fragment.
		if continues && (widgetHasActionablePageBreak(widget) || widget.Bottom() > bottom) {
			containerFull = true
			widget.SetVisible(false)
			continue
		}
		if widget.Display() == DisplayOnce {
			contentPlaced = true
		}
		cx += widget.Width() + style.HPadding()
		maxY = max(maxY, widget.Height())
	}
	if container.Height() == 0 && maxY > 0 {
		container.ResolveHeight(cy + maxY + NonContentHeight(container))
	}
	return layoutPositionedChildren(container, writer)
}

func hasOrdinaryFlowContentAfter(widgets []Widget, start int) bool {
	for _, widget := range widgets[start:] {
		if isPageBreak(widget) || widgetZeroFootprint(widget) {
			continue
		}
		if widget.Display() == DisplayOnce {
			return true
		}
	}
	return false
}
