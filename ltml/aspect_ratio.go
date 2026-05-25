// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import "math"

type IntrinsicAspectRatioProvider interface {
	IntrinsicAspectRatio(writer Writer) (widthOverHeight float64, ok bool)
}

type aspectWidthResolver interface {
	ResolveAspectWidth(value float64)
}

type aspectHeightResolver interface {
	ResolveAspectHeight(value float64)
}

type aspectDimensionState interface {
	WidthAspectInferred() bool
	HeightAspectInferred() bool
}

func prepareAspectRatioDimensions(container Container, writer Writer) {
	for _, widget := range container.Widgets() {
		resolveAspectRatioDimensions(widget, writer)
	}
}

func resolveAspectRatioDimensions(widget Widget, writer Writer) {
	provider, ok := widget.(IntrinsicAspectRatioProvider)
	if !ok {
		return
	}
	aspectRatio, ok := provider.IntrinsicAspectRatio(writer)
	if !ok || aspectRatio <= 0 || math.IsNaN(aspectRatio) || math.IsInf(aspectRatio, 0) {
		return
	}

	widthSpecified := widgetWidthAuthored(widget)
	heightSpecified := widgetHeightAuthored(widget)
	if widthSpecified == heightSpecified {
		return
	}
	width, height := aspectRatioOuterSize(widget, aspectRatio)
	if widthSpecified && !widgetAutoHeight(widget) {
		if resolver, ok := widget.(interface{ ResolveWidth(float64) }); ok {
			resolver.ResolveWidth(width)
		}
		if resolver, ok := widget.(aspectHeightResolver); ok {
			resolver.ResolveAspectHeight(height)
		} else {
			widget.ResolveHeight(height)
		}
	} else if heightSpecified && !widgetAutoWidth(widget) {
		if resolver, ok := widget.(aspectWidthResolver); ok {
			resolver.ResolveAspectWidth(width)
		} else {
			widget.ResolveWidth(width)
		}
		if resolver, ok := widget.(interface{ ResolveHeight(float64) }); ok {
			resolver.ResolveHeight(height)
		}
	}
}

func aspectRatioOuterSize(widget Widget, aspectRatio float64) (width, height float64) {
	nonContentWidth := NonContentWidth(widget)
	nonContentHeight := NonContentHeight(widget)
	widthSpecified := widgetWidthAuthored(widget)
	heightSpecified := widgetHeightAuthored(widget)

	contentWidth := 0.0
	contentHeight := 0.0
	if widthSpecified {
		contentWidth = max(widget.Width()-nonContentWidth, 0)
		contentHeight = contentWidth / aspectRatio
	} else if heightSpecified {
		contentHeight = max(widget.Height()-nonContentHeight, 0)
		contentWidth = contentHeight * aspectRatio
	}

	maxWidth, maxHeight := aspectRatioContentBounds(widget, nonContentWidth, nonContentHeight)
	scale := 1.0
	if maxWidth >= 0 && contentWidth > maxWidth && contentWidth > 0 {
		scale = min(scale, maxWidth/contentWidth)
	}
	if maxHeight >= 0 && contentHeight > maxHeight && contentHeight > 0 {
		scale = min(scale, maxHeight/contentHeight)
	}
	return contentWidth*scale + nonContentWidth, contentHeight*scale + nonContentHeight
}

func aspectRatioContentBounds(widget Widget, nonContentWidth, nonContentHeight float64) (maxWidth, maxHeight float64) {
	maxWidth = math.Inf(1)
	maxHeight = math.Inf(1)
	if widget.MaxWidthIsSet() {
		maxWidth = max(widget.MaxWidth()-nonContentWidth, 0)
	}
	if widget.MaxHeightIsSet() {
		maxHeight = max(widget.MaxHeight()-nonContentHeight, 0)
	}
	if math.IsInf(maxWidth, 1) {
		maxWidth = -1
	}
	if math.IsInf(maxHeight, 1) {
		maxHeight = -1
	}
	return maxWidth, maxHeight
}

func widgetAspectWidthInferred(widget Widget) bool {
	state, ok := widget.(aspectDimensionState)
	return ok && state.WidthAspectInferred()
}

func widgetAspectHeightInferred(widget Widget) bool {
	state, ok := widget.(aspectDimensionState)
	return ok && state.HeightAspectInferred()
}
