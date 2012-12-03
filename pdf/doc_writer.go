// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"io"
)

type DocWriter struct {
	wr          io.Writer
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
}

func NewDocWriter(wr io.Writer) *DocWriter {
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
	return &DocWriter{wr: wr, nextSeq: nextSeq, file: file, catalog: catalog, resources: resources, fontSources: fontSources}
}

func nextSeqFunc() func() int {
	var nextValue = 0
	return func() int {
		nextValue++
		return nextValue
	}
}

func (dw *DocWriter) AddFont(family string, options Options) ([]*Font, error) {
	return dw.curPage.AddFont(family, options)
}

func (dw *DocWriter) AddFontSource(fontSource FontSource, subType string) {
	dw.fontSources[subType] = fontSource
}

func (dw *DocWriter) Close() {
	if len(dw.pages) == 0 {
		dw.OpenPageWithOptions(Options{})
	}
	for _, pw := range dw.pages {
		pw.Close()
	}
	dw.curPage = nil
	dw.file.write(dw.wr)
}

func (dw *DocWriter) ClosePage() {
	if dw.inPage() {
		dw.curPage.Close()
		dw.curPage = nil
	}
}

func (dw *DocWriter) FontColor() Color {
	return dw.curPage.FontColor()
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

func (dw *DocWriter) LineTo(x, y float64) {
	dw.curPage.LineTo(x, y)
}

func (dw *DocWriter) LineWidth(units string) float64 {
	return dw.curPage.LineWidth(units)
}

func (dw *DocWriter) MoveTo(x, y float64) {
	dw.curPage.MoveTo(x, y)
}

func (dw *DocWriter) Open() {
	dw.OpenWithOptions(Options{})
}

func (dw *DocWriter) OpenWithOptions(options Options) {
	dw.options = options
}

func (dw *DocWriter) OpenPage() *PageWriter {
	if dw.curPage == nil {
		return dw.OpenPageWithOptions(Options{})
	}
	return dw.OpenPageAfter(dw.curPage)
}

func (dw *DocWriter) OpenPageAfter(pw *PageWriter) *PageWriter {
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

func (dw *DocWriter) OpenPageWithOptions(options Options) *PageWriter {
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

func (dw *DocWriter) ResetFonts() {
	dw.curPage.ResetFonts()
}

func (dw *DocWriter) SetFont(name string, size float64, options Options) ([]*Font, error) {
	return dw.curPage.SetFont(name, size, options)
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

func (dw *DocWriter) SetLineWidth(width float64, units string) (prev float64) {
	return dw.curPage.SetLineWidth(width, units)
}

func (dw *DocWriter) SetUnits(units string) {
	dw.curPage.SetUnits(units)
}
