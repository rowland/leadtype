// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

type StdSVG struct {
	StdComponent
	style string
}

func (svg *StdSVG) LayoutWidget(w Writer) {
	infoWidth, infoHeight, err := svg.svgDimensions(w)
	if err != nil || infoWidth <= 0 || infoHeight <= 0 {
		return
	}
	width, height := imageLikeLayoutSize(&svg.StdComponent.StdWidget, float64(infoWidth), float64(infoHeight))
	svg.ResolveWidth(width)
	svg.ResolveHeight(height)
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
			width, height := svg.placementSizeForWriter(w)
			if svg.hasInjectedStyle() {
				data, err := svg.styledSource(ref)
				if err != nil {
					return err
				}
				_, _, err = w.PrintSVG(data, ContentLeft(svg), ContentTop(svg), width, height)
				return err
			}
			_, _, err = w.PrintSVGFile(ref.identifier, ContentLeft(svg), ContentTop(svg), width, height)
			return err
		}
		body := svg.Body()
		if strings.TrimSpace(body) == "" {
			return fmt.Errorf("svg src or inline body must be specified")
		}
		width, height := svg.placementSizeForWriter(w)
		_, _, err := w.PrintSVG(svg.styledBody(body), ContentLeft(svg), ContentTop(svg), width, height)
		return err
	})
}

func (svg *StdSVG) PreferredHeight(w Writer) float64 {
	infoWidth, infoHeight, err := svg.svgDimensions(w)
	if err != nil || infoWidth == 0 {
		return NonContentHeight(svg)
	}
	_, height := imageLikeLayoutSize(&svg.StdComponent.StdWidget, float64(infoWidth), float64(infoHeight))
	return height
}

func (svg *StdSVG) PreferredWidth(w Writer) float64 {
	infoWidth, infoHeight, err := svg.svgDimensions(w)
	if err != nil || infoHeight == 0 {
		return NonContentWidth(svg)
	}
	width, _ := imageLikeLayoutSize(&svg.StdComponent.StdWidget, float64(infoWidth), float64(infoHeight))
	return width
}

func (svg *StdSVG) IntrinsicAspectRatio(w Writer) (float64, bool) {
	if w == nil {
		return 0, false
	}
	infoWidth, infoHeight, err := svg.svgDimensions(w)
	if err != nil || infoWidth <= 0 || infoHeight <= 0 {
		return 0, false
	}
	return float64(infoWidth) / float64(infoHeight), true
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
		if svg.hasInjectedStyle() {
			data, err := svg.styledSource(ref)
			if err != nil {
				return 0, 0, err
			}
			return w.SVGDimensions(data)
		}
		return w.SVGDimensionsFromFile(ref.identifier)
	}
	body := svg.Body()
	if strings.TrimSpace(body) == "" {
		return 0, 0, nil
	}
	return w.SVGDimensions(svg.styledBody(body))
}

func (svg *StdSVG) SetAttrs(attrs map[string]string) {
	svg.StdComponent.SetAttrs(attrs)
	if style, ok := attrs["style"]; ok {
		svg.style = strings.TrimSpace(style)
	}
}

func (svg *StdSVG) String() string {
	return fmt.Sprintf("StdSVG src=%s %s", svg.source.src, &svg.StdComponent.StdWidget)
}

func (svg *StdSVG) hasInjectedStyle() bool {
	return strings.TrimSpace(svg.style) != ""
}

func (svg *StdSVG) styledBody(body string) []byte {
	return injectSVGStyle([]byte(body), svg.style)
}

func (svg *StdSVG) styledSource(ref assetSourceRef) ([]byte, error) {
	data, err := readAssetSource(svg.doc, ref)
	if err != nil {
		return nil, err
	}
	return injectSVGStyle(data, svg.style), nil
}

func injectSVGStyle(data []byte, style string) []byte {
	style = strings.TrimSpace(style)
	if style == "" {
		return data
	}
	insert := svgStyleElement(style)
	offset := svgRootStartTagEnd(data)
	if offset < 0 {
		out := make([]byte, 0, len(insert)+len(data))
		out = append(out, insert...)
		out = append(out, data...)
		return out
	}
	if slash := svgRootSelfClosingSlash(data, offset); slash >= 0 {
		closeRoot := []byte("</svg>")
		out := make([]byte, 0, len(data)+len(insert)+len(closeRoot)-1)
		out = append(out, data[:slash]...)
		out = append(out, data[slash+1:offset]...)
		out = append(out, insert...)
		out = append(out, closeRoot...)
		out = append(out, data[offset:]...)
		return out
	}
	out := make([]byte, 0, len(data)+len(insert))
	out = append(out, data[:offset]...)
	out = append(out, insert...)
	out = append(out, data[offset:]...)
	return out
}

func svgStyleElement(style string) []byte {
	return []byte("<style><![CDATA[" + strings.ReplaceAll(style, "]]>", "]]]]><![CDATA[>") + "]]></style>")
}

func svgRootStartTagEnd(data []byte) int {
	text := string(data)
	start := strings.Index(text, "<svg")
	if start < 0 {
		return -1
	}
	quote := rune(0)
	for offset, r := range text[start:] {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			}
		case r == '\'' || r == '"':
			quote = r
		case r == '>':
			return start + offset + len(">")
		}
	}
	return -1
}

func svgRootSelfClosingSlash(data []byte, offset int) int {
	for i := offset - len(">") - 1; i >= 0; i-- {
		switch data[i] {
		case ' ', '\t', '\n', '\r':
			continue
		case '/':
			return i
		default:
			return -1
		}
	}
	return -1
}

func (svg *StdSVG) placementSizeForWriter(w Writer) (width, height *float64) {
	infoWidth, infoHeight, err := svg.svgDimensions(w)
	if err != nil {
		return imageLikeFallbackPlacementSize(&svg.StdComponent.StdWidget)
	}
	return imageLikePlacementSize(&svg.StdComponent.StdWidget, float64(infoWidth), float64(infoHeight))
}

func init() {
	registerTagExt(DefaultSpace, "svg", func() any { return &StdSVG{} })
}

var _ HasAttrs = (*StdSVG)(nil)
var _ Identifier = (*StdSVG)(nil)
var _ Printer = (*StdSVG)(nil)
var _ Component = (*StdSVG)(nil)
var _ IntrinsicAspectRatioProvider = (*StdSVG)(nil)
var _ WantsDoc = (*StdSVG)(nil)
var _ WantsContainer = (*StdSVG)(nil)
var _ WantsScope = (*StdSVG)(nil)
