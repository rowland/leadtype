// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
)

func floatPtr(v float64) *float64 {
	return &v
}

func mustReadTestImage(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/testimg.jpg")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mustReadTestPNG(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/eidetic.png")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func goldenPath(filename string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "golden", filename)
}

func mustEncodePNG(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func testSVGFixture() []byte {
	return []byte(`
<svg width="120" height="80" viewBox="0 0 120 80" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <clipPath id="clip">
      <rect x="6" y="6" width="108" height="68" rx="8" ry="8"/>
    </clipPath>
  </defs>
  <rect x="0" y="0" width="120" height="80" fill="#eef5ff"/>
  <g clip-path="url(#clip)">
    <path d="M 10 50 A 18 18 0 0 1 46 50" stroke="#cc3300" stroke-width="3" fill="none"/>
    <circle cx="28" cy="28" r="16" fill="#66aaff" stroke="#003366" stroke-width="2"/>
    <polygon points="60,12 108,12 92,34 108,56 60,56" fill="#88cc66" stroke="#225522" stroke-width="2"/>
    <path d="M 62 60 C 72 30, 98 30, 108 60" stroke="#7a32a8" stroke-width="3" fill="none"/>
  </g>
  <text x="60" y="70" text-anchor="middle" font-family="Helvetica" font-size="12" fill="#111111">SVG demo</text>
</svg>`)
}

func testSVGGradientOpacityFixture() []byte {
	return []byte(`
<svg width="120" height="40" viewBox="0 0 120 40" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <style>.band{opacity:0.8;fill:url(#grad);}</style>
    <linearGradient id="grad" x1="0" y1="0" x2="120" y2="0" gradientUnits="userSpaceOnUse">
      <stop offset="0" stop-color="#427b9a"/>
      <stop offset="1" stop-color="#f15c4e" stop-opacity="0.4"/>
    </linearGradient>
  </defs>
  <rect class="band" width="120" height="40"/>
</svg>`)
}

func testSVGUseMaskBlendFixture() []byte {
	return []byte(`
<svg width="120" height="80" viewBox="0 0 120 80" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
  <defs>
    <rect id="shape" x="10" y="10" width="100" height="60"/>
    <clipPath id="clip">
      <use xlink:href="#shape" />
    </clipPath>
    <linearGradient id="mask-grad" x1="10" y1="10" x2="110" y2="10" gradientUnits="userSpaceOnUse">
      <stop offset="0" stop-color="#ffffff"/>
      <stop offset="1" stop-color="#000000"/>
    </linearGradient>
    <mask id="fade" maskUnits="userSpaceOnUse">
      <g filter="url(#luminosity-noclip)">
        <rect x="10" y="10" width="100" height="60" fill="url(#mask-grad)"/>
      </g>
    </mask>
    <filter id="luminosity-noclip" x="0" y="0" width="120" height="80" filterUnits="userSpaceOnUse">
      <feFlood flood-color="#fff" result="bg"/>
      <feBlend in="SourceGraphic" in2="bg"/>
    </filter>
    <linearGradient id="fill-grad" x1="10" y1="10" x2="110" y2="70" gradientUnits="userSpaceOnUse">
      <stop offset="0" stop-color="#5290ae"/>
      <stop offset="1" stop-color="#cbcbcb"/>
    </linearGradient>
  </defs>
  <path clip-path="url(#clip)" mask="url(#fade)" style="mix-blend-mode:hard-light;fill:url(#fill-grad)" d="M10 10 H110 V70 H10 Z"/>
</svg>`)
}

func testSVGClippedUseGradientFixture() []byte {
	return []byte(`
<svg width="240" height="160" viewBox="0 0 240 160" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
  <defs>
    <rect id="hero-window" x="18" y="20" width="204" height="92" rx="24" />
    <path id="spark" d="M0 -12 L4 -4 L13 -2 L6 3 L8 12 L0 7 L-8 12 L-6 3 L-13 -2 L-4 -4 Z" />
    <clipPath id="hero-clip">
      <use xlink:href="#hero-window" />
    </clipPath>
    <linearGradient id="hero-grad" x1="10" y1="28" x2="230" y2="110" gradientUnits="userSpaceOnUse" gradientTransform="rotate(-12 120 70)">
      <stop offset="0" stop-color="#0b5d81" />
      <stop offset="0.55" stop-color="#56b8c7" stop-opacity="0.95" />
      <stop offset="1" stop-color="#f38b56" stop-opacity="0.22" />
    </linearGradient>
  </defs>
  <rect width="240" height="160" fill="#f7f2e7" />
  <g clip-path="url(#hero-clip)">
    <rect x="18" y="20" width="204" height="92" fill="url(#hero-grad)" />
    <use xlink:href="#spark" x="58" y="60" transform="scale(1.75)" fill="#f59d7e" stroke="#fff4e6" stroke-width="1.3" opacity="0.92" />
  </g>
</svg>`)
}

func testSVGTextGradientFixture() []byte {
	return []byte(`
<svg width="160" height="48" viewBox="0 0 160 48" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="title-grad" x1="0" y1="0" x2="160" y2="0" gradientUnits="userSpaceOnUse">
      <stop offset="0" stop-color="#1f6f8b"/>
      <stop offset="1" stop-color="#f39a5b"/>
    </linearGradient>
  </defs>
  <text x="16" y="30" font-family="Helvetica" font-size="20" fill="url(#title-grad)" opacity="0.8">LeadType</text>
</svg>`)
}

func testSVGTextGradientStrokeFixture() []byte {
	return []byte(`
<svg width="160" height="48" viewBox="0 0 160 48" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="stroke-grad" x1="0" y1="0" x2="160" y2="0" gradientUnits="userSpaceOnUse">
      <stop offset="0" stop-color="#1f6f8b"/>
      <stop offset="1" stop-color="#f39a5b"/>
    </linearGradient>
  </defs>
  <text x="16" y="30" font-family="Helvetica" font-size="20" fill="#222" stroke="url(#stroke-grad)">LeadType</text>
</svg>`)
}

func mutatePNGInterlace(t *testing.T, data []byte, interlace byte) []byte {
	t.Helper()
	if !isPNG(data) {
		t.Fatal("expected PNG data")
	}
	mutated := append([]byte(nil), data...)
	mutated[28] = interlace
	return mutated
}

func TestIsJPEG(t *testing.T) {
	if !isJPEG(mustReadTestImage(t)) {
		t.Fatal("expected fixture to be detected as JPEG")
	}
	if isJPEG([]byte("not a jpeg")) {
		t.Fatal("expected non-JPEG bytes to be rejected")
	}
}

func TestIsPNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	data := mustEncodePNG(t, img)
	if !isPNG(data) {
		t.Fatal("expected fixture to be detected as PNG")
	}
	if isPNG([]byte("not a png")) {
		t.Fatal("expected non-PNG bytes to be rejected")
	}
}

