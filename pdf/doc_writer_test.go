// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"image"
	"io"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/codepage"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/profile"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/ttf_fonts"
)

type limitedReadFS struct {
	name     string
	data     []byte
	maxChunk int
	limit    int
	reads    int
}

func (l *limitedReadFS) Open(name string) (fs.File, error) {
	if name != l.name {
		return nil, fs.ErrNotExist
	}
	return &limitedReadFile{fsys: l, reader: bytes.NewReader(l.data)}, nil
}

type limitedReadFile struct {
	fsys   *limitedReadFS
	reader *bytes.Reader
}

func (f *limitedReadFile) Stat() (fs.FileInfo, error) {
	return limitedFileInfo{name: f.fsys.name, size: int64(len(f.fsys.data))}, nil
}

func (f *limitedReadFile) Read(p []byte) (int, error) {
	remaining := f.fsys.limit - f.fsys.reads
	if remaining <= 0 {
		return 0, io.ErrUnexpectedEOF
	}
	if len(p) > f.fsys.maxChunk {
		p = p[:f.fsys.maxChunk]
	}
	if len(p) > remaining {
		p = p[:remaining]
	}
	n, err := f.reader.Read(p)
	f.fsys.reads += n
	return n, err
}

func (f *limitedReadFile) Close() error { return nil }

type limitedFileInfo struct {
	name string
	size int64
}

func (i limitedFileInfo) Name() string       { return i.name }
func (i limitedFileInfo) Size() int64        { return i.size }
func (i limitedFileInfo) Mode() fs.FileMode  { return 0o444 }
func (i limitedFileInfo) ModTime() time.Time { return time.Time{} }
func (i limitedFileInfo) IsDir() bool        { return false }
func (i limitedFileInfo) Sys() any           { return nil }

type countingFS struct {
	fsys  fs.FS
	opens int
}

type countingFontSource struct {
	base    font.FontSource
	selects int
}

func (s *countingFontSource) Select(family, weight, style string, ranges []string) (font.FontMetrics, error) {
	s.selects++
	return s.base.Select(family, weight, style, ranges)
}

func (s *countingFontSource) SubType() string {
	return s.base.SubType()
}

func (c *countingFS) Open(name string) (fs.File, error) {
	c.opens++
	return c.fsys.Open(name)
}

func TestNewDocWriter(t *testing.T) {
	dw := NewDocWriter()
	check(t, dw.nextSeq != nil, "DocWriter should have nextSeq func")
	check(t, dw.file != nil, "DocWriter should have file")
	check(t, dw.catalog != nil, "DocWriter should have catalog")
	check(t, len(dw.file.body.list) == 4, "DocWriter file body should be initialized")
	check(t, dw.file.trailer.dict["Root"] != nil, "DocWriter file trailer root should be set")
	check(t, dw.resources != nil, "DocWriter resources should be initialized")
	check(t, dw.PagesAcross() == 1, "DocWriter PagesAcross should default to 1")
	check(t, dw.PagesDown() == 1, "DocWriter PagesDown should default to 1")
	check(t, dw.PagesUp() == 1, "DocWriter PagesUp should default to 1")
	check(t, dw.pages == nil, "DocWriter pages should be nil")
	check(t, dw.curPage == nil, "DocWriter curPage should be nil")
	check(t, !dw.inPage(), "DocWriter should not be in page yet")
}

func TestDocWriter_ProfilerRecordsWritePhases(t *testing.T) {
	var buf bytes.Buffer
	profiler := profile.New()
	dw := NewDocWriter()
	dw.SetProfiler(profiler)
	if got := dw.Profiler(); got != profiler {
		t.Fatal("Profiler() did not return attached profiler")
	}
	dw.NewPage()
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	labels := map[string]bool{}
	for _, entry := range profiler.Entries() {
		labels[entry.Label] = true
	}
	for _, want := range []string{"pdf.write", "pdf.write.close_pages", "pdf.write.resolve_links", "pdf.write.flush_unicode_fonts", "pdf.write.file"} {
		if !labels[want] {
			t.Fatalf("profile labels = %#v, missing %q", labels, want)
		}
	}
}

func TestDocWriter_ProfilerDefaultsToNil(t *testing.T) {
	dw := NewDocWriter()
	if dw.Profiler() != nil {
		t.Fatal("new DocWriter profiler is not nil")
	}
}

func TestDocWriter_AddFontSource(t *testing.T) {
	dw := NewDocWriter()
	check(t, len(dw.fontSources) == 0, "No font sources should exist.")
	var fonts ttf_fonts.TtfFonts
	dw.AddFontSource(&fonts)
	check(t, dw.fontSources[0] == &fonts, "Font source should exist.")
}

func TestDocWriter_AddFontCachesSelections(t *testing.T) {
	afmFonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	source := &countingFontSource{base: afmFonts}
	dw := NewDocWriter()
	dw.AddFontSource(source)

	if _, err := dw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	if _, err := dw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}

	if source.selects != 1 {
		t.Fatalf("font selections = %d, want 1", source.selects)
	}
}

