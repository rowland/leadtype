// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"os"
	"os/exec"

	"github.com/rowland/leadtype/color"
	"github.com/rowland/leadtype/pdf"
)

const name = "test_008_rectangles.pdf"

func main() {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	doc := pdf.NewDocWriter()
	doc.NewPage()

	doc.SetLineColor(color.Blue)
	doc.SetLineWidth(3, "pt")
	doc.SetLineDashPattern("dashed")

	doc.SetUnits("in")
	doc.Rectangle(1, 1, 2, 3, true, false)

	doc.WriteTo(f)
	f.Close()
	exec.Command("open", name).Start()
}
