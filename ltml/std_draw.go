// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

// CanvasDrawer is an optional writer capability for placing LTML canvas assets.
type CanvasDrawer interface {
	DrawCanvas(key string, x, y, width, height, canvasWidth, canvasHeight float64, draw func(any) error) error
}

type StdDraw struct {
	StdWidget
	key string
}

func (d *StdDraw) DrawContent(w Writer) error {
	return withGraphicAccessibility(w, &d.StdWidget, "Figure", func() error {
		canvas, err := d.resolveCanvas()
		if err != nil {
			return err
		}
		drawer, ok := w.(CanvasDrawer)
		if !ok {
			return fmt.Errorf("draw requires a canvas-capable writer")
		}
		width, height := d.placementSize(canvas)
		if width <= 0 || height <= 0 {
			return nil
		}
		return drawer.DrawCanvas(
			canvas.Key(),
			ContentLeft(d),
			ContentTop(d),
			width,
			height,
			canvas.Width(),
			canvas.Height(),
			func(raw any) error {
				capture, ok := raw.(Writer)
				if !ok {
					return fmt.Errorf("canvas capture writer %T does not implement ltml.Writer", raw)
				}
				return renderCanvasCapture(canvas, capture)
			},
		)
	})
}

func (d *StdDraw) PreferredHeight(Writer) float64 {
	if d.height != 0 {
		return float64(d.height)
	}
	naturalWidth, naturalHeight, ok := d.naturalSize()
	if !ok || naturalWidth <= 0 {
		return NonContentHeight(d)
	}
	if d.width != 0 {
		return float64(d.width)*naturalHeight/naturalWidth + NonContentHeight(d)
	}
	return naturalHeight + NonContentHeight(d)
}

func (d *StdDraw) PreferredWidth(Writer) float64 {
	if d.width != 0 {
		return float64(d.width)
	}
	naturalWidth, naturalHeight, ok := d.naturalSize()
	if !ok || naturalHeight <= 0 {
		return NonContentWidth(d)
	}
	if d.height != 0 {
		return float64(d.height)*naturalWidth/naturalHeight + NonContentWidth(d)
	}
	return naturalWidth + NonContentWidth(d)
}

func (d *StdDraw) SetAttrs(attrs map[string]string) {
	d.StdWidget.SetAttrs(attrs)
	d.key = strings.TrimSpace(attrs["key"])
}

func (d *StdDraw) String() string {
	return fmt.Sprintf("StdDraw key=%s %s", d.key, &d.StdWidget)
}

func (d *StdDraw) placementSize(canvas *StdCanvas) (width, height float64) {
	if canvas == nil {
		return 0, 0
	}
	if d.WidthIsSet() && d.HeightIsSet() {
		return ContentWidth(d), ContentHeight(d)
	}
	if d.WidthIsSet() {
		width = ContentWidth(d)
		if canvas.Width() == 0 {
			return width, 0
		}
		return width, width * canvas.Height() / canvas.Width()
	}
	if d.HeightIsSet() {
		height = ContentHeight(d)
		if canvas.Height() == 0 {
			return 0, height
		}
		return height * canvas.Width() / canvas.Height(), height
	}
	return canvas.Width(), canvas.Height()
}

func (d *StdDraw) naturalSize() (width, height float64, ok bool) {
	canvas, err := d.resolveCanvas()
	if err != nil || canvas == nil {
		return 0, 0, false
	}
	return canvas.Width(), canvas.Height(), true
}

func (d *StdDraw) resolveCanvas() (*StdCanvas, error) {
	if strings.TrimSpace(d.key) == "" {
		return nil, fmt.Errorf("<draw> requires a key")
	}
	doc := documentForContainer(d.container)
	if doc == nil {
		return nil, fmt.Errorf("draw document is not set")
	}
	canvas := doc.Canvas(d.key)
	if canvas == nil {
		return nil, fmt.Errorf("missing canvas key %q", d.key)
	}
	return canvas, nil
}

func init() {
	registerTag(DefaultSpace, "draw", func() any { return &StdDraw{} })
}

var _ HasAttrs = (*StdDraw)(nil)
var _ Identifier = (*StdDraw)(nil)
var _ Printer = (*StdDraw)(nil)
var _ WantsContainer = (*StdDraw)(nil)
var _ WantsDoc = (*StdDraw)(nil)
var _ WantsScope = (*StdDraw)(nil)
