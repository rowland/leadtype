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

func TestPageWriter_PrintImage_DefaultSize(t *testing.T) {
	data := mustReadTestImage(t)
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})
	width, height, err := pw.PrintImage(data, 10, 20, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if width != float64(cfg.Width) || height != float64(cfg.Height) {
		t.Fatalf("expected intrinsic size %dx%d, got %.2fx%.2f", cfg.Width, cfg.Height, width, height)
	}
	if !strings.Contains(pw.stream.String(), "/Im0 Do\n") {
		t.Fatalf("expected image draw operator, got:\n%s", pw.stream.String())
	}
	if dw.resources.xObjects["Im0"] == nil {
		t.Fatal("expected image XObject registered in resources")
	}
}

func TestPageWriter_PrintImage_AspectRatio(t *testing.T) {
	data := mustReadTestImage(t)
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{"units": "in"})
	width := 2.0
	actualWidth, actualHeight, err := pw.PrintImage(data, 1, 2, floatPtr(width), nil)
	if err != nil {
		t.Fatal(err)
	}
	expectedHeight := width * float64(cfg.Height) / float64(cfg.Width)
	expectFdelta(t, width, actualWidth, 0.0001)
	expectFdelta(t, expectedHeight, actualHeight, 0.0001)
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