func TestJPEGInfo(t *testing.T) {
	data := mustReadTestImage(t)
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	info, err := jpegInfo(data)
	if err != nil {
		t.Fatal(err)
	}
	if info.width != cfg.Width {
		t.Fatalf("expected width %d, got %d", cfg.Width, info.width)
	}
	if info.height != cfg.Height {
		t.Fatalf("expected height %d, got %d", cfg.Height, info.height)
	}
	if info.components != 3 {
		t.Fatalf("expected 3 components, got %d", info.components)
	}
	if info.bitsPerComponent != 8 {
		t.Fatalf("expected 8 bits per component, got %d", info.bitsPerComponent)
	}
}

func TestJPEGInfo_NotJPEG(t *testing.T) {
	if _, err := jpegInfo([]byte("bogus")); err != errNotJPEG {
		t.Fatalf("expected errNotJPEG, got %v", err)
	}
}

func TestPNGInfo(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	data := mustEncodePNG(t, img)
	info, err := pngInfo(data)
	if err != nil {
		t.Fatal(err)
	}
	if info.width != 3 || info.height != 2 {
		t.Fatalf("expected dimensions 3x2, got %dx%d", info.width, info.height)
	}
	if info.components != 4 {
		t.Fatalf("expected 4 components for RGBA PNG, got %d", info.components)
	}
	if info.bitsPerComponent != 8 {
		t.Fatalf("expected 8 bits per component, got %d", info.bitsPerComponent)
	}
}

