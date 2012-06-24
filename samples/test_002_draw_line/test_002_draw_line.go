// Copyright 2011-2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"../../pdf/_obj/go-pdf/pdf"
	"os"
	"exec"
)

const name = "test_002_draw_line.pdf"

func main() {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	doc := pdf.NewDocWriter(f)
	doc.Open(pdf.Options{})
	doc.OpenPage(pdf.Options{})

	doc.SetLineColor(pdf.Blue)
	doc.SetLineWidth(3, "pt")
	doc.SetLineDashPattern("dashed")

	doc.SetUnits("in")
	doc.MoveTo(1, 1)
	doc.LineTo(2, 2)

	doc.ClosePage()
	doc.Close()
	f.Close()
	exec.Command("open", name).Start()
}
