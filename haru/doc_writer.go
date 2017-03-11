// Copyright 2017 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package haru

import (
	"fmt"
	"io"
	"strings"

	"github.com/rowland/hpdf"

	"github.com/rowland/leadtype/codepage"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

type DocWriter struct {
	pdf *hpdf.PDF

	pages       []*PageWriter
	pagesAcross int
	pagesDown   int
	curPage     *PageWriter
	options     options.Options
	fontSources font.FontSources
	fontHandles map[string]*hpdf.Font
	fontNames   map[string]string
	// fontEncodings map[string]*fontEncoding
}

func NewDocWriter() *DocWriter {
	pdf, _ := hpdf.New()

	fontSources := make(font.FontSources, 0, 2)
	fontHandles := make(map[string]*hpdf.Font)
	fontNames := make(map[string]string)
	// fontEncodings := make(map[string]*fontEncoding)
	return &DocWriter{
		pdf:         pdf,
		options:     options.Options{},
		fontSources: fontSources,
		fontHandles: fontHandles,
		fontNames:   fontNames,
		// fontEncodings: fontEncodings
	}
}

func nextSeqFunc() func() int {
	var nextValue = 0
	return func() int {
		nextValue++
		return nextValue
	}
}

func (dw *DocWriter) AddFont(family string, options options.Options) ([]*font.Font, error) {
	return dw.CurPage().AddFont(family, options)
}

func (dw *DocWriter) AddFontSource(fontSource font.FontSource) {
	dw.fontSources = append(dw.fontSources, fontSource)
}

func (dw *DocWriter) CurPage() *PageWriter {
	if dw.curPage == nil {
		return dw.NewPage()
	}
	return dw.curPage
}

func (dw *DocWriter) FontColor() colors.Color {
	return dw.CurPage().FontColor()
}

func toHaruEncoding(encoding string) string {
	// These encodings do not match exactly, but libharu does not include ISO-8859-1.
	if encoding == "ISO-8859-1" {
		return "WinAnsiEncoding"
	}
	// libharu does not use the dash after ISO.
	return strings.Replace(encoding, "ISO-8859", "ISO8859", 1)
}

var standardEncodedFonts = map[string]bool{
	"Symbol":       true,
	"ZapfDingbats": true,
}

func (dw *DocWriter) fontHandle(f *font.Font, cpi codepage.CodepageIndex) *hpdf.Font {
	if f == nil {
		panic("fontHandle: no font specified.")
	}
	if !f.HasMetrics() {
		panic("fontHandle: font missing metrics.")
	}
	name := fmt.Sprintf("%s/%s-%s", f.PostScriptName(), cpi, f.SubType())
	if handle, ok := dw.fontHandles[name]; ok {
		return handle
	}
	fontName, ok := dw.fontNames[f.Filename()]
	if !ok {
		var err error
		if standardEncodedFonts[f.PostScriptName()] {
			fontName = f.PostScriptName()
		} else {
			// fmt.Printf("dw.pdf.LoadType1FontFromFile(%s)\n", f.Filename())
			fontName, err = dw.pdf.LoadType1FontFromFile(f.Filename())
			if err != nil {
				panic("fontHandle: " + err.Error())
			}
		}
		dw.fontNames[f.Filename()] = fontName
	}
	var err error
	var handle *hpdf.Font
	if standardEncodedFonts[f.PostScriptName()] {
		// fmt.Printf("dw.pdf.GetFont(%s)\n", fontName)
		handle, err = dw.pdf.GetFont(fontName)
	} else {
		encoding := toHaruEncoding(cpi.String())
		// fmt.Printf("dw.pdf.GetFont(%s, %s)\n", fontName, encoding)
		handle, err = dw.pdf.GetFont(fontName, encoding)
	}
	if err != nil {
		panic("fontHandle: " + err.Error())
	}
	dw.fontHandles[name] = handle
	return handle
}

func (dw *DocWriter) Fonts() []*font.Font {
	return dw.CurPage().Fonts()
}

func (dw *DocWriter) FontSize() float64 {
	return dw.CurPage().FontSize()
}

func (dw *DocWriter) FontSources() font.FontSources {
	return dw.fontSources
}

func (dw *DocWriter) FontStyle() string {
	return dw.CurPage().FontStyle()
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
	return dw.CurPage().LineCapStyle()
}

func (dw *DocWriter) LineColor() colors.Color {
	return dw.CurPage().LineColor()
}

func (dw *DocWriter) LineDashPattern() string {
	return dw.CurPage().LineDashPattern()
}

func (dw *DocWriter) LineSpacing() float64 {
	return dw.CurPage().LineSpacing()
}

func (dw *DocWriter) LineTo(x, y float64) {
	dw.CurPage().LineTo(x, y)
}

func (dw *DocWriter) LineWidth(units string) float64 {
	return dw.CurPage().LineWidth(units)
}

func (dw *DocWriter) Loc() (x, y float64) {
	return dw.CurPage().Loc()
}

func (dw *DocWriter) MoveTo(x, y float64) {
	dw.CurPage().MoveTo(x, y)
}

func (dw *DocWriter) NewPage() *PageWriter {
	if dw.curPage == nil {
		return dw.NewPageWithOptions(options.Options{})
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

func (dw *DocWriter) NewPageWithOptions(options options.Options) *PageWriter {
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

func (dw *DocWriter) PageHeight() float64 {
	return dw.CurPage().PageHeight()
}

func (dw *DocWriter) PagesUp() int {
	return dw.PagesAcross() * dw.PagesDown()
}

func (dw *DocWriter) PageWidth() float64 {
	return dw.CurPage().PageWidth()
}

func (dw *DocWriter) Print(text string) (err error) {
	return dw.CurPage().Print(text)
}

func (dw *DocWriter) PrintParagraph(para []*rich_text.RichText, options options.Options) {
	dw.CurPage().PrintParagraph(para, options)
}

func (dw *DocWriter) PrintRichText(text *rich_text.RichText) {
	dw.CurPage().PrintRichText(text)
}

func (dw *DocWriter) PrintWithOptions(text string, options options.Options) (err error) {
	return dw.CurPage().PrintWithOptions(text, options)
}

func (dw *DocWriter) Rectangle(x, y, width, height float64, border bool, fill bool) {
	dw.CurPage().Rectangle(x, y, width, height, border, fill)
}

func (dw *DocWriter) Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool) {
	dw.CurPage().Rectangle2(x, y, width, height, border, fill, corners, path, reverse)
}

func (dw *DocWriter) ResetFonts() {
	dw.CurPage().ResetFonts()
}

func (dw *DocWriter) SetFillColor(color interface{}) (prev colors.Color) {
	return dw.CurPage().SetFillColor(color)
}

func (dw *DocWriter) SetFont(name string, size float64, options options.Options) ([]*font.Font, error) {
	return dw.CurPage().SetFont(name, size, options)
}

func (dw *DocWriter) SetFontColor(color interface{}) (lastColor colors.Color) {
	return dw.CurPage().SetFontColor(color)
}

func (dw *DocWriter) SetFontSize(size float64) (prev float64) {
	return dw.CurPage().SetFontSize(size)
}

func (dw *DocWriter) SetFontStyle(style string) (prev string, err error) {
	return dw.CurPage().SetFontStyle(style)
}

func (dw *DocWriter) SetLineColor(color colors.Color) (prev colors.Color) {
	return dw.CurPage().SetLineColor(color)
}

func (dw *DocWriter) SetLineDashPattern(lineDashPattern string) (prev string) {
	return dw.CurPage().SetLineDashPattern(lineDashPattern)
}

func (dw *DocWriter) SetLineSpacing(lineSpacing float64) (prev float64) {
	return dw.CurPage().SetLineSpacing(lineSpacing)
}

func (dw *DocWriter) SetStrikeout(strikeout bool) (prev bool) {
	return dw.CurPage().SetStrikeout(strikeout)
}

func (dw *DocWriter) SetLineWidth(width float64, units string) (prev float64) {
	return dw.CurPage().SetLineWidth(width, units)
}

func (dw *DocWriter) SetOptions(options options.Options) {
	dw.options = options
}

func (dw *DocWriter) SetUnderline(underline bool) (prev bool) {
	return dw.CurPage().SetUnderline(underline)
}

func (dw *DocWriter) SetUnits(units string) {
	dw.options["units"] = units
	if dw.curPage != nil {
		dw.curPage.SetUnits(units)
	}
}

func (dw *DocWriter) Strikeout() bool {
	return dw.CurPage().Strikeout()
}

func (dw *DocWriter) Underline() bool {
	return dw.CurPage().Underline()
}

func (dw *DocWriter) Write(text []byte) (n int, err error) {
	return dw.CurPage().Write(text)
}

// WriteTo implements io.WriterTo.
func (dw *DocWriter) WriteTo(wr io.Writer) (int64, error) {
	if len(dw.pages) == 0 {
		dw.NewPage()
	}
	for _, pw := range dw.pages {
		pw.close()
	}
	dw.curPage = nil
	err := dw.pdf.SaveToStream(wr)
	return 0, err
}

func (dw *DocWriter) X() float64 {
	return dw.CurPage().X()
}

func (dw *DocWriter) Y() float64 {
	return dw.CurPage().Y()
}