func TestPNGInfo_UnsupportedInterlace(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	data := mutatePNGInterlace(t, mustEncodePNG(t, img), 1)
	if _, err := pngInfo(data); err != errUnsupportedPNG {
		t.Fatalf("expected errUnsupportedPNG, got %v", err)
	}
}

func TestDecodePNG_RGB(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xFF})
	img.Set(1, 0, color.RGBA{R: 0x40, G: 0x50, B: 0x60, A: 0xFF})
	data := mustEncodePNG(t, img)

	decoded, err := decodePNG(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.filter != "FlateDecode" {
		t.Fatalf("expected FlateDecode filter, got %q", decoded.filter)
	}
	if decoded.info.components != 3 && decoded.info.components != 4 {
		t.Fatalf("expected RGB or RGBA PNG info, got %d components", decoded.info.components)
	}
	if len(decoded.alphaData) != 0 {
		t.Fatalf("expected opaque PNG to omit alpha mask, got %d alpha bytes", len(decoded.alphaData))
	}
	want := []byte{0x10, 0x20, 0x30, 0x40, 0x50, 0x60}
	if !bytes.Equal(decoded.data, want) {
		t.Fatalf("expected RGB data %v, got %v", want, decoded.data)
	}
}

func TestDecodePNG_GrayAlpha(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.NRGBA{R: 0x22, G: 0x22, B: 0x22, A: 0x80})
	img.Set(1, 0, color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0x40})
	data := mustEncodePNG(t, img)

	decoded, err := decodePNG(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded.data) != 6 {
		t.Fatalf("expected RGB image data for NRGBA PNG, got %d bytes", len(decoded.data))
	}
	if len(decoded.alphaData) != 2 {
		t.Fatalf("expected 2 alpha bytes, got %d", len(decoded.alphaData))
	}
	if decoded.alphaData[0] != 0x80 || decoded.alphaData[1] != 0x40 {
		t.Fatalf("unexpected alpha mask %v", decoded.alphaData)
	}
}

func TestDecodePNG_Gray(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 2, 1))
	img.SetGray(0, 0, color.Gray{Y: 0x11})
	img.SetGray(1, 0, color.Gray{Y: 0x99})
	data := mustEncodePNG(t, img)

	decoded, err := decodePNG(data)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x11, 0x99}
	if !bytes.Equal(decoded.data, want) {
		t.Fatalf("expected gray data %v, got %v", want, decoded.data)
	}
	if len(decoded.alphaData) != 0 {
		t.Fatalf("expected no alpha data, got %d bytes", len(decoded.alphaData))
	}
}

func TestPageWriter_PrintImage_PNG_DefaultSize(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 2))
	data := mustEncodePNG(t, img)

	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})
	width, height, err := pw.PrintImage(data, 10, 20, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if width != 3 || height != 2 {
		t.Fatalf("expected intrinsic size 3x2, got %.2fx%.2f", width, height)
	}
	if !strings.Contains(pw.stream.String(), "/Im0 Do\n") {
		t.Fatalf("expected image draw operator, got:\n%s", pw.stream.String())
	}
}

func TestDocWriter_PrintImage_PNG_RGBA_Integration(t *testing.T) {
	var buf bytes.Buffer

	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.NRGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0x80})
	img.Set(1, 0, color.NRGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0xFF})
	data := mustEncodePNG(t, img)

	dw := NewDocWriter()
	dw.SetUnits("in")
	dw.NewPage()
	if _, _, err := dw.PrintImage(data, 1, 1, nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdf := buf.String()
	for _, fragment := range []string{
		"/Subtype /Image",
		"/Filter /FlateDecode",
		"/ColorSpace /DeviceRGB",
		"/ColorSpace /DeviceGray",
		"/SMask ",
		"/XObject <<",
		"/Im0 ",
	} {
		if !strings.Contains(pdf, fragment) {
			t.Fatalf("expected generated PDF to contain %q, got:\n%s", fragment, pdf)
		}
	}
}

