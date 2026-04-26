// Copyright 2026 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"image"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

func TestDocWriter_MemoizeForm_ReusesFormAcrossPages(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.SetUnits("in")
	fonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fonts)

	render := func(pw *PageWriter) error {
		pw.MoveTo(0.15, 0.3)
		return pw.Print("Memo form")
	}

	pw1 := dw.NewPage()
	if _, err := pw1.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	if err := pw1.MemoizeForm("memo-card", 1, 1, 2.25, 0.6, render); err != nil {
		t.Fatal(err)
	}

	pw2 := dw.NewPage()
	if _, err := pw2.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	if err := pw2.MemoizeForm("memo-card", 0.75, 1.5, 2.25, 0.6, render); err != nil {
		t.Fatal(err)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	if count := strings.Count(pdfText, "/Subtype /Form"); count != 1 {
		t.Fatalf("expected one memoized form, got %d\n%s", count, pdfText)
	}
	if count := strings.Count(pdfText, "/Mf0 Do"); count != 2 {
		t.Fatalf("expected two placements of memoized form, got %d\n%s", count, pdfText)
	}
}

func TestDocWriter_MemoizeFormOnCanvas_ReusesFormAcrossDifferentPlacementSizes(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.SetUnits("pt")
	pw := dw.NewPage()

	render := func(pw *PageWriter) error {
		pw.SetLineWidth(2, "pt")
		pw.SetLineColor(colors.DarkBlue)
		return pw.Circle(120, 120, 100, true, false, false)
	}

	if err := pw.MemoizeFormOnCanvas("memo-board", 72, 72, 96, 96, 240, 240, render); err != nil {
		t.Fatal(err)
	}
	if err := pw.MemoizeFormOnCanvas("memo-board", 204, 72, 48, 48, 240, 240, render); err != nil {
		t.Fatal(err)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	if count := strings.Count(pdfText, "/Subtype /Form"); count != 1 {
		t.Fatalf("expected one memoized form for shared canvas geometry, got %d\n%s", count, pdfText)
	}
	if count := strings.Count(pdfText, "/Mf0 Do"); count != 2 {
		t.Fatalf("expected two placements of shared memoized form, got %d\n%s", count, pdfText)
	}
}

func TestDocWriter_MemoizeForm_DifferentSizesDoNotCollide(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.SetUnits("in")
	fonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fonts)

	pw := dw.NewPage()
	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	render := func(pw *PageWriter) error {
		pw.MoveTo(0.1, 0.25)
		return pw.Print("Sized form")
	}

	if err := pw.MemoizeForm("memo-card", 1, 1, 2, 0.5, render); err != nil {
		t.Fatal(err)
	}
	if err := pw.MemoizeForm("memo-card", 1, 2, 2.75, 0.5, render); err != nil {
		t.Fatal(err)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	if count := strings.Count(pdfText, "/Subtype /Form"); count != 2 {
		t.Fatalf("expected two memoized forms for distinct sizes, got %d\n%s", count, pdfText)
	}
	if !strings.Contains(pdfText, "/Mf0 Do") || !strings.Contains(pdfText, "/Mf1 Do") {
		t.Fatalf("expected both memoized form placements, got:\n%s", pdfText)
	}
}

func TestDocWriter_MemoizeFormOnCanvas_DifferentCanvasSizesDoNotCollide(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.SetUnits("pt")
	pw := dw.NewPage()

	render := func(canvas float64) func(*PageWriter) error {
		return func(pw *PageWriter) error {
			pw.SetLineWidth(2, "pt")
			pw.SetLineColor(colors.FireBrick)
			return pw.Circle(canvas/2, canvas/2, canvas*0.4, true, false, false)
		}
	}

	if err := pw.MemoizeFormOnCanvas("memo-board", 72, 72, 60, 60, 180, 180, render(180)); err != nil {
		t.Fatal(err)
	}
	if err := pw.MemoizeFormOnCanvas("memo-board", 150, 72, 60, 60, 240, 240, render(240)); err != nil {
		t.Fatal(err)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	if count := strings.Count(pdfText, "/Subtype /Form"); count != 2 {
		t.Fatalf("expected two memoized forms for distinct canvas sizes, got %d\n%s", count, pdfText)
	}
	if !strings.Contains(pdfText, "/Mf0 Do") || !strings.Contains(pdfText, "/Mf1 Do") {
		t.Fatalf("expected both memoized form placements, got:\n%s", pdfText)
	}
}

func TestDocWriter_MemoizeForm_ReusesNestedImageResources(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.SetUnits("in")
	dw.SetAssetFS(fstest.MapFS{
		"logo.png": &fstest.MapFile{Data: mustEncodePNG(t, image.NewGray(image.Rect(0, 0, 8, 4)))},
	})
	pw := dw.NewPage()

	render := func(pw *PageWriter) error {
		width := 1.0
		_, _, err := pw.PrintImageFile("logo.png", 0, 0, &width, nil)
		return err
	}

	if err := pw.MemoizeForm("memo-image", 1, 1, 1, 0.5, render); err != nil {
		t.Fatal(err)
	}
	if err := pw.MemoizeForm("memo-image", 2.5, 1, 1, 0.5, render); err != nil {
		t.Fatal(err)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	if count := strings.Count(pdfText, "/Subtype /Form"); count != 1 {
		t.Fatalf("expected one memoized form, got %d\n%s", count, pdfText)
	}
	if count := strings.Count(pdfText, "/Subtype /Image"); count != 1 {
		t.Fatalf("expected one nested image resource, got %d\n%s", count, pdfText)
	}
}

func TestDocWriter_MemoizeForm_SuppressesInnerTargetLinks(t *testing.T) {
	var buf bytes.Buffer

	dw := newAFMDocWriter(t)
	pw := dw.NewPage()

	if err := pw.MemoizeForm("memo-target", 1, 1, 2, 0.5, func(form *PageWriter) error {
		form.RegisterDestination("memo-target", 0.1, 0.1)
		return form.AddTargetLink(0.1, 0.1, 0.75, 0.2, "memo-target")
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	if strings.Contains(pdfText, "/Subtype /Link") {
		t.Fatalf("memoized form should suppress inner target annotations, got:\n%s", pdfText)
	}
	if strings.Contains(pdfText, "/S /GoTo") {
		t.Fatalf("memoized form should suppress inner destinations, got:\n%s", pdfText)
	}
}

func TestDocWriter_MemoizeForm_SuppressesInnerLinksAndAccessibility(t *testing.T) {
	var buf bytes.Buffer

	dw := NewDocWriter()
	dw.SetUnits("in")
	dw.EnableTaggedPDF(true)
	fonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fonts)

	pw := dw.NewPage()
	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	rt, err := rich_text.New("Memoized link", pw.Fonts(), pw.FontSize(), options.Options{
		"link_uri": "https://example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	var memoErr error
	if err := pw.WithAccessibilityTag("Figure", AccessibilityOptions{ID: "outer-form"}, func() {
		memoErr = pw.MemoizeForm("memo-link", 1, 1, 2.5, 0.5, func(form *PageWriter) error {
			form.MoveTo(0.1, 0.3)
			form.PrintRichText(rt)
			return nil
		})
	}); err != nil {
		t.Fatal(err)
	}
	if memoErr != nil {
		t.Fatal(memoErr)
	}

	if _, err := dw.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	pdfText := buf.String()
	if strings.Contains(pdfText, "/Subtype /Link") {
		t.Fatalf("memoized form should suppress inner link annotations, got:\n%s", pdfText)
	}
	if !strings.Contains(pdfText, "/StructTreeRoot") {
		t.Fatalf("expected tagged PDF structure, got:\n%s", pdfText)
	}
	if count := strings.Count(pdfText, "BDC\n"); count != 1 {
		t.Fatalf("expected one outer tagged placement and no inner tagged text, got %d\n%s", count, pdfText)
	}
}
