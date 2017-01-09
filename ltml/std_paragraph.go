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
	bullet     *BulletStyle
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

func (p *StdParagraph) Bullet() *BulletStyle {
	if p.bullet != nil {
		return p.bullet
	}
	if ps := p.ParagraphStyle(); ps != nil {
		return ps.Bullet()
	}
	return nil
}

func (p *StdParagraph) bulletWidth() float64 {
	if b := p.Bullet(); b != nil {
		return b.Width()
	}
	return 0
}

func (p *StdParagraph) DrawContent(w Writer) error {
	fmt.Println(p.RichText(w))
	para := p.Lines(w, ContentWidth(p)-p.bulletWidth())
	if len(para) == 0 {
		return nil
	}
	w.MoveTo(ContentLeft(p), ContentTop(p)+para[0].Ascent())
	if b := p.Bullet(); b != nil {
		x, y := w.Loc()
		b.Apply(w)
		w.Print(b.Text())
		w.MoveTo(x+b.Width(), y)
	}
	w.PrintParagraph(para, options.Options{
		"text-align": p.ParagraphStyle().textAlign.String(),
		"width":      ContentWidth(p),
	})
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

	height := NonContentHeight(p)
	for _, line := range para {
		height += line.Leading() * w.LineSpacing()
	}
	if len(para) > 0 {
		height -= para[len(para)-1].Height() * (w.LineSpacing() - 1)
		height -= para[len(para)-1].LineGap()
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
				"color":     w.FontColor(),
				"strikeout": w.Strikeout(),
				"underline": w.Underline()})
		if err != nil {
			fmt.Fprintf(os.Stderr, "StdParagraph.RichText: %v", err)
		}
	}
	return p.richText
}

func (p *StdParagraph) SetAttrs(attrs map[string]string) {
	p.StdContainer.SetAttrs(attrs)
	if style, ok := attrs["style"]; ok {
		p.paragraphStyle = ParagraphStyleFor(style, p.scope)
	}
	if MapHasKeyPrefix(attrs, "style.") {
		p.paragraphStyle = p.ParagraphStyle().Clone()
		p.paragraphStyle.SetAttrs("style.", attrs)
	}
	if bullet, ok := attrs["bullet"]; ok {
		p.bullet = BulletStyleFor(bullet, p.scope)
	}
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
