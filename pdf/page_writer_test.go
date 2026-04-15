// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"math"
	"strings"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/font"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/ttf_fonts"
	"github.com/rowland/leadtype/wordbreaking"
)

const (
	black = 0
	red   = 0xFF0000
	green = 0x00FF00
)

func TestPageWriter_checkSetFillColor(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.fillColor == black, "fillColor should default to black")
	check(t, pw.last.fillColor == black, "last.fillColor should default to black")

	prev := pw.SetFillColor(red)
	check(t, prev == black, "Previous color was black")
	check(t, pw.fillColor == red, "fillColor should now be red")
	check(t, pw.last.fillColor == black, "last.fillColor should still be black")
	pw.checkSetFillColor()
	check(t, pw.last.fillColor == red, "last.fillColor should now be red")

	expectS(t, "1 0 0 rg\n", pw.stream.String())
	// TODO: test for autoPath behavior
}

func TestPageWriter_checkSetFontColor(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.fontColor == black, "fontColor should default to black")
	check(t, pw.last.fillColor == black, "last.fillColor should default to black")

	prev := pw.SetFontColor(red)
	check(t, prev == black, "Previous color was black")
	check(t, pw.fontColor == red, "fontColor should now be red")
	check(t, pw.last.fillColor == black, "last.fontColor should still be black")
	pw.checkSetFontColor()
	check(t, pw.last.fillColor == red, "last.fontColor should now be red")

	expectS(t, "1 0 0 rg\n", pw.stream.String())
	// TODO: test for autoPath behavior
}

func testLinearGradient() *LinearGradient {
	return &LinearGradient{
		X0: 0, Y0: 0, X1: 100, Y1: 0,
		Stops: []GradientStop{
			{Position: 0, Color: colors.Red},
			{Position: 1, Color: colors.Blue},
		},
	}
}

func testRadialGradient() *RadialGradient {
	return &RadialGradient{
		X0: 50, Y0: 50, R0: 0,
		X1: 50, Y1: 50, R1: 50,
		Stops: []GradientStop{
			{Position: 0, Color: colors.White},
			{Position: 1, Color: colors.Black},
		},
	}
}

func TestPageWriter_checkSetLineColor(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.lineColor == black, "lineColor should default to black")
	check(t, pw.last.lineColor == black, "last.lineColor should default to black")
	prev := pw.SetLineColor(red)
	check(t, prev == black, "Previous color was black")
	check(t, pw.lineColor == red, "lineColor should now be red")
	check(t, pw.last.lineColor == black, "last.lineColor should still be black")
	pw.checkSetLineColor()
	check(t, pw.last.lineColor == red, "last.lineColor should now be red")

	expectS(t, "1 0 0 RG\n", pw.stream.String())
	// TODO: test for autoPath behavior
}

func TestPageWriter_checkSetLineDashPattern(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.lineDashPattern == "", "lineDashPattern should default to unset")
	check(t, pw.last.lineDashPattern == "", "last.lineDashPattern should default to unset")

	prevLDP := pw.SetLineDashPattern("solid")
	check(t, prevLDP == "", "Previous pattern was unset")
	check(t, pw.lineDashPattern == "solid", "lineDashPattern should now be solid")
	check(t, pw.last.lineDashPattern == "", "last.lineDashPattern should still be unset")

	check(t, pw.lineCapStyle == ButtCap, "lineCapStyle should default to ButtCap")
	check(t, pw.last.lineCapStyle == ButtCap, "last.lineCapStyle should default to ButtCap")

	prevLCS := pw.SetLineCapStyle(RoundCap)
	check(t, prevLCS == ButtCap, "Previous style was ButtCap")
	check(t, pw.lineCapStyle == RoundCap, "lineCapStyle should now be RoundCap")
	check(t, pw.last.lineCapStyle == ButtCap, "last.lineCapStyle should still be ButtCap")

	pw.checkSetLineDashPattern()
	check(t, pw.last.lineDashPattern == "solid", "last.lineDashPattern should now be solid")
	check(t, pw.last.lineCapStyle == RoundCap, "last.lineCapStyle should now be RoundCap")

	expectS(t, "[] 0 d\n1 J\n", pw.stream.String())
}

func TestPageWriter_checkSetLineWidth(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.lineWidth == 0, "lineWidth should default to 0")
	check(t, pw.last.lineWidth == 0, "last.lineWidth should default to 0")

	prev := pw.SetLineWidth(72, "pt")
	check(t, prev == 0, "Previous width was 0")
	check(t, pw.lineWidth == 72, "lineWidth should now be red")
	check(t, pw.last.lineWidth == 0, "last.lineWidth should still be 0")

	pw.checkSetLineWidth()
	check(t, pw.last.lineWidth == 72, "last.lineWidth should now be 72")

	prev = pw.SetLineWidth(2, "in")
	check(t, prev == 1, "Previous width was 1 inch")
	check(t, pw.lineWidth == 144, "lineWidth should be stored in points")

	expectS(t, "72 w\n", pw.stream.String())
	// TODO: test for autoPath behavior
}

func TestPageWriter_MeasureText(t *testing.T) {
	dw := NewDocWriter()
	fonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fonts)
	pw := newPageWriter(dw, options.Options{"units": "in"})
	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}

	rt, err := rich_text.New("Hello", pw.Fonts(), pw.FontSize(), options.Options{
		"color": pw.fontColor, "strikeout": pw.strikeout, "underline": pw.underline,
	})
	if err != nil {
		t.Fatal(err)
	}

	metrics, err := pw.MeasureText("Hello")
	if err != nil {
		t.Fatal(err)
	}

	expectFdelta(t, UnitConversions["in"].fromPts(rt.Width()), metrics.Width, 0.0001)
	expectFdelta(t, UnitConversions["in"].fromPts(rt.Height()), metrics.Height, 0.0001)
	expectFdelta(t, UnitConversions["in"].fromPts(rt.Ascent()), metrics.Ascent, 0.0001)
	expectFdelta(t, UnitConversions["in"].fromPts(rt.Descent()), metrics.Descent, 0.0001)
}

func TestPageWriter_PrintParagraph_RightAlignConvertsWidthFromCurrentUnits(t *testing.T) {
	dw := NewDocWriter()
	fonts, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fonts)
	pw := newPageWriter(dw, options.Options{"units": "in"})
	fontsUsed, err := pw.SetFont("Helvetica", 12, options.Options{})
	if err != nil {
		t.Fatal(err)
	}

	rt, err := rich_text.New("Hello", fontsUsed, pw.FontSize(), nil)
	if err != nil {
		t.Fatal(err)
	}

	pw.MoveTo(0.75, 1.0)
	pw.PrintParagraph([]*rich_text.RichText{rt}, options.Options{
		"text-align": "right",
		"width":      6.0,
	})

	want := pw.units.toPts(0.75+6.0) - rt.Width()
	if math.Abs(pw.last.loc.X-want) > 0.001 {
		t.Fatalf("right-aligned paragraph started at %v pt, want %v pt", pw.last.loc.X, want)
	}
}

func TestClonePageWriter(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	pw.SetLineCapStyle(ProjectingSquareCap)
	pw.SetLineColor(colors.AliceBlue)
	pw.SetLineDashPattern("dotted")
	pw.SetLineWidth(42, "pt")

	pwc := clonePageWriter(pw)
	check(t, pwc.LineCapStyle() == ProjectingSquareCap, "LineCapStyle should be ProjectingSquareCap")
	check(t, pwc.LineColor() == colors.AliceBlue, "LineColor should be AliceBlue")
	check(t, pwc.LineDashPattern() == "dotted", "LineDashPattern should be 'dotted'")
	check(t, pwc.LineWidth("pt") == 42, "LineWidth should be 42")
}

func TestPageWriter_Line(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{"units": "in"})

	pw.Line(1, 1, 0, 2)

	expectS(t, "72 720 m\n216 720 l\nS\n", pw.stream.String())
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

func TestPageWriter_PointsForCircle(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	points := pw.PointsForCircle(2, 3, 1)
	expectI(t, 13, len(points))
	expectF(t, 3, points[0].X)
	expectF(t, 3, points[0].Y)
	expectF(t, 3, points[len(points)-1].X)
	expectF(t, 3, points[len(points)-1].Y)
}

