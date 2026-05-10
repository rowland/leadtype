package ltml

import "math"

func imageLikeLayoutSize(widget *StdWidget, naturalWidth, naturalHeight float64) (width, height float64) {
	if naturalWidth <= 0 || naturalHeight <= 0 {
		return widget.Width(), widget.Height()
	}
	if width, height, ok := imageLikeLegacyLayoutSize(widget, naturalWidth, naturalHeight); ok {
		return width, height
	}

	contentWidth, contentHeight := imageLikeContentSize(widget, naturalWidth, naturalHeight)
	return contentWidth + NonContentWidth(widget), contentHeight + NonContentHeight(widget)
}

func imageLikeLegacyLayoutSize(widget *StdWidget, naturalWidth, naturalHeight float64) (width, height float64, ok bool) {
	authoredWidth := widgetWidthSpecified(widget)
	authoredHeight := widgetHeightSpecified(widget)
	if widget.MaxWidthIsSet() || widget.MaxHeightIsSet() ||
		((widget.widthValid || widget.heightValid) && !authoredWidth && !authoredHeight) {
		return 0, 0, false
	}
	switch {
	case authoredWidth && authoredHeight:
		return widget.Width(), widget.Height(), true
	case authoredWidth:
		return widget.Width(), widget.Width()*naturalHeight/naturalWidth + NonContentHeight(widget), true
	case authoredHeight:
		return widget.Height()*naturalWidth/naturalHeight + NonContentWidth(widget), widget.Height(), true
	default:
		return naturalWidth + NonContentWidth(widget), naturalHeight + NonContentHeight(widget), true
	}
}

func imageLikeContentSize(widget *StdWidget, naturalWidth, naturalHeight float64) (width, height float64) {
	nonContentWidth := NonContentWidth(widget)
	nonContentHeight := NonContentHeight(widget)
	authoredWidth := widgetWidthSpecified(widget)
	authoredHeight := widgetHeightSpecified(widget)

	switch {
	case authoredWidth && authoredHeight:
		width = max(widget.uncappedWidth()-nonContentWidth, 0)
		height = max(widget.uncappedHeight()-nonContentHeight, 0)
	case authoredWidth:
		width = max(widget.uncappedWidth()-nonContentWidth, 0)
		height = width * naturalHeight / naturalWidth
	case authoredHeight:
		height = max(widget.uncappedHeight()-nonContentHeight, 0)
		width = height * naturalWidth / naturalHeight
	default:
		width = naturalWidth
		height = naturalHeight
	}

	maxWidth, maxHeight := imageLikeContentBounds(widget, nonContentWidth, nonContentHeight, authoredWidth, authoredHeight)
	scale := 1.0
	if maxWidth >= 0 && width > maxWidth && width > 0 {
		scale = min(scale, maxWidth/width)
	}
	if maxHeight >= 0 && height > maxHeight && height > 0 {
		scale = min(scale, maxHeight/height)
	}
	return width * scale, height * scale
}

func imageLikePlacementSize(widget *StdWidget, naturalWidth, naturalHeight float64) (width, height *float64) {
	if naturalWidth <= 0 || naturalHeight <= 0 {
		return imageLikeFallbackPlacementSize(widget)
	}

	authoredWidth := widgetWidthSpecified(widget)
	authoredHeight := widgetHeightSpecified(widget)
	constrained := widget.MaxWidthIsSet() || widget.MaxHeightIsSet() ||
		((widget.widthValid || widget.heightValid) && !authoredWidth && !authoredHeight)

	if constrained || (authoredWidth && authoredHeight) {
		contentWidth, contentHeight := imageLikeContentSize(widget, naturalWidth, naturalHeight)
		return &contentWidth, &contentHeight
	}
	if authoredWidth {
		contentWidth := max(widget.Width()-NonContentWidth(widget), 0)
		return &contentWidth, nil
	}
	if authoredHeight {
		contentHeight := max(widget.Height()-NonContentHeight(widget), 0)
		return nil, &contentHeight
	}
	return nil, nil
}

func imageLikeFallbackPlacementSize(widget *StdWidget) (width, height *float64) {
	authoredWidth := widgetWidthSpecified(widget)
	authoredHeight := widgetHeightSpecified(widget)
	if authoredWidth || (widget.widthValid && !authoredWidth) {
		contentWidth := max(ContentWidth(widget), 0)
		width = &contentWidth
	}
	if authoredHeight || (widget.heightValid && !authoredHeight) {
		contentHeight := max(ContentHeight(widget), 0)
		height = &contentHeight
	}
	return width, height
}

func imageLikeContentBounds(widget *StdWidget, nonContentWidth, nonContentHeight float64, authoredWidth, authoredHeight bool) (maxWidth, maxHeight float64) {
	maxWidth = math.Inf(1)
	maxHeight = math.Inf(1)
	if widget.MaxWidthIsSet() {
		maxWidth = max(widget.MaxWidth()-nonContentWidth, 0)
	}
	if widget.MaxHeightIsSet() {
		maxHeight = max(widget.MaxHeight()-nonContentHeight, 0)
	}
	if widget.widthValid && !authoredWidth && !authoredHeight {
		maxWidth = min(maxWidth, max(widget.Width()-nonContentWidth, 0))
	}
	if widget.heightValid && !authoredWidth && !authoredHeight {
		maxHeight = min(maxHeight, max(widget.Height()-nonContentHeight, 0))
	}
	if math.IsInf(maxWidth, 1) {
		maxWidth = -1
	}
	if math.IsInf(maxHeight, 1) {
		maxHeight = -1
	}
	return maxWidth, maxHeight
}