func TestDocWriter_PrintImage_PNG_Golden(t *testing.T) {
	var buf bytes.Buffer

	rgba := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	rgba.Set(0, 0, color.NRGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0x80})
	rgba.Set(1, 0, color.NRGBA{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF})
	rgba.Set(0, 1, color.NRGBA{R: 0x00, G: 0x00, B: 0xFF, A: 0x40})
	rgba.Set(1, 1, color.NRGBA{R: 0xFF, G: 0xFF, B: 0x00, A: 0xC0})

	gray := image.NewGray(image.Rect(0, 0, 2, 2))
	gray.SetGray(0, 0, color.Gray{Y: 0x11})
	gray.SetGray(1, 0, color.Gray{Y: 0x44})
	gray.SetGray(0, 1, color.Gray{Y: 0x88})
	gray.SetGray(1, 1, color.Gray{Y: 0xCC})

	dw := NewDocWriter()
	dw.SetUnits("in")
	dw.NewPage()
	if _, _, err := dw.PrintImage(mustEncodePNG(t, rgba), 1, 1, floatPtr(1.0), nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := dw.PrintImage(mustEncodePNG(t, gray), 2.5, 1, floatPtr(1.0), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	compareGolden(t, buf.Bytes(), goldenPath("png_images.pdf"))
}

func TestImageDimensions_PNG(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 4, 3))
	data := mustEncodePNG(t, img)
	width, height, err := imageDimensions(data)
	if err != nil {
		t.Fatal(err)
	}
	if width != 4 || height != 3 {
		t.Fatalf("expected dimensions 4x3, got %dx%d", width, height)
	}
}

func TestImageDimensions_PNG_Fixture(t *testing.T) {
	width, height, err := imageDimensions(mustReadTestPNG(t))
	if err != nil {
		t.Fatal(err)
	}
	if width != 226 || height != 79 {
		t.Fatalf("expected dimensions 226x79, got %dx%d", width, height)
	}
}

func TestSVGDimensions(t *testing.T) {
	width, height, err := svgDimensions(testSVGFixture())
	if err != nil {
		t.Fatal(err)
	}
	if width != 120 || height != 80 {
		t.Fatalf("expected dimensions 120x80, got %dx%d", width, height)
	}
}

func TestPageWriter_PrintImage_SVG(t *testing.T) {
	dw := NewDocWriter()
	afm, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(afm)
	pw := newPageWriter(dw, options.Options{"units": "in"})
	width := 2.0
	actualWidth, actualHeight, err := pw.PrintImage(testSVGFixture(), 1, 1, &width, nil)
	if err != nil {
		t.Fatal(err)
	}
	if actualWidth != 2.0 {
		t.Fatalf("actualWidth = %.2f, want 2.0", actualWidth)
	}
	if actualHeight <= 0 {
		t.Fatalf("actualHeight = %.2f, want > 0", actualHeight)
	}
	pw.close()

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if strings.Count(got, "/Subtype /Form") != 1 {
		t.Fatalf("expected one SVG form XObject, got:\n%s", got)
	}
	if !strings.Contains(got, "/Fm0 Do") {
		t.Fatalf("expected page content to place cached SVG form, got:\n%s", got)
	}
	if strings.Contains(got, "/Im0 Do") {
		t.Fatalf("expected SVG rendering to avoid image XObjects, got:\n%s", got)
	}
	for _, fragment := range []string{" m\n", " c\n", "W\n", "BT\n", "Tj\n"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected SVG form stream to contain %q, got:\n%s", fragment, got)
		}
	}
}

