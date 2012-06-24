// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package pdf

import (
	"fmt"
	"io"
)

type miscWriter struct {
	wr io.Writer
}

func newMiscWriter(wr io.Writer) *miscWriter {
	return &miscWriter{wr}
}

func (mw *miscWriter) setCmykColorFill(c, m, y, k float64) {
	fmt.Fprintf(mw.wr, "%s %s %s %s k\n", g(c), g(m), g(y), g(k))
}

func (mw *miscWriter) setCmykColorStroke(c, m, y, k float64) {
	fmt.Fprintf(mw.wr, "%s %s %s %s K\n", g(c), g(m), g(y), g(k))
}

func (mw *miscWriter) setColorSpaceFill(name string) {
	fmt.Fprintf(mw.wr, "/%s cs\n", name)
}

func (mw *miscWriter) setColorSpaceStroke(name string) {
	fmt.Fprintf(mw.wr, "/%s CS\n", name)
}

func (mw *miscWriter) setColorFill(colors []float64) {
	fmt.Fprintf(mw.wr, "%s sc\n", float64Slice(colors).join(" "))
}

func (mw *miscWriter) setColorStroke(colors []float64) {
	fmt.Fprintf(mw.wr, "%s SC\n", float64Slice(colors).join(" "))
}

// TODO: scn, SCN: patterns and separations
func (mw *miscWriter) setColorRenderingIntent(intent string) {
	fmt.Fprintf(mw.wr, "/%s ri\n", intent)
}

func (mw *miscWriter) setGrayFill(gray float64) {
	fmt.Fprintf(mw.wr, "%s g\n", g(gray))
}

func (mw *miscWriter) setGrayStroke(gray float64) {
	fmt.Fprintf(mw.wr, "%s G\n", g(gray))
}

func (mw *miscWriter) setRgbColorFill(red, green, blue float64) {
	fmt.Fprintf(mw.wr, "%s %s %s rg\n", g(red), g(green), g(blue))
}

func (mw *miscWriter) setRgbColorStroke(red, green, blue float64) {
	fmt.Fprintf(mw.wr, "%s %s %s RG\n", g(red), g(green), g(blue))
}

func (mw *miscWriter) xObject(name string) {
	fmt.Fprintf(mw.wr, "/%s Do\n", name)
}
