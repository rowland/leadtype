// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"github.com/rowland/leadtype/colors"
	"github.com/rowland/leadtype/pdf"
)

func init() {
	registerSample("test_008_rectangles", "draw a dashed blue rectangle", runTest008Rectangles)
}

func runTest008Rectangles() (string, error) {
	return writeDoc("test_008_rectangles.pdf", func(doc *pdf.DocWriter) error {
		doc.NewPage()
		doc.SetLineColor(colors.Blue)
		doc.SetLineWidth(3, "pt")
		doc.SetLineDashPattern("dashed")
		doc.SetUnits("in")
		doc.Rectangle(1, 1, 2, 3, true, false)
		return nil
	})
}
