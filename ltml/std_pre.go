// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
	"strings"

	"github.com/rowland/leadtype/rich_text"
)

const preTabWidth = 4

type StdPre struct {
	StdWidget
	source  textSource
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
	lines, err := p.resolvedLines()
	if err != nil {
		return err
	}
	return withWidgetRoleAccessibility(w, &p.StdWidget, "", p.AccessibilityText(), func() error {
		if len(lines) == 0 {
			lines = []string{""}
		}

		applyWidgetFont(w, p)
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
	})
}

func (p *StdPre) BeforePrint(Writer) error {
	_, err := p.resolvedLines()
	return err
}

func (p *StdPre) DefaultAttrs(HasScope) map[string]string {
	return map[string]string{"font": "fixed"}
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
	if fixed, ok := defaultStyles["fixed"].(*FontStyle); ok && fixed != nil {
		return fixed
	}
	return defaultFont
}

func (p *StdPre) PreferredHeight(w Writer) (float64, error) {
	if p.height != 0 {
		return float64(p.height), nil
	}
	lines := p.Lines()
	if len(lines) == 0 {
		lines = []string{""}
	}
	return float64(len(lines))*p.lineHeight(w) + NonContentHeight(p), nil
}

func (p *StdPre) PreferredWidth(w Writer) (float64, error) {
	if p.width != 0 {
		return float64(p.width), nil
	}
	applyWidgetFont(w, p)
	maxWidth := 0.0
	for _, line := range p.Lines() {
		rt, err := p.richTextForLine(line, w)
		if err != nil {
			return 0, err
		}
		if width := rt.Width(); width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth + NonContentWidth(p), nil
}

func (p *StdPre) AccessibilityText() string {
	lines, err := p.resolvedLines()
	if err != nil {
		return ""
	}
	return strings.Join(lines, "\n")
}

func (p *StdPre) SetAttrs(attrs map[string]string) {
	p.StdWidget.SetAttrs(attrs)
	p.source.SetAttrs(attrs)
}

func (p *StdPre) Lines() []string {
	lines, err := p.resolvedLines()
	if err != nil {
		return []string{""}
	}
	return lines
}

func (p *StdPre) resolvedLines() ([]string, error) {
	if p.source.Explicit() {
		text, err := p.source.Text(p.doc, p.container, p.rawText, "pre")
		if err != nil {
			return nil, err
		}
		return normalizedPreLines(text), nil
	}
	if p.lines == nil {
		p.lines = normalizedPreLines(p.rawText)
	}
	return p.lines, nil
}

func (p *StdPre) String() string {
	return fmt.Sprintf("StdPre src=%s %s", p.source.src, &p.StdWidget)
}

func (p *StdPre) lineHeight(w Writer) float64 {
	rt, err := p.richTextForLine("M", w)
	if err != nil {
		return effectiveFontSizeForWidget(p) * w.LineSpacing()
	}
	return rt.Leading() * w.LineSpacing()
}

func (p *StdPre) richTextForLine(line string, w Writer) (*rich_text.RichText, error) {
	font := p.Font()
	applyWidgetFont(w, p)
	rt, err := rich_text.New(line, w.Fonts(), w.FontSize(), font.RichTextOptions())
	if err != nil {
		debugf("StdPre.richTextForLine: %v", err)
		return nil, err
	}
	return rt, nil
}

func init() {
	registerTag(DefaultSpace, "pre", func() any { return &StdPre{} })
}

func normalizedPreLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.ReplaceAll(text, "\t", strings.Repeat(" ", preTabWidth))
	if text == "" {
		return []string{""}
	}

	lines := strings.Split(text, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return []string{""}
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
	return lines
}

var _ HasAttrs = (*StdPre)(nil)
var _ HasText = (*StdPre)(nil)
var _ Identifier = (*StdPre)(nil)
var _ Printer = (*StdPre)(nil)
var _ WantsContainer = (*StdPre)(nil)
var _ WantsDoc = (*StdPre)(nil)
var _ HasDefaultAttrs = (*StdPre)(nil)
