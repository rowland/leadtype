// Copyright 2026 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"regexp"
	"strconv"
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

func TestStdImage_TableLayoutWithAssetFS_UsesNonZeroInferredDimensions(t *testing.T) {
	doc, err := Parse([]byte(`
<ltml units="in" margin="0.5">
  <page>
    <table order="rows" cols="1" width="100%">
      <div layout="vbox" padding="6pt">
        <image src="logo.png" width="1.5" border="dotted" padding="3pt" />
        <image src="logo.png" height="0.75" border="dotted" padding="3pt" />
        <image src="logo.png" border="dotted" padding="3pt" />
      </div>
    </table>
  </page>
</ltml>`))
	if err != nil {
		t.Fatal(err)
	}

	pdf := renderWithAssetFS(t, doc)
	matches := imageMatrixRE.FindAllStringSubmatch(pdf, -1)
	if len(matches) != 3 {
		t.Fatalf("image draw count = %d, want 3; pdf:\n%s", len(matches), pdf)
	}

	for i, match := range matches {
		width, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			t.Fatalf("parsing width %q: %v", match[1], err)
		}
		height, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			t.Fatalf("parsing height %q: %v", match[2], err)
		}
		if width <= 0 || height <= 0 {
			t.Fatalf("image %d matrix width=%v height=%v, want both > 0; pdf:\n%s", i, width, height, pdf)
		}
	}
}

var imageMatrixRE = regexp.MustCompile(`(?m)^([0-9.]+) 0 0 ([0-9.]+) [0-9.]+ [0-9.]+ cm\n/Im[0-9]+ Do$`)
