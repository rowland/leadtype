// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"testing"
)

func TestMiscWriter_setCmykColorFill(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setCmykColorFill(0.5, 0.5, 0.5, 0.5)
	expectS(t, "0.5 0.5 0.5 0.5 k\n", buf.String())
}

func TestMiscWriter_setCmykColorStroke(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setCmykColorStroke(0.5, 0.5, 0.5, 0.5)
	expectS(t, "0.5 0.5 0.5 0.5 K\n", buf.String())
}

func TestMiscWriter_setColorFill(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setColorFill([]float64{0.1, 0.2, 0.3, 0.4})
	expectS(t, "0.1 0.2 0.3 0.4 sc\n", buf.String())
}

// TODO: scn, SCN: patterns and separations
func TestMiscWriter_setColorRenderingIntent(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setColorRenderingIntent("RelativeColorimetric")
	expectS(t, "/RelativeColorimetric ri\n", buf.String())
}

func TestMiscWriter_setColorSpaceFill(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setColorSpaceFill("DeviceGray")
	expectS(t, "/DeviceGray cs\n", buf.String())
}

func TestMiscWriter_setColorSpaceStroke(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setColorSpaceStroke("DeviceGray")
	expectS(t, "/DeviceGray CS\n", buf.String())
}

func TestMiscWriter_setColorStroke(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setColorStroke([]float64{0.1, 0.2, 0.3, 0.4})
	expectS(t, "0.1 0.2 0.3 0.4 SC\n", buf.String())
}

func TestMiscWriter_setGrayFill(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setGrayFill(0.4)
	expectS(t, "0.4 g\n", buf.String())
}

func TestMiscWriter_setGrayStroke(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setGrayStroke(0.9)
	expectS(t, "0.9 G\n", buf.String())
}

func TestMiscWriter_setRgbColorFill(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setRgbColorFill(0.3, 0.6, 0.9)
	expectS(t, "0.3 0.6 0.9 rg\n", buf.String())
}

func TestMiscWriter_setRgbColorStroke(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.setRgbColorStroke(0.3, 0.6, 0.9)
	expectS(t, "0.3 0.6 0.9 RG\n", buf.String())
}

func TestMiscWriter_xObject(t *testing.T) {
	var buf bytes.Buffer
	mw := newMiscWriter(&buf)
	mw.xObject("Image1")
	expectS(t, "/Image1 Do\n", buf.String())
}
