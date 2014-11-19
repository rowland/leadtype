// Copyright 2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"leadtype/pdf"
	"os"
	"os/exec"
)

const name = "test_006_rich_text.pdf"

type RT struct {
	s       string
	options pdf.Options
}

var line1 = []RT{
	{"Here is some ", pdf.Options{}},
	{"Bold ", pdf.Options{"weight": "bold"}},
	{"text.", pdf.Options{}},
}
var line2 = []RT{
	{"Here is some ", pdf.Options{}},
	{"Italic", pdf.Options{"style": "italic"}},
	{" text.", pdf.Options{}},
}
var line3 = []RT{
	{"Here is some ", pdf.Options{}},
	{"Bold, Italic", pdf.Options{"weight": "bold", "style": "italic"}},
	{" text.", pdf.Options{}},
}
var line4 = []RT{
	{"Here is some ", pdf.Options{}},
	{"Red", pdf.Options{"color": "red"}},
	{" text.", pdf.Options{}},
}
var line5 = []RT{
	{"Here is some ", pdf.Options{}},
	{"Underlined", pdf.Options{"underline": true}},
	{" text.", pdf.Options{}},
}
var line6 = []RT{
	{"This text is ", pdf.Options{}},
	{"Bold", pdf.Options{"weight": "bold"}},
	{". This text is ", pdf.Options{}},
	{"Underlined", pdf.Options{"underline": true}},
	{". This text is ", pdf.Options{}},
	{"Red", pdf.Options{"color": "Red"}},
	{". This text is normal.", pdf.Options{}},
}

func makeRtLine(doc *pdf.DocWriter, pieces []RT) *pdf.RichText {
	rt := &pdf.RichText{}
	for _, p := range pieces {
		fonts, err := doc.SetFont("Arial", 12, "TrueType", p.options)
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
	ttfc, err := pdf.NewTtfFontCollection("/Library/Fonts/*.ttf")
	if err != nil {
		panic(err)
	}
	doc.AddFontSource(ttfc, "TrueType")
	doc.SetUnits("in")

	doc.NewPage()
	doc.MoveTo(1, 1)
	doc.SetFont("Courier New", 12, "TrueType", pdf.Options{})
	doc.SetUnderline(true)
	doc.Print("Rich Text\n\n")
	doc.SetUnderline(false)

	lines := []*pdf.RichText{
		makeRtLine(doc, line1),
		makeRtLine(doc, line2),
		makeRtLine(doc, line3),
		makeRtLine(doc, line4),
		makeRtLine(doc, line5),
		makeRtLine(doc, line6),
	}
	doc.PrintParagraph(lines)
	doc.WriteTo(f)
	f.Close()
	exec.Command("open", name).Start()
}
