// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
)

func renderCanvasCapture(canvas *StdCanvas, w Writer) error {
	if canvas == nil {
		return nil
	}
	doc := documentForContainer(canvas.Container())
	if doc == nil {
		return fmt.Errorf("canvas %q document is not set", canvas.Key())
	}
	if err := doc.pushCanvasCapture(canvas.Key()); err != nil {
		return err
	}
	defer doc.popCanvasCapture()

	resetCanvasWidgetRenderState(canvas)
	return withDocumentVisualCapture(doc, func() error {
		canvas.LayoutWidget(w)
		return Print(canvas, w)
	})
}

func withDocumentVisualCapture(doc *StdDocument, fn func() error) error {
	if fn == nil {
		return nil
	}
	if doc == nil {
		return fn()
	}
	doc.visualCaptureDepth++
	defer func() {
		doc.visualCaptureDepth--
	}()
	return fn()
}

func documentVisualCaptureActive(doc *StdDocument) bool {
	return doc != nil && doc.visualCaptureDepth > 0
}

func resetCanvasWidgetRenderState(root Widget) {
	walkWidgets(root, func(widget Widget) bool {
		widget.SetPrinted(false)
		widget.SetVisible(true)
		widget.SetDisabled(false)
		switch value := widget.(type) {
		case *StdContainer:
			value.activeChildren = nil
		case *StdIndex:
			value.clearSplitOverride()
		}
		return true
	})
}

func (d *StdDocument) pushCanvasCapture(key string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil
	}
	for _, active := range d.canvasCaptureStack {
		if active == key {
			path := append(append([]string(nil), d.canvasCaptureStack...), key)
			return fmt.Errorf("recursive canvas draw detected: %s", strings.Join(path, " -> "))
		}
	}
	d.canvasCaptureStack = append(d.canvasCaptureStack, key)
	return nil
}

func (d *StdDocument) popCanvasCapture() {
	if d == nil || len(d.canvasCaptureStack) == 0 {
		return
	}
	d.canvasCaptureStack = d.canvasCaptureStack[:len(d.canvasCaptureStack)-1]
}
