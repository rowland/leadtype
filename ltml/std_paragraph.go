// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
	"github.com/rowland/leadtype/wordbreaking"
)

type StdParagraph struct {
	StdContainer
	textPieces         []textPiece
	richText           *rich_text.RichText
	bullet             *BulletStyle
	splitEnabled       bool
	splitExplicit      bool
	orphans            int
	widows             int
	splitLines         []*rich_text.RichText
	suppressBullet     bool
	continuationIndent float64
}

func (p *StdParagraph) AddText(text string) {
	p.AddTextWithFont(text, p.Font())
}

func (p *StdParagraph) AddTextWithFont(text string, font *FontStyle) {
	text = normalizeXMLText(text)
	if text == "" {
		return
	}
	if len(p.textPieces) == 0 {
		text = strings.TrimLeftFunc(text, unicode.IsSpace)
		if text == "" {
			return
		}
	} else if last := &p.textPieces[len(p.textPieces)-1]; strings.HasSuffix(last.ResolvedText(nil), " ") && strings.HasPrefix(text, " ") {
		text = text[1:]
	}
	p.richText = nil
	if len(p.textPieces) > 0 && p.textPieces[len(p.textPieces)-1].font == font && !p.textPieces[len(p.textPieces)-1].Dynamic() {
		lastText := p.textPieces[len(p.textPieces)-1].ResolvedText(nil)
		p.textPieces[len(p.textPieces)-1].content = staticInlineText(lastText + text)
		return
	}
	p.textPieces = append(p.textPieces, newStaticTextPiece(text, font))
}

func (p *StdParagraph) AddInlineWithFont(content inlineText, font *FontStyle) {
	p.richText = nil
	p.textPieces = append(p.textPieces, textPiece{content: content, font: font})
}

func normalizeXMLText(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	inSpace := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			if !inSpace {
				b.WriteByte(' ')
				inSpace = true
			}
			continue
		}
		b.WriteRune(r)
		inSpace = false
	}
	return b.String()
}

func (p *StdParagraph) BeforePrint(w Writer) error {
	// fmt.Printf("Printing %s\n", p)
	return nil
}

func (p *StdParagraph) LayoutWidget(Writer) {
	// Paragraphs are text leaf widgets. Inline descendants contribute rich text
	// rather than participating in container child layout.
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
	para := p.Lines(w, p.lineWidth())
	if len(para) == 0 {
		return nil
	}
	indent := p.textIndent()
	x := ContentLeft(p)
	if p.suppressBullet {
		x += indent
	}
	w.MoveTo(x, ContentTop(p)+para[0].Ascent())
	if b := p.Bullet(); b != nil && !p.suppressBullet {
		x, y := w.Loc()
		b.Apply(w)
		w.Print(b.Text())
		w.MoveTo(x+b.Width(), y)
	}
	w.PrintParagraph(para, options.Options{
		"text-align": p.ParagraphStyle().textAlign.String(),
		"width":      ContentWidth(p) - indent,
	})
	return nil
}

func (p *StdParagraph) Lines(w Writer, width float64) []*rich_text.RichText {
	if p.splitLines != nil {
		return p.splitLines
	}
	rt := p.RichText(w)
	flags := make([]wordbreaking.Flags, rt.Len())
	wordbreaking.MarkRuneAttributes(rt.String(), flags)
	return rt.WrapToWidth(width, flags, false)
}

func (p *StdParagraph) PreferredHeight(w Writer) float64 {
	if p.height != 0 {
		return p.height
	}
	return p.heightForLines(p.Lines(w, p.lineWidth()), w)
}

func (p *StdParagraph) PreferredWidth(w Writer) float64 {
	if p.width != 0 {
		return p.width
	}
	return p.RichText(w).Width() + p.bulletWidth() + NonContentWidth(p) + 1
}

