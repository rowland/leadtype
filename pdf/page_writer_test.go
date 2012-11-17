// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"testing"
)

const (
	black = 0
	red   = 0xFF0000
	green = 0x00FF00
)

func TestPageWriter_checkSetLineColor(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	check(t, pw.lineColor == black, "lineColor should default to black")
	check(t, pw.last.lineColor == black, "last.lineColor should default to black")
	pw.SetLineColor(red)
	check(t, pw.lineColor == red, "lineColor should now be red")
	check(t, pw.last.lineColor == black, "last.lineColor should still be black")
	pw.checkSetLineColor()
	check(t, pw.last.lineColor == red, "last.lineColor should now be red")

	expectS(t, "1 0 0 RG\n", pw.stream.String())
	// TODO: test for autoPath behavior
}

func TestPageWriter_checkSetLineDashPattern(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	check(t, pw.lineDashPattern == "", "lineDashPattern should default to unset")
	check(t, pw.last.lineDashPattern == "", "last.lineDashPattern should default to unset")

	pw.SetLineDashPattern("solid")
	check(t, pw.lineDashPattern == "solid", "lineDashPattern should now be solid")
	check(t, pw.last.lineDashPattern == "", "last.lineDashPattern should still be unset")

	check(t, pw.lineCapStyle == ButtCap, "lineCapStyle should default to ButtCap")
	check(t, pw.last.lineCapStyle == ButtCap, "last.lineCapStyle should default to ButtCap")

	pw.SetLineCapStyle(RoundCap)
	check(t, pw.lineCapStyle == RoundCap, "lineCapStyle should now be RoundCap")
	check(t, pw.last.lineCapStyle == ButtCap, "last.lineCapStyle should still be ButtCap")

	pw.checkSetLineDashPattern()
	check(t, pw.last.lineDashPattern == "solid", "last.lineDashPattern should now be solid")
	check(t, pw.last.lineCapStyle == RoundCap, "last.lineCapStyle should now be RoundCap")

	expectS(t, "[] 0 d\n1 J\n", pw.stream.String())
}

func TestPageWriter_checkSetLineWidth(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	check(t, pw.lineWidth == 0, "lineWidth should default to 0")
	check(t, pw.last.lineWidth == 0, "last.lineWidth should default to 0")

	pw.SetLineWidth(72, "pt")
	check(t, pw.lineWidth == 72, "lineWidth should now be red")
	check(t, pw.last.lineWidth == 0, "last.lineWidth should still be 0")

	pw.checkSetLineWidth()
	check(t, pw.last.lineWidth == 72, "last.lineWidth should now be 72")

	pw.SetLineWidth(2, "in")
	check(t, pw.lineWidth == 144, "lineWidth should be stored in points")

	expectS(t, "72 w\n", pw.stream.String())
	// TODO: test for autoPath behavior
}

func TestclonePageWriter(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	pw.SetLineCapStyle(ProjectingSquareCap)
	pw.SetLineColor(AliceBlue)
	pw.SetLineDashPattern("dotted")
	pw.SetLineWidth(42, "pt")

	pwc := clonePageWriter(pw)
	check(t, pwc.LineCapStyle() == ProjectingSquareCap, "LineCapStyle should be ProjectingSquareCap")
	check(t, pwc.LineColor() == AliceBlue, "LineColor should be AliceBlue")
	check(t, pwc.LineDashPattern() == "dotted", "LineDashPattern should be 'dotted'")
	check(t, pwc.LineWidth("pt") == 42, "LineWidth should be 42")
}

func TestPageWriter_LineCapStyle(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	check(t, pw.LineCapStyle() == ButtCap, "Should default to ButtCap")

	last := pw.SetLineCapStyle(RoundCap)
	check(t, pw.LineCapStyle() == RoundCap, "Should have changed to RoundCap")
	check(t, last == ButtCap, "Previous style was ButtCap")

	last = pw.SetLineCapStyle(ProjectingSquareCap)
	check(t, last == RoundCap, "Previous style was RoundCap")
	check(t, pw.LineCapStyle() == ProjectingSquareCap, "New style is ProjectingSquareCap")
}

