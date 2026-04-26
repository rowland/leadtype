// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rowland/leadtype/pdf"
)

type StdDocument struct {
	StdPage
	documentPageNo             int
	physicalPageNo             int
	pendingStart               *int
	ua                         bool
	networkAssets              bool
	compressPages              bool
	compressToUnicode          bool
	compressEmbeddedFonts      bool
	svgGradientStopOpacityMode string
	renderContext              *documentRenderContext
	canvases                   map[string]*StdCanvas
	canvasCaptureStack         []string
	visualCaptureDepth         int
}

func (d *StdDocument) Font() *FontStyle {
	if d.font == nil {
		return defaultFont
	}
	return d.font
}

func (d *StdDocument) Page(i int) *StdPage {
	if i < len(d.children) {
		return d.children[i].(*StdPage)
	}
	return nil
}

func (d *StdDocument) PageStyle() *PageStyle {
	if d.pageStyle == nil {
		style := PageStyleFor("letter", d.scope)
		if style == nil {
			panic("default page style missing")
		}
		return style
	}
	return d.pageStyle
}

func (d *StdDocument) ParagraphStyle() *ParagraphStyle {
	if d.paragraphStyle == nil {
		return defaultParagraphStyle
	}
	return d.paragraphStyle
}

func (d *StdDocument) CurrentPageNo() int {
	return d.documentPageNo
}

func (d *StdDocument) CurrentPhysicalPageNo() int {
	return d.physicalPageNo
}

func (d *StdDocument) SetCurrentPageStart(start int) {
	d.documentPageNo = start - 1
}

func (d *StdDocument) SetPendingStart(start int) {
	d.pendingStart = &start
}

func (d *StdDocument) Print(w Writer) error {
	if err := d.validateCanvasAssets(); err != nil {
		return err
	}
	applyWriterAccessibility(w, d)
	d.applyWriterCompression(w)
	d.applyWriterSVGCompatibility(w)
	return d.printWithIndexes(w)
}

func (d *StdDocument) Canvas(key string) *StdCanvas {
	if d == nil || d.canvases == nil {
		return nil
	}
	return d.canvases[key]
}

func (d *StdDocument) registerCanvas(canvas *StdCanvas) error {
	if d == nil || canvas == nil {
		return nil
	}
	if err := canvas.validateDefinition(); err != nil {
		return err
	}
	if d.canvases == nil {
		d.canvases = make(map[string]*StdCanvas)
	}
	key := canvas.Key()
	if _, exists := d.canvases[key]; exists {
		return fmt.Errorf("duplicate canvas key %q", key)
	}
	d.canvases[key] = canvas
	return nil
}

func (d *StdDocument) eachCanvas(fn func(*StdCanvas)) {
	if d == nil || len(d.canvases) == 0 || fn == nil {
		return
	}
	keys := make([]string, 0, len(d.canvases))
	for key := range d.canvases {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fn(d.canvases[key])
	}
}

func (d *StdDocument) validateCanvasAssets() error {
	if d == nil {
		return nil
	}
	missing := make(map[string]struct{})
	walkWidgets(d, func(widget Widget) bool {
		draw, ok := widget.(*StdDraw)
		if !ok {
			return true
		}
		key := strings.TrimSpace(draw.key)
		if key == "" {
			missing["<blank>"] = struct{}{}
			return true
		}
		if d.Canvas(key) == nil {
			missing[key] = struct{}{}
		}
		return true
	})
	if len(missing) == 0 {
		return nil
	}
	keys := make([]string, 0, len(missing))
	for key := range missing {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) == 1 && keys[0] == "<blank>" {
		return fmt.Errorf("<draw> requires a key")
	}
	return fmt.Errorf("missing canvas definition(s): %s", strings.Join(keys, ", "))
}

func (d *StdDocument) NetworkAssetsEnabled() bool {
	return d.networkAssets
}

func (d *StdDocument) SetAttrs(attrs map[string]string) {
	d.StdPage.SetAttrs(attrs)
	if value, ok := attrs["compress-pages"]; ok {
		d.compressPages = value == "true"
	}
	if value, ok := attrs["compress-to-unicode"]; ok {
		d.compressToUnicode = value == "true"
	}
	if value, ok := attrs["compress-embedded-fonts"]; ok {
		d.compressEmbeddedFonts = value == "true"
	}
	if value, ok := attrs["ua"]; ok {
		d.ua = value == "true"
	}
	if value, ok := attrs["network-assets"]; ok {
		d.networkAssets = value == "true"
	}
	if value, ok := attrs["svg-gradient-stop-opacity-mode"]; ok {
		d.svgGradientStopOpacityMode = value
	}
}

func (d *StdDocument) applyWriterCompression(w Writer) {
	if cw, ok := w.(interface{ CompressPages(bool) *pdf.DocWriter }); ok {
		cw.CompressPages(d.compressPages)
	}
	if cw, ok := w.(interface{ CompressToUnicode(bool) *pdf.DocWriter }); ok {
		cw.CompressToUnicode(d.compressToUnicode)
	}
	if cw, ok := w.(interface{ CompressEmbeddedFonts(bool) *pdf.DocWriter }); ok {
		cw.CompressEmbeddedFonts(d.compressEmbeddedFonts)
	}
}

func (d *StdDocument) applyWriterSVGCompatibility(w Writer) {
	if d.svgGradientStopOpacityMode == "" {
		return
	}
	if sw, ok := w.(interface{ SetSVGGradientStopOpacityMode(string) string }); ok {
		sw.SetSVGGradientStopOpacityMode(d.svgGradientStopOpacityMode)
	}
}

func (d *StdDocument) String() string {
	return fmt.Sprintf("StdDocument %s units=%s margin=%s", &d.Identity, d.units, &d.margin)
}

func init() {
	registerTag(DefaultSpace, "ltml", func() any { return &StdDocument{} })
}

var _ HasAttrs = (*StdDocument)(nil)
var _ HasScope = (*StdDocument)(nil)
var _ Identifier = (*StdDocument)(nil)
