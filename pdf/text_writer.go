// Copyright 2014 Brent Rowland.
// Use of this source code is governed by the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
	"io"
)

type textWriter struct {
	wr io.Writer
}

func newTextWriter(wr io.Writer) *textWriter {
	return &textWriter{wr}
}

func (tw *textWriter) open() {
	fmt.Fprintf(tw.wr, "BT\n")
}

func (tw *textWriter) close() {
	fmt.Fprintf(tw.wr, "ET\n")
}

func (tw *textWriter) setCharSpacing(charSpace float64) {
	fmt.Fprintf(tw.wr, "%s Tc\n", g(charSpace))
}

func (tw *textWriter) setWordSpacing(wordSpace float64) {
	fmt.Fprintf(tw.wr, "%s Tw\n", g(wordSpace))
}

func (tw *textWriter) setHorizScaling(scale float64) {
	fmt.Fprintf(tw.wr, "%s Tz\n", g(scale))
}

func (tw *textWriter) setLeading(leading float64) {
	fmt.Fprintf(tw.wr, "%s TL\n", g(leading))
}

func (tw *textWriter) setFontAndSize(fontName string, size float64) {
	fmt.Fprintf(tw.wr, "/%s %s Tf\n", fontName, g(size))
}

func (tw *textWriter) setRenderingMode(render int) {
	fmt.Fprintf(tw.wr, "%d Tr\n", render)
}

func (tw *textWriter) setRise(rise float64) {
	fmt.Fprintf(tw.wr, "%s Ts\n", g(rise))
}

func (tw *textWriter) moveBy(tx, ty float64) {
	fmt.Fprintf(tw.wr, "%s %s Td\n", g(tx), g(ty))
}

func (tw *textWriter) moveByAndSetLeading(tx, ty float64) {
	fmt.Fprintf(tw.wr, "%s %s TD\n", g(tx), g(ty))
}

func (tw *textWriter) setMatrix(a, b, c, d, x, y float64) {
	fmt.Fprintf(tw.wr, "%s %s %s %s %s %s Tm\n", g(a), g(b), g(c), g(d), g(x), g(y))
}

func (tw *textWriter) nextLine() {
	fmt.Fprintf(tw.wr, "T*\n")
}

func (tw *textWriter) show(s []byte) {
	fmt.Fprintf(tw.wr, "(%s) Tj\n", str(s).escape())
}

func (tw *textWriter) nextLineShow(s []byte) {
	fmt.Fprintf(tw.wr, "(%s) '", str(s).escape())
}

func (tw *textWriter) setSpacingNextLineShow(charSpace, wordSpace float64, s string) {
	fmt.Fprintf(tw.wr, "%s %s (%s) \"\n", g(charSpace), g(wordSpace), str(s).escape())
}

func (tw *textWriter) showWithDispacements(elements array) {
	elements.write(tw.wr)
	fmt.Fprint(tw.wr, "TJ\n")
}