func TestDocWriter_PrintSVG_Golden(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter()
	afm, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(afm)
	dw.SetUnits("in")
	dw.NewPage()
	width := 3.0
	if _, _, err := dw.PrintSVG(testSVGFixture(), 1, 1, &width, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	compareGolden(t, buf.Bytes(), goldenPath("svg_image.pdf"))
}

func TestPageWriter_PrintSVG_RestoresFontState(t *testing.T) {
	dw := NewDocWriter()
	afm, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(afm)
	pw := newPageWriter(dw, options.Options{"units": "in"})
	fonts, err := pw.SetFont("Helvetica", 10, options.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(fonts) == 0 {
		t.Fatal("expected initial font")
	}
	savedFonts := append([]*font.Font(nil), pw.Fonts()...)
	savedSize := pw.FontSize()
	savedColor := pw.FontColor()

	data := []byte(`<svg width="80" height="20" xmlns="http://www.w3.org/2000/svg"><text x="10" y="12" font-family="Helvetica" font-size="4" fill="#333333">tiny</text></svg>`)
	width := 1.5
	if _, _, err := pw.PrintSVG(data, 1, 1, &width, nil); err != nil {
		t.Fatal(err)
	}

	if got := pw.FontSize(); got != savedSize {
		t.Fatalf("font size after SVG = %v, want %v", got, savedSize)
	}
	if got := pw.FontColor(); got != savedColor {
		t.Fatalf("font color after SVG = %v, want %v", got, savedColor)
	}
	if len(pw.Fonts()) != len(savedFonts) {
		t.Fatalf("font count after SVG = %d, want %d", len(pw.Fonts()), len(savedFonts))
	}
	if len(savedFonts) > 0 && pw.Fonts()[0].PostScriptName() != savedFonts[0].PostScriptName() {
		t.Fatalf("font after SVG = %s, want %s", pw.Fonts()[0].PostScriptName(), savedFonts[0].PostScriptName())
	}
}

func TestDocWriter_PrintSVG_ReusesFormAcrossPages(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter()
	afm, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(afm)
	dw.SetUnits("in")

	width1 := 3.0
	if _, _, err := dw.PrintSVG(testSVGFixture(), 1, 1, &width1, nil); err != nil {
		t.Fatal(err)
	}
	dw.NewPage()
	height2 := 1.25
	if _, _, err := dw.PrintSVG(testSVGFixture(), 0.5, 0.75, nil, &height2); err != nil {
		t.Fatal(err)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if count := strings.Count(got, "/Subtype /Form"); count != 1 {
		t.Fatalf("expected one cached SVG form, got %d\n%s", count, got)
	}
	if count := strings.Count(got, "/Fm0 Do"); count != 2 {
		t.Fatalf("expected two form placements, got %d\n%s", count, got)
	}
}

func TestDocWriter_PrintSVG_CacheSeparatesStopOpacityModes(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter()
	width := 120.0
	if _, _, err := dw.PrintSVG(testSVGGradientOpacityFixture(), 0, 0, &width, nil); err != nil {
		t.Fatal(err)
	}
	dw.SetSVGGradientStopOpacityMode("compatibility")
	dw.NewPage()
	if _, _, err := dw.PrintSVG(testSVGGradientOpacityFixture(), 0, 0, &width, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if count := strings.Count(got, "/Subtype /Form"); count != 3 {
		t.Fatalf("expected two cached SVG forms plus one soft-mask helper form, got %d\n%s", count, got)
	}
	for _, fragment := range []string{"/Fm0 Do", "/Fm1 Do"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected output to contain %q, got:\n%s", fragment, got)
		}
	}
}

func TestPageWriter_PrintSVG_GradientOpacityUsesSoftMask(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{"units": "pt"})
	width := 120.0
	if _, _, err := pw.PrintSVG(testSVGGradientOpacityFixture(), 0, 0, &width, nil); err != nil {
		t.Fatal(err)
	}
	pw.close()

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, fragment := range []string{"/ExtGState", "/SMask", "/Luminosity", "/Sh"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected SVG gradient opacity output to contain %q, got:\n%s", fragment, got)
		}
	}
	for _, fragment := range []string{"/BC [0 0 0", "/I true", "/CS /DeviceRGB"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected SVG gradient opacity soft mask output to contain %q, got:\n%s", fragment, got)
		}
	}
	for _, fragment := range []string{"/Pattern", " scn"} {
		if strings.Contains(got, fragment) {
			t.Fatalf("expected varying-alpha SVG fill to avoid pattern fill, got unexpected %q in:\n%s", fragment, got)
		}
	}
	if strings.Contains(got, "\nET\n") {
		t.Fatalf("expected SVG-only output to avoid stray ET operators, got:\n%s", got)
	}
}

func TestPageWriter_PrintSVG_GradientOpacityCompatibilityModeUsesFlatAlpha(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{"units": "pt"})
	if prev := pw.SetSVGGradientStopOpacityMode("compatibility"); prev != svgGradientStopOpacityModeSoftMask {
		t.Fatalf("expected default mode %q, got %q", svgGradientStopOpacityModeSoftMask, prev)
	}
	width := 120.0
	if _, _, err := pw.PrintSVG(testSVGGradientOpacityFixture(), 0, 0, &width, nil); err != nil {
		t.Fatal(err)
	}
	pw.close()

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "/ca 0.56") {
		t.Fatalf("expected compatibility mode to collapse stop-opacity to flat alpha 0.56, got:\n%s", got)
	}
	for _, fragment := range []string{"/SMask", "/Luminosity"} {
		if strings.Contains(got, fragment) {
			t.Fatalf("expected compatibility mode to avoid %q, got:\n%s", fragment, got)
		}
	}
}

