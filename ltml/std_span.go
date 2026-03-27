// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdSpan struct {
	StdContainer
}

func (s *StdSpan) AddText(text string) {
	s.AddTextWithFont(text, s.Font())
}

func (s *StdSpan) AddTextWithFont(text string, font *FontStyle) {
	s.container.(AddTextWithFonter).AddTextWithFont(text, font)
}

func (s *StdSpan) AddInlineWithFont(content inlineText, font *FontStyle) {
	s.container.(AddInlineWithFonter).AddInlineWithFont(content, font)
}

func (s *StdSpan) SetContainer(container Container) error {
	switch container.(type) {
	case *StdSpan:
	case *StdParagraph:
	case *StdLabel:
	default:
		return fmt.Errorf("span must be child of p, label or another span")
	}
	s.container = container
	return nil
}

func init() {
	registerTag(DefaultSpace, "span", func() interface{} { return &StdSpan{} })
}
