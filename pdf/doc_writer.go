// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
	"io"

	"leadtype/codepage"
)

type DocWriter struct {
	pages       []*PageWriter
	nextSeq     func() int
	file        *file
	catalog     *catalog
	resources   *resources
	pagesAcross int
	pagesDown   int
	curPage     *PageWriter
	options     Options
	fontSources map[string]FontSource
	fontKeys    map[string]string
}

func NewDocWriter() *DocWriter {
	nextSeq := nextSeqFunc()
	file := newFile()
	pages := newPages(nextSeq(), 0)
	outlines := newOutlines(nextSeq(), 0)
	catalog := newCatalog(nextSeq(), 0, "UseNone", pages, outlines)
	file.body.add(pages, outlines, catalog)
	file.trailer.setRoot(catalog)
	resources := newResources(nextSeq(), 0)
	resources.setProcSet(nameArray("PDF", "Text", "ImageB", "ImageC"))
	file.body.add(resources)
	fontSources := make(map[string]FontSource, 2)
	fontKeys := make(map[string]string)
	return &DocWriter{nextSeq: nextSeq, file: file, catalog: catalog, resources: resources, options: Options{}, fontSources: fontSources, fontKeys: fontKeys}
}

func nextSeqFunc() func() int {
	var nextValue = 0
	return func() int {
		nextValue++
		return nextValue
	}
}

func (dw *DocWriter) AddFont(family string, subType string, options Options) ([]*Font, error) {
	return dw.curPage.AddFont(family, subType, options)
}

func (dw *DocWriter) AddFontSource(fontSource FontSource, subType string) {
	dw.fontSources[subType] = fontSource
}

func (dw *DocWriter) WriteTo(wr io.Writer) {
	if len(dw.pages) == 0 {
		dw.NewPageWithOptions(Options{})
	}
	for _, pw := range dw.pages {
		pw.close()
	}
	dw.curPage = nil
	dw.file.write(wr)
}

func (dw *DocWriter) FontColor() Color {
	return dw.curPage.FontColor()
}

func (dw *DocWriter) fontKey(f *Font, cpi codepage.CodepageIndex) string {
	if f == nil {
		panic("fontKey: No font specified.")
	}
	if f.metrics == nil {
		panic("fontKey: font missing metrics.")
	}
	name := fmt.Sprintf("%s/%s-%s", f.metrics.PostScriptName(), cpi, f.subType)
	if key, ok := dw.fontKeys[name]; ok {
		return key
	}
	descriptor := newFontDescriptor(
		dw.nextSeq(), 0,
		f.metrics.PostScriptName(), f.metrics.Family(),
		0, // flags
		f.metrics.BoundingBox(),
		0, // missingWidth
		f.metrics.StemV(),
		0, // stemH
		f.metrics.ItalicAngle(),
		f.metrics.CapHeight(),
		f.metrics.XHeight(),
		f.metrics.Ascent(),
		f.metrics.Descent(),
		f.metrics.Leading(),
		0, 0) // maxWidth, avgWidth
	dw.file.body.add(descriptor)
	key := fmt.Sprintf("F%d", len(dw.fontKeys))
	dw.fontKeys[name] = key
	widths := dw.widthsForFontCodepage(f, cpi)
	dw.file.body.add(widths)
	font := newSimpleFont(
		dw.nextSeq(), 0,
		f.subType,
		f.metrics.PostScriptName(),
		32, 255, widths,
		descriptor)
	dw.file.body.add(font)
	dw.resources.fonts[key] = &indirectObjectRef{font}
	return key
}

func (dw *DocWriter) Fonts() []*Font {
	return dw.curPage.Fonts()
}

func (dw *DocWriter) FontSize() float64 {
	return dw.curPage.FontSize()
}

func (dw *DocWriter) FontStyle() string {
	return dw.curPage.FontStyle()
}

func (dw *DocWriter) indexOfPage(pw *PageWriter) int {
	for i, v := range dw.pages {
		if v == pw {
			return i
		}
	}
	return -1
}

func (dw *DocWriter) inPage() bool {
	return dw.curPage != nil
}

func (dw *DocWriter) insertPage(pw *PageWriter, index int) {
	pages := make([]*PageWriter, 0, len(dw.pages)+1)
	pages = append(pages, dw.pages[:index+1]...)
	pages = append(pages, pw)
	pages = append(pages, dw.pages[index+1:]...)
	dw.pages = pages
}

func (dw *DocWriter) LineCapStyle() LineCapStyle {
	return dw.curPage.LineCapStyle()
}