func TestPageWriter_PointsForArc(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	points := pw.PointsForArc(0, 0, 1, 0, 180)
	expectI(t, 7, len(points))
	expectFdelta(t, 1, points[0].X, 0.0001)
	expectFdelta(t, 0, points[0].Y, 0.0001)
	expectFdelta(t, -1, points[len(points)-1].X, 0.0001)
	expectFdelta(t, 0, points[len(points)-1].Y, 0.0001)
}

func TestPageWriter_Polygon_InvalidSides(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	if err := pw.Polygon(1, 1, 1, 2, true, false, false, 0); err != errInvalidPolygonSides {
		t.Fatalf("expected errInvalidPolygonSides, got %v", err)
	}
}

func TestPageWriter_Star_InvalidPoints(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	if err := pw.Star(1, 1, 1, 0.5, 4, true, false, false, 0); err != errInvalidStarPoints {
		t.Fatalf("expected errInvalidStarPoints, got %v", err)
	}
}

func TestPageWriter_CircleFillAndStroke(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})
	pw.SetFillColor(red)
	pw.SetLineColor(green)

	if err := pw.Circle(2, 2, 1, true, true, false); err != nil {
		t.Fatalf("Circle returned error: %v", err)
	}
	got := pw.stream.String()
	if !strings.Contains(got, "1 0 0 rg\n") {
		t.Fatalf("expected fill color command, got:\n%s", got)
	}
	if !strings.Contains(got, "0 1 0 RG\n") {
		t.Fatalf("expected line color command, got:\n%s", got)
	}
	if !strings.Contains(got, "B\n") {
		t.Fatalf("expected fill-and-stroke operator, got:\n%s", got)
	}
}

func TestPageWriter_CircleUsesTranslatedMoveTo(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{"units": "in"})

	if err := pw.Circle(2, 2, 1, true, false, false); err != nil {
		t.Fatalf("Circle returned error: %v", err)
	}
	if !strings.Contains(pw.stream.String(), "216 648 m\n") {
		t.Fatalf("expected translated moveTo at circle start, got:\n%s", pw.stream.String())
	}
}

func TestPageWriter_Rotate(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{"units": "in"})

	if err := pw.Rotate(90, 1, 2, func() {
		pw.Line(1, 1, 0, 1)
	}); err != nil {
		t.Fatalf("Rotate returned error: %v", err)
	}

	expectS(t, "q\n0 1 -1 0 720 576 cm\n72 720 m\n144 720 l\nS\nQ\n", pw.stream.String())
}

func TestPageWriter_RotateFlushesQueuedRichTextInsideTransform(t *testing.T) {
	dw := NewDocWriter()
	afmfc, err := afm_fonts.Default()
	if err != nil {
		t.Fatalf("Default AFM fonts returned error: %v", err)
	}
	dw.AddFontSource(afmfc)
	pw := newPageWriter(dw, options.Options{"units": "in"})

	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatalf("SetFont returned error: %v", err)
	}

	rt, err := pw.richTextForString("rotate me")
	if err != nil {
		t.Fatalf("richTextForString returned error: %v", err)
	}

	if err := pw.Rotate(25, 1, 2, func() {
		pw.PrintRichText(rt)
	}); err != nil {
		t.Fatalf("Rotate returned error: %v", err)
	}

	got := pw.stream.String()
	idxQ := strings.Index(got, "Q\n")
	idxBT := strings.Index(got, "BT\n")
	idxET := strings.Index(got, "ET\n")
	if idxBT < 0 || idxET < 0 || idxQ < 0 {
		t.Fatalf("expected rotated text operators in stream, got:\n%s", got)
	}
	if idxBT > idxQ {
		t.Fatalf("expected text block to begin before graphics restore, got:\n%s", got)
	}
	if idxET > idxQ {
		t.Fatalf("expected text block to end before graphics restore, got:\n%s", got)
	}
}

func TestPageWriter_RotateRestoresGraphicsStateCache(t *testing.T) {
	dw := NewDocWriter()
	afmfc, err := afm_fonts.Default()
	if err != nil {
		t.Fatalf("Default AFM fonts returned error: %v", err)
	}
	dw.AddFontSource(afmfc)
	pw := newPageWriter(dw, options.Options{"units": "in"})

	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatalf("SetFont returned error: %v", err)
	}

	pw.SetFillColor(colors.LemonChiffon)
	pw.Rectangle(1, 1, 2, 0.5, false, true)

	if err := pw.Rotate(30, 2, 2, func() {
		pw.MoveTo(1.2, 1.4)
		if err := pw.Print("inside"); err != nil {
			t.Fatalf("Print(inside) returned error: %v", err)
		}
	}); err != nil {
		t.Fatalf("Rotate returned error: %v", err)
	}

	pw.MoveTo(1, 3)
	if err := pw.Print("after"); err != nil {
		t.Fatalf("Print(after) returned error: %v", err)
	}
	pw.flushText()

	got := pw.stream.String()
	lastBT := strings.LastIndex(got, "BT\n")
	if lastBT < 0 {
		t.Fatalf("expected trailing text object, got:\n%s", got)
	}
	lastBlock := got[lastBT:]
	if !strings.Contains(lastBlock, "0 0 0 rg\n") {
		t.Fatalf("expected post-rotate text to reissue black fill color, got:\n%s", lastBlock)
	}
}

func TestPageWriter_StartTextReissuesFontInsideFreshTextObject(t *testing.T) {
	dw := NewDocWriter()
	afmfc, err := afm_fonts.Default()
	if err != nil {
		t.Fatalf("Default AFM fonts returned error: %v", err)
	}
	dw.AddFontSource(afmfc)
	pw := newPageWriter(dw, options.Options{"units": "in"})

	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatalf("SetFont returned error: %v", err)
	}

	if err := pw.Print("before"); err != nil {
		t.Fatalf("Print(before) returned error: %v", err)
	}
	pw.endText()

	pw.MoveTo(1, 2)
	if err := pw.Print("after"); err != nil {
		t.Fatalf("Print(after) returned error: %v", err)
	}
	pw.flushText()

	got := pw.stream.String()
	blocks := strings.Split(got, "BT\n")
	if len(blocks) < 3 {
		t.Fatalf("expected two text objects, got:\n%s", got)
	}
	second := blocks[2]
	if !strings.Contains(second, "/F0 12 Tf\n") {
		t.Fatalf("expected second text object to reissue font selection, got:\n%s", second)
	}
}

func TestPageWriter_Scale(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{"units": "in"})

	if err := pw.Scale(1, 2, 2, 0.5, func() {
		pw.Rectangle(1, 1, 1, 0.5, true, false)
	}); err != nil {
		t.Fatalf("Scale returned error: %v", err)
	}

	expectS(t, "q\n2 0 0 0.5 -72 324 cm\n72 684 72 36 re\nS\nQ\n", pw.stream.String())
}

func TestPageWriter_RotateInsideManualPath(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.Path(func() {
		if rotateErr := pw.Rotate(45, 1, 1, func() {}); rotateErr != errTransformInsideManualPath {
			t.Fatalf("expected errTransformInsideManualPath, got %v", rotateErr)
		}
		_ = pw.Fill()
	})
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}
}

func TestPageWriter_flushText(t *testing.T) {
	skipIfNoTTFFonts(t)
	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}

	dw := NewDocWriter()
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	pw.SetFont("Arial", 12, options.Options{})
	pw.Print("Hello")
	pw.Print(", World!")
	pw.flushText()
	expectS(t, "BT\n/F0 12 Tf\n(Hello, World!) Tj\n", pw.stream.String())
}