func TestDocWriter_AddFontCachesFailedSelections(t *testing.T) {
	afmFonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	source := &countingFontSource{base: afmFonts}
	dw := NewDocWriter()
	dw.AddFontSource(source)

	if _, err := dw.SetFont("Missing Font", 12, options.Options{}); err == nil {
		t.Fatal("first missing font selection succeeded")
	}
	if _, err := dw.SetFont("Missing Font", 12, options.Options{}); err == nil {
		t.Fatal("cached missing font selection succeeded")
	}

	if source.selects != 1 {
		t.Fatalf("font selections = %d, want 1 cached failure", source.selects)
	}
}

func TestDocWriter_AddFontSourceClearsSelectionCache(t *testing.T) {
	afmFonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	source := &countingFontSource{base: afmFonts}
	dw := NewDocWriter()
	dw.AddFontSource(source)

	if _, err := dw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(source)
	if _, err := dw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}

	if source.selects != 2 {
		t.Fatalf("font selections = %d, want 2 after adding a source", source.selects)
	}
}

func TestDocWriter_AddFontSourceClearsFailedSelectionCache(t *testing.T) {
	afmFonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	missingSource := &countingFontSource{base: &ttf_fonts.TtfFonts{}}
	dw := NewDocWriter()
	dw.AddFontSource(missingSource)

	if _, err := dw.SetFont("Helvetica", 12, options.Options{}); err == nil {
		t.Fatal("font selection unexpectedly succeeded before adding AFM source")
	}
	dw.AddFontSource(afmFonts)
	if _, err := dw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatalf("font selection after adding AFM source: %v", err)
	}

	if missingSource.selects != 2 {
		t.Fatalf("font selections = %d, want 2 after invalidating cached failure", missingSource.selects)
	}
}

func TestDocWriter_ShareFontSelectionCacheFrom(t *testing.T) {
	afmFonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	source := &countingFontSource{base: afmFonts}
	first := NewDocWriter()
	first.AddFontSource(source)
	second := NewDocWriter()
	second.AddFontSource(source)
	second.ShareFontSelectionCacheFrom(first)

	if _, err := first.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	if _, err := second.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}

	if source.selects != 1 {
		t.Fatalf("font selections = %d, want 1 shared cache miss", source.selects)
	}
}

func TestDocWriter_ShareFontSelectionCacheFromSharesFailures(t *testing.T) {
	afmFonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	source := &countingFontSource{base: afmFonts}
	first := NewDocWriter()
	first.AddFontSource(source)
	second := NewDocWriter()
	second.AddFontSource(source)
	second.ShareFontSelectionCacheFrom(first)

	if _, err := first.SetFont("Missing Font", 12, options.Options{}); err == nil {
		t.Fatal("first missing font selection succeeded")
	}
	if _, err := second.SetFont("Missing Font", 12, options.Options{}); err == nil {
		t.Fatal("shared cached missing font selection succeeded")
	}

	if source.selects != 1 {
		t.Fatalf("font selections = %d, want 1 shared cached failure", source.selects)
	}
}

func TestDocWriter_Close(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter()
	dw.NewPage()
	dw.WriteTo(&buf)
	check(t, !dw.inPage(), "DocWriter should not be in page anymore")

	dw2 := NewDocWriter()
	dw2.WriteTo(&buf)
	check(t, len(dw.pages) == 1, "DocWriter pages should have minimum of 1 page after Close")
}

func TestDocWriter_CompressPages(t *testing.T) {
	dw := NewDocWriter()
	dw.CompressPages(true)
	fonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fonts)
	pw := dw.NewPage()
	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	pw.MoveTo(72, 720)
	if err := pw.Print("Hello"); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdf := buf.String()
	if !strings.Contains(pdf, "/Filter /FlateDecode") {
		t.Fatalf("expected compressed page stream, got pdf excerpt:\n%s", pdf)
	}
}

func TestDocWriter_MeasureText(t *testing.T) {
	dw := NewDocWriter()
	fonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fonts)
	dw.SetUnits("cm")
	dw.NewPage()
	if _, err := dw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}

	metrics, err := dw.MeasureText("Hello")
	if err != nil {
		t.Fatal(err)
	}

	rt, err := rich_text.New("Hello", dw.Fonts(), dw.FontSize(), options.Options{
		"color": dw.FontColor(),
	})
	if err != nil {
		t.Fatal(err)
	}

	expectFdelta(t, UnitConversions["cm"].fromPts(rt.Width()), metrics.Width, 0.0001)
	expectFdelta(t, UnitConversions["cm"].fromPts(rt.Height()), metrics.Height, 0.0001)
	expectFdelta(t, UnitConversions["cm"].fromPts(rt.Ascent()), metrics.Ascent, 0.0001)
	expectFdelta(t, UnitConversions["cm"].fromPts(rt.Descent()), metrics.Descent, 0.0001)
}

