// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/ttf_fonts"
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

	expectedRise := -float64(fonts[0].Ascent()) * 0.001 * 12
	if !strings.Contains(pw.stream.String(), "/F0 12 Tf\n") {
		t.Fatalf("expected font selection, got:\n%s", pw.stream.String())
	}
	if !strings.Contains(pw.stream.String(), g(expectedRise)+" Ts\n") {
		t.Fatalf("expected rise command %q, got:\n%s", g(expectedRise)+" Ts\n", pw.stream.String())
	}
}

func TestPageWriter_flushText_VTextAlignAdjustsUnderline(t *testing.T) {
	newWriter := func(vTextAlign string) (*PageWriter, float64) {
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
		expectedY := float64(fonts[0].UnderlinePosition())*12/1000.0 - pw.vTextAlignPts
		return pw, expectedY
	}

	baseWriter, baseY := newWriter("base")
	topWriter, topY := newWriter("top")

	if !strings.Contains(baseWriter.stream.String(), g(baseY)+" l\n") {
		t.Fatalf("expected base underline to include y coordinate %q, got:\n%s", g(baseY)+" l\n", baseWriter.stream.String())
	}
	if !strings.Contains(topWriter.stream.String(), g(topY)+" l\n") {
		t.Fatalf("expected top-aligned underline to include y coordinate %q, got:\n%s", g(topY)+" l\n", topWriter.stream.String())
	}
	if g(baseY) == g(topY) {
		t.Fatalf("expected underline positions to differ between base and top alignment")
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
	// [4 2] 0 = dashed
	// 0 J = ButtCap
	// 72 = 1 inch
	// 792 - 72 = 720 = 1 inch from top
	// 792 - 144 = 648 = 2 inches from top
	expectS(t, "0 0 1 RG\n3 w\n[4 2] 0 d\n0 J\n72 720 m\n144 648 l\n", pw.stream.String())
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
