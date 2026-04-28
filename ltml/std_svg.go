// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

type StdSVG struct {
	StdComponent
}

func (svg *StdSVG) LayoutWidget(w Writer) {
}

func (svg *StdSVG) DrawContent(w Writer) error {
	return withGraphicAccessibility(w, &svg.StdComponent.StdWidget, "Figure", func() error {
		if svg.source.Explicit() {
			ref, err := svg.assetSource()
			if err != nil {
				return err
			}
			if ref.identifier == "" {
				return fmt.Errorf("svg src or inline body must be specified")
			}
			_, _, err = w.PrintSVGFile(ref.identifier, ContentLeft(svg), ContentTop(svg), svg.widthForWriter(), svg.heightForWriter())
			return err
		}
		body := svg.Body()
		if strings.TrimSpace(body) == "" {
			return fmt.Errorf("svg src or inline body must be specified")
		}
		_, _, err := w.PrintSVG([]byte(body), ContentLeft(svg), ContentTop(svg), svg.widthForWriter(), svg.heightForWriter())
		return err
	})
}

func (svg *StdSVG) PreferredHeight(w Writer) float64 {
	if svg.height != 0 {
		return float64(svg.height)
	}
	infoWidth, infoHeight, err := svg.svgDimensions(w)
	if err != nil || infoWidth == 0 {
		return NonContentHeight(svg)
	}
	if svg.width != 0 {
		return float64(svg.width)*float64(infoHeight)/float64(infoWidth) + NonContentHeight(svg)
	}
	return float64(infoHeight) + NonContentHeight(svg)
}

func (svg *StdSVG) PreferredWidth(w Writer) float64 {
	if svg.width != 0 {
		return float64(svg.width)
	}
	infoWidth, infoHeight, err := svg.svgDimensions(w)
	if err != nil || infoHeight == 0 {
		return NonContentWidth(svg)
	}
	if svg.height != 0 {
		return float64(svg.height)*float64(infoWidth)/float64(infoHeight) + NonContentWidth(svg)
	}
	return float64(infoWidth) + NonContentWidth(svg)
}

func (svg *StdSVG) svgDimensions(w Writer) (width, height int, err error) {
	if svg.source.Explicit() {
		ref, err := svg.assetSource()
		if err != nil {
			return 0, 0, err
		}
		if ref.identifier == "" {
			return 0, 0, nil
		}
		return w.SVGDimensionsFromFile(ref.identifier)
	}
	body := svg.Body()
	if strings.TrimSpace(body) == "" {
		return 0, 0, nil
	}
	return w.SVGDimensions([]byte(body))
}

func (svg *StdSVG) SetAttrs(attrs map[string]string) {
	svg.StdComponent.SetAttrs(attrs)
}

func (svg *StdSVG) String() string {
	return fmt.Sprintf("StdSVG src=%s %s", svg.source.src, &svg.StdComponent.StdWidget)
}

func (svg *StdSVG) widthForWriter() *float64 {
	if svg.WidthIsSet() {
		width := ContentWidth(svg)
		return &width
	}
	return nil
}

func (svg *StdSVG) heightForWriter() *float64 {
	if svg.HeightIsSet() {
		height := ContentHeight(svg)
		return &height
	}
	return nil
}

func init() {
	registerTagExt(DefaultSpace, "svg", func() any { return &StdSVG{} })
}

var _ HasAttrs = (*StdSVG)(nil)
var _ Identifier = (*StdSVG)(nil)
var _ Printer = (*StdSVG)(nil)
var _ Component = (*StdSVG)(nil)
var _ WantsDoc = (*StdSVG)(nil)
var _ WantsContainer = (*StdSVG)(nil)
var _ WantsScope = (*StdSVG)(nil)