func TestPageWriter_PrintSVG_UseMaskAndBlendMode(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{"units": "pt"})
	width := 120.0
	if _, _, err := pw.PrintSVG(testSVGUseMaskBlendFixture(), 0, 0, &width, nil); err != nil {
		t.Fatal(err)
	}
	pw.close()

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, fragment := range []string{"/HardLight", "/SMask", "/Luminosity", "W\n", "/Fm0 Do"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected SVG use/mask/blend output to contain %q, got:\n%s", fragment, got)
		}
	}
}

func TestPageWriter_PrintSVG_WarnsForIgnoredFilterRef(t *testing.T) {
	msg := captureStderr(t, func() {
		dw := NewDocWriter()
		pw := newPageWriter(dw, options.Options{"units": "pt"})
		width := 120.0
		if _, _, err := pw.PrintSVG(testSVGUseMaskBlendFixture(), 0, 0, &width, nil); err != nil {
			t.Fatal(err)
		}
		pw.close()
	})
	if !strings.Contains(msg, "svg: <g> filter: filter #luminosity-noclip is parsed but not yet rendered") {
		t.Fatalf("expected ignored filter warning, got %q", msg)
	}
}

func TestPageWriter_PrintSVG_UseDoesNotReclipReferencedNode(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{"units": "pt"})
	width := 240.0
	if _, _, err := pw.PrintSVG(testSVGClippedUseGradientFixture(), 0, 0, &width, nil); err != nil {
		t.Fatal(err)
	}
	pw.close()

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if count := strings.Count(got, "q\nW\nn\n"); count != 3 {
		t.Fatalf("expected one alpha-mask clip, one group clip, and one direct-shading clip, got %d clip scopes\n%s", count, got)
	}
	if strings.Contains(got, "73.5 125 m\n346.5 125 l\n") {
		t.Fatalf("expected <use> rendering to avoid replaying a transformed parent clip, got:\n%s", got)
	}
}

func TestPageWriter_PrintSVG_TextGradientUsesTextClip(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := newPageWriter(dw, options.Options{"units": "pt"})
	width := 160.0
	if _, _, err := pw.PrintSVG(testSVGTextGradientFixture(), 0, 0, &width, nil); err != nil {
		t.Fatal(err)
	}
	pw.close()

	var buf bytes.Buffer
	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, fragment := range []string{"7 Tr\n", "/Sh", "sh\n", "/ca 0.8"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected SVG text gradient output to contain %q, got:\n%s", fragment, got)
		}
	}
}

func TestPageWriter_PrintSVG_TextGradientStrokeWarns(t *testing.T) {
	msg := captureStderr(t, func() {
		dw := NewDocWriter()
		fc, err := afm_fonts.Default()
		if err != nil {
			t.Fatal(err)
		}
		dw.AddFontSource(fc)
		pw := newPageWriter(dw, options.Options{"units": "pt"})
		width := 160.0
		if _, _, err := pw.PrintSVG(testSVGTextGradientStrokeFixture(), 0, 0, &width, nil); err != nil {
			t.Fatal(err)
		}
		pw.close()
	})
	if !strings.Contains(msg, "svg: <text> stroke: gradient stroke on text is not yet implemented") {
		t.Fatalf("expected text stroke gradient warning, got %q", msg)
	}
}

func TestPNGInfo_UnsupportedBitDepth(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	data := mustEncodePNG(t, img)
	mutated := append([]byte(nil), data...)
	mutated[24] = 16
	if _, err := pngInfo(mutated); err != errUnsupportedPNG {
		t.Fatalf("expected errUnsupportedPNG, got %v", err)
	}
}

func TestPNGInfo_BadHeader(t *testing.T) {
	data := mustEncodePNG(t, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	mutated := append([]byte(nil), data...)
	binary.BigEndian.PutUint32(mutated[8:12], 12)
	if _, err := pngInfo(mutated); err != errBadPNG {
		t.Fatalf("expected errBadPNG, got %v", err)
	}
}