func TestDocWriter_fontKey(t *testing.T) {
	skipIfNoTTFFonts(t)
	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}
	dw := NewDocWriter()
	dw.AddFontSource(fc)
	dw.NewPage()
	fonts, err := dw.AddFont("Arial", options.Options{})
	if err != nil {
		t.Fatal(err)
	}

	key1 := dw.fontKey(fonts[0], codepage.CodepageIndex(0))
	check(t, key1 == "F0", "1st fontKey should be F0.")
	key2 := dw.fontKey(fonts[0], codepage.CodepageIndex(0))
	check(t, key2 == "F0", "Same font and cpi should yield same key.")
	key3 := dw.fontKey(fonts[0], codepage.CodepageIndex(1))
	check(t, key3 == "F1", "2nd fontKey should be F1.")
}

func TestDocWriter_indexOfPage(t *testing.T) {
	dw := NewDocWriter()

	p1 := dw.NewPage()
	p2 := dw.NewPage()
	p3 := dw.NewPage()

	i1 := dw.indexOfPage(p1)
	check(t, i1 == 0, "1st PageWriter should have index 0")
	i2 := dw.indexOfPage(p2)
	check(t, i2 == 1, "2nd PageWriter should have index 1")
	i3 := dw.indexOfPage(p3)
	check(t, i3 == 2, "3rd PageWriter should have index 2")
}

func TestDocWriter_NewPage(t *testing.T) {
	dw := NewDocWriter()
	dw.NewPage()
	check(t, dw.curPage != nil, "DocWriter curPage should not be nil")
	check(t, dw.inPage(), "DocWriter should be in page now")
	check(t, len(dw.pages) == 1, "DocWriter should have 1 page now")
}

func TestDocWriter_NewPageAfter(t *testing.T) {
	dw := NewDocWriter()
	p1 := dw.NewPage()

	checkFatal(t, p1 != nil, "NewPageAfter should return a valid reference p1")
	checkFatal(t, len(dw.pages) == 1, "pages should have len 1")
	check(t, p1 == dw.pages[0], "p1 should be in 1st slot")

	check(t, p1.LineCapStyle() != ProjectingSquareCap, "LineCapStyle shouldn't default to ProjectingSquareCap")
	check(t, p1.LineColor() != colors.AliceBlue, "LineColor shouldn't default to AliceBlue")
	check(t, p1.LineDashPattern() != "dotted", "LineDashPattern shouldn't default to 'dotted'")
	check(t, p1.LineWidth("pt") != 42, "LineWidth shouldn't default to 42")

	p1.SetLineCapStyle(ProjectingSquareCap)
	p1.SetLineColor(colors.AliceBlue)
	p1.SetLineDashPattern("dotted")
	p1.SetLineWidth(42, "pt")

	p3 := dw.NewPageAfter(p1)

	checkFatal(t, p3 != nil, "NewPageAfter should return a valid reference p3")
	checkFatal(t, len(dw.pages) == 2, "pages should have len 2")
	check(t, p3 == dw.pages[1], "p3 should be in 2nd slot for now")

	check(t, p3.LineCapStyle() == ProjectingSquareCap, "LineCapStyle should still be ProjectingSquareCap")
	check(t, p3.LineColor() == colors.AliceBlue, "LineColor should still be AliceBlue")
	check(t, p3.LineDashPattern() == "dotted", "LineDashPattern should still be 'dotted'")
	check(t, p3.LineWidth("pt") == 42, "LineWidth should still be 42")

	p3.SetLineCapStyle(RoundCap)
	p3.SetLineColor(colors.AntiqueWhite)
	p3.SetLineDashPattern("dashed")
	p3.SetLineWidth(36, "pt")

	p2 := dw.NewPageAfter(p1)
	checkFatal(t, p2 != nil, "NewPageAfter should return a valid reference p2")
	checkFatal(t, len(dw.pages) == 3, "pages should have len 3")
	check(t, p2 == dw.pages[1], "p2 should now be in 2nd slot")
	check(t, p3 == dw.pages[2], "p3 should be in 3rd slot now")

	check(t, p2.LineCapStyle() == ProjectingSquareCap, "LineCapStyle should be ProjectingSquareCap")
	check(t, p2.LineColor() == colors.AliceBlue, "LineColor should be AliceBlue")
	check(t, p2.LineDashPattern() == "dotted", "LineDashPattern should be 'dotted'")
	check(t, p2.LineWidth("pt") == 42, "LineWidth should be 42")
}

func TestDocWriter_Open(t *testing.T) {
	dw := NewDocWriter()
	check(t, len(dw.options) == 0, "Default page options should be empty")
}

func TestDocWriter_PageHeight(t *testing.T) {
	dw := NewDocWriter()
	// height in (default) points
	expectF(t, 792, dw.PageHeight())
	dw.SetUnits("in")
	expectF(t, 11, dw.PageHeight())
	dw.SetUnits("cm")
	expectFdelta(t, 792*25.4/720, dw.PageHeight(), 1e-9)
	// dp = thousandths of an inch (792 / 0.072)
	dw.SetUnits("dp")
	expectF(t, 11000, dw.PageHeight())
}

