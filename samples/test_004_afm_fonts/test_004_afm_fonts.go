// Copyright 2012 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"leadtype/pdf"
	"os"
	"os/exec"
)

const name = "test_004_afm_fonts.pdf"

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

	doc.Open()
	doc.OpenPage()
	doc.SetUnits("in")

	for i, info := range afmfc.FontInfos {
		offset := i % 20
		if offset == 0 && i > 0 {
			doc.OpenPage()
		}
		doc.MoveTo(1, 1+float64(offset)*0.5)
		fmt.Println("<" + info.PostScriptName() + "><" + info.Weight() + ">")
		_, err = doc.SetFont(info.PostScriptName(), 12, pdf.Options{"sub_type": "Type1"})
		if err != nil {
			panic(err)
		}
		doc.Print(info.FullName())
	}
	doc.Close()
	f.Close()
	exec.Command("open", name).Start()
}
