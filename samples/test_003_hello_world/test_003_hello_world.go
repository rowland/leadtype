// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"os"
	"os/exec"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

const name = "test_003_hello_world.pdf"

func main() {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	doc := pdf.NewDocWriter()
	afmfc, err := pdf.NewAfmFontCollection("../../afm/data/fonts/*.afm")
	if err != nil {
		panic(err)
	}
	doc.AddFontSource(afmfc)
	ttfc, err := pdf.NewTtfFontCollection("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	doc.AddFontSource(ttfc)

	doc.NewPage()
	doc.SetUnits("in")

	_, err = doc.SetFont("Helvetica", 12, options.Options{})
	if err != nil {
		panic(err)
	}
	doc.MoveTo(1, 1)
	doc.Print("Hello, World!")

	_, err = doc.SetFont("Arial", 14, options.Options{})
	if err != nil {
		panic(err)
	}
	doc.MoveTo(2, 2)
	doc.Print("Goodbye, cruel world...")

	doc.WriteTo(f)
	f.Close()
	exec.Command("open", name).Start()
}