func TestDocWriter_PageWidth(t *testing.T) {
	dw := NewDocWriter()
	// width in (default) points
	expectF(t, 612, dw.PageWidth())
	dw.SetUnits("in")
	expectF(t, 8.5, dw.PageWidth())
	dw.SetUnits("cm")
	expectFdelta(t, 612*25.4/720, dw.PageWidth(), 1e-9)
	dw.SetUnits("dp")
	expectF(t, 8500, dw.PageWidth())
}

func TestDocWriter_PrintImageFile_Integration(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.SetUnits("in")
	dw.NewPage()
	actualWidth, actualHeight, err := dw.PrintImageFile("testdata/testimg.jpg", 1, 1, floatPtr(2.0), nil)
	if err != nil {
		t.Fatal(err)
	}
	expectFdelta(t, 2.0, actualWidth, 0.0001)
	if actualHeight <= 0 {
		t.Fatalf("expected positive inferred height, got %f", actualHeight)
	}
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdf := buf.String()
	for _, fragment := range []string{
		"/Subtype /Image",
		"/Filter /DCTDecode",
		"/ColorSpace /DeviceRGB",
		"/XObject <<",
		"/Im0 ",
		" cm\n",
		"/Im0 Do\n",
	} {
		if !strings.Contains(pdf, fragment) {
			t.Fatalf("expected generated PDF to contain %q, got:\n%s", fragment, pdf)
		}
	}
}

func TestDocWriter_PrintImageFile_PNG_Integration(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.SetUnits("in")
	dw.NewPage()
	actualWidth, actualHeight, err := dw.PrintImageFile("testdata/eidetic.png", 1, 1, floatPtr(2.0), nil)
	if err != nil {
		t.Fatal(err)
	}
	expectFdelta(t, 2.0, actualWidth, 0.0001)
	if actualHeight <= 0 {
		t.Fatalf("expected positive inferred height, got %f", actualHeight)
	}
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdf := buf.String()
	for _, fragment := range []string{
		"/Subtype /Image",
		"/Filter /FlateDecode",
		"/ColorSpace /DeviceRGB",
		"/XObject <<",
		"/Im0 ",
		" cm\n",
		"/Im0 Do\n",
	} {
		if !strings.Contains(pdf, fragment) {
			t.Fatalf("expected generated PDF to contain %q, got:\n%s", fragment, pdf)
		}
	}
}

func TestDocWriter_PrintImageFile_UsesAssetFS(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	data := mustEncodePNG(t, img)

	dw := NewDocWriter()
	dw.SetAssetFS(fstest.MapFS{
		"logo.png": &fstest.MapFile{Data: data},
	})
	dw.NewPage()

	width, height, err := dw.PrintImageFile("logo.png", 1, 1, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if width != 3 || height != 2 {
		t.Fatalf("expected intrinsic size 3x2, got %.2fx%.2f", width, height)
	}
}

func TestDocWriter_ImageDimensionsFromFile_UsesAssetFS(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 4, 3))
	data := mustEncodePNG(t, img)

	dw := NewDocWriter()
	dw.SetAssetFS(fstest.MapFS{
		"nested/logo.png": &fstest.MapFile{Data: data},
	})

	width, height, err := dw.ImageDimensionsFromFile("nested/logo.png")
	if err != nil {
		t.Fatal(err)
	}
	if width != 4 || height != 3 {
		t.Fatalf("expected dimensions 4x3, got %dx%d", width, height)
	}
}

func TestDocWriter_ImageDimensionsFromFile_CachesAssetFSDimensions(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 4, 3))
	data := mustEncodePNG(t, img)
	assetFS := &countingFS{fsys: fstest.MapFS{
		"logo.png": &fstest.MapFile{Data: data},
	}}

	dw := NewDocWriter()
	dw.SetAssetFS(assetFS)

	for i := 0; i < 2; i++ {
		width, height, err := dw.ImageDimensionsFromFile("logo.png")
		if err != nil {
			t.Fatal(err)
		}
		if width != 4 || height != 3 {
			t.Fatalf("expected dimensions 4x3, got %dx%d", width, height)
		}
	}
	if assetFS.opens != 1 {
		t.Fatalf("opened asset %d times, want 1", assetFS.opens)
	}
}

