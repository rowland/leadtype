// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

type StdLabel struct {
	StdContainer
	textPieces []textPiece
	richText   *rich_text.RichText
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
	} else if last := &l.textPieces[len(l.textPieces)-1]; strings.HasSuffix(last.text, " ") && strings.HasPrefix(text, " ") {
		text = text[1:]
	}
	l.richText = nil
	if len(l.textPieces) > 0 && l.textPieces[len(l.textPieces)-1].font == font {
		l.textPieces[len(l.textPieces)-1].text += text
		return
	}
	l.textPieces = append(l.textPieces, textPiece{text, font})
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

func (l *StdLabel) DrawContent(w Writer) error {
	rt := l.RichText(w)
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
	rt := l.RichText(w)
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
	if l.richText != nil {
		return l.richText
	}
	l.richText = &rich_text.RichText{}
	for _, piece := range l.textPieces {
		piece.font.Apply(w)
		var err error
		l.richText, err = l.richText.Add(
			piece.text,
			w.Fonts(),
			w.FontSize(),
			options.Options{
				"color":     w.FontColor(),
				"strikeout": w.Strikeout(),
				"underline": w.Underline(),
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "StdLabel.RichText: %v", err)
		}
	}
	return l.richText
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
