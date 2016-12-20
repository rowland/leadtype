// Copyright 2012-2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/ttf_fonts"
)

const name = "test_005_ttf_fonts.pdf"

func main() {
	f, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	doc := pdf.NewDocWriter()
	ttfc, err := ttf_fonts.New("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	doc.AddFontSource(ttfc)

	doc.SetUnits("in")
	doc.SetLineSpacing(1.5)
	doc.MoveTo(1, 1)

	for _, info := range ttfc.FontInfos {
		if doc.Y() > 10 {
			doc.NewPage()
			doc.MoveTo(1, 1)
		}
		_, err = doc.SetFont(info.Family(), 12, options.Options{"style": info.Style()})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Setting font to %s/%s: %v\n", info.Family(), info.Style(), err)
			continue
		}
		fmt.Fprintln(doc, info.FullName())
	}
	doc.WriteTo(f)
	f.Close()
	exec.Command("open", name).Start()
}
