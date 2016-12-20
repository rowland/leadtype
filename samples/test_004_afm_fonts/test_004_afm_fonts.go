// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rowland/leadtype/afm_fonts"
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
)

const name = "test_004_afm_fonts.pdf"

func main() {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	doc := pdf.NewDocWriter()
	afmfc, err := afm_fonts.New("../../afm/data/fonts/*.afm")
	if err != nil {
		panic(err)
	}
	doc.AddFontSource(afmfc)

	doc.SetUnits("in")
	doc.SetLineSpacing(1.5)
	doc.MoveTo(1, 1)

	for _, info := range afmfc.FontInfos {
		if doc.Y() > 10 {
			doc.NewPage()
			doc.MoveTo(1, 1)
		}
		_, err = doc.SetFont(info.PostScriptName(), 12, options.Options{})
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(doc, info.FullName())
	}
	doc.WriteTo(f)
	f.Close()
	exec.Command("open", name).Start()
}