func TestPageWriter_LineColor(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	check(t, pw.LineColor() == black, "Should default to black")

	last := pw.SetLineColor(red) // previous color was black
	check(t, pw.LineColor() == red, "Should have changed to red")
	check(t, last == black, "Previous color was black")

	last = pw.SetLineColor(green)
	check(t, last == red, "Previous color was red")
	check(t, pw.LineColor() == green, "New color is green")
}

func TestPageWriter_LineDashPattern(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	check(t, pw.LineDashPattern() == "", "Should default to unset")

	last := pw.SetLineDashPattern("solid")
	check(t, pw.LineDashPattern() == "solid", "Should have changed to solid")
	check(t, last == "", "Previous pattern was unset")

	last = pw.SetLineDashPattern("dotted")
	check(t, pw.LineDashPattern() == "dotted", "Should have changed to solid")
	check(t, last == "solid", "Previous pattern was solid")
}

func TestPageWriter_LineTo(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	check(t, pw.stream.String() == "", "Stream should default to empty")
	check(t, pw.inPath == false, "Should not be in path by default")
	pw.SetLineColor(Blue)
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
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	check(t, pw.LineWidth("pt") == 0, "Should default to zero")
	prev := pw.SetLineWidth(72, "pt")
	check(t, prev == 0, "Should return previous value")
	check(t, pw.LineWidth("pt") == 72, "Should now be 72")
	check(t, pw.LineWidth("in") == 1, "72 points is 1 inch")
	check(t, pw.SetLineWidth(200, "in") == 1, "Previous value was 1 inch")
}

func TestPageWriter_MoveTo(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})
	expectF(t, 0, pw.loc.x)
	expectF(t, 0, pw.loc.y)
	pw.MoveTo(222, 333)
	expectF(t, 222, pw.loc.x)
	expectF(t, 792-333, pw.loc.y)
}

func TestPageWriter_PageHeight(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})
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

func TestPageWriter_SetFont(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)

	fc, err := NewTtfFontCollection("/Library/Fonts/*.ttf")
	if err != nil {
		t.Fatal(err)
	}
	dw.AddFontSource(fc, "TrueType")

	pw := dw.OpenPage()

	check(t, pw.Fonts() == nil, "Page font list should be empty by default")

	fonts, _ := pw.SetFont("Arial", 12, Options{})

	checkFatal(t, len(fonts) == 1, "length of fonts should be 1")
	expectS(t, "Arial", fonts[0].family)
	expectF(t, 12, fonts[0].size)
	check(t, fonts[0] == pw.Fonts()[0], "SetFont should return new font list")
	check(t, fonts[0] == dw.Fonts()[0], "SetFont changes to font list should be global")
}

func TestPageWriter_SetUnits(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})
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
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	check(t, pw.gw == nil, "GraphWriter should be nil by default.")
	check(t, !pw.inGraph, "Should not be in graph mode by default.")
	pw.last.loc = location{7, 7} // prove last.loc gets reset

	pw.startGraph()

	check(t, pw.inGraph, "Should be in graph mode now.")
	check(t, pw.last.loc.equal(location{0, 0}), "startGraph should reset location")
	check(t, pw.gw != nil, "GraphWriter should no longer be nil.")
}

func TestPageWriter_translate(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})
	loc := pw.translate(100, 200)
	expectF(t, 100, loc.x)
	expectF(t, 792-200, loc.y)
}

func TestPageWriter_units(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	// defaults to points
	pw1 := newPageWriter(dw, Options{})
	expectS(t, "pt", pw1.units.name)
	expectF(t, 1, pw1.units.ratio)
	// inches
	pw2 := newPageWriter(dw, Options{"units": "in"})
	expectS(t, "in", pw2.units.name)
	expectF(t, 72, pw2.units.ratio)
	// centimeters
	pw3 := newPageWriter(dw, Options{"units": "cm"})
	expectS(t, "cm", pw3.units.name)
	expectF(t, 28.35, pw3.units.ratio)
	// custom: "Dave points"
	UnitConversions.Add("dp", 0.072)
	pw4 := newPageWriter(dw, Options{"units": "dp"})
	expectS(t, "dp", pw4.units.name)
	expectF(t, 0.072, pw4.units.ratio)
}
