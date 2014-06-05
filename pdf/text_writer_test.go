// Copyright 2014 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"bytes"
	"testing"
)

func TestTextWriter_open(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.open()
	expectS(t, "BT\n", buf.String())
}

func TestTextWriter_close(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.close()
	expectS(t, "ET\n", buf.String())
}

func TestTextWriter_setCharSpacing(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.setCharSpacing(5)
	expectS(t, "5 Tc\n", buf.String())
}

func TestTextWriter_setWordSpacing(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.setWordSpacing(5)
	expectS(t, "5 Tw\n", buf.String())
}

func TestTextWriter_setHorizScaling(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.setHorizScaling(90)
	expectS(t, "90 Tz\n", buf.String())
}

func TestTextWriter_setLeading(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.setLeading(8)
	expectS(t, "8 TL\n", buf.String())
}

func TestTextWriter_setFontAndSize(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.setFontAndSize("Arial", 12)
	expectS(t, "/Arial 12 Tf\n", buf.String())
}

func TestTextWriter_setRenderingMode(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.setRenderingMode(0)
	expectS(t, "0 Tr\n", buf.String())
}

func TestTextWriter_setRise(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.setRise(0)
	expectS(t, "0 Ts\n", buf.String())
}

func TestTextWriter_moveBy(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.moveBy(7, 11.5)
	expectS(t, "7 11.5 Td\n", buf.String())
}

func TestTextWriter_moveByAndSetLeading(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.moveByAndSetLeading(5.5, 12)
	expectS(t, "5.5 12 TD\n", buf.String())
}

func TestTextWriter_setMatrix(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.setMatrix(1.1, 2, 3.3, 4, 5.5, 6)
	expectS(t, "1.1 2 3.3 4 5.5 6 Tm\n", buf.String())
}

func TestTextWriter_nextLine(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.nextLine()
	expectS(t, "T*\n", buf.String())
}

func TestTextWriter_show(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.show([]byte("Hello"))
	expectS(t, "(Hello) Tj\n", buf.String())
}

func TestTextWriter_nextLineShow(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.nextLineShow([]byte("Goodbye"))
	expectS(t, "(Goodbye) '", buf.String())
}

func TestTextWriter_setSpacingNextLineShow(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	tw.setSpacingNextLineShow(5, 11.5, "Hello and goodbye")
	expectS(t, "5 11.5 (Hello and goodbye) \"\n", buf.String())
}

func TestTextWriter_showWithDispacements(t *testing.T) {
	var buf bytes.Buffer
	tw := newTextWriter(&buf)
	a := array{
		str("H"),
		integer(120),
		str("e"),
		integer(80),
		str("y"),
	}
	tw.showWithDispacements(a)
	expectS(t, "[(H) 120 (e) 80 (y) ] TJ\n", buf.String())
}
