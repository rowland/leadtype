// Copyright 2016 Brent Rowland.
// Use of this source code is governed the Apache License, Version 2.0, as described in the LICENSE file.

package ltml

import (
	"fmt"
)

type StdSpan struct {
	StdContainer
}

type AddTextWithFonter interface {
	AddTextWithFont(text string, font *FontStyle)
}

func (s *StdSpan) AddText(text string) {
	s.AddTextWithFont(text, s.Font())
}

func (s *StdSpan) AddTextWithFont(text string, font *FontStyle) {
	s.container.(AddTextWithFonter).AddTextWithFont(text, font)
}

func (s *StdSpan) SetContainer(container Container) error {
	switch container.(type) {
	case *StdSpan:
	case *StdParagraph:
	default:
		return fmt.Errorf("span must be child of p, pabel or another span.")
	}
	s.container = container
	return nil
}

func init() {
	registerTag(DefaultSpace, "span", func() interface{} { return &StdSpan{} })
}
