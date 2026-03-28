// Copyright 2011-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/rowland/leadtype/codepage"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

type DocWriter struct {
	pages                 []*PageWriter
	nextSeq               func() int
	file                  *file
	catalog               *catalog
	resources             *resources
	pagesAcross           int
	pagesDown             int
	curPage               *PageWriter
	options               options.Options
	fontSources           font.FontSources
	fontKeys              map[string]string
	fontEncodings         map[string]*fontEncoding
	glyphRecorders        map[string]*glyphRecorder  // keyed by font PostScript name
	unicodeFonts          map[string]*font.Font      // PostScript name → font, for width lookup at Close
	cidFonts              map[string]*cidFont        // PostScript name → CID font, for /W update at Close
	type0Fonts            map[string]*type0Font      // PostScript name → Type0 font, for ToUnicode at Close
	fontDescriptors       map[string]*fontDescriptor // PostScript name → descriptor, for FontFile2 at Close
	images                map[string]*cachedImage
	assetFS               fs.FS
	compressPages         bool
	compressToUnicode     bool
	compressEmbeddedFonts bool
}

type cachedImage struct {
	image *pdfImage
	name  string
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
	fontSources := make(font.FontSources, 0, 2)
	fontKeys := make(map[string]string)
	fontEncodings := make(map[string]*fontEncoding)
	return &DocWriter{
		nextSeq:         nextSeq,
		file:            file,
		catalog:         catalog,
		resources:       resources,
		options:         options.Options{},
		fontSources:     fontSources,
		fontKeys:        fontKeys,
		fontEncodings:   fontEncodings,
		glyphRecorders:  make(map[string]*glyphRecorder),
		unicodeFonts:    make(map[string]*font.Font),
		cidFonts:        make(map[string]*cidFont),
		type0Fonts:      make(map[string]*type0Font),
		fontDescriptors: make(map[string]*fontDescriptor),
		images:          make(map[string]*cachedImage),
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

func (dw *DocWriter) CompressPages(value bool) *DocWriter {
	dw.compressPages = value
	return dw
}

func (dw *DocWriter) CompressToUnicode(value bool) *DocWriter {
	dw.compressToUnicode = value
	return dw
}

func (dw *DocWriter) CompressEmbeddedFonts(value bool) *DocWriter {
	dw.compressEmbeddedFonts = value
	return dw
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

func (dw *DocWriter) SetAssetFS(assetFS fs.FS) {
	dw.assetFS = assetFS
}

func (dw *DocWriter) readImageFile(filename string) ([]byte, error) {
	if dw.assetFS == nil {
		return os.ReadFile(filename)
	}
	if !validAssetPath(filename) {
		return nil, fmt.Errorf("invalid asset path %q", filename)
	}
	info, err := fs.Stat(dw.assetFS, filename)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("asset path %q is a directory", filename)
	}
	return fs.ReadFile(dw.assetFS, filename)
}

func useStandardEncoding(family string) bool {
	switch family {
	case "Symbol", "ZapfDingbats":
		return true
	default:
		return false
	}
}

func (dw *DocWriter) fontKey(f *font.Font, cpi codepage.CodepageIndex) string {
	if f == nil {
		panic("fontKey: No font specified.")
	}
	if !f.HasMetrics() {
		panic("fontKey: font missing metrics.")
	}
	if f.SubType() == "TrueType" {
		return dw.fontKeyUnicode(f)
	}
	name := fmt.Sprintf("%s/%s-%s", f.PostScriptName(), cpi, f.SubType())
	if key, ok := dw.fontKeys[name]; ok {
		return key
	}
	descriptor := newFontDescriptor(
		dw.nextSeq(), 0,
		f.PostScriptName(), f.Family(),
		f.Flags(),
		f.BoundingBox(),
		0, // missingWidth
		f.StemV(),
		0, // stemH
		f.ItalicAngle(),
		f.CapHeight(),
		f.XHeight(),
		f.Ascent(),
		f.Descent(),
		f.Leading(),
		0, 0) // maxWidth, avgWidth
	dw.file.body.add(descriptor)
	key := fmt.Sprintf("F%d", len(dw.fontKeys))
	dw.fontKeys[name] = key
	widths := dw.widthsForFontCodepage(f, cpi)
	dw.file.body.add(widths)
	var font *simpleFont
	switch f.SubType() {
	case "Type1":
		if useStandardEncoding(f.Family()) {
			font = newType1Font(
				dw.nextSeq(), 0,
				f.PostScriptName(),
				32, 255, widths,
				descriptor, nil)
		} else {
			encoding, ok := dw.fontEncodings[cpi.String()]
			if !ok {
				differences := glyphDiffs(codepage.Idx_CP1252, cpi, 32, 255)
				encoding = newFontEncoding(dw.nextSeq(), 0, "WinAnsiEncoding", differences)
				dw.file.body.add(encoding)
				dw.fontEncodings[cpi.String()] = encoding
			}
			font = newType1Font(
				dw.nextSeq(), 0,
				f.PostScriptName(),
				32, 255, widths,
				descriptor, &indirectObjectRef{encoding})
		}
	case "TrueType":
		encoding, ok := dw.fontEncodings[cpi.String()]
		if !ok {
			differences := glyphDiffs(codepage.Idx_CP1252, cpi, 32, 255)
			encoding = newFontEncoding(dw.nextSeq(), 0, "WinAnsiEncoding", differences)
			dw.file.body.add(encoding)
			dw.fontEncodings[cpi.String()] = encoding
		}
		font = newTrueTypeFont(
			dw.nextSeq(), 0,
			f.PostScriptName(),
			32, 255, widths,
			descriptor, &indirectObjectRef{encoding})
	}
	toUnicodeData := toUnicodeCMapData(cpi.Map())
	toUnicodeStream := newStream(dw.nextSeq(), 0, toUnicodeData)
	if dw.compressToUnicode {
		if err := toUnicodeStream.compress(); err != nil {
			panic(err)
		}
	}
	dw.file.body.add(toUnicodeStream)
	font.setToUnicode(&indirectObjectRef{toUnicodeStream})

	dw.file.body.add(font)
	dw.resources.fonts[key] = &indirectObjectRef{font}
	return key
}

// fontKeyUnicode registers a Type0/CIDFontType2 composite font for the given
// TrueType font and returns the PDF resource key (e.g. "F0"). The /W and
// ToUnicode entries are left empty and filled in by flushUnicodeFonts at
// document close, once all glyph usages have been recorded.
func (dw *DocWriter) fontKeyUnicode(f *font.Font) string {
	psName := f.PostScriptName()
	cacheName := psName + "/unicode-Type0"
	if key, ok := dw.fontKeys[cacheName]; ok {
		return key
	}
	descriptor := newFontDescriptor(
		dw.nextSeq(), 0,
		psName, f.Family(),
		f.Flags(),
		f.BoundingBox(),
		0, // missingWidth
		f.StemV(),
		0, // stemH
		f.ItalicAngle(),
		f.CapHeight(),
		f.XHeight(),
		f.Ascent(),
		f.Descent(),
		f.Leading(),
		0, 0) // maxWidth, avgWidth
	dw.file.body.add(descriptor)

	key := fmt.Sprintf("F%d", len(dw.fontKeys))
	dw.fontKeys[cacheName] = key

	cid := newCIDFont(dw.nextSeq(), 0, psName, descriptor, 1000, array{})
	dw.file.body.add(cid)

	t0 := newType0Font(dw.nextSeq(), 0, psName, cid)
	dw.file.body.add(t0)

	dw.resources.fonts[key] = &indirectObjectRef{t0}
	dw.glyphRecorders[psName] = newGlyphRecorder()
	dw.unicodeFonts[psName] = f
	dw.cidFonts[psName] = cid
	dw.type0Fonts[psName] = t0
	dw.fontDescriptors[psName] = descriptor
	return key
}

// subsetTag generates a 6-character random uppercase tag for a font subset,
// per PDF spec §9.6.4: the tag is prefixed to the PostScript name as
// "ABCDEF+FontName" to signal an embedded subset.
func subsetTag() string {
	b := make([]byte, 6)
	rand.Read(b)
	for i := range b {
		b[i] = 'A' + b[i]%26
	}
	return string(b)
}

// flushUnicodeFonts is called from WriteTo before serialising the PDF.
// It fills in the /W width arrays, ToUnicode CMap streams, and embedded
// subset font streams for every Type0 composite font that was used during
// rendering.
func (dw *DocWriter) flushUnicodeFonts() {
	for psName, gr := range dw.glyphRecorders {
		mapping := gr.mapping()
		if len(mapping) == 0 {
			continue
		}
		f := dw.unicodeFonts[psName]
		upm := f.UnitsPerEm()

		// Build /W width array from recorded glyph IDs.
		glyphWidths := make(map[uint16]int, len(mapping))
		for gid := range mapping {
			w := f.AdvanceWidthForGlyph(gid)
			if upm > 0 {
				w = w * 1000 / upm
			}
			glyphWidths[gid] = w
		}
		defWidth := mostCommonWidth(glyphWidths)
		dw.cidFonts[psName].setDefaultWidth(defWidth)
		dw.cidFonts[psName].setWidths(buildCIDWidthArray(glyphWidths, defWidth))

		// Build ToUnicode CMap stream.
		tuData := toUnicodeCMapDataComposite(mapping)
		tuStream := newStream(dw.nextSeq(), 0, tuData)
		if dw.compressToUnicode {
			if err := tuStream.compress(); err != nil {
				panic(err)
			}
		}
		dw.file.body.add(tuStream)
		dw.type0Fonts[psName].setToUnicode(&indirectObjectRef{tuStream})

		// Embed a font subset as /FontFile2 in the descriptor.
		glyphIDs := make([]uint16, 0, len(mapping))
		for gid := range mapping {
			glyphIDs = append(glyphIDs, gid)
		}
		if subsetData, err := f.SubsetBytes(glyphIDs); err == nil {
			fontStream := newStream(dw.nextSeq(), 0, subsetData)
			fontStream.setLength1(len(subsetData))
			if dw.compressEmbeddedFonts {
				if err := fontStream.compress(); err != nil {
					panic(err)
				}
			}
			dw.file.body.add(fontStream)
			dw.fontDescriptors[psName].setFontFile2(&indirectObjectRef{fontStream})

			// Apply the 6-char subset tag to all three name occurrences:
			// FontDescriptor/FontName, CIDFont/BaseFont, Type0/BaseFont.
			taggedName := subsetTag() + "+" + psName
			dw.fontDescriptors[psName].dict["FontName"] = name(taggedName)
			dw.cidFonts[psName].setBaseFont(taggedName)
			dw.type0Fonts[psName].setBaseFont(taggedName)
		}
	}
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

func (dw *DocWriter) Line(x, y, angle, length float64) {
	dw.CurPage().Line(x, y, angle, length)
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

func (dw *DocWriter) PointsForCircle(x, y, r float64) []Location {
	return dw.CurPage().PointsForCircle(x, y, r)
}

func (dw *DocWriter) Circle(x, y, r float64, border, fill, reverse bool) error {
	return dw.CurPage().Circle(x, y, r, border, fill, reverse)
}

func (dw *DocWriter) PointsForEllipse(x, y, rx, ry float64) []Location {
	return dw.CurPage().PointsForEllipse(x, y, rx, ry)
}

func (dw *DocWriter) Ellipse(x, y, rx, ry float64, border, fill, reverse bool) error {
	return dw.CurPage().Ellipse(x, y, rx, ry, border, fill, reverse)
}

func (dw *DocWriter) PointsForArc(x, y, r, startAngle, endAngle float64) []Location {
	return dw.CurPage().PointsForArc(x, y, r, startAngle, endAngle)
}

func (dw *DocWriter) Arc(x, y, r, startAngle, endAngle float64, moveToStart bool) error {
	return dw.CurPage().Arc(x, y, r, startAngle, endAngle, moveToStart)
}

func (dw *DocWriter) Pie(x, y, r, startAngle, endAngle float64, border, fill, reverse bool) error {
	return dw.CurPage().Pie(x, y, r, startAngle, endAngle, border, fill, reverse)
}

func (dw *DocWriter) Arch(x, y, r1, r2, startAngle, endAngle float64, border, fill, reverse bool) error {
	return dw.CurPage().Arch(x, y, r1, r2, startAngle, endAngle, border, fill, reverse)
}

func (dw *DocWriter) PointsForPolygon(x, y, r float64, sides int, rotation float64) []Location {
	return dw.CurPage().PointsForPolygon(x, y, r, sides, rotation)
}

func (dw *DocWriter) Polygon(x, y, r float64, sides int, border, fill, reverse bool, rotation float64) error {
	return dw.CurPage().Polygon(x, y, r, sides, border, fill, reverse, rotation)
}

func (dw *DocWriter) Star(x, y, r1, r2 float64, points int, border, fill, reverse bool, rotation float64) error {
	return dw.CurPage().Star(x, y, r1, r2, points, border, fill, reverse, rotation)
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

func (dw *DocWriter) ImageDimensions(data []byte) (width, height int, err error) {
	return imageDimensions(data)
}

func (dw *DocWriter) ImageDimensionsFromFile(filename string) (width, height int, err error) {
	data, err := dw.readImageFile(filename)
	if err != nil {
		return 0, 0, err
	}
	return imageDimensions(data)
}

func (dw *DocWriter) Path(fn func()) error {
	return dw.CurPage().Path(fn)
}

func (dw *DocWriter) Fill() error {
	return dw.CurPage().Fill()
}

func (dw *DocWriter) Stroke() error {
	return dw.CurPage().Stroke()
}

func (dw *DocWriter) FillAndStroke() error {
	return dw.CurPage().FillAndStroke()
}

func (dw *DocWriter) Clip(fn func()) error {
	return dw.CurPage().Clip(fn)
}

func (dw *DocWriter) Rotate(angle, x, y float64, fn func()) error {
	return dw.CurPage().Rotate(angle, x, y, fn)
}

func (dw *DocWriter) Scale(x, y, scaleX, scaleY float64, fn func()) error {
	return dw.CurPage().Scale(x, y, scaleX, scaleY, fn)
}

func (dw *DocWriter) Rectangle(x, y, width, height float64, border bool, fill bool) {
	dw.CurPage().Rectangle(x, y, width, height, border, fill)
}

func (dw *DocWriter) Rectangle2(x, y, width, height float64, border bool, fill bool, corners []float64, path, reverse bool) {
	dw.CurPage().Rectangle2(x, y, width, height, border, fill, corners, path, reverse)
}

func (dw *DocWriter) PrintImage(data []byte, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return dw.CurPage().PrintImage(data, x, y, width, height)
}

func (dw *DocWriter) PrintImageFile(filename string, x, y float64, width, height *float64) (actualWidth, actualHeight float64, err error) {
	return dw.CurPage().PrintImageFile(filename, x, y, width, height)
}

func (dw *DocWriter) loadImage(data []byte, key string) (*pdfImage, string, error) {
	if cached, ok := dw.images[key]; ok {
		return cached.image, cached.name, nil
	}
	decoded, err := decodeImage(data)
	if err != nil {
		return nil, "", err
	}
	components := decoded.info.components
	if len(decoded.alphaData) > 0 && (components == imageComponentsGrayAlpha || components == imageComponentsRGBA) {
		components--
	}
	colorSpace, err := imageColorSpace(components)
	if err != nil {
		return nil, "", err
	}
	image := newPDFImage(dw.nextSeq(), 0, decoded.data)
	image.setDimensions(decoded.info.width, decoded.info.height)
	image.setBitsPerComponent(decoded.info.bitsPerComponent)
	image.setColorSpace(colorSpace)
	if decoded.filter == "FlateDecode" {
		if err := image.compress(); err != nil {
			return nil, "", err
		}
	} else if decoded.filter != "" {
		image.setFilter(decoded.filter)
	}
	if len(decoded.alphaData) > 0 {
		mask := newPDFImage(dw.nextSeq(), 0, decoded.alphaData)
		mask.setDimensions(decoded.info.width, decoded.info.height)
		mask.setBitsPerComponent(decoded.info.bitsPerComponent)
		mask.setColorSpace("DeviceGray")
		if err := mask.compress(); err != nil {
			return nil, "", err
		}
		dw.file.body.add(mask)
		image.setSMask(&indirectObjectRef{mask})
	}

	name := fmt.Sprintf("Im%d", len(dw.images))
	dw.file.body.add(image)
	dw.resources.setXObject(name, &indirectObjectRef{image})
	dw.images[key] = &cachedImage{image: image, name: name}
	return image, name, nil
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

func (dw *DocWriter) SetVTextAlign(vTextAlign string) (prev string) {
	return dw.CurPage().SetVTextAlign(vTextAlign)
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

func (dw *DocWriter) VTextAlign() string {
	return dw.CurPage().VTextAlign()
}

func (dw *DocWriter) Underline() bool {
	return dw.CurPage().Underline()
}

func (dw *DocWriter) widthsForFontCodepage(f *font.Font, cpi codepage.CodepageIndex) *indirectObject {
	var widths [256]int
	upm := f.UnitsPerEm()
	// Avoid divide by zero error for unusual fonts.
	if upm > 0 {
		for i, r := range cpi.Map() {
			designWidth, _ := f.AdvanceWidth(r)
			widths[i] = designWidth * 1000 / upm
		}
	}
	pdfWidths := arrayFromInts(widths[32:])
	ioWidths := &indirectObject{dw.nextSeq(), 0, &pdfWidths}
	return ioWidths
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
	dw.flushUnicodeFonts()
	dw.file.write(wr)
	return 0, nil
}

func (dw *DocWriter) X() float64 {
	return dw.CurPage().X()
}

func (dw *DocWriter) Y() float64 {
	return dw.CurPage().Y()
}
