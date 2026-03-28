// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"

	"github.com/rowland/leadtype/pdf"
)

type StdDocument struct {
	StdPage
	documentPageNo        int
	physicalPageNo        int
	pendingStart          *int
	compressPages         bool
	compressToUnicode     bool
	compressEmbeddedFonts bool
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
	d.applyWriterCompression(w)
	d.documentPageNo = 0
	d.physicalPageNo = 0
	d.pendingStart = nil
	return d.DrawContent(w)
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

func (d *StdDocument) String() string {
	return fmt.Sprintf("StdDocument %s units=%s margin=%s", &d.Identity, d.units, &d.margin)
}

func init() {
	registerTag(DefaultSpace, "ltml", func() interface{} { return &StdDocument{} })
}

var _ HasAttrs = (*StdDocument)(nil)
var _ HasScope = (*StdDocument)(nil)
var _ Identifier = (*StdDocument)(nil)
