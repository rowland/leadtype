package pdf

import (
	"testing"
	"bytes"
)

const (
	black = 0
	red = 0xFF0000
	green = 0x00FF00
)

func TestPageWriter_checkSetLineColor(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})
	check(t, pw.lineColor == black, "lineColor should default to black")
	check(t, pw.lastLineColor == black, "lastLineColor should default to black")
	pw.SetLineColor(red)
	check(t, pw.lineColor == red, "lineColor should now be red")
	check(t, pw.lastLineColor == black, "lastLineColor should still be black")
	pw.checkSetLineColor()
	check(t, pw.lastLineColor == red, "lastLineColor should now be red")
	// TODO: test for autoPath behavior
}

func TestPageWriter_LineColor(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})

	check(t, pw.LineColor() == black, "Should default to black") // defaults to black

	last := pw.SetLineColor(red) // previous color was black
	check(t, pw.LineColor() == red, "Should have changed to red")
	check(t, last == black, "Previous color was black")

	last = pw.SetLineColor(green)
	check(t, last == red, "Previous color was red")
	check(t, pw.LineColor() == green, "New color is green")
}

func TestPageWriter_MoveTo(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})
	expectF(t, 0, pw.loc.x)
	expectF(t, 0, pw.loc.y)
	pw.MoveTo(222, 333)
	expectF(t, 222, pw.loc.x)
	expectF(t, 792 - 333, pw.loc.y)
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
	pw.lastLoc = location{7, 7} // prove lastLoc gets reset

	pw.startGraph()

	check(t, pw.inGraph, "Should be in graph mode now.")
	check(t, pw.lastLoc.equal(location{0, 0}), "startGraph should reset location")
	check(t, pw.gw != nil, "GraphWriter should no longer be nil.")
}

func TestPageWriter_translate(t *testing.T) {
	var buf bytes.Buffer
	dw := NewDocWriter(&buf)
	pw := newPageWriter(dw, Options{})
	loc := pw.translate(100, 200)
	expectF(t, 100, loc.x)
	expectF(t, 792 - 200, loc.y)
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