func TestDocWriter_SVGDimensionsFromFile_CachesAssetFSDimensions(t *testing.T) {
	assetFS := &countingFS{fsys: fstest.MapFS{
		"icon.svg": &fstest.MapFile{Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="7" height="5"></svg>`)},
	}}

	dw := NewDocWriter()
	dw.SetAssetFS(assetFS)

	for i := 0; i < 2; i++ {
		width, height, err := dw.SVGDimensionsFromFile("icon.svg")
		if err != nil {
			t.Fatal(err)
		}
		if width != 7 || height != 5 {
			t.Fatalf("expected dimensions 7x5, got %dx%d", width, height)
		}
	}
	if assetFS.opens != 1 {
		t.Fatalf("opened asset %d times, want 1", assetFS.opens)
	}
}

func TestDocWriter_DimensionsFromFile_CachesErrors(t *testing.T) {
	assetFS := &countingFS{fsys: fstest.MapFS{}}

	dw := NewDocWriter()
	dw.SetAssetFS(assetFS)

	for i := 0; i < 2; i++ {
		width, height, err := dw.ImageDimensionsFromFile("missing.png")
		if err == nil {
			t.Fatal("expected error")
		}
		if width != 0 || height != 0 {
			t.Fatalf("expected zero dimensions for error, got %dx%d", width, height)
		}
	}
	if assetFS.opens != 1 {
		t.Fatalf("opened asset %d times, want 1", assetFS.opens)
	}
}

func TestDocWriter_SetAssetFS_ClearsDimensionCache(t *testing.T) {
	dw := NewDocWriter()
	dw.SetAssetFS(fstest.MapFS{
		"logo.png": &fstest.MapFile{Data: mustEncodePNG(t, image.NewGray(image.Rect(0, 0, 4, 3)))},
	})

	width, height, err := dw.ImageDimensionsFromFile("logo.png")
	if err != nil {
		t.Fatal(err)
	}
	if width != 4 || height != 3 {
		t.Fatalf("expected dimensions 4x3, got %dx%d", width, height)
	}

	dw.SetAssetFS(fstest.MapFS{
		"logo.png": &fstest.MapFile{Data: mustEncodePNG(t, image.NewGray(image.Rect(0, 0, 6, 2)))},
	})

	width, height, err = dw.ImageDimensionsFromFile("logo.png")
	if err != nil {
		t.Fatal(err)
	}
	if width != 6 || height != 2 {
		t.Fatalf("expected dimensions 6x2 after SetAssetFS, got %dx%d", width, height)
	}
}

func TestDocWriter_ImageDimensionsFromFile_PNGReadsOnlyMetadata(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 4, 3))
	data := append(mustEncodePNG(t, img), bytes.Repeat([]byte{0xAA}, 4096)...)
	assetFS := &limitedReadFS{name: "logo.png", data: data, maxChunk: 5, limit: 40}

	dw := NewDocWriter()
	dw.SetAssetFS(assetFS)

	width, height, err := dw.ImageDimensionsFromFile("logo.png")
	if err != nil {
		t.Fatal(err)
	}
	if width != 4 || height != 3 {
		t.Fatalf("expected dimensions 4x3, got %dx%d", width, height)
	}
	if assetFS.reads > 40 {
		t.Fatalf("read %d bytes, want metadata-only read", assetFS.reads)
	}
}

func TestDocWriter_ImageDimensionsFromFile_JPEGReadsOnlyMetadata(t *testing.T) {
	data := append([]byte{
		0xFF, 0xD8,
		0xFF, 0xE0, 0x00, 0x04, 0x00, 0x00,
		0xFF, 0xC0, 0x00, 0x11, 0x08, 0x00, 0x03, 0x00, 0x04, 0x03,
	}, bytes.Repeat([]byte{0xBB}, 4096)...)
	assetFS := &limitedReadFS{name: "photo.jpg", data: data, maxChunk: 3, limit: 24}

	dw := NewDocWriter()
	dw.SetAssetFS(assetFS)

	width, height, err := dw.ImageDimensionsFromFile("photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if width != 4 || height != 3 {
		t.Fatalf("expected dimensions 4x3, got %dx%d", width, height)
	}
	if assetFS.reads > 24 {
		t.Fatalf("read %d bytes, want metadata-only read", assetFS.reads)
	}
}

func TestDocWriter_ImageAssetPathValidation(t *testing.T) {
	dw := NewDocWriter()
	dw.SetAssetFS(fstest.MapFS{
		"logo.png": &fstest.MapFile{Data: mustEncodePNG(t, image.NewRGBA(image.Rect(0, 0, 1, 1)))},
		"dir":      &fstest.MapFile{Mode: fs.ModeDir},
	})

	bad := []string{"", ".", "./logo.png", "a/../logo.png", "/tmp/logo.png", "dir"}
	for _, name := range bad {
		t.Run(name, func(t *testing.T) {
			if _, _, err := dw.PrintImageFile(name, 1, 1, nil, nil); err == nil {
				t.Fatalf("expected error for %q, got nil", name)
			}
		})
	}
}

func TestDocWriter_SetFont_TrueType(t *testing.T) {
	skipIfNoTTFFonts(t)
	dw := NewDocWriter()

	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)

	dw.NewPage()

	st := SuperTest{t}
	st.True(dw.Fonts() == nil, "Document font list should be empty by default.")

	fonts, _ := dw.SetFont("Courier New", 10, options.Options{"weight": "Bold", "style": "Italic", "color": "AliceBlue"})
	st.Must(len(fonts) == 1, "length of fonts should be 1")
	st.Equal("Courier New", fonts[0].Family())
	st.True(10 == dw.FontSize())
	st.True(1.0 == fonts[0].RelativeSize)
	st.Equal("Bold", fonts[0].Weight)
	st.Equal("Italic", fonts[0].Style())
	st.True(fonts[0] == dw.Fonts()[0], "SetFont result should match new font list.")
	st.True(fonts[0].SubType() == "TrueType", "Font subType should be TrueType.")
}

func TestDocWriter_SetFont_Type1(t *testing.T) {
	dw := NewDocWriter()

	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)

	dw.NewPage()

	check(t, dw.Fonts() == nil, "Document font list should be empty by default.")

	fonts, _ := dw.SetFont("Courier", 10, options.Options{"weight": "Bold", "style": "Italic", "color": "AliceBlue"})
	checkFatal(t, len(fonts) == 1, "length of fonts should be 1")
	expectNS(t, "family", "Courier", fonts[0].Family())
	expectF(t, 10, dw.FontSize())
	expectF(t, 1.0, fonts[0].RelativeSize)
	expectNS(t, "weight", "Bold", fonts[0].Weight)
	expectNS(t, "style", "Italic", fonts[0].Style())
	check(t, fonts[0] == dw.Fonts()[0], "SetFont result should match new font list.")
	check(t, fonts[0].SubType() == "Type1", "Font subType should be Type1.")
}

func TestDocWriter_SetOptions(t *testing.T) {
	dw := NewDocWriter()
	dw.SetOptions(options.Options{"units": "in"})
	check(t, len(dw.options) == 1, "Default page options should have 1 option")
	check(t, dw.options["units"] == "in", "Default units should be in")
}

func TestDocWriter_SetSVGGradientStopOpacityMode(t *testing.T) {
	dw := NewDocWriter()
	if prev := dw.SetSVGGradientStopOpacityMode(SVGGradientStopOpacityModeCompatibility); prev != SVGGradientStopOpacityModeSoftMask {
		t.Fatalf("expected previous mode %q, got %q", SVGGradientStopOpacityModeSoftMask, prev)
	}
	if got := svgGradientStopOpacityMode(dw.options); got != SVGGradientStopOpacityModeCompatibility {
		t.Fatalf("expected doc mode %q, got %q", SVGGradientStopOpacityModeCompatibility, got)
	}
	pw := dw.NewPage()
	if got := svgGradientStopOpacityMode(pw.options); got != SVGGradientStopOpacityModeCompatibility {
		t.Fatalf("expected page to inherit mode %q, got %q", SVGGradientStopOpacityModeCompatibility, got)
	}
	if prev := dw.SetSVGGradientStopOpacityMode(SVGGradientStopOpacityModeSoftMask); prev != SVGGradientStopOpacityModeCompatibility {
		t.Fatalf("expected previous mode %q, got %q", SVGGradientStopOpacityModeCompatibility, prev)
	}
	if got := svgGradientStopOpacityMode(pw.options); got != SVGGradientStopOpacityModeSoftMask {
		t.Fatalf("expected current page mode %q after doc update, got %q", SVGGradientStopOpacityModeSoftMask, got)
	}
}

func TestParseSVGGradientStopOpacityMode(t *testing.T) {
	valid := []struct {
		value string
		want  SVGGradientStopOpacityMode
	}{
		{"soft-mask", SVGGradientStopOpacityModeSoftMask},
		{"compatibility", SVGGradientStopOpacityModeCompatibility},
	}
	for _, tt := range valid {
		got, ok := ParseSVGGradientStopOpacityMode(tt.value)
		if !ok || got != tt.want {
			t.Fatalf("ParseSVGGradientStopOpacityMode(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
	for _, value := range []string{"", "default", "softmask", "flat", "compatible", "Compatibility", " compatibility "} {
		if got, ok := ParseSVGGradientStopOpacityMode(value); ok {
			t.Fatalf("ParseSVGGradientStopOpacityMode(%q) = %q, want invalid", value, got)
		}
	}
}

func TestDocWriter_SetSVGBlendMode(t *testing.T) {
	dw := NewDocWriter()
	if prev := dw.SetSVGBlendMode(SVGBlendModeIgnore); prev != SVGBlendModeRespect {
		t.Fatalf("expected previous mode %q, got %q", SVGBlendModeRespect, prev)
	}
	if got := svgBlendMode(dw.options); got != SVGBlendModeIgnore {
		t.Fatalf("expected doc mode %q, got %q", SVGBlendModeIgnore, got)
	}
	pw := dw.NewPage()
	if got := svgBlendMode(pw.options); got != SVGBlendModeIgnore {
		t.Fatalf("expected page to inherit mode %q, got %q", SVGBlendModeIgnore, got)
	}
	if prev := dw.SetSVGBlendMode(SVGBlendModeRespect); prev != SVGBlendModeIgnore {
		t.Fatalf("expected previous mode %q, got %q", SVGBlendModeIgnore, prev)
	}
	if got := svgBlendMode(pw.options); got != SVGBlendModeRespect {
		t.Fatalf("expected current page mode %q after doc update, got %q", SVGBlendModeRespect, got)
	}
}

func TestDocWriter_NewPageWithOptionsOverridesSVGDefaults(t *testing.T) {
	dw := NewDocWriter()
	dw.SetSVGGradientStopOpacityMode(SVGGradientStopOpacityModeCompatibility)
	dw.SetSVGBlendMode(SVGBlendModeIgnore)

	pw := dw.NewPageWithOptions(options.Options{
		SVGGradientStopOpacityModeOption: SVGGradientStopOpacityModeSoftMask,
		SVGBlendModeOption:               SVGBlendModeRespect,
	})

	if got := svgGradientStopOpacityMode(pw.options); got != SVGGradientStopOpacityModeSoftMask {
		t.Fatalf("expected page stop opacity mode %q, got %q", SVGGradientStopOpacityModeSoftMask, got)
	}
	if got := svgBlendMode(pw.options); got != SVGBlendModeRespect {
		t.Fatalf("expected page blend mode %q, got %q", SVGBlendModeRespect, got)
	}
	if got := svgGradientStopOpacityMode(dw.options); got != SVGGradientStopOpacityModeCompatibility {
		t.Fatalf("expected doc stop opacity default %q, got %q", SVGGradientStopOpacityModeCompatibility, got)
	}
	if got := svgBlendMode(dw.options); got != SVGBlendModeIgnore {
		t.Fatalf("expected doc blend mode default %q, got %q", SVGBlendModeIgnore, got)
	}
}

func TestParseSVGBlendMode(t *testing.T) {
	valid := []struct {
		value string
		want  SVGBlendMode
	}{
		{"respect", SVGBlendModeRespect},
		{"ignore", SVGBlendModeIgnore},
	}
	for _, tt := range valid {
		got, ok := ParseSVGBlendMode(tt.value)
		if !ok || got != tt.want {
			t.Fatalf("ParseSVGBlendMode(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
	for _, value := range []string{"", "default", "honor", "keep", "drop", "off", "none", "compatibility", "Ignore", " ignore "} {
		if got, ok := ParseSVGBlendMode(value); ok {
			t.Fatalf("ParseSVGBlendMode(%q) = %q, want invalid", value, got)
		}
	}
}

func TestDocWriter_ClipText(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	if _, err := dw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	dw.MoveTo(72, 720)
	if err := dw.ClipText("Hello", func() {
		dw.Rectangle(0, 0, 18, 12, false, true)
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dw.CurPage().stream.String(), "7 Tr\n") {
		t.Fatalf("expected DocWriter ClipText to delegate to current page, got:\n%s", dw.CurPage().stream.String())
	}
}

func TestDocWriter_FillStrokeClipText(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	if _, err := dw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	dw.MoveTo(72, 720)
	if err := dw.FillStrokeClipText("Hello", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dw.CurPage().stream.String(), "6 Tr\n") {
		t.Fatalf("expected DocWriter FillStrokeClipText to delegate to current page, got:\n%s", dw.CurPage().stream.String())
	}
}

func TestDocWriter_PathAPI_Integration(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.Path(func() {
		dw.MoveTo(72, 720)
		dw.LineTo(144, 720)
		dw.LineTo(144, 648)
		dw.LineTo(72, 720)
		if err := dw.Clip(func() {
			dw.Rectangle(90, 666, 72, 36, true, false)
		}); err != nil {
			t.Fatalf("Clip returned error: %v", err)
		}
	})

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}

	pdf := buf.String()

	if !strings.HasPrefix(pdf, "%PDF-1.7\n") {
		t.Fatalf("expected PDF header, got:\n%s", pdf[:min(32, len(pdf))])
	}
	if !strings.Contains(pdf, "/Type /Catalog") {
		t.Fatalf("expected catalog object in generated PDF")
	}
	if !strings.Contains(pdf, "/Type /Page") {
		t.Fatalf("expected page object in generated PDF")
	}
	if !strings.Contains(pdf, "/Contents ") {
		t.Fatalf("expected page contents reference in generated PDF")
	}
	if !strings.Contains(pdf, "xref\n") {
		t.Fatalf("expected xref table in generated PDF")
	}
	if !strings.Contains(pdf, "startxref\n") {
		t.Fatalf("expected startxref in generated PDF")
	}
	if !strings.HasSuffix(pdf, "%%EOF\n") {
		t.Fatalf("expected PDF EOF marker")
	}

	// Public path API should produce path construction, clipping, scoped graphics
	// state, and drawing operators in the emitted content stream.
	for _, fragment := range []string{
		"72 72 m\n",
		"144 72 l\n",
		"144 144 l\n",
		"q\n",
		"W\nn\n",
		"90 90 72 36 re\n",
		"S\n",
		"Q\n",
	} {
		if !strings.Contains(pdf, fragment) {
			t.Fatalf("expected generated PDF to contain %q, got:\n%s", fragment, pdf)
		}
	}
}

func TestDocWriter_ShapesIntegration(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.SetUnits("in")
	dw.NewPage()
	if err := dw.Circle(1.5, 1.5, 0.5, true, false, false); err != nil {
		t.Fatalf("Circle returned error: %v", err)
	}
	if err := dw.Ellipse(3.0, 1.5, 0.75, 0.5, true, false, false); err != nil {
		t.Fatalf("Ellipse returned error: %v", err)
	}
	if err := dw.Pie(1.5, 3.5, 0.75, 0, 135, true, true, false); err != nil {
		t.Fatalf("Pie returned error: %v", err)
	}
	if err := dw.Arch(1.5, 5.0, 0.35, 0.7, 220, 20, true, true, false); err != nil {
		t.Fatalf("Arch returned error: %v", err)
	}
	if err := dw.Polygon(3.5, 3.5, 0.7, 6, true, false, false, 0); err != nil {
		t.Fatalf("Polygon returned error: %v", err)
	}
	if err := dw.Star(5.5, 3.5, 0.8, 0.35, 5, true, false, false, 0); err != nil {
		t.Fatalf("Star returned error: %v", err)
	}
	if err := dw.Circle(1.5, 6.0, 0.5, true, true, false); err != nil {
		t.Fatalf("filled Circle returned error: %v", err)
	}
	if err := dw.Ellipse(3.0, 6.0, 0.75, 0.5, true, true, false); err != nil {
		t.Fatalf("filled Ellipse returned error: %v", err)
	}
	if err := dw.Polygon(4.8, 5.5, 0.6, 5, true, true, false, 0); err != nil {
		t.Fatalf("filled Polygon returned error: %v", err)
	}
	if err := dw.Star(5.8, 5.7, 0.55, 0.25, 5, true, true, false, 0); err != nil {
		t.Fatalf("filled Star returned error: %v", err)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	pdf := buf.String()

	if !strings.Contains(pdf, "%PDF-1.7\n") {
		t.Fatalf("expected PDF header")
	}
	if !strings.Contains(pdf, " c\n") {
		t.Fatalf("expected curve operators from ellipse/circle/arc-based shapes, got:\n%s", pdf)
	}
	if !strings.Contains(pdf, " l\n") {
		t.Fatalf("expected line operators from polygonal shapes, got:\n%s", pdf)
	}
	if !strings.Contains(pdf, "B\n") {
		t.Fatalf("expected fill-and-stroke operators in generated PDF, got:\n%s", pdf)
	}
	if !strings.Contains(pdf, "S\n") {
		t.Fatalf("expected stroke operators in generated PDF, got:\n%s", pdf)
	}
}

func TestDocWriter_TransformsIntegration(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.SetUnits("in")
	dw.NewPage()
	if err := dw.Rotate(30, 2, 2, func() {
		dw.Rectangle(1.5, 1.7, 1.0, 0.5, true, false)
	}); err != nil {
		t.Fatalf("Rotate returned error: %v", err)
	}
	if err := dw.Scale(4.5, 2.5, 1.5, 0.75, func() {
		dw.Line(4.0, 2.5, 0, 1.0)
	}); err != nil {
		t.Fatalf("Scale returned error: %v", err)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	pdf := buf.String()

	for _, fragment := range []string{
		"%PDF-1.7\n",
		"q\n",
		" cm\n",
		" re\n",
		"S\n",
		"Q\n",
	} {
		if !strings.Contains(pdf, fragment) {
			t.Fatalf("expected generated PDF to contain %q, got:\n%s", fragment, pdf)
		}
	}
}

func TestDocWriter_VTextAlignIntegration(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	dw.NewPage()
	fonts, err := dw.SetFont("Courier", 12, options.Options{})
	if err != nil {
		t.Fatal(err)
	}

	dw.SetVTextAlign("base")
	if err := dw.Print("Base\n"); err != nil {
		t.Fatal(err)
	}
	dw.SetVTextAlign("top")
	if err := dw.Print("Top"); err != nil {
		t.Fatal(err)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}
	pdf := buf.String()
	scale := 12 * 0.001
	if upm := fonts[0].UnitsPerEm(); upm > 0 {
		scale = 12 / float64(upm)
	}
	topRise := -float64(fonts[0].CapHeight()) * scale
	if topRise == 0 {
		topRise = -float64(fonts[0].Ascent()) * scale
	}

	for _, fragment := range []string{
		"%PDF-1.7\n",
		"0 Ts\n",
		g(topRise) + " Ts\n",
		"(Base) Tj\n",
		"(Top) Tj\n",
	} {
		if !strings.Contains(pdf, fragment) {
			t.Fatalf("expected generated PDF to contain %q, got:\n%s", fragment, pdf)
		}
	}
}

// TODO: TestPagesAcross
// TODO: TestPagesDown
// TODO: TestPagesUp