func (p *StdParagraph) RichText(w Writer) *rich_text.RichText {
	doc := documentForContainer(p)
	if p.richText != nil && !p.hasDynamicText() {
		return p.richText
	}
	rt := &rich_text.RichText{}
	lastText := ""
	for _, piece := range p.textPieces {
		font := piece.Font(p.Font())
		font.Apply(w)
		text := piece.ResolvedText(doc)
		if text == "" {
			continue
		}
		if strings.HasSuffix(lastText, " ") && strings.HasPrefix(text, " ") {
			text = text[1:]
		}
		if text == "" {
			continue
		}
		var err error
		rt, err = rt.Add(
			text,
			w.Fonts(),
			w.FontSize(),
			font.RichTextOptions(),
		)
		if err != nil {
			debugf("StdParagraph.RichText: %v", err)
		}
		lastText = text
	}
	if !p.hasDynamicText() {
		p.richText = rt
	}
	return rt
}

func (p *StdParagraph) hasDynamicText() bool {
	for _, piece := range p.textPieces {
		if piece.Dynamic() {
			return true
		}
	}
	return false
}

func (p *StdParagraph) SetAttrs(attrs map[string]string) {
	p.StdContainer.SetAttrs(attrs)
	p.splitEnabled = true
	p.orphans = 2
	p.widows = 2
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
	if split, ok := attrs["split"]; ok {
		p.splitExplicit = true
		p.splitEnabled = split != "false"
	}
	if orphans, ok := attrs["orphans"]; ok {
		if value, err := strconv.Atoi(orphans); err == nil {
			p.orphans = value
		}
	}
	if widows, ok := attrs["widows"]; ok {
		if value, err := strconv.Atoi(widows); err == nil {
			p.widows = value
		}
	}
}

func (p *StdParagraph) SplitForHeight(avail float64, w Writer) (*SplitResult, error) {
	if !p.splitEnabled {
		return nil, nil
	}
	lines := p.Lines(w, p.lineWidth())
	if len(lines) < 2 {
		return nil, nil
	}
	avail -= NonContentHeight(p)
	if avail <= 0 {
		return nil, nil
	}
	fit := 0
	for i := 1; i <= len(lines); i++ {
		if p.contentHeightForLines(lines[:i], w) <= avail {
			fit = i
			continue
		}
		break
	}
	if fit < p.orphanCount() || len(lines)-fit < p.widowCount() {
		return nil, nil
	}
	head := p.cloneForSplit(lines[:fit], p.suppressBullet, p.continuationIndent)
	tail := p.cloneForSplit(lines[fit:], true, p.textIndent())
	return &SplitResult{Head: head, Tail: tail}, nil
}

func (p *StdParagraph) cloneForSplit(lines []*rich_text.RichText, suppressBullet bool, continuationIndent float64) *StdParagraph {
	clone := *p
	clone.splitLines = append([]*rich_text.RichText(nil), lines...)
	clone.suppressBullet = suppressBullet
	clone.continuationIndent = continuationIndent
	clone.richText = nil
	clone.printed = false
	clone.invisible = false
	clone.disabled = false
	clone.path = ""
	return &clone
}

func (p *StdParagraph) textIndent() float64 {
	if p.suppressBullet {
		return p.continuationIndent
	}
	return p.bulletWidth()
}

func (p *StdParagraph) lineWidth() float64 {
	if p.Width() == 0 {
		return ContentWidth(p.container) - p.textIndent() - NonContentWidth(p)
	}
	return ContentWidth(p) - p.textIndent()
}

func (p *StdParagraph) heightForLines(lines []*rich_text.RichText, w Writer) float64 {
	return NonContentHeight(p) + p.contentHeightForLines(lines, w)
}

func (p *StdParagraph) contentHeightForLines(lines []*rich_text.RichText, w Writer) float64 {
	height := 0.0
	for _, line := range lines {
		height += line.Leading() * w.LineSpacing()
	}
	if len(lines) > 0 {
		height -= lines[len(lines)-1].Height() * (w.LineSpacing() - 1)
		height -= lines[len(lines)-1].LineGap()
	}
	return height
}

func (p *StdParagraph) orphanCount() int {
	if p.orphans < 1 {
		return 1
	}
	return p.orphans
}

func (p *StdParagraph) widowCount() int {
	if p.widows < 1 {
		return 1
	}
	return p.widows
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
