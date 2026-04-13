// Copyright 2014 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package main

import (
	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/pdf"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/ttf_fonts"
)

func init() {
	registerSample("test_006_rich_text", "demonstrate rich text styling", runTest006RichText)
}

type rtLinePiece struct {
	s       string
	options options.Options
}

var line1 = []rtLinePiece{
	{"Here is some ", options.Options{}},
	{"Bold ", options.Options{"weight": "bold"}},
	{"text.", options.Options{}},
}

var line2 = []rtLinePiece{
	{"Here is some ", options.Options{}},
	{"Italic", options.Options{"style": "italic"}},
	{" text.", options.Options{}},
}

var line3 = []rtLinePiece{
	{"Here is some ", options.Options{}},
	{"Bold, Italic", options.Options{"weight": "bold", "style": "italic"}},
	{" text.", options.Options{}},
}

var line4 = []rtLinePiece{
	{"Here is some ", options.Options{}},
	{"Red", options.Options{"color": "red"}},
	{" text.", options.Options{}},
}

var line5 = []rtLinePiece{
	{"Here is some ", options.Options{}},
	{"Underlined", options.Options{"underline": true}},
	{" text.", options.Options{}},
}

var line6 = []rtLinePiece{
	{"This text is ", options.Options{}},
	{"Bold", options.Options{"weight": "bold"}},
	{". This text is ", options.Options{}},
	{"Underlined", options.Options{"underline": true}},
	{". This text is ", options.Options{}},
	{"Red", options.Options{"color": "Red"}},
	{". This text is normal.", options.Options{}},
}

var line7 = []rtLinePiece{
	{"This text is ", options.Options{}},
	{"correct", options.Options{"underline": true, "underline_color": "green"}},
	{".", options.Options{}},
}

var line8 = []rtLinePiece{
	{"This text is ", options.Options{}},
	{"incorrect", options.Options{"strikeout": true, "strikeout_color": "red"}},
	{".", options.Options{}},
}

var line9 = []rtLinePiece{
	{"Round cap underline: ", options.Options{}},
	{"Sample", options.Options{
		"underline":           true,
		"underline_color":     "blue",
		"underline_cap":       "round_cap",
		"underline_thickness": 280,
	}},
}

var line10 = []rtLinePiece{
	{"Projecting square cap underline: ", options.Options{}},
	{"Sample", options.Options{
		"underline":           true,
		"underline_color":     "blue",
		"underline_cap":       "projecting_square_cap",
		"underline_thickness": 280,
	}},
}

var line11 = []rtLinePiece{
	{"Butt cap underline: ", options.Options{}},
	{"Sample", options.Options{
		"underline":           true,
		"underline_color":     "blue",
		"underline_cap":       "butt_cap",
		"underline_thickness": 280,
	}},
}

var line12 = []rtLinePiece{
	{"Thick underline: ", options.Options{}},
	{"Sample", options.Options{
		"underline":           true,
		"underline_color":     "purple",
		"underline_thickness": 500,
	}},
}

var line13 = []rtLinePiece{
	{"Thick strikeout: ", options.Options{}},
	{"Sample", options.Options{
		"strikeout":           true,
		"strikeout_color":     "orange",
		"strikeout_thickness": 500,
	}},
}

func runTest006RichText() (string, error) {
	return writeDoc("test_006_rich_text.pdf", func(doc *pdf.DocWriter) error {
		ttfc, err := ttf_fonts.NewFromSystemFonts()
		if err != nil {
			return err
		}
		doc.AddFontSource(ttfc)
		doc.SetUnits("in")

		doc.NewPage()
		doc.MoveTo(1, 1)
		if err := mustHaveFonts(doc.SetFont("Courier New", 12, options.Options{})); err != nil {
			return err
		}
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
			makeRtLine(doc, line7),
			makeRtLine(doc, line8),
			makeSizedRtLine(doc, line9, 24),
			makeSizedRtLine(doc, line10, 24),
			makeSizedRtLine(doc, line11, 24),
			makeSizedRtLine(doc, line12, 20),
			makeSizedRtLine(doc, line13, 20),
		}
		doc.PrintParagraph(lines, options.Options{})
		return nil
	})
}

func makeRtLine(doc *pdf.DocWriter, pieces []rtLinePiece) *rich_text.RichText {
	return makeSizedRtLine(doc, pieces, 12)
}

func makeSizedRtLine(doc *pdf.DocWriter, pieces []rtLinePiece, size float64) *rich_text.RichText {
	rt := &rich_text.RichText{}
	for _, p := range pieces {
		fonts, err := doc.SetFont("Arial", size, p.options)
		if err != nil {
			panic(err)
		}
		rt, err = rt.Add(p.s, fonts, size, p.options)
		if err != nil {
			panic(err)
		}
	}
	return rt
}
