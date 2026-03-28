// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"

	"github.com/rowland/leadtype/options"
	"github.com/rowland/leadtype/rich_text"
)

const preTabWidth = 4

type StdPre struct {
	StdWidget
	rawText string
	lines   []string
}

func (p *StdPre) AddText(text string) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	p.rawText += text
	p.lines = nil
}

func (p *StdPre) DrawContent(w Writer) error {
	lines := p.Lines()
	if len(lines) == 0 {
		lines = []string{""}
	}

	p.Font().Apply(w)
	lineHeight := p.lineHeight(w)
	y := ContentTop(p)
	for _, line := range lines {
		if line != "" {
			rt, err := p.richTextForLine(line, w)
			if err != nil {
				return err
			}
			w.MoveTo(ContentLeft(p), y+rt.Ascent())
			w.PrintRichText(rt)
		}
		y += lineHeight
	}
	return nil
}

func (p *StdPre) Font() *FontStyle {
	if p.font != nil {
		return p.font
	}
	if p.scope != nil {
		if fixed := FontStyleFor("fixed", p.scope); fixed != nil {
			return fixed
		}
	}
	if fixed, ok := defaultStyles["fixed"].(*FontStyle); ok {
		return fixed
	}
	return defaultFont
}

func (p *StdPre) PreferredHeight(w Writer) float64 {
	if p.height != 0 {
		return p.height
	}
	lines := p.Lines()
	if len(lines) == 0 {
		lines = []string{""}
	}
	return float64(len(lines))*p.lineHeight(w) + NonContentHeight(p)
}

func (p *StdPre) PreferredWidth(w Writer) float64 {
	if p.width != 0 {
		return p.width
	}
	p.Font().Apply(w)
	maxWidth := 0.0
	for _, line := range p.Lines() {
		rt, err := p.richTextForLine(line, w)
		if err != nil {
			continue
		}
		if width := rt.Width(); width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth + NonContentWidth(p)
}

func (p *StdPre) SetAttrs(attrs map[string]string) {
	p.StdWidget.SetAttrs(attrs)
}

func (p *StdPre) Lines() []string {
	if p.lines != nil {
		return p.lines
	}
	text := strings.ReplaceAll(p.rawText, "\t", strings.Repeat(" ", preTabWidth))
	if text == "" {
		p.lines = []string{""}
		return p.lines
	}

	lines := strings.Split(text, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		p.lines = []string{""}
		return p.lines
	}

	indent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		count := 0
		for count < len(line) && line[count] == ' ' {
			count++
		}
		if indent == -1 || count < indent {
			indent = count
		}
	}
	if indent > 0 {
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				lines[i] = ""
				continue
			}
			lines[i] = line[indent:]
		}
	}
	p.lines = lines
	return p.lines
}

func (p *StdPre) String() string {
	return fmt.Sprintf("StdPre %s", &p.StdWidget)
}

func (p *StdPre) lineHeight(w Writer) float64 {
	rt, err := p.richTextForLine("M", w)
	if err != nil {
		return p.Font().size * w.LineSpacing()
	}
	return rt.Leading() * w.LineSpacing()
}

func (p *StdPre) richTextForLine(line string, w Writer) (*rich_text.RichText, error) {
	p.Font().Apply(w)
	rt, err := rich_text.New(line, w.Fonts(), w.FontSize(), options.Options{
		"color":     w.FontColor(),
		"strikeout": w.Strikeout(),
		"underline": w.Underline(),
	})
	if err != nil {
		debugf("StdPre.richTextForLine: %v", err)
		return nil, err
	}
	return rt, nil
}

func init() {
	registerTag(DefaultSpace, "pre", func() interface{} { return &StdPre{} })
}

var _ HasAttrs = (*StdPre)(nil)
var _ HasText = (*StdPre)(nil)
var _ Identifier = (*StdPre)(nil)
var _ Printer = (*StdPre)(nil)
var _ WantsContainer = (*StdPre)(nil)
