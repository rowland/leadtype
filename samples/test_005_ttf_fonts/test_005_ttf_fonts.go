// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"leadtype/pdf"
	"os"
	"os/exec"
)

const name = "test_005_ttf_fonts.pdf"

func main() {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	doc := pdf.NewDocWriter()
	ttfc, err := pdf.NewTtfFontCollection("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	doc.AddFontSource(ttfc)

	doc.SetUnits("in")

	for i, info := range ttfc.FontInfos {
		offset := i % 20
		if offset == 0 {
			doc.NewPage()
		}
		doc.MoveTo(1, 1+float64(offset)*0.5)
		_, err = doc.SetFont(info.Family(), 12, pdf.Options{"style": info.Style()})
		if err != nil {
			panic(err)
		}
		doc.Print(info.FullName())
	}
	doc.WriteTo(f)
	f.Close()
	exec.Command("open", name).Start()
}
