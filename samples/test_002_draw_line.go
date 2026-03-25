// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_002_draw_line", "draw a dashed blue line", runTest002DrawLine)
}

func runTest002DrawLine() (string, error) {
	return writeDoc("test_002_draw_line.pdf", func(doc *pdf.DocWriter) error {
		doc.NewPage()
		doc.SetLineColor(colors.Blue)
		doc.SetLineWidth(3, "pt")
		doc.SetLineDashPattern("dashed")
		doc.SetUnits("in")
		doc.MoveTo(1, 1)
		doc.LineTo(2, 2)
		return nil
	})
}
