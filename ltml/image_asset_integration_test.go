// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rowland/leadtype/ltml/ltpdf"
)

func encodeTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xFF})
	img.Set(1, 0, color.RGBA{R: 0x40, G: 0x50, B: 0x60, A: 0xFF})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func renderWithAssetFS(t *testing.T, doc *Doc) string {
	t.Helper()
	imageFS := fstest.MapFS{
		"logo.png": &fstest.MapFile{Data: encodeTestPNG(t)},
	}

	w := ltpdf.NewDocWriter()
	w.SetAssetFS(imageFS)
	if err := doc.Print(w); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := w.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func renderDirectPDFWithAssetFS(t *testing.T) string {
	t.Helper()
	w := ltpdf.NewDocWriter()
	w.SetAssetFS(fstest.MapFS{
		"logo.png": &fstest.MapFile{Data: encodeTestPNG(t)},
	})
	w.NewPage()
	if _, _, err := w.PrintImageFile("logo.png", 1, 1, nil, nil); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := w.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestStdImage_UsesWriterAssetFS(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml>
  <page>
    <image src="logo.png" />
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	ltmlPDF := renderWithAssetFS(t, doc)
	directPDF := renderDirectPDFWithAssetFS(t)

	for _, pdf := range []string{ltmlPDF, directPDF} {
		if !strings.Contains(pdf, "/Subtype /Image") {
			t.Fatalf("expected rendered PDF to contain an image object, got:\n%s", pdf)
		}
		if !strings.Contains(pdf, "/Im0 Do\n") {
			t.Fatalf("expected rendered PDF to draw an image, got:\n%s", pdf)
		}
	}
	if strings.Count(ltmlPDF, "/Subtype /Image") != strings.Count(directPDF, "/Subtype /Image") {
		t.Fatalf("expected LTML and direct PDF paths to embed the same number of image objects")
	}
}
