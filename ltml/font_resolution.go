package ltml

type explicitFontCarrier interface {
	explicitFont() *FontStyle
}

func applyContainerFont(w Writer, container Container) {
	font := fontForContainer(container)
	if font == nil {
		return
	}
	font.applyWithSize(w, effectiveFontSizeForContainer(container))
}

func applyExplicitFontForContainer(w Writer, container Container, font *FontStyle) {
	if font == nil {
		return
	}
	font.applyWithSize(w, explicitFontSizeForContainer(container, font))
}

func applyTextPieceFontForContainer(w Writer, container Container, piece textPiece, fallback *FontStyle) *FontStyle {
	font, explicit := piece.Font(fallback)
	if explicit {
		applyExplicitFontForContainer(w, container, font)
	} else {
		applyContainerFont(w, container)
	}
	return font
}

func applyWidgetFont(w Writer, widget Widget) {
	font := fontForWidget(widget)
	if font == nil {
		return
	}
	font.applyWithSize(w, effectiveFontSizeForWidget(widget))
}

func effectiveFontSizeForWidget(widget Widget) float64 {
	if widget == nil {
		return defaultFontSize
	}
	if container, ok := widget.(Container); ok {
		return effectiveFontSizeForContainer(container)
	}
	if withFont, ok := widget.(explicitFontCarrier); ok {
		if font := withFont.explicitFont(); font != nil {
			return font.ResolveAgainstBase(rootFontSizeForWidget(widget))
		}
	}
	if withContainer, ok := widget.(interface{ Container() Container }); ok {
		return effectiveFontSizeForContainer(withContainer.Container())
	}
	return defaultFontSize
}

func rootFontSizeForWidget(widget Widget) float64 {
	if widget == nil {
		return defaultFontSize
	}
	if container, ok := widget.(Container); ok {
		return rootFontSizeForContainer(container)
	}
	if withContainer, ok := widget.(interface{ Container() Container }); ok {
		return rootFontSizeForContainer(withContainer.Container())
	}
	return defaultFontSize
}

func fontForWidget(widget Widget) *FontStyle {
	if widget == nil {
		return defaultFont
	}
	if container, ok := widget.(Container); ok {
		return fontForContainer(container)
	}
	if withFont, ok := widget.(HasFont); ok {
		return withFont.Font()
	}
	return defaultFont
}

func explicitFontSizeForContainer(container Container, font *FontStyle) float64 {
	return font.ResolveAgainstBase(rootFontSizeForContainer(container))
}

func effectiveFontSizeForContainer(container Container) float64 {
	if container == nil {
		return defaultFontSize
	}
	switch value := container.(type) {
	case *StdDocument:
		return documentRootFontSize(value)
	case *StdPage:
		return pageRootFontSize(value)
	}
	if withFont, ok := container.(explicitFontCarrier); ok {
		if font := withFont.explicitFont(); font != nil {
			return explicitFontSizeForContainer(container, font)
		}
	}
	return effectiveFontSizeForContainer(container.Container())
}

func rootFontSizeForContainer(container Container) float64 {
	if container == nil {
		return defaultFontSize
	}
	switch value := container.(type) {
	case *StdDocument:
		return documentRootFontSize(value)
	case *StdPage:
		return pageRootFontSize(value)
	}
	if page := pageForContainer(container); page != nil {
		return pageRootFontSize(page)
	}
	if doc := documentForContainer(container); doc != nil {
		return documentRootFontSize(doc)
	}
	return defaultFontSize
}

func documentRootFontSize(doc *StdDocument) float64 {
	if doc == nil || doc.font == nil {
		return defaultFontSize
	}
	return doc.font.ResolveAgainstBase(defaultFontSize)
}

func pageRootFontSize(page *StdPage) float64 {
	if page == nil {
		return defaultFontSize
	}
	base := documentRootFontSize(page.document())
	if page.font == nil {
		return base
	}
	return page.font.ResolveAgainstBase(base)
}

func pageForContainer(container Container) *StdPage {
	for container != nil {
		if page, ok := container.(*StdPage); ok {
			return page
		}
		container = container.Container()
	}
	return nil
}

func fontForContainer(container Container) *FontStyle {
	if container == nil {
		return defaultFont
	}
	switch value := container.(type) {
	case *StdDocument:
		return value.Font()
	case *StdPage:
		return value.Font()
	default:
		if withFont, ok := container.(HasFont); ok {
			return withFont.Font()
		}
		return defaultFont
	}
}
