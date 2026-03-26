// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"image/jpeg"
	"os"
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

func TestIsJPEG(t *testing.T) {
	if !isJPEG(mustReadTestImage(t)) {
		t.Fatal("expected fixture to be detected as JPEG")
	}
	if isJPEG([]byte("not a jpeg")) {
		t.Fatal("expected non-JPEG bytes to be rejected")
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