func TestPageWriter_ClipText(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	pw.MoveTo(72, 720)
	called := false
	if err := pw.ClipText("Hello", func() {
		called = true
		pw.SetFillColor(red)
	}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected ClipText callback to run")
	}

	got := pw.stream.String()
	for _, fragment := range []string{"q\nBT\n", "/F0 12 Tf\n", "7 Tr\n", "(Hello) Tj\n", "ET\n", "Q\n"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected clipped text output to contain %q, got:\n%s", fragment, got)
		}
	}
	metrics, err := pw.MeasureText("Hello")
	if err != nil {
		t.Fatal(err)
	}
	if delta := math.Abs(pw.loc.X - (72 + pw.units.toPts(metrics.Width))); delta > 0.001 {
		t.Fatalf("expected ClipText to advance location to %.3f pt, got %.3f pt", 72+pw.units.toPts(metrics.Width), pw.loc.X)
	}
}

func TestPageWriter_FillStrokeClipText(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	pw.SetLineColor(red)
	pw.SetLineWidth(2, "pt")
	pw.MoveTo(72, 720)
	if err := pw.FillStrokeClipText("Hello", func() {
		pw.SetFillColor(green)
	}); err != nil {
		t.Fatal(err)
	}

	got := pw.stream.String()
	for _, fragment := range []string{"6 Tr\n", "1 0 0 RG\n", "2 w\n", "(Hello) Tj\n"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected fill+stroke clipped text output to contain %q, got:\n%s", fragment, got)
		}
	}
}

func TestPageWriter_ClipRichTextSuppressesDecorationsLinksAndAccessibility(t *testing.T) {
	dw := NewDocWriter()
	dw.EnableTaggedPDF(true)
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	fonts, err := pw.SetFont("Helvetica", 12, options.Options{})
	if err != nil {
		t.Fatal(err)
	}
	rt, err := rich_text.New("Hello", fonts, 12, options.Options{
		"underline":    true,
		"strikeout":    true,
		"link_uri":     "https://example.com",
		"char_spacing": 0.5,
	})
	if err != nil {
		t.Fatal(err)
	}
	pw.MoveTo(72, 720)
	if err := pw.WithAccessibilityTag("P", AccessibilityOptions{}, func() {
		if err := pw.ClipRichText(rt, nil); err != nil {
			t.Fatal(err)
		}
	}); err != nil {
		t.Fatal(err)
	}

	got := pw.stream.String()
	for _, fragment := range []string{" BDC\n", "/ActualText", " S\n"} {
		if strings.Contains(got, fragment) {
			t.Fatalf("expected clipped rich text to suppress %q, got:\n%s", fragment, got)
		}
	}
	if len(pw.page.annots) != 0 {
		t.Fatalf("expected clipped rich text to avoid link annotations, got %d", len(pw.page.annots))
	}
}

func TestPageWriter_ClipRichText_UnicodePositionedGlyphs(t *testing.T) {
	skipIfNoTTFFonts(t)
	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}

	dw := NewDocWriter()
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	fonts, err := pw.SetFont("Arial", 12, options.Options{})
	if err != nil {
		t.Fatal(err)
	}
	rt, err := rich_text.New("Hi", fonts, 12, options.Options{"char_spacing": 1.0})
	if err != nil {
		t.Fatal(err)
	}
	pw.MoveTo(72, 720)
	if err := pw.ClipRichText(rt, nil); err != nil {
		t.Fatal(err)
	}

	got := pw.stream.String()
	for _, fragment := range []string{"7 Tr\n", " Tm\n", "<"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("expected clipped rich text positioned glyph output to contain %q, got:\n%s", fragment, got)
		}
	}
}

func TestPageWriter_SetVTextAlign(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	expectS(t, "base", pw.VTextAlign())
	prev := pw.SetVTextAlign("top")
	expectS(t, "base", prev)
	expectS(t, "top", pw.VTextAlign())
}

func TestPageWriter_flushText_VTextAlignRise(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	fonts, err := pw.SetFont("Courier", 12, options.Options{})
	if err != nil {
		t.Fatal(err)
	}
	pw.SetVTextAlign("top")
	if err := pw.Print("Hi"); err != nil {
		t.Fatal(err)
	}
	pw.flushText()

	scale := 12 * 0.001
	if upm := fonts[0].UnitsPerEm(); upm > 0 {
		scale = 12 / float64(upm)
	}
	expectedRise := -float64(fonts[0].CapHeight()) * scale
	if expectedRise == 0 {
		expectedRise = -float64(fonts[0].Ascent()) * scale
	}
	if !strings.Contains(pw.stream.String(), "/F0 12 Tf\n") {
		t.Fatalf("expected font selection, got:\n%s", pw.stream.String())
	}
	if !strings.Contains(pw.stream.String(), g(expectedRise)+" Ts\n") {
		t.Fatalf("expected rise command %q, got:\n%s", g(expectedRise)+" Ts\n", pw.stream.String())
	}
}

func TestPageWriter_flushText_VTextAlignAdjustsUnderline(t *testing.T) {
	newWriter := func(vTextAlign string) (*PageWriter, float64, float64) {
		dw := NewDocWriter()
		fc, err := afm_fonts.Default()
		if err != nil {
			t.Fatal(err)
		}
		dw.AddFontSource(fc)
		pw := dw.NewPage()
		fonts, err := pw.SetFont("Courier", 12, options.Options{})
		if err != nil {
			t.Fatal(err)
		}
		pw.SetUnderline(true)
		pw.SetVTextAlign(vTextAlign)
		if err := pw.Print("Hi"); err != nil {
			t.Fatal(err)
		}
		pw.flushText()
		rise := 0.0
		scale := 12 * 0.001
		if upm := fonts[0].UnitsPerEm(); upm > 0 {
			scale = 12 / float64(upm)
		}
		top := float64(fonts[0].CapHeight()) * scale
		if top == 0 {
			top = float64(fonts[0].Ascent()) * scale
		}
		descent := float64(fonts[0].Descent()) * scale
		switch vTextAlign {
		case "above":
			rise = -(top - descent)
		case "top":
			rise = -top
		case "middle":
			rise = -((top + descent) / 2.0)
		case "below":
			rise = -descent
		}
		hWidth, _ := fonts[0].AdvanceWidth('H')
		iWidth, _ := fonts[0].AdvanceWidth('i')
		expectedY := float64(fonts[0].UnderlinePosition())*12/1000.0 + rise
		return pw, expectedY, 12 * (float64(hWidth+iWidth) / 1000.0)
	}

	baseWriter, baseY, baseWidth := newWriter("base")
	topWriter, topY, topWidth := newWriter("top")

	baseUnderline := "0 " + g(baseY) + " m\n" + g(baseWidth) + " " + g(baseY) + " l\n"
	topUnderline := "0 " + g(topY) + " m\n" + g(topWidth) + " " + g(topY) + " l\n"

	if !strings.Contains(baseWriter.stream.String(), baseUnderline) {
		t.Fatalf("expected base underline segment %q, got:\n%s", baseUnderline, baseWriter.stream.String())
	}
	if !strings.Contains(topWriter.stream.String(), topUnderline) {
		t.Fatalf("expected top-aligned underline segment %q, got:\n%s", topUnderline, topWriter.stream.String())
	}
	if g(baseY) == g(topY) {
		t.Fatalf("expected underline positions to differ between base and top alignment")
	}
}

