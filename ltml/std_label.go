// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

type StdLabel struct {
	StdContainer
	textPieces  []textPiece
	richText    *rich_text.RichText
	shrinkToFit bool
}

func (l *StdLabel) AddText(text string) {
	l.AddTextWithFont(text, l.Font())
}

func (l *StdLabel) AddTextWithFont(text string, font *FontStyle) {
	text = normalizeLabelXMLText(text)
	if text == "" {
		return
	}
	if len(l.textPieces) == 0 {
		text = strings.TrimLeftFunc(text, unicode.IsSpace)
		if text == "" {
			return
		}
	} else if last := &l.textPieces[len(l.textPieces)-1]; strings.HasSuffix(last.ResolvedText(nil), " ") && strings.HasPrefix(text, " ") {
		text = text[1:]
	}
	l.richText = nil
	if len(l.textPieces) > 0 && l.textPieces[len(l.textPieces)-1].font == font && !l.textPieces[len(l.textPieces)-1].Dynamic() {
		lastText := l.textPieces[len(l.textPieces)-1].ResolvedText(nil)
		l.textPieces[len(l.textPieces)-1].content = staticInlineText(lastText + text)
		return
	}
	l.textPieces = append(l.textPieces, newStaticTextPiece(text, font))
}

func (l *StdLabel) AddInlineWithFont(content inlineText, font *FontStyle) {
	l.richText = nil
	l.textPieces = append(l.textPieces, textPiece{content: content, font: font})
}

func normalizeLabelXMLText(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	inSpace := false
	for _, r := range text {
		if r == '\n' || r == '\r' {
			continue
		}
		if r == '\t' {
			r = ' '
		}
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

func (l *StdLabel) LayoutWidget(Writer) {
	// Labels are leaf widgets even though they embed StdContainer to collect
	// inline children like <span> and <pageno>.
}

func (l *StdLabel) DrawContent(w Writer) error {
	rt := l.fittedRichText(w)
	if rt.Len() == 0 {
		return nil
	}
	l.Font().Apply(w)
	w.MoveTo(ContentLeft(l), ContentTop(l)+rt.Ascent())
	w.PrintRichText(rt)
	return nil
}

func (l *StdLabel) PreferredHeight(w Writer) float64 {
	if l.height != 0 {
		return l.height
	}
	rt := l.fittedRichText(w)
	if rt.Len() == 0 {
		return l.Font().size*w.LineSpacing() + NonContentHeight(l)
	}
	return rt.Leading()*w.LineSpacing() + NonContentHeight(l)
}

func (l *StdLabel) PreferredWidth(w Writer) float64 {
	if l.width != 0 {
		return l.width
	}
	rt := l.RichText(w)
	return rt.Width() + NonContentWidth(l)
}

func (l *StdLabel) RichText(w Writer) *rich_text.RichText {
	doc := documentForContainer(l)
	if l.richText != nil && !l.hasDynamicText() {
		return l.richText
	}
	rt := &rich_text.RichText{}
	lastText := ""
	for _, piece := range l.textPieces {
		font := piece.Font(l.Font())
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
			options.Options{
				"color":     w.FontColor(),
				"strikeout": w.Strikeout(),
				"underline": w.Underline(),
			},
		)
		if err != nil {
			debugf("StdLabel.RichText: %v", err)
		}
		lastText = text
	}
	if !l.hasDynamicText() {
		l.richText = rt
	}
	return rt
}

func (l *StdLabel) SetAttrs(attrs map[string]string) {
	l.StdContainer.SetAttrs(attrs)
	l.shrinkToFit = attrs["fit"] == "shrink"
}

func (l *StdLabel) fittedRichText(w Writer) *rich_text.RichText {
	rt := l.RichText(w)
	if rt == nil || rt.Len() == 0 || !l.shrinkToFit || l.width == 0 {
		return rt
	}
	availableWidth := ContentWidth(l)
	if availableWidth <= 0 {
		return rt
	}
	measuredWidth := rt.Width()
	if measuredWidth <= 0 || measuredWidth <= availableWidth {
		return rt
	}
	return rt.Scale(availableWidth/measuredWidth, 6.0)
}

func (l *StdLabel) hasDynamicText() bool {
	for _, piece := range l.textPieces {
		if piece.Dynamic() {
			return true
		}
	}
	return false
}

func (l *StdLabel) String() string {
	return fmt.Sprintf("StdLabel %s", &l.StdContainer)
}

func init() {
	registerTag(DefaultSpace, "label", func() interface{} { return &StdLabel{} })
}

var _ Container = (*StdLabel)(nil)
var _ HasAttrs = (*StdLabel)(nil)
var _ HasText = (*StdLabel)(nil)
var _ Identifier = (*StdLabel)(nil)
var _ Printer = (*StdLabel)(nil)
var _ WantsContainer = (*StdLabel)(nil)
