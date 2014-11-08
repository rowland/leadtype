// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"leadtype/pdf"
	"os"
	"os/exec"
)

const name = "test_003_hello_world.pdf"

func main() {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	doc := pdf.NewDocWriter(f)
	afmfc, err := pdf.NewAfmFontCollection("../../afm/data/fonts/*.afm")
	if err != nil {
		panic(err)
	}
	doc.AddFontSource(afmfc, "Type1")
	ttfc, err := pdf.NewTtfFontCollection("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	doc.AddFontSource(ttfc, "TrueType")

	doc.Open()
	doc.OpenPage()
	doc.SetUnits("in")

	_, err = doc.SetFont("Helvetica", 12, pdf.Options{"sub_type": "Type1"})
	if err != nil {
		panic(err)
	}
	doc.MoveTo(1, 1)
	doc.Print("Hello, World!")

	_, err = doc.SetFont("Arial", 14, pdf.Options{"sub_type": "TrueType"})
	if err != nil {
		panic(err)
	}
	doc.MoveTo(2, 2)
	doc.Print("Goodbye, cruel world...")

	doc.ClosePage()
	doc.Close()
	f.Close()
	exec.Command("open", name).Start()
}