func TestPageWriter_flushText_DecorationPayloadCapOverrideRestoresLineCap(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	pw.SetLineCapStyle(RoundCap)
	pw.MoveTo(72, 760)
	pw.LineTo(96, 760)

	if _, err := pw.SetFont("Courier", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	rt, err := rich_text.New("Hi", pw.Fonts(), 12, options.Options{
		"underline": true,
		"decoration": &rich_text.DecorationOverrides{
			Underline: rich_text.DecorationLineOverrides{CapStyle: "butt_cap"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	pw.MoveTo(72, 720)
	pw.PrintRichText(rt)
	pw.flushText()

	pw.MoveTo(72, 700)
	pw.LineTo(96, 700)

	got := pw.stream.String()
	firstRound := strings.Index(got, "1 J\n")
	if firstRound < 0 {
		t.Fatalf("expected initial round cap command, got:\n%s", got)
	}
	butt := strings.Index(got[firstRound+1:], "0 J\n")
	if butt < 0 {
		t.Fatalf("expected underline cap override to emit butt cap command, got:\n%s", got)
	}
	butt += firstRound + 1
	secondRound := strings.Index(got[butt+1:], "1 J\n")
	if secondRound < 0 {
		t.Fatalf("expected line cap to be restored after underline, got:\n%s", got)
	}
}

func TestPageWriter_flushText_DecorationWithoutCapOverridePreservesAmbientLineCap(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	pw.SetLineCapStyle(RoundCap)
	pw.MoveTo(72, 760)
	pw.LineTo(96, 760)

	if _, err := pw.SetFont("Courier", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	pw.SetUnderline(true)
	pw.MoveTo(72, 720)
	if err := pw.Print("Hi"); err != nil {
		t.Fatal(err)
	}
	pw.flushText()

	if strings.Contains(pw.stream.String(), "0 J\n") {
		t.Fatalf("expected underline without cap override to preserve ambient line cap, got:\n%s", pw.stream.String())
	}
}

func TestPageWriter_flushText_DecorationPayloadColorOverrideRestoresLineColor(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	pw.SetLineColor(colors.Black)
	pw.MoveTo(72, 760)
	pw.LineTo(96, 760)

	if _, err := pw.SetFont("Courier", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	rt, err := rich_text.New("Hi", pw.Fonts(), 12, options.Options{
		"underline": true,
		"decoration": &rich_text.DecorationOverrides{
			Underline: rich_text.DecorationLineOverrides{Color: colors.Red, HasColor: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	pw.MoveTo(72, 720)
	pw.PrintRichText(rt)
	pw.flushText()

	pw.MoveTo(72, 700)
	pw.LineTo(96, 700)

	got := pw.stream.String()
	firstRed := strings.Index(got, "1 0 0 RG\n")
	if firstRed < 0 {
		t.Fatalf("expected underline color override to emit red stroke color, got:\n%s", got)
	}
	restoredBlack := strings.Index(got[firstRed+1:], "0 0 0 RG\n")
	if restoredBlack < 0 {
		t.Fatalf("expected line color to be restored after underline, got:\n%s", got)
	}
}

func TestPageWriter_flushText_DecorationPayloadColorOverrideTemporarilySuppressesLineGradient(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	if err := pw.SetLineLinearGradient(testLinearGradient()); err != nil {
		t.Fatal(err)
	}
	pw.MoveTo(72, 760)
	pw.LineTo(96, 760)

	if _, err := pw.SetFont("Courier", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	rt, err := rich_text.New("Hi", pw.Fonts(), 12, options.Options{
		"underline": true,
		"decoration": &rich_text.DecorationOverrides{
			Underline: rich_text.DecorationLineOverrides{Color: colors.Red, HasColor: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	pw.MoveTo(72, 720)
	pw.PrintRichText(rt)
	pw.flushText()

	pw.MoveTo(72, 700)
	pw.LineTo(96, 700)

	got := pw.stream.String()
	expectI(t, 2, strings.Count(got, "/Pattern CS"))
	expectI(t, 2, strings.Count(got, "SCN"))
	expectI(t, 1, strings.Count(got, "1 0 0 RG\n"))
}

func TestPageWriter_flushText_DecorationPayloadOverridesAmbientDecorationSettings(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	pw.SetLineCapStyle(ButtCap)
	pw.SetLineColor(colors.Black)
	pw.SetLineDashPattern("solid")
	if _, err := pw.SetFont("Courier", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}

	rt, err := rich_text.New("Hi", pw.Fonts(), 12, options.Options{
		"underline": true,
		"decoration": &rich_text.DecorationOverrides{
			Underline: rich_text.DecorationLineOverrides{
				Color:       colors.Red,
				HasColor:    true,
				Width:       3,
				HasWidth:    true,
				Pattern:     "dashed",
				HasPattern:  true,
				CapStyle:    "round_cap",
				Position:    6,
				HasPosition: true,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	pw.MoveTo(72, 720)
	pw.PrintRichText(rt)
	pw.flushText()

	got := pw.stream.String()
	expectedY := pw.translate(720) + 6
	expectedLine := g(72.0) + " " + g(expectedY) + " m\n" + g(72.0+rt.Width()) + " " + g(expectedY) + " l\n"
	if !strings.Contains(got, expectedLine) {
		t.Fatalf("expected payload position override to move underline to %q, got:\n%s", expectedLine, got)
	}
	if strings.Count(got, "1 0 0 RG\n") != 1 {
		t.Fatalf("expected payload color override to apply, got:\n%s", got)
	}
	if !strings.Contains(got, "1 J\n") {
		t.Fatalf("expected payload cap override to emit round cap, got:\n%s", got)
	}
	if !strings.Contains(got, "[12 6] 0 d\n") {
		t.Fatalf("expected payload dash override to emit dashed pattern scaled by line width, got:\n%s", got)
	}
}

func TestPageWriter_MoveToPreservesPendingTextSettings(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	if _, err := pw.SetFont("Courier", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	pw.SetFontColor(colors.FireBrick)
	pw.SetVTextAlign("above")
	if err := pw.Print("First"); err != nil {
		t.Fatal(err)
	}

	pw.SetFontColor(colors.DarkBlue)
	pw.SetVTextAlign("top")
	pw.MoveTo(1, 2)
	if err := pw.Print("Second"); err != nil {
		t.Fatal(err)
	}
	pw.flushText()

	got := pw.stream.String()
	if !strings.Contains(got, "0 0 0.5451 rg\n") {
		t.Fatalf("expected second line color command for DarkBlue, got:\n%s", got)
	}
	scale := 12 * 0.001
	if upm := pw.fonts[0].UnitsPerEm(); upm > 0 {
		scale = 12 / float64(upm)
	}
	topRiseValue := -float64(pw.fonts[0].CapHeight()) * scale
	if topRiseValue == 0 {
		topRiseValue = -float64(pw.fonts[0].Ascent()) * scale
	}
	topRise := g(topRiseValue)
	if !strings.Contains(got, topRise+" Ts\n") {
		t.Fatalf("expected second line top rise %q, got:\n%s", topRise+" Ts\n", got)
	}
}

func TestPageWriter_SetVTextAlignFlushesPendingText(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	fonts, err := pw.SetFont("Courier", 12, options.Options{})
	if err != nil {
		t.Fatal(err)
	}
	pw.SetVTextAlign("above")
	if err := pw.Print("Above"); err != nil {
		t.Fatal(err)
	}
	pw.SetVTextAlign("top")
	if err := pw.Print("Top"); err != nil {
		t.Fatal(err)
	}
	pw.flushText()

	scale := 12 * 0.001
	if upm := fonts[0].UnitsPerEm(); upm > 0 {
		scale = 12 / float64(upm)
	}
	topValue := float64(fonts[0].CapHeight()) * scale
	if topValue == 0 {
		topValue = float64(fonts[0].Ascent()) * scale
	}
	descent := float64(fonts[0].Descent()) * scale
	aboveRise := g(-(topValue - descent))
	topRise := g(-topValue)
	got := pw.stream.String()

	if !strings.Contains(got, aboveRise+" Ts\n(Above) Tj\n") {
		t.Fatalf("expected pending text to flush with above rise before alignment change, got:\n%s", got)
	}
	if !strings.Contains(got, topRise+" Ts\n(Top) Tj\n") {
		t.Fatalf("expected later text to use top rise, got:\n%s", got)
	}
}

func TestPageWriter_PrintParagraph_JustifyLeavesLastLineAtLeftEdge(t *testing.T) {
	skipIfNoTTFFonts(t)
	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}

	dw := NewDocWriter()
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	if _, err := pw.SetFont("Arial", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	rt, err := pw.richTextForString("one two three four five six seven eight nine ten eleven")
	if err != nil {
		t.Fatal(err)
	}
	flags := make([]wordbreaking.Flags, rt.Len())
	wordbreaking.MarkRuneAttributes(rt.String(), flags)
	para := rt.WrapToWidth(140, flags, false)
	if len(para) < 2 {
		t.Fatalf("expected wrapped paragraph to produce multiple lines, got %d", len(para))
	}

	pw.MoveTo(72, 720)
	pw.PrintParagraph(para, options.Options{"text-align": "justify", "width": 140})
	pw.flushText()

	got := pw.stream.String()
	if strings.Count(got, " 720 Td\n") > 1 {
		t.Fatalf("expected later lines not to drift right from the left edge, got:\n%s", got)
	}
	lastBT := strings.LastIndex(got, "BT\n")
	if lastBT < 0 {
		t.Fatalf("expected text output, got:\n%s", got)
	}
	lastBlock := got[lastBT:]
	if strings.Contains(lastBlock, " Tc\n") || strings.Contains(lastBlock, " Tw\n") {
		t.Fatalf("expected final line not to carry justification spacing, got:\n%s", lastBlock)
	}
}

func TestPageWriter_PrintParagraph_UsesNextLineLeadingForAdvance(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	fonts12, err := pw.SetFont("Courier", 12, options.Options{})
	if err != nil {
		t.Fatal(err)
	}
	line1, err := rich_text.New("small", fonts12, 12, options.Options{})
	if err != nil {
		t.Fatal(err)
	}
	fonts24, err := pw.SetFont("Courier", 24, options.Options{})
	if err != nil {
		t.Fatal(err)
	}
	line2, err := rich_text.New("BIG", fonts24, 24, options.Options{})
	if err != nil {
		t.Fatal(err)
	}

	pw.MoveTo(72, 720)
	startY := pw.loc.Y
	pw.PrintParagraph([]*rich_text.RichText{line1, line2}, options.Options{})

	expectedY := startY - (2 * line2.Leading() * pw.lineSpacing)
	if delta := math.Abs(pw.loc.Y - expectedY); delta > 0.001 {
		t.Fatalf("expected final Y %.3f pt using next-line leading, got %.3f pt", expectedY, pw.loc.Y)
	}
}

func TestPageWriter_FlushText_PositionedTrueTypeLeafSetsFontBeforeGlyphs(t *testing.T) {
	skipIfNoTTFFonts(t)
	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}

	dw := NewDocWriter()
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	regular, err := font.New("Arial", options.Options{}, dw.fontSources)
	if err != nil {
		t.Fatal(err)
	}
	bold, err := font.New("Arial", options.Options{"weight": "Bold"}, dw.fontSources)
	if err != nil {
		t.Fatal(err)
	}

	first, err := rich_text.New("plain ", []*font.Font{regular}, 12, options.Options{
		"char_spacing": 1.0,
		"word_spacing": 2.0,
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := rich_text.New("bold", []*font.Font{bold}, 12, options.Options{
		"char_spacing": 1.0,
		"color":        colors.Red,
	})
	if err != nil {
		t.Fatal(err)
	}

	pw.MoveTo(72, 720)
	pw.PrintRichText(first.AddPiece(second))
	pw.flushText()

	got := pw.stream.String()
	secondTm := fmt.Sprintf("1 0 0 1 %s 720 Tm\n", g(72+first.Width()))
	secondTmIndex := strings.Index(got, secondTm)
	if secondTmIndex < 0 {
		t.Fatalf("expected positioned glyph matrix for second leaf, got:\n%s", got)
	}

	boldKey := dw.fontKeyUnicode(bold)
	boldTf := fmt.Sprintf("/%s 12 Tf\n", boldKey)
	boldTfIndex := strings.Index(got, boldTf)
	if boldTfIndex < 0 {
		t.Fatalf("expected bold font selection in stream, got:\n%s", got)
	}
	if boldTfIndex > secondTmIndex {
		t.Fatalf("expected second leaf font selection before positioned glyphs, got:\n%s", got)
	}

	redIndex := strings.Index(got, "1 0 0 rg\n")
	if redIndex < 0 {
		t.Fatalf("expected second leaf color in stream, got:\n%s", got)
	}
	if redIndex > secondTmIndex {
		t.Fatalf("expected second leaf color before positioned glyphs, got:\n%s", got)
	}
}

func TestPageWriter_PathDiscardRestoresState(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.Path(func() {
		pw.SetLineColor(red)
		pw.MoveTo(1, 1)
		pw.LineTo(2, 2)
	})
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}

	if len(pw.pathStates) != 0 {
		t.Fatal("manual path stack should be empty after Path returns")
	}
	if !pw.autoPath {
		t.Fatal("autoPath should be restored after Path returns")
	}
	if pw.inPath {
		t.Fatal("inPath should be cleared after discarding unfinished path")
	}
	if pw.LineColor() != black {
		t.Fatalf("expected line color restored to black, got %v", pw.LineColor())
	}
	if !strings.Contains(pw.stream.String(), "n\n") {
		t.Fatalf("expected discarded path to emit new path operator, got:\n%s", pw.stream.String())
	}
}

func TestPageWriter_ClipTextInsideManualPath(t *testing.T) {
	dw := NewDocWriter()
	fc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw := dw.NewPage()

	if _, err := pw.SetFont("Helvetica", 12, options.Options{}); err != nil {
		t.Fatal(err)
	}
	err = pw.Path(func() {
		pw.MoveTo(1, 1)
		pw.LineTo(2, 2)
		if clipErr := pw.ClipText("Hello", nil); clipErr != errTextClipInsideManualPath {
			t.Fatalf("expected errTextClipInsideManualPath, got %v", clipErr)
		}
	})
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}
}

func TestPageWriter_FillOutsidePathErrors(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	if err := pw.Fill(); err != errNoActivePath {
		t.Fatalf("expected errNoActivePath, got %v", err)
	}
	if err := pw.Stroke(); err != errNoActivePath {
		t.Fatalf("expected errNoActivePath, got %v", err)
	}
	if err := pw.FillAndStroke(); err != errNoActivePath {
		t.Fatalf("expected errNoActivePath, got %v", err)
	}
	if err := pw.Clip(nil); err != errNoActivePath {
		t.Fatalf("expected errNoActivePath, got %v", err)
	}
}

func TestPageWriter_PathStroke(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.Path(func() {
		pw.MoveTo(1, 1)
		pw.LineTo(2, 2)
		if err := pw.Stroke(); err != nil {
			t.Fatalf("Stroke returned error: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}

	expectS(t, "1 791 m\n2 790 l\nS\n", pw.stream.String())
}

func TestPageWriter_PathFillAndStrokeRestoresColors(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.Path(func() {
		pw.SetFillColor(red)
		pw.SetLineColor(green)
		pw.MoveTo(1, 1)
		pw.LineTo(2, 2)
		if err := pw.FillAndStroke(); err != nil {
			t.Fatalf("FillAndStroke returned error: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}

	if pw.fillColor != black {
		t.Fatalf("expected fill color restored to black, got %v", pw.fillColor)
	}
	if pw.LineColor() != black {
		t.Fatalf("expected line color restored to black, got %v", pw.LineColor())
	}
	if !strings.Contains(pw.stream.String(), "1 0 0 rg\n") {
		t.Fatalf("expected fill color command in stream, got:\n%s", pw.stream.String())
	}
	if !strings.Contains(pw.stream.String(), "0 1 0 RG\n") {
		t.Fatalf("expected line color command in stream, got:\n%s", pw.stream.String())
	}
	if !strings.Contains(pw.stream.String(), "B\n") {
		t.Fatalf("expected fill-and-stroke operator in stream, got:\n%s", pw.stream.String())
	}
}

func TestPageWriter_PathClipScopesGraphicsState(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.Path(func() {
		pw.MoveTo(1, 1)
		pw.LineTo(2, 2)
		pw.LineTo(1, 2)
		pw.LineTo(1, 1)
		if err := pw.Clip(func() {
			pw.Rectangle(3, 3, 1, 1, true, false)
		}); err != nil {
			t.Fatalf("Clip returned error: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("Path returned error: %v", err)
	}

	got := pw.stream.String()
	if !strings.Contains(got, "q\n") || !strings.Contains(got, "Q\n") {
		t.Fatalf("expected save/restore graphics state in clip output, got:\n%s", got)
	}
	if !strings.Contains(got, "W\nn\n") {
		t.Fatalf("expected clip and path reset operators, got:\n%s", got)
	}
	if !strings.Contains(got, "re\nS\n") {
		t.Fatalf("expected clipped drawing inside scope, got:\n%s", got)
	}
}

func TestPageWriter_FontSize(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.FontSize() == 0, "Should default to zero")
	prev := pw.SetFontSize(12.5)
	check(t, prev == 0, "Should return previous value")
	check(t, pw.FontSize() == 12.5, "Should now be 12.5")
}

func TestPageWriter_FontStyle(t *testing.T) {
	skipIfNoTTFFonts(t)
	dw := NewDocWriter()

	ttfc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(ttfc)

	afmfc, err := afm_fonts.Default()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(afmfc)

	pw := dw.NewPage()

	check(t, pw.Fonts() == nil, "Page font list should be empty by default")
	check(t, pw.FontStyle() == "", "Should default to empty string")

	prev, err := pw.SetFontStyle("Italic")
	expectS(t, "No current font to apply style Italic to.", err.Error())
	check(t, prev == "", "Should return empty string when no font is set")

	_, err = pw.AddFont("Arial", options.Options{})
	check(t, err == nil, "Failed to add Arial")
	_, err = pw.AddFont("Helvetica", options.Options{})
	check(t, err == nil, "Failed to add Helvetica")

	fonts := pw.Fonts()
	checkFatal(t, len(fonts) == 2, "length of fonts should be 2")
	expectS(t, "", fonts[0].Style())
	expectS(t, "", fonts[1].Style())

	prev, err = pw.SetFontStyle("Italic")
	check(t, err == nil, "Error setting style to Italic.")
	check(t, pw.FontStyle() == "Italic", "FontStyle should now be Italic")
	fonts = pw.Fonts()
	checkFatal(t, len(fonts) == 2, "length of fonts should be 2")

	expectS(t, "Italic", fonts[0].Style())
	expectS(t, "Italic", fonts[1].Style())
}

func TestPageWriter_LineCapStyle(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.LineCapStyle() == ButtCap, "Should default to ButtCap")

	last := pw.SetLineCapStyle(RoundCap)
	check(t, pw.LineCapStyle() == RoundCap, "Should have changed to RoundCap")
	check(t, last == ButtCap, "Previous style was ButtCap")

	last = pw.SetLineCapStyle(ProjectingSquareCap)
	check(t, last == RoundCap, "Previous style was RoundCap")
	check(t, pw.LineCapStyle() == ProjectingSquareCap, "New style is ProjectingSquareCap")
}

func TestPageWriter_LineColor(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.LineColor() == black, "Should default to black")

	last := pw.SetLineColor(red) // previous color was black
	check(t, pw.LineColor() == red, "Should have changed to red")
	check(t, last == black, "Previous color was black")

	last = pw.SetLineColor(green)
	check(t, last == red, "Previous color was red")
	check(t, pw.LineColor() == green, "New color is green")
}

func TestPageWriter_LineDashPattern(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.LineDashPattern() == "", "Should default to unset")

	last := pw.SetLineDashPattern("solid")
	check(t, pw.LineDashPattern() == "solid", "Should have changed to solid")
	check(t, last == "", "Previous pattern was unset")

	last = pw.SetLineDashPattern("dotted")
	check(t, pw.LineDashPattern() == "dotted", "Should have changed to solid")
	check(t, last == "solid", "Previous pattern was solid")
}

func TestPageWriter_Strikeout(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.Strikeout() == false, "Should default to false")

	last := pw.SetStrikeout(true)
	check(t, pw.Strikeout() == true, "Should have changed to true")
	check(t, last == false, "Previous setting was false")
}

func TestPageWriter_LineTo(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.stream.String() == "", "Stream should default to empty")
	check(t, pw.inPath == false, "Should not be in path by default")
	pw.SetLineColor(colors.Blue)
	pw.SetLineWidth(3, "pt")
	pw.SetLineDashPattern("dashed")

	pw.SetUnits("in")
	pw.MoveTo(1, 1)
	pw.LineTo(2, 2)

	check(t, pw.inPath == true, "Should now be in path")
	// 0 0 1 RG = Blue
	// 3 w = 3 pts wide
	// [12 6] 0 = dashed scaled by 3 pt line width
	// 0 J = ButtCap
	// 72 = 1 inch
	// 792 - 72 = 720 = 1 inch from top
	// 792 - 144 = 648 = 2 inches from top
	expectS(t, "0 0 1 RG\n3 w\n[12 6] 0 d\n0 J\n72 720 m\n144 648 l\n", pw.stream.String())
}

func TestPageWriter_Line_CustomDashPatternScalesWithLineWidth(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	pw.SetLineColor(colors.Blue)
	pw.SetLineWidth(2, "pt")
	pw.SetLineDashPattern("[2 2 4 4 6 6 8 8] 0")
	pw.SetUnits("in")
	pw.MoveTo(1, 1)
	pw.LineTo(2, 1)

	expectS(t, "0 0 1 RG\n2 w\n[4 4 8 8 12 12 16 16] 0 d\n0 J\n72 720 m\n144 720 l\n", pw.stream.String())
}

func TestPageWriter_LineWidth(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.LineWidth("pt") == 0, "Should default to zero")
	prev := pw.SetLineWidth(72, "pt")
	check(t, prev == 0, "Should return previous value")
	check(t, pw.LineWidth("pt") == 72, "Should now be 72")
	check(t, pw.LineWidth("in") == 1, "72 points is 1 inch")
	check(t, pw.SetLineWidth(200, "in") == 1, "Previous value was 1 inch")
}

func TestPageWriter_MoveTo(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})
	expectF(t, 0, pw.loc.X)
	expectF(t, 0, pw.loc.Y)
	pw.MoveTo(222, 333)
	expectF(t, 222, pw.loc.X)
	expectF(t, 792-333, pw.loc.Y)
}

func TestPageWriter_PageHeight(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})
	// height in (default) points
	expectF(t, 792, pw.PageHeight())
	pw.SetUnits("in")
	expectF(t, 11, pw.PageHeight())
	pw.SetUnits("cm")
	expectFdelta(t, 27.93, pw.PageHeight(), 0.01)
	// custom: "Dave points"
	UnitConversions.Add("dp", 0.072)
	pw.SetUnits("dp")
	expectF(t, 11000, pw.PageHeight())
}

func TestPageWriter_PageWidth(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})
	// width in (default) points
	expectF(t, 612, pw.PageWidth())
	pw.SetUnits("in")
	expectF(t, 8.5, pw.PageWidth())
	pw.SetUnits("cm")
	expectFdelta(t, 21.58, pw.PageWidth(), 0.01)
	// custom: "Dave points"
	UnitConversions.Add("dp", 0.072)
	pw.SetUnits("dp")
	expectF(t, 8500, pw.PageWidth())
}

func TestPageWriter_Print(t *testing.T) {
	skipIfNoTTFFonts(t)
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw.SetFont("Arial", 12, options.Options{})

	check(t, pw.line == nil, "line should start out nil")
	pw.Print("Hello")
	check(t, pw.line != nil, "line should no longer be nil")
	expectS(t, "Hello", pw.line.String())
	pw.Print(", World!")
	expectS(t, "Hello, World!", pw.line.String())
}

func TestPageWriter_SetFont(t *testing.T) {
	skipIfNoTTFFonts(t)
	dw := NewDocWriter()

	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)

	pw := dw.NewPage()

	check(t, pw.Fonts() == nil, "Page font list should be empty by default")

	fonts, _ := pw.SetFont("Arial", 12, options.Options{"color": green})

	checkFatal(t, len(fonts) == 1, "length of fonts should be 1")
	expectS(t, "Arial", fonts[0].Family())
	check(t, pw.FontColor() == green, "FontColor should now be green")
	expectF(t, 12, pw.FontSize())
	expectF(t, 1.0, fonts[0].RelativeSize)
	check(t, fonts[0] == pw.Fonts()[0], "SetFont should return new font list")
	check(t, fonts[0] == dw.Fonts()[0], "SetFont changes to font list should be global")
}

func TestPageWriter_SetUnits(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})
	// defaults to points
	expectS(t, "pt", pw.Units())
	// inches
	pw.SetUnits("in")
	expectS(t, "in", pw.Units())
	// centimeters
	pw.SetUnits("cm")
	expectS(t, "cm", pw.Units())
	// custom: "Dave points"
	UnitConversions.Add("dp", 0.072)
	pw.SetUnits("dp")
	expectS(t, "dp", pw.Units())
}

func TestPageWriter_startGraph(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.gw != nil, "GraphWriter should not be nil.")
	check(t, !pw.inGraph, "Should not be in graph mode by default.")
	pw.last.loc = Location{7, 7} // prove last.loc gets reset

	pw.startGraph()

	check(t, pw.inGraph, "Should be in graph mode now.")
	check(t, pw.last.loc.equal(Location{0, 0}), "startGraph should reset location")
}

func TestPageWriter_translate(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})
	actual := pw.translate(200)
	expectF(t, 792-200, actual)
}

func TestPageWriter_Underline(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.Underline() == false, "Should default to false")

	last := pw.SetUnderline(true)
	check(t, pw.Underline() == true, "Should have changed to true")
	check(t, last == false, "Previous setting was false")
}

func TestPageWriter_units(t *testing.T) {
	dw := NewDocWriter()
	// defaults to points
	pw1 := newPageWriter(dw, options.Options{})
	expectS(t, "pt", pw1.units.name)
	expectF(t, 1, pw1.units.ratio)
	// inches
	pw2 := newPageWriter(dw, options.Options{"units": "in"})
	expectS(t, "in", pw2.units.name)
	expectF(t, 72, pw2.units.ratio)
	// centimeters
	pw3 := newPageWriter(dw, options.Options{"units": "cm"})
	expectS(t, "cm", pw3.units.name)
	expectF(t, 28.35, pw3.units.ratio)
	// custom: "Dave points"
	UnitConversions.Add("dp", 0.072)
	pw4 := newPageWriter(dw, options.Options{"units": "dp"})
	expectS(t, "dp", pw4.units.name)
	expectF(t, 0.072, pw4.units.ratio)
}

func TestPageWriter_Write(t *testing.T) {
	skipIfNoTTFFonts(t)
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	fc, err := ttf_fonts.NewFromSystemFonts()
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc)
	pw.SetFont("Arial", 12, options.Options{})

	check(t, pw.line == nil, "line should start out nil")
	fmt.Fprint(pw, "Hello")
	check(t, pw.line != nil, "line should no longer be nil")
	expectS(t, "Hello", pw.line.String())
	fmt.Fprint(pw, ", World!")
	expectS(t, "Hello, World!", pw.line.String())
}

func TestPageWriter_checkSetFillGradient(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.fillGradient == "", "fillGradient should default to empty")
	err := pw.SetFillLinearGradient(testLinearGradient())
	check(t, err == nil, "SetFillLinearGradient should not return error")
	check(t, pw.fillGradient != "", "fillGradient should be set")

	pw.checkSetFillColor()
	s := pw.stream.String()
	check(t, strings.Contains(s, "/Pattern cs"), "should set Pattern color space")
	check(t, strings.Contains(s, "scn"), "should use scn operator")
}

func TestPageWriter_FillGradientRevertToSolid(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	pw.SetFillLinearGradient(testLinearGradient())
	pw.checkSetFillColor()

	// Now clear and set solid color
	pw.ClearFillGradient()
	pw.SetFillColor(colors.Green)
	pw.checkSetFillColor()

	s := pw.stream.String()
	check(t, strings.Contains(s, "rg"), "should revert to rg operator for solid color")
}

func TestPageWriter_FillGradientAfterFontColorReset(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.SetFillLinearGradient(testLinearGradient()) == nil, "gradient should be set")
	pw.checkSetFillColor()
	pw.SetFontColor(colors.Green)
	pw.checkSetFontColor()
	pw.checkSetFillColor()

	s := pw.stream.String()
	expectI(t, 2, strings.Count(s, "/Pattern cs"))
	expectI(t, 2, strings.Count(s, "scn"))
	expectI(t, 1, strings.Count(s, " rg\n"))
}

func TestPageWriter_checkSetLineGradient(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.SetLineLinearGradient(testLinearGradient())
	check(t, err == nil, "SetLineLinearGradient should not return error")
	check(t, pw.lineGradient != "", "lineGradient should be set")

	pw.checkSetLineColor()
	s := pw.stream.String()
	check(t, strings.Contains(s, "/Pattern CS"), "should set Pattern color space for stroke")
	check(t, strings.Contains(s, "SCN"), "should use SCN operator")
}

func TestPageWriter_LineGradientAfterSolidStrokeReset(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.SetLineLinearGradient(testLinearGradient()) == nil, "gradient should be set")
	pw.checkSetLineColor()
	pw.ClearLineGradient()
	pw.SetLineColor(colors.Green)
	pw.checkSetLineColor()
	check(t, pw.SetLineLinearGradient(testLinearGradient()) == nil, "gradient should be set again")
	pw.checkSetLineColor()

	s := pw.stream.String()
	expectI(t, 2, strings.Count(s, "/Pattern CS"))
	expectI(t, 2, strings.Count(s, "SCN"))
	expectI(t, 1, strings.Count(s, " RG\n"))
}

func TestPageWriter_SetFillRadialGradient(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.SetFillRadialGradient(testRadialGradient())
	check(t, err == nil, "SetFillRadialGradient should not return error")
	check(t, pw.fillGradient != "", "fillGradient should be set for radial")
}

func TestPageWriter_SetFillGradient_ValidationError(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.SetFillLinearGradient(&LinearGradient{
		X0: 0, Y0: 0, X1: 100, Y1: 0,
		Stops: []GradientStop{
			{Position: 0, Color: colors.Red},
		},
	})
	check(t, err != nil, "single stop should return validation error")
}

func TestPageWriter_PaintLinearGradient(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.PaintLinearGradient(testLinearGradient())
	check(t, err == nil, "PaintLinearGradient should not return error")
	s := pw.stream.String()
	check(t, strings.Contains(s, "sh"), "should contain sh operator")
}

func TestPageWriter_GradientCaching(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	lg := testLinearGradient()
	pw.SetFillLinearGradient(lg)
	name1 := pw.fillGradient
	pw.ClearFillGradient()
	pw.SetFillLinearGradient(lg)
	name2 := pw.fillGradient
	expectS(t, name1, name2)
}

func TestPageWriter_SetGradient_NilErrors(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.SetFillLinearGradient(nil) != nil, "nil fill linear gradient should error")
	check(t, pw.SetFillRadialGradient(nil) != nil, "nil fill radial gradient should error")
	check(t, pw.SetLineLinearGradient(nil) != nil, "nil line linear gradient should error")
	check(t, pw.SetLineRadialGradient(nil) != nil, "nil line radial gradient should error")
	check(t, pw.PaintLinearGradient(nil) != nil, "nil linear shading should error")
	check(t, pw.PaintRadialGradient(nil) != nil, "nil radial shading should error")
}

func TestPageWriter_PathRestoresGradientStateAfterFill(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.SetFillLinearGradient(testLinearGradient()) == nil, "outer gradient should be set")
	outer := pw.fillGradient

	err := pw.Path(func() {
		check(t, pw.SetFillRadialGradient(testRadialGradient()) == nil, "inner gradient should be set")
		pw.MoveTo(0, 0)
		pw.LineTo(10, 0)
		pw.LineTo(10, 10)
		pw.LineTo(0, 10)
		pw.LineTo(0, 0)
		check(t, pw.Fill() == nil, "Fill should succeed")
	})
	check(t, err == nil, "Path should succeed")
	expectS(t, outer, pw.fillGradient)
}

func TestPageWriter_PathRestoresGradientStateAfterStroke(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.SetLineLinearGradient(testLinearGradient()) == nil, "outer line gradient should be set")
	outer := pw.lineGradient

	err := pw.Path(func() {
		pw.ClearLineGradient()
		pw.SetLineColor(colors.Green)
		pw.MoveTo(0, 0)
		pw.LineTo(10, 10)
		check(t, pw.Stroke() == nil, "Stroke should succeed")
	})
	check(t, err == nil, "Path should succeed")
	expectS(t, outer, pw.lineGradient)
}

func TestPageWriter_PathRestoresGradientStateAfterFillAndStroke(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.SetFillLinearGradient(testLinearGradient()) == nil, "outer fill gradient should be set")
	check(t, pw.SetLineLinearGradient(testLinearGradient()) == nil, "outer line gradient should be set")
	outerFill := pw.fillGradient
	outerLine := pw.lineGradient

	err := pw.Path(func() {
		pw.ClearFillGradient()
		pw.SetFillColor(colors.LightGray)
		pw.ClearLineGradient()
		pw.SetLineColor(colors.Black)
		pw.MoveTo(0, 0)
		pw.LineTo(10, 0)
		pw.LineTo(10, 10)
		pw.LineTo(0, 10)
		pw.LineTo(0, 0)
		check(t, pw.FillAndStroke() == nil, "FillAndStroke should succeed")
	})
	check(t, err == nil, "Path should succeed")
	expectS(t, outerFill, pw.fillGradient)
	expectS(t, outerLine, pw.lineGradient)
}

func TestPageWriter_PathRestoresGradientStateAfterClip(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.SetFillLinearGradient(testLinearGradient()) == nil, "outer fill gradient should be set")
	check(t, pw.SetLineLinearGradient(testLinearGradient()) == nil, "outer line gradient should be set")
	outerFill := pw.fillGradient
	outerLine := pw.lineGradient

	err := pw.Path(func() {
		pw.ClearFillGradient()
		pw.SetFillColor(colors.Green)
		pw.ClearLineGradient()
		pw.SetLineColor(colors.Black)
		pw.MoveTo(0, 0)
		pw.LineTo(10, 0)
		pw.LineTo(10, 10)
		pw.LineTo(0, 10)
		pw.LineTo(0, 0)
		check(t, pw.Clip(func() {
			check(t, pw.PaintLinearGradient(testLinearGradient()) == nil, "PaintLinearGradient should succeed")
		}) == nil, "Clip should succeed")
	})
	check(t, err == nil, "Path should succeed")
	expectS(t, outerFill, pw.fillGradient)
	expectS(t, outerLine, pw.lineGradient)

	s := pw.stream.String()
	check(t, strings.Contains(s, "q\nW\nn\n"), "clip should save graphics state, clip, and clear the path")
	check(t, strings.Contains(s, "sh\nQ\n"), "clip should paint shading and restore graphics state")
}

func TestPageWriter_ClipRestoresLastStateForFontColor(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	pw.SetFillColor(colors.HoneyDew)
	pw.checkSetFillColor()

	err := pw.Path(func() {
		pw.MoveTo(0, 0)
		pw.LineTo(10, 0)
		pw.LineTo(10, 10)
		pw.LineTo(0, 10)
		pw.LineTo(0, 0)
		check(t, pw.Clip(func() {
			pw.SetFontColor(colors.Black)
			pw.checkSetFontColor()
		}) == nil, "Clip should succeed")
	})
	check(t, err == nil, "Path should succeed")
	check(t, pw.last.fillColor == colors.HoneyDew, "clip should restore cached outer fill color")

	pw.SetFontColor(colors.Black)
	pw.checkSetFontColor()
	expectI(t, 2, strings.Count(pw.stream.String(), "0 0 0 rg\n"))
}

func TestPageWriter_PaintSweepBandConstantColorSegment(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.PaintSweepBand(&SweepBand{
		X: 10, Y: 10,
		InnerRadius: 2,
		OuterRadius: 4,
		Segments: []SweepBandSegment{
			{StartAngle: 0, EndAngle: 45, StartColor: colors.Red, EndColor: colors.Red},
		},
	})
	check(t, err == nil, "PaintSweepBand should succeed")
	s := pw.stream.String()
	check(t, strings.Contains(s, "q\nW\nn\n"), "constant-color segment should clip before painting")
	check(t, strings.Contains(s, "sh\nQ\n"), "constant-color segment should use the same shading paint path and restore graphics state")
}

func TestPageWriter_PaintSweepBandGradientSegment(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.PaintSweepBand(&SweepBand{
		X: 10, Y: 10,
		InnerRadius: 2,
		OuterRadius: 4,
		Segments: []SweepBandSegment{
			{StartAngle: 45, EndAngle: 135, StartColor: colors.Red, EndColor: colors.Blue},
		},
	})
	check(t, err == nil, "PaintSweepBand should succeed")
	s := pw.stream.String()
	check(t, strings.Contains(s, "q\nW\nn\n"), "gradient segment should clip before painting")
	check(t, strings.Contains(s, "sh\nQ\n"), "gradient segment should paint shading and restore graphics state")
}

func TestPageWriter_PaintSweepBandDoesNotLeakState(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	check(t, pw.SetFillLinearGradient(testLinearGradient()) == nil, "outer fill gradient should be set")
	check(t, pw.SetLineLinearGradient(testLinearGradient()) == nil, "outer line gradient should be set")
	savedState := pw.drawState
	savedLast := pw.last

	err := pw.PaintSweepBand(&SweepBand{
		X: 10, Y: 10,
		InnerRadius: 2,
		OuterRadius: 4,
		Segments: []SweepBandSegment{
			{StartAngle: 0, EndAngle: 45, StartColor: colors.Red, EndColor: colors.Red},
			{StartAngle: 45, EndAngle: 90, StartColor: colors.Red, EndColor: colors.Yellow},
		},
	})
	check(t, err == nil, "PaintSweepBand should succeed")
	expectV(t, savedState, pw.drawState)
	expectV(t, savedLast, pw.last)
}

func TestPageWriter_PaintSweepBandGradientCoords(t *testing.T) {
	dw := NewDocWriter()
	pw := dw.NewPage()

	err := pw.PaintSweepBand(&SweepBand{
		X: 10, Y: 10,
		InnerRadius: 2,
		OuterRadius: 4,
		Segments: []SweepBandSegment{
			{StartAngle: 45, EndAngle: 135, StartColor: colors.Red, EndColor: colors.Blue},
			{StartAngle: -45, EndAngle: 45, StartColor: colors.Yellow, EndColor: colors.Green},
			{StartAngle: 135, EndAngle: 180, StartColor: colors.Blue, EndColor: colors.Red},
		},
	})
	check(t, err == nil, "PaintSweepBand should succeed")

	var buf bytes.Buffer
	_, err = dw.WriteTo(&buf)
	check(t, err == nil, "WriteTo should succeed")
	pdf := buf.String()
	check(t, strings.Contains(pdf, "/Coords [12.1213 784.1213 7.8787 784.1213 ]"), "top segment should use a horizontal gradient axis")
	check(t, strings.Contains(pdf, "/Coords [12.1213 779.8787 12.1213 784.1213 ]"), "right segment should use a vertical gradient axis")
	check(t, strings.Contains(pdf, "/Coords [7.8787 784.1213 7 782 ]"), "corner segment should use a diagonal gradient axis")
}

func TestPageWriter_PaintSweepBandStepsEmitMultipleShadings(t *testing.T) {
	dw := NewDocWriter()
	pw := newPageWriter(dw, options.Options{})

	err := pw.PaintSweepBand(&SweepBand{
		X: 10, Y: 10,
		InnerRadius: 2,
		OuterRadius: 4,
		Segments: []SweepBandSegment{
			{StartAngle: 0, EndAngle: 90, StartColor: colors.Red, EndColor: colors.Blue, Steps: 4},
		},
	})
	check(t, err == nil, "PaintSweepBand should succeed")
	s := pw.stream.String()
	expectI(t, 4, strings.Count(s, "q\nW\nn\n"))
	expectI(t, 4, strings.Count(s, "sh\nQ\n"))
}

func TestPageWriter_PaintSweepBandStepsPreserveCoverage(t *testing.T) {
	seg := SweepBandSegment{
		StartAngle: 45,
		EndAngle:   135,
		StartColor: colors.Red,
		EndColor:   colors.Blue,
		Steps:      3,
	}
	expanded := seg.expand()
	expectFdelta(t, 45, expanded[0].StartAngle, 0.0001)
	expectFdelta(t, 135, expanded[len(expanded)-1].EndAngle, 0.0001)
}
