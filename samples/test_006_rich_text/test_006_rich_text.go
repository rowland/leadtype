// Copyright 2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"os"
	"os/exec"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/ttf_fonts"
)

const name = "test_006_rich_text.pdf"

type RT struct {
	s       string
	options options.Options
}

var line1 = []RT{
	{"Here is some ", options.Options{}},
	{"Bold ", options.Options{"weight": "bold"}},
	{"text.", options.Options{}},
}
var line2 = []RT{
	{"Here is some ", options.Options{}},
	{"Italic", options.Options{"style": "italic"}},
	{" text.", options.Options{}},
}
var line3 = []RT{
	{"Here is some ", options.Options{}},
	{"Bold, Italic", options.Options{"weight": "bold", "style": "italic"}},
	{" text.", options.Options{}},
}
var line4 = []RT{
	{"Here is some ", options.Options{}},
	{"Red", options.Options{"color": "red"}},
	{" text.", options.Options{}},
}
var line5 = []RT{
	{"Here is some ", options.Options{}},
	{"Underlined", options.Options{"underline": true}},
	{" text.", options.Options{}},
}
var line6 = []RT{
	{"This text is ", options.Options{}},
	{"Bold", options.Options{"weight": "bold"}},
	{". This text is ", options.Options{}},
	{"Underlined", options.Options{"underline": true}},
	{". This text is ", options.Options{}},
	{"Red", options.Options{"color": "Red"}},
	{". This text is normal.", options.Options{}},
}

func makeRtLine(doc *pdf.DocWriter, pieces []RT) *rich_text.RichText {
	rt := &rich_text.RichText{}
	for _, p := range pieces {
		fonts, err := doc.SetFont("Arial", 12, p.options)
		if err != nil {
			panic(err)
		}
		if len(fonts) == 0 {
			panic("No fonts returned.")
		}
		rt, err = rt.Add(p.s, fonts, 12, p.options)
	}
	return rt
}

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

	doc.NewPage()
	doc.MoveTo(1, 1)
	doc.SetFont("Courier New", 12, options.Options{})
	doc.SetUnderline(true)
	doc.Print("Rich Text\n\n")
	doc.SetUnderline(false)
	doc.SetLineSpacing(1.2)

	lines := []*rich_text.RichText{
		makeRtLine(doc, line1),
		makeRtLine(doc, line2),
		makeRtLine(doc, line3),
		makeRtLine(doc, line4),
		makeRtLine(doc, line5),
		makeRtLine(doc, line6),
	}
	doc.PrintParagraph(lines, options.Options{})
	doc.WriteTo(f)
	f.Close()
	exec.Command("open", name).Start()
}
