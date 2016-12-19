// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"os"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/wordbreaking"
)

type textPiece struct {
	text string
	font *FontStyle
}

type StdParagraph struct {
	StdContainer
	textPieces []textPiece
	richText   *rich_text.RichText
}

func (p *StdParagraph) AddText(text string) {
	p.AddTextWithFont(text, p.Font())
}

func (p *StdParagraph) AddTextWithFont(text string, font *FontStyle) {
	p.richText = nil
	p.textPieces = append(p.textPieces, textPiece{text, font})
}

func (p *StdParagraph) BeforePrint(w Writer) error {
	fmt.Printf("Printing %s\n", p)
	return nil
}

func (p *StdParagraph) bulletWidth() float64 {
	return 0 // TODO
}

func (p *StdParagraph) DrawContent(w Writer) error {
	fmt.Println(p.RichText(w))
	para := p.Lines(w, p.Width())
	if len(para) == 0 {
		return nil
	}
	w.MoveTo(ContentLeft(p), ContentTop(p)+para[0].Ascent())
	w.PrintParagraph(para)
	return nil
}

func (p *StdParagraph) Lines(w Writer, width float64) []*rich_text.RichText {
	rt := p.RichText(w)
	flags := make([]wordbreaking.Flags, rt.Len())
	wordbreaking.MarkRuneAttributes(rt.String(), flags)
	return rt.WrapToWidth(width, flags, false)
}

func (p *StdParagraph) PreferredHeight(w Writer) float64 {
	if p.height != 0 {
		return p.height
	}
	var width float64
	if p.Width() == 0 {
		width = ContentWidth(p.container) - p.bulletWidth() - NonContentWidth(p)
	} else {
		width = ContentWidth(p) - p.bulletWidth()
	}

	para := p.Lines(w, width)

	height := 0.0
	for _, line := range para {
		height += line.Height()
	}
	return height
}

func (p *StdParagraph) PreferredWidth(w Writer) float64 {
	if p.width != 0 {
		return p.width
	}
	return p.RichText(w).Width() + p.bulletWidth() + NonContentWidth(p) + 1
}

func (p *StdParagraph) RichText(w Writer) *rich_text.RichText {
	if p.richText != nil {
		return p.richText
	}
	p.richText = &rich_text.RichText{}
	for _, piece := range p.textPieces {
		piece.font.Apply(w)
		var err error
		p.richText, err = p.richText.Add(
			piece.text,
			w.Fonts(),
			w.FontSize(), options.Options{
				"color":        w.FontColor(),
				"line_through": w.LineThrough(),
				"underline":    w.Underline()})
		if err != nil {
			fmt.Fprintf(os.Stderr, "StdParagraph.RichText: %v", err)
		}
	}
	return p.richText
}

func (p *StdParagraph) SetAttrs(attrs map[string]string) {
	p.StdContainer.SetAttrs(attrs)
}

func (p *StdParagraph) String() string {
	return fmt.Sprintf("StdParagraph %s", &p.StdContainer)
}

func init() {
	registerTag(DefaultSpace, "p", func() interface{} { return &StdParagraph{} })
}

var _ Container = (*StdParagraph)(nil)
var _ HasAttrs = (*StdParagraph)(nil)
var _ Identifier = (*StdParagraph)(nil)
var _ Printer = (*StdParagraph)(nil)
var _ WantsContainer = (*StdParagraph)(nil)