func (dw *DocWriter) LineColor() Color {
	return dw.curPage.LineColor()
}

func (dw *DocWriter) LineDashPattern() string {
	return dw.curPage.LineDashPattern()
}

func (dw *DocWriter) LineThrough() bool {
	return dw.curPage.LineThrough()
}

func (dw *DocWriter) LineTo(x, y float64) {
	dw.curPage.LineTo(x, y)
}

func (dw *DocWriter) LineWidth(units string) float64 {
	return dw.curPage.LineWidth(units)
}

func (dw *DocWriter) MoveTo(x, y float64) {
	dw.curPage.MoveTo(x, y)
}

func (dw *DocWriter) NewPage() *PageWriter {
	if dw.curPage == nil {
		return dw.NewPageWithOptions(Options{})
	}
	return dw.NewPageAfter(dw.curPage)
}

func (dw *DocWriter) NewPageAfter(pw *PageWriter) *PageWriter {
	var i int
	if pw == nil {
		i = len(dw.pages)
	} else {
		i = dw.indexOfPage(pw)
	}
	if i >= 0 {
		dw.curPage = clonePageWriter(pw)
		dw.insertPage(dw.curPage, i)
		return dw.curPage
	}
	return nil
}

func (dw *DocWriter) NewPageWithOptions(options Options) *PageWriter {
	dw.curPage = newPageWriter(dw, dw.options.Merge(options))
	dw.pages = append(dw.pages, dw.curPage)
	return dw.curPage
}

func (dw *DocWriter) PagesAcross() int {
	if dw.pagesAcross < 1 {
		return 1
	}
	return dw.pagesAcross
}

func (dw *DocWriter) PagesDown() int {
	if dw.pagesDown < 1 {
		return 1
	}
	return dw.pagesDown
}

func (dw *DocWriter) PagesUp() int {
	return dw.PagesAcross() * dw.PagesDown()
}

func (dw *DocWriter) Print(text string) (err error) {
	return dw.curPage.Print(text)
}

func (dw *DocWriter) PrintRichText(text *RichText) {
	dw.curPage.PrintRichText(text)
}

func (dw *DocWriter) ResetFonts() {
	dw.curPage.ResetFonts()
}

func (dw *DocWriter) SetFont(name string, size float64, subType string, options Options) ([]*Font, error) {
	return dw.curPage.SetFont(name, size, subType, options)
}

func (dw *DocWriter) SetFontColor(color interface{}) (lastColor Color) {
	return dw.curPage.SetFontColor(color)
}

func (dw *DocWriter) SetFontSize(size float64) (prev float64) {
	return dw.curPage.SetFontSize(size)
}

func (dw *DocWriter) SetFontStyle(style string) (prev string, err error) {
	return dw.curPage.SetFontStyle(style)
}

func (dw *DocWriter) SetLineColor(color Color) (prev Color) {
	return dw.curPage.SetLineColor(color)
}

func (dw *DocWriter) SetLineDashPattern(lineDashPattern string) (prev string) {
	return dw.curPage.SetLineDashPattern(lineDashPattern)
}

func (dw *DocWriter) SetLineThrough(lineThrough bool) (prev bool) {
	return dw.curPage.SetLineThrough(lineThrough)
}

func (dw *DocWriter) SetLineWidth(width float64, units string) (prev float64) {
	return dw.curPage.SetLineWidth(width, units)
}

func (dw *DocWriter) SetOptions(options Options) {
	dw.options = options
}

func (dw *DocWriter) SetUnderline(underline bool) (prev bool) {
	return dw.curPage.SetUnderline(underline)
}

func (dw *DocWriter) SetUnits(units string) {
	dw.options["units"] = units
	if dw.curPage != nil {
		dw.curPage.SetUnits(units)
	}
}

func (dw *DocWriter) Underline() bool {
	return dw.curPage.Underline()
}

func (dw *DocWriter) widthsForFontCodepage(f *Font, cpi codepage.CodepageIndex) *indirectObject {
	var widths [256]int
	upm := f.metrics.UnitsPerEm()
	// Avoid divide by zero error for unusual fonts.
	if upm > 0 {
		for i, r := range cpi.Map() {
			designWidth, _ := f.metrics.AdvanceWidth(r)
			widths[i] = designWidth * 1000 / upm
		}
	}
	pdfWidths := arrayFromInts(widths[32:])
	ioWidths := &indirectObject{dw.nextSeq(), 0, &pdfWidths}
	return ioWidths
}

func (dw *DocWriter) Write(text []byte) (n int, err error) {
	return dw.curPage.Write(text)
}
